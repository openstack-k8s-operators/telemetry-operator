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

package ceilometercentral

import (
	corev1 "k8s.io/api/core/v1"

	telemetry "github.com/openstack-k8s-operators/telemetry-operator/pkg/telemetry"
)

// getVolumes - service volumes
func getVolumes(name string) []corev1.Volume {
	return append(telemetry.GetVolumes(name), corev1.Volume{
		Name: "sg-core-conf-yaml",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				Items: []corev1.KeyToPath{{
					Key:  "sg-core.conf.yaml",
					Path: "sg-core.conf.yaml",
				}},
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name + "-config-data",
				},
			},
		},
	})
}

// getVolumeMountsSgCore - VolumeMounts for SGCore container
func getVolumeMountsSgCore() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "sg-core-conf-yaml",
			MountPath: "/etc/sg-core.conf.yaml",
			SubPath:   "sg-core.conf.yaml",
		},
	}
}
