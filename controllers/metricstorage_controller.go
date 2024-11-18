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

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/apimachinery/pkg/api/meta"

	logr "github.com/go-logr/logr"
	"github.com/openstack-k8s-operators/lib-common/modules/ansible"
	common "github.com/openstack-k8s-operators/lib-common/modules/common"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	object "github.com/openstack-k8s-operators/lib-common/modules/common/object"
	tls "github.com/openstack-k8s-operators/lib-common/modules/common/tls"

	infranetworkv1 "github.com/openstack-k8s-operators/infra-operator/apis/network/v1beta1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	availability "github.com/openstack-k8s-operators/telemetry-operator/pkg/availability"
	ceilometer "github.com/openstack-k8s-operators/telemetry-operator/pkg/ceilometer"
	"github.com/openstack-k8s-operators/telemetry-operator/pkg/dashboards"
	metricstorage "github.com/openstack-k8s-operators/telemetry-operator/pkg/metricstorage"
	telemetry "github.com/openstack-k8s-operators/telemetry-operator/pkg/telemetry"
	rabbitmqv1 "github.com/rabbitmq/cluster-operator/api/v1beta1"
	monv1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1"
	monv1alpha1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1alpha1"
	obov1 "github.com/rhobs/observability-operator/pkg/apis/monitoring/v1alpha1"
	obsui "github.com/rhobs/observability-operator/pkg/apis/uiplugin/v1alpha1"
)

// fields to index to reconcile when change
const (
	prometheusCaBundleSecretNameField = ".spec.prometheusTls.caBundleSecretName"
	prometheusTLSField                = ".spec.prometheusTls.secretName"
)

