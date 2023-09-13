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
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
)

// PasswordsSelector to identify the Service password from the Secret
type PasswordsSelector struct {
	// Service - Selector to get the ceilometer service password from the Secret
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=CeilometerPassword
	Service string `json:"service"`
}

// TelemetrySpec defines the desired state of Telemetry
type TelemetrySpec struct {
	// +kubebuilder:validation:Required
	// Autoscaling - Spec definition for the Autoscaling service of this Telemetry deployment
	Autoscaling AutoscalingSpec `json:"autoscaling"`

	// +kubebuilder:validation:Required
	// Ceilometer - Spec definition for the Ceilometer service of this Telemetry deployment
	Ceilometer CeilometerSpec `json:"ceilometer"`
}

// TelemetryStatus defines the observed state of Telemetry
type TelemetryStatus struct {
	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`

	// ReadyCount of Ceilometer instance
	CeilometerReadyCount int32 `json:"ceilometerReadyCount,omitempty"`

	// ReadyCount of Autoscaling instance
	AutoscalingReadyCount int32 `json:"autoscalingReadyCount,omitempty"`
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

// IsReady - returns true if Telemetry is reconciled successfully
func (instance Telemetry) IsReady() bool {
	return instance.Status.Conditions.IsTrue(condition.ReadyCondition)
}

func init() {
	SchemeBuilder.Register(&Telemetry{}, &TelemetryList{})
}

// SetupDefaultsTelemetry - initializes any CRD field defaults based on environment variables (the defaulting mechanism itself is implemented via webhooks)
func SetupDefaultsTelemetry() {
	// Acquire environmental defaults and initialize Telemetry defaults with them
	telemetryDefaults := TelemetryDefaults{
		CentralContainerImageURL:      util.GetEnvVar("RELATED_IMAGE_CEILOMETER_CENTRAL_IMAGE_URL_DEFAULT", CeilometerCentralContainerImage),
		ComputeContainerImageURL:      util.GetEnvVar("RELATED_IMAGE_CEILOMETER_COMPUTE_IMAGE_URL_DEFAULT", CeilometerComputeContainerImage),
		IpmiContainerImageURL:         util.GetEnvVar("RELATED_IMAGE_CEILOMETER_IPMI_IMAGE_URL_DEFAULT", CeilometerIpmiContainerImage),
		NotificationContainerImageURL: util.GetEnvVar("RELATED_IMAGE_CEILOMETER_NOTIFICATION_IMAGE_URL_DEFAULT", CeilometerNotificationContainerImage),
		NodeExporterContainerImageURL: util.GetEnvVar("RELATED_IMAGE_TELEMETRY_NODE_EXPORTER_IMAGE_URL_DEFAULT", NodeExporterContainerImage),
		SgCoreContainerImageURL:       util.GetEnvVar("RELATED_IMAGE_CEILOMETER_SGCORE_IMAGE_URL_DEFAULT", CeilometerSgCoreContainerImage),
	}

	SetupTelemetryDefaults(telemetryDefaults)
}
