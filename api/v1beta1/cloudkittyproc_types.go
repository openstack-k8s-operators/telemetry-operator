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
	topologyv1 "github.com/openstack-k8s-operators/infra-operator/apis/topology/v1beta1"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CloudKittyProcTemplateCore defines the input parameters for the CloudKitty Processor service
type CloudKittyProcTemplateCore struct {
	// Common input parameters for the CloudKitty Processor service
	CloudKittyServiceTemplate `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	// Replicas - CloudKitty API Replicas
	Replicas *int32 `json:"replicas"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// TLS - Parameters related to the TLS
	TLS tls.SimpleService `json:"tls,omitempty"`
}

// CloudKittyProcTemplate defines the input parameters for the CloudKitty Processor service
type CloudKittyProcTemplate struct {
	// +kubebuilder:validation:Required
	// ContainerImage - CloudKitty Container Image URL (will be set to environmental default if empty)
	ContainerImage string `json:"containerImage"`

	CloudKittyProcTemplateCore `json:",inline"`
}

// CloudKittyProcSpec defines the desired state of CloudKitty Processor
type CloudKittyProcSpec struct {
	// Common input parameters for all CloudKitty services
	CloudKittyTemplate `json:",inline"`

	// Input parameters for the CloudKitty Processor service
	CloudKittyProcTemplate `json:",inline"`

	// +kubebuilder:validation:Required
	// DatabaseHostname - CloudKitty Database Hostname
	DatabaseHostname string `json:"databaseHostname"`

	// +kubebuilder:validation:Required
	// Secret containing RabbitMq transport URL
	TransportURLSecret string `json:"transportURLSecret"`

	// +kubebuilder:validation:Required
	// ServiceAccount - service account name used internally to provide CloudKitty services the default SA name
	ServiceAccount string `json:"serviceAccount"`
}

// CloudKittyProcStatus defines the observed state of CloudKitty Processor
type CloudKittyProcStatus struct {
	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`

	// ReadyCount of CloudKitty Processor instances
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=0
	ReadyCount int32 `json:"readyCount"`

	// NetworkAttachments status of the deployment pods
	NetworkAttachments map[string][]string `json:"networkAttachments,omitempty"`

	// ObservedGeneration - the most recent generation observed for this service.
	// If the observed generation is different than the spec generation, then the
	// controller has not started processing the latest changes, and the status
	// and its conditions are likely stale.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// LastAppliedTopology - the last applied Topology
	LastAppliedTopology *topologyv1.TopoRef `json:"lastAppliedTopology,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="NetworkAttachments",type="string",JSONPath=".status.networkAttachments",description="NetworkAttachments"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[0].status",description="Status"
//+kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.conditions[0].message",description="Message"

// CloudKittyProc is the Schema for the cloudkittprocs API
type CloudKittyProc struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudKittyProcSpec   `json:"spec,omitempty"`
	Status CloudKittyProcStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CloudKittyProcList contains a list of CloudKittyProc
type CloudKittyProcList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudKittyProc `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudKittyProc{}, &CloudKittyProcList{})
}

// IsReady - returns true if service is ready to serve requests
func (instance CloudKittyProc) IsReady() bool {
	return instance.Generation == instance.Status.ObservedGeneration &&
		instance.Status.ReadyCount == *instance.Spec.Replicas &&
		(instance.Status.Conditions.IsTrue(condition.DeploymentReadyCondition) ||
			(instance.Status.Conditions.IsFalse(condition.DeploymentReadyCondition) && *instance.Spec.Replicas == 0))
}

// GetSpecTopologyRef - Returns the LastAppliedTopology Set in the Status
func (instance *CloudKittyProc) GetSpecTopologyRef() *topologyv1.TopoRef {
	return instance.Spec.TopologyRef
}

// GetLastAppliedTopology - Returns the LastAppliedTopology Set in the Status
func (instance *CloudKittyProc) GetLastAppliedTopology() *topologyv1.TopoRef {
	return instance.Status.LastAppliedTopology
}

// SetLastAppliedTopology - Sets the LastAppliedTopology value in the Status
func (instance *CloudKittyProc) SetLastAppliedTopology(topologyRef *topologyv1.TopoRef) {
	instance.Status.LastAppliedTopology = topologyRef
}
