/*
Copyright 2025.

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
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	telemetryv1beta1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
)

// nolint:unused
// log is for logging in this package.
var metricstoragelog = logf.Log.WithName("metricstorage-resource")

// SetupMetricStorageWebhookWithManager registers the webhook for MetricStorage in the manager.
func SetupMetricStorageWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&telemetryv1beta1.MetricStorage{}).
		WithValidator(&MetricStorageCustomValidator{}).
		WithDefaulter(&MetricStorageCustomDefaulter{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-telemetry-openstack-org-v1beta1-metricstorage,mutating=true,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=metricstorages,verbs=create;update,versions=v1beta1,name=mmetricstorage-v1beta1.kb.io,admissionReviewVersions=v1

// MetricStorageCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind MetricStorage when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type MetricStorageCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &MetricStorageCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind MetricStorage.
func (d *MetricStorageCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	metricstorage, ok := obj.(*telemetryv1beta1.MetricStorage)

	if !ok {
		return fmt.Errorf("%w: expected an MetricStorage object but got %T", ErrUnexpectedObjectType, obj)
	}
	metricstoragelog.Info("Defaulting for MetricStorage", "name", metricstorage.GetName())

	// Call the Default method on the MetricStorage type
	metricstorage.Default()

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-telemetry-openstack-org-v1beta1-metricstorage,mutating=false,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=metricstorages,verbs=create;update,versions=v1beta1,name=vmetricstorage-v1beta1.kb.io,admissionReviewVersions=v1

// MetricStorageCustomValidator struct is responsible for validating the MetricStorage resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type MetricStorageCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &MetricStorageCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type MetricStorage.
func (v *MetricStorageCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	metricstorage, ok := obj.(*telemetryv1beta1.MetricStorage)
	if !ok {
		return nil, fmt.Errorf("%w: expected a MetricStorage object but got %T", ErrUnexpectedObjectType, obj)
	}
	metricstoragelog.Info("Validation for MetricStorage upon creation", "name", metricstorage.GetName())

	// Call the ValidateCreate method on the MetricStorage type
	return metricstorage.ValidateCreate()
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type MetricStorage.
func (v *MetricStorageCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	metricstorage, ok := newObj.(*telemetryv1beta1.MetricStorage)
	if !ok {
		return nil, fmt.Errorf("%w: expected a MetricStorage object for the newObj but got %T", ErrUnexpectedObjectType, newObj)
	}
	metricstoragelog.Info("Validation for MetricStorage upon update", "name", metricstorage.GetName())

	// Call the ValidateUpdate method on the MetricStorage type
	return metricstorage.ValidateUpdate(oldObj)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type MetricStorage.
func (v *MetricStorageCustomValidator) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	metricstorage, ok := obj.(*telemetryv1beta1.MetricStorage)
	if !ok {
		return nil, fmt.Errorf("%w: expected a MetricStorage object but got %T", ErrUnexpectedObjectType, obj)
	}
	metricstoragelog.Info("Validation for MetricStorage upon deletion", "name", metricstorage.GetName())

	// Call the ValidateDelete method on the MetricStorage type
	return metricstorage.ValidateDelete()
}
