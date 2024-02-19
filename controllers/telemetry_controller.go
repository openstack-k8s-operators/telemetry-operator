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
	common "github.com/openstack-k8s-operators/lib-common/modules/common"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"k8s.io/client-go/kubernetes"

	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	ceilometer "github.com/openstack-k8s-operators/telemetry-operator/pkg/ceilometer"
	logging "github.com/openstack-k8s-operators/telemetry-operator/pkg/logging"
	telemetry "github.com/openstack-k8s-operators/telemetry-operator/pkg/telemetry"
	obov1 "github.com/rhobs/observability-operator/pkg/apis/monitoring/v1alpha1"
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
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=telemetries/finalizers,verbs=update
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=autoscalings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=autoscalings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=autoscalings/finalizers,verbs=update;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometers/finalizers,verbs=update;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=metricstorages,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=metricstorages/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=metricstorages/finalizers,verbs=update;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=loggings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=loggings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=loggings/finalizers,verbs=update;delete
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
			_err = err
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
			condition.UnknownCondition(telemetryv1.CeilometerReadyCondition, condition.InitReason, telemetryv1.CeilometerReadyInitMessage),
			condition.UnknownCondition(telemetryv1.AutoscalingReadyCondition, condition.InitReason, telemetryv1.AutoscalingReadyInitMessage),
			condition.UnknownCondition(telemetryv1.MetricStorageReadyCondition, condition.InitReason, telemetryv1.MetricStorageReadyInitMessage),
			condition.UnknownCondition(telemetryv1.LoggingReadyCondition, condition.InitReason, telemetryv1.LoggingReadyInitMessage),
		)

		instance.Status.Conditions.Init(&cl)

		// Register overall status immediately to have an early feedback e.g. in the cli
		return ctrl.Result{}, nil
	}

	if instance.Status.Hash == nil {
		instance.Status.Hash = map[string]string{}
	}

	serviceLabels := map[string]string{
		common.AppSelector: telemetry.ServiceName,
	}

	// Handle service init
	ctrlResult, err := r.reconcileInit(ctx, instance, helper, serviceLabels)
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
	instance *telemetryv1.Telemetry,
	helper *helper.Helper,
	serviceLabels map[string]string,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info("Reconciling Service init")

	Log.Info("Reconciled Service init successfully")
	return ctrl.Result{}, nil
}

