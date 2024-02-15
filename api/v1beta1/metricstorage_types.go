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
	infranetworkv1 "github.com/openstack-k8s-operators/infra-operator/apis/network/v1beta1"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	obov1 "github.com/rhobs/observability-operator/pkg/apis/monitoring/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PersistentStorage defines storage options used for persistent storage
type PersistentStorage struct {
	// PvcStorageRequest The amount of storage to request in PVC
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="20G"
	PvcStorageRequest string `json:"pvcStorageRequest"`

	// PvcStorageSelector The Label selector to specify in PVCs
	// +kubebuilder:validation:Optional
	PvcStorageSelector metav1.LabelSelector `json:"pvcStorageSelector,omitempty"`

	// PvcStorageClass The storage class to use for storing metrics
	// +kubebuilder:validation:Optional
	PvcStorageClass string `json:"pvcStorageClass,omitempty"`
}

// Storage defines the options used for storage of metrics
type Storage struct {
	// Strategy to use for storage. Can be "persistent"
	// or empty, in which case a COO default is used
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=persistent
	// +kubebuilder:default=persistent
	Strategy string `json:"strategy"`

	// Retention time for metrics
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="24h"
	Retention string `json:"retention"`

	// Used to specify the options of persistent storage when
	// strategy = "persistent"
	// +kubebuilder:validation:Optional
	Persistent PersistentStorage `json:"persistent"`
}

// MonitoringStack defines the options for a Red Hat supported metric storage
type MonitoringStack struct {
	// AlertingEnabled allows to enable or disable alertmanager
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	AlertingEnabled bool `json:"alertingEnabled"`

	// ScrapeInterval sets the interval between scrapes
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="30s"
	ScrapeInterval string `json:"scrapeInterval"`

	// DataplaneNetwork defines the network that will be used to scrape dataplane node_exporter endpoints
	// +kubebuilder:default=ctlplane
	DataplaneNetwork infranetworkv1.NetNameStr `json:"dataplaneNetwork"`

	// Storage allows to define options for how to store metrics
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={strategy: persistent, retention: "24h", persistent: {pvcStorageRequest: "20G"}}
	Storage `json:"storage"`
}

// MetricStorageSpec defines the desired state of MetricStorage
type MetricStorageSpec struct {
	// MonitoringStack allows to define a metric storage with
	// options supported by Red Hat
	// +kubebuilder:validation:Optional
	// +nullable
	MonitoringStack *MonitoringStack `json:"monitoringStack,omitempty"`

	// CustomMonitoringStack allows to deploy a custom monitoring
	// stack when the options in "MonitoringStack" aren't
	// enough
	// +kubebuilder:validation:Optional
	// +nullable
	CustomMonitoringStack *obov1.MonitoringStackSpec `json:"customMonitoringStack,omitempty"`
}

// MetricStorageStatus defines the observed state of MetricStorage
type MetricStorageStatus struct {
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MetricStorage is the Schema for the metricstorages API
type MetricStorage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MetricStorageSpec   `json:"spec,omitempty"`
	Status MetricStorageStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MetricStorageList contains a list of MetricStorage
type MetricStorageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MetricStorage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MetricStorage{}, &MetricStorageList{})
}

// IsReady - returns true if MetricStorage is reconciled successfully
func (instance MetricStorage) IsReady() bool {
	return instance.Status.Conditions.IsTrue(condition.ReadyCondition)
}
