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
)

var (
	configMode int32 = 0640
	scriptMode int32 = 0740
)

func getVolumes() []corev1.Volume {
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
		},
		{
			Name: "run-httpd",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{Medium: ""},
			},
		},
		{
			Name: "log-httpd",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{Medium: ""},
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

// getHttpdVolumeMounts - Returns the VolumeMounts used by the httpd sidecar
func getHttpdVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "config-data",
			MountPath: "/etc/httpd/conf/httpd.conf",
			SubPath:   "httpd.conf",
			ReadOnly:  true,
		},
		{
			Name:      "config-data",
			MountPath: "/etc/httpd/conf.d/ssl.conf",
			SubPath:   "ssl.conf",
			ReadOnly:  true,
		},
		{
			Name:      "run-httpd",
			MountPath: "/run/httpd",
			ReadOnly:  false,
		},
		{
			Name:      "log-httpd",
			MountPath: "/var/log/httpd",
			ReadOnly:  false,
		},
	}
}
