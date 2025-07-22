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
	"slices"
	"sort"
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
	statefulset "github.com/openstack-k8s-operators/lib-common/modules/common/statefulset"
	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	util "github.com/openstack-k8s-operators/lib-common/modules/common/util"

	rabbitmqv1 "github.com/openstack-k8s-operators/infra-operator/apis/rabbitmq/v1beta1"
	topologyv1 "github.com/openstack-k8s-operators/infra-operator/apis/topology/v1beta1"
	keystonev1 "github.com/openstack-k8s-operators/keystone-operator/api/v1beta1"
	mariadbv1 "github.com/openstack-k8s-operators/mariadb-operator/api/v1beta1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	availability "github.com/openstack-k8s-operators/telemetry-operator/pkg/availability"
	ceilometer "github.com/openstack-k8s-operators/telemetry-operator/pkg/ceilometer"
	mysqldexporter "github.com/openstack-k8s-operators/telemetry-operator/pkg/mysqldexporter"
	utils "github.com/openstack-k8s-operators/telemetry-operator/pkg/utils"
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
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometers/finalizers,verbs=update;patch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete;
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete;
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete;
// +kubebuilder:rbac:groups=keystone.openstack.org,resources=keystoneservices,verbs=get;list;watch;create;update;patch;delete;
// +kubebuilder:rbac:groups=rabbitmq.openstack.org,resources=transporturls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=keystone.openstack.org,resources=keystoneapis,verbs=get;list;watch;
// +kubebuilder:rbac:groups=mariadb.openstack.org,resources=mariadbdatabases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mariadb.openstack.org,resources=mariadbdatabases/finalizers,verbs=update;patch
// +kubebuilder:rbac:groups=mariadb.openstack.org,resources=mariadbaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mariadb.openstack.org,resources=mariadbaccounts/finalizers,verbs=update;patch
// +kubebuilder:rbac:groups=mariadb.openstack.org,resources=galeras,verbs=get;list;watch
// service account, role, rolebinding
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;list;watch;create;update;patch
// service account permissions that are needed to grant permission to the above
// +kubebuilder:rbac:groups="security.openshift.io",resourceNames=anyuid,resources=securitycontextconstraints,verbs=use
// +kubebuilder:rbac:groups=topology.openstack.org,resources=topologies,verbs=get;list;watch;update

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
	isNewInstance := instance.Status.Conditions == nil
	if isNewInstance {
		instance.Status.Conditions = condition.Conditions{}
	}

	// Save a copy of the conditions so that we can restore the LastTransitionTime
	// when a condition's state doesn't change.
	savedConditions := instance.Status.Conditions.DeepCopy()

	// Always patch the instance status when exiting this function so we can
	// persist any changes.
	defer func() {
		// Don't update the status, if reconciler Panics
		if r := recover(); r != nil {
			Log.Info(fmt.Sprintf("panic during reconcile %v\n", r))
			panic(r)
		}
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
	// initialize conditions used later as Status=Unknown
	cl := condition.CreateList(
		condition.UnknownCondition(condition.ReadyCondition, condition.InitReason, condition.ReadyInitMessage),
		condition.UnknownCondition(condition.InputReadyCondition, condition.InitReason, condition.InputReadyInitMessage),
		condition.UnknownCondition(condition.ServiceConfigReadyCondition, condition.InitReason, condition.ServiceConfigReadyInitMessage),
		condition.UnknownCondition(condition.DeploymentReadyCondition, condition.InitReason, condition.DeploymentReadyInitMessage),
		// right now we have no dedicated KeystoneServiceReadyInitMessage
		condition.UnknownCondition(condition.KeystoneServiceReadyCondition, condition.InitReason, ""),
		condition.UnknownCondition(condition.TLSInputReadyCondition, condition.InitReason, condition.InputReadyInitMessage),

		// MysqldExporter conditions
		condition.UnknownCondition(telemetryv1.MysqldExporterDBReadyCondition, condition.InitReason, condition.DBReadyInitMessage),
		condition.UnknownCondition(telemetryv1.MysqldExporterDeploymentReadyCondition, condition.InitReason, condition.DeploymentReadyInitMessage),
		condition.UnknownCondition(telemetryv1.MysqldExporterMariaDBAccountReadyCondition, condition.InitReason, mariadbv1.MariaDBAccountReadyInitMessage),
		condition.UnknownCondition(telemetryv1.MysqldExporterServiceConfigReadyCondition, condition.InitReason, condition.ServiceConfigReadyInitMessage),
		condition.UnknownCondition(telemetryv1.MysqldExporterTLSInputReadyCondition, condition.InitReason, condition.InputReadyInitMessage),

		// kube-state-metrics conditions
		condition.UnknownCondition(telemetryv1.KSMDeploymentReadyCondition, condition.InitReason, condition.DeploymentReadyInitMessage),
		condition.UnknownCondition(telemetryv1.KSMCreateServiceReadyCondition, condition.InitReason, condition.CreateServiceReadyInitMessage),
		condition.UnknownCondition(telemetryv1.KSMServiceConfigReadyCondition, condition.InitReason, condition.ServiceConfigReadyInitMessage),
		condition.UnknownCondition(telemetryv1.KSMTLSInputReadyCondition, condition.InitReason, condition.InputReadyInitMessage),
	)

	// Init Topology condition if there's a reference
	if instance.Spec.TopologyRef != nil {
		c := condition.UnknownCondition(condition.TopologyReadyCondition, condition.InitReason, condition.TopologyReadyInitMessage)
		cl.Set(c)
	}

	instance.Status.Conditions.Init(&cl)
	instance.Status.ObservedGeneration = instance.Generation

	if instance.Status.Hash == nil {
		instance.Status.Hash = map[string]string{}
	}

	if instance.Status.MysqldExporterHash == nil {
		instance.Status.MysqldExporterHash = map[string]string{}
	}

	if instance.Status.KSMHash == nil {
		instance.Status.KSMHash = map[string]string{}
	}

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

// fields to index to reconcile when change
const (
	ceilometerPasswordSecretField         = ".spec.secret"
	ceilometerCaBundleSecretNameField     = ".spec.tls.caBundleSecretName"
	ceilometerTLSField                    = ".spec.tls.secretName"
	ksmCaBundleSecretNameField            = ".spec.ksmTls.caBundleSecretName"
	ksmTLSField                           = ".spec.ksmTls.secretName"
	mysqldExporterCaBundleSecretNameField = ".spec.mysqldExporterTls.caBundleSecretName"
	mysqldExporterTLSField                = ".spec.mysqldExporterTls.secretName"
)

var (
	ceilometerWatchFields = []string{
		ceilometerPasswordSecretField,
		ceilometerCaBundleSecretNameField,
		ceilometerTLSField,
		ksmCaBundleSecretNameField,
		ksmTLSField,
		mysqldExporterCaBundleSecretNameField,
		mysqldExporterTLSField,
		topologyField,
	}
)

func (r *CeilometerReconciler) mysqldExporterDeleteDBResources(ctx context.Context, instance *telemetryv1.Ceilometer, helper *helper.Helper, galera string) (ctrl.Result, error) {
	databaseName := fmt.Sprintf("%s-%s", mysqldexporter.ServiceName, galera)
	accountName := fmt.Sprintf("%s-%s", instance.Spec.MysqldExporterDatabaseAccountPrefix, galera)

	db, err := mariadbv1.GetDatabaseByNameAndAccount(ctx, helper, databaseName, accountName, instance.Namespace)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if !k8s_errors.IsNotFound(err) {
		if err := db.DeleteFinalizer(ctx, helper); err != nil {
			return ctrl.Result{}, err
		}
		if res, err := utils.EnsureDeleted(ctx, helper, db.GetDatabase()); err != nil {
			return res, err
		}
		if res, err := utils.EnsureDeleted(ctx, helper, db.GetAccount()); err != nil {
			return res, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *CeilometerReconciler) reconcileDeleteMysqldExporter(ctx context.Context, instance *telemetryv1.Ceilometer, helper *helper.Helper) (ctrl.Result, error) {
	// NOTE: We need to delete all created resources explicitly to make the `.spec.mysqldExporterEnabled = false` work.
	for _, galera := range instance.Status.MysqldExporterExportedGaleras {
		if res, err := r.mysqldExporterDeleteDBResources(ctx, instance, helper, galera); err != nil {
			return res, err
		}
	}

	exporterSts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mysqldexporter.ServiceName,
			Namespace: instance.Namespace,
		},
	}
	exporterSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mysqldexporter.ServiceName,
			Namespace: instance.Namespace,
		},
	}
	exporterSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-config-data", mysqldexporter.ServiceName),
			Namespace: instance.Namespace,
		},
	}
	if res, err := utils.EnsureDeleted(ctx, helper, exporterSts); err != nil {
		return res, err
	}
	if res, err := utils.EnsureDeleted(ctx, helper, exporterSvc); err != nil {
		return res, err
	}
	if res, err := utils.EnsureDeleted(ctx, helper, exporterSecret); err != nil {
		return res, err
	}

	instance.Status.Conditions.MarkTrue(telemetryv1.MysqldExporterDBReadyCondition, telemetryv1.MysqldExporterDisabledMessage)
	instance.Status.Conditions.MarkTrue(telemetryv1.MysqldExporterDeploymentReadyCondition, telemetryv1.MysqldExporterDisabledMessage)
	instance.Status.Conditions.MarkTrue(telemetryv1.MysqldExporterMariaDBAccountReadyCondition, telemetryv1.MysqldExporterDisabledMessage)
	instance.Status.Conditions.MarkTrue(telemetryv1.MysqldExporterServiceConfigReadyCondition, telemetryv1.MysqldExporterDisabledMessage)
	instance.Status.Conditions.MarkTrue(telemetryv1.MysqldExporterTLSInputReadyCondition, telemetryv1.MysqldExporterDisabledMessage)

	instance.Status.MysqldExporterExportedGaleras = []string{}
	instance.Status.MysqldExporterReadyCount = 0
	return ctrl.Result{}, nil

}

