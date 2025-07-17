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
	topologyv1 "github.com/openstack-k8s-operators/infra-operator/apis/topology/v1beta1"
	"github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// CloudKittyUserID - Kolla's cloudkitty UID comes from the 'cloudkitty-user' in
	// https://github.com/openstack/kolla/blob/master/kolla/common/users.py
	CloudKittyUserID = 42408
	// CloudKittyGroupID - Kolla's cloudkitty GID
	CloudKittyGroupID = 42408

	// CloudKittyAPIContainerImage - default fall-back image for CloudKitty API
	CloudKittyAPIContainerImage = "quay.io/podified-master-centos9/openstack-cloudkitty-api:current-podified"
	// CloudKittyProcContainerImage - default fall-back image for CloudKitty Processor
	CloudKittyProcContainerImage = "quay.io/podified-master-centos9/openstack-cloudkitty-processor:current-podified"
	// CloudKittyDbSyncHash hash
	CKDbSyncHash = "ckdbsync"
	// CKStorageInitHash hash
	CKStorageInitHash = "ckstorageinit"
	// CloudKittyReplicas - The number of replicas per each service deployed
	CloudKittyReplicas = 1
)

type CloudKittySpecBase struct {
	CloudKittyTemplate `json:",inline"`

	// +kubebuilder:validation:Required
	// MariaDB instance name
	// Right now required by the maridb-operator to get the credentials from the instance to create the DB
	// Might not be required in future
	DatabaseInstance string `json:"databaseInstance"`

	// +kubebuilder:validation:Required
	// +kubebuilder:default=rabbitmq
	// RabbitMQ instance name
	// Needed to request a transportURL that is created and used in CloudKitty
	RabbitMqClusterName string `json:"rabbitMqClusterName"`

	// +kubebuilder:validation:Required
	// +kubebuilder:default=memcached
	// Memcached instance name.
	MemcachedInstance string `json:"memcachedInstance"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// PreserveJobs - do not delete jobs after they finished e.g. to check logs
	PreserveJobs bool `json:"preserveJobs"`

	// +kubebuilder:validation:Optional
	// CustomServiceConfig - customize the service config for all CloudKitty services using this parameter to change service defaults,
	// or overwrite rendered information using raw OpenStack config format. The content gets added to
	// to /etc/<service>/<service>.conf.d directory as a custom config file.
	CustomServiceConfig string `json:"customServiceConfig,omitempty"`

	// +kubebuilder:validation:Optional
	// NodeSelector to target subset of worker nodes running this service. Setting
	// NodeSelector here acts as a default value and can be overridden by service
	// specific NodeSelector Settings.
	NodeSelector *map[string]string `json:"nodeSelector,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=60
	// +kubebuilder:validation:Minimum=10
	// APITimeout for HAProxy, Apache, and rpc_response_timeout
	APITimeout int `json:"apiTimeout"`

	// +kubebuilder:validation:Optional
	// TopologyRef to apply the Topology defined by the associated CR referenced
	// by name
	TopologyRef *topologyv1.TopoRef `json:"topologyRef,omitempty"`

	// Host of user deployed prometheus
	// +kubebuilder:validation:Optional
	PrometheusHost string `json:"prometheusHost,omitempty"`

	// Port of user deployed prometheus
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Optional
	PrometheusPort int32 `json:"prometheusPort,omitempty"`

	// If defined, specifies which CA certificate to use for user deployed prometheus
	// +kubebuilder:validation:Optional
	// +nullable
	PrometheusTLSCaCertSecret *corev1.SecretKeySelector `json:"prometheusTLSCaCertSecret,omitempty"`
}

// CloudKittySpecCore the same as CloudKittySpec without ContainerImage references
type CloudKittySpecCore struct {
	CloudKittySpecBase `json:",inline"`

	// +kubebuilder:validation:Required
	// CloudKittyAPI - Spec definition for the API service of this CloudKitty deployment
	CloudKittyAPI CloudKittyAPITemplateCore `json:"cloudKittyAPI"`

	// +kubebuilder:validation:Required
	// CloudKittyProc - Spec definition for the Scheduler service of this CloudKitty deployment
	CloudKittyProc CloudKittyProcTemplateCore `json:"cloudKittyProc"`
}

// CloudKittySpec defines the desired state of CloudKitty
type CloudKittySpec struct {
	CloudKittySpecBase `json:",inline"`

	// +kubebuilder:validation:Required
	// CloudKittyAPI - Spec definition for the API service of this CloudKitty deployment
	CloudKittyAPI CloudKittyAPITemplate `json:"cloudKittyAPI"`

	// +kubebuilder:validation:Required
	// CloudKittyProc - Spec definition for the Scheduler service of this CloudKitty deployment
	CloudKittyProc CloudKittyProcTemplate `json:"cloudKittyProc"`
}

// CloudKittyTemplate defines common input parameters used by all CloudKitty services
type CloudKittyTemplate struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=cloudkitty
	// ServiceUser - optional username used for this service to register in cloudkitty
	ServiceUser string `json:"serviceUser"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=cloudkitty
	// DatabaseAccount - optional MariaDBAccount used for cloudkitty DB, defaults to cloudkitty
	DatabaseAccount string `json:"databaseAccount"`

	// +kubebuilder:validation:Required
	// Secret containing OpenStack password information
	Secret string `json:"secret"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={cloudKittyService: CloudKittyPassword}
	// PasswordsSelectors - Selectors to identify the ServiceUser password from the Secret
	PasswordSelectors PasswordsSelector `json:"passwordSelector"`
}

