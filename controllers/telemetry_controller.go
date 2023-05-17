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
	"time"

	corev1 "k8s.io/api/core/v1"
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

	rabbitmqv1 "github.com/openstack-k8s-operators/infra-operator/apis/rabbitmq/v1beta1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	"github.com/openstack-k8s-operators/telemetry-operator/pkg/telemetry"
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
			condition.UnknownCondition(condition.InputReadyCondition, condition.InitReason, condition.InputReadyInitMessage),
			condition.UnknownCondition(telemetryv1.TelemetryRabbitMqTransportURLReadyCondition, condition.InitReason, telemetryv1.TelemetryRabbitMqTransportURLReadyInitMessage),
			condition.UnknownCondition(telemetryv1.CeilometerCentralReadyCondition, condition.InitReason, telemetryv1.CeilometerCentralReadyInitMessage),
			condition.UnknownCondition(telemetryv1.CeilometerComputeReadyCondition, condition.InitReason, telemetryv1.CeilometerComputeReadyInitMessage),
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

	//
	// create RabbitMQ transportURL CR and get the actual URL from the associated secret that is created
	//

	transportURL, op, err := r.transportURLCreateOrUpdate(instance)

	if err != nil {
		r.Log.Info("Error getting transportURL. Setting error condition on status and returning")
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.TelemetryRabbitMqTransportURLReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			telemetryv1.TelemetryRabbitMqTransportURLReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}

	if op != controllerutil.OperationResultNone {
		r.Log.Info(fmt.Sprintf("TransportURL %s successfully reconciled - operation: %s", transportURL.Name, string(op)))
	}

	instance.Status.TransportURLSecret = transportURL.Status.SecretName

	if instance.Status.TransportURLSecret == "" {
		r.Log.Info(fmt.Sprintf("Waiting for TransportURL %s secret to be created", transportURL.Name))
		instance.Status.Conditions.Set(condition.FalseCondition(
			telemetryv1.TelemetryRabbitMqTransportURLReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			telemetryv1.TelemetryRabbitMqTransportURLReadyRunningMessage))
		return ctrl.Result{RequeueAfter: time.Duration(10) * time.Second}, nil
	}

	instance.Status.Conditions.MarkTrue(telemetryv1.TelemetryRabbitMqTransportURLReadyCondition, telemetryv1.TelemetryRabbitMqTransportURLReadyRunningMessage)

	// end transportURL

	// create ConfigMap with all the telemetry playbooks ready for ansibleEE execution

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
	r.Log.Info("Deploying ceilometercompute")
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
	r.Log.Info("Deploying infracompute")
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

func (r *TelemetryReconciler) transportURLCreateOrUpdate(instance *telemetryv1.Telemetry) (*rabbitmqv1.TransportURL, controllerutil.OperationResult, error) {
	transportURL := &rabbitmqv1.TransportURL{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-telemetry-transport", instance.Name),
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

func (r *TelemetryReconciler) ceilometerCentralCreateOrUpdate(instance *telemetryv1.Telemetry) (*telemetryv1.CeilometerCentral, controllerutil.OperationResult, error) {
	ccentral := &telemetryv1.CeilometerCentral{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-ceilometer-central", instance.Name),
			Namespace: instance.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, ccentral, func() error {
		ccentral.Spec = instance.Spec.CeilometerCentral
		ccentral.Spec.TransportURLSecret = instance.Status.TransportURLSecret

		err := controllerutil.SetControllerReference(instance, ccentral, r.Scheme)
		if err != nil {
			return err
		}

		return nil
	})

	return ccentral, op, err
}

func (r *TelemetryReconciler) ceilometerComputeCreateOrUpdate(instance *telemetryv1.Telemetry) (*telemetryv1.CeilometerCompute, controllerutil.OperationResult, error) {
	r.Log.Info("Inside ceilometerComputeCreateOrUpdate")
	ccompute := &telemetryv1.CeilometerCompute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-ceilometer-compute", instance.Name),
			Namespace: instance.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.Client, ccompute, func() error {
		ccompute.Spec = instance.Spec.CeilometerCompute
		ccompute.Spec.TransportURLSecret = instance.Status.TransportURLSecret

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

func (r *TelemetryReconciler) ensurePlaybooks(ctx context.Context, h *helper.Helper, instance *telemetryv1.Telemetry) error {
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

func (r *TelemetryReconciler) ensureExtravars(ctx context.Context, h *helper.Helper, instance *telemetryv1.Telemetry) error {
	data := make(map[string]string)
	data["extravars"] = fmt.Sprintf("telemetry_node_exporter_image: %s", instance.Spec.InfraCompute.NodeExporterImage)

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
func (r *TelemetryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1.Telemetry{}).
		Owns(&telemetryv1.CeilometerCentral{}).
		Owns(&telemetryv1.CeilometerCompute{}).
		Owns(&telemetryv1.InfraCompute{}).
		Owns(&rabbitmqv1.TransportURL{}).
		Complete(r)
}
