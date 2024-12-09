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

	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
)

const (
	// CeilometerCentralContainerImage - default fall-back image for Ceilometer Central
	CeilometerCentralContainerImage = "quay.io/podified-antelope-centos9/openstack-ceilometer-central:current-podified"
	// CeilometerNotificationContainerImage - default fall-back image for Ceilometer Notification
	CeilometerNotificationContainerImage = "quay.io/podified-antelope-centos9/openstack-ceilometer-notification:current-podified"
	// CeilometerSgCoreContainerImage - default fall-back image for Ceilometer SgCore
	CeilometerSgCoreContainerImage = "quay.io/openstack-k8s-operators/sg-core:v6.0.0"
	// CeilometerComputeContainerImage - default fall-back image for Ceilometer Compute
	CeilometerComputeContainerImage = "quay.io/podified-antelope-centos9/openstack-ceilometer-compute:current-podified"
	// CeilometerIpmiContainerImage - default fall-back image for Ceilometer Ipmi
	CeilometerIpmiContainerImage = "quay.io/podified-antelope-centos9/openstack-ceilometer-ipmi:current-podified"
	// CeilometerProxyContainerImage - default fall-back image for proxy container
	// CeilometerProxyContainerImage = "registry.redhat.io/ubi9/httpd-24:latest"
	CeilometerProxyContainerImage = "quay.io/podified-antelope-centos9/openstack-aodh-api:current-podified"
	// KubeStateMetricsImage - default fall-back image for KSM
	KubeStateMetricsImage = "registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.10.0"
)

// CeilometerSpec defines the desired state of Ceilometer
type CeilometerSpec struct {
	CeilometerSpecCore `json:",inline"`

	// +kubebuilder:validation:Required
	CentralImage string `json:"centralImage"`

	// +kubebuilder:validation:Required
	NotificationImage string `json:"notificationImage"`

	// +kubebuilder:validation:Required
	SgCoreImage string `json:"sgCoreImage"`

	// +kubebuilder:validation:Required
	ComputeImage string `json:"computeImage"`

	// +kubebuilder:validation:Required
	IpmiImage string `json:"ipmiImage"`

	// +kubebuilder:validation:Required
	ProxyImage string `json:"proxyImage"`

	// +kubebuilder:validation:Optional
	KSMImage string `json:"ksmImage"`
}

// CeilometerSpecCore defines the desired state of Ceilometer. This version is used by the OpenStackControlplane (no image parameters)
type CeilometerSpecCore struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=60
	// APITimeout for Apache
	APITimeout int `json:"apiTimeout"`

	// RabbitMQ instance name
	// Needed to request a transportURL that is created and used in Telemetry
	// +kubebuilder:default=rabbitmq
	RabbitMqClusterName string `json:"rabbitMqClusterName,omitempty"`

	// PasswordSelectors - Selectors to identify the service from the Secret
	// +kubebuilder:default:={ceilometerService: CeilometerPassword}
	PasswordSelectors PasswordsSelector `json:"passwordSelector,omitempty"`

	// ServiceUser - optional username used for this service to register in keystone
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=ceilometer
	ServiceUser string `json:"serviceUser"`

	// Secret containing OpenStack password information for ceilometer
	// +kubebuilder:validation:Required
	// +kubebuilder:default=osp-secret
	Secret string `json:"secret"`

	// CustomServiceConfig - customize the service config using this parameter to change service defaults,
	// or overwrite rendered information using raw OpenStack config format. The content gets added to
	// to /etc/<service>/<service>.conf.d directory as custom.conf file.
	// +kubebuilder:default:="# add your customization here"
	CustomServiceConfig string `json:"customServiceConfig,omitempty"`

	// ConfigOverwrite - interface to overwrite default config files like e.g. logging.conf or policy.json.
	// But can also be used to add additional files. Those get added to the service config dir in /etc/<service> .
	// TODO: -> implement
	DefaultConfigOverwrite map[string]string `json:"defaultConfigOverwrite,omitempty"`

	// NetworkAttachmentDefinitions list of network attachment definitions the service pod gets attached to
	NetworkAttachmentDefinitions []string `json:"networkAttachmentDefinitions,omitempty"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// TLS - Parameters related to the TLS
	TLS tls.SimpleService `json:"tls,omitempty"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// KSMTLS - Parameters related to the TLS for kube-state-metrics
	KSMTLS tls.SimpleService `json:"ksmTls,omitempty"`

	// +kubebuilder:validation:Optional
	// NodeSelector to target subset of worker nodes running this service
	NodeSelector *map[string]string `json:"nodeSelector,omitempty"`
}

