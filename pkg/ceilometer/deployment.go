/*

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

package ceilometer

import (
	"fmt"

	ceilometerv1 "github.com/openstack-k8s-operators/ceilometer-operator/api/v1beta1"
	"github.com/openstack-k8s-operators/lib-common/modules/common/annotations"
	"github.com/openstack-k8s-operators/lib-common/modules/common/env"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// ServiceCommand -
	ServiceCommand = "/usr/local/bin/kolla_set_configs && /usr/local/bin/kolla_start"
)

// Deployment func
func Deployment(
	instance *ceilometerv1.Ceilometer,
	configHash string,
	labels map[string]string,
) (*appsv1.Deployment, error) {
	runAsUser := int64(0)

	// TO-DO Probes
	livenessProbe := &corev1.Probe{
		// TODO might need tuning
		TimeoutSeconds:      5,
		PeriodSeconds:       3,
		InitialDelaySeconds: 3,
	}
	readinessProbe := &corev1.Probe{
		// TODO might need tuning
		TimeoutSeconds:      5,
		PeriodSeconds:       5,
		InitialDelaySeconds: 5,
	}

	args := []string{"-c"}
	args = append(args, ServiceCommand)

	livenessProbe.HTTPGet = &corev1.HTTPGetAction{
		Path: "/v3",
		Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(CeilometerPrometheusPort)},
	}
	readinessProbe.HTTPGet = &corev1.HTTPGetAction{
		Path: "/v3",
		Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(CeilometerPrometheusPort)},
	}

	envVarsCentral := map[string]env.Setter{}
	envVarsCentral["KOLLA_CONFIG_FILE"] = env.SetValue(KollaConfigCentral)
	envVarsCentral["KOLLA_CONFIG_STRATEGY"] = env.SetValue("COPY_ALWAYS")
	envVarsCentral["CONFIG_HASH"] = env.SetValue(configHash)

	envVarsNotification := map[string]env.Setter{}
	envVarsNotification["KOLLA_CONFIG_FILE"] = env.SetValue(KollaConfigNotification)
	envVarsNotification["KOLLA_CONFIG_STRATEGY"] = env.SetValue("COPY_ALWAYS")
	envVarsNotification["CONFIG_HASH"] = env.SetValue(configHash)

	var replicas int32 = 1

	centralAgentContainer := corev1.Container{
		ImagePullPolicy: "Always",
		Command: []string{
			"/bin/bash",
		},
		Args:  args,
		Image: instance.Spec.CentralImage,
		Name:  "ceilometer-central-agent",
		Env:   env.MergeEnvs([]corev1.EnvVar{}, envVarsCentral),
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &runAsUser,
		},
		VolumeMounts: getVolumeMounts(),
	}
	notificationAgentContainer := corev1.Container{
		ImagePullPolicy: "Always",
		Command: []string{
			"/bin/bash",
		},
		Args:  args,
		Image: instance.Spec.NotificationImage,
		Name:  "ceilometer-notification-agent",
		Env:   env.MergeEnvs([]corev1.EnvVar{}, envVarsNotification),
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &runAsUser,
		},
		VolumeMounts: getVolumeMounts(),
	}
	sgCoreContainer := corev1.Container{
		ImagePullPolicy: "Always",
		Image:           instance.Spec.SgCoreImage,
		Name:            "sg-core",
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &runAsUser,
		},
		VolumeMounts:   getVolumeMountsSgCore(),
		ReadinessProbe: readinessProbe,
		LivenessProbe:  livenessProbe,
	}

	pod := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				centralAgentContainer,
				notificationAgentContainer,
				sgCoreContainer,
			},
		},
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: pod,
		},
	}

	deployment.Spec.Template.Spec.Volumes = getVolumes(instance.Name)

	// networks to attach to
	nwAnnotation, err := annotations.GetNADAnnotation(instance.Namespace, instance.Spec.NetworkAttachmentDefinitions)
	if err != nil {
		return nil, fmt.Errorf("failed create network annotation from %s: %w",
			instance.Spec.NetworkAttachmentDefinitions, err)
	}
	deployment.Spec.Template.Annotations = util.MergeStringMaps(deployment.Spec.Template.Annotations, nwAnnotation)

	initContainerDetails := APIDetails{
		ContainerImage:           instance.Spec.InitImage,
		RabbitMQSecret:           instance.Spec.RabbitMQSecret,
		RabbitMQHostSelector:     instance.Spec.PasswordSelectors.Host,
		RabbitMQUsernameSelector: instance.Spec.PasswordSelectors.Username,
		RabbitMQPasswordSelector: instance.Spec.PasswordSelectors.Password,
		OSPSecret:                instance.Spec.Secret,
		ServiceSelector:          instance.Spec.PasswordSelectors.Service,
		VolumeMounts:             getInitVolumeMounts(),
	}
	deployment.Spec.Template.Spec.InitContainers = initContainer(initContainerDetails)

	return deployment, nil
}
