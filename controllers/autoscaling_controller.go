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
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	logr "github.com/go-logr/logr"
	common "github.com/openstack-k8s-operators/lib-common/modules/common"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	endpoint "github.com/openstack-k8s-operators/lib-common/modules/common/endpoint"
	env "github.com/openstack-k8s-operators/lib-common/modules/common/env"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	labels "github.com/openstack-k8s-operators/lib-common/modules/common/labels"
	common_rbac "github.com/openstack-k8s-operators/lib-common/modules/common/rbac"
	secret "github.com/openstack-k8s-operators/lib-common/modules/common/secret"
	util "github.com/openstack-k8s-operators/lib-common/modules/common/util"

	heatv1 "github.com/openstack-k8s-operators/heat-operator/api/v1beta1"
	memcachedv1 "github.com/openstack-k8s-operators/infra-operator/apis/memcached/v1beta1"
	rabbitmqv1 "github.com/openstack-k8s-operators/infra-operator/apis/rabbitmq/v1beta1"
	keystonev1 "github.com/openstack-k8s-operators/keystone-operator/api/v1beta1"
	mariadbv1 "github.com/openstack-k8s-operators/mariadb-operator/api/v1beta1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	autoscaling "github.com/openstack-k8s-operators/telemetry-operator/pkg/autoscaling"
)

// AutoscalingReconciler reconciles a Autoscaling object
type AutoscalingReconciler struct {
	client.Client
	Kclient kubernetes.Interface
	Scheme  *runtime.Scheme
}

// GetLogger returns a logger object with a prefix of "conroller.name" and aditional controller context fields
func (r *AutoscalingReconciler) GetLogger(ctx context.Context) logr.Logger {
	return log.FromContext(ctx).WithName("Controllers").WithName("Autoscaling")
}

// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=autoscalings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=autoscalings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=autoscalings/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete;
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete;
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete;
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;create;update;patch;delete;watch
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list
// +kubebuilder:rbac:groups=keystone.openstack.org,resources=keystoneservices,verbs=get;list;watch;create;update;patch;delete;
// +kubebuilder:rbac:groups=keystone.openstack.org,resources=keystoneapis,verbs=get;list;watch;
// +kubebuilder:rbac:groups=keystone.openstack.org,resources=keystoneendpoints,verbs=get;list;watch;create;update;patch;delete;
// +kubebuilder:rbac:groups=rabbitmq.openstack.org,resources=transporturls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mariadb.openstack.org,resources=mariadbdatabases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mariadb.openstack.org,resources=mariadbdatabases/finalizers,verbs=update
// +kubebuilder:rbac:groups=mariadb.openstack.org,resources=mariadbaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mariadb.openstack.org,resources=mariadbaccounts/finalizers,verbs=update
// +kubebuilder:rbac:groups=memcached.openstack.org,resources=memcacheds,verbs=get;list;watch;
// +kubebuilder:rbac:groups=heat.openstack.org,resources=heats,verbs=get;list;watch;
// service account, role, rolebinding
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;list;watch;create;update
// service account permissions that are needed to grant permission to the above
// +kubebuilder:rbac:groups="security.openshift.io",resourceNames=anyuid,resources=securitycontextconstraints,verbs=use

