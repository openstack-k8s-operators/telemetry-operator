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

package metricstorage

import (
	"fmt"
	"reflect"

	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	monv1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1"
	obov1 "github.com/rhobs/observability-operator/pkg/apis/monitoring/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MonitoringStack defines a monitoringstack for metricStorage
func MonitoringStack(
	instance *telemetryv1.MetricStorage,
	labels map[string]string,
) (*obov1.MonitoringStack, error) {
	if instance.Spec.MonitoringStack == nil {
		return nil, fmt.Errorf("monitoringStack is set to nil")
	}
	pvc, err := getPVCSpec(instance)
	if err != nil {
		return nil, err
	}
	monitoringStack := &obov1.MonitoringStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: obov1.MonitoringStackSpec{
			AlertmanagerConfig: obov1.AlertmanagerConfig{
				Disabled: !instance.Spec.MonitoringStack.AlertingEnabled,
			},
			PrometheusConfig: &obov1.PrometheusConfig{
				Replicas: &telemetryv1.PrometheusReplicas,
				// NOTE: unsupported before OBOv0.0.21, but we can set the value
				//       in the ServiceMonitor, so this isn't a big deal.
				//ScrapeInterval: instance.Spec.MonitoringStack.ScrapeInterval,
				PersistentVolumeClaim: pvc,
			},
			Retention: monv1.Duration(instance.Spec.MonitoringStack.Storage.Retention),
			ResourceSelector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
		},
	}
	return monitoringStack, nil
}

func getPVCSpec(instance *telemetryv1.MetricStorage) (*corev1.PersistentVolumeClaimSpec, error) {
	if instance.Spec.MonitoringStack.Storage.Strategy == "persistent" {
		persistentSpec := &instance.Spec.MonitoringStack.Storage.Persistent
		pvc := corev1.PersistentVolumeClaimSpec{}
		if !reflect.DeepEqual(persistentSpec.PvcStorageSelector, metav1.LabelSelector{}) {
			pvc.Selector = &persistentSpec.PvcStorageSelector
		}
		if persistentSpec.PvcStorageClass != "" {
			pvc.StorageClassName = &persistentSpec.PvcStorageClass
		}
		var quantity resource.Quantity
		var err error
		if persistentSpec.PvcStorageRequest == "" {
			// NOTE: We can't rely on defaults in the CRD here, because we want
			// to be able to omit the spec.storage.persistent field. In case
			// the persistent field is omited, we won't have anything in
			// persistentSpec.PvcStorageRequest and we need to set the default
			// value like this here.
			quantity, err = resource.ParseQuantity(telemetryv1.DefaultPvcStorageRequest)
		} else {
			quantity, err = resource.ParseQuantity(persistentSpec.PvcStorageRequest)
		}
		if err != nil {
			return nil, err
		}
		pvc.Resources = corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				"storage": quantity,
			},
		}
		return &pvc, nil
	}
	return nil, nil
}
