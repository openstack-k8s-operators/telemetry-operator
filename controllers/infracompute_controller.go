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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	logr "github.com/go-logr/logr"
	common "github.com/openstack-k8s-operators/lib-common/modules/common"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	ansibleeev1 "github.com/openstack-k8s-operators/openstack-ansibleee-operator/api/v1alpha1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	infracompute "github.com/openstack-k8s-operators/telemetry-operator/pkg/infracompute"
)

// InfraComputeReconciler reconciles a InfraCompute object
type InfraComputeReconciler struct {
	client.Client
	Kclient kubernetes.Interface
	Log     logr.Logger
	Scheme  *runtime.Scheme
}

//+kubebuilder:rbac:groups=telemetry.openstack.org,resources=infracomputes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=telemetry.openstack.org,resources=infracomputes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=telemetry.openstack.org,resources=infracomputes/finalizers,verbs=update
//+kubebuilder:rbac:groups=ansibleee.openstack.org,resources=openstackansibleees,verbs=get;list;watch;create;update;patch;delete;

func (r *InfraComputeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, _err error) {
	_ = r.Log.WithValues("infra-compute", req.NamespacedName)

	// Fetch the InfraCompute instance
	instance := &telemetryv1.InfraCompute{}
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
			condition.UnknownCondition(condition.InputReadyCondition, condition.InitReason, condition.InputReadyInitMessage),
			condition.UnknownCondition(condition.ServiceConfigReadyCondition, condition.InitReason, condition.ServiceConfigReadyInitMessage),
			condition.UnknownCondition(condition.DeploymentReadyCondition, condition.InitReason, condition.DeploymentReadyInitMessage),
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

func (r *InfraComputeReconciler) reconcileDelete(ctx context.Context, instance *telemetryv1.InfraCompute, helper *helper.Helper) (ctrl.Result, error) {
	r.Log.Info("Reconciling Service delete")

	// Service is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(instance, helper.GetFinalizer())

	// Delete the openstack-ansibleee when the infracompute is deleted
	ansibleee := &ansibleeev1.OpenStackAnsibleEE{}
	err := r.Get(ctx, types.NamespacedName{Name: infracompute.ServiceName, Namespace: instance.Namespace}, ansibleee)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = helper.GetClient().Delete(ctx, ansibleee)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *InfraComputeReconciler) reconcileNormal(ctx context.Context, instance *telemetryv1.InfraCompute, helper *helper.Helper) (ctrl.Result, error) {
	r.Log.Info(fmt.Sprintf("Reconciling Service '%s'", instance.Name))

	instance.Status.Conditions.MarkTrue(condition.ServiceConfigReadyCondition, condition.ServiceConfigReadyMessage)

	serviceLabels := map[string]string{
		common.AppSelector: infracompute.ServiceName,
	}

	_, err := r.createAnsibleExecution(ctx, instance, serviceLabels)
	if err != nil {
		return ctrl.Result{}, err
	}

	r.Log.Info("Reconciled Service successfully")
	return ctrl.Result{}, nil
}

func (r *InfraComputeReconciler) createAnsibleExecution(ctx context.Context, instance *telemetryv1.InfraCompute,
	labels map[string]string) (ctrl.Result, error) {
	// Define a new AnsibleEE
	ansibleEE, err := infracompute.AnsibleEE(instance, labels)
	if err != nil {
		fmt.Println("Error defining new ansibleEE")
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.AnsibleEECondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.AnsibleEEReadyErrorMessage,
			err.Error(),
		))
		return ctrl.Result{}, err
	}

	err = r.Client.Get(ctx, types.NamespacedName{Namespace: ansibleEE.Namespace, Name: ansibleEE.Name}, ansibleEE)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			err = r.Create(ctx, ansibleEE)
			if err != nil {
				fmt.Println("Error creating ansibleEE")
				instance.Status.Conditions.Set(condition.FalseCondition(
					condition.AnsibleEECondition,
					condition.ErrorReason,
					condition.SeverityWarning,
					condition.AnsibleEEReadyErrorMessage,
					err.Error(),
				))
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	fmt.Println("Returning...")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InfraComputeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1.InfraCompute{}).
		Owns(&ansibleeev1.OpenStackAnsibleEE{}).
		Complete(r)
}