// Reconcile reconciles an Autoscaling
func (r *AutoscalingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)

	// Fetch the Autoscaling instance
	instance := &telemetryv1.Autoscaling{}
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
			// service account, role, rolebinding conditions
			condition.UnknownCondition(condition.ServiceAccountReadyCondition,
				condition.InitReason,
				condition.ServiceAccountReadyInitMessage),
			condition.UnknownCondition(condition.RoleReadyCondition, condition.InitReason, condition.RoleReadyInitMessage),
			condition.UnknownCondition(condition.RoleBindingReadyCondition,
				condition.InitReason,
				condition.RoleBindingReadyInitMessage),

			// Prometheus, Aodh, Heat conditions
			condition.UnknownCondition(telemetryv1.HeatReadyCondition, condition.InitReason, telemetryv1.HeatReadyInitMessage),
			condition.UnknownCondition(condition.MemcachedReadyCondition, condition.InitReason, condition.MemcachedReadyInitMessage),

			condition.UnknownCondition(condition.ServiceConfigReadyCondition, condition.InitReason, condition.ServiceConfigReadyInitMessage),
			condition.UnknownCondition(condition.DBReadyCondition, condition.InitReason, condition.DBReadyInitMessage),
			condition.UnknownCondition(condition.DBSyncReadyCondition, condition.InitReason, condition.DBSyncReadyInitMessage),
			condition.UnknownCondition(condition.RabbitMqTransportURLReadyCondition, condition.InitReason, condition.RabbitMqTransportURLReadyInitMessage),

			condition.UnknownCondition(condition.DeploymentReadyCondition, condition.InitReason, condition.DeploymentReadyInitMessage),
			// right now we have no dedicated KeystoneServiceReadyInitMessage
			condition.UnknownCondition(condition.KeystoneServiceReadyCondition, condition.InitReason, ""),
			condition.UnknownCondition(condition.KeystoneEndpointReadyCondition, condition.InitReason, ""),
			condition.UnknownCondition(condition.TLSInputReadyCondition, condition.InitReason, condition.InputReadyInitMessage),
		)

		instance.Status.Conditions.Init(&cl)

		// Register overall status immediately to have an early feedback e.g. in the cli
		return ctrl.Result{}, nil
	}

	if instance.Status.Hash == nil {
		instance.Status.Hash = map[string]string{}
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
	passwordSecretField     = ".spec.secret"
	caBundleSecretNameField = ".spec.tls.caBundleSecretName"
	tlsAPIInternalField     = ".spec.tls.api.internal.secretName"
	tlsAPIPublicField       = ".spec.tls.api.public.secretName"
	tlsField                = ".spec.tls.secretName"
)

var (
	allWatchFields = []string{
		passwordSecretField,
		caBundleSecretNameField,
		tlsAPIInternalField,
		tlsAPIPublicField,
		tlsField,
	}
)

func (r *AutoscalingReconciler) reconcileDelete(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info("Reconciling Service delete")
	ctrlResult, err := r.reconcileDeleteAodh(ctx, instance, helper)
	if err != nil {
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}
	// Service is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(instance, helper.GetFinalizer())
	Log.Info(fmt.Sprintf("Reconciled Service '%s' delete successfully", autoscaling.ServiceName))

	return ctrl.Result{}, nil
}

func (r *AutoscalingReconciler) reconcileInit(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
	serviceLabels map[string]string,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info("Reconciling Service init")
	ctrlResult, err := r.reconcileInitAodh(ctx, instance, helper, serviceLabels)
	if err != nil {
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}
	Log.Info("Reconciled Service init successfully")
	return ctrl.Result{}, nil
}

