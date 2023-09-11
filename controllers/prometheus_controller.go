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
	"time"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	common "github.com/openstack-k8s-operators/lib-common/modules/common"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	service "github.com/openstack-k8s-operators/lib-common/modules/common/service"

	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	autoscaling "github.com/openstack-k8s-operators/telemetry-operator/pkg/autoscaling"
)

func (r *AutoscalingReconciler) reconcileDisabledPrometheus(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
) (ctrl.Result, error) {
	r.Log.Info("Reconciling Service Prometheus disabled")
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

func (r *AutoscalingReconciler) reconcileDeletePrometheus(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
) (ctrl.Result, error) {
	r.Log.Info("Reconciling Service Prometheus delete")
	// TODO: finalizer prometheus
	r.Log.Info(fmt.Sprintf("Reconciled Service '%s' delete successfully", autoscaling.ServiceName))

	return ctrl.Result{}, nil
}

func (r *AutoscalingReconciler) reconcileInitPrometheus(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
	serviceLabels map[string]string,
) (ctrl.Result, error) {
	// TODO: init?
	return ctrl.Result{}, nil
}

func (r *AutoscalingReconciler) reconcileNormalPrometheus(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
) (ctrl.Result, error) {
	serviceLabels := map[string]string{
		common.AppSelector: autoscaling.ServiceName,
	}
	prom := autoscaling.Prometheus(instance, serviceLabels)
	r.Log.Info(fmt.Sprintf("Reconciling Service Aodh '%s'", prom.Name))

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
			promHost = fmt.Sprintf("%s.%s.svc", serviceName, instance.Namespace)
			promPort = service.GetServicesPortDetails(promSvc, "web").Port
		} else {
			return ctrl.Result{RequeueAfter: time.Duration(10) * time.Second}, fmt.Errorf("Prometheus %s isn't ready", prom.Name)
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

	instance.Status.PrometheusHost = promHost
	instance.Status.PrometheusPort = promPort

	r.Log.Info(fmt.Sprintf("Reconciled Service Aodh '%s' successfully", prom.Name))
	return ctrl.Result{}, nil
}
