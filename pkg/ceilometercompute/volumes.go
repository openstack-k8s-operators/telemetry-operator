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

package ceilometercompute

import (
	corev1 "k8s.io/api/core/v1"

	storage "github.com/openstack-k8s-operators/lib-common/modules/storage"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	telemetry "github.com/openstack-k8s-operators/telemetry-operator/pkg/telemetry"
)

var (
	// ScriptsVolumeDefaultMode is the default permissions mode for Scripts volume
	ScriptsVolumeDefaultMode int32 = 0755
	// Config0640AccessMode is the 640 permissions mode
	Config0640AccessMode int32 = 0640
)

// getVolumes - service volumes
func getVolumes(name string) []corev1.Volume {
	return []corev1.Volume{
		{
			Name: "scripts",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode: &ScriptsVolumeDefaultMode,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: name + "-scripts",
					},
				},
			},
		}, {
			Name: "config-data",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode: &Config0640AccessMode,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: name + "-config-data",
					},
				},
			},
		}, {
			Name: "config-data-merged",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{Medium: ""},
			},
		},
	}
}

// getInitVolumeMounts - general init task VolumeMounts
func getInitVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "scripts",
			MountPath: "/usr/local/bin/container-scripts",
			ReadOnly:  true,
		},
		{
			Name:      "config-data",
			MountPath: "/var/lib/config-data/default",
			ReadOnly:  true,
		},
		{
			Name:      "config-data-merged",
			MountPath: "/var/lib/config-data/merged",
			ReadOnly:  false,
		},
	}
}

// getVolumeMounts - general VolumeMounts
func getVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "scripts",
			MountPath: "/usr/local/bin/container-scripts",
			ReadOnly:  true,
		},
		{
			Name:      "config-data-merged",
			MountPath: "/var/lib/config-data/merged",
			ReadOnly:  false,
		},
	}
}

func getExtraMounts(name string, instance *telemetryv1.CeilometerCompute) []storage.VolMounts {
	volumes := []corev1.Volume{
		{
			Name: "sshkey",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: instance.Spec.DataplaneSSHSecret,
					Items: []corev1.KeyToPath{
						{
							Key:  "ssh-privatekey",
							Path: "ssh_key",
						},
					},
				},
			},
		}, {
			Name: "inventory",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode: &Config0640AccessMode,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: instance.Spec.DataplaneInventoryConfigMap,
					},
				},
			},
		}, {
			Name: "extravars",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: telemetry.ServiceName + "-ceilometer-extravars",
					},
				},
			},
		},
	}

	mounts := []corev1.VolumeMount{
		{
			Name:      "sshkey",
			MountPath: "/runner/env/ssh_key",
			SubPath:   "ssh_key",
		}, {
			Name:      "inventory",
			MountPath: "/runner/inventory/hosts",
			SubPath:   "inventory",
		},
		{
			Name:      "extravars",
			MountPath: "/runner/env/extravars",
			SubPath:   "extravars",
		},
	}

	return []storage.VolMounts{
		{
			Volumes: append(getVolumes(name), volumes...),
			Mounts:  append(getVolumeMounts(), mounts...),
		},
	}
}
