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
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	"github.com/openstack-k8s-operators/telemetry-operator/pkg/cloudkitty"
	"github.com/openstack-k8s-operators/telemetry-operator/pkg/cloudkittyapi"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	topologyv1 "github.com/openstack-k8s-operators/infra-operator/apis/topology/v1beta1"
	keystonev1 "github.com/openstack-k8s-operators/keystone-operator/api/v1beta1"
	"github.com/openstack-k8s-operators/lib-common/modules/common"
	"github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	"github.com/openstack-k8s-operators/lib-common/modules/common/endpoint"
	"github.com/openstack-k8s-operators/lib-common/modules/common/env"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/labels"
	nad "github.com/openstack-k8s-operators/lib-common/modules/common/networkattachment"
	"github.com/openstack-k8s-operators/lib-common/modules/common/secret"
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"github.com/openstack-k8s-operators/lib-common/modules/common/statefulset"
	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
)

// GetClient -
func (r *CloudKittyAPIReconciler) GetClient() client.Client {
	return r.Client
}

// GetKClient -
func (r *CloudKittyAPIReconciler) GetKClient() kubernetes.Interface {
	return r.Kclient
}

// GetScheme -
func (r *CloudKittyAPIReconciler) GetScheme() *runtime.Scheme {
	return r.Scheme
}

// CloudKittyAPIReconciler reconciles a CloudKittyAPI object
type CloudKittyAPIReconciler struct {
	client.Client
	Kclient kubernetes.Interface
	Scheme  *runtime.Scheme
}

// GetLogger returns a logger object with a logging prefix of "controller.name" and additional controller context fields
func (r *CloudKittyAPIReconciler) GetLogger(ctx context.Context) logr.Logger {
	return log.FromContext(ctx).WithName("Controllers").WithName("CloudKittyAPI")
}

var keystoneServices = []map[string]string{
	{
		"type": cloudkitty.ServiceType,
		"name": cloudkitty.ServiceName,
		"desc": "CloudKitty V2 Service",
	},
}

//+kubebuilder:rbac:groups=telemetry.openstack.org,resources=cloudkittyapis,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=telemetry.openstack.org,resources=cloudkittyapis/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=telemetry.openstack.org,resources=cloudkittyapis/finalizers,verbs=update;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;create;update;patch;delete;watch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;create;update;patch;delete;watch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;create;update;patch;delete;watch
// +kubebuilder:rbac:groups=keystone.openstack.org,resources=keystoneservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=keystone.openstack.org,resources=keystoneendpoints,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=k8s.cni.cncf.io,resources=network-attachment-definitions,verbs=get;list;watch
// +kubebuilder:rbac:groups=topology.openstack.org,resources=topologies,verbs=get;list;watch;update

