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
	"fmt"

	endpoint "github.com/openstack-k8s-operators/lib-common/modules/common/endpoint"
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AetosService returns a K8s Service that exposes the Aetos sidecar port on the Prometheus pod
func AetosService(
	instance *telemetryv1.MetricStorage,
	endpointType endpoint.Endpoint,
) corev1.Service {
	endpointTypeStr := string(endpointType)
	svcName := fmt.Sprintf("%s-%s", AetosServiceName, endpointTypeStr)

	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: instance.Namespace,
			Annotations: map[string]string{
				service.AnnotationEndpointKey:      endpointTypeStr,
				service.AnnotationIngressCreateKey: "false",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app.kubernetes.io/component": "prometheus",
				"app.kubernetes.io/part-of":   instance.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:     AetosServiceName,
					Port:     int32(AetosPort),
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}
	return svc
}
