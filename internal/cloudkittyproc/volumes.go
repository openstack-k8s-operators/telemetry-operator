package cloudkittyproc

import (
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	"github.com/openstack-k8s-operators/telemetry-operator/internal/cloudkitty"
	"github.com/openstack-k8s-operators/telemetry-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
)

var config0440AccessMode int32 = 0440

// GetVolumes -
func GetVolumes(parentName string, name string, instance *telemetryv1.CloudKittyProc) []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: "config-data-custom",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &config0440AccessMode,
					SecretName:  name + "-config-data",
				},
			},
		},
	}

	if instance.Spec.CustomConfigsSecretName != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "custom-config",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					DefaultMode: &config0440AccessMode,
					SecretName:  instance.Spec.CustomConfigsSecretName,
				},
			},
		})
	}

	return append(cloudkitty.GetVolumes(parentName), volumes...)
}

// GetVolumeMounts - CloudKitty Proc VolumeMounts
// customConfigKeys are the keys from the CustomConfigsSecret; each is
// SubPath-mounted into /etc/cloudkitty/ (matching the API contract:
// "files from this secret will get copied into /etc/cloudkitty/").
func GetVolumeMounts(customConfigKeys []string) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "config-data-custom",
			MountPath: "/etc/cloudkitty/cloudkitty.conf.d/00-cloudkitty.conf",
			SubPath:   cloudkitty.DefaultsConfigFileName,
			ReadOnly:  true,
		},
		{
			Name:      "config-data-custom",
			MountPath: "/etc/cloudkitty/cloudkitty.conf.d/" + cloudkitty.CustomConfigFileName,
			SubPath:   cloudkitty.CustomConfigFileName,
			ReadOnly:  true,
		},
		{
			Name:      "config-data-custom",
			MountPath: "/etc/cloudkitty/cloudkitty.conf.d/" + cloudkitty.CustomServiceConfigFileName,
			SubPath:   cloudkitty.CustomServiceConfigFileName,
			ReadOnly:  true,
		},
		{
			Name:      "config-data-custom",
			MountPath: "/etc/cloudkitty/cloudkitty.conf.d/" + cloudkitty.CustomServiceConfigSecretsFileName,
			SubPath:   cloudkitty.CustomServiceConfigSecretsFileName,
			ReadOnly:  true,
		},
	}

	base := append(cloudkitty.GetVolumeMounts(), volumeMounts...)

	customConfigMounts := make([]corev1.VolumeMount, 0, len(customConfigKeys))
	for _, key := range customConfigKeys {
		customConfigMounts = append(customConfigMounts, corev1.VolumeMount{
			Name:      "custom-config",
			MountPath: "/etc/cloudkitty/" + key,
			SubPath:   key,
			ReadOnly:  true,
		})
	}

	// custom-config files override the default file mounted at the same path
	return utils.MergeCustomConfigMounts(base, customConfigMounts)
}
