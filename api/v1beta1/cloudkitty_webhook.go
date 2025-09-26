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
	"fmt"
	"slices"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// CloudKittyDefaults -
type CloudKittyDefaults struct {
	APIContainerImageURL  string
	ProcContainerImageURL string
}

var cloudKittyDefaults CloudKittyDefaults

// log is for logging in this package.
var cloudKittyLog = logf.Log.WithName("cloudkitty-resource")

// SetupCloudKittyDefaults - initialize CloudKitty spec defaults for use with either internal or external webhooks
func SetupCloudKittyDefaults(defaults CloudKittyDefaults) {
	cloudKittyDefaults = defaults
	cloudKittyLog.Info("CloudKitty defaults initialized", "defaults", defaults)
}

// SetupWebhookWithManager - setups webhook with the adequate manager
func (r *CloudKitty) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-telemetry-openstack-org-v1beta1-cloudkitty,mutating=true,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=cloudkitties,verbs=create;update,versions=v1beta1,name=mcloudkitty.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CloudKitty{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CloudKitty) Default() {
	cloudKittyLog.Info("default", "name", r.Name)

	r.Spec.Default()
}

// Default - set defaults for this CloudKitty spec
func (spec *CloudKittySpec) Default() {
	if spec.CloudKittyAPI.ContainerImage == "" {
		spec.CloudKittyAPI.ContainerImage = cloudKittyDefaults.APIContainerImageURL
	}
	if spec.CloudKittyProc.ContainerImage == "" {
		spec.CloudKittyProc.ContainerImage = cloudKittyDefaults.ProcContainerImageURL
	}

}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-telemetry-openstack-org-v1beta1-cloudkitty,mutating=false,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=cloudkitties,verbs=create;update,versions=v1beta1,name=vcloudkitty.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CloudKitty{}

func (r *ObjectStorageSpec) Validate(basePath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	// NOTE: Having 0 schemas is allowed. LokiStack has a default for that
	for _, s := range r.Schemas {
		if s.EffectiveDate == "" {
			allErrs = append(
				allErrs,
				field.Invalid(
					basePath.Child("schemas").Child("effectiveDate"), "", "effectiveDate field should not be empty"),
			)
		}
		if s.Version == "" {
			allErrs = append(
				allErrs,
				field.Invalid(
					basePath.Child("schemas").Child("version"), "", "version field should not be empty"),
			)
		}
	}

	if r.Secret.Name == "" {
		allErrs = append(
			allErrs,
			field.Invalid(
				basePath.Child("secret").Child("name"), "", "name field should not be empty"),
		)
	}

	if r.Secret.Type == "" {
		allErrs = append(
			allErrs,
			field.Invalid(
				basePath.Child("secret").Child("type"), "", "type field should not be empty"),
		)
	}
	validTypes := []string{"azure", "gcs", "s3", "swift", "alibabacloud"}
	if !slices.Contains(validTypes, r.Secret.Type) {
		allErrs = append(
			allErrs,
			field.Invalid(
				basePath.Child("secret").Child("type"), r.Secret.Type, fmt.Sprintf("type field needs to be one of %s", validTypes)),
		)
	}

	if r.TLS != nil && r.TLS.CASpec.CA == "" {
		allErrs = append(
			allErrs,
			field.Invalid(
				basePath.Child("tls").Child("caName"), "", "caName field should not be empty"),
		)
	}
	return allErrs
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CloudKitty) ValidateCreate() (admission.Warnings, error) {
	cloudKittyLog.Info("validate create", "name", r.Name)

	allErrs := r.Spec.ValidateCreate(field.NewPath("spec"), r.Namespace)

	if len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(
			schema.GroupKind{Group: "cloudkitties.telemetry.openstack.org", Kind: "CloudKitty"},
			r.Name, allErrs)
	}

	return nil, nil
}

// ValidateCreate validates the CloudKittySpec during the webhook invocation.
func (r *CloudKittySpec) ValidateCreate(basePath *field.Path, namespace string) field.ErrorList {
	return r.CloudKittySpecBase.ValidateCreate(basePath, namespace)
}

// ValidateCreate validates the CloudKittySpecCore during the webhook invocation. It is
// expected to be called by the validation webhook in the higher level telemetry webhook
func (r *CloudKittySpecCore) ValidateCreate(basePath *field.Path, namespace string) field.ErrorList {
	return r.CloudKittySpecBase.ValidateCreate(basePath, namespace)
}

// ValidateCreate validates the CloudKittySpecBase during the webhook invocation.
func (r *CloudKittySpecBase) ValidateCreate(basePath *field.Path, namespace string) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, r.S3StorageConfig.Validate(basePath.Child("s3StorageConfig"))...)

	// TODO: Add other CK spec field validations as needed

	return allErrs
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CloudKitty) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {

	cloudKittyLog.Info("validate update", "name", r.Name)
	oldCloudKitty, ok := old.(*CloudKitty)
	if !ok || oldCloudKitty == nil {
		return nil, apierrors.NewInternalError(fmt.Errorf("unable to convert existing object"))
	}

	allErrs := r.Spec.ValidateUpdate(oldCloudKitty.Spec, field.NewPath("spec"), r.Namespace)

	if len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(
			schema.GroupKind{Group: "cloudkitties.telemetry.openstack.org", Kind: "CloudKitty"},
			r.Name, allErrs)
	}

	return nil, nil

}

// ValidateCreate validates the CloudKittySpec during the webhook invocation.
func (r *CloudKittySpec) ValidateUpdate(old CloudKittySpec, basePath *field.Path, namespace string) field.ErrorList {
	return r.CloudKittySpecBase.ValidateUpdate(old.CloudKittySpecBase, basePath, namespace)
}

// ValidateUpdate validates the CloudKittySpecCore during the webhook invocation. It is
// expected to be called by the validation webhook in the higher level telemetry webhook
func (r *CloudKittySpecCore) ValidateUpdate(old CloudKittySpecCore, basePath *field.Path, namespace string) field.ErrorList {
	return r.CloudKittySpecBase.ValidateUpdate(old.CloudKittySpecBase, basePath, namespace)
}

// ValidateCreate validates the CloudKittySpecBase during the webhook invocation.
func (r *CloudKittySpecBase) ValidateUpdate(old CloudKittySpecBase, basePath *field.Path, namespace string) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, r.S3StorageConfig.Validate(basePath.Child("s3StorageConfig"))...)

	// TODO: Add other CK spec field validations as needed

	return allErrs
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CloudKitty) ValidateDelete() (admission.Warnings, error) {
	cloudKittyLog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
