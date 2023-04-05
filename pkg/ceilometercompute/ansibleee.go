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

package ceilometer

import (
	ceilometerv1 "github.com/openstack-k8s-operators/ceilometer-operator/api/v1beta1"
	ansibleeev1 "github.com/openstack-k8s-operators/openstack-ansibleee-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
)

// Deployment func
func AnsibleEE(
	instance *ceilometerv1.CeilometerCompute,
	configHash string,
	labels map[string]string,
) (*ansibleeev1.OpenStackAnsibleEE, error) {

	ansibleee := &ansibleeev1.OpenStackAnsibleEE{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-ansibleee",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: ansibleeev1.OpenStackAnsibleEESpec{
			Play:      instance.Spec.Play,
			Inventory: instance.Spec.Inventory,
			Env: []corev1.EnvVar{
				{Name: "ANSIBLE_FORCE_COLOR", Value: "True"},
				{Name: "ANSIBLE_SSH_ARGS", Value: "-C -o ControlMaster=auto -o ControlPersist=80s"},
				{Name: "ANSIBLE_ENABLE_TASK_DEBUGGER", Value: "True"},
				{Name: "ANSIBLE_VERBOSITY", Value: "1"},
			},
			ExtraMounts: getExtraMounts(instance.Name),
		},
	}

	return ansibleee, nil
}