func (r *CeilometerReconciler) reconcileDeleteKSM(ctx context.Context, instance *telemetryv1.Ceilometer, helper *helper.Helper) (ctrl.Result, error) {
	// NOTE: We need to delete all created resources explicitly to make the `.spec.ksmEnabled = false` work.
	sfset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      availability.KSMServiceName,
			Namespace: instance.Namespace,
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      availability.KSMServiceName,
			Namespace: instance.Namespace,
		},
	}
	scrt := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-tls-config", availability.KSMServiceName),
			Namespace: instance.Namespace,
		},
	}

	for _, obj := range []client.Object{svc, sfset, scrt} {
		if res, err := utils.EnsureDeleted(ctx, helper, obj); err != nil {
			return res, err
		}
	}

	//NOTE(mmagr): To keep ConditionReady state also in case of KSM deletion
	instance.Status.Conditions.MarkTrue(telemetryv1.KSMDeploymentReadyCondition, telemetryv1.KSMDisabledMessage)
	instance.Status.Conditions.MarkTrue(telemetryv1.KSMCreateServiceReadyCondition, telemetryv1.KSMDisabledMessage)
	instance.Status.Conditions.MarkTrue(telemetryv1.KSMServiceConfigReadyCondition, telemetryv1.KSMDisabledMessage)
	instance.Status.Conditions.MarkTrue(telemetryv1.KSMTLSInputReadyCondition, telemetryv1.KSMDisabledMessage)

	instance.Status.KSMReadyCount = 0
	return ctrl.Result{}, nil

}

