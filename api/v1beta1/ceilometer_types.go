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
	rabbitmqv1 "github.com/openstack-k8s-operators/infra-operator/apis/rabbitmq/v1beta1"
	topologyv1 "github.com/openstack-k8s-operators/infra-operator/apis/topology/v1beta1"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	// CeilometerCentralContainerImage - default fall-back image for Ceilometer Central
	CeilometerCentralContainerImage = "quay.io/podified-antelope-centos9/openstack-ceilometer-central:current-podified"
	// CeilometerNotificationContainerImage - default fall-back image for Ceilometer Notification
	CeilometerNotificationContainerImage = "quay.io/podified-antelope-centos9/openstack-ceilometer-notification:current-podified"
	// CeilometerSgCoreContainerImage - default fall-back image for Ceilometer SgCore
	CeilometerSgCoreContainerImage = "quay.io/openstack-k8s-operators/sg-core:latest"
	// CeilometerComputeContainerImage - default fall-back image for Ceilometer Compute
	CeilometerComputeContainerImage = "quay.io/podified-antelope-centos9/openstack-ceilometer-compute:current-podified"
	// CeilometerIpmiContainerImage - default fall-back image for Ceilometer Ipmi
	CeilometerIpmiContainerImage = "quay.io/podified-antelope-centos9/openstack-ceilometer-ipmi:current-podified"
	// CeilometerProxyContainerImage - default fall-back image for proxy container
	// CeilometerProxyContainerImage = "registry.redhat.io/ubi9/httpd-24:latest"
	CeilometerProxyContainerImage = "quay.io/podified-antelope-centos9/openstack-aodh-api:current-podified"
	// KubeStateMetricsImage - default fall-back image for KSM
	KubeStateMetricsImage = "registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.10.0"
	// MysqldExporterImage - default fall-back image for mysqld_exporter
	MysqldExporterContainerImage = "quay.io/prometheus/mysqld-exporter:v0.15.1"
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

	// +kubebuilder:validation:Optional
	MysqldExporterImage string `json:"mysqldExporterImage"`
}

// CeilometerSpecCore defines the desired state of Ceilometer. This version is used by the OpenStackControlplane (no image parameters)
type CeilometerSpecCore struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=60
	// APITimeout for Apache
	APITimeout int `json:"apiTimeout"`

	// +kubebuilder:validation:Optional
	// NotificationsBus configuration (username, vhost, and cluster) for notifications
	NotificationsBus *rabbitmqv1.RabbitMqConfig `json:"notificationsBus,omitempty"`

	// +kubebuilder:validation:Optional
	// RabbitMQ instance name
	// Needed to request a transportURL that is created and used in Telemetry
	// Deprecated: Use NotificationsBus.Cluster instead
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
	// +listType=atomic
	NetworkAttachmentDefinitions []string `json:"networkAttachmentDefinitions,omitempty"`

	// Whether kube-state-metrics should be deployed
	// +kubebuilder:validation:optional
	// +kubebuilder:default=true
	KSMEnabled *bool `json:"ksmEnabled,omitempty"`

	// Whether mysqld_exporter should be deployed
	// +kubebuilder:validation:optional
	MysqldExporterEnabled *bool `json:"mysqldExporterEnabled,omitempty"`

	// MysqldExporterDatabaseAccountPrefix - Database account prefix for the mysqld-exporter.
	// A mariadbaccount CR named "<mysqldExporterDatabaseAccountPrefix>-<galera CR name>" for each
	// galera instance needs to be either created by the user or if it's missing, it'll be
	// created by the telemetry-operator automatically.
	// +kubebuilder:validation:optional
	// +kubebuilder:default=mysqld-exporter
	MysqldExporterDatabaseAccountPrefix string `json:"mysqldExporterDatabaseAccountPrefix,omitempty"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// TLS - Parameters related to the TLS
	TLS tls.SimpleService `json:"tls,omitempty"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// KSMTLS - Parameters related to the TLS for kube-state-metrics
	KSMTLS tls.SimpleService `json:"ksmTls,omitempty"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// MysqldExporterTLS - Parameters related to the TLS for mysqld_exporter
	MysqldExporterTLS tls.SimpleService `json:"mysqldExporterTLS,omitempty"`

	// +kubebuilder:validation:Optional
	// NodeSelector to target subset of worker nodes running this service
	NodeSelector *map[string]string `json:"nodeSelector,omitempty"`

	// +kubebuilder:validation:Optional
	// TopologyRef to apply the Topology defined by the associated CR referenced
	// by name
	TopologyRef *topologyv1.TopoRef `json:"topologyRef,omitempty"`

	// +kubebuilder:validation:Optional
	// A name of a secret containing custom configuration files. Files
	// from this secret will get copied into /etc/ceilometer/ and they'll
	// overwrite any default files already present there.
	CustomConfigsSecretName string `json:"customConfigsSecretName,omitempty"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// Auth - authentication settings for keystone integration
	Auth AuthSpec `json:"auth,omitempty"`
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

	// NotificationsURLSecret - Secret containing RabbitMQ notification transportURL
	NotificationsURLSecret *string `json:"notificationsURLSecret,omitempty"`

	// Networks in addtion to the cluster network, the service is attached to
	Networks []string `json:"networks,omitempty"`

	// ObservedGeneration - the most recent generation observed for this
	// service. If the observed generation is less than the spec generation,
	// then the controller has not processed the latest changes injected by
	// the openstack-operator in the top-level CR (e.g. the ContainerImage)
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// ReadyCount of mysqld_exporter instances
	MysqldExporterReadyCount int32 `json:"mysqldExporterReadyCount,omitempty"`

	// Map of hashes to track e.g. job status
	MysqldExporterHash map[string]string `json:"mysqldExporterHash,omitempty"`

	// List of galera CRs, which are being exported with mysqld_exporter
	// +listType=atomic
	MysqldExporterExportedGaleras []string `json:"mysqldExporterExportedGaleras,omitempty"`

	// ReadyCount of kube-state-metrics instances
	KSMReadyCount int32 `json:"ksmReadyCount,omitempty"`

	// Map of hashes to track e.g. job status
	KSMHash map[string]string `json:"ksmHash,omitempty"`

	// LastAppliedTopology - the last applied Topology
	LastAppliedTopology *topologyv1.TopoRef `json:"lastAppliedTopology,omitempty"`
}

