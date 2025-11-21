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
var telemetrylog = logf.Log.WithName("telemetry-resource")

// SetupTelemetryWebhookWithManager registers the webhook for Telemetry in the manager.
func SetupTelemetryWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&telemetryv1beta1.Telemetry{}).
		WithValidator(&TelemetryCustomValidator{}).
		WithDefaulter(&TelemetryCustomDefaulter{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-telemetry-openstack-org-v1beta1-telemetry,mutating=true,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=telemetries,verbs=create;update,versions=v1beta1,name=mtelemetry-v1beta1.kb.io,admissionReviewVersions=v1

// TelemetryCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Telemetry when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type TelemetryCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &TelemetryCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Telemetry.
func (d *TelemetryCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	telemetry, ok := obj.(*telemetryv1beta1.Telemetry)

	if !ok {
		return fmt.Errorf("%w: expected an Telemetry object but got %T", ErrUnexpectedObjectType, obj)
	}
	telemetrylog.Info("Defaulting for Telemetry", "name", telemetry.GetName())

	// Call the Default method on the Telemetry type
	telemetry.Default()

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-telemetry-openstack-org-v1beta1-telemetry,mutating=false,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=telemetries,verbs=create;update,versions=v1beta1,name=vtelemetry-v1beta1.kb.io,admissionReviewVersions=v1

// TelemetryCustomValidator struct is responsible for validating the Telemetry resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type TelemetryCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &TelemetryCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Telemetry.
func (v *TelemetryCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	telemetry, ok := obj.(*telemetryv1beta1.Telemetry)
	if !ok {
		return nil, fmt.Errorf("%w: expected a Telemetry object but got %T", ErrUnexpectedObjectType, obj)
	}
	telemetrylog.Info("Validation for Telemetry upon creation", "name", telemetry.GetName())

	// Call the ValidateCreate method on the Telemetry type
	return telemetry.ValidateCreate()
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Telemetry.
func (v *TelemetryCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	telemetry, ok := newObj.(*telemetryv1beta1.Telemetry)
	if !ok {
		return nil, fmt.Errorf("%w: expected a Telemetry object for the newObj but got %T", ErrUnexpectedObjectType, newObj)
	}
	telemetrylog.Info("Validation for Telemetry upon update", "name", telemetry.GetName())

	// Call the ValidateUpdate method on the Telemetry type
	return telemetry.ValidateUpdate(oldObj)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Telemetry.
func (v *TelemetryCustomValidator) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	telemetry, ok := obj.(*telemetryv1beta1.Telemetry)
	if !ok {
		return nil, fmt.Errorf("%w: expected a Telemetry object but got %T", ErrUnexpectedObjectType, obj)
	}
	telemetrylog.Info("Validation for Telemetry upon deletion", "name", telemetry.GetName())

	// Call the ValidateDelete method on the Telemetry type
	return telemetry.ValidateDelete()
}