func (r *CeilometerReconciler) reconcileDelete(ctx context.Context, instance *telemetryv1.Ceilometer, helper *helper.Helper) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info("Reconciling Service delete")
	ctrlResult, err := r.reconcileDeleteMysqldExporter(ctx, instance, helper)
	if (err != nil || ctrlResult != ctrl.Result{}) {
		return ctrlResult, err
	}

	ctrlResult, err = r.reconcileDeleteKSM(ctx, instance, helper)
	if (err != nil || ctrlResult != ctrl.Result{}) {
		return ctrlResult, err
	}

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
			Log.Info(fmt.Sprintf("OpenStack secret %s not found", instance.Spec.Secret))
			return ctrl.Result{RequeueAfter: time.Duration(10) * time.Second}, nil
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
		instance.Status.Conditions.Set(c)
	}

	if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	if instance.Status.Hash == nil {
		instance.Status.Hash = map[string]string{}
	}

	if instance.Status.MysqldExporterHash == nil {
		instance.Status.MysqldExporterHash = map[string]string{}
	}

	if instance.Status.KSMHash == nil {
		instance.Status.KSMHash = map[string]string{}
	}

	Log.Info("Reconciled Service init successfully")
	return ctrl.Result{}, nil
}

func (r *CeilometerReconciler) reconcileNormal(ctx context.Context, instance *telemetryv1.Ceilometer, helper *helper.Helper) (ctrl.Result, error) {
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

	ksmRes, err := r.reconcileKSM(ctx, instance, helper, topology)
	if err != nil {
		return ksmRes, err
	}

	mysqldRes, err := r.reconcileMysqldExporter(ctx, instance, helper, topology)
	if (err != nil || mysqldRes != ctrl.Result{}) {
		return mysqldRes, err
	}

	// NOTE(mmagr): Ceilometer reconciliation has to be the last as this is the (only) place
	//              where condition ReadyCondition is/should be evaluated
	return r.reconcileCeilometer(ctx, instance, helper, topology)
}

func (r *CeilometerReconciler) reconcileCeilometer(
	ctx context.Context,
	instance *telemetryv1.Ceilometer,
	helper *helper.Helper,
	topology *topologyv1.Topology,
) (ctrl.Result, error) {
	// ConfigMap
	configMapVars := make(map[string]env.Setter)

	Log := r.GetLogger(ctx)
	Log.Info(fmt.Sprintf(msgReconcileStart, ceilometer.ServiceName))

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

	//
	// check for required OpenStack secret holding passwords for service/admin user and add hash to the vars map
	//
	ctrlResult, err := r.getSecret(ctx, helper, instance, instance.Spec.Secret, instance.Spec.PasswordSelectors.CeilometerService, &configMapVars)
	if err != nil {
		return ctrlResult, err
	}
	// run check OpenStack secret - end

	//
	// check for required TransportURL secret holding transport URL string
	//
	ctrlResult, err = r.getSecret(ctx, helper, instance, instance.Status.TransportURLSecret, "transport_url", &configMapVars)
	if err != nil {
		return ctrlResult, err
	}
	// run check TransportURL secret - end

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
			configMapVars[tls.CABundleKey] = env.SetValue(hash)
		}
	}

	// Validate metadata service cert secret
	if instance.Spec.TLS.Enabled() {
		hash, err := instance.Spec.TLS.ValidateCertSecret(ctx, helper, instance.Namespace)
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
		configMapVars[tls.TLSHashName] = env.SetValue(hash)
	}
	// all cert input checks out so report InputReady
	instance.Status.Conditions.MarkTrue(condition.TLSInputReadyCondition, condition.InputReadyMessage)

	//
	// create Configmap required for ceilometer input
	// - %-scripts configmap holding scripts to e.g. bootstrap the service
	// - %-config configmap holding minimal ceilometer config required to get the service up, user can add additional files to be added to the service
	// - parameters which has passwords gets added from the OpenStack secret via the init container
	//
	err = r.generateServiceConfig(ctx, helper, instance, &configMapVars)
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
	// create Configmap required for ceilometer-compute input
	// - %-scripts configmap holding scripts to e.g. bootstrap the service
	// - %-config configmap holding minimal ceilometer-compute config required to get the service up, user can add additional files to be added to the service
	// - parameters which has passwords gets added from the OpenStack secret via the init container
	//
	err = r.generateComputeServiceConfig(ctx, helper, instance, &configMapVars)
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
	configMapVars["endpointurls"] = env.SetValue(hash)

	//
	// create hash over all the different input resources to identify if any those changed
	// and a restart/recreate is required.
	//
	inputHash, hashChanged, err := r.createHashOfInputHashes(ctx, &instance.Status.Hash, configMapVars)
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

	instance.Status.Conditions.MarkTrue(condition.ServiceConfigReadyCondition, condition.ServiceConfigReadyMessage)

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
	sfsetDef, err := ceilometer.StatefulSet(instance, inputHash, serviceLabels, topology)
	if err != nil {
		return ctrl.Result{}, err
	}
	sfset := statefulset.NewStatefulSet(
		sfsetDef,
		time.Duration(5)*time.Second,
	)

	ctrlResult, err = sfset.CreateOrPatch(ctx, helper)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DeploymentReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.DeploymentReadyErrorMessage,
			err.Error()))
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DeploymentReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.DeploymentReadyRunningMessage))
		return ctrlResult, nil
	}

	if err := controllerutil.SetControllerReference(instance, sfsetDef, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Evaluate the last part of the reconciliation only if we see the last
	// version of the CR
	if sfset.GetStatefulSet().Generation == sfset.GetStatefulSet().Status.ObservedGeneration {
		instance.Status.ReadyCount = sfset.GetStatefulSet().Status.ReadyReplicas
		instance.Status.Networks = instance.Spec.NetworkAttachmentDefinitions
		svc, op, err := ceilometer.Service(instance, helper, ceilometer.CeilometerPrometheusPort, serviceLabels)
		if err != nil {
			return ctrl.Result{}, err
		}
		if op != controllerutil.OperationResultNone {
			Log.Info(fmt.Sprintf(msgOperation, svc.Name, string(op)))
		}
		if instance.Status.ReadyCount > 0 {
			instance.Status.Conditions.MarkTrue(condition.DeploymentReadyCondition, condition.DeploymentReadyMessage)
		}
		if instance.Status.Conditions.AllSubConditionIsTrue() {
			instance.Status.Conditions.MarkTrue(
				condition.ReadyCondition, condition.ReadyMessage)
		}
		Log.Info(fmt.Sprintf(msgReconcileSuccess, ceilometer.ServiceName))
	}
	return ctrl.Result{}, nil
}

