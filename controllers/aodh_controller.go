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

	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	common "github.com/openstack-k8s-operators/lib-common/modules/common"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	endpoint "github.com/openstack-k8s-operators/lib-common/modules/common/endpoint"
	env "github.com/openstack-k8s-operators/lib-common/modules/common/env"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	job "github.com/openstack-k8s-operators/lib-common/modules/common/job"
	secret "github.com/openstack-k8s-operators/lib-common/modules/common/secret"
	service "github.com/openstack-k8s-operators/lib-common/modules/common/service"
	statefulset "github.com/openstack-k8s-operators/lib-common/modules/common/statefulset"
	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	util "github.com/openstack-k8s-operators/lib-common/modules/common/util"
	mariadbv1 "github.com/openstack-k8s-operators/mariadb-operator/api/v1beta1"

	keystonev1 "github.com/openstack-k8s-operators/keystone-operator/api/v1beta1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	autoscaling "github.com/openstack-k8s-operators/telemetry-operator/pkg/autoscaling"
)

func (r *AutoscalingReconciler) reconcileDeleteAodh(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info("Reconciling Service Aodh delete")

	// remove db finalizer first
	db, err := mariadbv1.GetDatabaseByName(ctx, helper, autoscaling.DatabaseName)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if !k8s_errors.IsNotFound(err) {
		if err := db.DeleteFinalizer(ctx, helper); err != nil {
			return ctrl.Result{}, err
		}
	}

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

	// Remove the finalizer from our KeystoneEndpoint CR
	keystoneEndpoint, err := keystonev1.GetKeystoneEndpointWithName(ctx, helper, autoscaling.ServiceName, instance.Namespace)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}
	if err == nil {
		if controllerutil.RemoveFinalizer(keystoneEndpoint, helper.GetFinalizer()) {
			err = r.Update(ctx, keystoneEndpoint)
			if err != nil && !k8s_errors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
			util.LogForObject(helper, "Removed finalizer from our KeystoneEndpoint", instance)
		}
	}
	Log.Info(fmt.Sprintf("Reconciled Service Aodh '%s' delete successfully", autoscaling.ServiceName))

	return ctrl.Result{}, nil
}

func (r *AutoscalingReconciler) reconcileInitAodh(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
	serviceLabels map[string]string,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info("Reconciling Service Aodh init")
	_, _, err := secret.GetSecret(ctx, helper, instance.Spec.Aodh.Secret, instance.Namespace)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: time.Duration(10) * time.Second}, fmt.Errorf("OpenStack secret %s not found", instance.Spec.Aodh.Secret)
		}
		return ctrl.Result{}, err
	}

	ksSvcSpec := keystonev1.KeystoneServiceSpec{
		ServiceType:        autoscaling.ServiceType,
		ServiceName:        autoscaling.ServiceName,
		ServiceDescription: "Aodh for autoscaling Service",
		Enabled:            true,
		ServiceUser:        instance.Spec.Aodh.ServiceUser,
		Secret:             instance.Spec.Aodh.Secret,
		PasswordSelector:   instance.Spec.Aodh.PasswordSelectors.AodhService,
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
	//
	// create service DB instance
	//
	db := mariadbv1.NewDatabaseWithNamespace(
		// instance.Name
		// TODO: We might want to change the db name.
		// The mariadb-operator is currently implemented
		// in a way, that the db name needs to be the
		// same as the user
		instance.Spec.Aodh.DatabaseUser,
		instance.Spec.Aodh.DatabaseUser,
		instance.Spec.Aodh.Secret,
		map[string]string{
			"dbName": instance.Spec.Aodh.DatabaseInstance,
		},
		autoscaling.ServiceName,
		instance.Namespace,
	)
	// create or patch the DB
	ctrlResult, err = db.CreateOrPatchDBByName(
		ctx,
		helper,
		instance.Spec.Aodh.DatabaseInstance,
	)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DBReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.DBReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}
	if (ctrlResult != ctrl.Result{}) {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DBReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.DBReadyRunningMessage))
		return ctrlResult, nil
	}
	// wait for the DB to be setup
	ctrlResult, err = db.WaitForDBCreated(ctx, helper)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DBReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.DBReadyErrorMessage,
			err.Error()))
		return ctrlResult, err
	}
	if (ctrlResult != ctrl.Result{}) {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DBReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.DBReadyRunningMessage))
		return ctrlResult, nil
	}
	// update Status.DatabaseHostname, used to config the service
	instance.Status.DatabaseHostname = db.GetDatabaseHostname()
	instance.Status.Conditions.MarkTrue(condition.DBReadyCondition, condition.DBReadyMessage)
	// create service DB - end

	//
	// run Aodh db sync
	dbSyncHash := instance.Status.Hash[telemetryv1.DbSyncHash]
	jobDef := autoscaling.DbSyncJob(instance, serviceLabels)

	dbSyncjob := job.NewJob(
		jobDef,
		telemetryv1.DbSyncHash,
		instance.Spec.Aodh.PreserveJobs,
		time.Duration(5)*time.Second,
		dbSyncHash,
	)
	ctrlResult, err = dbSyncjob.DoJob(
		ctx,
		helper,
	)
	if (ctrlResult != ctrl.Result{}) {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DBSyncReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.DBSyncReadyRunningMessage))
		return ctrlResult, nil
	}
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DBSyncReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.DBSyncReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}
	if dbSyncjob.HasChanged() {
		instance.Status.Hash[telemetryv1.DbSyncHash] = dbSyncjob.GetHash()
		Log.Info(fmt.Sprintf("Service '%s' - Job %s hash added - %s", instance.Name, jobDef.Name, instance.Status.Hash[telemetryv1.DbSyncHash]))
	}
	instance.Status.Conditions.MarkTrue(condition.DBSyncReadyCondition, condition.DBSyncReadyMessage)

	// run Aodh db sync - end
	Log.Info("Reconciled Service Aodh init successfully")
	return ctrl.Result{}, nil
}

