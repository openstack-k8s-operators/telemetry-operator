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
var cloudkittylog = logf.Log.WithName("cloudkitty-resource")

// SetupCloudKittyWebhookWithManager registers the webhook for CloudKitty in the manager.
func SetupCloudKittyWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&telemetryv1beta1.CloudKitty{}).
		WithValidator(&CloudKittyCustomValidator{}).
		WithDefaulter(&CloudKittyCustomDefaulter{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-telemetry-openstack-org-v1beta1-cloudkitty,mutating=true,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=cloudkitties,verbs=create;update,versions=v1beta1,name=mcloudkitty-v1beta1.kb.io,admissionReviewVersions=v1

// CloudKittyCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind CloudKitty when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type CloudKittyCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &CloudKittyCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind CloudKitty.
func (d *CloudKittyCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	cloudkitty, ok := obj.(*telemetryv1beta1.CloudKitty)

	if !ok {
		return fmt.Errorf("%w: expected an CloudKitty object but got %T", ErrUnexpectedObjectType, obj)
	}
	cloudkittylog.Info("Defaulting for CloudKitty", "name", cloudkitty.GetName())

	// Call the Default method on the CloudKitty type
	cloudkitty.Default()

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-telemetry-openstack-org-v1beta1-cloudkitty,mutating=false,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=cloudkitties,verbs=create;update,versions=v1beta1,name=vcloudkitty-v1beta1.kb.io,admissionReviewVersions=v1

// CloudKittyCustomValidator struct is responsible for validating the CloudKitty resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type CloudKittyCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &CloudKittyCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type CloudKitty.
func (v *CloudKittyCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cloudkitty, ok := obj.(*telemetryv1beta1.CloudKitty)
	if !ok {
		return nil, fmt.Errorf("%w: expected a CloudKitty object but got %T", ErrUnexpectedObjectType, obj)
	}
	cloudkittylog.Info("Validation for CloudKitty upon creation", "name", cloudkitty.GetName())

	// Call the ValidateCreate method on the CloudKitty type
	return cloudkitty.ValidateCreate()
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type CloudKitty.
func (v *CloudKittyCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	cloudkitty, ok := newObj.(*telemetryv1beta1.CloudKitty)
	if !ok {
		return nil, fmt.Errorf("%w: expected a CloudKitty object for the newObj but got %T", ErrUnexpectedObjectType, newObj)
	}
	cloudkittylog.Info("Validation for CloudKitty upon update", "name", cloudkitty.GetName())

	// Call the ValidateUpdate method on the CloudKitty type
	return cloudkitty.ValidateUpdate(oldObj)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type CloudKitty.
func (v *CloudKittyCustomValidator) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cloudkitty, ok := obj.(*telemetryv1beta1.CloudKitty)
	if !ok {
		return nil, fmt.Errorf("%w: expected a CloudKitty object but got %T", ErrUnexpectedObjectType, obj)
	}
	cloudkittylog.Info("Validation for CloudKitty upon deletion", "name", cloudkitty.GetName())

	// Call the ValidateDelete method on the CloudKitty type
	return cloudkitty.ValidateDelete()
}