func (r *CeilometerReconciler) reconcileMysqldExporter(
	ctx context.Context,
	instance *telemetryv1.Ceilometer,
	helper *helper.Helper,
	topology *topologyv1.Topology,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info(fmt.Sprintf(msgReconcileStart, mysqldexporter.ServiceName))

	if instance.Spec.MysqldExporterEnabled == nil || !*instance.Spec.MysqldExporterEnabled {
		return r.reconcileDeleteMysqldExporter(ctx, instance, helper)
	}
	if instance.Spec.MysqldExporterImage == "" {
		Log.Info("MysqldExporter is enabled, but MysqldExporterImage isn't set")
		return r.reconcileDeleteMysqldExporter(ctx, instance, helper)
	}

	configMapVars := make(map[string]env.Setter)
	//
	// TLS input validation
	//
	// Validate the CA cert secret if provided
	if instance.Spec.MysqldExporterTLS.CaBundleSecretName != "" {
		hash, err := tls.ValidateCACertSecret(
			ctx,
			helper.GetClient(),
			types.NamespacedName{
				Name:      instance.Spec.MysqldExporterTLS.CaBundleSecretName,
				Namespace: instance.Namespace,
			},
		)
		if err != nil {
			if k8s_errors.IsNotFound(err) {
				instance.Status.Conditions.Set(condition.FalseCondition(
					telemetryv1.MysqldExporterTLSInputReadyCondition,
					condition.RequestedReason,
					condition.SeverityInfo,
					fmt.Sprintf(condition.TLSInputReadyWaitingMessage, instance.Spec.MysqldExporterTLS.CaBundleSecretName)))
				return ctrl.Result{}, nil
			}
			instance.Status.Conditions.Set(condition.FalseCondition(
				telemetryv1.MysqldExporterTLSInputReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.TLSInputErrorMessage,
				err.Error()))
			return ctrl.Result{}, err
		}

		if hash != "" {
			configMapVars[tls.CABundleKey] = env.SetValue(hash)
		}
	}

	// Validate metadata service cert secret
	if instance.Spec.MysqldExporterTLS.Enabled() {
		hash, err := instance.Spec.MysqldExporterTLS.ValidateCertSecret(ctx, helper, instance.Namespace)
		if err != nil {
			if k8s_errors.IsNotFound(err) {
				instance.Status.Conditions.Set(condition.FalseCondition(
					telemetryv1.MysqldExporterTLSInputReadyCondition,
					condition.RequestedReason,
					condition.SeverityInfo,
					fmt.Sprintf(condition.TLSInputReadyWaitingMessage, err.Error())))
				return ctrl.Result{}, nil
			}
			instance.Status.Conditions.Set(condition.FalseCondition(
				telemetryv1.MysqldExporterTLSInputReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.TLSInputErrorMessage,
				err.Error()))
			return ctrl.Result{}, err
		}
		configMapVars[tls.TLSHashName] = env.SetValue(hash)
	}
	// all cert input checks out so report InputReady
	instance.Status.Conditions.MarkTrue(telemetryv1.MysqldExporterTLSInputReadyCondition, condition.InputReadyMessage)

	//
	// create Configmap required for mysqld_exporter input
	// - %-config configmap holding minimal mysqld_exporter config required to get the service up
	//
	result, err := r.generateMysqldExporterServiceConfig(ctx, helper, instance, &configMapVars)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.MysqldExporterServiceConfigReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.ServiceConfigReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	} else if (result != ctrl.Result{}) {
		return result, nil
	}

	//
	// create hash over all the different input resources to identify if any those changed
	// and a restart/recreate is required.
	//
	inputHash, hashChanged, err := r.createHashOfInputHashes(ctx, &instance.Status.MysqldExporterHash, configMapVars)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.MysqldExporterServiceConfigReadyCondition,
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

	instance.Status.MysqldExporterHash[common.InputHashName] = inputHash

	instance.Status.Conditions.MarkTrue(telemetryv1.MysqldExporterServiceConfigReadyCondition, condition.ServiceConfigReadyMessage)

	serviceLabels := map[string]string{
		common.AppSelector:   mysqldexporter.ServiceName,
		common.OwnerSelector: instance.Name,
	}

	// Define a new StatefulSet object
	sfsetDef, err := mysqldexporter.StatefulSet(instance, inputHash, serviceLabels, topology)
	if err != nil {
		return ctrl.Result{}, err
	}
	sfset := statefulset.NewStatefulSet(
		sfsetDef,
		time.Duration(5)*time.Second,
	)

	ctrlResult, err := sfset.CreateOrPatch(ctx, helper)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.MysqldExporterDeploymentReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.DeploymentReadyErrorMessage,
			err.Error()))
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.MysqldExporterDeploymentReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.DeploymentReadyRunningMessage))
		return ctrlResult, nil
	}

	if err := controllerutil.SetControllerReference(instance, sfsetDef, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Evaluate the last part of the reconciliation only if we see the last
	// version of the CR
	if sfset.GetStatefulSet().Generation == sfset.GetStatefulSet().Status.ObservedGeneration {
		instance.Status.MysqldExporterReadyCount = sfset.GetStatefulSet().Status.ReadyReplicas
		svc, op, err := mysqldexporter.Service(instance, helper, serviceLabels)
		if err != nil {
			return ctrl.Result{}, err
		}
		if op != controllerutil.OperationResultNone {
			Log.Info(fmt.Sprintf(msgOperation, svc.Name, string(op)))
		}
		if instance.Status.MysqldExporterReadyCount > 0 {
			instance.Status.Conditions.MarkTrue(telemetryv1.MysqldExporterDeploymentReadyCondition, condition.DeploymentReadyMessage)
		}
		Log.Info(fmt.Sprintf(msgReconcileSuccess, mysqldexporter.ServiceName))
	}

	return ctrl.Result{}, nil
}

