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
	"fmt"

	tls "github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	"github.com/openstack-k8s-operators/telemetry-operator/pkg/telemetry"
	monv1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RabbitMQPodMonitor creates a PodMonitor CR for monitoring a RabbitMQ cluster
func RabbitMQPodMonitor(
	instance *telemetryv1.MetricStorage,
	labels map[string]string,
	rabbitMQClusterName string,
	tlsEnabled bool,
) *monv1.PodMonitor {
	metricRelabelConfigs := []*monv1.RelabelConfig{
		{
			Action:       "labeldrop",
			Regex:        "namespace",
			SourceLabels: []monv1.LabelName{},
		},
		{
			Action:       "labeldrop",
			Regex:        "job",
			SourceLabels: []monv1.LabelName{},
		},
		{
			Action:       "labeldrop",
			Regex:        "publisher",
			SourceLabels: []monv1.LabelName{},
		},
	}
	podMonitorName := fmt.Sprintf("%s-%s", telemetry.ServiceName, rabbitMQClusterName)
	selector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app.kubernetes.io/name": rabbitMQClusterName,
		},
	}
	serverName := fmt.Sprintf("%s.%s.svc", rabbitMQClusterName, instance.Namespace)
	return PodMonitor(instance, labels, podMonitorName, metricRelabelConfigs,
		selector, serverName, RabbitMQPrometheusPortName, tlsEnabled)
}

// PodMonitor creates a PodMonitor CR
// NOTE: Current implementation allows single metric endpoint per pod
func PodMonitor(
	instance *telemetryv1.MetricStorage,
	labels map[string]string,
	name string,
	metricRelabelConfigs []*monv1.RelabelConfig,
	selector metav1.LabelSelector,
	serverName string,
	port string,
	tlsEnabled bool,
) *monv1.PodMonitor {
	var scrapeInterval monv1.Duration
	if instance.Spec.MonitoringStack != nil && instance.Spec.MonitoringStack.ScrapeInterval != "" {
		scrapeInterval = monv1.Duration(instance.Spec.MonitoringStack.ScrapeInterval)
	} else if instance.Spec.CustomMonitoringStack != nil &&
		instance.Spec.CustomMonitoringStack.PrometheusConfig != nil &&
		instance.Spec.CustomMonitoringStack.PrometheusConfig.ScrapeInterval != nil &&
		*instance.Spec.CustomMonitoringStack.PrometheusConfig.ScrapeInterval != monv1.Duration("") {
		scrapeInterval = *instance.Spec.CustomMonitoringStack.PrometheusConfig.ScrapeInterval
	} else {
		scrapeInterval = telemetryv1.DefaultScrapeInterval
	}

	podMonitor := &monv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: monv1.PodMonitorSpec{
			Selector: selector,
			PodMetricsEndpoints: []monv1.PodMetricsEndpoint{
				{
					Interval:             scrapeInterval,
					MetricRelabelConfigs: metricRelabelConfigs,
					Scheme:               "http",
					Port:                 port,
				},
			},
		},
	}

	if tlsEnabled {
		tlsConfig := monv1.PodMetricsEndpointTLSConfig{
			SafeTLSConfig: monv1.SafeTLSConfig{
				CA: monv1.SecretOrConfigMap{
					Secret: &v1.SecretKeySelector{
						Key: tls.CABundleKey,
						LocalObjectReference: v1.LocalObjectReference{
							Name: instance.Spec.PrometheusTLS.CaBundleSecretName,
						},
					},
				},
				ServerName: serverName,
			},
		}
		podMonitor.Spec.PodMetricsEndpoints[0].TLSConfig = &tlsConfig
		podMonitor.Spec.PodMetricsEndpoints[0].Scheme = "https"
	}

	return podMonitor
}
