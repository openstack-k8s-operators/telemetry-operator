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

package autoscaling

import (
	"github.com/openstack-k8s-operators/lib-common/modules/common/volume"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	"github.com/openstack-k8s-operators/telemetry-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
)

const (
	configVolume = "aodh-config-data"
)

var config0440AccessMode int32 = 0440

// getVolumes - service volumes
func getVolumes(instance *telemetryv1.Autoscaling) []corev1.Volume {
	vols := []corev1.Volume{
		{
			Name: "config-data",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &config0440AccessMode,
					SecretName:  configVolume,
				},
			},
		},
		volume.WritableDirVolume(volume.RunHttpdVolumeName),
		volume.WritableDirVolume(volume.VarLogHttpdVolumeName),
	}

	if instance.Spec.Aodh.CustomConfigsSecretName != "" {
		vols = append(vols, corev1.Volume{
			Name: "custom-config",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &config0440AccessMode,
					SecretName:  instance.Spec.Aodh.CustomConfigsSecretName,
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
			MountPath: "/etc/aodh/" + key,
			SubPath:   key,
			ReadOnly:  true,
		})
	}
	return mounts
}

// getAPIVolumeMounts - aodh-api (httpd) VolumeMounts
func getAPIVolumeMounts(customConfigKeys []string) []corev1.VolumeMount {
	vm := []corev1.VolumeMount{
		{
			Name:      "config-data",
			MountPath: "/etc/aodh/aodh.conf",
			SubPath:   "aodh.conf",
			ReadOnly:  true,
		},
		{
			Name:      "config-data",
			MountPath: "/etc/aodh/aodh.conf.d/01-aodh-custom.conf",
			SubPath:   "custom.conf",
			ReadOnly:  true,
		},
		{
			Name:      "config-data",
			MountPath: "/etc/httpd/conf.d/00wsgi-aodh.conf",
			SubPath:   "wsgi-aodh.conf",
			ReadOnly:  true,
		},
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
			Name:      "config-data",
			MountPath: "/etc/my.cnf",
			SubPath:   "my.cnf",
			ReadOnly:  true,
		},
		volume.WritableDirVolumeMount(volume.RunHttpdVolumeName, volume.RunHttpdMountPath),
		volume.WritableDirVolumeMount(volume.VarLogHttpdVolumeName, volume.VarLogHttpdMountPath),
	}
	// custom-config files override the default file mounted at the same path
	return utils.MergeCustomConfigMounts(vm, customConfigMounts(customConfigKeys))
}

// getEvaluatorVolumeMounts - aodh-evaluator VolumeMounts
func getEvaluatorVolumeMounts(customConfigKeys []string) []corev1.VolumeMount {
	vm := []corev1.VolumeMount{
		{
			Name:      "config-data",
			MountPath: "/etc/aodh/aodh.conf",
			SubPath:   "aodh.conf",
			ReadOnly:  true,
		},
		{
			Name:      "config-data",
			MountPath: "/etc/aodh/aodh.conf.d/01-aodh-custom.conf",
			SubPath:   "custom.conf",
			ReadOnly:  true,
		},
		{
			Name:      "config-data",
			MountPath: "/etc/openstack/prometheus.yaml",
			SubPath:   "prometheus.yaml",
			ReadOnly:  true,
		},
		{
			Name:      "config-data",
			MountPath: "/etc/my.cnf",
			SubPath:   "my.cnf",
			ReadOnly:  true,
		},
	}
	// custom-config files override the default file mounted at the same path
	return utils.MergeCustomConfigMounts(vm, customConfigMounts(customConfigKeys))
}

// getWorkerVolumeMounts - aodh-notifier/listener/dbsync VolumeMounts
func getWorkerVolumeMounts(customConfigKeys []string) []corev1.VolumeMount {
	vm := []corev1.VolumeMount{
		{
			Name:      "config-data",
			MountPath: "/etc/aodh/aodh.conf",
			SubPath:   "aodh.conf",
			ReadOnly:  true,
		},
		{
			Name:      "config-data",
			MountPath: "/etc/aodh/aodh.conf.d/01-aodh-custom.conf",
			SubPath:   "custom.conf",
			ReadOnly:  true,
		},
		{
			Name:      "config-data",
			MountPath: "/etc/my.cnf",
			SubPath:   "my.cnf",
			ReadOnly:  true,
		},
	}
	// custom-config files override the default file mounted at the same path
	return utils.MergeCustomConfigMounts(vm, customConfigMounts(customConfigKeys))
}

// getCustomPrometheusCaVolume - Volume for CA certificate of user deployed Prometheus
func getCustomPrometheusCaVolume(secretName string) corev1.Volume {
	return corev1.Volume{
		Name: "custom-prometheus-ca",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	}
}

// getCustomPrometheusCaVolumeMount - VolumeMount for CA certificate of user deployed Prometheus
func getCustomPrometheusCaVolumeMount(fileName string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      "custom-prometheus-ca",
		MountPath: CustomPrometheusCaCertFolderPath + fileName,
		SubPath:   fileName,
		ReadOnly:  true,
	}
}
