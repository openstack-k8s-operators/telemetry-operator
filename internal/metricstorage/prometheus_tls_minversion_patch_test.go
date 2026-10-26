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
	"encoding/json"
	"reflect"
	"testing"

	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
)

// The SSA apply must claim only minVersion; claiming
// keySecret/cert/client_ca conflicts with COO's ForceOwnership.
func TestPrometheusTLSMinVersionClaimsOnlyTLSVersions(t *testing.T) {
	secretName := "cert-metric-storage-prometheus-svc" //nolint:gosec // G101: not credentials, a cert secret name
	instance := &telemetryv1.MetricStorage{}
	instance.Name = "telemetry-kuttl-metricstorage"
	instance.Namespace = "telemetry-kuttl-tests"
	instance.Spec.PrometheusTLS.SecretName = &secretName

	raw, err := json.Marshal(PrometheusTLSMinVersion(instance).Object)
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatal(err)
	}

	spec := obj["spec"].(map[string]interface{})
	web := spec["web"].(map[string]interface{})
	tlsConfig := web["tlsConfig"].(map[string]interface{})

	want := map[string]interface{}{"minVersion": "TLS13"}
	if !reflect.DeepEqual(tlsConfig, want) {
		t.Errorf("tlsConfig = %v, want %v (must not claim cert/keySecret/client_ca)", tlsConfig, want)
	}
}
