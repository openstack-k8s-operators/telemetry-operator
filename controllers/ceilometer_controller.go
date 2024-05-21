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
	"sort"
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
	projects "github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	common "github.com/openstack-k8s-operators/lib-common/modules/common"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	endpoint "github.com/openstack-k8s-operators/lib-common/modules/common/endpoint"
	env "github.com/openstack-k8s-operators/lib-common/modules/common/env"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	labels "github.com/openstack-k8s-operators/lib-common/modules/common/labels"
	common_rbac "github.com/openstack-k8s-operators/lib-common/modules/common/rbac"
	rolebinding "github.com/openstack-k8s-operators/lib-common/modules/common/rolebinding"
	secret "github.com/openstack-k8s-operators/lib-common/modules/common/secret"
	serviceaccount "github.com/openstack-k8s-operators/lib-common/modules/common/serviceaccount"
	statefulset "github.com/openstack-k8s-operators/lib-common/modules/common/statefulset"
	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	util "github.com/openstack-k8s-operators/lib-common/modules/common/util"
	openstack "github.com/openstack-k8s-operators/lib-common/modules/openstack"

	rabbitmqv1 "github.com/openstack-k8s-operators/infra-operator/apis/rabbitmq/v1beta1"
	keystonev1 "github.com/openstack-k8s-operators/keystone-operator/api/v1beta1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	availability "github.com/openstack-k8s-operators/telemetry-operator/pkg/availability"
	ceilometer "github.com/openstack-k8s-operators/telemetry-operator/pkg/ceilometer"
)

const (
	msgOperation        = "Service '%s' successfully changed - operation: %s"
	msgReconcileStart   = "Reconciling Service '%s'"
	msgReconcileSuccess = "Reconciled Service '%s' successfully"
)

// CeilometerReconciler reconciles a Ceilometer object
type CeilometerReconciler struct {
	client.Client
	Kclient kubernetes.Interface
	Scheme  *runtime.Scheme
}

// GetLogger returns a logger object with a prefix of "conroller.name" and aditional controller context fields
func (r *CeilometerReconciler) GetLogger(ctx context.Context) logr.Logger {
	return log.FromContext(ctx).WithName("Controllers").WithName("Ceilometer")
}

// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometers/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete;
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete;
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete;
// +kubebuilder:rbac:groups=keystone.openstack.org,resources=keystoneservices,verbs=get;list;watch;create;update;patch;delete;
// +kubebuilder:rbac:groups=rabbitmq.openstack.org,resources=transporturls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=keystone.openstack.org,resources=keystoneapis,verbs=get;list;watch;
// service account, role, rolebinding
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;list;watch;create;update
// service account permissions that are needed to grant permission to the above
// +kubebuilder:rbac:groups="security.openshift.io",resourceNames=anyuid,resources=securitycontextconstraints,verbs=use