var (
	prometheusAllWatchFields = []string{
		prometheusCaBundleSecretNameField,
		prometheusTLSField,
	}
	serviceLabels = map[string]string{
		common.AppSelector: "metricStorage",
	}
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

// ConnectionInfo holds information about connection to a compute node
type ConnectionInfo struct {
	IP       string
	Hostname string
	TLS      bool
	FQDN     string
}

// GetLogger returns a logger object with a prefix of "conroller.name" and aditional controller context fields
func (r *MetricStorageReconciler) GetLogger(ctx context.Context) logr.Logger {
	return log.FromContext(ctx).WithName("Controllers").WithName("MetricStorage")
}

//+kubebuilder:rbac:groups=telemetry.openstack.org,resources=metricstorages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=telemetry.openstack.org,resources=metricstorages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=telemetry.openstack.org,resources=metricstorages/finalizers,verbs=update;patch
//+kubebuilder:rbac:groups=monitoring.rhobs,resources=monitoringstacks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitoring.rhobs,resources=servicemonitors,verbs=get;list;delete
//+kubebuilder:rbac:groups=monitoring.rhobs,resources=scrapeconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitoring.rhobs,resources=prometheusrules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitoring.rhobs,resources=prometheuses,verbs=get;list;watch;update;patch;delete
//+kubebuilder:rbac:groups=monitoring.rhobs,resources=alertmanagers,verbs=get;list;watch;update;patch;delete
//+kubebuilder:rbac:groups=network.openstack.org,resources=ipsets,verbs=get;list;watch
//+kubebuilder:rbac:groups=rabbitmq.com,resources=rabbitmqclusters,verbs=get;list;watch
//+kubebuilder:rbac:groups=observability.openshift.io,resources=uiplugins,verbs=get;list;watch;create;patch
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile reconciles MetricStorage
func (r *MetricStorageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, _err error) {
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

	// initialize status if Conditions is nil, but do not reset if it already
	// exists
	isNewInstance := instance.Status.Conditions == nil
	if isNewInstance {
		instance.Status.Conditions = condition.Conditions{}
	}

	// Save a copy of the condtions so that we can restore the LastTransitionTime
	// when a condition's state doesn't change.
	savedConditions := instance.Status.Conditions.DeepCopy()

	// Always patch the instance status when exiting this function so we can
	// persist any changes.
	defer func() {
		condition.RestoreLastTransitionTimes(
			&instance.Status.Conditions, savedConditions)
		if instance.Status.Conditions.IsUnknown(condition.ReadyCondition) {
			instance.Status.Conditions.Set(
				instance.Status.Conditions.Mirror(condition.ReadyCondition))
		}
		err := helper.PatchInstance(ctx, instance)
		if err != nil {
			_err = err
			return
		}
	}()

	//
	// initialize status
	//
	cl := condition.CreateList(
		condition.UnknownCondition(condition.ReadyCondition, condition.InitReason, condition.ReadyInitMessage),
		condition.UnknownCondition(telemetryv1.MonitoringStackReadyCondition, condition.InitReason, telemetryv1.MonitoringStackReadyInitMessage),
		condition.UnknownCondition(telemetryv1.ScrapeConfigReadyCondition, condition.InitReason, telemetryv1.ScrapeConfigReadyInitMessage),
		condition.UnknownCondition(telemetryv1.DashboardPrometheusRuleReadyCondition, condition.InitReason, telemetryv1.DashboardPrometheusRuleReadyInitMessage),
		condition.UnknownCondition(telemetryv1.DashboardPluginReadyCondition, condition.InitReason, telemetryv1.DashboardPluginReadyInitMessage),
		condition.UnknownCondition(telemetryv1.DashboardDatasourceReadyCondition, condition.InitReason, telemetryv1.DashboardDatasourceReadyInitMessage),
		condition.UnknownCondition(telemetryv1.DashboardDefinitionReadyCondition, condition.InitReason, telemetryv1.DashboardDefinitionReadyInitMessage),
		condition.UnknownCondition(telemetryv1.PrometheusReadyCondition, condition.InitReason, telemetryv1.PrometheusReadyInitMessage),
		condition.UnknownCondition(condition.TLSInputReadyCondition, condition.InitReason, condition.InputReadyInitMessage),
	)

	instance.Status.Conditions.Init(&cl)
	instance.Status.ObservedGeneration = instance.Generation

	// If we're not deleting this and the service object doesn't have our finalizer, add it.
	if instance.DeletionTimestamp.IsZero() && controllerutil.AddFinalizer(instance, helper.GetFinalizer()) || isNewInstance {
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

	if res, err := metricstorage.DeleteDashboardObjects(ctx, instance, helper); err != nil {
		return res, err
	}

	// Service is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(instance, helper.GetFinalizer())
	Log.Info(fmt.Sprintf("Reconciled Service '%s' delete successfully", instance.Name))

	return ctrl.Result{}, nil
}

// TODO: call the function appropriately
//
//nolint:all
func (r *MetricStorageReconciler) reconcileUpdate(
	ctx context.Context,
	instance *telemetryv1.MetricStorage,
	helper *helper.Helper,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info("Reconciling Service update")

	err := r.deleteOldServiceMonitors(ctx, instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	Log.Info(fmt.Sprintf("Reconciled Service '%s' update successfully", instance.Name))

	return ctrl.Result{}, nil
}

// Delete old ServiceMonitors
// ServiceMonitors were used when deploying with RHOSO 18 GA telemetry-operator.
//
//nolint:all
func (r *MetricStorageReconciler) deleteOldServiceMonitors(
	ctx context.Context,
	instance *telemetryv1.MetricStorage,
) error {
	monitorList := &monv1.ServiceMonitorList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.GetNamespace()),
	}
	err := r.Client.List(ctx, monitorList, listOpts...)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return err
	}
	for _, monitor := range monitorList.Items {
		if object.CheckOwnerRefExist(instance.ObjectMeta.UID, monitor.ObjectMeta.OwnerReferences) {
			err = r.Client.Delete(ctx, monitor)
			if err != nil {
				return err
			}
		}
	}
	return nil
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
			condition.Reason("Can't own MonitoringStack resource. The Cluster Observability Operator probably isn't installed"),
			condition.SeverityError,
			telemetryv1.MonitoringStackUnableToOwnMessage, err)
		Log.Info("Can't own MonitoringStack resource. The Cluster Observability Operator probably isn't installed")
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

	if instance.Spec.PrometheusTLS.Enabled() {
		// Patch Prometheus to add TLS
		prometheusWatchFn := func(_ context.Context, o client.Object) []reconcile.Request {
			name := client.ObjectKey{
				Namespace: o.GetNamespace(),
				Name:      o.GetName(),
			}
			return []reconcile.Request{{NamespacedName: name}}
		}
		err = r.ensureWatches(ctx, "prometheuses.monitoring.rhobs", &monv1.Prometheus{}, handler.EnqueueRequestsFromMapFunc(prometheusWatchFn))
		if err != nil {
			instance.Status.Conditions.MarkFalse(telemetryv1.PrometheusReadyCondition,
				condition.Reason("Can't watch prometheus resource. The Cluster Observability Operator probably isn't installed"),
				condition.SeverityError,
				telemetryv1.PrometheusUnableToWatchMessage, err)
			Log.Info("Can't watch Prometheus resource. The Cluster Observability Operator probably isn't installed")
			return ctrl.Result{RequeueAfter: telemetryv1.PauseBetweenWatchAttempts}, nil
		}
		prometheusTLSPatch := metricstorage.PrometheusTLS(instance)
		err = r.Client.Patch(context.Background(), &prometheusTLSPatch, client.Apply, client.FieldOwner("telemetry-operator"))
		if err != nil {
			Log.Error(err, "Can't patch Prometheus resource")
			return ctrl.Result{}, err
		}
		instance.Status.PrometheusTLSPatched = true
	} else if instance.Status.PrometheusTLSPatched {
		// Delete the prometheus CR, so it can be automatically restored without the TLS patch
		prometheus := monv1.Prometheus{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: instance.Namespace,
				Name:      instance.Name,
			},
		}
		err = r.Client.Delete(context.Background(), &prometheus)
		if err != nil && !k8s_errors.IsNotFound(err) {
			instance.Status.Conditions.MarkFalse(telemetryv1.PrometheusReadyCondition,
				condition.Reason("Can't delete old Prometheus CR to remove TLS configuration"),
				condition.SeverityError,
				telemetryv1.PrometheusUnableToRemoveTLSMessage, err)
			Log.Error(err, "Can't delete old Prometheus CR to remove TLS configuration")
			return ctrl.Result{}, err
		}
		instance.Status.PrometheusTLSPatched = false
	}
	instance.Status.Conditions.MarkTrue(telemetryv1.PrometheusReadyCondition, condition.ReadyMessage)

	// Patch Prometheus service to add route creation
	prometheusServicePatch := metricstorage.PrometheusService(instance)
	err = r.Client.Patch(context.Background(), &prometheusServicePatch, client.Apply, client.FieldOwner("telemetry-operator"))
	if err != nil {
		Log.Error(err, "Can't patch Prometheus service resource")
		return ctrl.Result{}, err
	}

	// Patch Alertmanager service to add route creation
	if instance.Spec.MonitoringStack != nil && instance.Spec.MonitoringStack.AlertingEnabled {
		alertmanagerServicePatch := metricstorage.AlertmanagerService(instance)
		err = r.Client.Patch(context.Background(), &alertmanagerServicePatch, client.Apply, client.FieldOwner("telemetry-operator"))
		if err != nil {
			Log.Error(err, "Can't patch Alertmanager service resource")
			return ctrl.Result{}, err
		}
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

	// Deploy ScrapeConfigs
	if res, err := r.createScrapeConfigs(ctx, instance, eventHandler, helper); err != nil {
		return res, err
	}

	if !instance.Spec.DashboardsEnabled {
		if res, err := metricstorage.DeleteDashboardObjects(ctx, instance, helper); err != nil {
			return res, err
		}
		instance.Status.Conditions.MarkTrue(telemetryv1.DashboardPrometheusRuleReadyCondition, telemetryv1.DashboardsNotEnabledMessage)
		instance.Status.Conditions.MarkTrue(telemetryv1.DashboardDatasourceReadyCondition, telemetryv1.DashboardsNotEnabledMessage)
		instance.Status.Conditions.MarkTrue(telemetryv1.DashboardDefinitionReadyCondition, telemetryv1.DashboardsNotEnabledMessage)
		instance.Status.Conditions.MarkTrue(telemetryv1.DashboardPluginReadyCondition, telemetryv1.DashboardsNotEnabledMessage)
	} else {
		if res, err := r.createDashboardObjects(ctx, instance, helper, eventHandler); err != nil {
			return res, err
		}
	}
	//
	// TLS input validation
	//
	// Validate the CA cert secret if provided
	if instance.Spec.PrometheusTLS.CaBundleSecretName != "" {
		_, err := tls.ValidateCACertSecret(
			ctx,
			helper.GetClient(),
			types.NamespacedName{
				Name:      instance.Spec.PrometheusTLS.CaBundleSecretName,
				Namespace: instance.Namespace,
			},
		)
		if err != nil {
			if k8s_errors.IsNotFound(err) {
				instance.Status.Conditions.Set(condition.FalseCondition(
					condition.TLSInputReadyCondition,
					condition.RequestedReason,
					condition.SeverityInfo,
					fmt.Sprintf(condition.TLSInputReadyWaitingMessage, instance.Spec.PrometheusTLS.CaBundleSecretName)))
				return ctrl.Result{}, nil
			}
			instance.Status.Conditions.Set(condition.FalseCondition(
				condition.TLSInputReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.TLSInputErrorMessage,
				err.Error()))
			return ctrl.Result{}, err
		}
	}

	// Validate API service certs secrets
	if instance.Spec.PrometheusTLS.Enabled() {
		_, err := instance.Spec.PrometheusTLS.ValidateCertSecret(ctx, helper, instance.Namespace)
		if err != nil {
			if k8s_errors.IsNotFound(err) {
				instance.Status.Conditions.Set(condition.FalseCondition(
					condition.TLSInputReadyCondition,
					condition.RequestedReason,
					condition.SeverityInfo,
					fmt.Sprintf(condition.TLSInputReadyWaitingMessage, err.Error())))
				return ctrl.Result{}, nil
			}
			instance.Status.Conditions.Set(condition.FalseCondition(
				condition.TLSInputReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.TLSInputErrorMessage,
				err.Error()))
			return ctrl.Result{}, err
		}
	}

	// all cert input checks out so report InputReady
	instance.Status.Conditions.MarkTrue(condition.TLSInputReadyCondition, condition.InputReadyMessage)

	if instance.Status.Conditions.AllSubConditionIsTrue() {
		instance.Status.Conditions.MarkTrue(
			condition.ReadyCondition, condition.ReadyMessage)
	}
	Log.Info("Reconciled Service successfully")
	return ctrl.Result{}, nil
}