// Reconcile -
func (r *CloudKittyAPIReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, _err error) {
	Log := r.GetLogger(ctx)

	// Fetch the CloudKittyAPI instance
	instance := &telemetryv1.CloudKittyAPI{}
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

	//
	// initialize status
	//
	isNewInstance := instance.Status.Conditions == nil
	if isNewInstance {
		instance.Status.Conditions = condition.Conditions{}
	}

	// Save a copy of the condtions so that we can restore the LastTransitionTime
	// when a condition's state doesn't change.
	savedConditions := instance.Status.Conditions.DeepCopy()

	// Always patch the instance status when exiting this function so we can persist any changes.
	defer func() {
		// Don't update the status, if reconciler Panics
		if r := recover(); r != nil {
			Log.Info(fmt.Sprintf("panic during reconcile %v\n", r))
			panic(r)
		}
		condition.RestoreLastTransitionTimes(&instance.Status.Conditions, savedConditions)
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

	// Always initialize conditions used later as Status=Unknown
	cl := condition.CreateList(
		condition.UnknownCondition(condition.ReadyCondition, condition.InitReason, condition.ReadyInitMessage),
		condition.UnknownCondition(condition.CreateServiceReadyCondition, condition.InitReason, condition.CreateServiceReadyInitMessage),
		condition.UnknownCondition(condition.InputReadyCondition, condition.InitReason, condition.InputReadyInitMessage),
		condition.UnknownCondition(condition.ServiceConfigReadyCondition, condition.InitReason, condition.ServiceConfigReadyInitMessage),
		condition.UnknownCondition(condition.DeploymentReadyCondition, condition.InitReason, condition.DeploymentReadyInitMessage),
		// right now we have no dedicated KeystoneServiceReadyInitMessage and KeystoneEndpointReadyInitMessage
		condition.UnknownCondition(condition.KeystoneServiceReadyCondition, condition.InitReason, ""),
		condition.UnknownCondition(condition.KeystoneEndpointReadyCondition, condition.InitReason, ""),
		condition.UnknownCondition(condition.NetworkAttachmentsReadyCondition, condition.InitReason, condition.NetworkAttachmentsReadyInitMessage),
		condition.UnknownCondition(condition.TLSInputReadyCondition, condition.InitReason, condition.InputReadyInitMessage),
	)
	instance.Status.Conditions.Init(&cl)
	// Always mark the Generation as observed early on
	instance.Status.ObservedGeneration = instance.Generation

	// If we're not deleting this and the service object doesn't have our finalizer, add it.
	if (instance.DeletionTimestamp.IsZero() && controllerutil.AddFinalizer(instance, helper.GetFinalizer())) || isNewInstance {
		// Register overall status immediately to have an early feedback e.g. in the cli
		return ctrl.Result{}, nil
	}

	if instance.Status.Hash == nil {
		instance.Status.Hash = map[string]string{}
	}
	if instance.Status.APIEndpoints == nil {
		instance.Status.APIEndpoints = map[string]map[string]string{}
	}
	if instance.Status.ServiceIDs == nil {
		instance.Status.ServiceIDs = map[string]string{}
	}
	if instance.Status.NetworkAttachments == nil {
		instance.Status.NetworkAttachments = map[string][]string{}
	}

	// Handle service delete
	if !instance.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, instance, helper)
	}

	// Init Topology condition if there's a reference
	if instance.Spec.TopologyRef != nil {
		c := condition.UnknownCondition(condition.TopologyReadyCondition, condition.InitReason, condition.TopologyReadyInitMessage)
		cl.Set(c)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, instance, helper)
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudKittyAPIReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	Log := r.GetLogger(ctx)

	// Watch for changes to secrets we don't own. Global secrets
	// (e.g. TransportURLSecret) are handled by the main cloudkitty controller.
	secretFn := func(_ context.Context, o client.Object) []reconcile.Request {
		var namespace string = o.GetNamespace()
		var secretName string = o.GetName()
		result := []reconcile.Request{}

		// get all API CRs
		apis := &telemetryv1.CloudKittyAPIList{}
		listOpts := []client.ListOption{
			client.InNamespace(namespace),
		}
		if err := r.Client.List(context.Background(), apis, listOpts...); err != nil {
			Log.Error(err, "Unable to retrieve API CRs %v")
			return nil
		}

		// Watch for changes to secrets where the owner label AND the
		// CR.Spec.ManagingCrName label matches
		label := o.GetLabels()
		if l, ok := label[labels.GetOwnerNameLabelSelector(labels.GetGroupLabel(cloudkitty.ServiceName))]; ok {
			for _, cr := range apis.Items {
				// return reconcile event for the CR where the owner label AND the parentCloudKittyName matches
				if l == cloudkitty.GetOwningCloudKittyName(&cr) {
					// return namespace and Name of CR
					name := client.ObjectKey{
						Namespace: o.GetNamespace(),
						Name:      cr.Name,
					}
					Log.Info(fmt.Sprintf("Secret %s and CR %s marked with label: %s", o.GetName(), cr.Name, l))

					result = append(result, reconcile.Request{NamespacedName: name})
				}
			}
		}

		// Watch for changes to any CustomServiceConfigSecrets
		for _, cr := range apis.Items {
			for _, v := range cr.Spec.CustomServiceConfigSecrets {
				if v == secretName {
					name := client.ObjectKey{
						Namespace: namespace,
						Name:      cr.Name,
					}
					Log.Info(fmt.Sprintf("Secret %s is used by CloudKitty CR %s", secretName, cr.Name))
					result = append(result, reconcile.Request{NamespacedName: name})
				}
			}
		}
		if len(result) > 0 {
			return result
		}
		return nil
	}

	// index passwordSecretField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.CloudKittyAPI{}, cloudKittyPasswordSecretField, func(rawObj client.Object) []string {
		// Extract the secret name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.CloudKittyAPI)
		if cr.Spec.Secret == "" {
			return nil
		}
		return []string{cr.Spec.Secret}
	}); err != nil {
		return err
	}

	// index caBundleSecretNameField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.CloudKittyAPI{}, cloudKittyCaBundleSecretNameField, func(rawObj client.Object) []string {
		// Extract the secret name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.CloudKittyAPI)
		if cr.Spec.TLS.CaBundleSecretName == "" {
			return nil
		}
		return []string{cr.Spec.TLS.CaBundleSecretName}
	}); err != nil {
		return err
	}

	// index tlsAPIInternalField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.CloudKittyAPI{}, cloudKittyTlsAPIInternalField, func(rawObj client.Object) []string {
		// Extract the secret name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.CloudKittyAPI)
		if cr.Spec.TLS.API.Internal.SecretName == nil {
			return nil
		}
		return []string{*cr.Spec.TLS.API.Internal.SecretName}
	}); err != nil {
		return err
	}

	// index tlsAPIPublicField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.CloudKittyAPI{}, cloudKittyTlsAPIPublicField, func(rawObj client.Object) []string {
		// Extract the secret name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.CloudKittyAPI)
		if cr.Spec.TLS.API.Public.SecretName == nil {
			return nil
		}
		return []string{*cr.Spec.TLS.API.Public.SecretName}
	}); err != nil {
		return err
	}

	// index topologyField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.CloudKittyAPI{}, cloudKittyTopologyField, func(rawObj client.Object) []string {
		// Extract the topology name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.CloudKittyAPI)
		if cr.Spec.TopologyRef == nil {
			return nil
		}
		return []string{cr.Spec.TopologyRef.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1.CloudKittyAPI{}).
		Owns(&keystonev1.KeystoneService{}).
		Owns(&keystonev1.KeystoneEndpoint{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		// watch the secrets we don't own
		Watches(&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(secretFn)).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForSrc),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Watches(&topologyv1.Topology{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForSrc),
			builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}

func (r *CloudKittyAPIReconciler) findObjectsForSrc(ctx context.Context, src client.Object) []reconcile.Request {
	requests := []reconcile.Request{}

	l := log.FromContext(ctx).WithName("Controllers").WithName("CloudKittyAPI")

	for _, field := range cloudKittyAPIWatchFields {
		crList := &telemetryv1.CloudKittyAPIList{}
		listOps := &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(field, src.GetName()),
			Namespace:     src.GetNamespace(),
		}
		err := r.List(ctx, crList, listOps)
		if err != nil {
			l.Error(err, fmt.Sprintf("listing %s for field: %s - %s", crList.GroupVersionKind().Kind, field, src.GetNamespace()))
			return requests
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

func (r *CloudKittyAPIReconciler) reconcileDelete(ctx context.Context, instance *telemetryv1.CloudKittyAPI, helper *helper.Helper) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)

	Log.Info(fmt.Sprintf("Reconciling Service '%s' delete", instance.Name))

	// It's possible to get here before the endpoints have been set in the status, so check for this
	if instance.Status.APIEndpoints != nil {
		for _, ksSvc := range keystoneServices {

			// Remove the finalizer from our KeystoneEndpoint CR
			keystoneEndpoint, err := keystonev1.GetKeystoneEndpointWithName(ctx, helper, ksSvc["name"], instance.Namespace)
			if err != nil && !k8s_errors.IsNotFound(err) {
				return ctrl.Result{}, err
			}

			if err == nil {
				controllerutil.RemoveFinalizer(keystoneEndpoint, helper.GetFinalizer())
				if err = helper.GetClient().Update(ctx, keystoneEndpoint); err != nil && !k8s_errors.IsNotFound(err) {
					return ctrl.Result{}, err
				}
				util.LogForObject(helper, "Removed finalizer from our KeystoneEndpoint", instance)
			}

			// Remove the finalizer from our KeystoneService CR
			keystoneService, err := keystonev1.GetKeystoneServiceWithName(ctx, helper, ksSvc["name"], instance.Namespace)
			if err != nil && !k8s_errors.IsNotFound(err) {
				return ctrl.Result{}, err
			}

			if err == nil {
				controllerutil.RemoveFinalizer(keystoneService, helper.GetFinalizer())
				if err = helper.GetClient().Update(ctx, keystoneService); err != nil && !k8s_errors.IsNotFound(err) {
					return ctrl.Result{}, err
				}
				util.LogForObject(helper, "Removed finalizer from our KeystoneService", instance)
			}
		}
	}

	// Remove finalizer on the Topology CR
	if ctrlResult, err := topologyv1.EnsureDeletedTopologyRef(
		ctx,
		helper,
		instance.Status.LastAppliedTopology,
		instance.Name,
	); err != nil {
		return ctrlResult, err
	}

	// Service is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(instance, helper.GetFinalizer())
	Log.Info(fmt.Sprintf("Reconciled Service '%s' delete successfully", instance.Name))

	return ctrl.Result{}, nil
}

func (r *CloudKittyAPIReconciler) reconcileInit(
	ctx context.Context,
	instance *telemetryv1.CloudKittyAPI,
	helper *helper.Helper,
	serviceLabels map[string]string,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)

	Log.Info(fmt.Sprintf("Reconciling Service '%s' init", instance.Name))

	//
	// expose the service (create service and return the created endpoint URLs)
	//

	// V2
	publicEndpointData := endpoint.Data{
		Port: cloudkitty.CloudKittyPublicPort,
		Path: "/v2",
	}
	internalEndpointData := endpoint.Data{
		Port: cloudkitty.CloudKittyInternalPort,
		Path: "/v2",
	}
	cloudkittyEndpoints := map[service.Endpoint]endpoint.Data{
		service.EndpointPublic:   publicEndpointData,
		service.EndpointInternal: internalEndpointData,
	}

	apiEndpointsV3 := make(map[string]string)

	for endpointType, data := range cloudkittyEndpoints {
		endpointTypeStr := string(endpointType)
		endpointName := cloudkitty.ServiceName + "-" + endpointTypeStr
		svcOverride := instance.Spec.Override.Service[endpointType]
		if svcOverride.EmbeddedLabelsAnnotations == nil {
			svcOverride.EmbeddedLabelsAnnotations = &service.EmbeddedLabelsAnnotations{}
		}

		exportLabels := util.MergeStringMaps(
			serviceLabels,
			map[string]string{
				service.AnnotationEndpointKey: endpointTypeStr,
			},
		)

		// Create the service
		svc, err := service.NewService(
			service.GenericService(&service.GenericServiceDetails{
				Name:      endpointName,
				Namespace: instance.Namespace,
				Labels:    exportLabels,
				Selector:  serviceLabels,
				Port: service.GenericServicePort{
					Name:     endpointName,
					Port:     data.Port,
					Protocol: corev1.ProtocolTCP,
				},
			}),
			5,
			&svcOverride.OverrideSpec,
		)
		if err != nil {
			instance.Status.Conditions.Set(condition.FalseCondition(
				condition.CreateServiceReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.CreateServiceReadyErrorMessage,
				err.Error()))

			return ctrl.Result{}, err
		}

		svc.AddAnnotation(map[string]string{
			service.AnnotationEndpointKey: endpointTypeStr,
		})

		// add Annotation to whether creating an ingress is required or not
		if endpointType == service.EndpointPublic && svc.GetServiceType() == corev1.ServiceTypeClusterIP {
			svc.AddAnnotation(map[string]string{
				service.AnnotationIngressCreateKey: "true",
			})
		} else {
			svc.AddAnnotation(map[string]string{
				service.AnnotationIngressCreateKey: "false",
			})
			if svc.GetServiceType() == corev1.ServiceTypeLoadBalancer {
				svc.AddAnnotation(map[string]string{
					service.AnnotationHostnameKey: svc.GetServiceHostname(), // add annotation to register service name in dnsmasq
				})
			}
		}

		ctrlResult, err := svc.CreateOrPatch(ctx, helper)
		if err != nil {
			instance.Status.Conditions.Set(condition.FalseCondition(
				condition.CreateServiceReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.CreateServiceReadyErrorMessage,
				err.Error()))

			return ctrlResult, err
		} else if (ctrlResult != ctrl.Result{}) {
			instance.Status.Conditions.Set(condition.FalseCondition(
				condition.CreateServiceReadyCondition,
				condition.RequestedReason,
				condition.SeverityInfo,
				condition.CreateServiceReadyRunningMessage))
			return ctrlResult, nil
		}
		// create service - end

		// if TLS is enabled
		if instance.Spec.TLS.API.Enabled(endpointType) {
			// set endpoint protocol to https
			data.Protocol = ptr.To(service.ProtocolHTTPS)
		}

		apiEndpointsV3[string(endpointType)], err = svc.GetAPIEndpoint(
			svcOverride.EndpointURL, data.Protocol, data.Path)
		if err != nil {
			instance.Status.Conditions.MarkFalse(
				condition.CreateServiceReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.CreateServiceReadyErrorMessage,
				err.Error())
			return ctrl.Result{}, err
		}
	}
	instance.Status.Conditions.MarkTrue(condition.CreateServiceReadyCondition, condition.CreateServiceReadyMessage)

	//
	// Update instance status with service endpoint url from route host information
	//
	if instance.Status.APIEndpoints == nil {
		instance.Status.APIEndpoints = map[string]map[string]string{}
	}
	instance.Status.APIEndpoints[cloudkitty.ServiceName] = apiEndpointsV3
	// V2 - end

	// expose service - end

	//
	// create service and user in keystone - - https://docs.openstack.org/CloudKitty/latest/install/install-rdo.html#configure-user-and-endpoints
	// TODO: rework this
	//
	if instance.Status.ServiceIDs == nil {
		instance.Status.ServiceIDs = map[string]string{}
	}

	for _, ksSvc := range keystoneServices {
		ksSvcSpec := keystonev1.KeystoneServiceSpec{
			ServiceType:        ksSvc["type"],
			ServiceName:        ksSvc["name"],
			ServiceDescription: ksSvc["desc"],
			Enabled:            true,
			ServiceUser:        instance.Spec.ServiceUser,
			Secret:             instance.Spec.Secret,
			PasswordSelector:   instance.Spec.PasswordSelectors.CloudKittyService,
		}

		ksSvcObj := keystonev1.NewKeystoneService(ksSvcSpec, instance.Namespace, serviceLabels, cloudkitty.NormalDuration)
		ctrlResult, err := ksSvcObj.CreateOrPatch(ctx, helper)
		if err != nil {
			instance.Status.Conditions.MarkFalse(
				condition.KeystoneServiceReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				"Creating KeyStoneService CR %s",
				err.Error())
			return ctrlResult, err
		}

		// mirror the Status, Reason, Severity and Message of the latest keystoneservice condition
		// into a local condition with the type condition.KeystoneServiceReadyCondition
		c := ksSvcObj.GetConditions().Mirror(condition.KeystoneServiceReadyCondition)
		if c != nil {
			instance.Status.Conditions.Set(c)
		}

		if (ctrlResult != ctrl.Result{}) {
			return ctrlResult, nil
		}

		instance.Status.ServiceIDs[ksSvc["name"]] = ksSvcObj.GetServiceID()

		ksEndptSpec := keystonev1.KeystoneEndpointSpec{
			ServiceName: ksSvc["name"],
			Endpoints:   instance.Status.APIEndpoints[ksSvc["name"]],
		}

		ksEndptObj := keystonev1.NewKeystoneEndpoint(
			ksSvc["name"],
			instance.Namespace,
			ksEndptSpec,
			serviceLabels,
			cloudkitty.NormalDuration)
		ctrlResult, err = ksEndptObj.CreateOrPatch(ctx, helper)
		if err != nil {
			instance.Status.Conditions.MarkFalse(
				condition.KeystoneEndpointReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				"Creating KeyStoneEndpoint CR %s",
				err.Error())
			return ctrlResult, err
		}

		// mirror the Status, Reason, Severity and Message of the latest keystoneendpoint condition
		// into a local condition with the type condition.KeystoneEndpointReadyCondition
		c = ksEndptObj.GetConditions().Mirror(condition.KeystoneEndpointReadyCondition)
		if c != nil {
			instance.Status.Conditions.Set(c)
		}

		if (ctrlResult != ctrl.Result{}) {
			return ctrlResult, nil
		}
	}

	Log.Info(fmt.Sprintf("Reconciled Service '%s' init successfully", instance.Name))
	return ctrl.Result{}, nil
}