// Reconcile reconciles a Ceilometer
func (r *CeilometerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, _err error) {
	Log := r.GetLogger(ctx)

	// Fetch the Ceilometer instance
	instance := &telemetryv1.Ceilometer{}
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
	isNewInstance := instance.CeilometerStatus.Conditions == nil && instance.KSMStatus.Conditions == nil
	if isNewInstance {
		instance.CeilometerStatus.Conditions = condition.Conditions{}
		instance.KSMStatus.Conditions = condition.Conditions{}
	}

	// Save a copy of the conditions so that we can restore the LastTransitionTime
	// when a condition's state doesn't change.
	savedConditions := instance.CeilometerStatus.Conditions.DeepCopy()

	// Always patch the instance status when exiting this function so we can
	// persist any changes.
	defer func() {
		condition.RestoreLastTransitionTimes(
			&instance.CeilometerStatus.Conditions, savedConditions)
		if instance.CeilometerStatus.Conditions.IsUnknown(condition.ReadyCondition) {
			instance.CeilometerStatus.Conditions.Set(
				instance.CeilometerStatus.Conditions.Mirror(condition.ReadyCondition))
		}

		// update the Ready condition based on the sub conditions
		if instance.KSMStatus.Conditions.AllSubConditionIsTrue() {
			instance.KSMStatus.Conditions.MarkTrue(
				condition.ReadyCondition, condition.ReadyMessage)
		} else {
			// something is not ready so reset the Ready condition
			instance.KSMStatus.Conditions.MarkUnknown(
				condition.ReadyCondition, condition.InitReason, condition.ReadyInitMessage)
			// and recalculate it based on the state of the rest of the conditions
			instance.KSMStatus.Conditions.Set(
				instance.KSMStatus.Conditions.Mirror(condition.ReadyCondition))
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
	// initialize conditions used later as Status=Unknown
	cl := condition.CreateList(
		condition.UnknownCondition(condition.ReadyCondition, condition.InitReason, condition.ReadyInitMessage),
		condition.UnknownCondition(condition.InputReadyCondition, condition.InitReason, condition.InputReadyInitMessage),
		condition.UnknownCondition(condition.ServiceConfigReadyCondition, condition.InitReason, condition.ServiceConfigReadyInitMessage),
		condition.UnknownCondition(condition.DeploymentReadyCondition, condition.InitReason, condition.DeploymentReadyInitMessage),
		// right now we have no dedicated KeystoneServiceReadyInitMessage
		condition.UnknownCondition(condition.KeystoneServiceReadyCondition, condition.InitReason, ""),
		condition.UnknownCondition(condition.TLSInputReadyCondition, condition.InitReason, condition.InputReadyInitMessage),
	)
	instance.CeilometerStatus.Conditions.Init(&cl)
	instance.CeilometerStatus.ObservedGeneration = instance.Generation

	if instance.CeilometerStatus.Hash == nil {
		instance.CeilometerStatus.Hash = map[string]string{}
	}

	cl = condition.CreateList(
		condition.UnknownCondition(condition.ReadyCondition, condition.InitReason, condition.ReadyInitMessage),
		condition.UnknownCondition(condition.InputReadyCondition, condition.InitReason, condition.InputReadyInitMessage),
		condition.UnknownCondition(condition.ServiceConfigReadyCondition, condition.InitReason, condition.ServiceConfigReadyInitMessage),
		condition.UnknownCondition(condition.DeploymentReadyCondition, condition.InitReason, condition.DeploymentReadyInitMessage),
		// service account, role, rolebinding conditions
		condition.UnknownCondition(condition.ServiceAccountReadyCondition, condition.InitReason, condition.ServiceAccountReadyInitMessage),
		condition.UnknownCondition(condition.RoleReadyCondition, condition.InitReason, condition.RoleReadyInitMessage),
		condition.UnknownCondition(condition.RoleBindingReadyCondition, condition.InitReason, condition.RoleBindingReadyInitMessage),
	)
	instance.KSMStatus.Conditions.Init(&cl)
	instance.KSMStatus.ObservedGeneration = instance.Generation

	if instance.KSMStatus.Hash == nil {
		instance.KSMStatus.Hash = map[string]string{}
	}

	// If we're not deleting this and the service object doesn't have our finalizer, add it.
	if !instance.DeletionTimestamp.IsZero() && controllerutil.AddFinalizer(instance, helper.GetFinalizer()) || isNewInstance {
		return r.reconcileDelete(ctx, instance, helper)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, instance, helper)
}

// fields to index to reconcile when change
const (
	ceilometerPasswordSecretField     = ".spec.secret"
	ceilometerCaBundleSecretNameField = ".spec.tls.caBundleSecretName"
	ceilometerTlsField                = ".spec.tls.secretName"
	ksmCaBundleSecretNameField        = ".spec.ksmTls.caBundleSecretName"
	ksmTlsField                       = ".spec.ksmTls.secretName"
)

var (
	ceilometerWatchFields = []string{
		ceilometerPasswordSecretField,
		ceilometerCaBundleSecretNameField,
		ceilometerTlsField,
		ksmCaBundleSecretNameField,
		ksmTlsField,
	}
)

func (r *CeilometerReconciler) reconcileDelete(ctx context.Context, instance *telemetryv1.Ceilometer, helper *helper.Helper) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info("Reconciling Service delete")

	// Remove the finalizer from our KeystoneService CR
	keystoneService, err := keystonev1.GetKeystoneServiceWithName(ctx, helper, ceilometer.ServiceName, instance.Namespace)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if err == nil {
		if controllerutil.RemoveFinalizer(keystoneService, helper.GetFinalizer()) {
			err = r.Update(ctx, keystoneService)
			if err != nil && !k8s_errors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
			util.LogForObject(helper, "Removed finalizer from our KeystoneService", instance)
		}
	}

	// Service is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(instance, helper.GetFinalizer())
	Log.Info(fmt.Sprintf("Reconciled Service '%s' delete successfully", ceilometer.ServiceName))

	return ctrl.Result{}, nil
}

func (r *CeilometerReconciler) reconcileInit(
	ctx context.Context,
	instance *telemetryv1.Ceilometer,
	helper *helper.Helper,
	serviceLabels map[string]string,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info("Reconciling Service init")

	//
	// create Keystone service and users
	//
	_, _, err := secret.GetSecret(ctx, helper, instance.Spec.Secret, instance.Namespace)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: time.Duration(10) * time.Second}, fmt.Errorf("OpenStack secret %s not found", instance.Spec.Secret)
		}
		return ctrl.Result{}, err
	}

	ksSvcSpec := keystonev1.KeystoneServiceSpec{
		ServiceType:        ceilometer.ServiceType,
		ServiceName:        ceilometer.ServiceName,
		ServiceDescription: "Ceilometer Service",
		Enabled:            true,
		ServiceUser:        instance.Spec.ServiceUser,
		Secret:             instance.Spec.Secret,
		PasswordSelector:   instance.Spec.PasswordSelectors.CeilometerService,
	}

	ksSvc := keystonev1.NewKeystoneService(ksSvcSpec, instance.Namespace, serviceLabels, 10)
	ctrlResult, err := ksSvc.CreateOrPatch(ctx, helper)
	if err != nil {
		return ctrlResult, err
	}

	// mirror the Status, Reason, Severity and Message of the latest keystoneservice condition
	// into a local condition with the type condition.KeystoneServiceReadyCondition
	c := ksSvc.GetConditions().Mirror(condition.KeystoneServiceReadyCondition)
	if c != nil {
		instance.CeilometerStatus.Conditions.Set(c)
	}

	if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	if instance.CeilometerStatus.Hash == nil {
		instance.CeilometerStatus.Hash = map[string]string{}
	}

	if instance.KSMStatus.Hash == nil {
		instance.KSMStatus.Hash = map[string]string{}
	}

	Log.Info("Reconciled Service init successfully")
	return ctrl.Result{}, nil
}

