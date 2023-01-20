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
)

// CeilometerCommonSpec defines common fields for Ceilometer agents
type CeilometerCommonSpec struct {
}

// GetCommonFields allows to return all promoted fields as single structure
func (ccs CeilometerCommonSpec) GetCommonFields() *CeilometerCommonSpec {
	return &ccs
}

// CeilometerCentralAgentSpec defines the desired state of CeilometerCentralAgent
type CeilometerCentralAgentSpec struct {
	CeilometerCommonSpec `json:",inline"`

	// Foo is an example field of CeilometerCentralAgent. Edit ceilometercentralagent_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// CeilometerCentralAgentStatus defines the observed state of CeilometerCentralAgent
type CeilometerCentralAgentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CeilometerCentralAgent is the Schema for the ceilometercentralagents API
type CeilometerCentralAgent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CeilometerCentralAgentSpec   `json:"spec,omitempty"`
	Status CeilometerCentralAgentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CeilometerCentralAgentList contains a list of CeilometerCentralAgent
type CeilometerCentralAgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CeilometerCentralAgent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CeilometerCentralAgent{}, &CeilometerCentralAgentList{})
	SchemeBuilder.Register(&CeilometerNotificationAgent{}, &CeilometerNotificationAgentList{})
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CeilometerNotificationAgentSpec defines the desired state of CeilometerNotificationAgent
type CeilometerNotificationAgentSpec struct {
	CeilometerCommonSpec `json:",inline"`

	// Foo is an example field of CeilometerNotificationAgent. Edit ceilometernotificationagent_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// CeilometerNotificationAgentStatus defines the observed state of CeilometerNotificationAgent
type CeilometerNotificationAgentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CeilometerNotificationAgent is the Schema for the ceilometernotificationagents API
type CeilometerNotificationAgent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CeilometerNotificationAgentSpec   `json:"spec,omitempty"`
	Status CeilometerNotificationAgentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CeilometerNotificationAgentList contains a list of CeilometerNotificationAgent
type CeilometerNotificationAgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CeilometerNotificationAgent `json:"items"`
}
