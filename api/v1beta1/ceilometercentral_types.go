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
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PasswordsSelector to identify the Service password from the Secret
type PasswordsSelector struct {
	// Service - Selector to get the ceilometer service password from the Secret
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=CeilometerPassword
	Service string `json:"service,omitempty"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CeilometerCentralSpec defines the desired state of CeilometerCentral
type CeilometerCentralSpec struct {
	// The needed values to connect to RabbitMQ
	TransportURLSecret string `json:"transportURLSecret,omitempty"`

	// PasswordSelectors - Selectors to identify the service from the Secret
	// +kubebuilder:default:={service: CeilometerPassword}
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

	// +kubebuilder:default:="A ceilometer agent"
	Description string `json:"description,omitempty"`
}

// CeilometerCentralStatus defines the observed state of CeilometerCentral
type CeilometerCentralStatus struct {
	// ReadyCount of ceilometercentral instances
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

// CeilometerCentral is the Schema for the ceilometercentrals API
type CeilometerCentral struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CeilometerCentralSpec   `json:"spec,omitempty"`
	Status CeilometerCentralStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CeilometerCentralList contains a list of CeilometerCentral
type CeilometerCentralList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CeilometerCentral `json:"items"`
}

// IsReady - returns true if service is ready
func (instance CeilometerCentral) IsReady() bool {
	return instance.Status.Conditions.IsTrue(condition.DeploymentReadyCondition)
}

func init() {
	SchemeBuilder.Register(&CeilometerCentral{}, &CeilometerCentralList{})
}
