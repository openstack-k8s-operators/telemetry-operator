/*
Copyright 2025.

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

package metricstorage

import (
	env "github.com/openstack-k8s-operators/lib-common/modules/common/env"
	"github.com/openstack-k8s-operators/lib-common/modules/common/pod"
	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	monv1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	aetosConfigSecret = "aetos-config-data"

	// TODO: Switch to users.AetosUID / users.AetosGID once
	// https://github.com/openstack-k8s-operators/lib-common/pull/746 merges
	aetosUID int64 = 42404
	aetosGID int64 = 42404
)

var aetosConfigMode int32 = 0440

// PrometheusAetosSidecar defines patch for prometheus CR to add the Aetos sidecar container
func PrometheusAetosSidecar(
	instance *telemetryv1.MetricStorage,
	configHash string,
) monv1.Prometheus {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "aetos-config-data",
			MountPath: "/etc/aetos/aetos.conf",
			SubPath:   "aetos.conf",
			ReadOnly:  true,
		},
		{
			Name:      "aetos-config-data",
			MountPath: "/etc/httpd/conf/httpd.conf",
			SubPath:   "httpd.conf",
			ReadOnly:  true,
		},
		{
			Name:      "aetos-config-data",
			MountPath: "/etc/httpd/conf.d/wsgi-aetos.conf",
			SubPath:   "wsgi-aetos.conf",
			ReadOnly:  true,
		},
		{
			Name:      "aetos-config-data",
			MountPath: "/etc/httpd/conf.d/ssl.conf",
			SubPath:   "ssl.conf",
			ReadOnly:  true,
		},
		{
			Name:      "aetos-run-httpd",
			MountPath: "/run/httpd",
		},
		{
			Name:      "aetos-var-log-httpd",
			MountPath: "/var/log/httpd",
		},
	}

	volumes := []corev1.Volume{
		{
			Name: "aetos-config-data",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &aetosConfigMode,
					SecretName:  aetosConfigSecret,
				},
			},
		},
		{
			Name: "aetos-run-httpd",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "aetos-var-log-httpd",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	if instance.Spec.PrometheusTLS.CaBundleSecretName != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "aetos-ca-bundle",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: instance.Spec.PrometheusTLS.CaBundleSecretName,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "aetos-ca-bundle",
			MountPath: tls.DownstreamTLSCABundlePath,
			SubPath:   tls.CABundleKey,
			ReadOnly:  true,
		})
	}

	envVars := map[string]env.Setter{}
	envVars["CONFIG_HASH"] = env.SetValue(configHash)

	prom := monv1.Prometheus{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Prometheus",
			APIVersion: "monitoring.rhobs/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
		Spec: monv1.PrometheusSpec{
			CommonPrometheusFields: monv1.CommonPrometheusFields{
				Containers: []corev1.Container{
					{
						Name:            "aetos",
						Image:           instance.Spec.AetosImage,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Command:         []string{"/usr/sbin/httpd"},
						Args:            []string{"-DFOREGROUND"},
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 8989,
								Name:          "aetos",
							},
						},
						SecurityContext: pod.RestrictiveSecurityContext(aetosUID, aetosGID),
						Env:             env.MergeEnvs([]corev1.EnvVar{}, envVars),
						VolumeMounts:    volumeMounts,
					},
				},
				Volumes: volumes,
			},
		},
	}
	return prom
}
