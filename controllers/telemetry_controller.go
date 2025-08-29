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

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	logr "github.com/go-logr/logr"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"k8s.io/client-go/kubernetes"

	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	ceilometer "github.com/openstack-k8s-operators/telemetry-operator/pkg/ceilometer"
	logging "github.com/openstack-k8s-operators/telemetry-operator/pkg/logging"
	utils "github.com/openstack-k8s-operators/telemetry-operator/pkg/utils"
)

// TelemetryReconciler reconciles a Telemetry object
type TelemetryReconciler struct {
	client.Client
	Kclient kubernetes.Interface
	Scheme  *runtime.Scheme
}

// GetLogger returns a logger object with a prefix of "conroller.name" and aditional controller context fields
func (r *TelemetryReconciler) GetLogger(ctx context.Context) logr.Logger {
	return log.FromContext(ctx).WithName("Controllers").WithName("Telemetry")
}

// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=telemetries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=telemetries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=telemetries/finalizers,verbs=update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=autoscalings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=autoscalings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=autoscalings/finalizers,verbs=update;delete;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometers/finalizers,verbs=update;delete;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=metricstorages,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=metricstorages/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=metricstorages/finalizers,verbs=update;delete;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=loggings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=loggings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=loggings/finalizers,verbs=update;delete;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=cloudkitties,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=cloudkitties/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=cloudkitties/finalizers,verbs=update;delete;patch
// +kubebuilder:rbac:groups=rabbitmq.openstack.org,resources=transporturls,verbs=get;list;watch;create;update;patch;delete

// Reconcile reconciles a Telemetry
func (r *TelemetryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, _err error) {
	Log := r.GetLogger(ctx)

	// Fetch the Telemetry instance
	instance := &telemetryv1.Telemetry{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected.
			// For additional cleanup logic use finalizers. Return and don't requeue.
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		Log.Error(err, fmt.Sprintf("could not fetch instance %s", instance.Name))
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
		condition.UnknownCondition(telemetryv1.CeilometerReadyCondition, condition.InitReason, telemetryv1.CeilometerReadyInitMessage),
		condition.UnknownCondition(telemetryv1.AutoscalingReadyCondition, condition.InitReason, telemetryv1.AutoscalingReadyInitMessage),
		condition.UnknownCondition(telemetryv1.MetricStorageReadyCondition, condition.InitReason, telemetryv1.MetricStorageReadyInitMessage),
		condition.UnknownCondition(telemetryv1.LoggingReadyCondition, condition.InitReason, telemetryv1.LoggingReadyInitMessage),
		condition.UnknownCondition(telemetryv1.CloudKittyReadyCondition, condition.InitReason, telemetryv1.CloudKittyReadyInitMessage),
	)

	instance.Status.Conditions.Init(&cl)
	instance.Status.ObservedGeneration = instance.Generation

	// If we're not deleting this and the service object doesn't have our finalizer, add it.
	if instance.DeletionTimestamp.IsZero() && controllerutil.AddFinalizer(instance, helper.GetFinalizer()) || isNewInstance {
		return ctrl.Result{}, nil
	}

	if instance.Status.Hash == nil {
		instance.Status.Hash = map[string]string{}
	}

	// Handle service init
	ctrlResult, err := r.reconcileInit(ctx)
	if err != nil {
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	// Handle service delete
	if !instance.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, instance, helper)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, instance, helper)
}

func (r *TelemetryReconciler) reconcileDelete(ctx context.Context, instance *telemetryv1.Telemetry, helper *helper.Helper) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info(fmt.Sprintf("Reconciled Service '%s' delete", instance.Name))

	// Service is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(instance, helper.GetFinalizer())

	Log.Info(fmt.Sprintf("Reconciled Service '%s' delete successfully", instance.Name))
	return ctrl.Result{}, nil
}

func (r *TelemetryReconciler) reconcileInit(
	ctx context.Context,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info("Reconciling Service init")

	Log.Info("Reconciled Service init successfully")
	return ctrl.Result{}, nil
}

