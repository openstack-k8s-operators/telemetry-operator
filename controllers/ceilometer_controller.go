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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	configmap "github.com/openstack-k8s-operators/lib-common/modules/common/configmap"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	util "github.com/openstack-k8s-operators/lib-common/modules/common/util"
	"k8s.io/client-go/kubernetes"

	ceilometerv1beta1 "github.com/openstack-k8s-operators/ceilometer-operator/api/v1beta1"
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

	instance, err := r.getCeilometerInstance(ctx, req)
	if err != nil || instance.Name == "" {
		return ctrl.Result{}, err
	}

	// Check if the pod already exists, if not create a new one
	foundPod := &corev1.Pod{}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, foundPod)
	if err != nil && errors.IsNotFound(err) {
		// Define a new pod
		pod, err := r.podForCeilometer(instance)
		if err != nil {
			return ctrl.Result{}, err
		}
		fmt.Printf("Creating a new Pod: Pod.Namespace %s Pod.Name %s\n", pod.Namespace, pod.Name)
		err = r.Create(ctx, pod)
		if err != nil {
			fmt.Println(err.Error())
			return ctrl.Result{}, err
		}
		fmt.Println("pod created successfully - return and requeue")
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		fmt.Println(err.Error())
		//log.Error(err, "Failed to get Job")
		return ctrl.Result{}, err
	}

	err = r.generateServiceConfigMaps(ctx, instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CeilometerReconciler) getCeilometerInstance(ctx context.Context, req ctrl.Request) (*ceilometerv1beta1.Ceilometer, error) {
	// Fetch the Ceilometer instance
	instance := &ceilometerv1beta1.Ceilometer{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			fmt.Println("Ceilometer resource not found. Ignoring since object must be deleted")
			//log.Info("Ceilometer resource not found. Ignoring since object must be deleted")
			return &ceilometerv1beta1.Ceilometer{}, nil
		}
		// Error reading the object - requeue the request.
		fmt.Println(err.Error())
		//log.Error(err, "Failed to get Ceilometer")
		return &ceilometerv1beta1.Ceilometer{}, err
	}

	return instance, nil
}

// podForCeilometer returns a ceilometer Pod object
func (r *CeilometerReconciler) podForCeilometer(instance *ceilometerv1beta1.Ceilometer) (*corev1.Pod, error) {
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
	}
	notificationAgentContainer := corev1.Container{
		ImagePullPolicy: "Always",
		Image:           "quay.io/tripleomastercentos9/openstack-ceilometer-notification:current-tripleo",
		Name:            "ceilometer-notification-agent",
		Env:             envVars,
	}
	sgCoreContainer := corev1.Container{
		ImagePullPolicy: "Always",
		Image:           "quay.io/jlarriba/sg-core:latest",
		Name:            "sg-core",
	}

	pod := &corev1.Pod{
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
		},
	}

	// Set Ceilometer instance as the owner and controller
	err := ctrl.SetControllerReference(instance, pod, r.Scheme)
	if err != nil {
		return nil, err
	}

	return pod, nil
}

func (r *CeilometerReconciler) generateServiceConfigMaps(ctx context.Context, instance *ceilometerv1beta1.Ceilometer) error {

	helper, err := helper.NewHelper(
		instance,
		r.Client,
		r.Kclient,
		r.Scheme,
		r.Log,
	)
	if err != nil {
		return err
	}
	cmLabels := labelsForCeilometer(instance.Name)

	cms := []util.Template{
		{
			Name:         "ceilometer-conf",
			Namespace:    instance.Namespace,
			Type:         util.TemplateTypeConfig,
			InstanceType: "ceilometer",
			Labels:       cmLabels,
		},
	}

	err = configmap.EnsureConfigMaps(ctx, helper, instance, cms, nil)
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
		Complete(r)
}
