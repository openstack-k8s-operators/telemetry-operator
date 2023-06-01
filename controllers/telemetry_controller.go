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

	logr "github.com/go-logr/logr"
	common "github.com/openstack-k8s-operators/lib-common/modules/common"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	"github.com/openstack-k8s-operators/lib-common/modules/common/configmap"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	util "github.com/openstack-k8s-operators/lib-common/modules/common/util"
	"k8s.io/client-go/kubernetes"

	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	telemetry "github.com/openstack-k8s-operators/telemetry-operator/pkg/telemetry"
)

// TelemetryReconciler reconciles a Telemetry object
type TelemetryReconciler struct {
	client.Client
	Kclient kubernetes.Interface
	Log     logr.Logger
	Scheme  *runtime.Scheme
}

// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=telemetries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=telemetries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=telemetries/finalizers,verbs=update
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometercentrals,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometercentrals/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometercentrals/finalizers,verbs=update;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometercomputes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometercomputes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometercomputes/finalizers,verbs=update
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=infracomputes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=infracomputes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=infracomputes/finalizers,verbs=update
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=infracomputes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=infracomputes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=infracomputes/finalizers,verbs=update
// +kubebuilder:rbac:groups=rabbitmq.openstack.org,resources=transporturls,verbs=get;list;watch;create;update;patch;delete

// Reconcile reconciles a Telemetry
func (r *TelemetryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, _err error) {
	_ = r.Log.WithValues("telemetry", req.NamespacedName)

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
		r.Log,
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Always patch the instance status when exiting this function so we can persist any changes.
	defer func() {
		// update the overall status condition if service is ready
		if instance.IsReady() {
			instance.Status.Conditions.MarkTrue(condition.ReadyCondition, condition.ReadyMessage)
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
			condition.UnknownCondition(telemetryv1.TelemetryRabbitMqTransportURLReadyCondition, condition.InitReason, telemetryv1.TelemetryRabbitMqTransportURLReadyInitMessage),
			condition.UnknownCondition(telemetryv1.CeilometerCentralReadyCondition, condition.InitReason, telemetryv1.CeilometerCentralReadyInitMessage),
			condition.UnknownCondition(telemetryv1.CeilometerComputeReadyCondition, condition.InitReason, telemetryv1.CeilometerComputeReadyInitMessage),
			// service account, role, rolebinding conditions
			condition.UnknownCondition(condition.ServiceAccountReadyCondition, condition.InitReason, condition.ServiceAccountReadyInitMessage),
			condition.UnknownCondition(condition.RoleReadyCondition, condition.InitReason, condition.RoleReadyInitMessage),
			condition.UnknownCondition(condition.RoleBindingReadyCondition, condition.InitReason, condition.RoleBindingReadyInitMessage),
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
	r.Log.Info(fmt.Sprintf("Reconciled Service '%s' delete", instance.Name))

	// remove playbookCM
	playbookCMName := fmt.Sprintf("%s-compute-playbooks", telemetry.ServiceName)
	playbookCM, _, err := configmap.GetConfigMapAndHashWithName(ctx, helper, playbookCMName, instance.Namespace)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if err == nil {
		if err = helper.GetClient().Delete(ctx, playbookCM); err != nil && !k8s_errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		util.LogForObject(helper, "Removed our playbook configmap", instance)
	}
	// end remove playbookCM

	// remove extravarsCM
	extravarsCMName := fmt.Sprintf("%s-compute-extravars", telemetry.ServiceName)
	extravarsCM, _, err := configmap.GetConfigMapAndHashWithName(ctx, helper, extravarsCMName, instance.Namespace)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if err == nil {
		if err = helper.GetClient().Delete(ctx, extravarsCM); err != nil && !k8s_errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		util.LogForObject(helper, "Removed our extravars configmap", instance)
	}
	// end remove extravarsCM

	// Service is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(instance, helper.GetFinalizer())

	r.Log.Info(fmt.Sprintf("Reconciled Service '%s' delete successfully", instance.Name))
	return ctrl.Result{}, nil
}

func (r *TelemetryReconciler) reconcileInit(
	ctx context.Context,
	instance *telemetryv1.Telemetry,
	helper *helper.Helper,
	serviceLabels map[string]string,
) (ctrl.Result, error) {
	r.Log.Info("Reconciling Service init")

	r.Log.Info("Reconciled Service init successfully")
	return ctrl.Result{}, nil
}

func (r *TelemetryReconciler) reconcileNormal(ctx context.Context, instance *telemetryv1.Telemetry, helper *helper.Helper) (ctrl.Result, error) {
	r.Log.Info(fmt.Sprintf("Reconciling Service '%s'", instance.Name))

	// deploy ceilometercentral
	ceilometercentral, op, err := r.ceilometerCentralCreateOrUpdate(instance)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CeilometerCentralReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			telemetryv1.CeilometerCentralReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		r.Log.Info(fmt.Sprintf("Deployment %s successfully reconciled - operation: %s", instance.Name, string(op)))
	}

	// Mirror ceilometercompute's status ReadyCount to this parent CR
	instance.Status.CeilometerCentralReadyCount = ceilometercentral.Status.ReadyCount

	// Mirror ceilometercompute's condition status
	ccentral := ceilometercentral.Status.Conditions.Mirror(telemetryv1.CeilometerCentralReadyCondition)
	if ccentral != nil {
		instance.Status.Conditions.Set(ccentral)
	}
	// end deploy ceilometercentral

	// deploy ceilometercompute
	ceilometercompute, op, err := r.ceilometerComputeCreateOrUpdate(instance)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.CeilometerComputeReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			telemetryv1.CeilometerComputeReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		r.Log.Info(fmt.Sprintf("Deployment %s successfully reconciled - operation: %s", instance.Name, string(op)))
	}

	// Mirror ceilometercompute's status ReadyCount to this parent CR
	instance.Status.CeilometerComputeReadyCount = ceilometercompute.Status.ReadyCount

	// Mirror ceilometercompute's condition status
	ccompute := ceilometercompute.Status.Conditions.Mirror(telemetryv1.CeilometerComputeReadyCondition)
	if ccompute != nil {
		instance.Status.Conditions.Set(ccompute)
	}
	// end deploy ceilometercompute

	// deploy infracompute
	infracompute, op, err := r.infraComputeCreateOrUpdate(instance)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.InfraComputeReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			telemetryv1.InfraComputeReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		r.Log.Info(fmt.Sprintf("Deployment %s successfully reconciled - operation: %s", instance.Name, string(op)))
	}

	// Mirror infracompute's status ReadyCount to this parent CR
	instance.Status.InfraComputeReadyCount = infracompute.Status.ReadyCount

	// Mirror ceilometercompute's condition status
	icompute := infracompute.Status.Conditions.Mirror(telemetryv1.InfraComputeReadyCondition)
	if icompute != nil {
		instance.Status.Conditions.Set(icompute)
	}
	// end deploy infracompute

	r.Log.Info("Reconciled Service successfully")
	return ctrl.Result{}, nil
}