// CeilometerStatus defines the observed state of Ceilometer
type CeilometerStatus struct {
	// ReadyCount of ceilometer instances
	ReadyCount int32 `json:"readyCount,omitempty"`

	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`

	// TransportURLSecret - Secret containing RabbitMQ transportURL
	TransportURLSecret string `json:"transportURLSecret,omitempty"`

	// Networks in addtion to the cluster network, the service is attached to
	Networks []string `json:"networks,omitempty"`

	// ObservedGeneration - the most recent generation observed for this
	// service. If the observed generation is less than the spec generation,
	// then the controller has not processed the latest changes injected by
	// the openstack-operator in the top-level CR (e.g. the ContainerImage)
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// KSMStatus defines the observed state of kube-state-metrics
type KSMStatus struct {
	// ReadyCount of ksm instances
	ReadyCount int32 `json:"readyCount,omitempty"`

	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`

	// ObservedGeneration - the most recent generation observed for this
	// service. If the observed generation is less than the spec generation,
	// then the controller has not processed the latest changes injected by
	// the openstack-operator in the top-level CR (e.g. the ContainerImage)
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[0].status",description="Status"
//+kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.conditions[0].message",description="Message"

// Ceilometer is the Schema for the ceilometers API
type Ceilometer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec             CeilometerSpec   `json:"spec,omitempty"`
	CeilometerStatus CeilometerStatus `json:"status,omitempty"`
	KSMStatus        KSMStatus        `json:"ksmStatus,omitempty"`
}

//+kubebuilder:object:root=true

// CeilometerList contains a list of Ceilometer
type CeilometerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ceilometer `json:"items"`
}

// IsReady - returns true if Ceilometer is reconciled successfully
func (instance Ceilometer) IsReady() bool {
	return instance.CeilometerStatus.Conditions.IsTrue(condition.ReadyCondition) &&
		instance.KSMStatus.Conditions.IsTrue(condition.ReadyCondition)
}

func init() {
	SchemeBuilder.Register(&Ceilometer{}, &CeilometerList{})
}

// RbacConditionsSet - set the conditions for the rbac object
func (instance Ceilometer) RbacConditionsSet(c *condition.Condition) {
	instance.CeilometerStatus.Conditions.Set(c)
}

// RbacNamespace - return the namespace
func (instance Ceilometer) RbacNamespace() string {
	return instance.Namespace
}

// RbacResourceName - return the name to be used for rbac objects (serviceaccount, role, rolebinding)
func (instance Ceilometer) RbacResourceName() string {
	return "telemetry-" + instance.Name
}

// SetupDefaultsCeilometer - initializes any CRD field defaults based on environment variables (the defaulting mechanism itself is implemented via webhooks)
func SetupDefaultsCeilometer() {
	// Acquire environmental defaults and initialize Telemetry defaults with them
	ceilometerDefaults := CeilometerDefaults{
		CentralContainerImageURL:      util.GetEnvVar("RELATED_IMAGE_CEILOMETER_CENTRAL_IMAGE_URL_DEFAULT", CeilometerCentralContainerImage),
		SgCoreContainerImageURL:       util.GetEnvVar("RELATED_IMAGE_CEILOMETER_SGCORE_IMAGE_URL_DEFAULT", CeilometerSgCoreContainerImage),
		NotificationContainerImageURL: util.GetEnvVar("RELATED_IMAGE_CEILOMETER_NOTIFICATION_IMAGE_URL_DEFAULT", CeilometerNotificationContainerImage),
		ComputeContainerImageURL:      util.GetEnvVar("RELATED_IMAGE_CEILOMETER_COMPUTE_IMAGE_URL_DEFAULT", CeilometerComputeContainerImage),
		IpmiContainerImageURL:         util.GetEnvVar("RELATED_IMAGE_CEILOMETER_IPMI_IMAGE_URL_DEFAULT", CeilometerIpmiContainerImage),
		ProxyContainerImageURL:        util.GetEnvVar("RELATED_IMAGE_APACHE_IMAGE_URL_DEFAULT", CeilometerProxyContainerImage),
		KSMContainerImageURL:          util.GetEnvVar("RELATED_IMAGE_KSM_IMAGE_URL_DEFAULT", KubeStateMetricsImage),
	}

	SetupCeilometerDefaults(ceilometerDefaults)
}
