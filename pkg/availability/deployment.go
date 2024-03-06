/*
Copyright 2024.

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

package availability

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
)

// KSMDeployment requests Deployment of kube-state-metrics
func KSMDeployment(
	instance *telemetryv1.Ceilometer,
	labels map[string]string,
) (*appsv1.Deployment, error) {

	livenessProbe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/healthz",
				Port: intstr.FromInt(KSMHealthPort),
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      5,
	}

	readinessProbe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/",
				Port: intstr.FromInt(KSMMetricsPort),
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      5,
	}

	secCtx := &corev1.SecurityContext{
		AllowPrivilegeEscalation: pointer.BoolPtr(false),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"ALL"},
		},
		ReadOnlyRootFilesystem: pointer.BoolPtr(true),
		RunAsNonRoot:           pointer.BoolPtr(true),
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
	}

	container := corev1.Container{
		ImagePullPolicy: corev1.PullAlways,
		Image:           instance.Spec.KSMImage,
		Name:            KSMServiceName,
		Args: []string{
			"--resources=pods",
			fmt.Sprintf("--namespaces=%s", instance.Namespace),
		},
		SecurityContext: secCtx,
		LivenessProbe:   livenessProbe,
		ReadinessProbe:  readinessProbe,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: KSMMetricsPort,
				Name:          "http-metrics",
			},
			{
				ContainerPort: KSMHealthPort,
				Name:          "telemetry",
			},
		},
	}

	labels["app.kubernetes.io/component"] = "exporter"
	labels["app.kubernetes.io/name"] = KSMServiceName
	labels["app.kubernetes.io/version"] = instance.Spec.KSMImage[strings.LastIndex(instance.Spec.KSMImage, ":")+1:]

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KSMServiceName,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &KSMReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name": labels["app.kubernetes.io/name"],
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					AutomountServiceAccountToken: pointer.BoolPtr(true),
					ServiceAccountName:           KSMServiceName,
					Containers:                   []corev1.Container{container},
				},
			},
		},
	}

	return deployment, nil
}
