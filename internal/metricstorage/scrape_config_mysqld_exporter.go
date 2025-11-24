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
	"fmt"

	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	mysqldexporter "github.com/openstack-k8s-operators/telemetry-operator/internal/mysqldexporter"
	monv1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1"
	monv1alpha1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"k8s.io/utils/ptr"
)

// ScrapeConfigMysqldExporter creates a ScrapeConfig CR
func ScrapeConfigMysqldExporter(
	instance *telemetryv1.MetricStorage,
	labels map[string]string,
	targets []string,
	tlsEnabled bool,
) *monv1alpha1.ScrapeConfig {
	// Use our normal scrape config as a base, which will be extended
	scrapeConfig := ScrapeConfig(instance, labels, targets, tlsEnabled)

	scrapeConfig.Spec.MetricsPath = ptr.To("/probe")
	scrapeConfig.Spec.RelabelConfigs = []*monv1.RelabelConfig{
		{
			Action: "Replace",
			SourceLabels: []monv1.LabelName{
				"__address__",
			},
			TargetLabel: "__param_target",
		},
		{
			Action: "Replace",
			SourceLabels: []monv1.LabelName{
				"__address__",
			},
			TargetLabel: "__param_auth_module",
			Regex:       "(.*):(.*)",
			Replacement: "client.$1",
		},
		{
			Action: "Replace",
			SourceLabels: []monv1.LabelName{
				"__param_target",
			},
			TargetLabel: "instance",
		},
		{
			Action:      "Replace",
			TargetLabel: "__address__",
			Replacement: fmt.Sprintf("%s.%s.svc:%d", mysqldexporter.ServiceName, instance.Namespace, mysqldexporter.MysqldExporterPort),
		},
	}
	return scrapeConfig
}
