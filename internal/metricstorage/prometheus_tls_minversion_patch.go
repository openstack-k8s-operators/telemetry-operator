/*
Copyright 2026 Red Hat, Inc.

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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// PrometheusTLSMinVersion returns an SSA patch for the Prometheus CR that sets
// (or releases) web.tlsConfig.minVersion. Built as unstructured so the apply
// claims only this leaf: monv1.WebTLSConfig serializes
// keySecret/cert/client_ca even when unset (omitempty is a no-op on structs),
// which conflicts with COO's ForceOwnership of those fields. COO never sets
// minVersion, so it survives under FieldOwner "telemetry-operator".
func PrometheusTLSMinVersion(
	instance *telemetryv1.MetricStorage,
) unstructured.Unstructured {
	prom := unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "monitoring.rhobs/v1",
		"kind":       "Prometheus",
		"metadata": map[string]interface{}{
			"name":      instance.Name,
			"namespace": instance.Namespace,
		},
	}}
	if instance.Spec.PrometheusTLS.Enabled() {
		prom.Object["spec"] = map[string]interface{}{
			"web": map[string]interface{}{
				"tlsConfig": map[string]interface{}{
					"minVersion": string(monv1.TLSVersion13),
				},
			},
		}
	}
	return prom
}
