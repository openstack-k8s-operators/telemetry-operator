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
	"context"
	"fmt"
	"net"
	"sort"

	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	tls "github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	mysqldexporter "github.com/openstack-k8s-operators/telemetry-operator/internal/mysqldexporter"
	rabbitmqv1 "github.com/rabbitmq/cluster-operator/api/v1beta1"
	monv1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1"
	monv1alpha1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LabeledTarget represents a monitoring target with IP and FQDN labels
type LabeledTarget struct {
	IP   string
	FQDN string
}

func retrieveScrapeInterval(instance *telemetryv1.MetricStorage) monv1.Duration {
	if instance.Spec.MonitoringStack != nil && instance.Spec.MonitoringStack.ScrapeInterval != "" {
		return monv1.Duration(instance.Spec.MonitoringStack.ScrapeInterval)
	} else if instance.Spec.CustomMonitoringStack != nil &&
		instance.Spec.CustomMonitoringStack.PrometheusConfig != nil &&
		instance.Spec.CustomMonitoringStack.PrometheusConfig.ScrapeInterval != nil &&
		*instance.Spec.CustomMonitoringStack.PrometheusConfig.ScrapeInterval != monv1.Duration("") {
		return *instance.Spec.CustomMonitoringStack.PrometheusConfig.ScrapeInterval
	}
	return telemetryv1.DefaultScrapeInterval
}

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

// ScrapeConfigRabbitMQ creates a ScrapeConfig CR for scraping all nodes of
// a RabbitMQ cluster
func ScrapeConfigRabbitMQ(
	ctx context.Context,
	instance *telemetryv1.MetricStorage,
	labels map[string]string,
	helper *helper.Helper,
	rabbit rabbitmqv1.RabbitmqCluster,
	tlsEnabled bool,
) (*monv1alpha1.ScrapeConfig, error) {
	scrapeInterval := retrieveScrapeInterval(instance)

	// NOTE: we use the Endpointslices object to retrieve a list of all pods
	//       in each cluster, because RabbitMQ can run with multiple
	//       replicas per cluster. In that case, each replica will have
	//       different values to each metric, so we need to scrape
	//       all of them. Scraping the cluster based on its DNS name
	//       isn't enough.
	endpointSlices := discoveryv1.EndpointSliceList{}

	labelSelector := map[string]string{
		"kubernetes.io/service-name": rabbit.Name,
	}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(labelSelector),
	}
	err := helper.GetClient().List(ctx, &endpointSlices, listOpts...)
	if err != nil {
		return nil, err
	}

	var staticConfigs []monv1alpha1.StaticConfig
	for _, slice := range endpointSlices.Items {
		for _, endpoint := range slice.Endpoints {
			targets := []monv1alpha1.Target{}
			for _, address := range endpoint.Addresses {
				target := ""
				if tlsEnabled {
					target = net.JoinHostPort(address, RabbitMQPrometheusPortTLS)
				} else {
					target = net.JoinHostPort(address, RabbitMQPrometheusPortNoTLS)
				}
				targets = append(targets, monv1alpha1.Target(target))
			}
			staticConfigs = append(staticConfigs, monv1alpha1.StaticConfig{
				Targets: targets,
				Labels: map[monv1.LabelName]string{
					monv1.LabelName("pod"): endpoint.TargetRef.Name,
				},
			})
		}
	}

	scrapeConfig := &monv1alpha1.ScrapeConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: monv1alpha1.ScrapeConfigSpec{
			MetricRelabelConfigs: []*monv1.RelabelConfig{
				{
					Action:       "labeldrop",
					Regex:        "job",
					SourceLabels: []monv1.LabelName{},
				},
			},
			ScrapeInterval: &scrapeInterval,
			StaticConfigs:  staticConfigs,
		},
	}

	if tlsEnabled {
		tlsConfig := monv1.SafeTLSConfig{
			CA: monv1.SecretOrConfigMap{
				Secret: &corev1.SecretKeySelector{
					Key: tls.CABundleKey,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: instance.Spec.PrometheusTLS.CaBundleSecretName,
					},
				},
			},
			ServerName: fmt.Sprintf("%s.%s.svc", rabbit.Name, rabbit.Namespace),
		}
		scheme := "HTTPS"
		scrapeConfig.Spec.Scheme = &scheme
		scrapeConfig.Spec.TLSConfig = &tlsConfig
	}

	return scrapeConfig, nil
}

// ScrapeConfig creates a ScrapeConfig CR
func ScrapeConfig(
	instance *telemetryv1.MetricStorage,
	labels map[string]string,
	targets any,
	tlsEnabled bool,
) *monv1alpha1.ScrapeConfig {
	scrapeInterval := retrieveScrapeInterval(instance)

	var staticConfigs []monv1alpha1.StaticConfig
	if ips, ok := targets.([]string); ok {
		sort.Strings(ips)
		var convertedTargets []monv1alpha1.Target
		for _, t := range ips {
			convertedTargets = append(convertedTargets, monv1alpha1.Target(t))
		}
		staticConfigs = append(staticConfigs, monv1alpha1.StaticConfig{
			Targets: convertedTargets,
		})
	} else if labeledTargets, ok := targets.([]LabeledTarget); ok {
		sort.Slice(labeledTargets, func(i, j int) bool {
			return labeledTargets[i].IP < labeledTargets[j].IP
		})
		for _, t := range labeledTargets {
			staticConfigs = append(staticConfigs, monv1alpha1.StaticConfig{
				Targets: []monv1alpha1.Target{monv1alpha1.Target(t.IP)},
				Labels: map[monv1.LabelName]string{
					"fqdn": t.FQDN,
				},
			})
		}
	}

	scrapeConfig := &monv1alpha1.ScrapeConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: monv1alpha1.ScrapeConfigSpec{
			MetricRelabelConfigs: []*monv1.RelabelConfig{
				{
					Action:       "labeldrop",
					Regex:        "pod",
					SourceLabels: []monv1.LabelName{},
				},
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
			},
			ScrapeInterval: &scrapeInterval,
			StaticConfigs:  staticConfigs,
		},
	}

	if tlsEnabled {
		tlsConfig := monv1.SafeTLSConfig{
			CA: monv1.SecretOrConfigMap{
				Secret: &corev1.SecretKeySelector{
					Key: tls.CABundleKey,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: instance.Spec.PrometheusTLS.CaBundleSecretName,
					},
				},
			},
		}
		scheme := "HTTPS"
		scrapeConfig.Spec.Scheme = &scheme
		scrapeConfig.Spec.TLSConfig = &tlsConfig
	}

	return scrapeConfig
}
