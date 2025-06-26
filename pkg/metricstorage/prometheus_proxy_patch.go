package metricstorage

import (
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func PrometheusProxy(instance *telemetryv1.MetricStorage) *unstructured.Unstructured {
	prom := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "monitoring.rhobs/v1",
			"kind":       "Prometheus",
			"metadata": map[string]interface{}{
				"name":      instance.Name,
				"namespace": instance.Namespace,
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "kube-rbac-proxy",
						"image": "quay.io/openstack-k8s-operators/kube-rbac-proxy:v0.16.0",
						"args": []interface{}{
							"--secure-listen-address=0.0.0.0:8443",
							"--upstream=https://metric-storage-prometheus.openstack.svc:9090/",
							"--upstream-ca-file=/etc/tls/ca.crt",
							"--logtostderr=true",
							"--v=10",
						},
						"ports": []interface{}{
							map[string]interface{}{
								"name":          "https",
								"containerPort": 8443,
							},
						},
						"volumeMounts": []interface{}{
							map[string]interface{}{
								"name":      "ca-cert",
								"mountPath": "/etc/tls",
								"readOnly":  true,
							},
						},
					},
				},
				"volumes": []interface{}{
					map[string]interface{}{
						"name": "ca-cert",
						"secret": map[string]interface{}{
							"secretName": "cert-metric-storage-sample-prometheus-svc",
							"items": []interface{}{
								map[string]interface{}{
									"key":  "ca.crt",
									"path": "ca.crt",
								},
							},
						},
					},
				},
			},
		},
	}
	return prom
}