// NOTE(mmagr): remove KSMStatus with API version increment

// KSMStatus defines the observed state of kube-state-metrics [DEPRECATED, Status is used instead]
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

	Spec      CeilometerSpec   `json:"spec,omitempty"`
	Status    CeilometerStatus `json:"status,omitempty"`
	KSMStatus KSMStatus        `json:"ksmStatus,omitempty"`
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
	return instance.Status.Conditions.IsTrue(condition.ReadyCondition)
}

func init() {
	SchemeBuilder.Register(&Ceilometer{}, &CeilometerList{})
}

// RbacConditionsSet - set the conditions for the rbac object
func (instance Ceilometer) RbacConditionsSet(c *condition.Condition) {
	instance.Status.Conditions.Set(c)
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
		CentralContainerImageURL:        util.GetEnvVar("RELATED_IMAGE_CEILOMETER_CENTRAL_IMAGE_URL_DEFAULT", CeilometerCentralContainerImage),
		SgCoreContainerImageURL:         util.GetEnvVar("RELATED_IMAGE_CEILOMETER_SGCORE_IMAGE_URL_DEFAULT", CeilometerSgCoreContainerImage),
		NotificationContainerImageURL:   util.GetEnvVar("RELATED_IMAGE_CEILOMETER_NOTIFICATION_IMAGE_URL_DEFAULT", CeilometerNotificationContainerImage),
		ComputeContainerImageURL:        util.GetEnvVar("RELATED_IMAGE_CEILOMETER_COMPUTE_IMAGE_URL_DEFAULT", CeilometerComputeContainerImage),
		IpmiContainerImageURL:           util.GetEnvVar("RELATED_IMAGE_CEILOMETER_IPMI_IMAGE_URL_DEFAULT", CeilometerIpmiContainerImage),
		ProxyContainerImageURL:          util.GetEnvVar("RELATED_IMAGE_APACHE_IMAGE_URL_DEFAULT", CeilometerProxyContainerImage),
		KSMContainerImageURL:            util.GetEnvVar("RELATED_IMAGE_KSM_IMAGE_URL_DEFAULT", KubeStateMetricsImage),
		MysqldExporterContainerImageURL: util.GetEnvVar("RELATED_IMAGE_MYSQLD_EXPORTER_IMAGE_URL_DEFAULT", MysqldExporterContainerImage),
	}

	SetupCeilometerDefaults(ceilometerDefaults)
}

// GetSpecTopologyRef - Returns the LastAppliedTopology Set in the Status
func (instance *Ceilometer) GetSpecTopologyRef() *topologyv1.TopoRef {
	return instance.Spec.TopologyRef
}

// GetLastAppliedTopology - Returns the LastAppliedTopology Set in the Status
func (instance *Ceilometer) GetLastAppliedTopology() *topologyv1.TopoRef {
	return instance.Status.LastAppliedTopology
}

// SetLastAppliedTopology - Sets the LastAppliedTopology value in the Status
func (instance *Ceilometer) SetLastAppliedTopology(topologyRef *topologyv1.TopoRef) {
	instance.Status.LastAppliedTopology = topologyRef
}

// ValidateTopology -
func (instance *CeilometerSpecCore) ValidateTopology(
	basePath *field.Path,
	namespace string,
) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = append(allErrs, topologyv1.ValidateTopologyRef(
		instance.TopologyRef,
		*basePath.Child("topologyRef"), namespace)...)
	return allErrs
}