func (r *TelemetryReconciler) reconcileNormal(ctx context.Context, instance *telemetryv1.Telemetry, helper *helper.Helper) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info(fmt.Sprintf("Reconciling Service '%s'", instance.Name))

	ctrlResult, err := r.reconcileAutoscaling(ctx, instance, helper)
	if err != nil {
		return ctrl.Result{}, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	ctrlResult, err = r.reconcileCeilometer(ctx, instance, helper)
	if err != nil {
		return ctrl.Result{}, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	ctrlResult, err = r.reconcileMetricStorage(ctx, instance, helper)
	if err != nil {
		return ctrl.Result{}, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	ctrlResult, err = r.reconcileLogging(ctx, instance, helper)
	if err != nil {
		return ctrl.Result{}, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	ctrlResult, err = r.reconcileCloudKitty(ctx, instance, helper)
	if err != nil {
		return ctrl.Result{}, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	// We reached the end of the Reconcile, update the Ready condition based on
	// the sub conditions
	if instance.Status.Conditions.AllSubConditionIsTrue() {
		instance.Status.Conditions.MarkTrue(
			condition.ReadyCondition, condition.ReadyMessage)
	}
	Log.Info("Reconciled Service successfully")
	return ctrl.Result{}, nil
}

// reconcileCeilometer ...
func (r TelemetryReconciler) reconcileCeilometer(ctx context.Context, instance *telemetryv1.Telemetry, helper *helper.Helper) (ctrl.Result, error) {
	const (
		ceilometerNamespaceLabel = "Ceilometer.Namespace"
		ceilometerNameLabel      = "Ceilometer.Name"
	)
	ceilometerInstance := &telemetryv1.Ceilometer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ceilometer.ServiceName,
			Namespace: instance.Namespace,
		},
	}

	if instance.Spec.Ceilometer.Enabled == nil || !*instance.Spec.Ceilometer.Enabled {
		if res, err := utils.EnsureDeleted(ctx, helper, ceilometerInstance); err != nil {
			return res, err
		}
		instance.Status.Conditions.Remove(telemetryv1.CeilometerReadyCondition)
		return ctrl.Result{}, nil
	}

	if instance.Spec.Ceilometer.NodeSelector == nil {
		instance.Spec.Ceilometer.NodeSelector = instance.Spec.NodeSelector
	}

	if instance.Spec.Ceilometer.TopologyRef == nil {
		instance.Spec.Ceilometer.TopologyRef = instance.Spec.TopologyRef
	}

	helper.GetLogger().Info("Reconciling Ceilometer", ceilometerNamespaceLabel, instance.Namespace, ceilometerNameLabel, ceilometer.ServiceName)
	op, err := controllerutil.CreateOrPatch(ctx, helper.GetClient(), ceilometerInstance, func() error {
		instance.Spec.Ceilometer.CeilometerSpec.DeepCopyInto(&ceilometerInstance.Spec)

		err := controllerutil.SetControllerReference(helper.GetBeforeObject(), ceilometerInstance, helper.GetScheme())
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CeilometerReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			telemetryv1.CeilometerReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}

	// Check the observed Generation and mirror the condition from the underlying
	// resource reconciliation
	ceilObsGen, err := r.checkCeilometerGeneration(instance)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CeilometerReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			telemetryv1.CeilometerReadyRunningMessage))
		return ctrl.Result{}, nil
	}

	if !ceilObsGen {
		instance.Status.Conditions.Set(condition.UnknownCondition(
			telemetryv1.CeilometerReadyCondition,
			condition.InitReason,
			telemetryv1.CeilometerReadyRunningMessage,
		))
	} else {

		// Mirror Ceilometer's condition status
		c := ceilometerInstance.Status.Conditions.Mirror(telemetryv1.CeilometerReadyCondition)
		if c != nil {
			instance.Status.Conditions.Set(c)
		}
	}

	if op != controllerutil.OperationResultNone && ceilObsGen {
		helper.GetLogger().Info(fmt.Sprintf("%s %s - %s", ceilometer.ServiceName, ceilometerInstance.Name, op))
	}

	return ctrl.Result{}, nil
}

