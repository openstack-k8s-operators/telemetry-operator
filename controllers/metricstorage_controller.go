/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"regexp"

	discoveryv1 "k8s.io/api/discovery/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/apimachinery/pkg/api/meta"

	logr "github.com/go-logr/logr"
	common "github.com/openstack-k8s-operators/lib-common/modules/common"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"

	dataplanev1 "github.com/openstack-k8s-operators/dataplane-operator/api/v1beta1"
	infranetworkv1 "github.com/openstack-k8s-operators/infra-operator/apis/network/v1beta1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	ceilometer "github.com/openstack-k8s-operators/telemetry-operator/pkg/ceilometer"
	metricstorage "github.com/openstack-k8s-operators/telemetry-operator/pkg/metricstorage"
	monv1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1"
	obov1 "github.com/rhobs/observability-operator/pkg/apis/monitoring/v1alpha1"
)

// MetricStorageReconciler reconciles a MetricStorage object
type MetricStorageReconciler struct {
	client.Client
	Kclient    kubernetes.Interface
	Scheme     *runtime.Scheme
	Controller controller.Controller
	Watching   []string
	RESTMapper meta.RESTMapper
	Cache      cache.Cache
}

// GetLogger returns a logger object with a prefix of "conroller.name" and aditional controller context fields
func (r *MetricStorageReconciler) GetLogger(ctx context.Context) logr.Logger {
	return log.FromContext(ctx).WithName("Controllers").WithName("MetricStorage")
}

//+kubebuilder:rbac:groups=telemetry.openstack.org,resources=metricstorages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=telemetry.openstack.org,resources=metricstorages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=telemetry.openstack.org,resources=metricstorages/finalizers,verbs=update
//+kubebuilder:rbac:groups=monitoring.rhobs,resources=monitoringstacks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitoring.rhobs,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitoring.rhobs,resources=scrapeconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=network.openstack.org,resources=ipsets,verbs=get;list;watch
//+kubebuilder:rbac:groups=dataplane.openstack.org,resources=openstackdataplanenodesets,verbs=get;list;watch
//+kubebuilder:rbac:groups=dataplane.openstack.org,resources=openstackdataplaneservices,verbs=get;list;watch

// Reconcile reconciles MetricStorage
func (r *MetricStorageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)

	// Fetch the MetricStorage instance
	instance := &telemetryv1.MetricStorage{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected.
			// For additional cleanup logic use finalizers. Return and don't requeue.
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	helper, err := helper.NewHelper(
		instance,
		r.Client,
		r.Kclient,
		r.Scheme,
		Log,
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Always patch the instance status when exiting this function so we can persist any changes.
	defer func() {
		// update the Ready condition based on the sub conditions
		if instance.Status.Conditions.AllSubConditionIsTrue() {
			instance.Status.Conditions.MarkTrue(
				condition.ReadyCondition, condition.ReadyMessage)
		} else {
			// something is not ready so reset the Ready condition
			instance.Status.Conditions.MarkUnknown(
				condition.ReadyCondition, condition.InitReason, condition.ReadyInitMessage)
			// and recalculate it based on the state of the rest of the conditions
			instance.Status.Conditions.Set(
				instance.Status.Conditions.Mirror(condition.ReadyCondition))
		}
		err := helper.PatchInstance(ctx, instance)
		if err != nil {
			return
		}
	}()

	// If we're not deleting this and the service object doesn't have our finalizer, add it.
	if instance.DeletionTimestamp.IsZero() && controllerutil.AddFinalizer(instance, helper.GetFinalizer()) {
		return ctrl.Result{}, nil
	}

	//
	// initialize status
	//
	if instance.Status.Conditions == nil {
		instance.Status.Conditions = condition.Conditions{}
		// initialize conditions used later as Status=Unknown
		cl := condition.CreateList(
			condition.UnknownCondition(telemetryv1.MonitoringStackReadyCondition, condition.InitReason, telemetryv1.MonitoringStackReadyInitMessage),
			condition.UnknownCondition(telemetryv1.ServiceMonitorReadyCondition, condition.InitReason, telemetryv1.ServiceMonitorReadyInitMessage),
			condition.UnknownCondition(telemetryv1.ScrapeConfigReadyCondition, condition.InitReason, telemetryv1.ScrapeConfigReadyInitMessage),
			condition.UnknownCondition(telemetryv1.NodeSetReadyCondition, condition.InitReason, telemetryv1.NodeSetReadyInitMessage),
		)

		instance.Status.Conditions.Init(&cl)

		// Register overall status immediately to have an early feedback e.g. in the cli
		return ctrl.Result{}, nil
	}

	// Handle service delete
	if !instance.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, instance, helper)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, instance, helper)
}

