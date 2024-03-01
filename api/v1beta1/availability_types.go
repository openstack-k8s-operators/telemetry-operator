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

	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
)

const (
	// KubeStateMetricsImage - default fall-back image for KSM
	KubeStateMetricsImage = "registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.10.0"
)

// AvailabilitySpec defines the availability monitoring component spec
type AvailabilitySpec struct {
	// +kubebuilder:validation:Required
	KSMImage string `json:"apiImage"`
}

// AvailabilityStatus defines the observed state of Availability
type AvailabilityStatus struct {
	// ReadyCount of availability instances
	ReadyCount int32 `json:"readyCount,omitempty"`

	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Availability is the Schema for the availability API
type Availability struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AvailabilitySpec   `json:"spec,omitempty"`
	Status AvailabilityStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AvailabilityList contains a list of Availability
type AvailabilityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Availability `json:"items"`
}

// IsReady - returns true if Availability is reconciled successfully
func (instance Availability) IsReady() bool {
	return instance.Status.Conditions.IsTrue(condition.ReadyCondition)
}

func init() {
	SchemeBuilder.Register(&Availability{}, &AvailabilityList{})
}

// RbacConditionsSet - set the conditions for the rbac object
func (instance Availability) RbacConditionsSet(c *condition.Condition) {
	instance.Status.Conditions.Set(c)
}

// RbacNamespace - return the namespace
func (instance Availability) RbacNamespace() string {
	return instance.Namespace
}

// RbacResourceName - return the name to be used for rbac objects (serviceaccount, role, rolebinding)
func (instance Availability) RbacResourceName() string {
	return "telemetry-" + instance.Name
}

// SetupDefaultsAvailability - initializes any CRD field defaults based on environment variables (the defaulting mechanism itself is implemented via webhooks)
func SetupDefaultsAvailability() {
	// Acquire environmental defaults and initialize Telemetry defaults with them
	availabilityDefaults := AvailabilityDefaults{
		KSMImageURL: util.GetEnvVar("RELATED_IMAGE_KSM_IMAGE_URL_DEFAULT", KubeStateMetricsImage),
	}

	SetupAvailabilityDefaults(availabilityDefaults)
}