func (r *MetricStorageReconciler) createServiceScrapeConfig(
	ctx context.Context,
	instance *telemetryv1.MetricStorage,
	log logr.Logger,
	description string,
	serviceName string,
	targets interface{},
	tlsEnabled bool,
) error {
	scrapeConfig := &monv1alpha1.ScrapeConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: instance.Namespace,
		},
	}
	op, err := controllerutil.CreateOrPatch(ctx, r.Client, scrapeConfig, func() error {
		desiredScrapeConfig := metricstorage.ScrapeConfig(instance,
			serviceLabels,
			targets,
			tlsEnabled)
		desiredScrapeConfig.Spec.DeepCopyInto(&scrapeConfig.Spec)
		scrapeConfig.ObjectMeta.Labels = desiredScrapeConfig.ObjectMeta.Labels
		err := controllerutil.SetControllerReference(instance, scrapeConfig, r.Scheme)
		return err
	})

	if err == nil && op != controllerutil.OperationResultNone {
		log.Info(fmt.Sprintf("%s ScrapeConfig %s successfully changed - operation: %s", description, scrapeConfig.GetName(), string(op)))
	}
	return err
}

func (r *MetricStorageReconciler) createScrapeConfigs(
	ctx context.Context,
	instance *telemetryv1.MetricStorage,
	eventHandler handler.EventHandler,
	helper *helper.Helper,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	err := r.ensureWatches(ctx, "scrapeconfigs.monitoring.rhobs", &monv1alpha1.ScrapeConfig{}, eventHandler)
	if err != nil {
		instance.Status.Conditions.MarkFalse(telemetryv1.ScrapeConfigReadyCondition,
			condition.Reason("Can't own ScrapeConfig resource. The Cluster Observability Operator probably isn't installed"),
			condition.SeverityError,
			telemetryv1.ScrapeConfigUnableToOwnMessage, err)
		Log.Info("Can't own ScrapeConfig resource. The Cluster Observability Operator probably isn't installed")
		return ctrl.Result{RequeueAfter: telemetryv1.PauseBetweenWatchAttempts}, nil
	}

	// ScrapeConfig for ceilometer monitoring
	ceilometerRoute := fmt.Sprintf("%s-internal.%s.svc", ceilometer.ServiceName, instance.Namespace)
	ceilometerTarget := []string{fmt.Sprintf("%s:%d", ceilometerRoute, ceilometer.CeilometerPrometheusPort)}
	ceilometerCfgName := fmt.Sprintf("%s-ceilometer", telemetry.ServiceName)
	err = r.createServiceScrapeConfig(ctx, instance, Log, "Ceilometer",
		ceilometerCfgName, ceilometerTarget, instance.Spec.PrometheusTLS.Enabled())
	if err != nil {
		return ctrl.Result{}, err
	}

	// ScrapeConfig for kube-state-metrics
	ksmRoute := fmt.Sprintf("%s.%s.svc", availability.KSMServiceName, instance.Namespace)
	ksmTarget := []string{fmt.Sprintf("%s:%d", ksmRoute, availability.KSMMetricsPort)}
	ksmCfgName := fmt.Sprintf("%s-ksm", telemetry.ServiceName)
	err = r.createServiceScrapeConfig(ctx, instance, Log, "kube-state-metrics",
		ksmCfgName, ksmTarget, instance.Spec.PrometheusTLS.Enabled())
	if err != nil {
		return ctrl.Result{}, err
	}

	// ScrapeConfigs for RabbitMQ monitoring
	// NOTE: We're watching Rabbits and reconciling with each of their change
	//       that should keep the targets inside the ScrapeConfig always
	//       up to date.
	rabbitList := &rabbitmqv1.RabbitmqClusterList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.GetNamespace()),
	}
	err = r.Client.List(ctx, rabbitList, listOpts...)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}
	rabbitTargets := []string{}
	for _, rabbit := range rabbitList.Items {
		rabbitServerName := fmt.Sprintf("%s.%s.svc", rabbit.Name, rabbit.Namespace)
		rabbitTargets = append(rabbitTargets, fmt.Sprintf("%s:%d", rabbitServerName, metricstorage.RabbitMQPrometheusPort))
	}
	rabbitCfgName := fmt.Sprintf("%s-rabbitmq", telemetry.ServiceName)
	err = r.createServiceScrapeConfig(ctx, instance, Log, "RabbitMQ",
		rabbitCfgName, rabbitTargets, instance.Spec.PrometheusTLS.Enabled())
	if err != nil {
		return ctrl.Result{}, err
	}

	connectionInfo, err := getComputeNodesConnectionInfo(instance, helper, telemetry.ServiceName)
	if err != nil {
		Log.Info(fmt.Sprintf("Cannot get compute node connection info. Scrape configs not created. Error: %s", err))
	}

	// ScrapeConfigs for NodeExporters
	neTargetsTLS, neTargetsNonTLS := getNodeExporterTargets(connectionInfo)
	// ScrapeConfig for non-tls nodes
	neServiceName := fmt.Sprintf("%s-node-exporter", telemetry.ServiceName)
	err = r.createServiceScrapeConfig(ctx, instance, Log, "Node Exporter",
		neServiceName, neTargetsNonTLS, false)
	if err != nil {
		return ctrl.Result{}, err
	}

	// ScrapeConfig for tls nodes
	neServiceName = fmt.Sprintf("%s-node-exporter-tls", telemetry.ServiceName)
	err = r.createServiceScrapeConfig(ctx, instance, Log, "Node Exporter",
		neServiceName, neTargetsTLS, true)
	if err != nil {
		return ctrl.Result{}, err
	}

	connectionInfo, err = getComputeNodesConnectionInfo(instance, helper, telemetryv1.TelemetryPowerMonitoring)
	if err != nil {
		Log.Info(fmt.Sprintf("Cannot get compute node connection info. Scrape configs not created. Error: %s", err))
	}

	// kepler scrape endpoints
	keplerEndpoints, _ := getKeplerTargets(connectionInfo)
	if err != nil {
		Log.Info(fmt.Sprintf("Cannot get Kepler targets. Scrape configs not created. Error: %s", err))
	}

	// keplerEndpoint is reported as empty slice when telemetry-power-monitoring service is not enabled
	if len(keplerEndpoints) > 0 {
		// Kepler ScrapeConfig for non-tls nodes
		keplerServiceName := fmt.Sprintf("%s-kepler", telemetry.ServiceName)
		err = r.createServiceScrapeConfig(ctx, instance, Log, "Kepler",
			keplerServiceName, keplerEndpoints, false) // Currently Kepler doesn't support TLS so tlsEnabled is set to false
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	instance.Status.Conditions.MarkTrue(telemetryv1.ScrapeConfigReadyCondition, condition.ReadyMessage)
	return ctrl.Result{}, err
}