// CloudKittyServiceTemplate defines the input parameters that can be defined for a given
// CloudKitty service
type CloudKittyServiceTemplate struct {

	// +kubebuilder:validation:Optional
	// NodeSelector to target subset of worker nodes running this service. Setting here overrides
	// any global NodeSelector settings within the CloudKitty CR.
	NodeSelector *map[string]string `json:"nodeSelector,omitempty"`

	// +kubebuilder:validation:Optional
	// CustomServiceConfig - customize the service config using this parameter to change service defaults,
	// or overwrite rendered information using raw OpenStack config format. The content gets added to
	// to /etc/<service>/<service>.conf.d directory as a custom config file.
	CustomServiceConfig string `json:"customServiceConfig,omitempty"`

	// +kubebuilder:validation:Optional
	// CustomServiceConfigSecrets - customize the service config using this parameter to specify Secrets
	// that contain sensitive service config data. The content of each Secret gets added to the
	// /etc/<service>/<service>.conf.d directory as a custom config file.
	CustomServiceConfigSecrets []string `json:"customServiceConfigSecrets,omitempty"`

	// +kubebuilder:validation:Optional
	// Resources - Compute Resources required by this service (Limits/Requests).
	// https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// +kubebuilder:validation:Optional
	// NetworkAttachments is a list of NetworkAttachment resource names to expose the services to the given network
	NetworkAttachments []string `json:"networkAttachments,omitempty"`

	// +kubebuilder:validation:Optional
	// TopologyRef to apply the Topology defined by the associated CR referenced
	// by name
	TopologyRef *topologyv1.TopoRef `json:"topologyRef,omitempty"`
}

// CloudKittyStatus defines the observed state of CloudKitty
type CloudKittyStatus struct {
	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`

	// CloudKitty Database Hostname
	DatabaseHostname string `json:"databaseHostname,omitempty"`

	// TransportURLSecret - Secret containing RabbitMQ transportURL
	TransportURLSecret string `json:"transportURLSecret,omitempty"`

	// API endpoints
	APIEndpoints map[string]map[string]string `json:"apiEndpoints,omitempty"`

	// ServiceIDs
	ServiceIDs map[string]string `json:"serviceIDs,omitempty"`

	// ReadyCount of CloudKitty API instance
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=0
	CloudKittyAPIReadyCount int32 `json:"cloudKittyAPIReadyCount"`

	// ReadyCount of CloudKitty Processor instances
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=0
	CloudKittyProcReadyCount int32 `json:"cloudKittyProcReadyCounts"`

	// ObservedGeneration - the most recent generation observed for this service.
	// If the observed generation is different than the spec generation, then the
	// controller has not started processing the latest changes, and the status
	// and its conditions are likely stale.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// PrometheusHost - Hostname for prometheus used for autoscaling
	PrometheusHost string `json:"prometheusHostname,omitempty"`

	// PrometheusPort - Port for prometheus used for autoscaling
	PrometheusPort int32 `json:"prometheusPort,omitempty"`

	// PrometheusTLS - Determines if TLS should be used for accessing prometheus
	PrometheusTLS bool `json:"prometheusTLS,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CloudKitty is the Schema for the cloudkitties API
type CloudKitty struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudKittySpec   `json:"spec,omitempty"`
	Status CloudKittyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CloudKittyList contains a list of CloudKitty
type CloudKittyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudKitty `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudKitty{}, &CloudKittyList{})
}

// SetupDefaultsCloudKitty - initializes any CRD field defaults based on environment variables (the defaulting mechanism itself is implemented via webhooks)
func SetupDefaultsCloudKitty() {
	// Acquire environmental defaults and initialize Telemetry defaults with them
	cloudKittyDefaults := CloudKittyDefaults{
		APIContainerImageURL:  util.GetEnvVar("RELATED_IMAGE_CLOUDKITTY_API_IMAGE_URL_DEFAULT", CloudKittyAPIContainerImage),
		ProcContainerImageURL: util.GetEnvVar("RELATED_IMAGE_CLOUDKITTY_PROCESSOR_IMAGE_URL_DEFAULT", CloudKittyProcContainerImage),
	}

	SetupCloudKittyDefaults(cloudKittyDefaults)
}

// IsReady - returns true if all subresources Ready condition is true
func (instance CloudKitty) IsReady() bool {
	return instance.Generation == instance.Status.ObservedGeneration &&
		instance.Status.Conditions.IsTrue(CloudKittyAPIReadyCondition) &&
		instance.Status.Conditions.IsTrue(CloudKittyProcReadyCondition)
}

// RbacConditionsSet - set the conditions for the rbac object
func (instance CloudKitty) RbacConditionsSet(c *condition.Condition) {
	instance.Status.Conditions.Set(c)
}

// RbacNamespace - return the namespace
func (instance CloudKitty) RbacNamespace() string {
	return instance.Namespace
}

// RbacResourceName - return the name to be used for rbac objects (serviceaccount, role, rolebinding)
func (instance CloudKitty) RbacResourceName() string {
	return "cloudkitty-" + instance.Name
}
