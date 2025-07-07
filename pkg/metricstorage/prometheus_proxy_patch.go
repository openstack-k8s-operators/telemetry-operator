package metricstorage

import (
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func PrometheusProxy(instance *telemetryv1.MetricStorage) *unstructured.Unstructured {
	tlsEnabled := instance.Spec.PrometheusTLS.Enabled()

	args := []interface{}{
		"--secure-listen-address=0.0.0.0:8443",
		"--logtostderr=true",
		"--v=10",
	}

	if tlsEnabled {
		args = append(args,
			"--upstream=https://metric-storage-prometheus.openstack.svc:9090/",
			"--upstream-ca-file=/etc/tls/internal-ca-bundle.pem",
		)
	} else {
		args = append(args, "--upstream=http://metric-storage-prometheus.openstack.svc:9090/")
	}

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
						"ports": []interface{}{
							map[string]interface{}{
								"name":          "https",
								"containerPort": 8443,
							},
						},
					},
				},
			},
		},
	}

	spec := prom.Object["spec"].(map[string]interface{})
	containers := spec["containers"].([]interface{})
	container := containers[0].(map[string]interface{})
	container["args"] = args

	if tlsEnabled {
		container["volumeMounts"] = []interface{}{
			map[string]interface{}{
				"name":      "ca-cert",
				"mountPath": "/etc/tls",
				"readOnly":  true,
			},
		}

		spec["volumes"] = []interface{}{
			map[string]interface{}{
				"name": "ca-cert",
				"secret": map[string]interface{}{
					"secretName": "combined-ca-bundle",
					"items": []interface{}{
						map[string]interface{}{
							"key":  "internal-ca-bundle.pem",
							"path": "internal-ca-bundle.pem",
						},
					},
				},
			},
		}
	}

	return prom
}