func getNodeExporterTargets(nodes []ConnectionInfo) ([]metricstorage.LabeledTarget, []metricstorage.LabeledTarget) {
	tls := []metricstorage.LabeledTarget{}
	nonTLS := []metricstorage.LabeledTarget{}
	for _, node := range nodes {
		target := metricstorage.LabeledTarget{
			IP:   fmt.Sprintf("%s:%d", node.IP, telemetryv1.DefaultNodeExporterPort),
			FQDN: node.FQDN,
		}
		if node.TLS {
			tls = append(tls, target)
		} else {
			nonTLS = append(nonTLS, target)
		}
	}
	return tls, nonTLS
}

func getKeplerTargets(nodes []ConnectionInfo) ([]metricstorage.LabeledTarget, []metricstorage.LabeledTarget) {
	tls := []metricstorage.LabeledTarget{}
	nonTLS := []metricstorage.LabeledTarget{}
	for _, node := range nodes {
		target := metricstorage.LabeledTarget{
			IP:   fmt.Sprintf("%s:%d", node.IP, telemetryv1.DefaultKeplerPort),
			FQDN: node.FQDN,
		}
		if node.TLS {
			tls = append(tls, target)
		} else {
			nonTLS = append(nonTLS, target)
		}
	}
	return tls, nonTLS
}

