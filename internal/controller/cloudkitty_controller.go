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

package controller

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	lokistackv1 "github.com/grafana/loki/operator/api/loki/v1"

	"github.com/openstack-k8s-operators/telemetry-operator/internal/cloudkitty"
	"github.com/openstack-k8s-operators/telemetry-operator/internal/metricstorage"
	"github.com/openstack-k8s-operators/telemetry-operator/internal/utils"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	certmgrv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/go-logr/logr"
	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	memcachedv1 "github.com/openstack-k8s-operators/infra-operator/apis/memcached/v1beta1"
	rabbitmqv1 "github.com/openstack-k8s-operators/infra-operator/apis/rabbitmq/v1beta1"
	keystonev1 "github.com/openstack-k8s-operators/keystone-operator/api/v1beta1"
	"github.com/openstack-k8s-operators/lib-common/modules/certmanager"
	"github.com/openstack-k8s-operators/lib-common/modules/common"
	"github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	"github.com/openstack-k8s-operators/lib-common/modules/common/configmap"
	"github.com/openstack-k8s-operators/lib-common/modules/common/endpoint"
	"github.com/openstack-k8s-operators/lib-common/modules/common/env"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/job"
	"github.com/openstack-k8s-operators/lib-common/modules/common/labels"
	nad "github.com/openstack-k8s-operators/lib-common/modules/common/networkattachment"
	common_rbac "github.com/openstack-k8s-operators/lib-common/modules/common/rbac"
	"github.com/openstack-k8s-operators/lib-common/modules/common/secret"
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	mariadbv1 "github.com/openstack-k8s-operators/mariadb-operator/api/v1beta1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetClient -
func (r *CloudKittyReconciler) GetClient() client.Client {
	return r.Client
}

// GetKClient -
func (r *CloudKittyReconciler) GetKClient() kubernetes.Interface {
	return r.Kclient
}

// GetScheme -
func (r *CloudKittyReconciler) GetScheme() *runtime.Scheme {
	return r.Scheme
}

// CloudKittyReconciler reconciles a CloudKitty object
type CloudKittyReconciler utils.ConditionalWatchingReconciler

// GetLogger returns a logger object with a logging prefix of "controller.name" and additional controller context fields
func (r *CloudKittyReconciler) GetLogger(ctx context.Context) logr.Logger {
	return log.FromContext(ctx).WithName("Controllers").WithName("CloudKitty")
}

// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=cloudkitties,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=cloudkitties/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=cloudkitties/finalizers,verbs=update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=cloudkittyapis,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=cloudkittyapis/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=cloudkittyapis/finalizers,verbs=update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=cloudkittyprocs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=cloudkittyprocs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=cloudkittyprocs/finalizers,verbs=update;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;create;update;patch;delete;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;create;update;patch;delete;watch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;create;update;patch;delete;watch
// +kubebuilder:rbac:groups=mariadb.openstack.org,resources=mariadbdatabases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mariadb.openstack.org,resources=mariadbaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mariadb.openstack.org,resources=mariadbaccounts/finalizers,verbs=update;patch
// +kubebuilder:rbac:groups=memcached.openstack.org,resources=memcacheds,verbs=get;list;watch;
// +kubebuilder:rbac:groups=keystone.openstack.org,resources=keystoneapis,verbs=get;list;watch
// +kubebuilder:rbac:groups=rabbitmq.openstack.org,resources=transporturls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=k8s.cni.cncf.io,resources=network-attachment-definitions,verbs=get;list;watch
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch
// +kubebuilder:rbac:groups=loki.grafana.com,resources=lokistacks,verbs=get;list;watch;create;update;patch;delete

// service account, role, rolebinding
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;list;watch;create;update;patch
// service account permissions that are needed to grant permission to the above
// +kubebuilder:rbac:groups="security.openshift.io",resourceNames=anyuid;privileged,resources=securitycontextconstraints,verbs=use
// +kubebuilder:rbac:groups="",resources=pods,verbs=create;delete;get;list;patch;update;watch

// Reconcile -
func (r *CloudKittyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, _err error) {
	Log := r.GetLogger(ctx)

	// Fetch the CloudKitty instance
	instance := &telemetryv1.CloudKitty{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected.
			// For additional cleanup logic use finalizers. Return and don't requeue.
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		Log.Error(err, fmt.Sprintf("could not fetch CloudKitty instance %s", instance.Name))
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
		Log.Error(err, fmt.Sprintf("could not instantiate helper for instance %s", instance.Name))
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
		condition.UnknownCondition(condition.DBReadyCondition, condition.InitReason, condition.DBReadyInitMessage),
		condition.UnknownCondition(condition.DBSyncReadyCondition, condition.InitReason, condition.DBSyncReadyInitMessage),
		condition.UnknownCondition(telemetryv1.CloudKittyStorageInitReadyCondition, condition.InitReason, telemetryv1.CloudKittyStorageInitReadyInitMessage),
		condition.UnknownCondition(condition.RabbitMqTransportURLReadyCondition, condition.InitReason, condition.RabbitMqTransportURLReadyInitMessage),
		condition.UnknownCondition(condition.MemcachedReadyCondition, condition.InitReason, condition.MemcachedReadyInitMessage),
		condition.UnknownCondition(condition.InputReadyCondition, condition.InitReason, condition.InputReadyInitMessage),
		condition.UnknownCondition(condition.ServiceConfigReadyCondition, condition.InitReason, condition.ServiceConfigReadyInitMessage),
		condition.UnknownCondition(telemetryv1.CloudKittyAPIReadyCondition, condition.InitReason, telemetryv1.CloudKittyAPIReadyInitMessage),
		condition.UnknownCondition(telemetryv1.CloudKittyProcReadyCondition, condition.InitReason, telemetryv1.CloudKittyProcReadyInitMessage),
		condition.UnknownCondition(telemetryv1.CloudKittyClientCertReadyCondition, condition.InitReason, telemetryv1.CloudKittyClientCertReadyInitMessage),
		condition.UnknownCondition(telemetryv1.CloudKittyLokiStackReadyCondition, condition.InitReason, telemetryv1.CloudKittyLokiStackReadyInitMessage),
		condition.UnknownCondition(condition.NetworkAttachmentsReadyCondition, condition.InitReason, condition.NetworkAttachmentsReadyInitMessage),
		// service account, role, rolebinding conditions
		condition.UnknownCondition(condition.ServiceAccountReadyCondition, condition.InitReason, condition.ServiceAccountReadyInitMessage),
		condition.UnknownCondition(condition.RoleReadyCondition, condition.InitReason, condition.RoleReadyInitMessage),
		condition.UnknownCondition(condition.RoleBindingReadyCondition, condition.InitReason, condition.RoleBindingReadyInitMessage),
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

	// Handle service delete
	if !instance.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, instance, helper)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, instance, helper)
}

