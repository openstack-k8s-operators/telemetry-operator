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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var metricstoragelog = logf.Log.WithName("metricstorage-resource")

// SetupWebhookWithManager sets up the webhook with the Manager
func (r *MetricStorage) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-telemetry-openstack-org-v1beta1-metricstorage,mutating=true,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=metricstorages,verbs=create;update,versions=v1beta1,name=mmetricstorage.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &MetricStorage{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *MetricStorage) Default() {
	r.Spec.Default()
}

// Default - set defaults for the MetricStorage spec
func (spec *MetricStorageSpec) Default() {
	if spec.MonitoringStack == nil && spec.CustomMonitoringStack == nil {
		spec.MonitoringStack = &MonitoringStack{}
		// Set the AlertingEnabled to true here as the empty value means false
		// and we don't have a way of distinguishing if the value was left
		// empty or set to false by the user in the spec.MonitoringStack.Default()
		spec.MonitoringStack.AlertingEnabled = true
	}
	if spec.MonitoringStack != nil {
		spec.MonitoringStack.Default()
	}
}

// Default - set defaults for the MonitoringStack field
func (ms *MonitoringStack) Default() {
	if ms.ScrapeInterval == "" {
		ms.ScrapeInterval = "30s"
	}
	ms.Storage.Default()
}

// Default - set defaults for the Storage field
func (storage *Storage) Default() {
	if storage.Strategy == "" {
		storage.Strategy = "persistent"
	}
	if storage.Retention == "" {
		storage.Retention = "24h"
	}
	if storage.Strategy == "persistent" {
		storage.Persistent.Default()
	}
}

// Default - set defaults for the PersistentStorage field
func (ps *PersistentStorage) Default() {
	if ps.PvcStorageRequest == "" {
		ps.PvcStorageRequest = "20G"
	}
}

//+kubebuilder:webhook:path=/validate-telemetry-openstack-org-v1beta1-metricstorage,mutating=false,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=metricstorages,verbs=create;update,versions=v1beta1,name=vmetricstorage.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &MetricStorage{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *MetricStorage) ValidateCreate() (admission.Warnings, error) {
	metricstoragelog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *MetricStorage) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	metricstoragelog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *MetricStorage) ValidateDelete() (admission.Warnings, error) {
	metricstoragelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
