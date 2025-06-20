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
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	topologyv1 "github.com/openstack-k8s-operators/infra-operator/apis/topology/v1beta1"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
)

// APIOverrideSpec to override the generated manifest of several child resources.
type APIOverrideSpec struct {
	// Override configuration for the Service created to serve traffic to the cluster.
	// The key must be the endpoint type (public, internal)
	Service map[service.Endpoint]service.RoutedOverrideSpec `json:"service,omitempty"`
}

// PasswordsSelector to identify the Service password from the Secret
type PasswordsSelector struct {
	// CeilometerService - Selector to get the ceilometer service password from the Secret
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=CeilometerPassword
	CeilometerService string `json:"ceilometerService"`

	// AodhService - Selector to get the aodh service password from the Secret
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=AodhPassword
	AodhService string `json:"aodhService"`

	// CloudKittyService - Selector to get the CloudKitty service password from the Secret
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=CloudKittyPassword
	CloudKittyService string `json:"cloudKittyService"`
}

// TelemetrySpec defines the desired state of Telemetry
type TelemetrySpec struct {
	TelemetrySpecBase `json:",inline"`

	// +kubebuilder:validation:Optional
	// Autoscaling - Parameters related to the autoscaling service
	Autoscaling AutoscalingSection `json:"autoscaling,omitempty"`

	// +kubebuilder:validation:Optional
	// Ceilometer - Parameters related to the ceilometer service
	Ceilometer CeilometerSection `json:"ceilometer,omitempty"`
}

// TelemetrySpecCore defines the desired state of Telemetry. This version has no image parameters and is used by OpenStackControlplane
type TelemetrySpecCore struct {
	TelemetrySpecBase `json:",inline"`

	// +kubebuilder:validation:Optional
	// Autoscaling - Parameters related to the autoscaling service
	Autoscaling AutoscalingSectionCore `json:"autoscaling,omitempty"`

	// +kubebuilder:validation:Optional
	// Ceilometer - Parameters related to the ceilometer service
	Ceilometer CeilometerSectionCore `json:"ceilometer,omitempty"`
}

// TelemetrySpecBase -
type TelemetrySpecBase struct {
	// +kubebuilder:validation:Optional
	// MetricStorage - Parameters related to the metricStorage
	MetricStorage MetricStorageSection `json:"metricStorage,omitempty"`

	// +kubebuilder:validation:Optional
	// Logging - Parameters related to the logging
	Logging LoggingSection `json:"logging,omitempty"`

	// +kubebuilder:validation:Optional
	// CloudKitty - Parameters related to the cloudkitty service
	CloudKitty CloudKittySection `json:"cloudkitty,omitempty"`

	// +kubebuilder:validation:Optional
	// NodeSelector to target subset of worker nodes running this service
	NodeSelector *map[string]string `json:"nodeSelector,omitempty"`

	// +kubebuilder:validation:Optional
	// TopologyRef to apply the Topology defined by the associated CR referenced
	// by name
	TopologyRef *topologyv1.TopoRef `json:"topologyRef,omitempty"`
}

// CeilometerSection defines the desired state of the ceilometer service
type CeilometerSection struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	// +operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch"}
	// Enabled - Whether OpenStack Ceilometer service should be deployed and managed
	Enabled *bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// Template - Overrides to use when creating the OpenStack Ceilometer service
	CeilometerSpec `json:",inline"`
}

// CeilometerSectionCore defines the desired state of the ceilometer service
type CeilometerSectionCore struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	// +operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch"}
	// Enabled - Whether OpenStack Ceilometer service should be deployed and managed
	Enabled *bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// Template - Overrides to use when creating the OpenStack Ceilometer service
	CeilometerSpecCore `json:",inline"`
}

// AutoscalingSection defines the desired state of the autoscaling service
type AutoscalingSection struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// +operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch"}
	// Enabled - Whether OpenStack autoscaling service should be deployed and managed
	Enabled *bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// Template - Overrides to use when creating the OpenStack autoscaling service
	AutoscalingSpec `json:",inline"`
}

