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
	corev1 "k8s.io/api/core/v1"
)

const (
	scriptVolume = "ceilometer-scripts"
	configVolume = "ceilometer-config-data"
	logVolume    = "logs"
	cacertVolume = "combined-ca-bundle"
)

var (
	configMode int32 = 0640
	scriptMode int32 = 0740
	cacertMode int32 = 0444
)

func getVolumes(name string) []corev1.Volume {
	return []corev1.Volume{
		{
			Name: "scripts",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &scriptMode,
					SecretName:  scriptVolume,
				},
			},
		}, {
			Name: "config-data",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &configMode,
					SecretName:  configVolume,
				},
			},
		}, {
			Name: "sg-core-conf-yaml",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &configMode,
					Items: []corev1.KeyToPath{{
						Key:  "sg-core.conf.yaml",
						Path: "sg-core.conf.yaml",
					}},
					SecretName: configVolume,
				},
			},
		}, {
			Name: "ca-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &cacertMode,
					SecretName:  cacertVolume,
				},
			},
		},
	}
}

// getVolumeMounts - general VolumeMounts
func getVolumeMounts(serviceName string) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "scripts",
			MountPath: "/var/lib/openstack/bin",
			ReadOnly:  true,
		},
		{
			Name:      "config-data",
			MountPath: "/var/lib/openstack/config",
			ReadOnly:  true,
		},
		{
			Name:      "config-data",
			MountPath: "/var/lib/kolla/config_files/config.json",
			SubPath:   serviceName + "-config.json",
			ReadOnly:  true,
		},
		{
			Name:      "ca-certs",
			MountPath: "/etc/pki/ca-trust/extracted/pem",
			ReadOnly:  true,
		},
	}
}

// getSgCoreVolumeMounts - VolumeMounts for SGCore container
func getSgCoreVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "sg-core-conf-yaml",
			MountPath: "/etc/sg-core.conf.yaml",
			SubPath:   "sg-core.conf.yaml",
		},
	}
}
