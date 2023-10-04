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

package autoscaling

import (
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	monv1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1"
	obov1 "github.com/rhobs/observability-operator/pkg/apis/monitoring/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Prometheus defines a monitoringstack with prometheus for autoscaling
func Prometheus(
	instance *telemetryv1.Autoscaling,
	labels map[string]string,
) *obov1.MonitoringStack {
	prometheus := &obov1.MonitoringStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-prometheus",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: obov1.MonitoringStackSpec{
			AlertmanagerConfig: obov1.AlertmanagerConfig{
				Disabled: true,
			},
			PrometheusConfig: &obov1.PrometheusConfig{
				Replicas: &PrometheusReplicas,
			},
			Retention: monv1.Duration(PrometheusRetention),
			LogLevel:  "debug",
			ResourceSelector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
		},
	}
	return prometheus
}