func (r *CeilometerReconciler) reconcileKSM(
	ctx context.Context,
	instance *telemetryv1.Ceilometer,
	helper *helper.Helper,
	topology *topologyv1.Topology,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info(fmt.Sprintf(msgReconcileStart, availability.KSMServiceName))

	if instance.Spec.KSMEnabled == nil || !*instance.Spec.KSMEnabled {
		return r.reconcileDeleteKSM(ctx, instance, helper)
	}
	if instance.Spec.KSMImage == "" {
		Log.Info("KSM is enabled, but KSMImage isn't set")
		return r.reconcileDeleteKSM(ctx, instance, helper)
	}

	// ConfigMap
	configMapVars := make(map[string]env.Setter)

	serviceLabels := map[string]string{
		common.AppSelector: availability.KSMServiceName,
	}

	// Validate the CA cert secret if provided
	if instance.Spec.KSMTLS.CaBundleSecretName != "" {
		hash, err := tls.ValidateCACertSecret(
			ctx,
			helper.GetClient(),
			types.NamespacedName{
				Name:      instance.Spec.KSMTLS.CaBundleSecretName,
				Namespace: instance.Namespace,
			},
		)
		if err != nil {
			if k8s_errors.IsNotFound(err) {
				instance.Status.Conditions.Set(condition.FalseCondition(
					telemetryv1.KSMTLSInputReadyCondition,
					condition.RequestedReason,
					condition.SeverityInfo,
					fmt.Sprintf(condition.TLSInputReadyWaitingMessage, instance.Spec.KSMTLS.CaBundleSecretName)))
				return ctrl.Result{}, nil
			}
			instance.Status.Conditions.Set(condition.FalseCondition(
				telemetryv1.KSMTLSInputReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.TLSInputErrorMessage,
				err.Error()))
			return ctrl.Result{}, err
		}

		if hash != "" {
			configMapVars[tls.CABundleKey] = env.SetValue(hash)
		}
	}

	tlsConfName := ""
	if instance.Spec.KSMTLS.Enabled() {
		// Validate metadata service cert secret
		hash, err := instance.Spec.KSMTLS.ValidateCertSecret(ctx, helper, instance.Namespace)
		if err != nil {
			instance.Status.Conditions.Set(condition.FalseCondition(
				telemetryv1.KSMTLSInputReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.TLSInputErrorMessage,
				err.Error()))
			return ctrl.Result{}, err
		}
		configMapVars[tls.TLSHashName] = env.SetValue(hash)

		// Create TLS conf for the service
		tlsConfDef := availability.KSMTLSConfig(instance, serviceLabels, true)
		tlsConfName = tlsConfDef.Name

		hash, op, err := secret.CreateOrPatchSecret(ctx, helper, instance, tlsConfDef)
		if err != nil {
			return ctrl.Result{}, err
		}
		if op != controllerutil.OperationResultNone {
			Log.Info(fmt.Sprintf("KSM TLS config %s successfully changed - operation: %s", tlsConfDef.Name, string(op)))
		}
		configMapVars[tlsConfDef.Name] = env.SetValue(hash)
	}

	instance.Status.Conditions.MarkTrue(telemetryv1.KSMTLSInputReadyCondition, condition.InputReadyMessage)

	//
	// create hash over all the different input resources to identify if any those changed
	// and a restart/recreate is required.
	//
	inputHash, hashChanged, err := r.createHashOfInputHashes(ctx, &instance.Status.KSMHash, configMapVars)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.KSMServiceConfigReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.ServiceConfigReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	} else if hashChanged {
		return ctrl.Result{}, nil
	}

	instance.Status.KSMHash[common.InputHashName] = inputHash
	instance.Status.Conditions.MarkTrue(telemetryv1.KSMServiceConfigReadyCondition, condition.InputReadyMessage)

	// create the kube-state-metrics statefulset
	ssDef, err := availability.KSMStatefulSet(instance, tlsConfName, serviceLabels, topology)
	if err != nil {
		return ctrl.Result{}, err
	}

	ss := statefulset.NewStatefulSet(ssDef, time.Duration(5)*time.Second)
	ctrlResult, err := ss.CreateOrPatch(ctx, helper)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.KSMDeploymentReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.DeploymentReadyErrorMessage,
			err.Error()))
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.KSMDeploymentReadyCondition,
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
		instance.Status.KSMReadyCount = ss.GetStatefulSet().Status.ReadyReplicas
		if instance.Status.KSMReadyCount > 0 {
			instance.Status.Conditions.MarkTrue(telemetryv1.KSMDeploymentReadyCondition, condition.DeploymentReadyMessage)
		}

		// Create the service
		svc, op, err := availability.KSMService(instance, helper, serviceLabels)
		if err != nil {
			instance.Status.Conditions.Set(condition.FalseCondition(
				telemetryv1.KSMCreateServiceReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.CreateServiceReadyErrorMessage,
				err.Error()))

			return ctrl.Result{}, err
		}
		if op != controllerutil.OperationResultNone {
			Log.Info(fmt.Sprintf(msgOperation, svc.Name, string(op)))
		}
		if op == controllerutil.OperationResultCreated || op == controllerutil.OperationResultUpdated {
			instance.Status.Conditions.MarkTrue(
				telemetryv1.KSMCreateServiceReadyCondition, condition.CreateServiceReadyMessage)
		}

		Log.Info(fmt.Sprintf(msgReconcileSuccess, availability.KSMServiceName))
	}

	return ctrl.Result{}, nil
}

