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

// PasswordsSelector to identify the Service password from the Secret
type PasswordsSelector struct {
	// Service - Selector to get the ceilometer service password from the Secret
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=CeilometerPassword
	Service string `json:"service,omitempty"`
}

// TelemetrySpec defines the desired state of Telemetry
type TelemetrySpec struct {
	// +kubebuilder:default:="A ceilometer agent"
	Description string `json:"description,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:default=rabbitmq
	// RabbitMQ instance name
	// Needed to request a transportURL that is created and used in Telemetry
	RabbitMqClusterName string `json:"rabbitMqClusterName"`

	// +kubebuilder:validation:Required
	// CeilometerCentral - Spec definition for the CeilometerCentral service of this Telemetry deployment
	CeilometerCentral CeilometerCentralSpec `json:"ceilometerCentral"`

	// +kubebuilder:validation:Required
	// CeilometerCompute - Spec definition for the CeilometerCompute service of this Telemetry deployment
	CeilometerCompute CeilometerComputeSpec `json:"ceilometerCompute"`
}

// TelemetryStatus defines the observed state of Telemetry
type TelemetryStatus struct {
	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`

	// TransportURLSecret - Secret containing RabbitMQ transportURL
	TransportURLSecret string `json:"transportURLSecret,omitempty"`

	// ReadyCount of CeilometerCentral instance
	CeilometerCentralReadyCount int32 `json:"ceilometerCentralReadyCount,omitempty"`

	// ReadyCount of CeilometerCompute instance
	CeilometerComputeReadyCount int32 `json:"ceilometerComputeReadyCount,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Telemetry is the Schema for the telemetry API
type Telemetry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TelemetrySpec   `json:"spec,omitempty"`
	Status TelemetryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TelemetryList contains a list of Telemetry
type TelemetryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Telemetry `json:"items"`
}

// IsReady - returns true if service is ready
func (instance Telemetry) IsReady() bool {
	return instance.Status.Conditions.IsTrue(condition.DeploymentReadyCondition)
}

func init() {
	SchemeBuilder.Register(&Telemetry{}, &TelemetryList{})
}
