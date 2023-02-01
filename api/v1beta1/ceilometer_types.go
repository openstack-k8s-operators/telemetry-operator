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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
)

// RabbitMQSelector to identify the DB and AdminUser password from the Secret
type RabbitMQSelector struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="host"
	// Host - Selector to get the host of the RabbitMQ connection
	Host string `json:"host,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="username"
	// Username - Selector to get the username of the RabbitMQ connection
	Username string `json:"username,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="password"
	// Password - Selector to get the password of the RabbitMQ connection
	Password string `json:"password,omitempty"`
}

// CeilometerSpec defines the desired state of Ceilometer
type CeilometerSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=rabbitmq-default-user
	// The needed values to connect to RabbitMQ
	RabbitMQSecret string `json:"rabbitMQSecret,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={username: username, password: password}
	// RabbitMQSelectors - Selectors to identify host, username and password from the Secret
	RabbitMQSelectors RabbitMQSelector `json:"rabbitMQSelector,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="# add your customization here"
	// CustomServiceConfig - customize the service config using this parameter to change service defaults,
	// or overwrite rendered information using raw OpenStack config format. The content gets added to
	// to /etc/<service>/<service>.conf.d directory as custom.conf file.
	CustomServiceConfig string `json:"customServiceConfig,omitempty"`

	// +kubebuilder:validation:Optional
	// ConfigOverwrite - interface to overwrite default config files like e.g. logging.conf or policy.json.
	// But can also be used to add additional files. Those get added to the service config dir in /etc/<service> .
	// TODO: -> implement
	DefaultConfigOverwrite map[string]string `json:"defaultConfigOverwrite,omitempty"`

	// +kubebuilder:validation:Optional
	// NetworkAttachmentDefinitions list of network attachment definitions the service pod gets attached to
	NetworkAttachmentDefinitions []string `json:"networkAttachmentDefinitions,omitempty"`
}

// CeilometerStatus defines the observed state of Ceilometer
type CeilometerStatus struct {
	// ReadyCount of keystone API instances
	ReadyCount int32 `json:"readyCount,omitempty"`

	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`

	// Networks in addtion to the cluster network, the service is attached to
	Networks []string `json:"networks,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Ceilometer is the Schema for the ceilometers API
type Ceilometer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CeilometerSpec   `json:"spec,omitempty"`
	Status CeilometerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CeilometerList contains a list of Ceilometer
type CeilometerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ceilometer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ceilometer{}, &CeilometerList{})
}
