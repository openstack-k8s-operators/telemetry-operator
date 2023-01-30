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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"

	logr "github.com/go-logr/logr"
	common "github.com/openstack-k8s-operators/lib-common/modules/common"
	configmap "github.com/openstack-k8s-operators/lib-common/modules/common/configmap"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	labels "github.com/openstack-k8s-operators/lib-common/modules/common/labels"
	util "github.com/openstack-k8s-operators/lib-common/modules/common/util"
	env "github.com/openstack-k8s-operators/lib-common/modules/common/env"
	oko_secret "github.com/openstack-k8s-operators/lib-common/modules/common/secret"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"

	ceilometerv1beta1 "github.com/openstack-k8s-operators/ceilometer-operator/api/v1beta1"
	"github.com/openstack-k8s-operators/ceilometer-operator/pkg/ceilometer"
	keystonev1 "github.com/openstack-k8s-operators/keystone-operator/api/v1beta1"
)

// CeilometerReconciler reconciles a Ceilometer object
type CeilometerReconciler struct {
	client.Client
	Kclient kubernetes.Interface
	Log     logr.Logger
	Scheme  *runtime.Scheme
}

//+kubebuilder:rbac:groups=ceilometer.openstack.org,resources=ceilometers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ceilometer.openstack.org,resources=ceilometers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ceilometer.openstack.org,resources=ceilometers/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete;
//+kubebuilder:rbac:groups=keystone.openstack.org,resources=keystoneservices,verbs=get;list;watch;create;update;patch;delete;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ceilometer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *CeilometerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	fmt.Printf("Log: %v", r.Log)
	_ = r.Log.WithValues("ceilometer", req.NamespacedName)

	// Fetch the Ceilometer instance
	instance := &ceilometerv1beta1.Ceilometer{}
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

func (r *CeilometerReconciler) reconcileDelete(ctx context.Context, instance *ceilometerv1beta1.Ceilometer, helper *helper.Helper) (ctrl.Result, error) {
	r.Log.Info("Reconciling Service delete")

	// do delete stuff

	return ctrl.Result{}, nil
}

func (r *CeilometerReconciler) reconcileInit(
	ctx context.Context,
	instance *ceilometerv1beta1.Ceilometer,
	helper *helper.Helper,
	serviceLabels map[string]string,
) (ctrl.Result, error) {
	r.Log.Info("Reconciling Service init")

	// init stuff
	// keystone add the user/password to be able to get a token

	r.Log.Info("Reconciled Service init successfully")
	return ctrl.Result{}, nil
}

