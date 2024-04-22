/*
Copyright 2024.

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

	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DashboardDatasourceData(ctx context.Context, c client.Client, instance *telemetryv1.MetricStorage, datasourceName string, namespace string) (map[string]string, error) {

	scheme := "http"
	certText := ""
	if instance.Spec.PrometheusTLS.Enabled() {
		scheme = "https"
		namespacedName := types.NamespacedName{
			Name:      *instance.Spec.PrometheusTLS.SecretName,
			Namespace: instance.Namespace,
		}
		caSecret := &corev1.Secret{}
		err := c.Get(ctx, namespacedName, caSecret)
		if err != nil {
			return nil, err
		}
		certText = string(caSecret.Data["ca.crt"])
	}

	return map[string]string{
		// WARNING: The value lines below MUST be indented with spaces, not tabs (because they are YAML strings)
		"dashboard-datasource-ca": certText,
		"dashboard-datasource.yaml": `
kind: "Datasource"
metadata:
    name: "` + datasourceName + `"
    project: "` + namespace + `"
spec:
    plugin:
        kind: "PrometheusDatasource"
        spec:
            direct_url: "` + scheme + `://metric-storage-prometheus.` + instance.Namespace + `.svc.cluster.local:9090"
`}, nil
}
