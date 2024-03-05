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
	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
)

const (
	// AodhAPIContainerImage - default fall-back image for Aodh API
	AodhAPIContainerImage = "quay.io/podified-antelope-centos9/openstack-aodh-api:current-podified"
	// AodhEvaluatorContainerImage - default fall-back image for Aodh Evaluator
	AodhEvaluatorContainerImage = "quay.io/podified-antelope-centos9/openstack-aodh-evaluator:current-podified"
	// AodhNotifierContainerImage - default fall-back image for Aodh Notifier
	AodhNotifierContainerImage = "quay.io/podified-antelope-centos9/openstack-aodh-notifier:current-podified"
	// AodhListenerContainerImage - default fall-back image for Aodh Listener
	AodhListenerContainerImage = "quay.io/podified-antelope-centos9/openstack-aodh-listener:current-podified"
	// DbSyncHash hash
	DbSyncHash = "dbsync"
)

// Aodh defines the aodh component spec
type Aodh struct {
	// RabbitMQ instance name
	// Needed to request a transportURL that is created and used in Aodh
	// +kubebuilder:default=rabbitmq
	RabbitMqClusterName string `json:"rabbitMqClusterName,omitempty"`

	// +kubebuilder:validation:Required
	// MariaDB instance name
	// Right now required by the maridb-operator to get the credentials from the instance to create the DB
	// Might not be required in future
	DatabaseInstance string `json:"databaseInstance"`

	// Database user name
	// Needed to connect to a database used by aodh
	// +kubebuilder:default=aodh
	DatabaseUser string `json:"databaseUser,omitempty"`

	// PasswordSelectors - Selectors to identify the service from the Secret
	// +kubebuilder:default:={aodhService: AodhPassword, database: AodhDatabasePassword}
	PasswordSelectors PasswordsSelector `json:"passwordSelector,omitempty"`

	// ServiceUser - optional username used for this service to register in keystone
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=aodh
	ServiceUser string `json:"serviceUser"`

	// Secret containing OpenStack password information for aodh
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
	// +kubebuilder:validation:Optional
	DefaultConfigOverwrite map[string]string `json:"defaultConfigOverwrite,omitempty"`

	// NetworkAttachmentDefinitions list of network attachment definitions the service pod gets attached to
	// +kubebuilder:validation:Optional
	NetworkAttachmentDefinitions []string `json:"networkAttachmentDefinitions,omitempty"`

	// Override, provides the ability to override the generated manifest of several child resources.
	Override APIOverrideSpec `json:"override,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// PreserveJobs - do not delete jobs after they finished e.g. to check logs
	PreserveJobs bool `json:"preserveJobs"`

	// +kubebuilder:validation:Required
	// +kubebuilder:default=memcached
	// Memcached instance name.
	MemcachedInstance string `json:"memcachedInstance"`

	// +kubebuilder:validation:Required
	APIImage string `json:"apiImage"`

	// +kubebuilder:validation:Required
	EvaluatorImage string `json:"evaluatorImage"`

	// +kubebuilder:validation:Required
	NotifierImage string `json:"notifierImage"`

	// +kubebuilder:validation:Required
	ListenerImage string `json:"listenerImage"`

	// +kubebuilder:validation:Optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// TLS - Parameters related to the TLS
	TLS tls.API `json:"tls,omitempty"`
}

// APIOverrideSpec to override the generated manifest of several child resources.
type APIOverrideSpec struct {
	// Override configuration for the Service created to serve traffic to the cluster.
	Service *service.RoutedOverrideSpec `json:"service,omitempty"`
}

// AutoscalingSpec defines the desired state of Autoscaling
type AutoscalingSpec struct {
	// Host of user deployed prometheus
	// +kubebuilder:validation:Optional
	PrometheusHost string `json:"prometheusHost,omitempty"`

	// Port of user deployed prometheus
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Optional
	PrometheusPort int32 `json:"prometheusPort,omitempty"`

	// Aodh spec
	Aodh Aodh `json:"aodh,omitempty"`

	// Heat instance name.
	// +kubebuilder:default=heat
	HeatInstance string `json:"heatInstance"`
}

// AutoscalingStatus defines the observed state of Autoscaling
type AutoscalingStatus struct {
	// ReadyCount of autoscaling instances
	ReadyCount int32 `json:"readyCount,omitempty"`

	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`

	// Networks in addtion to the cluster network, the service is attached to
	Networks []string `json:"networks,omitempty"`

	// TransportURLSecret - Secret containing RabbitMQ transportURL
	TransportURLSecret string `json:"transportURLSecret,omitempty"`

	// DatabaseHostname - Hostname for the database
	DatabaseHostname string `json:"databaseHostname,omitempty"`

	// PrometheusHost - Hostname for prometheus used for autoscaling
	PrometheusHost string `json:"prometheusHostname,omitempty"`

	// PrometheusPort - Port for prometheus used for autoscaling
	PrometheusPort int32 `json:"prometheusPort,omitempty"`

	// API endpoint
	APIEndpoints map[string]string `json:"apiEndpoint,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Autoscaling is the Schema for the autoscalings API
type Autoscaling struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutoscalingSpec   `json:"spec,omitempty"`
	Status AutoscalingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AutoscalingList contains a list of Autoscaling
type AutoscalingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Autoscaling `json:"items"`
}

// IsReady - returns true if Autescaling is reconciled successfully
func (instance Autoscaling) IsReady() bool {
	return instance.Status.Conditions.IsTrue(condition.ReadyCondition)
}

func init() {
	SchemeBuilder.Register(&Autoscaling{}, &AutoscalingList{})
}

// RbacConditionsSet - set the conditions for the rbac object
func (instance Autoscaling) RbacConditionsSet(c *condition.Condition) {
	instance.Status.Conditions.Set(c)
}

// RbacNamespace - return the namespace
func (instance Autoscaling) RbacNamespace() string {
	return instance.Namespace
}

// RbacResourceName - return the name to be used for rbac objects (serviceaccount, role, rolebinding)
func (instance Autoscaling) RbacResourceName() string {
	return "telemetry-" + instance.Name
}

// SetupDefaultsAutoscaling - initializes any CRD field defaults based on environment variables (the defaulting mechanism itself is implemented via webhooks)
func SetupDefaultsAutoscaling() {
	// Acquire environmental defaults and initialize Telemetry defaults with them
	autoscalingDefaults := AutoscalingDefaults{
		AodhAPIContainerImageURL:       util.GetEnvVar("RELATED_IMAGE_AODH_API_IMAGE_URL_DEFAULT", AodhAPIContainerImage),
		AodhEvaluatorContainerImageURL: util.GetEnvVar("RELATED_IMAGE_AODH_EVALUATOR_IMAGE_URL_DEFAULT", AodhEvaluatorContainerImage),
		AodhNotifierContainerImageURL:  util.GetEnvVar("RELATED_IMAGE_AODH_NOTIFIER_IMAGE_URL_DEFAULT", AodhNotifierContainerImage),
		AodhListenerContainerImageURL:  util.GetEnvVar("RELATED_IMAGE_AODH_LISTENER_IMAGE_URL_DEFAULT", AodhListenerContainerImage),
	}

	SetupAutoscalingDefaults(autoscalingDefaults)
}
