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

// TODO: We might want to split this to aodh and ceilometer and move it to appropriate files

// PasswordsSelector to identify the Service password from the Secret
type PasswordsSelector struct {
	// Service - Selector to get the ceilometer service password from the Secret
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=CeilometerPassword
	Service string `json:"service"`

	// AodhService - Selector to get the aodh service password from the Secret
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=AodhPassword
	AodhService string `json:"aodhService"`

	// Database - Selector to get the aodh database user password from the Secret
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=AodhDatabasePassword
	Database string `json:"database"`
}

// TelemetrySpec defines the desired state of Telemetry
type TelemetrySpec struct {
	// +kubebuilder:validation:Optional
	// Autoscaling - Parameters related to the autoscaling service
	Autoscaling AutoscalingSection `json:"autoscaling,omitempty"`

	// +kubebuilder:validation:Optional
	// Ceilometer - Parameters related to the ceilometer service
	Ceilometer CeilometerSection `json:"ceilometer,omitempty"`

	// +kubebuilder:validation:Optional
	// MetricStorage - Parameters related to the metricStorage
	MetricStorage MetricStorageSection `json:"metricStorage,omitempty"`

	// +kubebuilder:validation:Optional
	// Logging - Parameters related to the logging
	Logging LoggingSection `json:"logging,omitempty"`
}

// CeilometerSection defines the desired state of the ceilometer service
type CeilometerSection struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	// +operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch"}
	// Enabled - Whether OpenStack Ceilometer service should be deployed and managed
	Enabled bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// Template - Overrides to use when creating the OpenStack Ceilometer service
	CeilometerSpec `json:",inline"`
}

// AutoscalingSection defines the desired state of the autoscaling service
type AutoscalingSection struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// +operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch"}
	// Enabled - Whether OpenStack autoscaling service should be deployed and managed
	Enabled bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// Template - Overrides to use when creating the OpenStack autoscaling service
	AutoscalingSpec `json:",inline"`
}

// MetricStorageSection defines the desired state of the MetricStorage
type MetricStorageSection struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// +operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch"}
	// Enabled - Whether a MetricStorage should be deployed and managed
	Enabled bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// Template - Overrides to use when creating the MetricStorage
	MetricStorageSpec `json:",inline"`
}

// LoggingSection defines the desired state of the logging service
type LoggingSection struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	// +operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch"}
	// Enabled - Whether OpenStack logging service should be deployed and managed
	Enabled bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// Template - Overrides to use when creating the OpenStack Logging
	LoggingSpec `json:",inline"`
}

// TelemetryStatus defines the observed state of Telemetry
type TelemetryStatus struct {
	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`
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
		CentralContainerImageURL:       util.GetEnvVar("RELATED_IMAGE_CEILOMETER_CENTRAL_IMAGE_URL_DEFAULT", CeilometerCentralContainerImage),
		ComputeContainerImageURL:       util.GetEnvVar("RELATED_IMAGE_CEILOMETER_COMPUTE_IMAGE_URL_DEFAULT", CeilometerComputeContainerImage),
		IpmiContainerImageURL:          util.GetEnvVar("RELATED_IMAGE_CEILOMETER_IPMI_IMAGE_URL_DEFAULT", CeilometerIpmiContainerImage),
		NotificationContainerImageURL:  util.GetEnvVar("RELATED_IMAGE_CEILOMETER_NOTIFICATION_IMAGE_URL_DEFAULT", CeilometerNotificationContainerImage),
		SgCoreContainerImageURL:        util.GetEnvVar("RELATED_IMAGE_CEILOMETER_SGCORE_IMAGE_URL_DEFAULT", CeilometerSgCoreContainerImage),

		// Autoscaling
		AodhAPIContainerImageURL:       util.GetEnvVar("RELATED_IMAGE_AODH_API_IMAGE_URL_DEFAULT", AodhAPIContainerImage),
		AodhEvaluatorContainerImageURL: util.GetEnvVar("RELATED_IMAGE_AODH_EVALUATOR_IMAGE_URL_DEFAULT", AodhEvaluatorContainerImage),
		AodhNotifierContainerImageURL:  util.GetEnvVar("RELATED_IMAGE_AODH_NOTIFIER_IMAGE_URL_DEFAULT", AodhNotifierContainerImage),
		AodhListenerContainerImageURL:  util.GetEnvVar("RELATED_IMAGE_AODH_LISTENER_IMAGE_URL_DEFAULT", AodhListenerContainerImage),
		AodhInitContainerImageURL:      util.GetEnvVar("RELATED_IMAGE_AODH_API_IMAGE_URL_DEFAULT", AodhAPIContainerImage),

	}

	SetupTelemetryDefaults(telemetryDefaults)
}
