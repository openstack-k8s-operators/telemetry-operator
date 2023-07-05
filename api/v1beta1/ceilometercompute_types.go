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
	// CeilometerComputeContainerImage - default fall-back image for Ceilometer Compute
	CeilometerComputeContainerImage = "quay.io/podified-antelope-centos9/openstack-ceilometer-compute:current-podified"
	// CeilometerComputeInitContainerImage - default fall-back image for Ceilometer Compute Init
	CeilometerComputeInitContainerImage = "quay.io/podified-antelope-centos9/openstack-ceilometer-compute:current-podified"
	// CeilometerIpmiContainerImage - default fall-back image for Ceilometer Ipmi
	CeilometerIpmiContainerImage = "quay.io/podified-antelope-centos9/openstack-ceilometer-ipmi:current-podified"
)

// CeilometerComputeSpec defines the desired state of CeilometerCompute
type CeilometerComputeSpec struct {
	// RabbitMQ instance name
	// Needed to request a transportURL that is created and used in Telemetry
	// +kubebuilder:default=rabbitmq
	RabbitMqClusterName string `json:"rabbitMqClusterName,omitempty"`

	// TransportURLSecret contains the needed values to connect to RabbitMQ
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

	// +kubebuilder:default:="A ceilometer compute agent"
	Description string `json:"description,omitempty"`

	// CustomServiceConfig - customize the service config using this parameter to change service defaults,
	// or overwrite rendered information using raw OpenStack config format. The content gets added to
	// to /etc/<service>/<service>.conf.d directory as custom.conf file.
	// +kubebuilder:default:="# add your customization here"
	CustomServiceConfig string `json:"customServiceConfig,omitempty"`

	// DefaultConfigOverwrite - interface to overwrite default config files like e.g. logging.conf or policy.json.
	// But can also be used to add additional files. Those get added to the service config dir in /etc/<service> .
	// TODO: -> implement
	DefaultConfigOverwrite map[string]string `json:"defaultConfigOverwrite,omitempty"`

	// InitImage is the image used for the init container
	// +kubebuilder:validation:Required
	InitImage string `json:"initImage"`

	// ComputeImage is the image used for the ceilometer-agent-compute container
	// +kubebuilder:validation:Required
	ComputeImage string `json:"computeImage"`

	// IpmiImage is the image used for the ceilometer-agent-ipmi container
	// +kubebuilder:validation:Required
	IpmiImage string `json:"ipmiImage"`

	// DataplaneSSHSecret
	// +kubebuilder:validation:Required
	DataplaneSSHSecret string `json:"dataplaneSSHSecret"`

	// DataplaneInventoryConfigMap
	// +kubebuilder:validation:Required
	DataplaneInventoryConfigMap string `json:"dataplaneInventoryConfigMap"`

	// Playbook executed
	// +kubebuilder:default:="deploy_ceilometer.yml"
	Playbook string `json:"playbook,omitempty"`

	// ServiceAccount - service account name used internally to provide the default SA name
	ServiceAccount string `json:"serviceAccount"`
}

// CeilometerComputeStatus defines the observed state of CeilometerCompute
type CeilometerComputeStatus struct {
	// ReadyCount of ceilometercompute instances
	ReadyCount int32 `json:"readyCount,omitempty"`

	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CeilometerCompute is the Schema for the ceilometercomputes API
type CeilometerCompute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CeilometerComputeSpec   `json:"spec,omitempty"`
	Status CeilometerComputeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CeilometerComputeList contains a list of CeilometerCompute
type CeilometerComputeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CeilometerCompute `json:"items"`
}

// IsReady - returns true if CeilometerCompute is reconciled successfully
func (instance CeilometerCompute) IsReady() bool {
	return instance.Status.Conditions.IsTrue(condition.ReadyCondition)
}

func init() {
	SchemeBuilder.Register(&CeilometerCompute{}, &CeilometerComputeList{})
}

// RbacConditionsSet - set the conditions for the rbac object
func (instance CeilometerCompute) RbacConditionsSet(c *condition.Condition) {
	instance.Status.Conditions.Set(c)
}

// RbacNamespace - return the namespace
func (instance CeilometerCompute) RbacNamespace() string {
	return instance.Namespace
}

// RbacResourceName - return the name to be used for rbac objects (serviceaccount, role, rolebinding)
func (instance CeilometerCompute) RbacResourceName() string {
	return "telemetry-" + instance.Name
}

// SetupDefaultsCeilometerCompute - initializes any CRD field defaults based on environment variables (the defaulting mechanism itself is implemented via webhooks)
func SetupDefaultsCeilometerCompute() {
	// Acquire environmental defaults and initialize Telemetry defaults with them
	ceilometercomputeDefaults := CeilometerComputeDefaults{
		ComputeContainerImageURL:      util.GetEnvVar("CEILOMETER_COMPUTE_IMAGE_URL_DEFAULT", CeilometerComputeContainerImage),
		ComputeInitContainerImageURL:  util.GetEnvVar("CEILOMETER_COMPUTE_INIT_IMAGE_URL_DEFAULT", CeilometerComputeInitContainerImage),
		IpmiContainerImageURL:         util.GetEnvVar("CEILOMETER_IPMI_IMAGE_URL_DEFAULT", CeilometerIpmiContainerImage),
	}

	SetupCeilometerComputeDefaults(ceilometercomputeDefaults)
}