func (r *MetricStorageReconciler) reconcileDelete(
	ctx context.Context,
	instance *telemetryv1.MetricStorage,
	helper *helper.Helper,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info("Reconciling Service delete")
	// Service is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(instance, helper.GetFinalizer())
	Log.Info(fmt.Sprintf("Reconciled Service '%s' delete successfully", instance.Name))

	return ctrl.Result{}, nil
}

func (r *MetricStorageReconciler) reconcileNormal(
	ctx context.Context,
	instance *telemetryv1.MetricStorage,
	helper *helper.Helper,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info(fmt.Sprintf("Reconciling Service '%s'", instance.Name))

	var eventHandler handler.EventHandler = handler.EnqueueRequestForOwner(
		r.Scheme,
		r.RESTMapper,
		&telemetryv1.MetricStorage{},
		handler.OnlyControllerOwner(),
	)

	serviceLabels := map[string]string{
		common.AppSelector: "metricStorage",
	}

	if instance.Spec.CustomMonitoringStack == nil && instance.Spec.MonitoringStack == nil {
		Log.Info("Both fields: \"customMonitoringStack\", \"monitoringStack\" aren't set. Setting at least one is required.")
		instance.Status.Conditions.MarkFalse(telemetryv1.MonitoringStackReadyCondition,
			condition.Reason("MonitoringStack isn't configured properly"),
			condition.SeverityError,
			telemetryv1.MonitoringStackReadyMisconfiguredMessage, "Either \"customMonitoringStack\" or \"monitoringStack\" must be set, but both are nil.")
		return ctrl.Result{}, nil
	}

	// Deploy monitoring stack

	err := r.ensureWatches(ctx, "monitoringstacks.monitoring.rhobs", &obov1.MonitoringStack{}, eventHandler)
	if err != nil {
		instance.Status.Conditions.MarkFalse(telemetryv1.MonitoringStackReadyCondition,
			condition.Reason("Can't own MonitoringStack resource"),
			condition.SeverityError,
			telemetryv1.MonitoringStackUnableToOwnMessage, err)
		Log.Info("Can't own MonitoringStack resource")
		return ctrl.Result{RequeueAfter: telemetryv1.PauseBetweenWatchAttempts}, nil
	}

	monitoringStack := &obov1.MonitoringStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}
	op, err := controllerutil.CreateOrPatch(ctx, r.Client, monitoringStack, func() error {
		if reflect.DeepEqual(instance.Spec.CustomMonitoringStack, &obov1.MonitoringStackSpec{}) || instance.Spec.CustomMonitoringStack == nil {
			Log.Info(fmt.Sprintf("Using MetricStorage exposed options for MonitoringStack %s definition", monitoringStack.Name))
			desiredMonitoringStack, err := metricstorage.MonitoringStack(instance, serviceLabels)
			if err != nil {
				return err
			}
			desiredMonitoringStack.Spec.DeepCopyInto(&monitoringStack.Spec)
		} else {
			Log.Info(fmt.Sprintf("Using CustomMonitoringStack for MonitoringStack %s definition", monitoringStack.Name))
			instance.Spec.CustomMonitoringStack.DeepCopyInto(&monitoringStack.Spec)
		}
		monitoringStack.ObjectMeta.Labels = serviceLabels
		err := controllerutil.SetControllerReference(instance, monitoringStack, r.Scheme)
		return err
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		Log.Info(fmt.Sprintf("MonitoringStack %s successfully changed - operation: %s", monitoringStack.Name, string(op)))
	}
	monitoringStackReady := true
	for _, c := range monitoringStack.Status.Conditions {
		if c.Status != "True" {
			instance.Status.Conditions.MarkFalse(telemetryv1.MonitoringStackReadyCondition,
				condition.Reason(c.Reason),
				condition.SeverityError,
				c.Message)
			monitoringStackReady = false
			break
		}
	}
	if len(monitoringStack.Status.Conditions) == 0 {
		monitoringStackReady = false
	}
	if monitoringStackReady {
		instance.Status.Conditions.MarkTrue(telemetryv1.MonitoringStackReadyCondition, condition.ReadyMessage)
	}

	// Deploy ServiceMonitor for ceilometer monitoring
	err = r.ensureWatches(ctx, "servicemonitors.monitoring.rhobs", &monv1.ServiceMonitor{}, eventHandler)

	if err != nil {
		instance.Status.Conditions.MarkFalse(telemetryv1.ServiceMonitorReadyCondition,
			condition.Reason("Can't own ServiceMonitor resource"),
			condition.SeverityError,
			telemetryv1.ServiceMonitorUnableToOwnMessage, err)
		Log.Info("Can't own ServiceMonitor resource")
		return ctrl.Result{RequeueAfter: telemetryv1.PauseBetweenWatchAttempts}, nil
	}
	ceilometerMonitor := &monv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}
	op, err = controllerutil.CreateOrPatch(ctx, r.Client, ceilometerMonitor, func() error {
		ceilometerLabels := map[string]string{
			common.AppSelector: ceilometer.ServiceName,
		}
		desiredCeilometerMonitor := metricstorage.ServiceMonitor(instance, serviceLabels, ceilometerLabels)
		desiredCeilometerMonitor.Spec.DeepCopyInto(&ceilometerMonitor.Spec)
		ceilometerMonitor.ObjectMeta.Labels = desiredCeilometerMonitor.ObjectMeta.Labels
		err = controllerutil.SetControllerReference(instance, ceilometerMonitor, r.Scheme)
		return err
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		Log.Info(fmt.Sprintf("Ceilometer ServiceMonitor %s successfully changed - operation: %s", ceilometerMonitor.Name, string(op)))
	}
	instance.Status.Conditions.MarkTrue(telemetryv1.ServiceMonitorReadyCondition, condition.ReadyMessage)
	// Deploy ScrapeConfig for NodeExporter monitoring
	nodeSetWatchFn := func(ctx context.Context, o client.Object) []reconcile.Request {
		// Reconcile all metricstorages when a nodeset changes
		result := []reconcile.Request{}

		// get all MetricStorage CRs
		metricstorages := &telemetryv1.MetricStorageList{}
		listOpts := []client.ListOption{
			client.InNamespace(o.GetNamespace()),
		}
		if err := r.Client.List(context.Background(), metricstorages, listOpts...); err != nil {
			Log.Error(err, "Unable to retrieve MetricStorage CRs %v")
			return nil
		}
		for _, cr := range metricstorages.Items {
			name := client.ObjectKey{
				Namespace: o.GetNamespace(),
				Name:      cr.Name,
			}
			result = append(result, reconcile.Request{NamespacedName: name})
		}
		if len(result) > 0 {
			return result
		}
		return nil
	}
	err = r.ensureWatches(ctx, "openstackdataplanenodesets.dataplane.openstack.org", &dataplanev1.OpenStackDataPlaneNodeSet{}, handler.EnqueueRequestsFromMapFunc(nodeSetWatchFn))
	if err != nil {
		instance.Status.Conditions.MarkFalse(telemetryv1.NodeSetReadyCondition,
			condition.Reason("Can't watch NodeSet resource"),
			condition.SeverityError,
			telemetryv1.NodeSetUnableToWatchMessage, err)
		Log.Info("Can't watch OpenStackDataPlaneNodeSet resource")
		return ctrl.Result{RequeueAfter: telemetryv1.PauseBetweenWatchAttempts}, nil
	}
	instance.Status.Conditions.MarkTrue(telemetryv1.NodeSetReadyCondition, condition.ReadyMessage)

	// TODO: Move to structured once OBO version is bumped
	scrapeConfig := &unstructured.Unstructured{}
	scrapeConfig.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "monitoring.rhobs",
		Version: "v1alpha1",
		Kind:    "ScrapeConfig",
	})
	scrapeConfig.SetName(instance.Name)
	scrapeConfig.SetNamespace(instance.Namespace)

	err = r.ensureWatches(ctx, "scrapeconfigs.monitoring.rhobs", scrapeConfig, eventHandler)

	if err != nil {
		instance.Status.Conditions.MarkFalse(telemetryv1.ScrapeConfigReadyCondition,
			condition.Reason("Can't own ScrapeConfig resource"),
			condition.SeverityError,
			telemetryv1.ScrapeConfigUnableToOwnMessage, err)
		Log.Info("Can't own ScrapeConfig resource")
		return ctrl.Result{RequeueAfter: telemetryv1.PauseBetweenWatchAttempts}, nil
	}
	op, err = controllerutil.CreateOrPatch(ctx, r.Client, scrapeConfig, func() error {
		endpoints, err := getNodeExporterTargets(instance, helper)
		if err != nil {
			return err
		}
		desiredScrapeConfig := metricstorage.ScrapeConfig(instance, serviceLabels, endpoints)
		scrapeConfig.UnstructuredContent()["spec"] = desiredScrapeConfig.UnstructuredContent()["spec"]
		scrapeConfig.SetLabels(desiredScrapeConfig.GetLabels())
		err = controllerutil.SetControllerReference(instance, scrapeConfig, r.Scheme)
		return err
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		Log.Info(fmt.Sprintf("Node Exporter ScrapeConfig %s successfully changed - operation: %s", scrapeConfig.GetName(), string(op)))
	}
	instance.Status.Conditions.MarkTrue(telemetryv1.ScrapeConfigReadyCondition, condition.ReadyMessage)

	Log.Info("Reconciled Service successfully")
	return ctrl.Result{}, nil
}