func (r *TelemetryReconciler) reconcileNormal(ctx context.Context, instance *telemetryv1.Telemetry, helper *helper.Helper) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info(fmt.Sprintf("Reconciling Service '%s'", instance.Name))

	ctrlResult, err := reconcileAutoscaling(ctx, instance, helper)
	if err != nil {
		return ctrl.Result{}, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	ctrlResult, err = reconcileCeilometer(ctx, instance, helper)
	if err != nil {
		return ctrl.Result{}, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	ctrlResult, err = reconcileMetricStorage(ctx, instance, helper)
	if err != nil {
		return ctrl.Result{}, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	ctrlResult, err = reconcileLogging(ctx, instance, helper)
	if err != nil {
		return ctrl.Result{}, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	Log.Info("Reconciled Service successfully")
	return ctrl.Result{}, nil
}

// ensureDeleted - Delete the object which in turn will clean the sub resources
func ensureDeleted(ctx context.Context, helper *helper.Helper, obj client.Object) (ctrl.Result, error) {
	key := client.ObjectKeyFromObject(obj)
	if err := helper.GetClient().Get(ctx, key, obj); err != nil {
		if k8s_errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	// Delete the object
	if obj.GetDeletionTimestamp().IsZero() {
		if err := helper.GetClient().Delete(ctx, obj); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil

}

// reconcileCeilometer ...
func reconcileCeilometer(ctx context.Context, instance *telemetryv1.Telemetry, helper *helper.Helper) (ctrl.Result, error) {
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

	if !instance.Spec.Ceilometer.Enabled {
		if res, err := ensureDeleted(ctx, helper, ceilometerInstance); err != nil {
			return res, err
		}
		instance.Status.Conditions.Remove(telemetryv1.CeilometerReadyCondition)
		return ctrl.Result{}, nil
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
	if op != controllerutil.OperationResultNone {
		helper.GetLogger().Info(fmt.Sprintf("%s %s - %s", ceilometer.ServiceName, ceilometerInstance.Name, op))
	}

	if ceilometerInstance.IsReady() {
		instance.Status.Conditions.MarkTrue(telemetryv1.CeilometerReadyCondition, telemetryv1.CeilometerReadyMessage)
	} else {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CeilometerReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			telemetryv1.CeilometerReadyRunningMessage))
	}

	return ctrl.Result{}, nil
}

// reconcileAutoscaling ...
func reconcileAutoscaling(ctx context.Context, instance *telemetryv1.Telemetry, helper *helper.Helper) (ctrl.Result, error) {
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

	if !instance.Spec.Autoscaling.Enabled {
		if res, err := ensureDeleted(ctx, helper, autoscalingInstance); err != nil {
			return res, err
		}
		instance.Status.Conditions.Remove(telemetryv1.AutoscalingReadyCondition)
		return ctrl.Result{}, nil
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
	if op != controllerutil.OperationResultNone {
		helper.GetLogger().Info(fmt.Sprintf("%s %s - %s", autoscalingName, autoscalingInstance.Name, op))
	}

	if autoscalingInstance.IsReady() {
		instance.Status.Conditions.MarkTrue(telemetryv1.AutoscalingReadyCondition, telemetryv1.AutoscalingReadyMessage)
	} else {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.AutoscalingReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			telemetryv1.AutoscalingReadyRunningMessage))
	}

	return ctrl.Result{}, nil
}

// reconcileMetricStorage ...
func reconcileMetricStorage(ctx context.Context, instance *telemetryv1.Telemetry, helper *helper.Helper) (ctrl.Result, error) {
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

	if !instance.Spec.MetricStorage.Enabled {
		if res, err := ensureDeleted(ctx, helper, metricStorageInstance); err != nil {
			return res, err
		}
		instance.Status.Conditions.Remove(telemetryv1.MetricStorageReadyCondition)
		return ctrl.Result{}, nil
	}

	helper.GetLogger().Info("Reconciling MetricStorage", metricStorageNamespaceLabel, instance.Namespace, metricStorageNameLabel, telemetryv1.DefaultServiceName)
	op, err := controllerutil.CreateOrPatch(ctx, helper.GetClient(), metricStorageInstance, func() error {
		if instance.Spec.MetricStorage.MetricStorageSpec.CustomMonitoringStack != nil {
			metricStorageInstance.Spec.CustomMonitoringStack = &obov1.MonitoringStackSpec{}
			instance.Spec.MetricStorage.MetricStorageSpec.CustomMonitoringStack.DeepCopyInto(metricStorageInstance.Spec.CustomMonitoringStack)
		}
		if instance.Spec.MetricStorage.MetricStorageSpec.MonitoringStack != nil {
			metricStorageInstance.Spec.MonitoringStack = &telemetryv1.MonitoringStack{}
			instance.Spec.MetricStorage.MetricStorageSpec.MonitoringStack.DeepCopyInto(metricStorageInstance.Spec.MonitoringStack)
		}

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
	if op != controllerutil.OperationResultNone {
		helper.GetLogger().Info(fmt.Sprintf("%s %s - %s", telemetryv1.DefaultServiceName, metricStorageInstance.Name, op))
	}

	if metricStorageInstance.IsReady() {
		instance.Status.Conditions.MarkTrue(telemetryv1.MetricStorageReadyCondition, telemetryv1.MetricStorageReadyMessage)
	} else {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.MetricStorageReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			telemetryv1.MetricStorageReadyRunningMessage))
	}

	return ctrl.Result{}, nil
}

// reconcileLogging ...
func reconcileLogging(ctx context.Context, instance *telemetryv1.Telemetry, helper *helper.Helper) (ctrl.Result, error) {
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

	if !instance.Spec.Logging.Enabled {
		if res, err := ensureDeleted(ctx, helper, loggingInstance); err != nil {
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
	if op != controllerutil.OperationResultNone {
		helper.GetLogger().Info(fmt.Sprintf("%s %s - %s", logging.ServiceName, loggingInstance.Name, op))
	}

	if loggingInstance.IsReady() {
		instance.Status.Conditions.MarkTrue(telemetryv1.LoggingReadyCondition, telemetryv1.LoggingReadyMessage)
	} else {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.LoggingReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			telemetryv1.LoggingReadyRunningMessage))
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
		Complete(r)
}