func (r *AutoscalingReconciler) reconcileNormal(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info(fmt.Sprintf("Reconciling Service '%s'", autoscaling.ServiceName))

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

	instance.Status.Conditions.MarkTrue(condition.ServiceConfigReadyCondition, condition.ServiceConfigReadyMessage)

	serviceLabels := map[string]string{
		common.AppSelector: autoscaling.ServiceName,
	}

	//
	// create RabbitMQ transportURL CR and get the actual URL from the associated secret that is created
	//
	transportURL, op, err := r.transportURLCreateOrUpdate(instance)
	if err != nil {
		Log.Info("Error getting transportURL. Setting error condition on status and returning")
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
		return ctrl.Result{RequeueAfter: time.Duration(10) * time.Second}, nil
	}

	instance.Status.Conditions.MarkTrue(condition.RabbitMqTransportURLReadyCondition, condition.RabbitMqTransportURLReadyMessage)
	// end transportURL

	configMapVars := make(map[string]env.Setter)

	//
	// check for required OpenStack secret holding passwords for service/admin user and add hash to the vars map
	//
	ctrlResult, err := r.getSecret(ctx, helper, instance, instance.Spec.Aodh.Secret, &configMapVars)
	if err != nil {
		return ctrlResult, err
	}
	// run check OpenStack secret - end

	//
	// Check for required memcached used for caching
	//
	memcached, err := r.getAutoscalingMemcached(ctx, helper, instance)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			instance.Status.Conditions.Set(condition.FalseCondition(
				condition.MemcachedReadyCondition,
				condition.RequestedReason,
				condition.SeverityInfo,
				condition.MemcachedReadyWaitingMessage))
			return ctrl.Result{RequeueAfter: time.Duration(10) * time.Second}, fmt.Errorf("memcached %s not found", instance.Spec.Aodh.MemcachedInstance)
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
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.MemcachedReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.MemcachedReadyWaitingMessage))
		return ctrl.Result{RequeueAfter: time.Duration(10) * time.Second}, fmt.Errorf("memcached %s is not ready", memcached.Name)
	}
	// Mark the Memcached Service as Ready if we get to this point with no errors
	instance.Status.Conditions.MarkTrue(
		condition.MemcachedReadyCondition, condition.MemcachedReadyMessage)
	// run check memcached - end

	//
	// Check for required heat used for autoscaling
	//
	heat, err := r.getAutoscalingHeat(ctx, helper, instance)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			instance.Status.Conditions.Set(condition.FalseCondition(
				telemetryv1.HeatReadyCondition,
				condition.RequestedReason,
				condition.SeverityInfo,
				telemetryv1.HeatReadyNotFoundMessage))
			return ctrl.Result{RequeueAfter: time.Duration(10) * time.Second}, fmt.Errorf("heat %s not found", instance.Spec.HeatInstance)
		}
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.HeatReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			telemetryv1.HeatReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}

	if !heat.IsReady() {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.HeatReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			telemetryv1.HeatReadyUnreadyMessage))
		return ctrl.Result{RequeueAfter: time.Duration(10) * time.Second}, fmt.Errorf("heat %s is not ready", heat.Name)
	}
	// Mark the Heat Service as Ready if we get to this point with no errors
	instance.Status.Conditions.MarkTrue(
		telemetryv1.HeatReadyCondition, condition.ReadyMessage)
	// run check heat - end

	//
	// check for required TransportURL secret holding transport URL string
	//
	ctrlResult, err = r.getSecret(ctx, helper, instance, instance.Status.TransportURLSecret, &configMapVars)
	if err != nil {
		return ctrlResult, err
	}
	// run check TransportURL secret - end

	//
	// Get correct prometheus host and port
	// NOTE: Always do this before calling the generateServiceConfig to get the newest values in the ServiceConfig
	//
	if instance.Spec.PrometheusHost == "" {
		instance.Status.PrometheusHost = fmt.Sprintf("%s-prometheus.%s.svc", telemetryv1.DefaultServiceName, instance.Namespace)
	} else {
		instance.Status.PrometheusHost = instance.Spec.PrometheusHost
	}
	if instance.Spec.PrometheusPort == 0 {
		instance.Status.PrometheusPort = telemetryv1.DefaultPrometheusPort
	} else {
		instance.Status.PrometheusPort = instance.Spec.PrometheusPort
	}

	//
	// create secret required for autoscaling input
	// - %-scripts configmap holding scripts to e.g. bootstrap the service
	// - %-config configmap holding minimal autoscaling config required to get the service up, user can add additional files to be added to the service
	//
	err = r.generateServiceConfig(ctx, helper, instance, &configMapVars, memcached)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.ServiceConfigReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.ServiceConfigReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}

	inputHash, hashChanged, err := r.createHashOfInputHashes(ctx, instance, configMapVars)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.ServiceConfigReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.ServiceConfigReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	} else if hashChanged {
		// Hash changed and instance status should be updated (which will be done by main defer func),
		// so we need to return and reconcile again
		return ctrl.Result{}, nil
	}

	instance.Status.Hash[common.InputHashName] = inputHash

	// Handle service init
	ctrlResult, err = r.reconcileInit(ctx, instance, helper, serviceLabels)
	if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}
	if err != nil {
		return ctrlResult, err
	}
	ctrlResult, err = r.reconcileNormalAodh(ctx, instance, helper, inputHash)
	if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}
	if err != nil {
		return ctrlResult, err
	}
	Log.Info("Reconciled Service successfully")
	return ctrl.Result{}, nil
}