// reconcileAutoscaling ...
func (r TelemetryReconciler) reconcileAutoscaling(ctx context.Context, instance *telemetryv1.Telemetry, helper *helper.Helper) (ctrl.Result, error) {
	const (
		autoscalingNamespaceLabel = "Autoscaling.Namespace"
		autoscalingNameLabel      = "Autoscaling.Name"
		autoscalingName           = "autoscaling"
	)
	autoscalingInstance := &telemetryv1.Autoscaling{
		ObjectMeta: metav1.ObjectMeta{
			Name:      autoscalingName,
			Namespace: instance.Namespace,
		},
	}

	if instance.Spec.Autoscaling.Enabled == nil || !*instance.Spec.Autoscaling.Enabled {
		if res, err := utils.EnsureDeleted(ctx, helper, autoscalingInstance); err != nil {
			return res, err
		}
		instance.Status.Conditions.Remove(telemetryv1.AutoscalingReadyCondition)
		return ctrl.Result{}, nil
	}

	if instance.Spec.Autoscaling.Aodh.NodeSelector == nil {
		instance.Spec.Autoscaling.Aodh.NodeSelector = instance.Spec.NodeSelector
	}

	if instance.Spec.Autoscaling.Aodh.TopologyRef == nil {
		instance.Spec.Autoscaling.Aodh.TopologyRef = instance.Spec.TopologyRef
	}

	helper.GetLogger().Info("Reconciling Autoscaling", autoscalingNamespaceLabel, instance.Namespace, autoscalingNameLabel, autoscalingName)
	op, err := controllerutil.CreateOrPatch(ctx, helper.GetClient(), autoscalingInstance, func() error {
		instance.Spec.Autoscaling.AutoscalingSpec.DeepCopyInto(&autoscalingInstance.Spec)

		err := controllerutil.SetControllerReference(helper.GetBeforeObject(), autoscalingInstance, helper.GetScheme())
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.AutoscalingReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			telemetryv1.AutoscalingReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}

	// Check the observed Generation and mirror the condition from the
	// underlying resource reconciliation
	autoObsGen, err := r.checkAutoscalingGeneration(instance)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.AutoscalingReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			telemetryv1.AutoscalingReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, nil
	}
	if !autoObsGen {
		instance.Status.Conditions.Set(condition.UnknownCondition(
			telemetryv1.AutoscalingReadyCondition,
			condition.InitReason,
			telemetryv1.AutoscalingReadyRunningMessage,
		))
	} else {
		// Mirror Autoscaling condition status
		c := autoscalingInstance.Status.Conditions.Mirror(telemetryv1.AutoscalingReadyCondition)
		if c != nil {
			instance.Status.Conditions.Set(c)
		}
	}
	if op != controllerutil.OperationResultNone && autoObsGen {
		helper.GetLogger().Info(fmt.Sprintf("%s %s - %s", autoscalingName, autoscalingInstance.Name, op))
	}

	return ctrl.Result{}, nil
}

// reconcileMetricStorage ...
func (r TelemetryReconciler) reconcileMetricStorage(ctx context.Context, instance *telemetryv1.Telemetry, helper *helper.Helper) (ctrl.Result, error) {
	const (
		metricStorageNamespaceLabel = "MetricStorage.Namespace"
		metricStorageNameLabel      = "MetricStorage.Name"
		metricStorageName           = telemetryv1.DefaultServiceName
	)
	metricStorageInstance := &telemetryv1.MetricStorage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      telemetryv1.DefaultServiceName,
			Namespace: instance.Namespace,
		},
	}

	if instance.Spec.MetricStorage.Enabled == nil || !*instance.Spec.MetricStorage.Enabled {
		if res, err := utils.EnsureDeleted(ctx, helper, metricStorageInstance); err != nil {
			return res, err
		}
		instance.Status.Conditions.Remove(telemetryv1.MetricStorageReadyCondition)
		return ctrl.Result{}, nil
	}

	helper.GetLogger().Info("Reconciling MetricStorage", metricStorageNamespaceLabel, instance.Namespace, metricStorageNameLabel, telemetryv1.DefaultServiceName)
	op, err := controllerutil.CreateOrPatch(ctx, helper.GetClient(), metricStorageInstance, func() error {
		instance.Spec.MetricStorage.MetricStorageSpec.DeepCopyInto(&metricStorageInstance.Spec)

		err := controllerutil.SetControllerReference(helper.GetBeforeObject(), metricStorageInstance, helper.GetScheme())
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.MetricStorageReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			telemetryv1.MetricStorageReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}

	// Check the observed Generation and mirror the condition from the underlying
	// resource reconciliation
	msObsGen, err := r.checkMetricsStorageGeneration(instance)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.MetricStorageReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			telemetryv1.MetricStorageReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, nil
	}

	if !msObsGen {
		instance.Status.Conditions.Set(condition.UnknownCondition(
			telemetryv1.MetricStorageReadyCondition,
			condition.InitReason,
			telemetryv1.MetricStorageReadyRunningMessage,
		))
	} else {
		// Mirror MetricsStorage's condition status
		c := metricStorageInstance.Status.Conditions.Mirror(telemetryv1.MetricStorageReadyCondition)
		if c != nil {
			instance.Status.Conditions.Set(c)
		}
	}

	if op != controllerutil.OperationResultNone && msObsGen {
		helper.GetLogger().Info(fmt.Sprintf("%s %s - %s", telemetryv1.DefaultServiceName, metricStorageInstance.Name, op))
	}

	return ctrl.Result{}, nil
}