func (r *MetricStorageReconciler) createDashboardObjects(ctx context.Context, instance *telemetryv1.MetricStorage, helper *helper.Helper, eventHandler handler.EventHandler) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	uiPluginObj := &obsui.UIPlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dashboards",
		},
	}
	op, err := controllerutil.CreateOrPatch(ctx, r.Client, uiPluginObj, func() error {
		uiPluginObj.Spec.Type = "Dashboards"
		return nil
	})
	if err != nil {
		Log.Error(err, fmt.Sprintf("Failed to update Dashboard Plugin definition %s - operation: %s", uiPluginObj.GetName(), string(op)))
		instance.Status.Conditions.MarkFalse(telemetryv1.DashboardPluginReadyCondition,
			condition.Reason("Can't create Dashboard Plugin definition"),
			condition.SeverityError,
			telemetryv1.DashboardPluginFailedMessage, err)
	} else {
		instance.Status.Conditions.MarkTrue(telemetryv1.DashboardPluginReadyCondition, condition.ReadyMessage)
	}
	if op != controllerutil.OperationResultNone {
		Log.Info(fmt.Sprintf("Dashboard Plugin definition %s successfully changed - operation: %s", uiPluginObj.GetName(), string(op)))
	}

	// Deploy PrometheusRule for dashboards
	err = r.ensureWatches(ctx, "prometheusrules.monitoring.rhobs", &monv1.PrometheusRule{}, eventHandler)
	if err != nil {
		instance.Status.Conditions.MarkFalse(telemetryv1.DashboardPrometheusRuleReadyCondition,
			condition.Reason("Can't own PrometheusRule resource. The Cluster Observability Operator probably isn't installed"),
			condition.SeverityError,
			telemetryv1.DashboardPrometheusRuleUnableToOwnMessage, err)
		Log.Info("Can't own PrometheusRule resource. The Cluster Observability Operator probably isn't installed")
		return ctrl.Result{RequeueAfter: telemetryv1.PauseBetweenWatchAttempts}, nil
	}
	prometheusRule := &monv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}
	op, err = controllerutil.CreateOrPatch(ctx, r.Client, prometheusRule, func() error {
		desiredPrometheusRule := metricstorage.DashboardPrometheusRule(instance, serviceLabels)
		desiredPrometheusRule.Spec.DeepCopyInto(&prometheusRule.Spec)
		prometheusRule.ObjectMeta.Labels = desiredPrometheusRule.ObjectMeta.Labels
		err = controllerutil.SetControllerReference(instance, prometheusRule, r.Scheme)
		return err
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		Log.Info(fmt.Sprintf("Prometheus Rules %s successfully changed - operation: %s", prometheusRule.Name, string(op)))
	}
	instance.Status.Conditions.MarkTrue(telemetryv1.DashboardPrometheusRuleReadyCondition, condition.ReadyMessage)

	// Deploy Configmap for Console UI Datasource
	datasourceName := instance.Namespace + "-" + instance.Name + "-datasource"
	datasourceCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      datasourceName,
			Namespace: metricstorage.DashboardArtifactsNamespace,
		},
	}
	dataSourceSuccess := false
	op, err = controllerutil.CreateOrPatch(ctx, r.Client, datasourceCM, func() error {
		datasourceCM.ObjectMeta.Labels = map[string]string{
			"console.openshift.io/dashboard-datasource": "true",
		}
		datasourceCM.Data, err = metricstorage.DashboardDatasourceData(ctx, r.Client, instance, datasourceName, metricstorage.DashboardArtifactsNamespace)
		return err
	})
	if err != nil {
		Log.Error(err, "Failed to update Console UI Datasource ConfigMap %s - operation: %s", datasourceCM.Name, string(op))
		instance.Status.Conditions.MarkFalse(telemetryv1.DashboardDatasourceReadyCondition,
			condition.Reason("Can't create Console UI Datasource ConfigMap"),
			condition.SeverityError,
			telemetryv1.DashboardDatasourceFailedMessage, err)
	} else {
		dataSourceSuccess = true
		instance.Status.Conditions.MarkTrue(telemetryv1.DashboardDatasourceReadyCondition, condition.ReadyMessage)
	}
	if op != controllerutil.OperationResultNone {
		Log.Info(fmt.Sprintf("Console UI Datasource ConfigMap %s successfully changed - operation: %s", datasourceCM.Name, string(op)))
	}

	// Deploy ConfigMaps for dashboards
	// NOTE: Dashboards installed without the custom datasource will default to the openshift-monitoring prometheus causing unexpected results
	if dataSourceSuccess {
		dashboardCMs := map[string]*corev1.ConfigMap{
			"grafana-dashboard-openstack-cloud":    dashboards.OpenstackCloud(datasourceName),
			"grafana-dashboard-openstack-node":     dashboards.OpenstackNode(datasourceName),
			"grafana-dashboard-openstack-vm":       dashboards.OpenstackVM(datasourceName),
			"grafana-dashboard-openstack-rabbitmq": dashboards.OpenstackRabbitmq(datasourceName),
		}

		// atleast one nodeset must have "telemetry-power-monitoring" service enabled for kepler dashboard to be created
		connectionInfo, err := getComputeNodesConnectionInfo(instance, helper, telemetryv1.TelemetryPowerMonitoring)
		if err != nil {
			Log.Info(fmt.Sprintf("Cannot get compute node connection info. Power monitoring dashboard not created. Error: %s", err))
		} else if len(connectionInfo) > 0 {
			dashboardCMs["grafana-dashboard-openstack-kepler"] = dashboards.OpenstackKepler(datasourceName)
		}

		for dashboardName, desiredCM := range dashboardCMs {
			dashboardCM := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      dashboardName,
					Namespace: metricstorage.DashboardArtifactsNamespace,
				},
			}
			op, err = controllerutil.CreateOrPatch(ctx, r.Client, dashboardCM, func() error {
				dashboardCM.ObjectMeta.Labels = desiredCM.ObjectMeta.Labels
				dashboardCM.Data = desiredCM.Data
				return nil
			})
			if err != nil {
				Log.Error(err, "Failed to update Dashboard ConfigMap %s - operation: %s", dashboardCM.Name, string(op))
				instance.Status.Conditions.MarkFalse(telemetryv1.DashboardDefinitionReadyCondition,
					condition.Reason("Can't create Console UI Dashboard ConfigMap"),
					condition.SeverityError,
					telemetryv1.DashboardDefinitionFailedMessage, err)
			} else {
				instance.Status.Conditions.MarkTrue(telemetryv1.DashboardDefinitionReadyCondition, condition.ReadyMessage)
			}
			if op != controllerutil.OperationResultNone {
				Log.Info(fmt.Sprintf("Dashboard ConfigMap %s successfully changed - operation: %s", dashboardCM.Name, string(op)))
			}
		}
	}
	return ctrl.Result{}, err
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

