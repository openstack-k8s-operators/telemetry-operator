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
	"fmt"

	env "github.com/openstack-k8s-operators/lib-common/modules/common/env"
	"github.com/openstack-k8s-operators/lib-common/modules/common/pod"
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	"github.com/openstack-k8s-operators/lib-common/modules/users"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	monv1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var aetosConfigMode int32 = 0440

// PrometheusAetosSidecar defines the merge patch for the Prometheus CR to add the Aetos sidecar container.
func PrometheusAetosSidecar(
	instance *telemetryv1.MetricStorage,
	configHash string,
) (monv1.Prometheus, error) {
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
					SecretName:  AetosConfigSecret,
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

	if instance.Spec.Aetos.TLS.CaBundleSecretName != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "aetos-ca-bundle",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: instance.Spec.Aetos.TLS.CaBundleSecretName,
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

	// Mount per-endpoint serving certs when Aetos API TLS is enabled. Both the
	// internal and public Services target this single httpd listener, so each
	// endpoint gets its own vhost (selected via SNI) with its own certificate.
	for _, endpt := range []service.Endpoint{service.EndpointInternal, service.EndpointPublic} {
		if !instance.Spec.Aetos.TLS.API.Enabled(endpt) {
			continue
		}
		var tlsEndptCfg tls.GenericService
		switch endpt {
		case service.EndpointPublic:
			tlsEndptCfg = instance.Spec.Aetos.TLS.API.Public
		case service.EndpointInternal:
			tlsEndptCfg = instance.Spec.Aetos.TLS.API.Internal
		}
		svc, err := tlsEndptCfg.ToService()
		if err != nil {
			return monv1.Prometheus{}, err
		}
		certMount := fmt.Sprintf("/etc/pki/tls/certs/%s.crt", endpt.String())
		keyMount := fmt.Sprintf("/etc/pki/tls/private/%s.key", endpt.String())
		svc.CertMount = &certMount
		svc.KeyMount = &keyMount
		// NOTE(jwysogla): secret files are root-owned, so mount them group-readable
		// for the non-root aetos user; this works only because COO sets the pod's
		// fsGroup, which kubelet applies to secret volumes.
		volumes = append(volumes, corev1.Volume{
			Name: endpt.String() + "-tls-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &aetosConfigMode,
					SecretName:  svc.SecretName,
				},
			},
		})
		volumeMounts = append(volumeMounts, svc.CreateVolumeMounts(endpt.String())...)
	}

	envVars := map[string]env.Setter{}
	envVars["CONFIG_HASH"] = env.SetValue(configHash)

	// kubelet probes send no SNI, so httpd serves the default (internal) vhost; match its TLS setting.
	probe := corev1.Probe{
		TimeoutSeconds:      30,
		PeriodSeconds:       30,
		InitialDelaySeconds: 5,
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/",
				Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(AetosPort)},
			},
		},
	}
	if instance.Spec.Aetos.TLS.API.Enabled(service.EndpointInternal) {
		probe.HTTPGet.Scheme = corev1.URISchemeHTTPS
	}
	readinessProbe := probe.DeepCopy()
	livenessProbe := probe.DeepCopy()

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
						Image:           instance.Spec.Aetos.Image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Command:         []string{"/usr/sbin/httpd"},
						Args:            []string{"-DFOREGROUND"},
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: int32(AetosPort),
								Name:          "aetos",
							},
						},
						SecurityContext: pod.RestrictiveSecurityContext(users.AetosUID, users.AetosGID),
						Env:             env.MergeEnvs([]corev1.EnvVar{}, envVars),
						VolumeMounts:    volumeMounts,
						ReadinessProbe:  readinessProbe,
						LivenessProbe:   livenessProbe,
					},
				},
				Volumes: volumes,
			},
		},
	}
	return prom, nil
}
