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
	// CeilometerCentralInitContainerImage - default fall-back image for Ceilometer Central Init
	CeilometerCentralInitContainerImage = "quay.io/podified-antelope-centos9/openstack-ceilometer-central:current-podified"
	// CeilometerNotificationContainerImage - default fall-back image for Ceilometer Notifcation
	CeilometerNotificationContainerImage = "quay.io/podified-antelope-centos9/openstack-ceilometer-notification:current-podified"
	// CeilometerSgCoreContainerImage - default fall-back image for Ceilometer SgCore
	CeilometerSgCoreContainerImage = "quay.io/infrawatch/sg-core:latest"
)

// CeilometerCentralSpec defines the desired state of CeilometerCentral
type CeilometerCentralSpec struct {
	// RabbitMQ instance name
	// Needed to request a transportURL that is created and used in Telemetry
	// +kubebuilder:default=rabbitmq
	RabbitMqClusterName string `json:"rabbitMqClusterName,omitempty"`

	// The needed values to connect to RabbitMQ
	TransportURLSecret string `json:"transportURLSecret,omitempty"`

	// PasswordSelectors - Selectors to identify the service from the Secret
	// +kubebuilder:default:={service: CeilometerPassword}
	PasswordSelectors PasswordsSelector `json:"passwordSelector,omitempty"`

	// ServiceUser - optional username used for this service to register in keystone
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=ceilometer
	ServiceUser string `json:"serviceUser"`

	// Secret containing OpenStack password information for ceilometer
	// +kubebuilder:validation:Optional
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

	// +kubebuilder:validation:Required
	CentralImage string `json:"centralImage"`

	// +kubebuilder:validation:Required
	NotificationImage string `json:"notificationImage"`

	// +kubebuilder:validation:Required
	SgCoreImage string `json:"sgCoreImage"`

	// +kubebuilder:validation:Required
	InitImage string `json:"initImage"`

	// +kubebuilder:default:="A ceilometer agent"
	Description string `json:"description,omitempty"`

	// ServiceAccount - service account name used internally to provide the default SA name
	ServiceAccount string `json:"serviceAccount"`
}

// CeilometerCentralStatus defines the observed state of CeilometerCentral
type CeilometerCentralStatus struct {
	// ReadyCount of ceilometercentral instances
	ReadyCount int32 `json:"readyCount,omitempty"`

	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`

	// Networks in addtion to the cluster network, the service is attached to
	Networks []string `json:"networks,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CeilometerCentral is the Schema for the ceilometercentrals API
type CeilometerCentral struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CeilometerCentralSpec   `json:"spec,omitempty"`
	Status CeilometerCentralStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CeilometerCentralList contains a list of CeilometerCentral
type CeilometerCentralList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CeilometerCentral `json:"items"`
}

// IsReady - returns true if CeilometerCentral is reconciled successfully
func (instance CeilometerCentral) IsReady() bool {
	return instance.Status.Conditions.IsTrue(condition.ReadyCondition)
}

func init() {
	SchemeBuilder.Register(&CeilometerCentral{}, &CeilometerCentralList{})
}

// RbacConditionsSet - set the conditions for the rbac object
func (instance CeilometerCentral) RbacConditionsSet(c *condition.Condition) {
	instance.Status.Conditions.Set(c)
}

// RbacNamespace - return the namespace
func (instance CeilometerCentral) RbacNamespace() string {
	return instance.Namespace
}

// RbacResourceName - return the name to be used for rbac objects (serviceaccount, role, rolebinding)
func (instance CeilometerCentral) RbacResourceName() string {
	return "telemetry-" + instance.Name
}

// SetupDefaultsCeilometerCentral - initializes any CRD field defaults based on environment variables (the defaulting mechanism itself is implemented via webhooks)
func SetupDefaultsCeilometerCentral() {
	// Acquire environmental defaults and initialize Telemetry defaults with them
	ceilometercentralDefaults := CeilometerCentralDefaults{
		CentralContainerImageURL:      util.GetEnvVar("CEILOMETER_CENTRAL_IMAGE_URL_DEFAULT", CeilometerCentralContainerImage),
		CentralInitContainerImageURL:  util.GetEnvVar("CEILOMETER_CENTRAL_INIT_IMAGE_URL_DEFAULT", CeilometerCentralInitContainerImage),
		SgCoreContainerImageURL:       util.GetEnvVar("CEILOMETER_SGCORE_IMAGE_URL_DEFAULT", CeilometerSgCoreContainerImage),
		NotificationContainerImageURL: util.GetEnvVar("CEILOMETER_NOTIFICATION_IMAGE_URL_DEFAULT", CeilometerNotificationContainerImage),
	}

	SetupCeilometerCentralDefaults(ceilometercentralDefaults)
}
