/*
Copyright 2022.

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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TODO: Once we update controller-runtime to v0.15+ ,
//       k8s.io/* to v0.27+ and
//       obo-prometheus-operator to at least v0.65.1-rhobs1,
//       switch this to structured definition

// ScrapeConfig creates a ScrapeConfig CR
func ScrapeConfig(
	instance *telemetryv1.MetricStorage,
	labels map[string]string,
	targets []string,
) *unstructured.Unstructured {
	var scrapeInterval string
	if instance.Spec.MonitoringStack != nil && instance.Spec.MonitoringStack.ScrapeInterval != "" {
		scrapeInterval = instance.Spec.MonitoringStack.ScrapeInterval
		// TODO: Uncomment the following else if once we update to OBOv0.0.21
		//} else if instance.Spec.CustomMonitoringStack.PrometheusConfig.ScrapeInterval {
		//	scrapeInterval = instance.Spec.CustomMonitoringStack.PrometheusConfig.ScrapeInterval
	} else {
		scrapeInterval = telemetryv1.DefaultScrapeInterval
	}
	scrapeConfig := &unstructured.Unstructured{}
	scrapeConfig.SetUnstructuredContent(map[string]interface{}{
		"spec": map[string]interface{}{
			"scrapeInterval": scrapeInterval,
			"staticConfigs": []interface{}{
				map[string]interface{}{
					"targets": targets,
				},
			},
		},
	})
	// NOTE: SetUnstructuredContent will overwrite any data, including GVK, name, ...
	//       Any other Set* function call must be done after SetUnstructuredContent
	scrapeConfig.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "monitoring.rhobs",
		Version: "v1alpha1",
		Kind:    "ScrapeConfig",
	})
	scrapeConfig.SetName(instance.Name)
	scrapeConfig.SetNamespace(instance.Namespace)
	scrapeConfig.SetLabels(labels)
	return scrapeConfig
}
