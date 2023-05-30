/*
Copyright 2022.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openstack-k8s-operators/lib-common/modules/common/condition"
)

// CeilometerComputeSpec defines the desired state of CeilometerCompute
type CeilometerComputeSpec struct {
	// TransportURLSecret contains the needed values to connect to RabbitMQ
	TransportURLSecret string `json:"transportURLSecret,omitempty"`

	// PasswordSelectors - Selectors to identify the service from the Secret
	// +kubebuilder:default:={service: CeilometerPassword}
	PasswordSelectors PasswordsSelector `json:"passwordSelector,omitempty"`

	// ServiceUser - optional username used for this service to register in keystone
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=ceilometer
	ServiceUser string `json:"serviceUser"`

	// Secret containing OpenStack password information for ceilometer
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=osp-secret
	Secret string `json:"secret"`

	// +kubebuilder:default:="A ceilometer compute agent"
	Description string `json:"description,omitempty"`

	// CustomServiceConfig - customize the service config using this parameter to change service defaults,
	// or overwrite rendered information using raw OpenStack config format. The content gets added to
	// to /etc/<service>/<service>.conf.d directory as custom.conf file.
	// +kubebuilder:default:="# add your customization here"
	CustomServiceConfig string `json:"customServiceConfig,omitempty"`

	// DefaultConfigOverwrite - interface to overwrite default config files like e.g. logging.conf or policy.json.
	// But can also be used to add additional files. Those get added to the service config dir in /etc/<service> .
	// TODO: -> implement
	DefaultConfigOverwrite map[string]string `json:"defaultConfigOverwrite,omitempty"`

	// InitImage is the image used for the init container
	InitImage string `json:"initImage,omitempty"`

	// ComputeImage is the image used for the ceilometer-agent-compute container
	ComputeImage string `json:"computeImage,omitempty"`

	// DataplaneSSHSecret
	// +kubebuilder:default:="dataplane-ansible-ssh-private-key-secret"
	DataplaneSSHSecret string `json:"dataplaneSSHSecret,omitempty"`

	// DataplaneInventoryConfigMap
	// +kubebuilder:default:="dataplanerole-edpm-compute"
	DataplaneInventoryConfigMap string `json:"dataplaneInventoryConfigMap,omitempty"`

	// Playbook executed
	// +kubebuilder:default:="deploy-ceilometer.yaml"
	Playbook string `json:"playbook,omitempty"`

	// ServiceAccount - service account name used internally to provide the default SA name
	ServiceAccount string `json:"serviceAccount"`
}

// CeilometerComputeStatus defines the observed state of CeilometerCompute
type CeilometerComputeStatus struct {
	// ReadyCount of ceilometercompute instances
	ReadyCount int32 `json:"readyCount,omitempty"`

	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`

	// Networks in addtion to the cluster network, the service is attached to
	Networks []string `json:"networks,omitempty"`

	// ServiceID
	ServiceID string `json:"serviceID,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CeilometerCompute is the Schema for the ceilometercomputes API
type CeilometerCompute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CeilometerComputeSpec   `json:"spec,omitempty"`
	Status CeilometerComputeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CeilometerComputeList contains a list of CeilometerCompute
type CeilometerComputeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CeilometerCompute `json:"items"`
}

// IsReady - returns true if service is ready
func (instance CeilometerCompute) IsReady() bool {
	return instance.Status.Conditions.IsTrue(condition.AnsibleEECondition)
}

func init() {
	SchemeBuilder.Register(&CeilometerCompute{}, &CeilometerComputeList{})
}
