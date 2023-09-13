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

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	logr "github.com/go-logr/logr"
	common "github.com/openstack-k8s-operators/lib-common/modules/common"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	common_rbac "github.com/openstack-k8s-operators/lib-common/modules/common/rbac"
	service "github.com/openstack-k8s-operators/lib-common/modules/common/service"
	util "github.com/openstack-k8s-operators/lib-common/modules/common/util"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	keystonev1 "github.com/openstack-k8s-operators/keystone-operator/api/v1beta1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	autoscaling "github.com/openstack-k8s-operators/telemetry-operator/pkg/autoscaling"
	obov1 "github.com/rhobs/observability-operator/pkg/apis/monitoring/v1alpha1"
)

// AutoscalingReconciler reconciles a Autoscaling object
type AutoscalingReconciler struct {
	client.Client
	Kclient kubernetes.Interface
	Log     logr.Logger
	Scheme  *runtime.Scheme
}

// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=autoscalings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=autoscalings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=telemetry.openstack.org,resources=autoscalings/finalizers,verbs=update

// Reconcile reconciles an Autoscaling
func (r *AutoscalingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues(autoscaling.ServiceName, req.NamespacedName)

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
			condition.UnknownCondition("PrometheusReady", condition.InitReason, "PrometheusNotStarted"),
			// right now we have no dedicated KeystoneServiceReadyInitMessage
			//condition.UnknownCondition(condition.KeystoneServiceReadyCondition, condition.InitReason, ""),
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

	// Handle service disabled
	if !instance.Spec.Enabled {
		return r.reconcileDisabled(ctx, instance, helper)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, instance, helper)
}

func (r *AutoscalingReconciler) reconcileDisabled(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
) (ctrl.Result, error) {
	r.Log.Info("Reconciling Service disabled")
	serviceLabels := map[string]string{
		common.AppSelector: autoscaling.ServiceName,
	}
	prom := autoscaling.Prometheus(instance, serviceLabels)
	err := r.Client.Delete(ctx, prom)
	if err != nil {
		if !k8s_errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
	}

	// Set the condition to true, since the service is disabled
	for _, c := range instance.Status.Conditions {
		instance.Status.Conditions.MarkTrue(c.Type, "Autoscaling disabled")
	}
	r.Log.Info(fmt.Sprintf("Reconciled Service '%s' disable successfully", autoscaling.ServiceName))
	return ctrl.Result{}, nil
}

func (r *AutoscalingReconciler) reconcileDelete(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
) (ctrl.Result, error) {
	r.Log.Info("Reconciling Service delete")

	// Remove the finalizer from our KeystoneService CR
	keystoneService, err := keystonev1.GetKeystoneServiceWithName(ctx, helper, autoscaling.ServiceName, instance.Namespace)
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
	r.Log.Info(fmt.Sprintf("Reconciled Service '%s' delete successfully", autoscaling.ServiceName))

	return ctrl.Result{}, nil
}

func (r *AutoscalingReconciler) reconcileInit(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
	serviceLabels map[string]string,
) (ctrl.Result, error) {
	r.Log.Info("Reconciling Service init")
	// TODO: This code is useles for prometheus, but it might
	//       be useful for aodh in the future. I'll leave it
	//       here for now.

	//
	// create Keystone service and users
	//
	/*
		        _, _, err := secret.GetSecret(ctx, helper, instance.Spec.Secret, instance.Namespace)
		        if err != nil {
		                if k8s_errors.IsNotFound(err) {
		                        return ctrl.Result{RequeueAfter: time.Duration(10) * time.Second}, fmt.Errorf("OpenStack secret %s not found", instance.Spec.Secret)
		                }
		                return ctrl.Result{}, err
		        }

			ksSvcSpec := keystonev1.KeystoneServiceSpec{
		                ServiceType:        autoscaling.ServiceType,
		                ServiceName:        autoscaling.ServiceName,
		                ServiceDescription: "Aodh for autoscaling Service",
		                Enabled:            true,
		                ServiceUser:        instance.Spec.ServiceUser,
		                Secret:             instance.Spec.Secret,
		                PasswordSelector:   instance.Spec.PasswordSelectors.Service,
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
	*/

	r.Log.Info("Reconciled Service init successfully")
	return ctrl.Result{}, nil

}

func (r *AutoscalingReconciler) reconcileNormal(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
) (ctrl.Result, error) {
	r.Log.Info(fmt.Sprintf("Reconciling Service '%s'", autoscaling.ServiceName))

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

	// Handle service init
	ctrlResult, err := r.reconcileInit(ctx, instance, helper, serviceLabels)
	if err != nil {
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	promResult, err := r.reconcilePrometheus(ctx, instance, helper, serviceLabels)
	if err != nil {
		return promResult, err
	}

	r.Log.Info("Reconciled Service successfully")
	return ctrl.Result{}, nil
}

func (r *AutoscalingReconciler) reconcilePrometheus(ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
	serviceLabels map[string]string,
) (ctrl.Result, error) {
	prom := autoscaling.Prometheus(instance, serviceLabels)

	var promHost string
	var promPort int32

	if instance.Spec.Prometheus.DeployPrometheus {
		op, err := controllerutil.CreateOrUpdate(ctx, r.Client, prom, func() error {
			err := controllerutil.SetControllerReference(instance, prom, r.Scheme)
			return err
		})
		if err != nil {
			return ctrl.Result{}, err
		}
		if op != controllerutil.OperationResultNone {
			r.Log.Info(fmt.Sprintf("Prometheus %s successfully reconciled - operation: %s", prom.Name, string(op)))
		}
		promReady := true
		for _, c := range prom.Status.Conditions {
			if c.Status != "True" {
				instance.Status.Conditions.MarkFalse("PrometheusReady",
					condition.Reason(c.Reason),
					condition.SeverityError,
					c.Message)
				promReady = false
				break
			}
		}
		if len(prom.Status.Conditions) == 0 {
			promReady = false
		}
		if promReady {
			instance.Status.Conditions.MarkTrue("PrometheusReady", "Prometheus is ready")
			serviceName := prom.Name + "-prometheus"
			promSvc, err := service.GetServiceWithName(ctx, helper, serviceName, instance.Namespace)
			if err != nil {
				return ctrl.Result{}, err
			}
			// TODO: Remove the nolint after adding aodh and using the variables
			promHost = fmt.Sprintf("%s.%s.svc", serviceName, instance.Namespace) //nolint:all
			promPort = service.GetServicesPortDetails(promSvc, "web").Port       //nolint:all
		}
	} else {
		err := r.Client.Delete(ctx, prom)
		if err != nil {
			if !k8s_errors.IsNotFound(err) {
				return ctrl.Result{}, nil
			}
		}
		promHost = instance.Spec.Prometheus.Host
		promPort = instance.Spec.Prometheus.Port
		if promHost == "" || promPort == 0 {
			instance.Status.Conditions.MarkFalse("PrometheusReady",
				condition.ErrorReason,
				condition.SeverityError,
				"deployPrometheus is false and either port or host isn't set")
		} else {
			instance.Status.Conditions.MarkTrue("PrometheusReady", "Prometheus is ready")
		}
	}

	// TODO: Pass the promHost and promPort variables to aodh

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AutoscalingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1.Autoscaling{}).
		Owns(&obov1.MonitoringStack{}).
		Complete(r)
}
