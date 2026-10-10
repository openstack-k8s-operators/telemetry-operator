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
	"fmt"

	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	persesv1alpha1 "github.com/perses/perses-operator/api/v1alpha1"
	persesv1 "github.com/perses/perses/pkg/model/api/v1"
	persescommon "github.com/perses/perses/pkg/model/api/v1/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// PrometheusDatasourceName is the name of the PersesDatasource the operator
	// manages and that the managed PersesDashboards reference.
	PrometheusDatasourceName = "prometheus"

	// prometheusDatasourceDisplayName is the human readable name shown in the
	// Perses UI for the managed Prometheus datasource.
	prometheusDatasourceDisplayName = "Prometheus Datasource"

	// combinedCABundleSecretName is the secret holding the cluster CA bundle
	// used to validate the Prometheus TLS certificate.
	combinedCABundleSecretName = "combined-ca-bundle" //nolint:gosec // G101: secret name, not an actual credential

	// combinedCABundleKey is the key inside combinedCABundleSecretName that
	// holds the CA bundle in PEM format.
	combinedCABundleKey = "tls-ca-bundle.pem"

	// persesDatasourceSecretName is the Perses Secret the perses-operator
	// generates from the datasource's client.tls config. The perses-operator
	// names it "<datasource>-secret" (its SecretNameSuffix) but does NOT wire it
	// into the datasource proxy automatically, so the proxy must reference it by
	// name for Perses to use the CA when querying Prometheus over TLS.
	persesDatasourceSecretName = PrometheusDatasourceName + "-secret"
)

// DashboardDatasource builds the PersesDatasource object that points the
// managed dashboards at the MetricStorage Prometheus instance. The datasource
// lives in the same namespace as the MetricStorage CR so it can be owned by it.
func DashboardDatasource(instance *telemetryv1.MetricStorage) *persesv1alpha1.PersesDatasource {
	scheme := "http"

	datasource := &persesv1alpha1.PersesDatasource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: persesv1alpha1.GroupVersion.String(),
			Kind:       "PersesDatasource",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      PrometheusDatasourceName,
			Namespace: instance.Namespace,
		},
	}

	if instance.Spec.PrometheusTLS.Enabled() {
		scheme = "https"
		datasource.Spec.Client = &persesv1alpha1.Client{
			TLS: &persesv1alpha1.TLS{
				Enable: true,
				CaCert: &persesv1alpha1.Certificate{
					SecretSource: persesv1alpha1.SecretSource{
						Type:      persesv1alpha1.SecretSourceTypeSecret,
						Name:      combinedCABundleSecretName,
						Namespace: instance.Namespace,
					},
					CertPath: combinedCABundleKey,
				},
			},
		}
	}

	url := fmt.Sprintf("%s://metric-storage-prometheus.%s.svc:9090", scheme, instance.Namespace)

	proxySpec := map[string]interface{}{
		"url": url,
	}
	if instance.Spec.PrometheusTLS.Enabled() {
		// Reference the Perses Secret the perses-operator creates from the
		// client.tls config above, so the Perses proxy verifies the Prometheus
		// serving certificate instead of failing with "certificate signed by
		// unknown authority".
		proxySpec["secret"] = persesDatasourceSecretName
	}

	datasource.Spec.Config = persesv1alpha1.Datasource{
		DatasourceSpec: persesv1.DatasourceSpec{
			Default: true,
			Display: &persescommon.Display{
				Name: prometheusDatasourceDisplayName,
			},
			Plugin: persescommon.Plugin{
				Kind: "PrometheusDatasource",
				Spec: map[string]interface{}{
					"proxy": map[string]interface{}{
						"kind": "HTTPProxy",
						"spec": proxySpec,
					},
				},
			},
		},
	}

	return datasource
}