// fields to index to reconcile when change
const (
	cloudKittyPasswordSecretField = ".spec.secret"
	//nolint:gosec // Not hardcoded credentials, just field name
	cloudKittyCaBundleSecretNameField  = ".spec.tls.caBundleSecretName"
	cloudKittyTLSAPIInternalField      = ".spec.tls.api.internal.secretName"
	cloudKittyTLSAPIPublicField        = ".spec.tls.api.public.secretName"
	cloudKittyAuthAppCredSecretField   = ".spec.auth.applicationCredentialSecret" //nolint:gosec // G101: Not actual credentials, just field path
	cloudKittyTopologyField            = ".spec.topologyRef.Name"
	cloudKittyCustomConfigsSecretField = ".spec.customConfigsSecretName" //nolint:gosec // G101: Not actual credentials, just field path
)

var (
	cloudKittyProcWatchFields = []string{
		cloudKittyPasswordSecretField,
		cloudKittyCaBundleSecretNameField,
		cloudKittyTopologyField,
		cloudKittyCustomConfigsSecretField,
	}
	cloudKittyAPIWatchFields = []string{
		cloudKittyPasswordSecretField,
		cloudKittyCaBundleSecretNameField,
		cloudKittyTLSAPIInternalField,
		cloudKittyTLSAPIPublicField,
		cloudKittyTopologyField,
		cloudKittyCustomConfigsSecretField,
	}
	cloudKittyWatchFields = []string{
		cloudKittyAuthAppCredSecretField,
	}
)

// SetupWithManager sets up the controller with the Manager.
func (r *CloudKittyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// index authAppCredSecretField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.CloudKitty{}, cloudKittyAuthAppCredSecretField, func(rawObj client.Object) []string {
		cr := rawObj.(*telemetryv1.CloudKitty)
		if cr.Spec.Auth.ApplicationCredentialSecret == "" {
			return nil
		}
		return []string{cr.Spec.Auth.ApplicationCredentialSecret}
	}); err != nil {
		return err
	}

	// transportURLSecretFn - Watch for changes made to the secret associated with the RabbitMQ
	// TransportURL created and used by CloudKitty CRs.  Watch functions return a list of namespace-scoped
	// CRs that then get fed  to the reconciler.  Hence, in this case, we need to know the name of the
	// CloudKitty CR associated with the secret we are examining in the function.  We could parse the name
	// out of the "%s-cloudkitty-transport" secret label, which would be faster than getting the list of
	// the CloudKitty CRs and trying to match on each one.  The downside there, however, is that technically
	// someone could randomly label a secret "something-cloudkitty-transport" where "something" actually
	// matches the name of an existing CloudKitty CR.  In that case changes to that secret would trigger
	// reconciliation for a CloudKitty CR that does not need it.
	//
	// TODO: We also need a watch func to monitor for changes to the secret referenced by CloudKitty.Spec.Secret
	transportURLSecretFn := func(ctx context.Context, o client.Object) []reconcile.Request {
		result := []reconcile.Request{}

		Log := r.GetLogger(ctx)

		// get all CloudKitty CRs
		cloudkitties := &telemetryv1.CloudKittyList{}
		listOpts := []client.ListOption{
			client.InNamespace(o.GetNamespace()),
		}
		if err := r.List(ctx, cloudkitties, listOpts...); err != nil {
			Log.Error(err, "Unable to retrieve CloudKitty CRs %v")
			return nil
		}

		for _, ownerRef := range o.GetOwnerReferences() {
			if ownerRef.Kind == "TransportURL" {
				for _, cr := range cloudkitties.Items {
					if ownerRef.Name == fmt.Sprintf("%s-cloudkitty-transport", cr.Name) {
						// return namespace and Name of CR
						name := client.ObjectKey{
							Namespace: o.GetNamespace(),
							Name:      cr.Name,
						}
						Log.Info(fmt.Sprintf("TransportURL Secret %s belongs to TransportURL belonging to CloudKitty CR %s", o.GetName(), cr.Name))
						result = append(result, reconcile.Request{NamespacedName: name})
					}
				}
			}
		}
		if len(result) > 0 {
			return result
		}
		return nil
	}

	memcachedFn := func(ctx context.Context, o client.Object) []reconcile.Request {
		Log := r.GetLogger(ctx)

		result := []reconcile.Request{}

		// get all CloudKitty CRs
		cloudkitties := &telemetryv1.CloudKittyList{}
		listOpts := []client.ListOption{
			client.InNamespace(o.GetNamespace()),
		}
		if err := r.List(ctx, cloudkitties, listOpts...); err != nil {
			Log.Error(err, "Unable to retrieve CloudKitty CRs %w")
			return nil
		}

		for _, cr := range cloudkitties.Items {
			if o.GetName() == cr.Spec.MemcachedInstance {
				name := client.ObjectKey{
					Namespace: o.GetNamespace(),
					Name:      cr.Name,
				}
				Log.Info(fmt.Sprintf("Memcached %s is used by CloudKitty CR %s", o.GetName(), cr.Name))
				result = append(result, reconcile.Request{NamespacedName: name})
			}
		}
		if len(result) > 0 {
			return result
		}
		return nil
	}

	prometheusEndpointSecretFn := func(ctx context.Context, o client.Object) []reconcile.Request {
		Log := r.GetLogger(ctx)

		result := []reconcile.Request{}

		// Only reconcile if this is the PrometheusEndpoint secret
		if o.GetName() != cloudkitty.PrometheusEndpointSecret {
			return nil
		}

		// get all CloudKitty CRs
		cloudkitties := &telemetryv1.CloudKittyList{}
		listOpts := []client.ListOption{
			client.InNamespace(o.GetNamespace()),
		}
		if err := r.List(ctx, cloudkitties, listOpts...); err != nil {
			Log.Error(err, "Unable to retrieve CloudKitty CRs %w")
			return nil
		}

		for _, cr := range cloudkitties.Items {
			// Only reconcile CloudKitty CRs that are using MetricStorage (PrometheusHost is empty)
			if cr.Spec.PrometheusHost == "" {
				name := client.ObjectKey{
					Namespace: o.GetNamespace(),
					Name:      cr.Name,
				}
				Log.Info(fmt.Sprintf("PrometheusEndpoint Secret %s is used by CloudKitty CR %s", o.GetName(), cr.Name))
				result = append(result, reconcile.Request{NamespacedName: name})
			}
		}
		if len(result) > 0 {
			return result
		}
		return nil
	}

	control, err := ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1.CloudKitty{}).
		Owns(&mariadbv1.MariaDBDatabase{}).
		Owns(&mariadbv1.MariaDBAccount{}).
		Owns(&telemetryv1.CloudKittyAPI{}).
		Owns(&telemetryv1.CloudKittyProc{}).
		Owns(&rabbitmqv1.TransportURL{}).
		Owns(&batchv1.Job{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&certmgrv1.Certificate{}).
		// Watch for input secrets using field indexers
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForSrc),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		// Watch for TransportURL Secrets which belong to any TransportURLs created by CloudKitty CRs
		Watches(&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(transportURLSecretFn)).
		// Watch for PrometheusEndpoint Secret created by MetricStorage
		Watches(&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(prometheusEndpointSecretFn)).
		Watches(&memcachedv1.Memcached{},
			handler.EnqueueRequestsFromMapFunc(memcachedFn)).
		Watches(&keystonev1.KeystoneAPI{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectForSrc),
			builder.WithPredicates(keystonev1.KeystoneAPIStatusChangedPredicate)).
		// LokiStack watch added dynamically inside the controller code.
		Build(r)
	r.Controller = control
	return err
}