func getComputeNodesConnectionInfo(
	instance *telemetryv1.MetricStorage,
	helper *helper.Helper,
	telemetryServiceName string,
) ([]ConnectionInfo, error) {
	ipSetList, err := getIPSetList(instance, helper)
	if err != nil {
		return []ConnectionInfo{}, err
	}
	inventorySecretList, err := getInventorySecretList(instance, helper)
	if err != nil {
		return []ConnectionInfo{}, err
	}
	var address string
	connectionInfo := []ConnectionInfo{}
	for _, secret := range inventorySecretList.Items {
		inventory, err := ansible.UnmarshalYAML(secret.Data["inventory"])
		if err != nil {
			return []ConnectionInfo{}, err
		}
		nodeSetGroup := inventory.Groups[secret.Labels["openstackdataplanenodeset"]]
		containsTargetService := false
		for _, svc := range nodeSetGroup.Vars["edpm_services"].([]interface{}) {
			if svc.(string) == telemetryServiceName {
				containsTargetService = true
			}
		}
		if !containsTargetService {
			// If Telemetry|TelemetryPowerMonitoring isn't
			// deployed on this nodeset there is no reason
			// to include these nodes for scraping by prometheus
			continue
		}
		for name, item := range nodeSetGroup.Hosts {
			namespacedName := &types.NamespacedName{
				Name:      name,
				Namespace: instance.GetNamespace(),
			}
			if len(ipSetList.Items) > 0 {
				// if we have IPSets, lets go to search for the IPs there
				address, _ = getAddressFromIPSet(instance, &item, namespacedName, helper)
			} else if _, ok := item.Vars["ansible_host"]; ok {
				address, _ = getAddressFromAnsibleHost(&item)
			} else {
				// we were unable to find an IP or HostName for a node, so we do not go further
				return connectionInfo, fmt.Errorf("failed to find an IP or HostName for node %s", name)
			}
			if address == "" {
				// we were unable to find an IP or HostName for a node, so we do not go further
				return connectionInfo, fmt.Errorf("failed to find an IP or HostName for node %s", name)
			}

			fqdn, _ := getCanonicalHostname(&item)

			if TLSEnabled, ok := nodeSetGroup.Vars["edpm_tls_certs_enabled"].(bool); ok && TLSEnabled {
				connectionInfo = append(connectionInfo, ConnectionInfo{
					IP:       address,
					Hostname: name,
					TLS:      true,
					FQDN:     fqdn,
				})
			} else {
				connectionInfo = append(connectionInfo, ConnectionInfo{
					IP:       address,
					Hostname: name,
					TLS:      false,
					FQDN:     fqdn,
				})
			}
		}
	}
	return connectionInfo, nil
}

