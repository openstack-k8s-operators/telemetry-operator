/*
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

package availability

import (
	"context"

	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	svc "github.com/openstack-k8s-operators/lib-common/modules/common/service"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1 "k8s.io/api/core/v1"
)

// KSMService exposes kube-state-metrics pod. Appropriate annotation for TLS enablement is added
func KSMService(
	instance *telemetryv1.Ceilometer,
	helper *helper.Helper,
	labels map[string]string,
) (*corev1.Service, controllerutil.OperationResult, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KSMServiceName,
			Namespace: instance.Namespace,
			Annotations: map[string]string{
				svc.AnnotationEndpointKey: string(svc.EndpointInternal),
			},
		},
	}

	op, err := controllerutil.CreateOrUpdate(context.TODO(), helper.GetClient(), service, func() error {
		service.Labels = labels
		service.Spec.Selector = labels
		service.Spec.Ports = []corev1.ServicePort{
			{
				Port: KSMMetricsPort,
				Name: "http-metrics",
			},
			{
				Port: KSMHealthPort,
				Name: "telemetry",
			},
		}

		return controllerutil.SetControllerReference(instance, service, helper.GetScheme())
	})

	return service, op, err
}