// createHashOfInputHashes - creates a hash of hashes which gets added to the resources which requires a restart
// if any of the input resources change, like configs, passwords, ...
//
// returns the hash, whether the hash changed (as a bool) and any error
func (r *AutoscalingReconciler) createHashOfInputHashes(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
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

func (r *AutoscalingReconciler) generateServiceConfig(
	ctx context.Context,
	h *helper.Helper,
	instance *telemetryv1.Autoscaling,
	envVars *map[string]env.Setter,
	mc *memcachedv1.Memcached,
) error {
	cmLabels := labels.GetLabels(instance, labels.GetGroupLabel(autoscaling.ServiceName), map[string]string{})
	customData := map[string]string{common.CustomServiceConfigFileName: instance.Spec.Aodh.CustomServiceConfig}
	for key, data := range instance.Spec.Aodh.DefaultConfigOverwrite {
		customData[key] = data
	}

	keystoneAPI, err := keystonev1.GetKeystoneAPI(ctx, h, instance.Namespace, map[string]string{})
	if err != nil {
		return err
	}

	keystoneInternalURL, err := keystoneAPI.GetEndpoint(endpoint.EndpointInternal)
	if err != nil {
		return err
	}

	ospSecret, _, err := secret.GetSecret(ctx, h, instance.Spec.Aodh.Secret, instance.Namespace)
	if err != nil {
		return err
	}

	transportURLSecret, _, err := secret.GetSecret(ctx, h, instance.Status.TransportURLSecret, instance.Namespace)
	if err != nil {
		return err
	}

	templateParameters := map[string]interface{}{
		"AodhUser":                 instance.Spec.Aodh.ServiceUser,
		"AodhPassword":             string(ospSecret.Data[instance.Spec.Aodh.PasswordSelectors.AodhService]),
		"KeystoneInternalURL":      keystoneInternalURL,
		"TransportURL":             string(transportURLSecret.Data["transport_url"]),
		"PrometheusHost":           instance.Status.PrometheusHost,
		"PrometheusPort":           instance.Status.PrometheusPort,
		"MemcachedServers":         strings.Join(mc.Status.ServerList, ","),
		"MemcachedServersWithInet": strings.Join(mc.Status.ServerListWithInet, ","),
		"DatabaseConnection": fmt.Sprintf("mysql+pymysql://%s:%s@%s/%s",
			instance.Spec.Aodh.DatabaseUser,
			string(ospSecret.Data[instance.Spec.Aodh.PasswordSelectors.Database]),
			instance.Status.DatabaseHostname,
			autoscaling.DatabaseName),
	}

	cms := []util.Template{
		// ScriptsSecret
		{
			Name:               fmt.Sprintf("%s-scripts", autoscaling.ServiceName),
			Namespace:          instance.Namespace,
			Type:               util.TemplateTypeScripts,
			InstanceType:       instance.Kind,
			AdditionalTemplate: map[string]string{"common.sh": "/common/common.sh"},
			Labels:             cmLabels,
		},
		// Secret
		{
			Name:          fmt.Sprintf("%s-config-data", autoscaling.ServiceName),
			Namespace:     instance.Namespace,
			Type:          util.TemplateTypeConfig,
			InstanceType:  instance.Kind,
			CustomData:    customData,
			ConfigOptions: templateParameters,
			Labels:        cmLabels,
		},
	}
	return secret.EnsureSecrets(ctx, h, instance, cms, envVars)
}

// getAutoscalingMemcached - gets the Memcached instance used for aodh cache backend
func (r *AutoscalingReconciler) getAutoscalingMemcached(
	ctx context.Context,
	h *helper.Helper,
	instance *telemetryv1.Autoscaling,
) (*memcachedv1.Memcached, error) {
	memcached := &memcachedv1.Memcached{}
	err := h.GetClient().Get(
		ctx,
		types.NamespacedName{
			Name:      instance.Spec.Aodh.MemcachedInstance,
			Namespace: instance.Namespace,
		},
		memcached)
	if err != nil {
		return nil, err
	}
	return memcached, err
}

func (r *AutoscalingReconciler) getAutoscalingHeat(
	ctx context.Context,
	h *helper.Helper,
	instance *telemetryv1.Autoscaling,
) (*heatv1.Heat, error) {
	heat := &heatv1.Heat{}
	err := h.GetClient().Get(
		ctx,
		types.NamespacedName{
			Name:      instance.Spec.HeatInstance,
			Namespace: instance.Namespace,
		},
		heat)
	if err != nil {
		return nil, err
	}
	return heat, err
}

// getSecret - get the specified secret, and add its hash to envVars
func (r *AutoscalingReconciler) getSecret(ctx context.Context, h *helper.Helper, instance *telemetryv1.Autoscaling, secretName string, envVars *map[string]env.Setter) (ctrl.Result, error) {
	secret, hash, err := secret.GetSecret(ctx, h, secretName, instance.Namespace)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			instance.Status.Conditions.Set(condition.FalseCondition(
				condition.InputReadyCondition,
				condition.RequestedReason,
				condition.SeverityInfo,
				condition.InputReadyWaitingMessage))
			return ctrl.Result{RequeueAfter: time.Duration(10) * time.Second}, fmt.Errorf("secret %s not found", secretName)
		}
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.InputReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.InputReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}

	// Add a prefix to the var name to avoid accidental collision with other non-secret
	// vars. The secret names themselves will be unique.
	(*envVars)["secret-"+secret.Name] = env.SetValue(hash)

	return ctrl.Result{}, nil
}

