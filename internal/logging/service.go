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
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1 "k8s.io/api/core/v1"
)

// Service creates a LoadBalancer service for openshift-logging.
// The RHOSO deployment namespace is included in the service name so that
// multiple RHOSO deployments living in different namespaces of the same
// OpenShift cluster can each create their own service in the shared
// CLONamespace without colliding.
func Service(
	instance *telemetryv1.Logging,
	helper *helper.Helper,
	labels map[string]string,
) (*corev1.Service, controllerutil.OperationResult, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "openstack-" + ServiceName + "-" + instance.Namespace,
			Namespace: instance.Spec.CLONamespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(context.TODO(), helper.GetClient(), service, func() error {
		//service.Labels = labels
		service.Spec.Ports = []corev1.ServicePort{{
			Protocol:   corev1.Protocol("TCP"),
			Port:       instance.Spec.Port,
			TargetPort: intstr.FromInt(instance.Spec.TargetPort),
		}}
		service.Spec.Selector = map[string]string{
			"app.kubernetes.io/component": "collector",
			"app.kubernetes.io/name":      "vector",
			"app.kubernetes.io/part-of":   "cluster-logging",
		}
		service.Annotations = instance.Spec.Annotations
		service.Labels = labels
		service.Spec.Type = "LoadBalancer"

		return nil
	})

	return service, op, err
}

// DeleteLegacyService removes the pre-upgrade, namespace-agnostic service
// ("openstack-logging") that was created before the RHOSO namespace was
// included in the service name. On upgrade this prevents an orphaned Service
// from being left behind in the shared CLONamespace. It is idempotent: a
// missing service is not treated as an error, so it is safe to call on every
// reconcile and when multiple deployments share a CLONamespace.
func DeleteLegacyService(
	ctx context.Context,
	instance *telemetryv1.Logging,
	helper *helper.Helper,
) error {
	legacyService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "openstack-" + ServiceName,
			Namespace: instance.Spec.CLONamespace,
		},
	}

	err := helper.GetClient().Delete(ctx, legacyService)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return err
	}

	return nil
}
