package cloudkittyapi

import (
	"github.com/openstack-k8s-operators/telemetry-operator/pkg/cloudkitty"
	corev1 "k8s.io/api/core/v1"
)

// GetVolumes -
func GetVolumes(parentName string, name string) []corev1.Volume {
	var config0644AccessMode int32 = 0644

	volumes := []corev1.Volume{
		{
			Name: "config-data-custom",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &config0644AccessMode,
					SecretName:  name + "-config-data",
				},
			},
		},
		{
			Name: "logs",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{Medium: ""},
			},
		},
	}

	return append(cloudkitty.GetVolumes(parentName), volumes...)
}

// GetVolumeMounts - CloudKitty API VolumeMounts
func GetVolumeMounts(parentName string) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "config-data-custom",
			MountPath: "/etc/cloudkitty/cloudkitty.conf.d",
			ReadOnly:  true,
		},
		GetLogVolumeMount(),
	}

	return append(cloudkitty.GetVolumeMounts(parentName), volumeMounts...)
}

// GetLogVolumeMount - CloudKitty API LogVolumeMount
func GetLogVolumeMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      "logs",
		MountPath: "/var/log/cloudkitty",
		ReadOnly:  false,
	}
}