func (r *TelemetryReconciler) ceilometerCentralCreateOrUpdate(instance *telemetryv1.Telemetry) (*telemetryv1.CeilometerCentral, controllerutil.OperationResult, error) {
	ccentral := &telemetryv1.CeilometerCentral{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-ceilometer-central", instance.Name),
			Namespace: instance.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, ccentral, func() error {
		ccentral.Spec = instance.Spec.CeilometerCentral

		err := controllerutil.SetControllerReference(instance, ccentral, r.Scheme)
		if err != nil {
			return err
		}

		return nil
	})

	return ccentral, op, err
}

func (r *TelemetryReconciler) ceilometerComputeCreateOrUpdate(instance *telemetryv1.Telemetry) (*telemetryv1.CeilometerCompute, controllerutil.OperationResult, error) {
	ccompute := &telemetryv1.CeilometerCompute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-ceilometer-compute", instance.Name),
			Namespace: instance.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, ccompute, func() error {
		ccompute.Spec = instance.Spec.CeilometerCompute

		err := controllerutil.SetControllerReference(instance, ccompute, r.Scheme)
		if err != nil {
			return err
		}

		return nil
	})

	r.Log.Info("Returning created ccompute")
	return ccompute, op, err
}

func (r *TelemetryReconciler) infraComputeCreateOrUpdate(instance *telemetryv1.Telemetry) (*telemetryv1.InfraCompute, controllerutil.OperationResult, error) {
	icompute := &telemetryv1.InfraCompute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-infra-compute", instance.Name),
			Namespace: instance.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, icompute, func() error {
		icompute.Spec = instance.Spec.InfraCompute

		err := controllerutil.SetControllerReference(instance, icompute, r.Scheme)
		if err != nil {
			return err
		}

		return nil
	})

	return icompute, op, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *TelemetryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1.Telemetry{}).
		Owns(&telemetryv1.CeilometerCentral{}).
		Owns(&telemetryv1.CeilometerCompute{}).
		Owns(&telemetryv1.InfraCompute{}).
		Complete(r)
}