func (r *CloudKittyAPIReconciler) reconcileNormal(ctx context.Context, instance *telemetryv1.CloudKittyAPI, helper *helper.Helper) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)

	Log.Info(fmt.Sprintf("Reconciling Service '%s'", instance.Name))

	configVars := make(map[string]env.Setter)

	//
	// check for required OpenStack secret holding passwords for service/admin user and add hash to the vars map
	//

	ctrlResult, err := cloudkitty.VerifyServiceSecret(
		ctx,
		types.NamespacedName{Namespace: instance.Namespace, Name: instance.Spec.Secret},
		[]string{
			instance.Spec.PasswordSelectors.CloudKittyService,
		},
		helper.GetClient(),
		&instance.Status.Conditions,
		cloudkitty.NormalDuration,
		&configVars,
	)
	if err != nil {
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	//
	// check for required Transport URL and config secrets
	//

	parentCloudKittyName := cloudkitty.GetOwningCloudKittyName(instance)
	secretNames := []string{
		instance.Spec.TransportURLSecret,                    // TransportURLSecret
		fmt.Sprintf("%s-scripts", parentCloudKittyName),     // ScriptsSecret
		fmt.Sprintf("%s-config-data", parentCloudKittyName), // ConfigSecret
	}
	// Append CustomServiceConfigSecrets that should be checked
	secretNames = append(secretNames, instance.Spec.CustomServiceConfigSecrets...)

	ctrlResult, err = cloudkitty.VerifyConfigSecrets(
		ctx,
		helper,
		&instance.Status.Conditions,
		secretNames,
		instance.Namespace,
		&configVars,
	)
	if err != nil {
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	instance.Status.Conditions.MarkTrue(condition.InputReadyCondition, condition.InputReadyMessage)

	//
	// TLS input validation
	//
	// Validate the CA cert secret if provided
	if instance.Spec.TLS.CaBundleSecretName != "" {
		hash, err := tls.ValidateCACertSecret(
			ctx,
			helper.GetClient(),
			types.NamespacedName{
				Name:      instance.Spec.TLS.CaBundleSecretName,
				Namespace: instance.Namespace,
			},
		)
		if err != nil {
			if k8s_errors.IsNotFound(err) {
				instance.Status.Conditions.Set(condition.FalseCondition(
					condition.TLSInputReadyCondition,
					condition.RequestedReason,
					condition.SeverityInfo,
					fmt.Sprintf(condition.TLSInputReadyWaitingMessage, instance.Spec.TLS.CaBundleSecretName)))
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

		if hash != "" {
			configVars[tls.CABundleKey] = env.SetValue(hash)
		}
	}

	// Validate API service certs secrets
	certsHash, err := instance.Spec.TLS.API.ValidateCertSecrets(ctx, helper, instance.Namespace)
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
	configVars[tls.TLSHashName] = env.SetValue(certsHash)

	// all cert input checks out so report InputReady
	instance.Status.Conditions.MarkTrue(condition.TLSInputReadyCondition, condition.InputReadyMessage)

	//
	// Create secrets required as input for the Service and calculate an overall hash of hashes
	//
	serviceLabels := map[string]string{
		common.AppSelector:       cloudkitty.ServiceName,
		common.ComponentSelector: cloudkittyapi.ComponentName,
	}

	//
	// create custom config for this cloudkitty service
	//
	err = r.generateServiceConfigs(ctx, helper, instance, &configVars, serviceLabels)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.ServiceConfigReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.ServiceConfigReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}

	instance.Status.Conditions.MarkTrue(condition.ServiceConfigReadyCondition, condition.ServiceConfigReadyMessage)

	//
	// TODO check when/if Init, Update, or Upgrade should/could be skipped
	//

	// networks to attach to
	nadList := []networkv1.NetworkAttachmentDefinition{}
	for _, netAtt := range instance.Spec.NetworkAttachments {
		nad, err := nad.GetNADWithName(ctx, helper, netAtt, instance.Namespace)
		if err != nil {
			if k8s_errors.IsNotFound(err) {
				Log.Info(fmt.Sprintf("network-attachment-definition %s not found", netAtt))
				instance.Status.Conditions.Set(condition.FalseCondition(
					condition.NetworkAttachmentsReadyCondition,
					condition.RequestedReason,
					condition.SeverityInfo,
					condition.NetworkAttachmentsReadyWaitingMessage,
					netAtt))
				return cloudkitty.ResultRequeue, fmt.Errorf("network-attachment-definition %s not found", netAtt)
			}
			instance.Status.Conditions.Set(condition.FalseCondition(
				condition.NetworkAttachmentsReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.NetworkAttachmentsReadyErrorMessage,
				err.Error()))
			return ctrl.Result{}, err
		}

		if nad != nil {
			nadList = append(nadList, *nad)
		}
	}

	serviceAnnotations, err := nad.EnsureNetworksAnnotation(nadList)
	if err != nil {
		err = fmt.Errorf("failed create network annotation from %s: %w", instance.Spec.NetworkAttachments, err)
		instance.Status.Conditions.MarkFalse(
			condition.NetworkAttachmentsReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.NetworkAttachmentsReadyErrorMessage,
			err.Error())
		return ctrl.Result{}, err
	}

	// Handle service init
	ctrlResult, err = r.reconcileInit(ctx, instance, helper, serviceLabels)
	if err != nil {
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	//
	// Handle Topology
	//
	topology, err := ensureTopology(
		ctx,
		helper,
		instance,      // topologyHandler
		instance.Name, // finalizer
		&instance.Status.Conditions,
		labels.GetLabelSelector(serviceLabels),
	)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.TopologyReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.TopologyReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, fmt.Errorf("waiting for Topology requirements: %w", err)
	}

	//
	// normal reconcile tasks
	//

	//
	// create hash over all the different input resources to identify if any those changed
	// and a restart/recreate is required.
	//
	inputHash, hashChanged, err := r.createHashOfInputHashes(ctx, instance, configVars)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.ServiceConfigReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.ServiceConfigReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	} else if hashChanged {
		Log.Info(fmt.Sprintf("%s... requeueing", condition.ServiceConfigReadyInitMessage))
		instance.Status.Conditions.MarkFalse(
			condition.ServiceConfigReadyCondition,
			condition.InitReason,
			condition.SeverityInfo,
			condition.ServiceConfigReadyInitMessage)
		// Hash changed and instance status should be updated (which will be done by main defer func),
		// so we need to return and reconcile again
		return ctrl.Result{}, nil
	}

	// Deploy a statefulset
	ssDef, err := cloudkittyapi.StatefulSet(instance, inputHash, serviceLabels, serviceAnnotations, topology)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DeploymentReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.DeploymentReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}
	ss := statefulset.NewStatefulSet(ssDef, cloudkitty.ShortDuration)

	var ssData appsv1.StatefulSet
	ctrlResult, err = ss.CreateOrPatch(ctx, helper)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DeploymentReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.DeploymentReadyErrorMessage,
			err.Error()))
		return ctrlResult, err

	} else if (ctrlResult == ctrl.Result{}) {
		// Wait until the data in the StatefulSet is for the current generation
		ssData = ss.GetStatefulSet()
		if ssData.Generation != ssData.Status.ObservedGeneration {
			ctrlResult = cloudkitty.ResultRequeue
			err = fmt.Errorf("waiting for Statefulset %s to start reconciling", ssData.Name)
		}
	}

	if (ctrlResult != ctrl.Result{}) {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DeploymentReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.DeploymentReadyRunningMessage))
		// If the deployment is not ready, then neither are the NADs
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.NetworkAttachmentsReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.NetworkAttachmentsReadyInitMessage))
		return ctrlResult, err
	}

	instance.Status.ReadyCount = ssData.Status.ReadyReplicas

	// verify if network attachment matches expectations
	networkReady := false
	networkAttachmentStatus := map[string][]string{}
	if *instance.Spec.Replicas > 0 {
		networkReady, networkAttachmentStatus, err = nad.VerifyNetworkStatusFromAnnotation(
			ctx,
			helper,
			instance.Spec.NetworkAttachments,
			serviceLabels,
			instance.Status.ReadyCount,
		)
		if err != nil {
			err = fmt.Errorf("verifying API NetworkAttachments (%s) %w", instance.Spec.NetworkAttachments, err)
			instance.Status.Conditions.MarkFalse(
				condition.NetworkAttachmentsReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.NetworkAttachmentsReadyErrorMessage,
				err.Error())
			return ctrl.Result{}, err
		}
	} else {
		networkReady = true
	}

	instance.Status.NetworkAttachments = networkAttachmentStatus
	if networkReady {
		instance.Status.Conditions.MarkTrue(condition.NetworkAttachmentsReadyCondition, condition.NetworkAttachmentsReadyMessage)
	} else {
		err := fmt.Errorf("not all pods have interfaces with ips as configured in NetworkAttachments: %s", instance.Spec.NetworkAttachments)
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.NetworkAttachmentsReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.NetworkAttachmentsReadyErrorMessage,
			err.Error()))

		return ctrl.Result{}, err
	}

	if instance.Status.ReadyCount > 0 {
		instance.Status.Conditions.MarkTrue(condition.DeploymentReadyCondition, condition.DeploymentReadyMessage)

	} else if *instance.Spec.Replicas > 0 {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DeploymentReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.DeploymentReadyRunningMessage))

	} else {
		instance.Status.Conditions.MarkFalse(
			condition.DeploymentReadyCondition,
			condition.NotRequestedReason,
			condition.SeverityInfo,
			condition.DeploymentReadyInitMessage)
	}
	// create StatefulSet - end

	Log.Info(fmt.Sprintf("Reconciled Service '%s' successfully", instance.Name))
	// update the overall status condition if service is ready
	if instance.IsReady() {
		instance.Status.Conditions.MarkTrue(condition.ReadyCondition, condition.ReadyMessage)
	}
	// For non ready we'll let the main defer func handle the status update using the Mirror function
	return ctrl.Result{}, nil
}