func (r *MetricStorageReconciler) ensureWatches(
	ctx context.Context,
	name string,
	kind client.Object,
	handler handler.EventHandler,
) error {
	Log := r.GetLogger(ctx)
	for _, item := range r.Watching {
		if item == name {
			// We are already watching the resource
			return nil
		}
	}
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apiextensions.k8s.io",
		Kind:    "CustomResourceDefinition",
		Version: "v1",
	})

	err := r.Client.Get(context.Background(), client.ObjectKey{
		Name: name,
	}, u)
	if err != nil {
		return err
	}

	Log.Info(fmt.Sprintf("Starting to watch %s", name))
	err = r.Controller.Watch(source.Kind(r.Cache, kind),
		handler,
	)
	if err == nil {
		r.Watching = append(r.Watching, name)
	}
	return err
}

func getNodeExporterTargets(
	instance *telemetryv1.MetricStorage,
	helper *helper.Helper,
) ([]string, error) {
	ipSetList, err := getIPSetList(instance, helper)
	if err != nil {
		return []string{}, err
	}
	nodeSetList, err := getNodeSetList(instance, helper)
	if err != nil {
		return []string{}, err
	}
	var address string
	addresses := []string{}
	for _, nodeSet := range nodeSetList.Items {
		telemetryServiceInNodeSet := false
		for _, service := range nodeSet.Spec.Services {
			if service == "telemetry" {
				telemetryServiceInNodeSet = true
				break
			}
		}
		if !telemetryServiceInNodeSet {
			// Telemetry isn't deployed on this nodeset
			// there is no reason to include these nodes
			// for scraping by prometheus
			continue
		}
		for name, item := range nodeSet.Spec.Nodes {
			namespacedName := &types.NamespacedName{
				Name:      name,
				Namespace: instance.GetNamespace(),
			}

			if len(ipSetList.Items) > 0 {
				// if we have IPSets, lets go to search for the IPs there
				address, _ = getAddressFromIPSet(instance, &item, namespacedName, helper)
			} else if len(item.Ansible.AnsibleHost) > 0 {
				address, _ = getAddressFromAnsibleHost(&item)
			} else {
				// we were unable to find an IP or HostName for a node, so we do not go further
				return addresses, nil
			}
			if address == "" {
				// we were unable to find an IP or HostName for a node, so we do not go further
				return addresses, nil
			}
			addresses = append(addresses, fmt.Sprintf("%s:%d", address, telemetryv1.DefaultNodeExporterPort))
		}
	}
	return addresses, nil
}

