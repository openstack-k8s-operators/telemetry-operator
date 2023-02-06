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

// PasswordsSelector to identify the DB and AdminUser password from the Secret
type PasswordsSelector struct {
	// Host - Selector to get the host of the RabbitMQ connection
	// +kubebuilder:default:="host"
	Host string `json:"host,omitempty"`
	// Username - Selector to get the username of the RabbitMQ connection
	// +kubebuilder:default:="username"
	Username string `json:"username,omitempty"`
	// Password - Selector to get the password of the RabbitMQ connection
	// +kubebuilder:default:="password"
	Password string `json:"password,omitempty"`
	// Service - Selector to get the ceilometer service password from the Secret
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=CeilometerPassword
	Service string `json:"service,omitempty"`
}

// CeilometerSpec defines the desired state of Ceilometer
type CeilometerSpec struct {
	// The needed values to connect to RabbitMQ
	// +kubebuilder:default:=rabbitmq-default-user
	RabbitMQSecret string `json:"rabbitMQSecret,omitempty"`

	// PasswordSelectors - Selectors to identify host, username and password from the Secret
	// +kubebuilder:default:={username: username, password: password}
	PasswordSelectors PasswordsSelector `json:"passwordSelector,omitempty"`

	// ServiceUser - optional username used for this service to register in keystone
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=ceilometer
	ServiceUser string `json:"serviceUser,omitempty"`

	// Secret containing OpenStack password information for ceilometer
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=osp-secret
	Secret string `json:"secret,omitempty"`

	// CustomServiceConfig - customize the service config using this parameter to change service defaults,
	// or overwrite rendered information using raw OpenStack config format. The content gets added to
	// to /etc/<service>/<service>.conf.d directory as custom.conf file.
	// +kubebuilder:default:="# add your customization here"
	CustomServiceConfig string `json:"customServiceConfig,omitempty"`

	// ConfigOverwrite - interface to overwrite default config files like e.g. logging.conf or policy.json.
	// But can also be used to add additional files. Those get added to the service config dir in /etc/<service> .
	// TODO: -> implement
	DefaultConfigOverwrite map[string]string `json:"defaultConfigOverwrite,omitempty"`

	// NetworkAttachmentDefinitions list of network attachment definitions the service pod gets attached to
	NetworkAttachmentDefinitions []string `json:"networkAttachmentDefinitions,omitempty"`

	// +kubebuilder:default:="quay.io/tripleomastercentos9/openstack-ceilometer-central:current-tripleo"
	CentralImage string `json:"centralImage,omitempty"`

	// +kubebuilder:default:="quay.io/tripleomastercentos9/openstack-ceilometer-notification:current-tripleo"
	NotificationImage string `json:"notificationImage,omitempty"`

	// +kubebuilder:default:="quay.io/infrawatch/sg-core:latest"
	SgCoreImage string `json:"sgCoreImage,omitempty"`

	// +kubebuilder:default:="quay.io/tripleomastercentos9/openstack-ceilometer-central:current-tripleo"
	InitImage string `json:"initImage,omitempty"`
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

	// ServiceID
	ServiceID string `json:"serviceID,omitempty"`
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
