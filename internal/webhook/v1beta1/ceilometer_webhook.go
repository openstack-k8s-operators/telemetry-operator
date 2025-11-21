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
var ceilometerlog = logf.Log.WithName("ceilometer-resource")

// SetupCeilometerWebhookWithManager registers the webhook for Ceilometer in the manager.
func SetupCeilometerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&telemetryv1beta1.Ceilometer{}).
		WithValidator(&CeilometerCustomValidator{}).
		WithDefaulter(&CeilometerCustomDefaulter{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-telemetry-openstack-org-v1beta1-ceilometer,mutating=true,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=ceilometers,verbs=create;update,versions=v1beta1,name=mceilometer-v1beta1.kb.io,admissionReviewVersions=v1

// CeilometerCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Ceilometer when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type CeilometerCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &CeilometerCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Ceilometer.
func (d *CeilometerCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	ceilometer, ok := obj.(*telemetryv1beta1.Ceilometer)

	if !ok {
		return fmt.Errorf("%w: expected an Ceilometer object but got %T", ErrUnexpectedObjectType, obj)
	}
	ceilometerlog.Info("Defaulting for Ceilometer", "name", ceilometer.GetName())

	// Call the Default method on the Ceilometer type
	ceilometer.Default()

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-telemetry-openstack-org-v1beta1-ceilometer,mutating=false,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=ceilometers,verbs=create;update,versions=v1beta1,name=vceilometer-v1beta1.kb.io,admissionReviewVersions=v1

// CeilometerCustomValidator struct is responsible for validating the Ceilometer resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type CeilometerCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &CeilometerCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Ceilometer.
func (v *CeilometerCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	ceilometer, ok := obj.(*telemetryv1beta1.Ceilometer)
	if !ok {
		return nil, fmt.Errorf("%w: expected a Ceilometer object but got %T", ErrUnexpectedObjectType, obj)
	}
	ceilometerlog.Info("Validation for Ceilometer upon creation", "name", ceilometer.GetName())

	// Call the ValidateCreate method on the Ceilometer type
	return ceilometer.ValidateCreate()
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Ceilometer.
func (v *CeilometerCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ceilometer, ok := newObj.(*telemetryv1beta1.Ceilometer)
	if !ok {
		return nil, fmt.Errorf("%w: expected a Ceilometer object for the newObj but got %T", ErrUnexpectedObjectType, newObj)
	}
	ceilometerlog.Info("Validation for Ceilometer upon update", "name", ceilometer.GetName())

	// Call the ValidateUpdate method on the Ceilometer type
	return ceilometer.ValidateUpdate(oldObj)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Ceilometer.
func (v *CeilometerCustomValidator) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	ceilometer, ok := obj.(*telemetryv1beta1.Ceilometer)
	if !ok {
		return nil, fmt.Errorf("%w: expected a Ceilometer object but got %T", ErrUnexpectedObjectType, obj)
	}
	ceilometerlog.Info("Validation for Ceilometer upon deletion", "name", ceilometer.GetName())

	// Call the ValidateDelete method on the Ceilometer type
	return ceilometer.ValidateDelete()
}
