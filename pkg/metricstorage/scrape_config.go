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

	tls "github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	monv1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1"
	monv1alpha1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ScrapeConfig creates a ScrapeConfig CR
func ScrapeConfig(
	instance *telemetryv1.MetricStorage,
	labels map[string]string,
	targets []string,
	tlsEnabled bool,
) *monv1alpha1.ScrapeConfig {
	var scrapeInterval monv1.Duration
	if instance.Spec.MonitoringStack != nil && instance.Spec.MonitoringStack.ScrapeInterval != "" {
		scrapeInterval = monv1.Duration(instance.Spec.MonitoringStack.ScrapeInterval)
	} else if instance.Spec.CustomMonitoringStack != nil && *instance.Spec.CustomMonitoringStack.PrometheusConfig.ScrapeInterval != monv1.Duration("") {
		scrapeInterval = *instance.Spec.CustomMonitoringStack.PrometheusConfig.ScrapeInterval
	} else {
		scrapeInterval = telemetryv1.DefaultScrapeInterval
	}
	var convertedTargets []monv1alpha1.Target
	for _, t := range targets {
		convertedTargets = append(convertedTargets, monv1alpha1.Target(t))
	}
	scrapeConfig := &monv1alpha1.ScrapeConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: monv1alpha1.ScrapeConfigSpec{
			ScrapeInterval: &scrapeInterval,
			StaticConfigs: []monv1alpha1.StaticConfig{
				{
					Targets: convertedTargets,
				},
			},
		},
	}

	if tlsEnabled {
		tlsConfig := monv1.SafeTLSConfig{
			CA: monv1.SecretOrConfigMap{
				Secret: &v1.SecretKeySelector{
					Key: tls.CABundleKey,
					LocalObjectReference: v1.LocalObjectReference{
						Name: instance.Spec.PrometheusTLS.CaBundleSecretName,
					},
				},
			},
		}
		scheme := "HTTPS"
		scrapeConfig.Spec.Scheme = &scheme
		scrapeConfig.Spec.TLSConfig = &tlsConfig
		scrapeConfig.ObjectMeta.Name = fmt.Sprintf("%s-tls", instance.Name)
	}

	return scrapeConfig
}
