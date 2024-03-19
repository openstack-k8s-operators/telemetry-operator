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

package availability

import (
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
)

// AodhStatefulSet func
func KSMServiceAccount(instance *telemetryv1.Ceilometer) (*corev1.ServiceAccount, *rbacv1.RoleBinding) {

	// Define a ServiceAccount
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KSMServiceName,
			Namespace: instance.Namespace,
		},
	}

	// Define a RoleBinding
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KSMServiceName,
			Namespace: instance.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "view",
			APIGroup: "rbac.authorization.k8s.io",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      KSMServiceName,
				Namespace: instance.Namespace,
			},
		},
	}

	return sa, rb
}