// getSecret - get the specified secret, and add its hash to envVars
func (r *CeilometerReconciler) getSecret(ctx context.Context, h *helper.Helper, instance *telemetryv1.Ceilometer, secretName string, expectedField string, envVars *map[string]env.Setter) (ctrl.Result, error) {
	secretHash, result, err := ensureSecret(
		ctx,
		types.NamespacedName{Namespace: instance.Namespace, Name: secretName},
		[]string{
			expectedField,
		},
		h.GetClient(),
		&instance.Status.Conditions,
		time.Duration(10)*time.Second,
	)
	if err != nil {
		return result, err
	}

	// Add a prefix to the var name to avoid accidental collision with other non-secret
	// vars. The secret names themselves will be unique.
	(*envVars)["secret-"+secretName] = env.SetValue(secretHash)

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

	transportURLSecret, _, err := secret.GetSecret(ctx, h, instance.Status.TransportURLSecret, instance.Namespace)
	if err != nil {
		return err
	}

	ceilometerPasswordSecret, _, err := secret.GetSecret(ctx, h, instance.Spec.Secret, instance.Namespace)
	if err != nil {
		return err
	}

	templateParameters := map[string]interface{}{
		"KeystoneInternalURL": keystoneInternalURL,
		"TransportURL":        string(transportURLSecret.Data["transport_url"]),
		"CeilometerPassword":  string(ceilometerPasswordSecret.Data["CeilometerPassword"]),
		"TLS":                 false, // Default to false. Change to true later if TLS enabled
		"SwiftRole":           false, //
		"Timeout":             instance.Spec.APITimeout,
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
			Name:         fmt.Sprintf("%s-scripts", ceilometer.ServiceName),
			Namespace:    instance.Namespace,
			Type:         util.TemplateTypeScripts,
			InstanceType: "ceilometercentral",
			AdditionalTemplate: map[string]string{
				"common.sh":             "/common/common.sh",
				"centralhealth.py":      "/ceilometercentral/bin/centralhealth.py",
				"notificationhealth.py": "/ceilometercentral/bin/notificationhealth.py",
			},
			Labels: cmLabels,
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
	ipmiLabels := labels.GetLabels(instance, labels.GetGroupLabel(ceilometer.IpmiServiceName), map[string]string{})
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

	transportURLSecret, _, err := secret.GetSecret(ctx, h, instance.Status.TransportURLSecret, instance.Namespace)
	if err != nil {
		return err
	}

	ceilometerPasswordSecret, _, err := secret.GetSecret(ctx, h, instance.Spec.Secret, instance.Namespace)
	if err != nil {
		return err
	}

	templateParameters := map[string]interface{}{
		"KeystoneInternalURL":      keystoneInternalURL,
		"TransportURL":             string(transportURLSecret.Data["transport_url"]),
		"CeilometerPassword":       string(ceilometerPasswordSecret.Data["CeilometerPassword"]),
		"ceilometer_compute_image": instance.Spec.ComputeImage,
		"ceilometer_ipmi_image":    instance.Spec.IpmiImage,
		"TLS":                      false,
	}

	if instance.Spec.TLS.Enabled() {
		templateParameters["TLS"] = true
		templateParameters["TlsCert"] = "/etc/ceilometer/tls/tls.crt"
		templateParameters["TlsKey"] = "/etc/ceilometer/tls/tls.key"
	}

	cms := []util.Template{
		// CeilometerCompute ScriptsConfigMap
		{
			Name:               fmt.Sprintf("%s-scripts", ceilometer.ComputeServiceName),
			Namespace:          instance.Namespace,
			Type:               util.TemplateTypeScripts,
			InstanceType:       "ceilometercompute",
			AdditionalTemplate: map[string]string{"common.sh": "/common/common.sh"},
			Labels:             cmLabels,
		},
		// CeilometerCompute ConfigMap
		{
			Name:          fmt.Sprintf("%s-config-data", ceilometer.ComputeServiceName),
			Namespace:     instance.Namespace,
			Type:          util.TemplateTypeConfig,
			InstanceType:  "ceilometercompute",
			CustomData:    customData,
			ConfigOptions: templateParameters,
			Labels:        cmLabels,
		},
		// CeilometerIpmi ScriptsConfigMap
		{
			Name:               fmt.Sprintf("%s-scripts", ceilometer.IpmiServiceName),
			Namespace:          instance.Namespace,
			Type:               util.TemplateTypeScripts,
			InstanceType:       "ceilometeripmi",
			AdditionalTemplate: map[string]string{"common.sh": "/common/common.sh"},
			Labels:             ipmiLabels,
		},
		// CeilometerIpmi ConfigMap
		{
			Name:          fmt.Sprintf("%s-config-data", ceilometer.IpmiServiceName),
			Namespace:     instance.Namespace,
			Type:          util.TemplateTypeConfig,
			InstanceType:  "ceilometeripmi",
			CustomData:    customData,
			ConfigOptions: templateParameters,
			Labels:        ipmiLabels,
		},
	}
	return secret.EnsureSecrets(ctx, h, instance, cms, envVars)
}

// mysqldExporterEnsureDB - create mysqld_exporter DB account
func (r *CeilometerReconciler) mysqldExporterEnsureDB(
	ctx context.Context,
	h *helper.Helper,
	instance *telemetryv1.Ceilometer,
	dbInstance string,
) (*mariadbv1.MariaDBAccount, *corev1.Secret, ctrl.Result, error) {
	accountName := fmt.Sprintf("%s-%s", instance.Spec.MysqldExporterDatabaseAccountPrefix, dbInstance)
	databaseName := fmt.Sprintf("%s-%s", mysqldexporter.ServiceName, dbInstance)

	// ensure MariaDBAccount exists.  This account record may be created by
	// openstack-operator or the cloud operator up front without a specific
	// MariaDBDatabase configured yet.   Otherwise, a MariaDBAccount CR is
	// created here with a generated username as well as a secret with
	// generated password.   The MariaDBAccount is created without being
	// yet associated with any MariaDBDatabase.
	account, secret, err := mariadbv1.EnsureMariaDBAccount(
		ctx, h, accountName,
		instance.Namespace, false, mysqldexporter.DatabaseUsernamePrefix,
	)

	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.MysqldExporterMariaDBAccountReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			mariadbv1.MariaDBAccountNotReadyMessage,
			err.Error()))

		return nil, nil, ctrl.Result{}, err
	}
	instance.Status.Conditions.MarkTrue(
		telemetryv1.MysqldExporterMariaDBAccountReadyCondition,
		mariadbv1.MariaDBAccountReadyMessage)

	// Create a DB for the account
	// TODO: remove creation of a db, when required mariadb-operator support exists.
	// We don't actually need any db created for the mysqld_exporter.
	// What we need is just some account with permission to execute: "SHOW GLOBAL STATUS" and "SHOW GLOBAL VARIABLES"
	// Unfortunatelly access to the database is granted only after creating a db.
	db := mariadbv1.NewDatabaseForAccount(
		dbInstance, // mariadb/galera service to target
		strings.Replace(databaseName, "-", "_", -1), // name used in CREATE DATABASE in mariadb
		databaseName,       // CR name for MariaDBDatabase
		accountName,        // CR name for MariaDBAccount
		instance.Namespace, // namespace
	)

	// create or patch the DB
	ctrlResult, err := db.CreateOrPatchAll(ctx, h)

	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.MysqldExporterDBReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.DBReadyErrorMessage,
			err.Error()))
		return nil, nil, ctrl.Result{}, err
	}
	if (ctrlResult != ctrl.Result{}) {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.MysqldExporterDBReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.DBReadyRunningMessage))
		return nil, nil, ctrlResult, nil
	}

	// wait for the DB to be setup
	ctrlResult, err = db.WaitForDBCreated(ctx, h)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.MysqldExporterDBReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.DBReadyErrorMessage,
			err.Error()))
		return nil, nil, ctrlResult, err
	}
	if (ctrlResult != ctrl.Result{}) {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.MysqldExporterDBReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.DBReadyRunningMessage))
		return nil, nil, ctrlResult, nil
	}

	return account, secret, ctrl.Result{}, nil
}