func (r *CloudKittyReconciler) findObjectForSrc(ctx context.Context, src client.Object) []reconcile.Request {
	requests := []reconcile.Request{}

	l := log.FromContext(ctx).WithName("Controllers").WithName("CloudKitty")

	crList := &telemetryv1.CloudKittyList{}
	listOps := &client.ListOptions{
		Namespace: src.GetNamespace(),
	}
	err := r.List(ctx, crList, listOps)
	if err != nil {
		l.Error(err, fmt.Sprintf("listing %s for namespace: %s", crList.GroupVersionKind().Kind, src.GetNamespace()))
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

	return requests
}

func (r *CloudKittyReconciler) findObjectsForSrc(ctx context.Context, src client.Object) []reconcile.Request {
	requests := []reconcile.Request{}

	l := log.FromContext(ctx).WithName("Controllers").WithName("CloudKitty")

	for _, field := range cloudKittyWatchFields {
		crList := &telemetryv1.CloudKittyList{}
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

func (r *CloudKittyReconciler) reconcileDelete(ctx context.Context, instance *telemetryv1.CloudKitty, helper *helper.Helper) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)

	Log.Info(fmt.Sprintf("Reconciling Service '%s' delete", instance.Name))

	// remove db finalizer first
	db, err := mariadbv1.GetDatabaseByNameAndAccount(ctx, helper, cloudkitty.DatabaseName, instance.Spec.DatabaseAccount, instance.Namespace)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if !k8s_errors.IsNotFound(err) {
		if err := db.DeleteFinalizer(ctx, helper); err != nil {
			return ctrl.Result{}, err
		}
	}

	// TODO: We might need to control how the sub-services (API and Proc) are
	// deleted (when their parent CloudKitty CR is deleted) once we further develop their functionality

	// Service is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(instance, helper.GetFinalizer())
	Log.Info(fmt.Sprintf("Reconciled Service '%s' delete successfully", instance.Name))

	return ctrl.Result{}, nil
}

func (r *CloudKittyReconciler) reconcileInit(
	ctx context.Context,
	instance *telemetryv1.CloudKitty,
	helper *helper.Helper,
	serviceLabels map[string]string,
	serviceAnnotations map[string]string,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)

	Log.Info(fmt.Sprintf("Reconciling Service '%s' init", instance.Name))

	//
	// run CloudKitty db sync
	//
	dbSyncHash := instance.Status.Hash[telemetryv1.CKDbSyncHash]
	jobDbSyncDef := cloudkitty.DbSyncJob(instance, serviceLabels, serviceAnnotations)

	dbSyncjob := job.NewJob(
		jobDbSyncDef,
		telemetryv1.CKDbSyncHash,
		instance.Spec.PreserveJobs,
		cloudkitty.ShortDuration,
		dbSyncHash,
	)
	ctrlResult, err := dbSyncjob.DoJob(
		ctx,
		helper,
	)
	if (ctrlResult != ctrl.Result{}) {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DBSyncReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.DBSyncReadyRunningMessage))
		return ctrlResult, nil
	}
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DBSyncReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.DBSyncReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}
	if dbSyncjob.HasChanged() {
		instance.Status.Hash[telemetryv1.CKDbSyncHash] = dbSyncjob.GetHash()
		Log.Info(fmt.Sprintf("Service '%s' - Job %s hash added - %s", instance.Name, jobDbSyncDef.Name, instance.Status.Hash[telemetryv1.CKDbSyncHash]))
	}
	instance.Status.Conditions.MarkTrue(condition.DBSyncReadyCondition, condition.DBSyncReadyMessage)

	// run CloudKitty db sync - end

	//
	// run CloudKitty Storage Init
	//
	ckStorageInitHash := instance.Status.Hash[telemetryv1.CKStorageInitHash]
	jobStorageInitDef := cloudkitty.StorageInitJob(instance, serviceLabels, serviceAnnotations)

	storageInitjob := job.NewJob(
		jobStorageInitDef,
		telemetryv1.CKStorageInitHash,
		instance.Spec.PreserveJobs,
		cloudkitty.ShortDuration,
		ckStorageInitHash,
	)
	ctrlResult, err = storageInitjob.DoJob(
		ctx,
		helper,
	)
	if (ctrlResult != ctrl.Result{}) {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CloudKittyStorageInitReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			telemetryv1.CloudKittyStorageInitReadyRunningMessage))
		return ctrlResult, nil
	}
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CloudKittyStorageInitReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			telemetryv1.CloudKittyStorageInitReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}
	if storageInitjob.HasChanged() {
		instance.Status.Hash[telemetryv1.CKStorageInitHash] = storageInitjob.GetHash()
		Log.Info(fmt.Sprintf("Service '%s' - Job %s hash added - %s", instance.Name, jobStorageInitDef.Name, instance.Status.Hash[telemetryv1.CKStorageInitHash]))
	}
	instance.Status.Conditions.MarkTrue(telemetryv1.CloudKittyStorageInitReadyCondition, telemetryv1.CloudKittyStorageInitReadyMessage)

	// run CloudKitty Storage Init - end

	Log.Info(fmt.Sprintf("Reconciled Service '%s' init successfully", instance.Name))
	return ctrl.Result{}, nil
}

