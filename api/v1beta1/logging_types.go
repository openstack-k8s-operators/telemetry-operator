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
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LoggingSpec defines the desired state of Logging
type LoggingSpec struct {
	// IPAddr is the address where the service will listen on
	// +kubebuilder:validation:Required
	IPAddr string `json:"ipaddr"`

	// Port is the port where the service will listen on
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=10514
	Port int32 `json:"port"`

	// TargetPort is the port where the logging syslog receiver is listening
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=10514
	TargetPort int `json:"targetPort"`

	// CLONamespace points to the namespace where the cluster-logging-operator is deployed
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=openshift-logging
	CLONamespace string `json:"cloNamespace"`

	// Annotations is a way to configure certain LoadBalancers, like MetalLB
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={metallb.universe.tf/address-pool: internalapi, metallb.universe.tf/allow-shared-ip: internalapi, metallb.universe.tf/loadBalancerIPs: "172.17.0.80"}
	Annotations map[string]string `json:"annotations"`

	// The number of retries rsyslog will attempt before abandoning
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=100
	RsyslogRetries int32 `json:"rsyslogRetries"`

	// The type of the local queue of logs
	// +kubebuilder:default=linkedList
	RsyslogQueueType string `json:"rsyslogQueueType"`

	// The size of the local queue of logs
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=10000
	RsyslogQueueSize int32 `json:"rsyslogQueueSize"`
}

// LoggingStatus defines the observed state of Logging
type LoggingStatus struct {
	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Logging is the Schema for the loggings API
type Logging struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoggingSpec   `json:"spec,omitempty"`
	Status LoggingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LoggingList contains a list of Logging
type LoggingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Logging `json:"items"`
}

// IsReady - returns true if Logging is reconciled successfully
func (instance Logging) IsReady() bool {
	return instance.Status.Conditions.IsTrue(condition.ReadyCondition)
}

func init() {
	SchemeBuilder.Register(&Logging{}, &LoggingList{})
}

// RbacConditionsSet - set the conditions for the rbac object
func (instance Logging) RbacConditionsSet(c *condition.Condition) {
	instance.Status.Conditions.Set(c)
}

// RbacNamespace - return the namespace
func (instance Logging) RbacNamespace() string {
	return instance.Namespace
}

// RbacResourceName - return the name to be used for rbac objects (serviceaccount, role, rolebinding)
func (instance Logging) RbacResourceName() string {
	return "telemetry-" + instance.Name
}
