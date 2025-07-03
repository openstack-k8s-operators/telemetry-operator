/*
Copyright 2025.

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

package metricstorage

import (
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	monv1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PrometheusNAD defines patch for prometheus CR to add NAD annotations
func PrometheusNAD(
	instance *telemetryv1.MetricStorage,
	networkAnnotations map[string]string,
) monv1.Prometheus {
	prom := monv1.Prometheus{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Prometheus",
			APIVersion: "monitoring.rhobs/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
		Spec: monv1.PrometheusSpec{
			CommonPrometheusFields: monv1.CommonPrometheusFields{
				PodMetadata: &monv1.EmbeddedObjectMetadata{
					Annotations: networkAnnotations,
				},
			},
		},
	}
	return prom
}