func (r *CeilometerReconciler) reconcileNormal(ctx context.Context, instance *telemetryv1.Ceilometer, helper *helper.Helper) (ctrl.Result, error) {
	// ConfigMap
	configMapVars := make(map[string]env.Setter)

	ksmRes, err := r.reconcileKSM(ctx, instance, helper, &configMapVars)
	if err != nil {
		return ksmRes, err
	}

	ceilRes, err := r.reconcileCeilometer(ctx, instance, helper, &configMapVars)
	if err != nil {
		return ceilRes, err
	}

	return ceilRes, nil
}

func (r *CeilometerReconciler) reconcileCeilometer(
	ctx context.Context,
	instance *telemetryv1.Ceilometer,
	helper *helper.Helper,
	configMapVars *map[string]env.Setter,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info(fmt.Sprintf(msgReconcileStart, ceilometer.ServiceName))

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

	//
	// create RabbitMQ transportURL CR and get the actual URL from the associated secret that is created
	//
	transportURL, op, err := r.transportURLCreateOrUpdate(instance)
	if err != nil {
		Log.Info("Error getting transportURL. Setting error condition on status and returning")
		instance.CeilometerStatus.Conditions.Set(condition.FalseCondition(
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

	instance.CeilometerStatus.TransportURLSecret = transportURL.Status.SecretName

	if instance.CeilometerStatus.TransportURLSecret == "" {
		Log.Info(fmt.Sprintf("Waiting for TransportURL %s secret to be created", transportURL.Name))
		instance.CeilometerStatus.Conditions.Set(condition.FalseCondition(
			condition.RabbitMqTransportURLReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.RabbitMqTransportURLReadyRunningMessage))
		return ctrl.Result{RequeueAfter: time.Duration(10) * time.Second}, nil
	}

	instance.CeilometerStatus.Conditions.MarkTrue(condition.RabbitMqTransportURLReadyCondition, condition.RabbitMqTransportURLReadyMessage)
	// end transportURL

	//
	// check for required OpenStack secret holding passwords for service/admin user and add hash to the vars map
	//
	ctrlResult, err := r.getSecret(ctx, helper, instance, instance.Spec.Secret, configMapVars)
	if err != nil {
		return ctrlResult, err
	}
	// run check OpenStack secret - end

	//
	// check for required TransportURL secret holding transport URL string
	//
	ctrlResult, err = r.getSecret(ctx, helper, instance, instance.CeilometerStatus.TransportURLSecret, configMapVars)
	if err != nil {
		return ctrlResult, err
	}
	// run check TransportURL secret - end

	instance.CeilometerStatus.Conditions.MarkTrue(condition.InputReadyCondition, condition.InputReadyMessage)

	//
	// TLS input validation
	//
	// Validate the CA cert secret if provided
	if instance.Spec.TLS.CaBundleSecretName != "" {
		hash, ctrlResult, err := tls.ValidateCACertSecret(
			ctx,
			helper.GetClient(),
			types.NamespacedName{
				Name:      instance.Spec.TLS.CaBundleSecretName,
				Namespace: instance.Namespace,
			},
		)
		if err != nil {
			instance.CeilometerStatus.Conditions.Set(condition.FalseCondition(
				condition.TLSInputReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.TLSInputErrorMessage,
				err.Error()))
			return ctrlResult, err
		} else if (ctrlResult != ctrl.Result{}) {
			return ctrlResult, nil
		}

		if hash != "" {
			(*configMapVars)[tls.CABundleKey] = env.SetValue(hash)
		}
	}

	// Validate metadata service cert secret
	if instance.Spec.TLS.Enabled() {
		hash, ctrlResult, err := instance.Spec.TLS.ValidateCertSecret(ctx, helper, instance.Namespace)
		if err != nil {
			instance.CeilometerStatus.Conditions.Set(condition.FalseCondition(
				condition.TLSInputReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.TLSInputErrorMessage,
				err.Error()))
			return ctrl.Result{}, err
		} else if (ctrlResult != ctrl.Result{}) {
			return ctrlResult, nil
		}
		(*configMapVars)[tls.TLSHashName] = env.SetValue(hash)
	}
	// all cert input checks out so report InputReady
	instance.CeilometerStatus.Conditions.MarkTrue(condition.TLSInputReadyCondition, condition.InputReadyMessage)

	//
	// create Configmap required for ceilometer input
	// - %-scripts configmap holding scripts to e.g. bootstrap the service
	// - %-config configmap holding minimal ceilometer config required to get the service up, user can add additional files to be added to the service
	// - parameters which has passwords gets added from the OpenStack secret via the init container
	//
	err = r.generateServiceConfig(ctx, helper, instance, configMapVars)
	if err != nil {
		instance.CeilometerStatus.Conditions.Set(condition.FalseCondition(
			condition.ServiceConfigReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.ServiceConfigReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}

	//
	// create Configmap required for ceilometer-compute input
	// - %-scripts configmap holding scripts to e.g. bootstrap the service
	// - %-config configmap holding minimal ceilometer-compute config required to get the service up, user can add additional files to be added to the service
	// - parameters which has passwords gets added from the OpenStack secret via the init container
	//
	err = r.generateComputeServiceConfig(ctx, helper, instance, configMapVars)
	if err != nil {
		instance.CeilometerStatus.Conditions.Set(condition.FalseCondition(
			condition.ServiceConfigReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.ServiceConfigReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}

	//
	// Swift Role for storage.* metrics
	//
	if r.roleExists(ctx, helper, instance, "SwiftSystemReader", Log) {
		err = r.ensureSwiftRole(ctx, helper, instance, Log)
		if err != nil {
			Log.Info("SwiftSystemReader cannot be assigned to Ceilometer. No storage metrics will be available.")
		}
	}

	// Hash all the endpointurls to trigger a redeployment everytime one of the internal endpoints changes or is added
	v := "internal"
	endpointurls, err := keystonev1.GetKeystoneEndpointUrls(ctx, helper, instance.Namespace, &v)
	sort.Strings(endpointurls)

	if err != nil {

		return ctrl.Result{}, err
	}
	hash, err := util.ObjectHash(endpointurls)
	if err != nil {
		return ctrl.Result{}, err
	}
	(*configMapVars)["endpointurls"] = env.SetValue(hash)

	//
	// create hash over all the different input resources to identify if any those changed
	// and a restart/recreate is required.
	//
	inputHash, hashChanged, err := r.createHashOfInputHashes(ctx, instance, *configMapVars)
	if err != nil {
		instance.CeilometerStatus.Conditions.Set(condition.FalseCondition(
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

	instance.CeilometerStatus.Hash[common.InputHashName] = inputHash

	instance.CeilometerStatus.Conditions.MarkTrue(condition.ServiceConfigReadyCondition, condition.ServiceConfigReadyMessage)

	serviceLabels := map[string]string{
		common.AppSelector:   ceilometer.ServiceName,
		common.OwnerSelector: instance.Name,
	}

	// Handle service init
	ctrlResult, err = r.reconcileInit(ctx, instance, helper, serviceLabels)
	if err != nil {
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	// Define a new StatefulSet object
	sfsetDef, err := ceilometer.StatefulSet(instance, inputHash, serviceLabels)
	if err != nil {
		return ctrl.Result{}, err
	}
	sfset := statefulset.NewStatefulSet(
		sfsetDef,
		time.Duration(5)*time.Second,
	)

	ctrlResult, err = sfset.CreateOrPatch(ctx, helper)
	if err != nil {
		instance.CeilometerStatus.Conditions.Set(condition.FalseCondition(
			condition.DeploymentReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.DeploymentReadyErrorMessage,
			err.Error()))
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		instance.CeilometerStatus.Conditions.Set(condition.FalseCondition(
			condition.DeploymentReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.DeploymentReadyRunningMessage))
		return ctrlResult, nil
	}

	err = controllerutil.SetControllerReference(instance, sfsetDef, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Evaluate the last part of the reconciliation only if we see the last
	// version of the CR
	if sfset.GetStatefulSet().Generation == sfset.GetStatefulSet().Status.ObservedGeneration {
		instance.CeilometerStatus.ReadyCount = sfset.GetStatefulSet().Status.ReadyReplicas
		instance.CeilometerStatus.Networks = instance.Spec.NetworkAttachmentDefinitions
		svc, op, err := ceilometer.Service(instance, helper, ceilometer.CeilometerPrometheusPort, serviceLabels)
		if err != nil {
			return ctrl.Result{}, err
		}
		if op != controllerutil.OperationResultNone {
			Log.Info(fmt.Sprintf(msgOperation, svc.Name, string(op)))
		}
		if instance.CeilometerStatus.ReadyCount > 0 {
			instance.CeilometerStatus.Conditions.MarkTrue(condition.DeploymentReadyCondition, condition.DeploymentReadyMessage)
		}
		if instance.CeilometerStatus.Conditions.AllSubConditionIsTrue() {
			instance.CeilometerStatus.Conditions.MarkTrue(
				condition.ReadyCondition, condition.ReadyMessage)
		}
		Log.Info(fmt.Sprintf(msgReconcileSuccess, ceilometer.ServiceName))
	}
	return ctrl.Result{}, nil
}

func (r *CeilometerReconciler) reconcileKSM(
	ctx context.Context,
	instance *telemetryv1.Ceilometer,
	helper *helper.Helper,
	configMapVars *map[string]env.Setter,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info(fmt.Sprintf(msgReconcileStart, availability.KSMServiceName))

	// create service account and role binding for the KSM service
	sa, rb := availability.KSMServiceAccount(instance)

	svcacc := serviceaccount.NewServiceAccount(sa, time.Duration(5)*time.Second)
	ctrlResult, err := svcacc.CreateOrPatch(ctx, helper)
	if err != nil {
		instance.KSMStatus.Conditions.Set(condition.FalseCondition(
			condition.ServiceAccountReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.ServiceAccountReadyErrorMessage,
			err.Error()))
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		instance.KSMStatus.Conditions.Set(condition.FalseCondition(
			condition.ServiceAccountReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.ServiceAccountReadyMessage))
		return ctrlResult, nil
	}

	err = controllerutil.SetControllerReference(instance, sa, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	roleBind := rolebinding.NewRoleBinding(rb, time.Duration(5)*time.Second)
	ctrlResult, err = roleBind.CreateOrPatch(ctx, helper)
	if err != nil {
		instance.KSMStatus.Conditions.Set(condition.FalseCondition(
			condition.RoleBindingReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.ServiceAccountReadyErrorMessage,
			err.Error()))
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		instance.KSMStatus.Conditions.Set(condition.FalseCondition(
			condition.RoleBindingReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.RoleBindingReadyMessage))
		return ctrlResult, nil
	}

	err = controllerutil.SetControllerReference(instance, rb, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	serviceLabels := map[string]string{
		common.AppSelector: availability.KSMServiceName,
	}

	// Validate the CA cert secret if provided
	if instance.Spec.KSMTLS.CaBundleSecretName != "" {
		hash, ctrlResult, err := tls.ValidateCACertSecret(
			ctx,
			helper.GetClient(),
			types.NamespacedName{
				Name:      instance.Spec.KSMTLS.CaBundleSecretName,
				Namespace: instance.Namespace,
			},
		)
		if err != nil {
			instance.KSMStatus.Conditions.Set(condition.FalseCondition(
				condition.TLSInputReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.TLSInputErrorMessage,
				err.Error()))
			return ctrlResult, err
		} else if (ctrlResult != ctrl.Result{}) {
			return ctrlResult, nil
		}

		if hash != "" {
			(*configMapVars)[fmt.Sprintf("ksm-%s", tls.CABundleKey)] = env.SetValue(hash)
		}
	}

	// Validate metadata service cert secret
	if instance.Spec.KSMTLS.Enabled() {
		hash, ctrlResult, err := instance.Spec.KSMTLS.ValidateCertSecret(ctx, helper, instance.Namespace)
		if err != nil {
			instance.KSMStatus.Conditions.Set(condition.FalseCondition(
				condition.TLSInputReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.TLSInputErrorMessage,
				err.Error()))
			return ctrl.Result{}, err
		} else if (ctrlResult != ctrl.Result{}) {
			return ctrlResult, nil
		}
		(*configMapVars)[fmt.Sprintf("ksm-%s", tls.TLSHashName)] = env.SetValue(hash)
	}

	tlsConfName := ""
	if instance.Spec.KSMTLS.Enabled() {
		tlsConfDef := availability.KSMTLSConfig(instance, serviceLabels, true)
		tlsConfName = tlsConfDef.Name

		hash, op, err := secret.CreateOrPatchSecret(ctx, helper, instance, tlsConfDef)
		if err != nil {
			return ctrl.Result{}, err
		}
		if op != controllerutil.OperationResultNone {
			Log.Info(fmt.Sprintf("KSM TLS config %s successfully changed - operation: %s", tlsConfDef.Name, string(op)))
		}
		(*configMapVars)[tlsConfDef.Name] = env.SetValue(hash)
	}

	// create the service
	ssDef, err := availability.KSMStatefulSet(instance, tlsConfName, serviceLabels)
	if err != nil {
		return ctrl.Result{}, err
	}

	ss := statefulset.NewStatefulSet(ssDef, time.Duration(5)*time.Second)
	ctrlResult, err = ss.CreateOrPatch(ctx, helper)
	if err != nil {
		instance.KSMStatus.Conditions.Set(condition.FalseCondition(
			condition.DeploymentReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.DeploymentReadyErrorMessage,
			err.Error()))
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		instance.KSMStatus.Conditions.Set(condition.FalseCondition(
			condition.DeploymentReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.DeploymentReadyRunningMessage))
		return ctrlResult, nil
	}

	err = controllerutil.SetControllerReference(instance, ssDef, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Evaluate the last part of the reconciliation only if we see the last
	// version of the CR
	ssobj := ss.GetStatefulSet()
	if ssobj.Generation == ssobj.Status.ObservedGeneration {
		instance.KSMStatus.ReadyCount = ss.GetStatefulSet().Status.ReadyReplicas
		if instance.KSMStatus.ReadyCount > 0 {
			instance.KSMStatus.Conditions.MarkTrue(condition.DeploymentReadyCondition, condition.DeploymentReadyMessage)
		}

		// Create the service
		svc, op, err := availability.KSMService(instance, helper, serviceLabels)
		if err != nil {
			instance.KSMStatus.Conditions.Set(condition.FalseCondition(
				condition.ExposeServiceReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.ExposeServiceReadyErrorMessage,
				err.Error()))

			return ctrl.Result{}, err
		}
		if op != controllerutil.OperationResultNone {
			Log.Info(fmt.Sprintf(msgOperation, svc.Name, string(op)))
		}

		if instance.CeilometerStatus.Conditions.AllSubConditionIsTrue() {
			instance.CeilometerStatus.Conditions.MarkTrue(
				condition.ReadyCondition, condition.ReadyMessage)
		}
		Log.Info(fmt.Sprintf(msgReconcileSuccess, availability.KSMServiceName))
	}

	return ctrl.Result{}, nil
}

// getSecret - get the specified secret, and add its hash to envVars
func (r *CeilometerReconciler) getSecret(ctx context.Context, h *helper.Helper, instance *telemetryv1.Ceilometer, secretName string, envVars *map[string]env.Setter) (ctrl.Result, error) {
	secret, hash, err := secret.GetSecret(ctx, h, secretName, instance.Namespace)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			instance.CeilometerStatus.Conditions.Set(condition.FalseCondition(
				condition.InputReadyCondition,
				condition.RequestedReason,
				condition.SeverityInfo,
				condition.InputReadyWaitingMessage))
			return ctrl.Result{RequeueAfter: time.Duration(10) * time.Second}, fmt.Errorf("secret %s not found", secretName)
		}
		instance.CeilometerStatus.Conditions.Set(condition.FalseCondition(
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

func (r *CeilometerReconciler) generateServiceConfig(
	ctx context.Context,
	h *helper.Helper,
	instance *telemetryv1.Ceilometer,
	envVars *map[string]env.Setter,
) error {
	cmLabels := labels.GetLabels(instance, labels.GetGroupLabel(ceilometer.ServiceName), map[string]string{})
	customData := map[string]string{common.CustomServiceConfigFileName: instance.Spec.CustomServiceConfig}
	for key, data := range instance.Spec.DefaultConfigOverwrite {
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

	transportURLSecret, _, _ := secret.GetSecret(ctx, h, instance.CeilometerStatus.TransportURLSecret, instance.Namespace)
	ceilometerPasswordSecret, _, _ := secret.GetSecret(ctx, h, instance.Spec.Secret, instance.Namespace)

	templateParameters := map[string]interface{}{
		"KeystoneInternalURL": keystoneInternalURL,
		"TransportURL":        string(transportURLSecret.Data["transport_url"]),
		"CeilometerPassword":  string(ceilometerPasswordSecret.Data["CeilometerPassword"]),
		"TLS":                 false, // Default to false. Change to true later if TLS enabled
		"SwiftRole":           false, //
	}

	// create httpd  vhost template parameters
	endptConfig := map[string]interface{}{}
	endptConfig["ServerName"] = fmt.Sprintf("%s-internal.%s.svc", ceilometer.ServiceName, instance.Namespace)
	if instance.Spec.TLS.Enabled() {
		templateParameters["TLS"] = true
		templateParameters["CAFile"] = tls.DownstreamTLSCABundlePath
		endptConfig["SSLCertificateFile"] = fmt.Sprintf("/etc/pki/tls/certs/%s", tls.CertKey)
		endptConfig["SSLCertificateKeyFile"] = fmt.Sprintf("/etc/pki/tls/private/%s", tls.PrivateKey)
	}
	templateParameters["vhost"] = endptConfig

	// If the Swift user exists, we add it to the polling
	if r.roleExists(ctx, h, instance, "SwiftSystemReader", h.GetLogger()) {
		templateParameters["SwiftRole"] = true
	}

	cms := []util.Template{
		// ScriptsSecrets
		{
			Name:               fmt.Sprintf("%s-scripts", ceilometer.ServiceName),
			Namespace:          instance.Namespace,
			Type:               util.TemplateTypeScripts,
			InstanceType:       "ceilometercentral",
			AdditionalTemplate: map[string]string{"common.sh": "/common/common.sh"},
			Labels:             cmLabels,
		},
		// Secrets
		{
			Name:          fmt.Sprintf("%s-config-data", ceilometer.ServiceName),
			Namespace:     instance.Namespace,
			Type:          util.TemplateTypeConfig,
			InstanceType:  "ceilometercentral",
			CustomData:    customData,
			ConfigOptions: templateParameters,
			Labels:        cmLabels,
		},
	}
	return secret.EnsureSecrets(ctx, h, instance, cms, envVars)
}

func (r *CeilometerReconciler) generateComputeServiceConfig(
	ctx context.Context,
	h *helper.Helper,
	instance *telemetryv1.Ceilometer,
	envVars *map[string]env.Setter,
) error {
	cmLabels := labels.GetLabels(instance, labels.GetGroupLabel(ceilometer.ComputeServiceName), map[string]string{})
	customData := map[string]string{common.CustomServiceConfigFileName: instance.Spec.CustomServiceConfig}
	for key, data := range instance.Spec.DefaultConfigOverwrite {
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

	transportURLSecret, _, _ := secret.GetSecret(ctx, h, instance.CeilometerStatus.TransportURLSecret, instance.Namespace)
	ceilometerPasswordSecret, _, _ := secret.GetSecret(ctx, h, instance.Spec.Secret, instance.Namespace)

	templateParameters := map[string]interface{}{
		"KeystoneInternalURL":      keystoneInternalURL,
		"TransportURL":             string(transportURLSecret.Data["transport_url"]),
		"CeilometerPassword":       string(ceilometerPasswordSecret.Data["CeilometerPassword"]),
		"ceilometer_compute_image": instance.Spec.ComputeImage,
		"ceilometer_ipmi_image":    instance.Spec.IpmiImage,
	}

	cms := []util.Template{
		// ScriptsConfigMap
		{
			Name:               fmt.Sprintf("%s-scripts", ceilometer.ComputeServiceName),
			Namespace:          instance.Namespace,
			Type:               util.TemplateTypeScripts,
			InstanceType:       "ceilometercompute",
			AdditionalTemplate: map[string]string{"common.sh": "/common/common.sh"},
			Labels:             cmLabels,
		},
		// ConfigMap
		{
			Name:          fmt.Sprintf("%s-config-data", ceilometer.ComputeServiceName),
			Namespace:     instance.Namespace,
			Type:          util.TemplateTypeConfig,
			InstanceType:  "ceilometercompute",
			CustomData:    customData,
			ConfigOptions: templateParameters,
			Labels:        cmLabels,
		},
	}
	return secret.EnsureSecrets(ctx, h, instance, cms, envVars)
}

// createHashOfInputHashes - creates a hash of hashes which gets added to the resources which requires a restart
// if any of the input resources change, like configs, passwords, ...
//
// returns the hash, whether the hash changed (as a bool) and any error
func (r *CeilometerReconciler) createHashOfInputHashes(
	ctx context.Context,
	instance *telemetryv1.Ceilometer,
	envVars map[string]env.Setter,
) (string, bool, error) {
	Log := r.GetLogger(ctx)
	changed := false
	hash, err := util.HashOfInputHashes(envVars)
	if err != nil {
		return hash, changed, err
	}
	if hashMap, changed := util.SetHash(instance.CeilometerStatus.Hash, common.InputHashName, hash); changed {
		instance.CeilometerStatus.Hash = hashMap
		Log.Info(fmt.Sprintf("Input maps hash %s - %s", common.InputHashName, hash))
	}
	return hash, changed, nil
}

func (r *CeilometerReconciler) transportURLCreateOrUpdate(instance *telemetryv1.Ceilometer) (*rabbitmqv1.TransportURL, controllerutil.OperationResult, error) {
	transportURL := &rabbitmqv1.TransportURL{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-transport", ceilometer.ServiceName),
			Namespace: instance.Namespace,
		},
	}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, transportURL, func() error {
		transportURL.Spec.RabbitmqClusterName = instance.Spec.RabbitMqClusterName
		err := controllerutil.SetControllerReference(instance, transportURL, r.Scheme)
		return err
	})
	return transportURL, op, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *CeilometerReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	// transportURLSecretFn - Watch for changes made to the secret associated with the RabbitMQ
	// TransportURL created and used by Ceilometer CRs.  Watch functions return a list of namespace-scoped
	// CRs that then get fed  to the reconciler.  Hence, in this case, we need to know the name of the
	// Ceilometer CR associated with the secret we are examining in the function.  We could parse the name
	// out of the "%s-transport" secret label, which would be faster than getting the list of
	// the Ceilometer CRs and trying to match on each one.  The downside there, however, is that technically
	// someone could randomly label a secret "something-transport" where "something" actually
	// matches the name of an existing Ceilometer CR.  In that case changes to that secret would trigger
	// reconciliation for a Ceilometer CR that does not need it.
	//
	// TODO: We also need a watch func to monitor for changes to the secret referenced by Ceilometer.Spec.Secret
	Log := r.GetLogger(ctx)
	transportURLSecretFn := func(ctx context.Context, o client.Object) []reconcile.Request {
		result := []reconcile.Request{}

		// get all Ceilometer CRs
		ceilometers := &telemetryv1.CeilometerList{}
		listOpts := []client.ListOption{
			client.InNamespace(o.GetNamespace()),
		}
		if err := r.Client.List(context.Background(), ceilometers, listOpts...); err != nil {
			Log.Error(err, "Unable to retrieve Ceilometer CRs %v")
			return nil
		}

		for _, ownerRef := range o.GetOwnerReferences() {
			if ownerRef.Kind == "TransportURL" {
				for _, cr := range ceilometers.Items {
					if ownerRef.Name == fmt.Sprintf("%s-transport", cr.Name) {
						// return namespace and Name of CR
						name := client.ObjectKey{
							Namespace: o.GetNamespace(),
							Name:      cr.Name,
						}
						Log.Info(fmt.Sprintf("TransportURL Secret %s belongs to TransportURL belonging to Ceilometer CR %s", o.GetName(), cr.Name))
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

	// Reconcile every time a keystoneendpoint is modified
	keystoneEndpointsWatchFn := func(ctx context.Context, o client.Object) []reconcile.Request {
		result := []reconcile.Request{}
		name := client.ObjectKey{
			Namespace: o.GetNamespace(),
			Name:      ceilometer.ServiceName,
		}
		result = append(result, reconcile.Request{NamespacedName: name})
		return result
	}

	// index ceilometerPasswordSecretField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.Ceilometer{}, ceilometerPasswordSecretField, func(rawObj client.Object) []string {
		// Extract the secret name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.Ceilometer)
		if cr.Spec.Secret == "" {
			return nil
		}
		return []string{cr.Spec.Secret}
	}); err != nil {
		return err
	}

	// index ceilometerCaBundleSecretNameField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.Ceilometer{}, ceilometerCaBundleSecretNameField, func(rawObj client.Object) []string {
		// Extract the secret name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.Ceilometer)
		if cr.Spec.TLS.CaBundleSecretName == "" {
			return nil
		}
		return []string{cr.Spec.TLS.CaBundleSecretName}
	}); err != nil {
		return err
	}

	// index ceilometerTlsField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.Ceilometer{}, ceilometerTlsField, func(rawObj client.Object) []string {
		// Extract the secret name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.Ceilometer)
		if cr.Spec.TLS.SecretName == nil {
			return nil
		}
		return []string{*cr.Spec.TLS.SecretName}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1.Ceilometer{}).
		Owns(&keystonev1.KeystoneService{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Owns(&rabbitmqv1.TransportURL{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		// Watch for TransportURL Secrets which belong to any TransportURLs created by Ceilometer CRs
		Watches(&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(transportURLSecretFn)).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForSrc),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Watches(
			&keystonev1.KeystoneEndpoint{},
			handler.EnqueueRequestsFromMapFunc(keystoneEndpointsWatchFn),
		).
		Complete(r)
}

func (r *CeilometerReconciler) findObjectsForSrc(ctx context.Context, src client.Object) []reconcile.Request {
	requests := []reconcile.Request{}

	l := log.FromContext(ctx).WithName("Controllers").WithName("Ceilometer")

	for _, field := range ceilometerWatchFields {
		crList := &telemetryv1.CeilometerList{}
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

func (r CeilometerReconciler) roleExists(
	ctx context.Context,
	helper *helper.Helper,
	instance *telemetryv1.Ceilometer,
	roleName string,
	log logr.Logger,
) bool {
	keystoneAPI, err := keystonev1.GetKeystoneAPI(ctx, helper, instance.Namespace, map[string]string{})
	if err != nil {
		return false
	}

	os, ctrlResult, err := keystonev1.GetAdminServiceClient(
		ctx,
		helper,
		keystoneAPI,
	)
	if err != nil || (ctrlResult != ctrl.Result{}) {
		return false
	}

	role, err := os.GetRole(log, roleName)
	if err != nil || role == nil {
		return false
	}
	return true
}

func (r *CeilometerReconciler) ensureSwiftRole(
	ctx context.Context,
	helper *helper.Helper,
	instance *telemetryv1.Ceilometer,
	log logr.Logger,
) error {
	keystoneAPI, err := keystonev1.GetKeystoneAPI(ctx, helper, instance.Namespace, map[string]string{})
	if err != nil {
		return err
	}

	os, ctrlResult, err := keystonev1.GetAdminServiceClient(
		ctx,
		helper,
		keystoneAPI,
	)
	if err != nil || (ctrlResult != ctrl.Result{}) {
		return err
	}

	// We are using the fixed domainID "default" because it is also fixed in ceilometer.conf
	user, err := os.GetUser(log, instance.Spec.ServiceUser, "default")
	if err != nil {
		return err
	}

	// We are using the fixed domainID "default" because it is also fixed in ceilometer.conf
	project, err := getProject(os, log, "service", "default")
	if err != nil {
		return err
	}

	err = os.AssignUserRole(
		log,
		"SwiftSystemReader",
		user.ID,
		project.ID)
	if err != nil {
		log.Error(err, "Cannot AssignUserRole")
		return err
	}

	return nil
}

// getProject - gets project with projectName
func getProject(
	o *openstack.OpenStack,
	log logr.Logger,
	projectName string,
	domainID string,
) (*projects.Project, error) {
	allPages, err := projects.List(o.GetOSClient(), projects.ListOpts{Name: projectName, DomainID: domainID}).AllPages()
	if err != nil {
		return nil, err
	}
	allProjects, err := projects.ExtractProjects(allPages)
	if err != nil {
		return nil, err
	}

	if len(allProjects) == 0 {
		log.Error(err, fmt.Sprintf("%s %s", projectName, "project not found"))
		return nil, err
	} else if len(allProjects) > 1 {
		log.Error(err, fmt.Sprintf("multiple project named \"%s\" found", projectName))
		return nil, err
	}

	return &allProjects[0], nil
}
