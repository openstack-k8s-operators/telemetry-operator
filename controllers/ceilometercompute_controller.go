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

	logr "github.com/go-logr/logr"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	common "github.com/openstack-k8s-operators/lib-common/modules/common"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	configmap "github.com/openstack-k8s-operators/lib-common/modules/common/configmap"
	env "github.com/openstack-k8s-operators/lib-common/modules/common/env"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	util "github.com/openstack-k8s-operators/lib-common/modules/common/util"

	ansibleeev1 "github.com/openstack-k8s-operators/openstack-ansibleee-operator/api/v1alpha1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	ceilometercompute "github.com/openstack-k8s-operators/telemetry-operator/pkg/ceilometercompute"
)

// CeilometerComputeReconciler reconciles a Ceilometer object
type CeilometerComputeReconciler struct {
	client.Client
	Kclient kubernetes.Interface
	Log     logr.Logger
	Scheme  *runtime.Scheme
}

// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometercomputes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometercomputes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=ceilometercomputes/finalizers,verbs=update
// +kubebuilder:rbac:groups=ansibleee.openstack.org,resources=openstackansibleees,verbs=get;list;watch;create;update;patch;delete;

// Reconcile reconciles a CeilometerCompute
func (r *CeilometerComputeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, _err error) {
	_ = r.Log.WithValues("ceilometer-compute", req.NamespacedName)

	// Fetch the CeilometerCompute instance
	instance := &telemetryv1.CeilometerCompute{}
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

func (r *CeilometerComputeReconciler) reconcileDelete(ctx context.Context, instance *telemetryv1.CeilometerCompute, helper *helper.Helper) (ctrl.Result, error) {
	r.Log.Info("Reconciling Service delete")

	// Service is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(instance, helper.GetFinalizer())

	// Delete the openstack-ansibleee when the ceilometercompute is deleted
	ansibleee := &ansibleeev1.OpenStackAnsibleEE{}
	err := r.Get(ctx, types.NamespacedName{Name: ceilometercompute.ServiceName, Namespace: instance.Namespace}, ansibleee)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = helper.GetClient().Delete(ctx, ansibleee)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CeilometerComputeReconciler) reconcileNormal(ctx context.Context, instance *telemetryv1.CeilometerCompute, helper *helper.Helper) (ctrl.Result, error) {
	r.Log.Info(fmt.Sprintf("Reconciling Service '%s'", instance.Name))
	// ConfigMap
	configMapVars := make(map[string]env.Setter)

	/*rabbitSecret, hash, err := oko_secret.GetSecret(ctx, helper, instance.Spec.RabbitMQSecret, instance.Namespace)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			instance.Status.Conditions.Set(condition.FalseCondition(
				condition.InputReadyCondition,
				condition.RequestedReason,
				condition.SeverityInfo,
				condition.InputReadyWaitingMessage))
			return ctrl.Result{RequeueAfter: time.Duration(10) * time.Second}, fmt.Errorf("RabbitMQ secret %s not found", instance.Spec.RabbitMQSecret)
		}
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.InputReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.InputReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}

	configMapVars[rabbitSecret.Name] = env.SetValue(hash)
	instance.Status.Conditions.MarkTrue(condition.InputReadyCondition, condition.InputReadyMessage)*/

	//
	// create Configmap required for ceilometer input
	// - %-scripts configmap holding scripts to e.g. bootstrap the service
	// - %-config configmap holding minimal ceilometer config required to get the service up, user can add additional files to be added to the service
	// - parameters which has passwords gets added from the OpenStack secret via the init container
	//
	err := r.generateServiceConfigMaps(ctx, helper, instance, &configMapVars)
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
	/*inputHash, hashChanged, err := r.createHashOfInputHashes(ctx, instance, configMapVars)
	fmt.Printf("hashChanged: %v\n", hashChanged)
	if err != nil {
		return ctrl.Result{}, err
	}*/
	/*if err != nil {
		return ctrl.Result{}, err
	} else if hashChanged {
		// Hash changed and instance status should be updated (which will be done by main defer func),
		// so we need to return and reconcile again
		return ctrl.Result{}, nil
	}
	fmt.Printf("MarkTrue\n")*/
	instance.Status.Conditions.MarkTrue(condition.ServiceConfigReadyCondition, condition.ServiceConfigReadyMessage)

	serviceLabels := map[string]string{
		common.AppSelector: ceilometercompute.ServiceName,
	}

	fmt.Println("Creating ansible execution")
	_, err = r.createAnsibleExecution(ctx, instance, serviceLabels)
	if err != nil {
		return ctrl.Result{}, err
	}

	r.Log.Info("Reconciled Service successfully")
	return ctrl.Result{}, nil
}

func (r *CeilometerComputeReconciler) createAnsibleExecution(ctx context.Context, instance *telemetryv1.CeilometerCompute,
	labels map[string]string) (ctrl.Result, error) {
	// Define a new AnsibleEE
	ansibleEE, err := ceilometercompute.AnsibleEE(instance, labels)
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

// getSecret - get the specified secret, and add its hash to envVars
/* func (r *CeilometerComputeReconciler) getSecret(ctx context.Context, h *helper.Helper, instance *telemetryv1.CeilometerCompute, secretName string, envVars *map[string]env.Setter) (ctrl.Result, error) {
	secret, hash, err := secret.GetSecret(ctx, h, secretName, instance.Namespace)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			instance.Status.Conditions.Set(condition.FalseCondition(
				condition.InputReadyCondition,
				condition.RequestedReason,
				condition.SeverityInfo,
				condition.InputReadyWaitingMessage))
			return ctrl.Result{RequeueAfter: time.Duration(10) * time.Second}, fmt.Errorf("Secret %s not found", secretName)
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
} */

func (r *CeilometerComputeReconciler) generateServiceConfigMaps(
	ctx context.Context,
	h *helper.Helper,
	instance *telemetryv1.CeilometerCompute,
	envVars *map[string]env.Setter,
) error {

	/* cmLabels := labels.GetLabels(instance, labels.GetGroupLabel(ceilometercompute.ServiceName), map[string]string{})
	customData := map[string]string{common.CustomServiceConfigFileName: instance.Spec.CustomServiceConfig}
	for key, data := range instance.Spec.DefaultConfigOverwrite {
		customData[key] = data
	}

	templateParameters := make(map[string]interface{}) */

	cms := []util.Template{
		// ScriptsConfigMap
		/* {
			Name:               fmt.Sprintf("%s-scripts", ceilometercompute.ServiceName),
			Namespace:          instance.Namespace,
			Type:               util.TemplateTypeScripts,
			InstanceType:       instance.Kind,
			AdditionalTemplate: map[string]string{"common.sh": "/common/common.sh"},
			Labels:             cmLabels,
		},
		// ConfigMap
		{
			Name:          fmt.Sprintf("%s-config-data", ceilometercompute.ServiceName),
			Namespace:     instance.Namespace,
			Type:          util.TemplateTypeConfig,
			InstanceType:  instance.Kind,
			CustomData:    customData,
			ConfigOptions: templateParameters,
			Labels:        cmLabels,
		}, */
	}
	return configmap.EnsureConfigMaps(ctx, h, instance, cms, envVars)
}

// createHashOfInputHashes - creates a hash of hashes which gets added to the resources which requires a restart
// if any of the input resources change, like configs, passwords, ...
//
// returns the hash, whether the hash changed (as a bool) and any error
/* func (r *CeilometerComputeReconciler) createHashOfInputHashes(
	ctx context.Context,
	instance *telemetryv1.CeilometerCompute,
	envVars map[string]env.Setter,
) (string, bool, error) {
	var hashMap map[string]string
	changed := false
	mergedMapVars := env.MergeEnvs([]corev1.EnvVar{}, envVars)
	hash, err := util.ObjectHash(mergedMapVars)
	if err != nil {
		return hash, changed, err
	}
	if hashMap, changed = util.SetHash(instance.Status.Hash, common.InputHashName, hash); changed {
		instance.Status.Hash = hashMap
		r.Log.Info(fmt.Sprintf("Input maps hash %s - %s", common.InputHashName, hash))
	}
	return hash, changed, nil
} */

// SetupWithManager sets up the controller with the Manager.
func (r *CeilometerComputeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1.CeilometerCompute{}).
		Owns(&ansibleeev1.OpenStackAnsibleEE{}).
		Complete(r)
}
