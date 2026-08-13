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
	"github.com/openstack-k8s-operators/lib-common/modules/common/volume"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	"github.com/openstack-k8s-operators/telemetry-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
)

const (
	scriptVolume = "ceilometer-scripts"
	configVolume = "ceilometer-config-data"
)

var (
	config0440AccessMode int32 = 0440
	script0550AccessMode int32 = 0550
)

func getVolumes(instance *telemetryv1.Ceilometer) []corev1.Volume {
	vols := []corev1.Volume{
		{
			Name: "scripts",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &script0550AccessMode,
					SecretName:  scriptVolume,
				},
			},
		}, {
			Name: "config-data",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &config0440AccessMode,
					SecretName:  configVolume,
				},
			},
		}, {
			Name: "sg-core-conf-yaml",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &config0440AccessMode,
					Items: []corev1.KeyToPath{{
						Key:  "sg-core.conf.yaml",
						Path: "sg-core.conf.yaml",
					}},
					SecretName: configVolume,
				},
			},
		},
		volume.WritableDirVolume(volume.RunHttpdVolumeName),
		volume.WritableDirVolume(volume.VarLogHttpdVolumeName),
	}

	if instance.Spec.CustomConfigsSecretName != "" {
		vols = append(vols, corev1.Volume{
			Name: "custom-config",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &config0440AccessMode,
					SecretName:  instance.Spec.CustomConfigsSecretName,
				},
			},
		})
	}
	return vols
}

func customConfigMounts(customConfigKeys []string) []corev1.VolumeMount {
	mounts := make([]corev1.VolumeMount, 0, len(customConfigKeys))
	for _, key := range customConfigKeys {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "custom-config",
			MountPath: "/etc/ceilometer/" + key,
			SubPath:   key,
			ReadOnly:  true,
		})
	}
	return mounts
}

// getCentralVolumeMounts - ceilometer-central VolumeMounts
func getCentralVolumeMounts(customConfigKeys []string) []corev1.VolumeMount {
	vm := []corev1.VolumeMount{
		{
			Name:      "config-data",
			MountPath: "/etc/ceilometer/ceilometer.conf",
			SubPath:   "ceilometer.conf",
			ReadOnly:  true,
		},
		{
			Name:      "config-data",
			MountPath: "/etc/ceilometer/polling.yaml",
			SubPath:   "polling.yaml.j2",
			ReadOnly:  true,
		},
		{
			Name:      "config-data",
			MountPath: "/etc/ceilometer/ceilometer.conf.d/01-ceilometer-custom.conf",
			SubPath:   "custom.conf",
			ReadOnly:  true,
		},
	}
	// custom-config files override the default file mounted at the same path
	return utils.MergeCustomConfigMounts(vm, customConfigMounts(customConfigKeys))
}

// getNotificationVolumeMounts - ceilometer-notification VolumeMounts
func getNotificationVolumeMounts(customConfigKeys []string) []corev1.VolumeMount {
	vm := []corev1.VolumeMount{
		{
			Name:      "config-data",
			MountPath: "/etc/ceilometer/ceilometer.conf",
			SubPath:   "ceilometer.conf",
			ReadOnly:  true,
		},
		{
			Name:      "config-data",
			MountPath: "/etc/ceilometer/pipeline.yaml",
			SubPath:   "pipeline.yaml",
			ReadOnly:  true,
		},
		{
			Name:      "config-data",
			MountPath: "/etc/ceilometer/event_pipeline.yaml",
			SubPath:   "event_pipeline.yaml",
			ReadOnly:  true,
		},
		{
			Name:      "config-data",
			MountPath: "/etc/ceilometer/ceilometer.conf.d/01-ceilometer-custom.conf",
			SubPath:   "custom.conf",
			ReadOnly:  true,
		},
	}
	// custom-config files override the default file mounted at the same path
	return utils.MergeCustomConfigMounts(vm, customConfigMounts(customConfigKeys))
}

func getSgCoreVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "sg-core-conf-yaml",
			MountPath: "/etc/sg-core.conf.yaml",
			SubPath:   "sg-core.conf.yaml",
		},
	}
}

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
		volume.WritableDirVolumeMount(volume.RunHttpdVolumeName, volume.RunHttpdMountPath),
		volume.WritableDirVolumeMount(volume.VarLogHttpdVolumeName, volume.VarLogHttpdMountPath),
	}
}

// getHealthCheckVolumeMounts - health check script SubPath mounts for central and notification agents
func getHealthCheckVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "scripts",
			MountPath: CentralHCScript,
			SubPath:   "centralhealth.py",
			ReadOnly:  true,
		},
		{
			Name:      "scripts",
			MountPath: NotificationHCScript,
			SubPath:   "notificationhealth.py",
			ReadOnly:  true,
		},
	}
}