// Original source:
// https://github.com/openstack-k8s-operators/openstack-operator/blob/cf133b39e91c05f53c57725d7c6f5a627d98dccd/pkg/openstack/ca.go#L687
func getCAFromSecret(
	ctx context.Context,
	instance *telemetryv1.CloudKitty,
	helper *helper.Helper,
	secretName string,
) (string, ctrl.Result, error) {
	caSecret, ctrlResult, err := secret.GetDataFromSecret(ctx, helper, secretName, time.Duration(5), "ca.crt")
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CloudKittyLokiStackReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			telemetryv1.CloudKittyLokiStackReadyErrorMessage,
			err.Error()))

		return "", ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CloudKittyLokiStackReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			telemetryv1.CloudKittyLokiStackReadyRunningMessage))

		return "", ctrlResult, nil
	}

	return caSecret, ctrl.Result{}, nil
}

func (r *CloudKittyReconciler) reconcileNormal(ctx context.Context, instance *telemetryv1.CloudKitty, helper *helper.Helper) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)

	Log.Info(fmt.Sprintf("Reconciling Service '%s'", instance.Name))

	serviceLabels := map[string]string{
		common.AppSelector: cloudkitty.ServiceName,
	}

	// Create cloudkitty client cert / key
	certIssuer, err := certmanager.GetIssuerByLabels(
		ctx, helper, instance.Namespace,
		map[string]string{certmanager.RootCAIssuerInternalLabel: ""},
	)
	if err != nil {
		Log.Error(err, "Failed to determine certificate issuer")
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CloudKittyClientCertReadyCondition,
			condition.ErrorReason,
			condition.SeverityError,
			telemetryv1.CloudKittyClientCertReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}
	certDefinition := cloudkitty.Certificate(
		instance, serviceLabels, certIssuer,
	)
	cert := certmanager.NewCertificate(certDefinition, 5*time.Second)
	ctrlResult, _, err := cert.CreateOrPatch(ctx, helper, nil)

	if err != nil {
		Log.Error(err, "Failed to create or patch cloudkitty client certificate")
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CloudKittyClientCertReadyCondition,
			condition.ErrorReason,
			condition.SeverityError,
			telemetryv1.CloudKittyClientCertReadyErrorMessage,
			err.Error()))
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		Log.Info("CloudKitty client certificate is being created")
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CloudKittyClientCertReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			telemetryv1.CloudKittyClientCertReadyRunningMessage))
		return ctrlResult, nil
	}

	caData, ctrlResult, err := getCAFromSecret(
		ctx, instance, helper, certDefinition.Spec.SecretName,
	)
	if err != nil {
		Log.Error(err, "Failed to get cloudkitty client certificate CA data")
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CloudKittyClientCertReadyCondition,
			condition.ErrorReason,
			condition.SeverityError,
			telemetryv1.CloudKittyClientCertReadyErrorMessage,
			err.Error()))
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	cms := []util.Template{
		{
			Name:         fmt.Sprintf("%s-%s", instance.Name, cloudkitty.CaConfigmapName),
			Namespace:    instance.Namespace,
			Type:         util.TemplateTypeNone,
			InstanceType: "cloudkitty",
			CustomData: map[string]string{
				cloudkitty.CaConfigmapKey: caData,
			},
		},
	}

	err = configmap.EnsureConfigMaps(
		ctx, helper, instance, cms, nil,
	)
	if err != nil {
		Log.Error(err, "Failed to create CA configmap for cloudkitty client cert verification")
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CloudKittyClientCertReadyCondition,
			condition.ErrorReason,
			condition.SeverityError,
			telemetryv1.CloudKittyClientCertReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}

	instance.Status.Conditions.MarkTrue(telemetryv1.CloudKittyClientCertReadyCondition, telemetryv1.CloudKittyClientCertReadyMessage)

	// Deploy Loki
	var eventHandler = handler.EnqueueRequestForOwner(
		r.Scheme,
		r.RESTMapper,
		&telemetryv1.CloudKitty{},
		handler.OnlyControllerOwner(),
	)

	err = utils.EnsureWatches(
		ctx, (*utils.ConditionalWatchingReconciler)(r),
		"lokistacks.loki.grafana.com",
		&lokistackv1.LokiStack{}, eventHandler, helper,
	)
	if err != nil {
		instance.Status.Conditions.MarkFalse(telemetryv1.CloudKittyLokiStackReadyCondition,
			condition.Reason("Can't own LokiStack resource. The loki-operator probably isn't installed"),
			condition.SeverityError,
			telemetryv1.CloudKittyLokiStackUnableToOwnMessage, err)
		Log.Info("Can't own LokiStack resource. The loki-operator probably isn't installed")
		return ctrl.Result{RequeueAfter: telemetryv1.PauseBetweenWatchAttempts}, nil
	}

	lokiStack := &lokistackv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-lokistack", instance.Name),
			Namespace: instance.Namespace,
		},
	}
	op, err := controllerutil.CreateOrPatch(ctx, r.Client, lokiStack, func() error {
		desiredLokiStack, err := cloudkitty.LokiStack(instance, serviceLabels)
		if err != nil {
			return err
		}
		desiredLokiStack.Spec.DeepCopyInto(&lokiStack.Spec)
		lokiStack.Labels = serviceLabels
		err = controllerutil.SetControllerReference(instance, lokiStack, r.Scheme)
		return err
	})
	if err != nil {
		Log.Error(err, "Failed to create or patch LokiStack")
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CloudKittyLokiStackReadyCondition,
			condition.ErrorReason,
			condition.SeverityError,
			telemetryv1.CloudKittyLokiStackReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		Log.Info(fmt.Sprintf("LokiStack %s successfully changed - operation: %s", lokiStack.Name, string(op)))
	}

	// Mirror LokiStacks's condition here. LokiStack uses conditions
	// a little differently than o-k-o. Whats more, it can have
	// multiple 'active' conditions, while we have only one 'master'
	// condition LokiStack here. So we mirror hopefully the one most
	// relevant active condition in this order of
	// priority: Ready > Failed > Degraded > Pending > Warning.

	order := []string{"Ready", "Failed", "Degraded", "Pending", "Warning"}
	index := len(order)
	reason := condition.InitReason
	message := telemetryv1.CloudKittyLokiStackReadyInitMessage
	for _, c := range lokiStack.Status.Conditions {
		conditionIndex := slices.Index(order, c.Type)
		if c.Status == "True" && conditionIndex < index {
			index = conditionIndex
			reason = c.Reason
			message = c.Message
		}
	}
	if index < len(order) && order[index] == "Ready" {
		instance.Status.Conditions.MarkTrue(telemetryv1.CloudKittyLokiStackReadyCondition, telemetryv1.CloudKittyLokiStackReadyMessage)
	} else {
		Log.Info("LokiStack not ready")
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CloudKittyLokiStackReadyCondition,
			condition.Reason(reason),
			condition.SeverityWarning,
			"LokiStack issue: %s", message))
	}

	// Service account, role, binding
	rbacRules := []rbacv1.PolicyRule{
		{
			APIGroups:     []string{"security.openshift.io"},
			ResourceNames: []string{"anyuid"},
			Resources:     []string{"securitycontextconstraints"},
			Verbs:         []string{"use"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"pods"},
			Verbs:     []string{"create", "get", "list", "watch", "update", "patch", "delete"},
		},
	}
	rbacResult, err := common_rbac.ReconcileRbac(ctx, helper, instance, rbacRules)
	if err != nil {
		return rbacResult, err
	} else if (rbacResult != ctrl.Result{}) {
		return rbacResult, nil
	}

	configVars := make(map[string]env.Setter)

	//
	// create RabbitMQ transportURL CR and get the actual URL from the associated secret that is created
	//

	transportURL, op, err := r.transportURLCreateOrUpdate(ctx, instance, serviceLabels)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.RabbitMqTransportURLReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.RabbitMqTransportURLReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}

	if op != controllerutil.OperationResultNone {
		Log.Info(fmt.Sprintf("TransportURL %s successfully reconciled - operation: %s", transportURL.Name, string(op)))
	}

	instance.Status.TransportURLSecret = transportURL.Status.SecretName

	if instance.Status.TransportURLSecret == "" {
		Log.Info(fmt.Sprintf("Waiting for TransportURL %s secret to be created", transportURL.Name))
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.RabbitMqTransportURLReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.RabbitMqTransportURLReadyRunningMessage))
		return cloudkitty.ResultRequeue, nil
	}

	instance.Status.Conditions.MarkTrue(condition.RabbitMqTransportURLReadyCondition, condition.RabbitMqTransportURLReadyMessage)

	// end transportURL

	//
	// Check for required memcached used for caching
	//
	memcached, err := memcachedv1.GetMemcachedByName(ctx, helper, instance.Spec.MemcachedInstance, instance.Namespace)
	if err != nil {
		Log.Info(fmt.Sprintf("%s... requeueing", condition.MemcachedReadyWaitingMessage))
		if k8s_errors.IsNotFound(err) {
			Log.Info(fmt.Sprintf("memcached %s not found", instance.Spec.MemcachedInstance))
			instance.Status.Conditions.Set(condition.FalseCondition(
				condition.MemcachedReadyCondition,
				condition.RequestedReason,
				condition.SeverityInfo,
				condition.MemcachedReadyWaitingMessage))
			return cloudkitty.ResultRequeue, nil
		}
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.MemcachedReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.MemcachedReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}

	if !memcached.IsReady() {
		Log.Info(fmt.Sprintf("%s... requeueing", condition.MemcachedReadyWaitingMessage))
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.MemcachedReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.MemcachedReadyWaitingMessage))
		return cloudkitty.ResultRequeue, nil
	}
	// Mark the Memcached Service as Ready if we get to this point with no errors
	instance.Status.Conditions.MarkTrue(
		condition.MemcachedReadyCondition, condition.MemcachedReadyMessage)
	// run check memcached - end

	//
	// Check for PrometheusEndpoint secret if using MetricStorage
	//
	if instance.Spec.PrometheusHost == "" {
		prometheusEndpointSecret := &corev1.Secret{}
		err = r.Get(ctx, client.ObjectKey{
			Name:      cloudkitty.PrometheusEndpointSecret,
			Namespace: instance.Namespace,
		}, prometheusEndpointSecret)
		if err != nil {
			if k8s_errors.IsNotFound(err) {
				Log.Info("PrometheusEndpoint Secret not found. CloudKitty will not be deployed until MetricStorage creates it.")
				instance.Status.Conditions.Set(condition.FalseCondition(
					condition.ServiceConfigReadyCondition,
					condition.Reason("PrometheusEndpoint secret not found. The MetricStorage probably hasn't been created yet or isn't ready"),
					condition.SeverityError,
					"PrometheusEndpoint secret %s not found. Waiting for MetricStorage to create it",
					cloudkitty.PrometheusEndpointSecret))
				return ctrl.Result{RequeueAfter: telemetryv1.PauseBetweenWatchAttempts}, nil
			}
			instance.Status.Conditions.Set(condition.FalseCondition(
				condition.ServiceConfigReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.ServiceConfigReadyErrorMessage,
				err.Error()))
			return ctrl.Result{}, err
		}
	}
	// run check PrometheusEndpoint secret - end

	//
	// check for required OpenStack secret holding passwords for service/admin user and add hash to the vars map
	//

	result, err := cloudkitty.VerifyServiceSecret(
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
		return result, err
	} else if (result != ctrl.Result{}) {
		return result, nil
	}
	instance.Status.Conditions.MarkTrue(condition.InputReadyCondition, condition.InputReadyMessage)
	// run check OpenStack secret - end

	db, result, err := r.ensureDB(ctx, helper, instance)
	if err != nil {
		return ctrl.Result{}, err
	} else if (result != ctrl.Result{}) {
		return result, nil
	}

	//
	// Create Secrets required as input for the Service and calculate an overall hash of hashes
	//
	err = r.generateServiceConfigs(ctx, helper, instance, &configVars, serviceLabels, memcached, db)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.ServiceConfigReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.ServiceConfigReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}

	//
	// create hash over all the different input resources to identify if any those changed
	// and a restart/recreate is required.
	//
	_, hashChanged, err := r.createHashOfInputHashes(ctx, instance, configVars)
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

	instance.Status.Conditions.MarkTrue(condition.ServiceConfigReadyCondition, condition.ServiceConfigReadyMessage)

	//
	// TODO check when/if Init, Update, or Upgrade should/could be skipped
	//

	// Check networks that the DBSync job will use in reconcileInit. The ones from the API service are always enough,
	// it doesn't need the storage specific ones that volume or backup may have.
	nadList := []networkv1.NetworkAttachmentDefinition{}
	for _, netAtt := range instance.Spec.CloudKittyAPI.NetworkAttachments {
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
				//nolint:err113 // Using condition message format from lib-common
				return cloudkitty.ResultRequeue, fmt.Errorf(condition.NetworkAttachmentsReadyWaitingMessage, netAtt)
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

	instance.Status.Conditions.MarkTrue(condition.NetworkAttachmentsReadyCondition, condition.NetworkAttachmentsReadyMessage)

	serviceAnnotations, err := nad.EnsureNetworksAnnotation(nadList)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed create network annotation from %s: %w",
			instance.Spec.CloudKittyAPI.NetworkAttachments, err)
	}

	// Handle service init
	ctrlResult, err = r.reconcileInit(ctx, instance, helper, serviceLabels, serviceAnnotations)
	if err != nil {
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	//
	// normal reconcile tasks
	//

	// deploy cloudkitty-api
	cloudKittyAPI, op, err := r.apiDeploymentCreateOrUpdate(ctx, instance)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CloudKittyAPIReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			telemetryv1.CloudKittyAPIReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		Log.Info(fmt.Sprintf("API CR for %s successfully %s", instance.Name, string(op)))
	}

	// Mirror values when the data in the StatefulSet is for the current generation
	if cloudKittyAPI.Generation == cloudKittyAPI.Status.ObservedGeneration {
		// Mirror CloudKittyAPI status' APIEndpoints and ReadyCount to this parent CR
		instance.Status.APIEndpoints = cloudKittyAPI.Status.APIEndpoints
		instance.Status.ServiceIDs = cloudKittyAPI.Status.ServiceIDs
		instance.Status.CloudKittyAPIReadyCount = cloudKittyAPI.Status.ReadyCount

		// Mirror CloudKittyAPI's condition status
		c := cloudKittyAPI.Status.Conditions.Mirror(telemetryv1.CloudKittyAPIReadyCondition)
		if c != nil {
			instance.Status.Conditions.Set(c)
		}
	}

	// deploy CloudKitty Processor
	cloudKittyProc, op, err := r.procDeploymentCreateOrUpdate(ctx, instance)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CloudKittyProcReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			telemetryv1.CloudKittyProcReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		Log.Info(fmt.Sprintf("Scheduler CR for %s successfully %s", instance.Name, string(op)))
	}

	// Mirror values when the data in the StatefulSet is for the current generation
	if cloudKittyProc.Generation == cloudKittyProc.Status.ObservedGeneration {
		// Mirror CloudKitty Processor status' ReadyCount to this parent CR
		instance.Status.CloudKittyProcReadyCount = cloudKittyProc.Status.ReadyCount

		// Mirror CloudKittyProc's condition status
		c := cloudKittyProc.Status.Conditions.Mirror(telemetryv1.CloudKittyProcReadyCondition)
		if c != nil {
			instance.Status.Conditions.Set(c)
		}
	}

	err = mariadbv1.DeleteUnusedMariaDBAccountFinalizers(ctx, helper, cloudkitty.DatabaseName, instance.Spec.DatabaseAccount, instance.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}

	Log.Info(fmt.Sprintf("Reconciled Service '%s' successfully", instance.Name))
	// update the overall status condition if service is ready
	if instance.IsReady() {
		instance.Status.Conditions.MarkTrue(condition.ReadyCondition, condition.ReadyMessage)
	}
	return ctrl.Result{}, nil
}

