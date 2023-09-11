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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	common "github.com/openstack-k8s-operators/lib-common/modules/common"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	deployment "github.com/openstack-k8s-operators/lib-common/modules/common/deployment"
	endpoint "github.com/openstack-k8s-operators/lib-common/modules/common/endpoint"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	job "github.com/openstack-k8s-operators/lib-common/modules/common/job"
	secret "github.com/openstack-k8s-operators/lib-common/modules/common/secret"
	service "github.com/openstack-k8s-operators/lib-common/modules/common/service"
	util "github.com/openstack-k8s-operators/lib-common/modules/common/util"
	database "github.com/openstack-k8s-operators/lib-common/modules/database"

	routev1 "github.com/openshift/api/route/v1"
	keystonev1 "github.com/openstack-k8s-operators/keystone-operator/api/v1beta1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	autoscaling "github.com/openstack-k8s-operators/telemetry-operator/pkg/autoscaling"
)

func (r *AutoscalingReconciler) reconcileDisabledAodh(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
) (ctrl.Result, error) {
	r.Log.Info("Reconciling Service Aodh disabled")
	serviceLabels := map[string]string{
		common.AppSelector: autoscaling.ServiceName,
	}

	// run the Delete to remove all finalizers
	ctrlResult, err := r.reconcileDeleteAodh(ctx, instance, helper)
	if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	// We are disabling autoscaling, not deleting it. We need to
	// manually delete all autoscaling resources,
	// that were created before

	// Delete deployment
	depl, err := autoscaling.AodhDeployment(instance, "", serviceLabels)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = r.Client.Delete(ctx, depl)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	// Delete db
	db, err := database.GetDatabaseByName(ctx, helper, autoscaling.ServiceName)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}
	if !k8s_errors.IsNotFound(err) {
		mariadbdatabase := db.GetDatabase()
		err = r.Client.Delete(ctx, mariadbdatabase)
		if err != nil && !k8s_errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	// Delete keystone service
	keystoneService, err := keystonev1.GetKeystoneServiceWithName(ctx, helper, autoscaling.ServiceName, instance.Namespace)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}
	if !k8s_errors.IsNotFound(err) {
		err = r.Client.Delete(ctx, keystoneService)
		if err != nil && !k8s_errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	// Delete keystone endpoint
	keystoneEndpoint, err := keystonev1.GetKeystoneEndpointWithName(ctx, helper, autoscaling.ServiceName, instance.Namespace)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}
	if !k8s_errors.IsNotFound(err) {
		err = r.Client.Delete(ctx, keystoneEndpoint)
		if err != nil && !k8s_errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	// Delete services
	svcs, err := service.GetServicesListWithLabel(ctx, helper, instance.Namespace, serviceLabels)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}
	if !k8s_errors.IsNotFound(err) {
		for _, svc := range svcs.Items {
			// NOTE: Using r.Client.Delete() to delete a service ends with:
			// "k8s.io/api/core/v1".Service does not implement
			// client.Object (DeepCopyObject method has pointer receiver)
			serviceClient := helper.GetKClient().CoreV1().Services(instance.Namespace)
			err = serviceClient.Delete(ctx, svc.ObjectMeta.Name, metav1.DeleteOptions{})
			if err != nil && !k8s_errors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
		}
	}

	// Delete public route
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: instance.Namespace,
			Name:      autoscaling.ServiceName + "-public",
		},
	}
	err = r.Client.Delete(ctx, route)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}
	instance.Status.Conditions = condition.Conditions{}
	r.Log.Info(fmt.Sprintf("Reconciled Service Aodh '%s' disable successfully", autoscaling.ServiceName))
	return ctrl.Result{}, nil
}

func (r *AutoscalingReconciler) reconcileDeleteAodh(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
) (ctrl.Result, error) {
	r.Log.Info("Reconciling Service Aodh delete")

	// remove db finalizer first
	db, err := database.GetDatabaseByName(ctx, helper, autoscaling.ServiceName)
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
	r.Log.Info(fmt.Sprintf("Reconciled Service Aodh '%s' delete successfully", autoscaling.ServiceName))

	return ctrl.Result{}, nil
}

func (r *AutoscalingReconciler) reconcileInitAodh(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
	serviceLabels map[string]string,
) (ctrl.Result, error) {
	r.Log.Info("Reconciling Service Aodh init")
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
	db := database.NewDatabaseWithNamespace(
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
		r.Log.Info(fmt.Sprintf("Service '%s' - Job %s hash added - %s", instance.Name, jobDef.Name, instance.Status.Hash[telemetryv1.DbSyncHash]))
	}
	instance.Status.Conditions.MarkTrue(condition.DBSyncReadyCondition, condition.DBSyncReadyMessage)

	// run Aodh db sync - end
	r.Log.Info("Reconciled Service Aodh init successfully")
	return ctrl.Result{}, nil
}

func (r *AutoscalingReconciler) reconcileNormalAodh(
	ctx context.Context,
	instance *telemetryv1.Autoscaling,
	helper *helper.Helper,
	inputHash string,
) (ctrl.Result, error) {
	r.Log.Info(fmt.Sprintf("Reconciling Service Aodh '%s'", autoscaling.ServiceName))
	serviceLabels := map[string]string{
		common.AppSelector: autoscaling.ServiceName,
	}

	deplDef, err := autoscaling.AodhDeployment(instance, inputHash, serviceLabels)
	if err != nil {
		return ctrl.Result{}, err
	}
	depl := deployment.NewDeployment(
		deplDef,
		time.Duration(5)*time.Second,
	)

	ctrlResult, err := depl.CreateOrPatch(ctx, helper)
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

	err = controllerutil.SetControllerReference(instance, deplDef, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	instance.Status.ReadyCount = depl.GetDeployment().Status.ReadyReplicas
	if instance.Status.ReadyCount > 0 {
		instance.Status.Conditions.MarkTrue(condition.DeploymentReadyCondition, condition.DeploymentReadyMessage)
	}
	instance.Status.Networks = instance.Spec.Aodh.NetworkAttachmentDefinitions

	ports := map[endpoint.Endpoint]endpoint.Data{}
	ports[endpoint.EndpointInternal] = endpoint.Data{
		Port: autoscaling.AodhAPIPort,
	}
	ports[endpoint.EndpointPublic] = endpoint.Data{
		Port: autoscaling.AodhAPIPort,
	}

	apiEndpoints, ctrlResult, err := endpoint.ExposeEndpoints(
		ctx,
		helper,
		autoscaling.ServiceName,
		serviceLabels,
		ports,
		time.Duration(5)*time.Second,
	)
	if err != nil {
		return ctrlResult, err
	}

	//
	// create keystone endpoints
	//

	ksEndpointSpec := keystonev1.KeystoneEndpointSpec{
		ServiceName: autoscaling.ServiceName,
		Endpoints:   apiEndpoints,
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

	r.Log.Info("Reconciled Service Aodh successfully")
	return ctrl.Result{}, nil
}