func (r *CeilometerReconciler) generateMysqldExporterServiceConfig(
	ctx context.Context,
	h *helper.Helper,
	instance *telemetryv1.Ceilometer,
	envVars *map[string]env.Setter,
) (ctrl.Result, error) {
	secretLabels := labels.GetLabels(instance, labels.GetGroupLabel(mysqldexporter.ServiceName), map[string]string{})

	// get all Galera CRs
	galeras := &mariadbv1.GaleraList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.GetNamespace()),
	}
	if err := r.Client.List(context.Background(), galeras, listOpts...); err != nil {
		return ctrl.Result{}, err
	}

	// Sort the galeras, so that we always generate the config in the same
	// order of galera instances. Otherwise the config hash would change
	// a lot, which would lead to a lot of unnecessary reconciles.
	sort.Slice(galeras.Items, func(i, j int) bool {
		return galeras.Items[i].GetName() < galeras.Items[j].GetName()
	})

	instance.Status.MysqldExporterExportedGaleras = []string{}

	databases := []map[string]interface{}{}
	for _, galera := range galeras.Items {
		galeraName := galera.GetName()

		if !galera.DeletionTimestamp.IsZero() {
			// This galera is trying to be deleted. So ensure we delete our resources so
			// we don't block its deletion.
			result, err := r.mysqldExporterDeleteDBResources(ctx, instance, h, galeraName)
			if (err != nil || result != ctrl.Result{}) {
				return result, err
			}
			continue
		}

		dbAccount, dbSecret, result, err := r.mysqldExporterEnsureDB(ctx, h, instance, galeraName)
		if err != nil {
			return ctrl.Result{}, err
		}
		if (result != ctrl.Result{}) {
			return result, nil
		}

		hostname, result, err := mariadbv1.GetServiceHostname(ctx, h, galeraName, instance.Namespace)
		if (err != nil || result != ctrl.Result{}) {
			return result, err
		}
		databaseParameters := map[string]interface{}{
			"Name":       fmt.Sprintf("client.%s.%s.svc", galeraName, galera.GetNamespace()),
			"Host":       hostname,
			"User":       dbAccount.Spec.UserName,
			"Password":   string(dbSecret.Data[mariadbv1.DatabasePasswordSelector]),
			"TLSEnabled": instance.Spec.MysqldExporterTLS.Enabled(),
		}
		databases = append(databases, databaseParameters)

		instance.Status.MysqldExporterExportedGaleras = append(
			instance.Status.MysqldExporterExportedGaleras,
			galeraName,
		)
	}

	instance.Status.Conditions.MarkTrue(telemetryv1.MysqldExporterDBReadyCondition, condition.DBReadyMessage)

	if len(databases) > 0 {
		// There needs to be a section called "client" in the config
		clientParameters := map[string]interface{}{
			"Name":       "client",
			"Host":       databases[0]["Host"],
			"User":       databases[0]["User"],
			"Password":   databases[0]["Password"],
			"TLSEnabled": databases[0]["TLSEnabled"],
		}
		databases = append(databases, clientParameters)
	}
	templateParameters := map[string]interface{}{
		"Databases": databases,
		"TLS": map[string]string{
			"Cert": fmt.Sprintf("/etc/pki/tls/certs/%s", tls.CertKey),
			"Key":  fmt.Sprintf("/etc/pki/tls/private/%s", tls.PrivateKey),
			"Ca":   tls.DownstreamTLSCABundlePath,
		},
	}

	secrets := []util.Template{
		{
			Name:          fmt.Sprintf("%s-config-data", mysqldexporter.ServiceName),
			Namespace:     instance.Namespace,
			Type:          util.TemplateTypeConfig,
			InstanceType:  "mysqldexporter",
			ConfigOptions: templateParameters,
			Labels:        secretLabels,
		},
	}

	err := secret.EnsureSecrets(ctx, h, instance, secrets, envVars)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// createHashOfInputHashes - creates a hash of hashes which gets added to the resources which requires a restart
