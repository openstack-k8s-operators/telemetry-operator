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
	"github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CloudKittyAPITemplateCore defines the input parameters for the CloudKitty API service
type CloudKittyAPITemplateCore struct {
	// Common input parameters for the CloudKitty API service
	CloudKittyServiceTemplate `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	// Replicas - CloudKitty API Replicas
	Replicas *int32 `json:"replicas"`

	// +kubebuilder:validation:Optional
	// Override, provides the ability to override the generated manifest of several child resources.
	Override APIOverrideSpec `json:"override,omitempty"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// TLS - Parameters related to the TLS
	TLS tls.API `json:"tls,omitempty"`
}

// CloudKittyAPITemplate defines the input parameters for the CloudKitty API service
type CloudKittyAPITemplate struct {
	// +kubebuilder:validation:Required
	// ContainerImage - CloudKitty Container Image URL (will be set to environmental default if empty)
	ContainerImage string `json:"containerImage"`

	CloudKittyAPITemplateCore `json:",inline"`
}

// CloudKittyAPISpec defines the desired state of CloudKittyAPI
type CloudKittyAPISpec struct {
	// Common input parameters for all CloudKitty services
	CloudKittyTemplate `json:",inline"`

	// Input parameters for the CloudKitty API service
	CloudKittyAPITemplate `json:",inline"`

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

// CloudKittyAPIStatus defines the observed state of CloudKittyAPI
type CloudKittyAPIStatus struct {
	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// API endpoints
	APIEndpoints map[string]map[string]string `json:"apiEndpoints,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`

	// ReadyCount of CloudKitty API instances
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=0
	ReadyCount int32 `json:"readyCount"`

	// ServiceIDs
	ServiceIDs map[string]string `json:"serviceIDs,omitempty"`

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

// CloudKittyAPI is the Schema for the cloudkittyapis API
type CloudKittyAPI struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudKittyAPISpec   `json:"spec,omitempty"`
	Status CloudKittyAPIStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CloudKittyAPIList contains a list of CloudKittyAPI
type CloudKittyAPIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudKittyAPI `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudKittyAPI{}, &CloudKittyAPIList{})
}

// IsReady - returns true if service is ready to serve requests
func (instance CloudKittyAPI) IsReady() bool {
	return instance.Generation == instance.Status.ObservedGeneration &&
		instance.Status.ReadyCount == *instance.Spec.Replicas &&
		(instance.Status.Conditions.IsTrue(condition.DeploymentReadyCondition) ||
			(instance.Status.Conditions.IsFalse(condition.DeploymentReadyCondition) && *instance.Spec.Replicas == 0))
}

// GetSpecTopologyRef - Returns the LastAppliedTopology Set in the Status
func (instance *CloudKittyAPI) GetSpecTopologyRef() *topologyv1.TopoRef {
	return instance.Spec.TopologyRef
}

// GetLastAppliedTopology - Returns the LastAppliedTopology Set in the Status
func (instance *CloudKittyAPI) GetLastAppliedTopology() *topologyv1.TopoRef {
	return instance.Status.LastAppliedTopology
}

// SetLastAppliedTopology - Sets the LastAppliedTopology value in the Status
func (instance *CloudKittyAPI) SetLastAppliedTopology(topologyRef *topologyv1.TopoRef) {
	instance.Status.LastAppliedTopology = topologyRef
}
