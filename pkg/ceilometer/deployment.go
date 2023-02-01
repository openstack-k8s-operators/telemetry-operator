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

	ceilometerv1beta1 "github.com/openstack-k8s-operators/ceilometer-operator/api/v1beta1"
	"github.com/openstack-k8s-operators/lib-common/modules/common/annotations"
	"github.com/openstack-k8s-operators/lib-common/modules/common/env"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ServiceCommand -
	ServiceCommand = "/usr/local/bin/kolla_set_configs && /usr/local/bin/kolla_start"
)

// Deployment func
func Deployment(
	instance *ceilometerv1beta1.Ceilometer,
	configHash string,
	labels map[string]string,
) (*appsv1.Deployment, error) {
	runAsUser := int64(0)

	// TO-DO Probes
	/*livenessProbe := &corev1.Probe{
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
	if instance.Spec.Debug.Service {
		args = append(args, common.DebugCommand)
		livenessProbe.Exec = &corev1.ExecAction{
			Command: []string{
				"/bin/true",
			},
		}

		readinessProbe.Exec = &corev1.ExecAction{
			Command: []string{
				"/bin/true",
			},
		}
	} else {
		args = append(args, ServiceCommand)

		//
		// https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/
		//
		livenessProbe.HTTPGet = &corev1.HTTPGetAction{
			Path: "/v3",
			Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(KeystonePublicPort)},
		}
		readinessProbe.HTTPGet = &corev1.HTTPGetAction{
			Path: "/v3",
			Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(KeystonePublicPort)},
		}
	}*/

	envVars := map[string]env.Setter{}
	envVars["KOLLA_CONFIG_FILE"] = env.SetValue(KollaConfig)
	envVars["KOLLA_CONFIG_STRATEGY"] = env.SetValue("COPY_ALWAYS")
	envVars["CONFIG_HASH"] = env.SetValue(configHash)

	var replicas int32 = 1

	centralAgentContainer := corev1.Container{
		ImagePullPolicy: "Always",
		Image:           "quay.io/tripleomastercentos9/openstack-ceilometer-central:current-tripleo",
		Name:            "ceilometer-central-agent",
		Env:             env.MergeEnvs([]corev1.EnvVar{}, envVars),
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &runAsUser,
		},
		VolumeMounts: []corev1.VolumeMount{{
			Name:      "ceilometer-conf",
			MountPath: "/var/lib/config-data/merged/ceilometer.conf",
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
		Env:             env.MergeEnvs([]corev1.EnvVar{}, envVars),
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &runAsUser,
		},
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
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &runAsUser,
		},
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

	initContainerDetails := CeilometerDetails{
		ContainerImage:           "quay.io/tripleomastercentos9/openstack-ceilometer-central:current-tripleo",
		RabbitMQSecret:           instance.Spec.RabbitMQSecret,
		RabbitMQHostSelector:     instance.Spec.RabbitMQSelectors.Host,
		RabbitMQUsernameSelector: instance.Spec.RabbitMQSelectors.Username,
		RabbitMQPasswordSelector: instance.Spec.RabbitMQSelectors.Password,
		VolumeMounts:             getInitVolumeMounts(),
	}
	deployment.Spec.Template.Spec.InitContainers = initContainer(initContainerDetails)

	return deployment, nil
}
