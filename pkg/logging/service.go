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

package logging

import (
	"context"

	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1 "k8s.io/api/core/v1"
)

// Service creates a LoadBalancer service for openshift-logging
func Service(
	instance *telemetryv1.Logging,
	helper *helper.Helper,
	labels map[string]string,
) (*corev1.Service, controllerutil.OperationResult, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "openstack-" + ServiceName,
			Namespace: instance.Spec.CLONamespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(context.TODO(), helper.GetClient(), service, func() error {
		//service.Labels = labels
		service.Spec.Ports = []corev1.ServicePort{{
			Protocol:   corev1.Protocol(instance.Spec.Protocol),
			Port:       instance.Spec.Port,
			TargetPort: intstr.FromInt(instance.Spec.TargetPort),
		}}
		service.Spec.Selector = map[string]string{
			"app.kubernetes.io/instance": "collector",
			"component":                  "collector",
			"provider":                   "openshift",
		}
		service.Annotations = map[string]string{
			"endpoint":                            "internal",
			"metallb.universe.tf/address-pool":    instance.Spec.Network,
			"metallb.universe.tf/allow-shared-ip": instance.Spec.Network,
			"metallb.universe.tf/loadBalancerIPs": instance.Spec.IPAddr,
		}
		service.Spec.Type = "LoadBalancer"

		return nil
	})

	return service, op, err
}
