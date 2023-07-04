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
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
)

const (
	// NodeExporterContainerImage - default fall-back image for node_exporter
	// NodeExporterContainerImage = "registry.redhat.io/openshift4/ose-prometheus-node-exporter:v4.13"
	NodeExporterContainerImage = "quay.io/prometheus/node-exporter:v1.5.0"
)

// InfraComputeSpec defines the desired state of InfraCompute
type InfraComputeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:default:="An infra compute agent"
	Description string `json:"description,omitempty"`

	// NodeExporterImage is the image used for the node_exporter container
	// +kubebuilder:validation:Required
	NodeExporterImage string `json:"nodeExporterImage"`

	// DataplaneSSHSecret
	// +kubebuilder:validation:Required
	DataplaneSSHSecret string `json:"dataplaneSSHSecret"`

	// DataplaneInventoryConfigMap
	// +kubebuilder:validation:Required
	DataplaneInventoryConfigMap string `json:"dataplaneInventoryConfigMap"`

	// Playbook executed
	// +kubebuilder:default:="deploy_infra.yml"
	Playbook string `json:"playbook,omitempty"`

	// The extravars ConfigMap to pass to ansible execution
	// +kubebuilder:default:="telemetry-infracompute-extravars"
	ExtravarsConfigMap string `json:"extravarsConfigMap,omitempty"`

	// ServiceAccount - service account name used internally to provide the default SA name
	ServiceAccount string `json:"serviceAccount"`
}

// InfraComputeStatus defines the observed state of InfraCompute
type InfraComputeStatus struct {
	// ReadyCount of ceilometercompute instances
	ReadyCount int32 `json:"readyCount,omitempty"`

	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`
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

// IsReady - returns true if InfraCompute is reconciled successfully
func (instance InfraCompute) IsReady() bool {
	return instance.Status.Conditions.IsTrue(condition.ReadyCondition)
}

func init() {
	SchemeBuilder.Register(&InfraCompute{}, &InfraComputeList{})
}

// RbacConditionsSet - set the conditions for the rbac object
func (instance InfraCompute) RbacConditionsSet(c *condition.Condition) {
	instance.Status.Conditions.Set(c)
}

// RbacNamespace - return the namespace
func (instance InfraCompute) RbacNamespace() string {
	return instance.Namespace
}

// RbacResourceName - return the name to be used for rbac objects (serviceaccount, role, rolebinding)
func (instance InfraCompute) RbacResourceName() string {
	return "telemetry-" + instance.Name
}

// SetupDefaultsInfraCompute - initializes any CRD field defaults based on environment variables (the defaulting mechanism itself is implemented via webhooks)
func SetupDefaultsInfraCompute() {
	// Acquire environmental defaults and initialize Telemetry defaults with them
	infracomputeDefaults := InfraComputeDefaults{
		NodeExporterContainerImageURL: util.GetEnvVar("TELEMETRY_NODE_EXPORTER_IMAGE_URL_DEFAULT", NodeExporterContainerImage),
	}

	SetupInfraComputeDefaults(infracomputeDefaults)
}