// if any of the input resources change, like configs, passwords, ...
//
// returns the hash, whether the hash changed (as a bool) and any error
func (r *CeilometerReconciler) createHashOfInputHashes(
	ctx context.Context,
	oldHashMap *map[string]string,
	envVars map[string]env.Setter,
) (string, bool, error) {
	Log := r.GetLogger(ctx)
	changed := false
	hash, err := util.HashOfInputHashes(envVars)
	if err != nil {
		return hash, changed, err
	}
	if hashMap, changed := util.SetHash(*oldHashMap, common.InputHashName, hash); changed {
		*oldHashMap = hashMap
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
	transportURLSecretFn := func(_ context.Context, o client.Object) []reconcile.Request {
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
	keystoneEndpointsWatchFn := func(_ context.Context, o client.Object) []reconcile.Request {
		result := []reconcile.Request{}
		name := client.ObjectKey{
			Namespace: o.GetNamespace(),
			Name:      ceilometer.ServiceName,
		}
		result = append(result, reconcile.Request{NamespacedName: name})
		return result
	}

	galeraWatchFn := func(_ context.Context, o client.Object) []reconcile.Request {
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

		for _, cr := range ceilometers.Items {
			// return namespace and Name of CR
			name := client.ObjectKey{
				Namespace: o.GetNamespace(),
				Name:      cr.Name,
			}
			if !slices.Contains(cr.Status.MysqldExporterExportedGaleras, o.GetName()) {
				Log.Info(fmt.Sprintf("There is a galera %s, which isn't exported by a ceilometer %s's mysqld_exporter yet.", o.GetName(), cr.Name))
				result = append(result, reconcile.Request{NamespacedName: name})
			} else if !o.GetDeletionTimestamp().IsZero() {
				Log.Info(fmt.Sprintf("There is a galera %s, which is exported by a ceilometer %s's mysqld_exporter, but it's being deleted.", o.GetName(), cr.Name))
				result = append(result, reconcile.Request{NamespacedName: name})
			}
		}
		if len(result) > 0 {
			return result
		}
		return nil
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

	// index ceilometerTLSField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.Ceilometer{}, ceilometerTLSField, func(rawObj client.Object) []string {
		// Extract the secret name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.Ceilometer)
		if cr.Spec.TLS.SecretName == nil {
			return nil
		}
		return []string{*cr.Spec.TLS.SecretName}
	}); err != nil {
		return err
	}

	// index ksmCaBundleSecretNameField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.Ceilometer{}, ksmCaBundleSecretNameField, func(rawObj client.Object) []string {
		// Extract the secret name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.Ceilometer)
		if cr.Spec.KSMTLS.CaBundleSecretName == "" {
			return nil
		}
		return []string{cr.Spec.KSMTLS.CaBundleSecretName}
	}); err != nil {
		return err
	}

	// index ksmTLSField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.Ceilometer{}, ksmTLSField, func(rawObj client.Object) []string {
		// Extract the secret name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.Ceilometer)
		if cr.Spec.KSMTLS.SecretName == nil {
			return nil
		}
		return []string{*cr.Spec.KSMTLS.SecretName}
	}); err != nil {
		return err
	}

	// index mysqldExporterCaBundleSecretNameField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.Ceilometer{}, mysqldExporterCaBundleSecretNameField, func(rawObj client.Object) []string {
		// Extract the secret name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.Ceilometer)
		if cr.Spec.MysqldExporterTLS.CaBundleSecretName == "" {
			return nil
		}
		return []string{cr.Spec.MysqldExporterTLS.CaBundleSecretName}
	}); err != nil {
		return err
	}

	// index mysqldExporterTLSField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.Ceilometer{}, mysqldExporterTLSField, func(rawObj client.Object) []string {
		// Extract the secret name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.Ceilometer)
		if cr.Spec.MysqldExporterTLS.SecretName == nil {
			return nil
		}
		return []string{*cr.Spec.MysqldExporterTLS.SecretName}
	}); err != nil {
		return err
	}

	// index topologyField
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &telemetryv1.Ceilometer{}, topologyField, func(rawObj client.Object) []string {
		// Extract the topology name from the spec, if one is provided
		cr := rawObj.(*telemetryv1.Ceilometer)
		if cr.Spec.TopologyRef == nil {
			return nil
		}
		return []string{cr.Spec.TopologyRef.Name}
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
		Owns(&mariadbv1.MariaDBDatabase{}).
		Owns(&mariadbv1.MariaDBAccount{}).
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
		Watches(
			&mariadbv1.Galera{},
			handler.EnqueueRequestsFromMapFunc(galeraWatchFn),
		).
		Watches(&topologyv1.Topology{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForSrc),
			builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&keystonev1.KeystoneAPI{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectForSrc),
			builder.WithPredicates(keystonev1.KeystoneAPIStatusChangedPredicate)).
		Complete(r)
}

func (r *CeilometerReconciler) findObjectsForSrc(ctx context.Context, src client.Object) []reconcile.Request {
	requests := []reconcile.Request{}

	Log := r.GetLogger(ctx)

	for _, field := range ceilometerWatchFields {
		crList := &telemetryv1.CeilometerList{}
		listOps := &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(field, src.GetName()),
			Namespace:     src.GetNamespace(),
		}
		err := r.Client.List(ctx, crList, listOps)
		if err != nil {
			Log.Error(err, fmt.Sprintf("listing %s for field: %s - %s", crList.GroupVersionKind().Kind, field, src.GetNamespace()))
			return requests
		}

		for _, item := range crList.Items {
			Log.Info(fmt.Sprintf("input source %s changed, reconcile: %s - %s", src.GetName(), item.GetName(), item.GetNamespace()))

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

func (r *CeilometerReconciler) findObjectForSrc(ctx context.Context, src client.Object) []reconcile.Request {
	requests := []reconcile.Request{}

	Log := r.GetLogger(ctx)

	crList := &telemetryv1.CeilometerList{}
	listOps := &client.ListOptions{
		Namespace: src.GetNamespace(),
	}
	err := r.Client.List(ctx, crList, listOps)
	if err != nil {
		Log.Error(err, fmt.Sprintf("listing %s for namespace: %s", crList.GroupVersionKind().Kind, src.GetNamespace()))
		return requests
	}

	for _, item := range crList.Items {
		Log.Info(fmt.Sprintf("input source %s changed, reconcile: %s - %s", src.GetName(), item.GetName(), item.GetNamespace()))

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
	project, err := os.GetProject(log, "service", "default")
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