// podForCeilometer returns a ceilometer Pod object
func (r *CeilometerReconciler) reconcileNormal(ctx context.Context, instance *ceilometerv1beta1.Ceilometer, helper *helper.Helper) (ctrl.Result, error) {
	r.Log.Info(fmt.Sprintf("Reconciling Service '%s'", instance.Name))
	// ConfigMap
	configMapVars := make(map[string]env.Setter)

	rabbitSecret, hash, err := oko_secret.GetSecret(ctx, helper, instance.Spec.RabbitMqSecret, instance.Namespace)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			instance.Status.Conditions.Set(condition.FalseCondition(
				condition.InputReadyCondition,
				condition.RequestedReason,
				condition.SeverityInfo,
				condition.InputReadyWaitingMessage))
			return ctrl.Result{RequeueAfter: time.Second * 10}, fmt.Errorf("RabbitMQ secret %s not found", instance.Spec.RabbitMqSecret)
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
	instance.Status.Conditions.MarkTrue(condition.InputReadyCondition, condition.InputReadyMessage)

	//
	// create Configmap required for ceilometer input
	// - %-scripts configmap holding scripts to e.g. bootstrap the service
	// - %-config configmap holding minimal keystone config required to get the service up, user can add additional files to be added to the service
	// - parameters which has passwords gets added from the OpenStack secret via the init container
	//
	err = r.generateServiceConfigMaps(ctx, helper, instance, &configMapVars)
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
	// TODO check when/if Init, Update, or Upgrade should/could be skipped
	//

	serviceLabels := map[string]string{
		common.AppSelector: ceilometer.ServiceName,
	}

	// Handle service init
	ctrlResult, err := r.reconcileInit(ctx, instance, helper, serviceLabels)
	if err != nil {
		return ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		return ctrlResult, nil
	}

	ls := labelsForCeilometer(instance.Name)
	var envVars []corev1.EnvVar
	var kollaEnvVar corev1.EnvVar
	kollaEnvVar.Name = "KOLLA_CONFIG_STRATEGY"
	kollaEnvVar.Value = "COPY_ALWAYS"
	envVars = append(envVars, kollaEnvVar)

	centralAgentContainer := corev1.Container{
		ImagePullPolicy: "Always",
		Image:           "quay.io/tripleomastercentos9/openstack-ceilometer-central:current-tripleo",
		Name:            "ceilometer-central-agent",
		Env:             envVars,
		VolumeMounts: []corev1.VolumeMount{{
			Name:      "ceilometer-conf",
			MountPath: "/var/lib/kolla/config_files/src/etc/ceilometer/ceilometer.conf",
			SubPath:   "ceilometer.conf",
		}, {
			Name:      "config-central-json",
			MountPath: "/var/lib/kolla/config_files/config.json",
			SubPath:   "config.json",
		}},
	}
	notificationAgentContainer := corev1.Container{
		ImagePullPolicy: "Always",
		Image:           "quay.io/tripleomastercentos9/openstack-ceilometer-notification:current-tripleo",
		Name:            "ceilometer-notification-agent",
		Env:             envVars,
		VolumeMounts: []corev1.VolumeMount{{
			Name:      "ceilometer-conf",
			MountPath: "/var/lib/kolla/config_files/src/etc/ceilometer/ceilometer.conf",
			SubPath:   "ceilometer.conf",
		}, {
			Name:      "pipeline-yaml",
			MountPath: "/var/lib/kolla/config_files/src/etc/ceilometer/pipeline.yaml",
			SubPath:   "pipeline.yaml",
		}, {
			Name:      "config-notification-json",
			MountPath: "/var/lib/kolla/config_files/config.json",
			SubPath:   "config.json",
		}},
	}
	sgCoreContainer := corev1.Container{
		ImagePullPolicy: "Always",
		Image:           "quay.io/jlarriba/sg-core:latest",
		Name:            "sg-core",
		VolumeMounts: []corev1.VolumeMount{{
			Name:      "sg-core-conf-yaml",
			MountPath: "/etc/sg-core.conf.yaml",
			SubPath:   "sg-core.conf.yaml",
		}},
	}

	pod := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    ls,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				centralAgentContainer,
				notificationAgentContainer,
				sgCoreContainer,
			},
			Volumes: []corev1.Volume{{
				Name: "ceilometer-conf",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						Items: []corev1.KeyToPath{{
							Key:  "ceilometer.conf",
							Path: "ceilometer.conf",
						}},
						LocalObjectReference: corev1.LocalObjectReference{
							Name: fmt.Sprintf("%s-config-data", instance.Name),
						},
					},
				},
			}, {
				Name: "config-central-json",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						Items: []corev1.KeyToPath{{
							Key:  "config-central.json",
							Path: "config.json",
						}},
						LocalObjectReference: corev1.LocalObjectReference{
							Name: fmt.Sprintf("%s-config-data", instance.Name),
						},
					},
				},
			}, {
				Name: "config-notification-json",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						Items: []corev1.KeyToPath{{
							Key:  "config-notification.json",
							Path: "config.json",
						}},
						LocalObjectReference: corev1.LocalObjectReference{
							Name: fmt.Sprintf("%s-config-data", instance.Name),
						},
					},
				},
			}, {
				Name: "pipeline-yaml",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						Items: []corev1.KeyToPath{{
							Key:  "pipeline.yaml",
							Path: "pipeline.yaml",
						}},
						LocalObjectReference: corev1.LocalObjectReference{
							Name: fmt.Sprintf("%s-config-data", instance.Name),
						},
					},
				},
			}, {
				Name: "sg-core-conf-yaml",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						Items: []corev1.KeyToPath{{
							Key:  "sg-core.conf.yaml",
							Path: "sg-core.conf.yaml",
						}},
						LocalObjectReference: corev1.LocalObjectReference{
							Name: fmt.Sprintf("%s-config-data", instance.Name),
						},
					},
				},
			}},
		},
	}

	var replicas int32 = 1

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    ls,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: pod,
		},
	}

	// Set Ceilometer instance as the owner and controller
	err = ctrl.SetControllerReference(instance, deployment, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, err
}

func (r *CeilometerReconciler) generateServiceConfigMaps(
	ctx context.Context, 
	h *helper.Helper, 
	instance *ceilometerv1beta1.Ceilometer,
	envVars *map[string]env.Setter,
	) error {


	cmLabels := labels.GetLabels(instance, labels.GetGroupLabel(ceilometer.ServiceName), map[string]string{})
	customData := map[string]string{common.CustomServiceConfigFileName: instance.Spec.CustomServiceConfig}
	for key, data := range instance.Spec.DefaultConfigOverwrite {
		customData[key] = data
	}

	templateParameters := make(map[string]interface{})

	cms := []util.Template{
		// ScriptsConfigMap
		{
			Name:               fmt.Sprintf("%s-scripts", instance.Name),
			Namespace:          instance.Namespace,
			Type:               util.TemplateTypeScripts,
			InstanceType:       instance.Kind,
			AdditionalTemplate: map[string]string{"common.sh": "/common/common.sh"},
			Labels:             cmLabels,
		},
		// ConfigMap
		{
			Name:          fmt.Sprintf("%s-config-data", instance.Name),
			Namespace:     instance.Namespace,
			Type:          util.TemplateTypeConfig,
			InstanceType:  instance.Kind,
			CustomData:    customData,
			ConfigOptions: templateParameters,
			Labels:        cmLabels,
		},
	}
	err := configmap.EnsureConfigMaps(ctx, h, instance, cms, envVars)
	if err != nil {
		return err
	}

	return nil
}

// labelsForCeilometer returns the labels for selecting the resources
// belonging to the given ceilometer CR name.
func labelsForCeilometer(name string) map[string]string {
	return map[string]string{"app": "ceilometer", "ceilometer_cr": name}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CeilometerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ceilometerv1beta1.Ceilometer{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&keystonev1.KeystoneService{}).
		Complete(r)
}
