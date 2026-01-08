package cloudkittyapi

import (
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	"github.com/openstack-k8s-operators/telemetry-operator/internal/cloudkitty"
	corev1 "k8s.io/api/core/v1"
)

var (
	configMode int32 = 0640
)

// GetVolumes -
func GetVolumes(parentName string, name string, instance *telemetryv1.CloudKittyAPI) []corev1.Volume {
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

	if instance.Spec.CustomConfigsSecretName != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "custom-config",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &configMode,
					SecretName:  instance.Spec.CustomConfigsSecretName,
				},
			},
		})
	}

	return append(cloudkitty.GetVolumes(parentName), volumes...)
}

// GetVolumeMounts - CloudKitty API VolumeMounts
func GetVolumeMounts(instance *telemetryv1.CloudKittyAPI) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "config-data-custom",
			MountPath: "/var/lib/openstack/service-config/",
			ReadOnly:  true,
		},
		GetLogVolumeMount(),
	}

	if instance.Spec.CustomConfigsSecretName != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "custom-config",
			MountPath: "/var/lib/openstack/custom-config",
			ReadOnly:  true,
		})
	}

	return append(cloudkitty.GetVolumeMounts(cloudkitty.ServiceName+"-api"), volumeMounts...)
}

// GetLogVolumeMount - CloudKitty API LogVolumeMount
func GetLogVolumeMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      "logs",
		MountPath: "/var/log/cloudkitty",
		ReadOnly:  false,
	}
}
