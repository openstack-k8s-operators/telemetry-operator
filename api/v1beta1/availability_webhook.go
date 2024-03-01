/*
Copyright 2023.

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

// AvailabilityDefaults -
type AvailabilityDefaults struct {
	KSMImageURL string
}

var availabilityDefaults AvailabilityDefaults

// log is for logging in this package.
var availabilitylog = logf.Log.WithName("availability-resource")

// SetupAvailabilityDefaults - initialize Availability spec defaults for use with either internal or external webhooks
func SetupAvailabilityDefaults(defaults AvailabilityDefaults) {
	availabilityDefaults = defaults
	availabilitylog.Info("Availability defaults initialized", "defaults", defaults)
}

// SetupWebhookWithManager - setups webhook with the adequate manager
func (r *Availability) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-telemetry-openstack-org-v1beta1-availabiltiy,mutating=true,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=availability,verbs=create;update,versions=v1beta1,name=mavailability.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Availability{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Availability) Default() {
	availabilitylog.Info("default", "name", r.Name)

	r.Spec.Default()
}

// Default - set defaults for this Availability spec
func (spec *AvailabilitySpec) Default() {
	if spec.KSMImage == "" {
		spec.KSMImage = availabilityDefaults.KSMImageURL
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-telemetry-openstack-org-v1beta1-availability,mutating=false,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=availabilities,verbs=create;update,versions=v1beta1,name=vavailability.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Availability{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Availability) ValidateCreate() (admission.Warnings, error) {
	availabilitylog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Availability) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	availabilitylog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Availability) ValidateDelete() (admission.Warnings, error) {
	availabilitylog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
