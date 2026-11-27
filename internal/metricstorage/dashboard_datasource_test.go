/*
Copyright 2026.

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
	"testing"

	"github.com/onsi/gomega"
	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// testLabels stands in for the service labels the MetricStorage controller
// puts on every resource it manages.
var testLabels = map[string]string{"service": "metricStorage"}

func newMetricStorage(prometheusTLS tls.SimpleService) *telemetryv1.MetricStorage {
	return &telemetryv1.MetricStorage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metric-storage",
			Namespace: "openstack",
		},
		Spec: telemetryv1.MetricStorageSpec{
			MetricStorageSpecBase: telemetryv1.MetricStorageSpecBase{
				PrometheusTLS: prometheusTLS,
			},
		},
	}
}

// TestDashboardDatasourceNoTLS checks the datasource uses plain HTTP and no
// client TLS config when Prometheus TLS is disabled, and targets the
// Prometheus Service of the MetricStorage instance.
func TestDashboardDatasourceNoTLS(t *testing.T) {
	g := gomega.NewWithT(t)

	instance := newMetricStorage(tls.SimpleService{})
	instance.Name = "telemetry-kuttl"
	datasource := DashboardDatasource(instance, testLabels)

	g.Expect(datasource.Spec.Client).To(gomega.BeNil())
	proxy := datasource.Spec.Config.Plugin.Spec.(map[string]interface{})["proxy"].(map[string]interface{})
	proxySpec := proxy["spec"].(map[string]interface{})
	g.Expect(proxySpec["url"]).To(gomega.Equal("http://telemetry-kuttl-prometheus.openstack.svc:9090"))
	g.Expect(proxySpec).NotTo(gomega.HaveKey("secret"))
}

// TestDashboardDatasourceTLSCABundle checks the datasource validates the
// Prometheus certificate against the configured CA bundle secret, the same
// one used by the MonitoringStack.
func TestDashboardDatasourceTLSCABundle(t *testing.T) {
	g := gomega.NewWithT(t)

	datasource := DashboardDatasource(newMetricStorage(tls.SimpleService{
		GenericService: tls.GenericService{SecretName: ptr.To("cert-metric-storage-prometheus-svc")},
		Ca:             tls.Ca{CaBundleSecretName: "my-ca-bundle"},
	}), testLabels)

	g.Expect(datasource.Spec.Client).NotTo(gomega.BeNil())
	caCert := datasource.Spec.Client.TLS.CaCert
	g.Expect(caCert.Name).To(gomega.Equal("my-ca-bundle"))
	g.Expect(caCert.Namespace).To(gomega.Equal("openstack"))
	g.Expect(caCert.CertPath).To(gomega.Equal(tls.CABundleKey))
	proxy := datasource.Spec.Config.Plugin.Spec.(map[string]interface{})["proxy"].(map[string]interface{})
	proxySpec := proxy["spec"].(map[string]interface{})
	g.Expect(proxySpec["url"]).To(gomega.Equal("https://metric-storage-prometheus.openstack.svc:9090"))
	g.Expect(proxySpec["secret"]).To(gomega.Equal(persesDatasourceSecretName))
}

// TestDashboardDatasourceTLSServiceCertCA checks the datasource falls back to
// the CA in the Prometheus serving certificate secret when no CA bundle secret
// is configured.
func TestDashboardDatasourceTLSServiceCertCA(t *testing.T) {
	g := gomega.NewWithT(t)

	datasource := DashboardDatasource(newMetricStorage(tls.SimpleService{
		GenericService: tls.GenericService{SecretName: ptr.To("cert-metric-storage-prometheus-svc")},
	}), testLabels)

	caCert := datasource.Spec.Client.TLS.CaCert
	g.Expect(caCert.Name).To(gomega.Equal("cert-metric-storage-prometheus-svc"))
	g.Expect(caCert.CertPath).To(gomega.Equal(serviceCertCAKey))
}

// TestDashboardDatasourceLabels checks the datasource carries the service
// labels, so it can be selected like the other resources the MetricStorage
// manages, and that it does not alias the caller's map.
func TestDashboardDatasourceLabels(t *testing.T) {
	g := gomega.NewWithT(t)

	labels := map[string]string{"service": "metricStorage"}
	datasource := DashboardDatasource(newMetricStorage(tls.SimpleService{}), labels)

	g.Expect(datasource.Labels).To(gomega.Equal(labels))
	datasource.Labels["extra"] = "value"
	g.Expect(labels).NotTo(gomega.HaveKey("extra"))
}