// generateServiceConfigs - create Secret which holds the service configuration
func (r *CloudKittyAPIReconciler) generateServiceConfigs(
	ctx context.Context,
	h *helper.Helper,
	instance *telemetryv1.CloudKittyAPI,
	envVars *map[string]env.Setter,
	serviceLabels map[string]string,
) error {
	//
	// create custom Secret for cloudkitty service-specific config input
	// - %-config-data holds custom config for the service
	//

	labels := labels.GetLabels(instance, labels.GetGroupLabel(cloudkitty.ServiceName), serviceLabels)

	// customData hold any customization for the service.
	customData := map[string]string{cloudkitty.CustomServiceConfigFileName: instance.Spec.CustomServiceConfig}

	// Fetch the two service config snippets (DefaultsConfigFileName and
	// CustomConfigFileName) from the Secret generated by the top level
	// cloudkitty controller, and add them to this service specific Secret.
	cloudkittySecretName := cloudkitty.GetOwningCloudKittyName(instance) + "-config-data"
	cloudkittySecret, _, err := secret.GetSecret(ctx, h, cloudkittySecretName, instance.Namespace)
	if err != nil {
		return err
	}
	customData[cloudkitty.DefaultsConfigFileName] = string(cloudkittySecret.Data[cloudkitty.DefaultsConfigFileName])
	customData[cloudkitty.CustomConfigFileName] = string(cloudkittySecret.Data[cloudkitty.CustomConfigFileName])

	customSecrets := ""
	for _, secretName := range instance.Spec.CustomServiceConfigSecrets {
		secret, _, err := secret.GetSecret(ctx, h, secretName, instance.Namespace)
		if err != nil {
			return err
		}
		for _, data := range secret.Data {
			customSecrets += string(data) + "\n"
		}
	}
	customData[cloudkitty.CustomServiceConfigSecretsFileName] = customSecrets

	templateParameters := map[string]interface{}{
		"LogFile": cloudkittyapi.LogFile,
	}

	configTemplates := []util.Template{
		{
			Name:          fmt.Sprintf("%s-config-data", instance.Name),
			Namespace:     instance.Namespace,
			Type:          util.TemplateTypeConfig,
			InstanceType:  instance.Kind,
			CustomData:    customData,
			ConfigOptions: templateParameters,
			Labels:        labels,
		},
	}

	return secret.EnsureSecrets(ctx, h, instance, configTemplates, envVars)
}

// createHashOfInputHashes - creates a hash of hashes which gets added to the resources which requires a restart
// if any of the input resources change, like configs, passwords, ...
//
// returns the hash, whether the hash changed (as a bool) and any error
func (r *CloudKittyAPIReconciler) createHashOfInputHashes(
	ctx context.Context,
	instance *telemetryv1.CloudKittyAPI,
	envVars map[string]env.Setter,
) (string, bool, error) {
	Log := r.GetLogger(ctx)

	var hashMap map[string]string
	changed := false
	mergedMapVars := env.MergeEnvs([]corev1.EnvVar{}, envVars)
	hash, err := util.ObjectHash(mergedMapVars)
	if err != nil {
		return hash, changed, err
	}
	if hashMap, changed = util.SetHash(instance.Status.Hash, common.InputHashName, hash); changed {
		instance.Status.Hash = hashMap
		Log.Info(fmt.Sprintf("Input maps hash %s - %s", common.InputHashName, hash))
	}
	return hash, changed, nil
}