// reconcileLogging ...
func (r TelemetryReconciler) reconcileLogging(ctx context.Context, instance *telemetryv1.Telemetry, helper *helper.Helper) (ctrl.Result, error) {
	const (
		loggingNamespaceLabel = "Logging.Namespace"
		loggingNameLabel      = "Logging.Name"
	)
	loggingInstance := &telemetryv1.Logging{
		ObjectMeta: metav1.ObjectMeta{
			Name:      logging.ServiceName,
			Namespace: instance.Namespace,
		},
	}

	if instance.Spec.Logging.Enabled == nil || !*instance.Spec.Logging.Enabled {
		if res, err := utils.EnsureDeleted(ctx, helper, loggingInstance); err != nil {
			return res, err
		}
		instance.Status.Conditions.Remove(telemetryv1.LoggingReadyCondition)
		return ctrl.Result{}, nil
	}

	helper.GetLogger().Info("Reconciling Logging", loggingNamespaceLabel, instance.Namespace, loggingNameLabel, logging.ServiceName)
	op, err := controllerutil.CreateOrPatch(ctx, helper.GetClient(), loggingInstance, func() error {
		instance.Spec.Logging.LoggingSpec.DeepCopyInto(&loggingInstance.Spec)

		err := controllerutil.SetControllerReference(helper.GetBeforeObject(), loggingInstance, helper.GetScheme())
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.LoggingReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			telemetryv1.LoggingReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}

	// Check the observed Generation and mirror the condition from the underlying
	// resource reconciliation
	logObsGen, err := r.checkLoggingGeneration(instance)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.LoggingReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			telemetryv1.LoggingReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, nil
	}
	if !logObsGen {
		instance.Status.Conditions.Set(condition.UnknownCondition(
			telemetryv1.LoggingReadyCondition,
			condition.InitReason,
			telemetryv1.LoggingReadyRunningMessage,
		))
	} else {

		// Mirror Logging's condition status
		c := loggingInstance.Status.Conditions.Mirror(telemetryv1.LoggingReadyCondition)
		if c != nil {
			instance.Status.Conditions.Set(c)
		}
	}
	if op != controllerutil.OperationResultNone && logObsGen {
		helper.GetLogger().Info(fmt.Sprintf("%s %s - %s", logging.ServiceName, loggingInstance.Name, op))
	}
	return ctrl.Result{}, nil
}

// reconcileAutoscaling ...
func (r TelemetryReconciler) reconcileCloudKitty(ctx context.Context, instance *telemetryv1.Telemetry, helper *helper.Helper) (ctrl.Result, error) {
	const (
		cloudKittyNamespaceLabel = "CloudKitty.Namespace"
		cloudKittyNameLabel      = "CloudKitty.Name"
		cloudKittyName           = "cloudkitty"
	)
	cloudKittyInstance := &telemetryv1.CloudKitty{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cloudKittyName,
			Namespace: instance.Namespace,
		},
	}

	if instance.Spec.CloudKitty.Enabled == nil || !*instance.Spec.CloudKitty.Enabled {
		if res, err := utils.EnsureDeleted(ctx, helper, cloudKittyInstance); err != nil {
			return res, err
		}
		instance.Status.Conditions.Remove(telemetryv1.CloudKittyReadyCondition)
		return ctrl.Result{}, nil
	}

	if instance.Spec.CloudKitty.NodeSelector == nil {
		instance.Spec.CloudKitty.NodeSelector = instance.Spec.NodeSelector
	}

	if instance.Spec.CloudKitty.TopologyRef == nil {
		instance.Spec.CloudKitty.TopologyRef = instance.Spec.TopologyRef
	}

	helper.GetLogger().Info("Reconciling CloudKitty", cloudKittyNamespaceLabel, instance.Namespace, cloudKittyNameLabel, cloudKittyName)
	op, err := controllerutil.CreateOrPatch(ctx, helper.GetClient(), cloudKittyInstance, func() error {
		instance.Spec.CloudKitty.CloudKittySpec.DeepCopyInto(&cloudKittyInstance.Spec)

		err := controllerutil.SetControllerReference(helper.GetBeforeObject(), cloudKittyInstance, helper.GetScheme())
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CloudKittyReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			telemetryv1.CloudKittyReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}

	// Check the observed Generation and mirror the condition from the
	// underlying resource reconciliation
	autoObsGen, err := r.checkCloudKittyGeneration(instance)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CloudKittyReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			telemetryv1.CloudKittyReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, nil
	}
	if !autoObsGen {
		instance.Status.Conditions.Set(condition.UnknownCondition(
			telemetryv1.CloudKittyReadyCondition,
			condition.InitReason,
			telemetryv1.AutoscalingReadyRunningMessage,
		))
	} else {
		// Mirror Autoscaling condition status
		c := cloudKittyInstance.Status.Conditions.Mirror(telemetryv1.CloudKittyReadyCondition)
		if c != nil {
			instance.Status.Conditions.Set(c)
		}
	}
	if op != controllerutil.OperationResultNone && autoObsGen {
		helper.GetLogger().Info(fmt.Sprintf("%s %s - %s", cloudKittyName, cloudKittyInstance.Name, op))
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TelemetryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1.Telemetry{}).
		Owns(&telemetryv1.Ceilometer{}).
		Owns(&telemetryv1.Autoscaling{}).
		Owns(&telemetryv1.MetricStorage{}).
		Owns(&telemetryv1.Logging{}).
		Owns(&telemetryv1.CloudKitty{}).
		Complete(r)
}