func getIPSetList(instance *telemetryv1.MetricStorage, helper *helper.Helper) (*infranetworkv1.IPSetList, error) {
	ipSets := &infranetworkv1.IPSetList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.GetNamespace()),
	}
	err := helper.GetClient().List(context.Background(), ipSets, listOpts...)
	return ipSets, err
}

func getInventorySecretList(instance *telemetryv1.MetricStorage, helper *helper.Helper) (*corev1.SecretList, error) {
	secrets := &corev1.SecretList{}
	labelSelector := map[string]string{
		"openstack.org/operator-name": "dataplane",
		"inventory":                   "true",
	}
	listOpts := []client.ListOption{
		client.InNamespace(instance.GetNamespace()),
		client.MatchingLabels(labelSelector),
	}
	err := helper.GetClient().List(context.Background(), secrets, listOpts...)
	return secrets, err
}

func getAddressFromIPSet(
	instance *telemetryv1.MetricStorage,
	item *ansible.Host,
	namespacedName *types.NamespacedName,
	helper *helper.Helper,
) (string, discoveryv1.AddressType) {
	ansibleHost := item.Vars["ansible_host"].(string)
	// we go search for an IPSet
	ipset := &infranetworkv1.IPSet{}
	err := helper.GetClient().Get(context.Background(), *namespacedName, ipset)
	if err != nil {
		// No IPsets found, lets try to get the HostName as last resource
		if isValidDomain(ansibleHost) {
			return ansibleHost, discoveryv1.AddressTypeFQDN
		}
		// No IP address or valid hostname found anywhere
		helper.GetLogger().Info("Did not found a valid hostname or IP address")
		return "", ""
	}
	// check that the reservations list is not empty
	if len(ipset.Status.Reservation) > 0 {
		// search for the network specified in the Spec
		for _, reservation := range ipset.Status.Reservation {
			if reservation.Network == *instance.Spec.DataplaneNetwork {
				return reservation.Address, discoveryv1.AddressTypeIPv4
			}
		}
	}
	// if the reservations list is empty, we go find if AnsibleHost exists
	return getAddressFromAnsibleHost(item)
}

func getAddressFromAnsibleHost(item *ansible.Host) (string, discoveryv1.AddressType) {
	ansibleHost := item.Vars["ansible_host"].(string)
	// check if ansiblehost is an IP
	addr := net.ParseIP(ansibleHost)
	if addr != nil {
		// it is an ip
		return ansibleHost, discoveryv1.AddressTypeIPv4
	}
	// it is not an ip, is it a valid hostname?
	if isValidDomain(ansibleHost) {
		// it is an valid domain name
		return ansibleHost, discoveryv1.AddressTypeFQDN
	}
	return "", ""
}

func getCanonicalHostname(item *ansible.Host) (string, discoveryv1.AddressType) {
	canonicalHostname := item.Vars["canonical_hostname"].(string)
	// is it a valid hostname?
	if isValidDomain(canonicalHostname) {
		// it is an valid domain name
		return canonicalHostname, discoveryv1.AddressTypeFQDN
	}
	return "", ""
}

