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
	"os"
	"path"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	logr "github.com/go-logr/logr"
	common "github.com/openstack-k8s-operators/lib-common/modules/common"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	common_rbac "github.com/openstack-k8s-operators/lib-common/modules/common/rbac"
	util "github.com/openstack-k8s-operators/lib-common/modules/common/util"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	ansibleeev1 "github.com/openstack-k8s-operators/openstack-ansibleee-operator/api/v1alpha1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	infracompute "github.com/openstack-k8s-operators/telemetry-operator/pkg/infracompute"
	telemetry "github.com/openstack-k8s-operators/telemetry-operator/pkg/telemetry"
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
// service account, role, rolebinding
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;list;watch;create;update
// service account permissions that are needed to grant permission to the above
// +kubebuilder:rbac:groups="security.openshift.io",resourceNames=anyuid,resources=securitycontextconstraints,verbs=use
// +kubebuilder:rbac:groups="",resources=pods,verbs=create;delete;get;list;patch;update;watch

// Reconcile reconciles an InfraCompute
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
			condition.UnknownCondition(condition.AnsibleEECondition, condition.InitReason, condition.AnsibleEEReadyInitMessage),
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

	// Create a configmap with the playbooks that will be run
	err = r.ensurePlaybooks(ctx, helper, instance)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.ServiceConfigReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.ServiceConfigReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}
	// end telemetry playbooks

	// Create the extravars configmap
	err = r.ensureExtravars(ctx, helper, instance)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.ServiceConfigReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.ServiceConfigReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}
	// end create extravars configmap

	serviceLabels := map[string]string{
		common.AppSelector: infracompute.ServiceName,
	}

	_, err = r.createAnsibleExecution(ctx, instance, serviceLabels)
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

	instance.Status.Conditions.MarkTrue(condition.AnsibleEECondition, condition.AnsibleEEReadyMessage)

	fmt.Println("Returning...")
	return ctrl.Result{}, nil
}

func (r *InfraComputeReconciler) ensurePlaybooks(ctx context.Context, h *helper.Helper, instance *telemetryv1.InfraCompute) error {
	playbooksPath, found := os.LookupEnv("OPERATOR_PLAYBOOKS")
	if !found {
		playbooksPath = "playbooks"
		os.Setenv("OPERATOR_PLAYBOOKS", playbooksPath)
		util.LogForObject(h, "OPERATOR_PLAYBOOKS not set in env when reconciling ", instance, "defaulting to ", playbooksPath)
	}

	util.LogForObject(h, "using playbooks for instance ", instance, "from ", playbooksPath)

	playbookDirEntries, err := os.ReadDir(playbooksPath)
	if err != nil {
		return err
	}
	// EnsureConfigMaps is not used as we do not want the templating behavior that adds.
	playbookCMName := fmt.Sprintf("%s-compute-playbooks", telemetry.ServiceName)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        playbookCMName,
			Namespace:   instance.Namespace,
			Annotations: map[string]string{},
		},
		Data: map[string]string{},
	}
	_, err = controllerutil.CreateOrPatch(ctx, h.GetClient(), configMap, func() error {
		for _, entry := range playbookDirEntries {
			filename := entry.Name()
			filePath := path.Join(playbooksPath, filename)
			data, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}
			configMap.Data[filename] = string(data)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error create/updating configmap: %w", err)
	}

	return nil
}

func (r *InfraComputeReconciler) ensureExtravars(ctx context.Context, h *helper.Helper, instance *telemetryv1.InfraCompute) error {
	data := make(map[string]string)
	data["extravars"] = fmt.Sprintf("telemetry_node_exporter_image: %s", instance.Spec.NodeExporterImage)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-compute-extravars", telemetry.ServiceName),
			Namespace: instance.Namespace,
		},
		Data: data,
	}
	_, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), configMap, func() error { return nil })
	if err != nil {
		return fmt.Errorf("error create/updating configmap: %w", err)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InfraComputeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1.InfraCompute{}).
		Owns(&ansibleeev1.OpenStackAnsibleEE{}).
		Complete(r)
}