func (r *AutoscalingReconciler) transportURLCreateOrUpdate(instance *telemetryv1.Autoscaling) (*rabbitmqv1.TransportURL, controllerutil.OperationResult, error) {
	transportURL := &rabbitmqv1.TransportURL{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-transport", autoscaling.ServiceName),
			Namespace: instance.Namespace,
		},
	}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, transportURL, func() error {
		transportURL.Spec.RabbitmqClusterName = instance.Spec.Aodh.RabbitMqClusterName
		err := controllerutil.SetControllerReference(instance, transportURL, r.Scheme)
		return err
	})
	return transportURL, op, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *AutoscalingReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	// transportURLSecretFn - Watch for changes made to the secret associated with the RabbitMQ
	// TransportURL created and used by Autoscaling CRs.  Watch functions return a list of namespace-scoped
	// CRs that then get fed  to the reconciler.  Hence, in this case, we need to know the name of the
	// Autoscaling CR associated with the secret we are examining in the function.  We could parse the name
	// out of the "%s-transport" secret label, which would be faster than getting the list of
	// the Autoscaling CRs and trying to match on each one.  The downside there, however, is that technically
	// someone could randomly label a secret "something-transport" where "something" actually
	// matches the name of an existing Autoscaling CR.  In that case changes to that secret would trigger
	// reconciliation for a Autoscaling CR that does not need it.
	//
	// TODO: We also need a watch func to monitor for changes to the secret referenced by Autoscaling.Spec.Secret
	Log := r.GetLogger(ctx)
	transportURLSecretFn := func(ctx context.Context, o client.Object) []reconcile.Request {
		result := []reconcile.Request{}

		// get all Autoscaling CRs
		autoscalings := &telemetryv1.AutoscalingList{}
		listOpts := []client.ListOption{
			client.InNamespace(o.GetNamespace()),
		}
		if err := r.Client.List(context.Background(), autoscalings, listOpts...); err != nil {
			Log.Error(err, "Unable to retrieve Autoscaling CRs %v")
			return nil
		}

		for _, ownerRef := range o.GetOwnerReferences() {
			if ownerRef.Kind == "TransportURL" {
				for _, cr := range autoscalings.Items {
					if ownerRef.Name == fmt.Sprintf("%s-transport", cr.Name) {
						// return namespace and Name of CR
						name := client.ObjectKey{
							Namespace: o.GetNamespace(),
							Name:      cr.Name,
						}
						Log.Info(fmt.Sprintf("TransportURL Secret %s belongs to TransportURL belonging to Autoscaling CR %s", o.GetName(), cr.Name))
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
		result := []reconcile.Request{}

		// get all autoscaling CRs
		autoscalings := &telemetryv1.AutoscalingList{}
		listOpts := []client.ListOption{
			client.InNamespace(o.GetNamespace()),
		}
		if err := r.Client.List(context.Background(), autoscalings, listOpts...); err != nil {
			Log.Error(err, "Unable to retrieve Autoscaling CRs %w")
			return nil
		}

		for _, cr := range autoscalings.Items {
			if o.GetName() == cr.Spec.Aodh.MemcachedInstance {
				name := client.ObjectKey{
					Namespace: o.GetNamespace(),
					Name:      cr.Name,
				}
				Log.Info(fmt.Sprintf("Memcached %s is used by Autoscaling CR %s", o.GetName(), cr.Name))
				result = append(result, reconcile.Request{NamespacedName: name})
			}
		}
		if len(result) > 0 {
			return result
		}
		return nil
	}
	// index caBundleSecretNameField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.Autoscaling{}, caBundleSecretNameField, func(rawObj client.Object) []string {
		// Extract the secret name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.Autoscaling)
		if cr.Spec.Aodh.TLS.CaBundleSecretName == "" {
			return nil
		}
		return []string{cr.Spec.Aodh.TLS.CaBundleSecretName}
	}); err != nil {
		return err
	}

	// index tlsAPIInternalField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.Autoscaling{}, tlsAPIInternalField, func(rawObj client.Object) []string {
		// Extract the secret name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.Autoscaling)
		if cr.Spec.Aodh.TLS.API.Internal.SecretName == nil {
			return nil
		}
		return []string{*cr.Spec.Aodh.TLS.API.Internal.SecretName}
	}); err != nil {
		return err
	}

	// index tlsAPIPublicField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.Autoscaling{}, tlsAPIPublicField, func(rawObj client.Object) []string {
		// Extract the secret name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.Autoscaling)
		if cr.Spec.Aodh.TLS.API.Public.SecretName == nil {
			return nil
		}
		return []string{*cr.Spec.Aodh.TLS.API.Public.SecretName}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1.Autoscaling{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&keystonev1.KeystoneService{}).
		Owns(&keystonev1.KeystoneEndpoint{}).
		Owns(&mariadbv1.MariaDBDatabase{}).
		Owns(&mariadbv1.MariaDBAccount{}).
		Owns(&rabbitmqv1.TransportURL{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		// Watch for TransportURL Secrets which belong to any TransportURLs created by Autoscaling CRs
		Watches(&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(transportURLSecretFn)).
		Watches(&memcachedv1.Memcached{},
			handler.EnqueueRequestsFromMapFunc(memcachedFn)).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForSrc),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}

func (r *AutoscalingReconciler) findObjectsForSrc(ctx context.Context, src client.Object) []reconcile.Request {
	requests := []reconcile.Request{}

	l := log.FromContext(context.Background()).WithName("Controllers").WithName("Autoscaling")

	for _, field := range allWatchFields {
		crList := &telemetryv1.AutoscalingList{}
		listOps := &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(field, src.GetName()),
			Namespace:     src.GetNamespace(),
		}
		err := r.Client.List(context.TODO(), crList, listOps)
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