// checkCeilometerGeneration -
func (r *TelemetryReconciler) checkCeilometerGeneration(
	instance *telemetryv1.Telemetry,
) (bool, error) {
	Log := r.GetLogger(context.Background())
	clm := &telemetryv1.CeilometerList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
	}
	if err := r.Client.List(context.Background(), clm, listOpts...); err != nil {
		Log.Error(err, "Unable to retrieve Ceilometer CR %w")
		return false, err
	}
	for _, item := range clm.Items {
		if item.Generation != item.Status.ObservedGeneration {
			return false, nil
		}
	}
	return true, nil
}

// checkAutoscalingGeneration -
func (r *TelemetryReconciler) checkAutoscalingGeneration(
	instance *telemetryv1.Telemetry,
) (bool, error) {
	Log := r.GetLogger(context.Background())
	as := &telemetryv1.AutoscalingList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
	}
	if err := r.Client.List(context.Background(), as, listOpts...); err != nil {
		Log.Error(err, "Unable to retrieve Autoscaling CR %w")
		return false, err
	}
	for _, item := range as.Items {
		if item.Generation != item.Status.ObservedGeneration {
			return false, nil
		}
	}
	return true, nil
}

// checkMetricsStorageGeneration -
func (r *TelemetryReconciler) checkMetricsStorageGeneration(
	instance *telemetryv1.Telemetry,
) (bool, error) {
	Log := r.GetLogger(context.Background())
	ms := &telemetryv1.MetricStorageList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
	}
	if err := r.Client.List(context.Background(), ms, listOpts...); err != nil {
		Log.Error(err, "Unable to retrieve MetricsStorage CR %w")
		return false, err
	}
	for _, item := range ms.Items {
		if item.Generation != item.Status.ObservedGeneration {
			return false, nil
		}
	}
	return true, nil
}

// checkLoggingGeneration -
func (r *TelemetryReconciler) checkLoggingGeneration(
	instance *telemetryv1.Telemetry,
) (bool, error) {
	Log := r.GetLogger(context.Background())
	l := &telemetryv1.LoggingList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
	}
	if err := r.Client.List(context.Background(), l, listOpts...); err != nil {
		Log.Error(err, "Unable to retrieve Logging CR %w")
		return false, err
	}
	for _, item := range l.Items {
		if item.Generation != item.Status.ObservedGeneration {
			return false, nil
		}
	}
	return true, nil
}

// checkCloudKittyGeneration -
func (r *TelemetryReconciler) checkCloudKittyGeneration(
	instance *telemetryv1.Telemetry,
) (bool, error) {
	Log := r.GetLogger(context.Background())
	clm := &telemetryv1.CloudKittyList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
	}
	if err := r.Client.List(context.Background(), clm, listOpts...); err != nil {
		Log.Error(err, "Unable to retrieve CloudKitty CR %w")
		return false, err
	}
	for _, item := range clm.Items {
		if item.Generation != item.Status.ObservedGeneration {
			return false, nil
		}
	}
	return true, nil
}
