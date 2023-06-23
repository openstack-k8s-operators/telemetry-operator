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

package infracompute

import (
	"fmt"
	"context"


	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	dataplanev1 "github.com/openstack-k8s-operators/dataplane-operator/api/v1beta1"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
)

// Service creates services in Kubernetes for the appropiate port in the passed node
func Service(
	instance *telemetryv1.InfraCompute,
	helper *helper.Helper,
	node *dataplanev1.OpenStackDataPlaneNode,
	port int,
	labels map[string]string,
) (*corev1.Service, controllerutil.OperationResult, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", instance.Name, node.Spec.HostName),
			Namespace: instance.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(context.TODO(), helper.GetClient(), service, func() error {
		labels["kubernetes.io/service-name"] = fmt.Sprintf("%s-%s", instance.Name, node.Spec.HostName)
		service.Labels = labels
		service.Spec.Ports = []corev1.ServicePort{{
			Protocol: "TCP",
			Port: int32(port),
			TargetPort: intstr.FromInt(port),
		}}

		err := controllerutil.SetControllerReference(instance, service, helper.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})

	return service, op, err
}

// EndpointSlice creates endpointslices in Kubernetes for the appropiate port in the passed node
func EndpointSlice(
	instance *telemetryv1.InfraCompute,
	helper *helper.Helper,
	node *dataplanev1.OpenStackDataPlaneNode,
	port int,
	labels map[string]string,
) (*discoveryv1.EndpointSlice, controllerutil.OperationResult, error) {
	endpointSlice := &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", instance.Name, node.Spec.HostName),
			Namespace: instance.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(context.TODO(), helper.GetClient(), endpointSlice, func() error {
		labels["kubernetes.io/service-name"] = fmt.Sprintf("%s-%s", instance.Name, node.Spec.HostName)
		endpointSlice.Labels = labels
		endpointSlice.AddressType = "IPv4"
		appProtocol := "http"
		protocol := corev1.ProtocolTCP
		port := int32(port)
		endpointSlice.Ports = []discoveryv1.EndpointPort{{
			AppProtocol: &appProtocol,
			Protocol: &protocol,
			Port: &port,
		}}
		endpointSlice.Endpoints = []discoveryv1.Endpoint{{
			Addresses: []string{
				node.Spec.AnsibleHost,
			},
		}}

		err := controllerutil.SetControllerReference(instance, endpointSlice, helper.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})

	return endpointSlice, op, err
}