// isValidDomain returns true if the domain is valid.
func isValidDomain(domain string) bool {
	domainRegexp := regexp.MustCompile(`^(?i)[a-z0-9-]+(\.[a-z0-9-]+)+\.?$`)
	return domainRegexp.MatchString(domain)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MetricStorageReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	Log := r.GetLogger(ctx)
	prometheusServiceWatchFn := func(_ context.Context, o client.Object) []reconcile.Request {
		result := []reconcile.Request{}

		// get all metricstorage CRs
		metricStorages := &telemetryv1.MetricStorageList{}
		listOpts := []client.ListOption{
			client.InNamespace(o.GetNamespace()),
		}
		if err := r.Client.List(context.Background(), metricStorages, listOpts...); err != nil {
			Log.Error(err, "Unable to retrieve MetricStorage CRs %w")
			return nil
		}

		for _, cr := range metricStorages.Items {
			if o.GetName() == fmt.Sprintf("%s-prometheus", cr.Name) {
				name := client.ObjectKey{
					Namespace: o.GetNamespace(),
					Name:      cr.Name,
				}
				Log.Info(fmt.Sprintf("Prometheus service %s is used by MetricStorage CR %s", o.GetName(), cr.Name))
				result = append(result, reconcile.Request{NamespacedName: name})
			}
		}
		if len(result) > 0 {
			return result
		}
		return nil
	}

	rabbitmqWatchFn := func(_ context.Context, o client.Object) []reconcile.Request {
		result := []reconcile.Request{}

		// get all metricstorage CRs
		metricStorages := &telemetryv1.MetricStorageList{}
		listOpts := []client.ListOption{
			client.InNamespace(o.GetNamespace()),
		}
		if err := r.Client.List(context.Background(), metricStorages, listOpts...); err != nil {
			Log.Error(err, "Unable to retrieve MetricStorage CRs %w")
			return nil
		}

		for _, cr := range metricStorages.Items {
			// Reconcile all metricstorages
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

	// index prometheusCaBundleSecretNameField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.MetricStorage{}, prometheusCaBundleSecretNameField, func(rawObj client.Object) []string {
		// Extract the secret name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.MetricStorage)
		if cr.Spec.PrometheusTLS.CaBundleSecretName == "" {
			return nil
		}
		return []string{cr.Spec.PrometheusTLS.CaBundleSecretName}
	}); err != nil {
		return err
	}

	// index prometheusTlsField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.MetricStorage{}, prometheusTLSField, func(rawObj client.Object) []string {
		// Extract the secret name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.MetricStorage)
		if cr.Spec.PrometheusTLS.SecretName == nil {
			return nil
		}
		return []string{*cr.Spec.PrometheusTLS.SecretName}
	}); err != nil {
		return err
	}
	inventoryPredicator, err := predicate.LabelSelectorPredicate(
		metav1.LabelSelector{
			MatchLabels: map[string]string{
				"openstack.org/operator-name": "dataplane",
				"inventory":                   "true",
			},
		},
	)
	if err != nil {
		return err
	}
	control, err := ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1.MetricStorage{}).
		Watches(&corev1.Service{},
			handler.EnqueueRequestsFromMapFunc(prometheusServiceWatchFn)).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForSrc),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.nodeSetWatchFn),
			builder.WithPredicates(inventoryPredicator),
		).
		Watches(
			&rabbitmqv1.RabbitmqCluster{},
			handler.EnqueueRequestsFromMapFunc(rabbitmqWatchFn),
		).
		Build(r)
	r.Controller = control
	return err
}

func (r *MetricStorageReconciler) findObjectsForSrc(ctx context.Context, src client.Object) []reconcile.Request {
	requests := []reconcile.Request{}

	l := log.FromContext(context.Background()).WithName("Controllers").WithName("MetricStorage")

	for _, field := range prometheusAllWatchFields {
		crList := &telemetryv1.MetricStorageList{}
		listOps := &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(field, src.GetName()),
			Namespace:     src.GetNamespace(),
		}
		err := r.Client.List(ctx, crList, listOps)
		if err != nil {
			return []reconcile.Request{}
		}

		for _, item := range crList.Items {
			l.Info(fmt.Sprintf("input source %s changed, reconcile: %s - %s", src.GetName(), item.GetName(), item.GetNamespace()))

			requests = append(requests,
				reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      item.GetName(),
						Namespace: item.GetNamespace(),
					},
				},
			)
		}
	}

	return requests
}

func (r *MetricStorageReconciler) nodeSetWatchFn(ctx context.Context, o client.Object) []reconcile.Request {
	l := log.FromContext(context.Background()).WithName("Controllers").WithName("MetricStorage")
	// Reconcile all metricstorages when a nodeset changes
	result := []reconcile.Request{}

	// get all MetricStorage CRs
	metricstorages := &telemetryv1.MetricStorageList{}
	listOpts := []client.ListOption{
		client.InNamespace(o.GetNamespace()),
	}
	if err := r.Client.List(ctx, metricstorages, listOpts...); err != nil {
		l.Error(err, "Unable to retrieve MetricStorage CRs %v")
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
