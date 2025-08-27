package cloudkitty

import (
	corev1 "k8s.io/api/core/v1"
)

var (
	// scriptMode is the default permissions mode for Scripts volume
	scriptMode int32 = 0740
	// configMode is the 640 permissions mode
	configMode int32 = 0640
)

// GetVolumes - service volumes
func GetVolumes(name string) []corev1.Volume {
	return []corev1.Volume{
		{
			Name: "scripts",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &scriptMode,
					SecretName:  name + "-scripts",
				},
			},
		}, {
			Name: "config-data",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &configMode,
					SecretName:  name + "-config-data",
				},
			},
		},
	}
}

// GetVolumeMounts - general VolumeMounts
func GetVolumeMounts(serviceName string) []corev1.VolumeMount {
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
