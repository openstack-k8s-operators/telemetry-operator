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
	"time"

	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

// AetosService returns a lib-common Service that exposes the Aetos sidecar
// running inside the Prometheus pod. For the public endpoint an ingress/route
// is requested so the service is reachable from outside the cluster, while the
// internal endpoint stays cluster-local.
func AetosService(
	instance *telemetryv1.MetricStorage,
	endpointType service.Endpoint,
	serviceLabels map[string]string,
	svcOverride service.RoutedOverrideSpec,
) (*service.Service, error) {
	endpointTypeStr := string(endpointType)
	svcName := fmt.Sprintf("%s-%s", AetosServiceName, endpointTypeStr)

	// The Aetos sidecar runs inside the Prometheus pod, so the Service must
	// select that pod rather than the MetricStorage controller's own labels.
	selector := map[string]string{
		"app.kubernetes.io/component": "prometheus",
		"app.kubernetes.io/part-of":   instance.Name,
	}

	exportLabels := util.MergeStringMaps(
		serviceLabels,
		map[string]string{
			service.AnnotationEndpointKey: endpointTypeStr,
		},
	)

	svc, err := service.NewService(
		service.GenericService(&service.GenericServiceDetails{
			Name:      svcName,
			Namespace: instance.Namespace,
			Labels:    exportLabels,
			Selector:  selector,
			Port: service.GenericServicePort{
				Name:     endpointTypeStr,
				Port:     AetosPort,
				Protocol: corev1.ProtocolTCP,
			},
		}),
		time.Duration(5)*time.Second,
		&svcOverride.OverrideSpec,
	)
	if err != nil {
		return nil, err
	}

	svc.AddAnnotation(map[string]string{
		service.AnnotationEndpointKey: endpointTypeStr,
	})

	if endpointType == service.EndpointPublic && svc.GetServiceType() == corev1.ServiceTypeClusterIP {
		svc.AddAnnotation(map[string]string{
			service.AnnotationIngressCreateKey: "true",
		})
	} else {
		svc.AddAnnotation(map[string]string{
			service.AnnotationIngressCreateKey: "false",
		})
		// The public endpoint's DNS is user-managed (its hostname is the
		// endpointURL host, not the in-cluster Service name), so skip
		// registering it in dnsmasq even when it is a LoadBalancer.
		if svc.GetServiceType() == corev1.ServiceTypeLoadBalancer && endpointType != service.EndpointPublic {
			svc.AddAnnotation(map[string]string{
				service.AnnotationHostnameKey: svc.GetServiceHostname(),
			})
		}
	}

	return svc, nil
}