func (r *AutoscalingReconciler) reconcileNormalAodh(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
	inputHash string,
) (ctrl.Result, error) {
	Log := r.GetLogger(ctx)
	Log.Info(fmt.Sprintf("Reconciling Service Aodh '%s'", autoscaling.ServiceName))
	serviceLabels := map[string]string{
		common.AppSelector: autoscaling.ServiceName,
	}

	// ConfigVars
	configVars := make(map[string]env.Setter)

	sfsetDef, err := autoscaling.AodhStatefulSet(instance, inputHash, serviceLabels)
	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			condition.DeploymentReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.DeploymentReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}
	sfset := statefulset.NewStatefulSet(
		sfsetDef,
		time.Duration(5)*time.Second,
	)

	ctrlResult, err := sfset.CreateOrPatch(ctx, helper)
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

	err = controllerutil.SetControllerReference(instance, sfsetDef, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	instance.Status.ReadyCount = sfset.GetStatefulSet().Status.ReadyReplicas
	if instance.Status.ReadyCount > 0 {
		instance.Status.Conditions.MarkTrue(condition.DeploymentReadyCondition, condition.DeploymentReadyMessage)
	}
	instance.Status.Networks = instance.Spec.Aodh.NetworkAttachmentDefinitions

	//
	// create service/s
	//

	aodhEndpoints := map[service.Endpoint]endpoint.Data{}
	aodhEndpoints[service.EndpointInternal] = endpoint.Data{
		Port: autoscaling.AodhAPIPort,
	}
	aodhEndpoints[service.EndpointPublic] = endpoint.Data{
		Port: autoscaling.AodhAPIPort,
	}

	apiEndpoints := make(map[string]string)

	for endpointType, data := range aodhEndpoints {
		endpointTypeStr := string(endpointType)
		endpointName := autoscaling.ServiceName + "-" + endpointTypeStr
		svcOverride := instance.Spec.Aodh.Override.Service
		if svcOverride == nil {
			svcOverride = &service.RoutedOverrideSpec{}
		}
		if svcOverride.EmbeddedLabelsAnnotations == nil {
			svcOverride.EmbeddedLabelsAnnotations = &service.EmbeddedLabelsAnnotations{}
		}

		exportLabels := util.MergeStringMaps(
			serviceLabels,
			map[string]string{
				service.AnnotationEndpointKey: endpointTypeStr,
			},
		)

		// Create the service
		svc, err := service.NewService(
			service.GenericService(&service.GenericServiceDetails{
				Name:      endpointName,
				Namespace: instance.Namespace,
				Labels:    exportLabels,
				Selector:  serviceLabels,
				Port: service.GenericServicePort{
					Name:     endpointName,
					Port:     data.Port,
					Protocol: corev1.ProtocolTCP,
				},
			}),
			5,
			&svcOverride.OverrideSpec,
		)
		if err != nil {
			instance.Status.Conditions.Set(condition.FalseCondition(
				condition.ExposeServiceReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.ExposeServiceReadyErrorMessage,
				err.Error()))

			return ctrl.Result{}, err
		}

		svc.AddAnnotation(map[string]string{
			service.AnnotationEndpointKey: endpointTypeStr,
		})

		// add Annotation to whether creating an ingress is required or not
		if endpointType == service.EndpointPublic && svc.GetServiceType() == corev1.ServiceTypeClusterIP {
			svc.AddAnnotation(map[string]string{
				service.AnnotationIngressCreateKey: "true",
			})
		} else {
			svc.AddAnnotation(map[string]string{
				service.AnnotationIngressCreateKey: "false",
			})
			if svc.GetServiceType() == corev1.ServiceTypeLoadBalancer {
				svc.AddAnnotation(map[string]string{
					service.AnnotationHostnameKey: svc.GetServiceHostname(), // add annotation to register service name in dnsmasq
				})
			}
		}

		ctrlResult, err := svc.CreateOrPatch(ctx, helper)
		if err != nil {
			instance.Status.Conditions.Set(condition.FalseCondition(
				condition.ExposeServiceReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.ExposeServiceReadyErrorMessage,
				err.Error()))

			return ctrlResult, err
		} else if (ctrlResult != ctrl.Result{}) {
			instance.Status.Conditions.Set(condition.FalseCondition(
				condition.ExposeServiceReadyCondition,
				condition.RequestedReason,
				condition.SeverityInfo,
				condition.ExposeServiceReadyRunningMessage))
			return ctrlResult, nil
		}
		// create service - end

		// if TLS is enabled
		if instance.Spec.Aodh.TLS.API.Enabled(endpointType) {
			// set endpoint protocol to https
			data.Protocol = ptr.To(service.ProtocolHTTPS)
		}

		apiEndpoints[string(endpointType)], err = svc.GetAPIEndpoint(
			svcOverride.EndpointURL, data.Protocol, data.Path)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	instance.Status.Conditions.MarkTrue(condition.ExposeServiceReadyCondition, condition.ExposeServiceReadyMessage)

	if instance.Status.APIEndpoints == nil {
		instance.Status.APIEndpoints = map[string]string{}
	}

	instance.Status.APIEndpoints = apiEndpoints

	//
	// create keystone endpoints
	//

	ksEndpointSpec := keystonev1.KeystoneEndpointSpec{
		ServiceName: autoscaling.ServiceName,
		Endpoints:   instance.Status.APIEndpoints,
	}

	ksEndptObj := keystonev1.NewKeystoneEndpoint(autoscaling.ServiceName, instance.Namespace, ksEndpointSpec, serviceLabels, time.Duration(10)*time.Second)
	ctrlResult, err = ksEndptObj.CreateOrPatch(ctx, helper)
	if err != nil {
		return ctrlResult, err
	}

	c := ksEndptObj.GetConditions().Mirror(condition.KeystoneEndpointReadyCondition)
	if c != nil {
		instance.Status.Conditions.Set(c)
	}

	if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	//
	// TLS input validation
	//
	// Validate the CA cert secret if provided
	if instance.Spec.Aodh.TLS.CaBundleSecretName != "" {
		hash, ctrlResult, err := tls.ValidateCACertSecret(
			ctx,
			helper.GetClient(),
			types.NamespacedName{
				Name:      instance.Spec.Aodh.TLS.CaBundleSecretName,
				Namespace: instance.Namespace,
			},
		)
		if err != nil {
			instance.Status.Conditions.Set(condition.FalseCondition(
				condition.TLSInputReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.TLSInputErrorMessage,
				err.Error()))
			return ctrlResult, err
		} else if (ctrlResult != ctrl.Result{}) {
			return ctrlResult, nil
		}

		if hash != "" {
			configVars[tls.CABundleKey] = env.SetValue(hash)
		}

		// Validate API service certs secrets
		certsHash, ctrlResult, err := instance.Spec.Aodh.TLS.API.ValidateCertSecrets(ctx, helper, instance.Namespace)
		if err != nil {
			instance.Status.Conditions.Set(condition.FalseCondition(
				condition.TLSInputReadyCondition,
				condition.ErrorReason,
				condition.SeverityWarning,
				condition.TLSInputErrorMessage,
				err.Error()))
			return ctrlResult, err
		} else if (ctrlResult != ctrl.Result{}) {
			return ctrlResult, nil
		}

		configVars[tls.TLSHashName] = env.SetValue(certsHash)
	}

	// all cert input checks out so report InputReady
	instance.Status.Conditions.MarkTrue(condition.TLSInputReadyCondition, condition.InputReadyMessage)

	Log.Info("Reconciled Service Aodh successfully")
	return ctrl.Result{}, nil
}
