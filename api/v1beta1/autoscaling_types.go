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
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
)

// Prometheus defines which prometheus to use for Autoscaling
type Prometheus struct {
	// Enables the deployment of autoscaling prometheus
	// +kubebuilder:default=true
	DeployPrometheus bool `json:"deployPrometheus,omitempty"`

	// Host of user deployed prometheus if deployPrometheus is set to false
	Host string `json:"host,omitempty"`

	// Port of user deployed prometheus if deployPrometheus is set to false
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port,omitempty"`
}

// AutoscalingSpec defines the desired state of Autoscaling
type AutoscalingSpec struct {
	// Specification of which prometheus to use for autoscaling
	Prometheus Prometheus `json:"prometheus,omitempty"`

	// Allows enabling and disabling the autoscaling feature
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`
}

// AutoscalingStatus defines the observed state of Autoscaling
type AutoscalingStatus struct {
	// ReadyCount of autoscaling instances
	ReadyCount int32 `json:"readyCount,omitempty"`

	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Autoscaling is the Schema for the autoscalings API
type Autoscaling struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutoscalingSpec   `json:"spec,omitempty"`
	Status AutoscalingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AutoscalingList contains a list of Autoscaling
type AutoscalingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Autoscaling `json:"items"`
}

// IsReady - returns true if CeilometerCentral is reconciled successfully
func (instance Autoscaling) IsReady() bool {
	return instance.Status.Conditions.IsTrue(condition.ReadyCondition)
}


func init() {
	SchemeBuilder.Register(&Autoscaling{}, &AutoscalingList{})
}

// RbacConditionsSet - set the conditions for the rbac object
func (instance Autoscaling) RbacConditionsSet(c *condition.Condition) {
	instance.Status.Conditions.Set(c)
}

// RbacNamespace - return the namespace
func (instance Autoscaling) RbacNamespace() string {
	return instance.Namespace
}

// RbacResourceName - return the name to be used for rbac objects (serviceaccount, role, rolebinding)
func (instance Autoscaling) RbacResourceName() string {
	return "telemetry-" + instance.Name
}
