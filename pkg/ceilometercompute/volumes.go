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
	"github.com/openstack-k8s-operators/lib-common/modules/storage"
	corev1 "k8s.io/api/core/v1"

	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	telemetry "github.com/openstack-k8s-operators/telemetry-operator/pkg/telemetry"
)

func getExtraMounts(name string, instance *telemetryv1.CeilometerCompute) []storage.VolMounts {
	volumes := []corev1.Volume{
		{
			Name: "sshkey",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: instance.Spec.DataplaneSshSecret,
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
					DefaultMode: &telemetry.Config0640AccessMode,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: instance.Spec.DataplaneInventoryConfigMap,
					},
				},
			},
		}, {
			Name: "playbook",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: name + "-external-compute-playbook",
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
			Name:      "playbook",
			MountPath: "/runner/project/deploy-ceilometer.yaml",
			SubPath:   "deploy-ceilometer.yaml",
		},
	}

	return []storage.VolMounts{
		{
			Volumes: append(telemetry.GetVolumes(name), volumes...),
			Mounts:  append(telemetry.GetVolumeMounts(), mounts...),
		},
	}
}
