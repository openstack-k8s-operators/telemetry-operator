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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InfraComputeSpec defines the desired state of InfraCompute
type InfraComputeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:default:="An infra compute agent"
	Description string `json:"description,omitempty"`

	// NodeExporterImage is the image used for the node_exporter container
	NodeExporterImage string `json:"nodeExporterImage,omitempty"`

	// DataplaneSSHSecret
	// +kubebuilder:default:="dataplane-ansible-ssh-private-key-secret"
	DataplaneSSHSecret string `json:"dataplaneSSHSecret,omitempty"`

	// DataplaneInventoryConfigMap
	// +kubebuilder:default:="dataplanerole-edpm-compute"
	DataplaneInventoryConfigMap string `json:"dataplaneInventoryConfigMap,omitempty"`

	// Playbook executed
	// +kubebuilder:default:="deploy-infra.yaml"
	Playbook string `json:"playbook,omitempty"`

	// The extravars ConfigMap to pass to ansible execution
	// +kubebuilder:default:="telemetry-infracompute-extravars"
	ExtravarsConfigMap string `json:"extravarsConfigMap,omitempty"`
}

// InfraComputeStatus defines the observed state of InfraCompute
type InfraComputeStatus struct {
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

// InfraCompute is the Schema for the infracomputes API
type InfraCompute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InfraComputeSpec   `json:"spec,omitempty"`
	Status InfraComputeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InfraComputeList contains a list of InfraCompute
type InfraComputeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InfraCompute `json:"items"`
}

// IsReady - returns true if service is ready
func (instance InfraCompute) IsReady() bool {
	return instance.Status.Conditions.IsTrue(condition.DeploymentReadyCondition)
}

func init() {
	SchemeBuilder.Register(&InfraCompute{}, &InfraComputeList{})
}
