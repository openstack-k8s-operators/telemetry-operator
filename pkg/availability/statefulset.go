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
	"path/filepath"

	"k8s.io/utils/ptr"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	topologyv1 "github.com/openstack-k8s-operators/infra-operator/apis/topology/v1beta1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
)

// KSMStatefulSet requests deployment of kube-state-metrics and creation of TLS config if it is necessary
func KSMStatefulSet(
	instance *telemetryv1.Ceilometer,
	tlsConfName string,
	labels map[string]string,
	topology *topologyv1.Topology,
) (*appsv1.StatefulSet, error) {

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
		AllowPrivilegeEscalation: ptr.To(false),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"ALL"},
		},
		ReadOnlyRootFilesystem: ptr.To(true),
		RunAsNonRoot:           ptr.To(true),
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
	}

	labels["app.kubernetes.io/component"] = "exporter"
	labels["app.kubernetes.io/name"] = KSMServiceName

	// kube-state-metrics relevant default arguments
	args := []string{
		"--resources=pods",
		fmt.Sprintf("--namespaces=%s", instance.Namespace),
	}

	volumes := []corev1.Volume{}
	mounts := []corev1.VolumeMount{}
	if instance.Spec.KSMTLS.Enabled() {
		svc, err := instance.Spec.KSMTLS.GenericService.ToService()
		if err != nil {
			return nil, err
		}

		tlsConfFullPath := filepath.Join(tlsConfPath, tlsConfKey)
		svc.CertMount = ptr.To(TLSCertPath)
		svc.KeyMount = ptr.To(TLSKeyPath)

		livenessProbe.HTTPGet.Scheme = corev1.URISchemeHTTPS
		readinessProbe.HTTPGet.Scheme = corev1.URISchemeHTTPS

		volumes = append(volumes, svc.CreateVolume(KSMServiceName), corev1.Volume{
			Name: fmt.Sprintf("%s-tls-config", KSMServiceName),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  tlsConfName,
					DefaultMode: ptr.To[int32](0400),
				},
			},
		})
		mounts = svc.CreateVolumeMounts(KSMServiceName)
		mounts = append(mounts, corev1.VolumeMount{
			Name:      fmt.Sprintf("%s-tls-config", KSMServiceName),
			MountPath: tlsConfFullPath,
			SubPath:   tlsConfKey,
			ReadOnly:  true,
		})

		// add CA cert if defined
		if instance.Spec.KSMTLS.CaBundleSecretName != "" {
			ca := instance.Spec.KSMTLS.Ca
			volumes = append(volumes, ca.CreateVolume())
			mounts = append(mounts, ca.CreateVolumeMounts(nil)...)
		}

		args = append(args, fmt.Sprintf("--tls-config=%s", tlsConfFullPath))
	}

	container := corev1.Container{
		ImagePullPolicy: corev1.PullAlways,
		Image:           instance.Spec.KSMImage,
		Name:            KSMServiceName,
		Args:            args,
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
		VolumeMounts: mounts,
	}

	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KSMServiceName,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &KSMReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      KSMServiceName,
					Namespace: instance.Namespace,
					Labels:    labels,
				},
				Spec: corev1.PodSpec{
					AutomountServiceAccountToken: ptr.To(true),
					ServiceAccountName:           instance.RbacResourceName(),
					Containers:                   []corev1.Container{container},
					Volumes:                      volumes,
				},
			},
		},
	}

	if instance.Spec.NodeSelector != nil {
		ss.Spec.Template.Spec.NodeSelector = *instance.Spec.NodeSelector
	}
	if topology != nil {
		topology.ApplyTo(&ss.Spec.Template)
	}

	return ss, nil
}