// AutoscalingSectionCore defines the desired state of the autoscaling service
type AutoscalingSectionCore struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// +operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch"}
	// Enabled - Whether OpenStack autoscaling service should be deployed and managed
	Enabled *bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// Template - Overrides to use when creating the OpenStack autoscaling service
	AutoscalingSpecCore `json:",inline"`
}

// MetricStorageSection defines the desired state of the MetricStorage
type MetricStorageSection struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// +operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch"}
	// Enabled - Whether a MetricStorage should be deployed and managed
	Enabled *bool `json:"enabled"`

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
	Enabled *bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// Template - Overrides to use when creating the OpenStack Logging
	LoggingSpec `json:",inline"`
}

// CloudKittySpec defines the desired state of the cloudkitty service
type CloudKittySection struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	// +operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch"}
	// Enabled - Whether OpenStack CloudKitty service should be deployed and managed
	Enabled *bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	//+operator-sdk:csv:customresourcedefinitions:type=spec
	// Template - Overrides to use when creating the OpenStack CloudKitty service
	CloudKittySpec `json:",inline"`
}

// TelemetryStatus defines the observed state of Telemetry
type TelemetryStatus struct {
	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`

	// ObservedGeneration - the most recent generation observed for this
	// service. If the observed generation is less than the spec generation,
	// then the controller has not processed the latest changes injected by
	// the openstack-operator in the top-level CR (e.g. the ContainerImage)
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// LastAppliedTopology - the last applied Topology
	LastAppliedTopology *topologyv1.TopoRef `json:"lastAppliedTopology,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[0].status",description="Status"
//+kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.conditions[0].message",description="Message"

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
		CentralContainerImageURL:        util.GetEnvVar("RELATED_IMAGE_CEILOMETER_CENTRAL_IMAGE_URL_DEFAULT", CeilometerCentralContainerImage),
		ComputeContainerImageURL:        util.GetEnvVar("RELATED_IMAGE_CEILOMETER_COMPUTE_IMAGE_URL_DEFAULT", CeilometerComputeContainerImage),
		IpmiContainerImageURL:           util.GetEnvVar("RELATED_IMAGE_CEILOMETER_IPMI_IMAGE_URL_DEFAULT", CeilometerIpmiContainerImage),
		NotificationContainerImageURL:   util.GetEnvVar("RELATED_IMAGE_CEILOMETER_NOTIFICATION_IMAGE_URL_DEFAULT", CeilometerNotificationContainerImage),
		SgCoreContainerImageURL:         util.GetEnvVar("RELATED_IMAGE_CEILOMETER_SGCORE_IMAGE_URL_DEFAULT", CeilometerSgCoreContainerImage),
		ProxyContainerImageURL:          util.GetEnvVar("RELATED_IMAGE_APACHE_IMAGE_URL_DEFAULT", CeilometerProxyContainerImage),
		MysqldExporterContainerImageURL: util.GetEnvVar("RELATED_IMAGE_CEILOMETER_MYSQLD_EXPORTER_IMAGE_URL_DEFAULT", MysqldExporterContainerImage),

		// Autoscaling
		AodhAPIContainerImageURL:       util.GetEnvVar("RELATED_IMAGE_AODH_API_IMAGE_URL_DEFAULT", AodhAPIContainerImage),
		AodhEvaluatorContainerImageURL: util.GetEnvVar("RELATED_IMAGE_AODH_EVALUATOR_IMAGE_URL_DEFAULT", AodhEvaluatorContainerImage),
		AodhNotifierContainerImageURL:  util.GetEnvVar("RELATED_IMAGE_AODH_NOTIFIER_IMAGE_URL_DEFAULT", AodhNotifierContainerImage),
		AodhListenerContainerImageURL:  util.GetEnvVar("RELATED_IMAGE_AODH_LISTENER_IMAGE_URL_DEFAULT", AodhListenerContainerImage),
	}

	SetupTelemetryDefaults(telemetryDefaults)
}