// generateServiceConfigs - create Secret which hold scripts and service configuration
func (r *CloudKittyReconciler) generateServiceConfigs(
	ctx context.Context,
	h *helper.Helper,
	instance *telemetryv1.CloudKitty,
	envVars *map[string]env.Setter,
	serviceLabels map[string]string,
	memcached *memcachedv1.Memcached,
	db *mariadbv1.Database,
) error {
	Log := r.GetLogger(ctx)
	//
	// create Secret required for cloudkitty input
	// - %-scripts holds scripts to e.g. bootstrap the service
	// - %-config holds minimal cloudkitty config required to get the service up
	//

	labels := labels.GetLabels(instance, labels.GetGroupLabel(cloudkitty.ServiceName), serviceLabels)

	var tlsCfg *tls.Service
	if instance.Spec.CloudKittyAPI.TLS.CaBundleSecretName != "" {
		tlsCfg = &tls.Service{}
	}

	// customData hold any customization for all cloudkitty services.
	customData := map[string]string{
		cloudkitty.CustomConfigFileName: instance.Spec.CustomServiceConfig,
		cloudkitty.MyCnfFileName:        db.GetDatabaseClientConfig(tlsCfg), //(mschuppert) for now just get the default my.cnf
	}

	keystoneAPI, err := keystonev1.GetKeystoneAPI(ctx, h, instance.Namespace, map[string]string{})
	if err != nil {
		return err
	}
	keystoneInternalURL, err := keystoneAPI.GetEndpoint(endpoint.EndpointInternal)
	if err != nil {
		return err
	}
	keystonePublicURL, err := keystoneAPI.GetEndpoint(endpoint.EndpointPublic)
	if err != nil {
		return err
	}

	ospSecret, _, err := secret.GetSecret(ctx, h, instance.Spec.Secret, instance.Namespace)
	if err != nil {
		return err
	}

	transportURLSecret, _, err := secret.GetSecret(ctx, h, instance.Status.TransportURLSecret, instance.Namespace)
	if err != nil {
		return err
	}

	if instance.Spec.PrometheusHost == "" {
		// We're using MetricStorage for Prometheus.
		// Note: The secret existence is already checked in reconcileNormal(), so we can safely get it here
		prometheusEndpointSecret := &corev1.Secret{}
		err = r.Get(ctx, client.ObjectKey{
			Name:      cloudkitty.PrometheusEndpointSecret,
			Namespace: instance.Namespace,
		}, prometheusEndpointSecret)
		if err != nil {
			return err
		}
		if prometheusEndpointSecret.Data != nil {
			instance.Status.PrometheusHost = string(prometheusEndpointSecret.Data[metricstorage.PrometheusHost])
			port, err := strconv.Atoi(string(prometheusEndpointSecret.Data[metricstorage.PrometheusPort]))
			if err != nil {
				return err
			}
			//nolint:gosec // G109: Port number is read from a secret and validated to be within valid range
			instance.Status.PrometheusPort = int32(port)

			metricStorage := &telemetryv1.MetricStorage{}
			err = r.Get(ctx, client.ObjectKey{
				Namespace: instance.Namespace,
				Name:      telemetryv1.DefaultServiceName,
			}, metricStorage)
			if err != nil {
				instance.Status.Conditions.Set(condition.FalseCondition(
					condition.ServiceConfigReadyCondition,
					condition.ErrorReason,
					condition.SeverityWarning,
					condition.ServiceConfigReadyErrorMessage,
					err.Error()))
				return err
			}
			instance.Status.PrometheusTLS = metricStorage.Spec.PrometheusTLS.Enabled()
		}
	} else {
		// We're using user-deployed Prometheus.
		instance.Status.PrometheusHost = instance.Spec.PrometheusHost
		instance.Status.PrometheusPort = instance.Spec.PrometheusPort
		instance.Status.PrometheusTLS = instance.Spec.PrometheusTLSCaCertSecret != nil
	}

	databaseAccount := db.GetAccount()
	dbSecret := db.GetSecret()

	lokiHost := fmt.Sprintf("%s-lokistack-gateway-http.%s.svc", instance.Name, instance.Namespace)

	templateParameters := make(map[string]interface{})
	templateParameters["ServiceUser"] = instance.Spec.ServiceUser
	templateParameters["ServicePassword"] = string(ospSecret.Data[instance.Spec.PasswordSelectors.CloudKittyService])
	templateParameters["KeystoneInternalURL"] = keystoneInternalURL

	// Try to get Application Credential from the secret specified in the CR
	if instance.Spec.Auth.ApplicationCredentialSecret != "" {
		acSecretObj, _, err := secret.GetSecret(ctx, h, instance.Spec.Auth.ApplicationCredentialSecret, instance.Namespace)
		if err != nil {
			if k8s_errors.IsNotFound(err) {
				Log.Info("ApplicationCredential secret not found, waiting", "secret", instance.Spec.Auth.ApplicationCredentialSecret)
				return fmt.Errorf("%w: %s", ErrACSecretNotFound, instance.Spec.Auth.ApplicationCredentialSecret)
			}
			Log.Error(err, "Failed to get ApplicationCredential secret", "secret", instance.Spec.Auth.ApplicationCredentialSecret)
			return err
		}
		acID, okID := acSecretObj.Data[keystonev1.ACIDSecretKey]
		acSecretData, okSecret := acSecretObj.Data[keystonev1.ACSecretSecretKey]
		if !okID || len(acID) == 0 || !okSecret || len(acSecretData) == 0 {
			Log.Info("ApplicationCredential secret missing required keys", "secret", instance.Spec.Auth.ApplicationCredentialSecret)
			return fmt.Errorf("%w: %s", ErrACSecretMissingKeys, instance.Spec.Auth.ApplicationCredentialSecret)
		}
		templateParameters["ACID"] = string(acID)
		templateParameters["ACSecret"] = string(acSecretData)
		Log.Info("Using ApplicationCredentials auth", "secret", instance.Spec.Auth.ApplicationCredentialSecret)
	}
	templateParameters["KeystonePublicURL"] = keystonePublicURL
	templateParameters["Region"] = keystoneAPI.GetRegion()
	templateParameters["TransportURL"] = string(transportURLSecret.Data["transport_url"])
	templateParameters["PrometheusHost"] = instance.Status.PrometheusHost
	templateParameters["PrometheusPort"] = instance.Status.PrometheusPort
	templateParameters["Period"] = instance.Spec.Period
	templateParameters["LokiHost"] = lokiHost
	templateParameters["LokiPort"] = 8080
	templateParameters["DatabaseConnection"] = fmt.Sprintf("mysql+pymysql://%s:%s@%s/%s?read_default_file=/etc/my.cnf",
		databaseAccount.Spec.UserName,
		string(dbSecret.Data[mariadbv1.DatabasePasswordSelector]),
		instance.Status.DatabaseHostname,
		cloudkitty.DatabaseName)
	templateParameters["MemcachedServersWithInet"] = memcached.GetMemcachedServerListWithInetString()
	templateParameters["TimeOut"] = instance.Spec.APITimeout

	templateParameters["TLS"] = false
	if instance.Spec.CloudKittyProc.TLS.Enabled() {
		templateParameters["TLS"] = true
		templateParameters["CAFile"] = tls.DownstreamTLSCABundlePath
	}

	// Set Prometheus TLS configuration
	templateParameters["PrometheusTLS"] = instance.Status.PrometheusTLS
	if instance.Status.PrometheusTLS {
		// For operator-managed Prometheus or user-deployed Prometheus with custom CA,
		// use the downstream TLS CA bundle path
		templateParameters["PrometheusCAFile"] = tls.DownstreamTLSCABundlePath
	}

	// create httpd  vhost template parameters
	httpdVhostConfig := map[string]interface{}{}
	for _, endpt := range []service.Endpoint{service.EndpointInternal, service.EndpointPublic} {
		endptConfig := map[string]interface{}{}
		endptConfig["ServerName"] = fmt.Sprintf("%s-%s.%s.svc", cloudkitty.ServiceName, endpt.String(), instance.Namespace)
		endptConfig["TLS"] = false // default TLS to false, and set it bellow to true if enabled
		if instance.Spec.CloudKittyAPI.TLS.API.Enabled(endpt) {
			endptConfig["TLS"] = true
			endptConfig["SSLCertificateFile"] = fmt.Sprintf("/etc/pki/tls/certs/%s.crt", endpt.String())
			endptConfig["SSLCertificateKeyFile"] = fmt.Sprintf("/etc/pki/tls/private/%s.key", endpt.String())
		}
		httpdVhostConfig[endpt.String()] = endptConfig
	}
	templateParameters["VHosts"] = httpdVhostConfig

	configTemplates := []util.Template{
		{
			Name:         fmt.Sprintf("%s-scripts", instance.Name),
			Namespace:    instance.Namespace,
			Type:         util.TemplateTypeScripts,
			InstanceType: instance.Kind,
			Labels:       labels,
		},
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
func (r *CloudKittyReconciler) createHashOfInputHashes(
	ctx context.Context,
	instance *telemetryv1.CloudKitty,
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

func (r *CloudKittyReconciler) transportURLCreateOrUpdate(
	ctx context.Context,
	instance *telemetryv1.CloudKitty,
	serviceLabels map[string]string,
) (*rabbitmqv1.TransportURL, controllerutil.OperationResult, error) {
	transportURL := &rabbitmqv1.TransportURL{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-transport", instance.Name),
			Namespace: instance.Namespace,
			Labels:    serviceLabels,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, transportURL, func() error {
		transportURL.Spec.RabbitmqClusterName = instance.Spec.RabbitMqClusterName

		err := controllerutil.SetControllerReference(instance, transportURL, r.Scheme)
		return err
	})

	return transportURL, op, err
}

func (r *CloudKittyReconciler) apiDeploymentCreateOrUpdate(ctx context.Context, instance *telemetryv1.CloudKitty) (*telemetryv1.CloudKittyAPI, controllerutil.OperationResult, error) {
	cloudkittyAPISpec := telemetryv1.CloudKittyAPISpec{
		CloudKittyTemplate:    instance.Spec.CloudKittyTemplate,
		CloudKittyAPITemplate: instance.Spec.CloudKittyAPI,
		DatabaseHostname:      instance.Status.DatabaseHostname,
		TransportURLSecret:    instance.Status.TransportURLSecret,
		ServiceAccount:        instance.RbacResourceName(),
	}

	if cloudkittyAPISpec.NodeSelector == nil {
		cloudkittyAPISpec.NodeSelector = instance.Spec.NodeSelector
	}

	// If topology is not present in the underlying CloudKittyAPI Spec,
	// inherit from the top-level CR
	if cloudkittyAPISpec.TopologyRef == nil {
		cloudkittyAPISpec.TopologyRef = instance.Spec.TopologyRef
	}

	deployment := &telemetryv1.CloudKittyAPI{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-api", instance.Name),
			Namespace: instance.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		deployment.Spec = cloudkittyAPISpec

		err := controllerutil.SetControllerReference(instance, deployment, r.Scheme)
		if err != nil {
			return err
		}

		return nil
	})

	return deployment, op, err
}

func (r *CloudKittyReconciler) procDeploymentCreateOrUpdate(ctx context.Context, instance *telemetryv1.CloudKitty) (*telemetryv1.CloudKittyProc, controllerutil.OperationResult, error) {
	cloudKittyProcSpec := telemetryv1.CloudKittyProcSpec{
		CloudKittyTemplate:     instance.Spec.CloudKittyTemplate,
		CloudKittyProcTemplate: instance.Spec.CloudKittyProc,
		DatabaseHostname:       instance.Status.DatabaseHostname,
		TransportURLSecret:     instance.Status.TransportURLSecret,
		ServiceAccount:         instance.RbacResourceName(),
		//TLS:                    instance.Spec.CloudKittyProc.TLS.Ca,
	}

	if cloudKittyProcSpec.NodeSelector == nil {
		cloudKittyProcSpec.NodeSelector = instance.Spec.NodeSelector
	}

	// If topology is not present in the underlying Scheduler Spec
	// inherit from the top-level CR
	if cloudKittyProcSpec.TopologyRef == nil {
		cloudKittyProcSpec.TopologyRef = instance.Spec.TopologyRef
	}

	deployment := &telemetryv1.CloudKittyProc{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-proc", instance.Name),
			Namespace: instance.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		deployment.Spec = cloudKittyProcSpec

		err := controllerutil.SetControllerReference(instance, deployment, r.Scheme)
		if err != nil {
			return err
		}

		return nil
	})

	return deployment, op, err
}

func (r *CloudKittyReconciler) ensureDB(
	ctx context.Context,
	h *helper.Helper,
	instance *telemetryv1.CloudKitty,
) (*mariadbv1.Database, ctrl.Result, error) {
	Log := r.GetLogger(ctx)

	// ensure MariaDBAccount exists.  This account record may be created by
	// openstack-operator or the cloud operator up front without a specific
	// MariaDBDatabase configured yet.   Otherwise, a MariaDBAccount CR is
	// created here with a generated username as well as a secret with
	// generated password.   The MariaDBAccount is created without being
	// yet associated with any MariaDBDatabase.
	_, _, err := mariadbv1.EnsureMariaDBAccount(
		ctx, h, instance.Spec.DatabaseAccount,
		instance.Namespace, false, "cloudkitty",
	)

	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			mariadbv1.MariaDBAccountReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			mariadbv1.MariaDBAccountNotReadyMessage,
			err.Error()))

		return nil, ctrl.Result{}, err
	}
	instance.Status.Conditions.MarkTrue(
		mariadbv1.MariaDBAccountReadyCondition,
		mariadbv1.MariaDBAccountReadyMessage,
	)

	db := mariadbv1.NewDatabaseForAccount(
		instance.Spec.DatabaseInstance, // mariadb/galera service to target
		cloudkitty.DatabaseName,        // name used in CREATE DATABASE in mariadb
		cloudkitty.DatabaseName,        // CR name for MariaDBDatabase
		instance.Spec.DatabaseAccount,  // CR name for MariaDBAccount
		instance.Namespace,             // namespace
	)

	// create or patch the DB
	ctrlResult, err := db.CreateOrPatchAll(ctx, h)

	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DBReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.DBReadyErrorMessage,
			err.Error()))
		return db, ctrl.Result{}, err
	}
	if (ctrlResult != ctrl.Result{}) {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DBReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.DBReadyRunningMessage))
		return db, ctrlResult, nil
	}
	// wait for the DB to be setup
	// (ksambor) should we use WaitForDBCreatedWithTimeout instead?
	ctrlResult, err = db.WaitForDBCreated(ctx, h)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DBReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.DBReadyErrorMessage,
			err.Error()))
		return db, ctrlResult, err
	}
	if (ctrlResult != ctrl.Result{}) {
		Log.Info(fmt.Sprintf("%s... requeueing", condition.DBReadyRunningMessage))
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DBReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.DBReadyRunningMessage))
		return db, ctrlResult, nil
	}

	// update Status.DatabaseHostname, used to config the service
	instance.Status.DatabaseHostname = db.GetDatabaseHostname()
	instance.Status.Conditions.MarkTrue(condition.DBReadyCondition, condition.DBReadyMessage)
	return db, ctrlResult, nil
}
