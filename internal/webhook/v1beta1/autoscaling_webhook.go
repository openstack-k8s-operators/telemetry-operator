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

// Package v1beta1 contains webhook implementations for telemetry v1beta1 API resources.
package v1beta1

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	telemetryv1beta1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
)

var (
	// ErrUnexpectedObjectType is returned when an unexpected object type is received by a webhook.
	ErrUnexpectedObjectType = errors.New("unexpected object type")
)

// nolint:unused
// log is for logging in this package.
var autoscalinglog = logf.Log.WithName("autoscaling-resource")

// SetupAutoscalingWebhookWithManager registers the webhook for Autoscaling in the manager.
func SetupAutoscalingWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&telemetryv1beta1.Autoscaling{}).
		WithValidator(&AutoscalingCustomValidator{}).
		WithDefaulter(&AutoscalingCustomDefaulter{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-telemetry-openstack-org-v1beta1-autoscaling,mutating=true,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=autoscalings,verbs=create;update,versions=v1beta1,name=mautoscaling-v1beta1.kb.io,admissionReviewVersions=v1

// AutoscalingCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Autoscaling when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type AutoscalingCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &AutoscalingCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Autoscaling.
func (d *AutoscalingCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	autoscaling, ok := obj.(*telemetryv1beta1.Autoscaling)

	if !ok {
		return fmt.Errorf("%w: expected an Autoscaling object but got %T", ErrUnexpectedObjectType, obj)
	}
	autoscalinglog.Info("Defaulting for Autoscaling", "name", autoscaling.GetName())

	// Call the Default method on the Autoscaling type
	autoscaling.Default()

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-telemetry-openstack-org-v1beta1-autoscaling,mutating=false,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=autoscalings,verbs=create;update,versions=v1beta1,name=vautoscaling-v1beta1.kb.io,admissionReviewVersions=v1

// AutoscalingCustomValidator struct is responsible for validating the Autoscaling resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type AutoscalingCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &AutoscalingCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Autoscaling.
func (v *AutoscalingCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	autoscaling, ok := obj.(*telemetryv1beta1.Autoscaling)
	if !ok {
		return nil, fmt.Errorf("%w: expected a Autoscaling object but got %T", ErrUnexpectedObjectType, obj)
	}
	autoscalinglog.Info("Validation for Autoscaling upon creation", "name", autoscaling.GetName())

	// Call the ValidateCreate method on the Autoscaling type
	return autoscaling.ValidateCreate()
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Autoscaling.
func (v *AutoscalingCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	autoscaling, ok := newObj.(*telemetryv1beta1.Autoscaling)
	if !ok {
		return nil, fmt.Errorf("%w: expected a Autoscaling object for the newObj but got %T", ErrUnexpectedObjectType, newObj)
	}
	autoscalinglog.Info("Validation for Autoscaling upon update", "name", autoscaling.GetName())

	// Call the ValidateUpdate method on the Autoscaling type
	return autoscaling.ValidateUpdate(oldObj)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Autoscaling.
func (v *AutoscalingCustomValidator) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	autoscaling, ok := obj.(*telemetryv1beta1.Autoscaling)
	if !ok {
		return nil, fmt.Errorf("%w: expected a Autoscaling object but got %T", ErrUnexpectedObjectType, obj)
	}
	autoscalinglog.Info("Validation for Autoscaling upon deletion", "name", autoscaling.GetName())

	// Call the ValidateDelete method on the Autoscaling type
	return autoscaling.ValidateDelete()
}