func getIPSetList(instance *telemetryv1.MetricStorage, helper *helper.Helper) (*infranetworkv1.IPSetList, error) {
	ipSets := &infranetworkv1.IPSetList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.GetNamespace()),
	}
	err := helper.GetClient().List(context.Background(), ipSets, listOpts...)
	return ipSets, err
}

func getNodeSetList(instance *telemetryv1.MetricStorage, helper *helper.Helper) (*dataplanev1.OpenStackDataPlaneNodeSetList, error) {
	nodeSets := &dataplanev1.OpenStackDataPlaneNodeSetList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.GetNamespace()),
	}
	err := helper.GetClient().List(context.Background(), nodeSets, listOpts...)
	return nodeSets, err
}

func getAddressFromIPSet(
	instance *telemetryv1.MetricStorage,
	item *dataplanev1.NodeSection,
	namespacedName *types.NamespacedName,
	helper *helper.Helper,
) (string, discoveryv1.AddressType) {
	// we go search for an IPSet
	ipset := &infranetworkv1.IPSet{}
	err := helper.GetClient().Get(context.Background(), *namespacedName, ipset)
	if err != nil {
		// No IPsets found, lets try to get the HostName as last resource
		if isValidDomain(item.HostName) {
			return item.HostName, discoveryv1.AddressTypeFQDN
		}
		// No IP address or valid hostname found anywhere
		helper.GetLogger().Info("Did not found a valid hostname or IP address")
		return "", ""
	}
	// check that the reservations list is not empty
	if len(ipset.Status.Reservation) > 0 {
		// search for the network specified in the Spec
		for _, reservation := range ipset.Status.Reservation {
			if reservation.Network == instance.Spec.MonitoringStack.DataplaneNetwork {
				return reservation.Address, discoveryv1.AddressTypeIPv4
			}
		}
	}
	// if the reservations list is empty, we go find if AnsibleHost exists
	return getAddressFromAnsibleHost(item)
}

func getAddressFromAnsibleHost(item *dataplanev1.NodeSection) (string, discoveryv1.AddressType) {
	// check if ansiblehost is an IP
	addr := net.ParseIP(item.Ansible.AnsibleHost)
	if addr != nil {
		// it is an ip
		return item.Ansible.AnsibleHost, discoveryv1.AddressTypeIPv4
	}
	// it is not an ip, is it a valid hostname?
	if isValidDomain(item.Ansible.AnsibleHost) {
		// it is an valid domain name
		return item.Ansible.AnsibleHost, discoveryv1.AddressTypeFQDN
	}
	// if the reservations list is empty, we go find if HostName is a valid domain
	if isValidDomain(item.HostName) {
		return item.HostName, discoveryv1.AddressTypeFQDN
	}
	return "", ""
}

// isValidDomain returns true if the domain is valid.
func isValidDomain(domain string) bool {
	domainRegexp := regexp.MustCompile(`^(?i)[a-z0-9-]+(\.[a-z0-9-]+)+\.?$`)
	return domainRegexp.MatchString(domain)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MetricStorageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	control, err := ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1.MetricStorage{}).
		Build(r)
	r.Controller = control
	return err
}
