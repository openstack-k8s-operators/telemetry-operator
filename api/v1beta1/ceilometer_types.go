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
	// CeilometerCentralContainerImage - default fall-back image for Ceilometer Central
	CeilometerCentralContainerImage = "quay.io/podified-antelope-centos9/openstack-ceilometer-central:current-podified"
	// CeilometerNotificationContainerImage - default fall-back image for Ceilometer Notification
	CeilometerNotificationContainerImage = "quay.io/podified-antelope-centos9/openstack-ceilometer-notification:current-podified"
	// CeilometerSgCoreContainerImage - default fall-back image for Ceilometer SgCore
	CeilometerSgCoreContainerImage = "quay.io/infrawatch/sg-core:v5.2.0-nextgen"
	// CeilometerComputeContainerImage - default fall-back image for Ceilometer Compute
	CeilometerComputeContainerImage = "quay.io/podified-antelope-centos9/openstack-ceilometer-compute:current-podified"
	// CeilometerIpmiContainerImage - default fall-back image for Ceilometer Ipmi
	CeilometerIpmiContainerImage = "quay.io/podified-antelope-centos9/openstack-ceilometer-ipmi:current-podified"
)

// CeilometerSpec defines the desired state of Ceilometer
type CeilometerSpec struct {
	// RabbitMQ instance name
	// Needed to request a transportURL that is created and used in Telemetry
	// +kubebuilder:default=rabbitmq
	RabbitMqClusterName string `json:"rabbitMqClusterName,omitempty"`

	// PasswordSelectors - Selectors to identify the service from the Secret
	// +kubebuilder:default:={service: CeilometerPassword}
	PasswordSelectors PasswordsSelector `json:"passwordSelector,omitempty"`

	// ServiceUser - optional username used for this service to register in keystone
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=ceilometer
	ServiceUser string `json:"serviceUser"`

	// Secret containing OpenStack password information for ceilometer
	// +kubebuilder:validation:Required
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
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Ceilometer is the Schema for the ceilometers API
type Ceilometer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CeilometerSpec   `json:"spec,omitempty"`
	Status CeilometerStatus `json:"status,omitempty"`
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
		CentralContainerImageURL:      util.GetEnvVar("RELATED_IMAGE_CEILOMETER_CENTRAL_IMAGE_URL_DEFAULT", CeilometerCentralContainerImage),
		SgCoreContainerImageURL:       util.GetEnvVar("RELATED_IMAGE_CEILOMETER_SGCORE_IMAGE_URL_DEFAULT", CeilometerSgCoreContainerImage),
		NotificationContainerImageURL: util.GetEnvVar("RELATED_IMAGE_CEILOMETER_NOTIFICATION_IMAGE_URL_DEFAULT", CeilometerNotificationContainerImage),
		ComputeContainerImageURL:      util.GetEnvVar("RELATED_IMAGE_CEILOMETER_COMPUTE_IMAGE_URL_DEFAULT", CeilometerComputeContainerImage),
		IpmiContainerImageURL:         util.GetEnvVar("RELATED_IMAGE_CEILOMETER_IPMI_IMAGE_URL_DEFAULT", CeilometerIpmiContainerImage),
	}

	SetupCeilometerDefaults(ceilometerDefaults)
}
