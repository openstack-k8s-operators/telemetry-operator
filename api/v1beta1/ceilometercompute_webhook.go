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
)

// CeilometerComputeDefaults -
type CeilometerComputeDefaults struct {
	ComputeContainerImageURL      string
	ComputeInitContainerImageURL  string
}

var ceilometercomputeDefaults CeilometerComputeDefaults

// log is for logging in this package.
var ceilometercomputelog = logf.Log.WithName("ceilometercompute-resource")

// SetupCeilometerComputeDefaults - initialize CeilometerCompute spec defaults for use with either internal or external webhooks
func SetupCeilometerComputeDefaults(defaults CeilometerComputeDefaults) {
	ceilometercomputeDefaults = defaults
	ceilometercomputelog.Info("CeilometerCompute defaults initialized", "defaults", defaults)
}

// SetupWebhookWithManager - setups webhook with the adequate manager
func (r *CeilometerCompute) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-telemetry-openstack-org-v1beta1-ceilometercentral,mutating=true,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=ceilometercomputes,verbs=create;update,versions=v1beta1,name=mceilometercompute.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CeilometerCompute{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CeilometerCompute) Default() {
	ceilometercentrallog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// Default - set defaults for this CeilometerCompute spec
func (spec *CeilometerComputeSpec) Default() {
	if spec.ComputeImage == "" {
		spec.ComputeImage = ceilometercomputeDefaults.ComputeContainerImageURL
	}
	if spec.InitImage == "" {
		spec.InitImage = ceilometercomputeDefaults.ComputeInitContainerImageURL
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-telemetry-openstack-org-v1beta1-ceilometercompute,mutating=false,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=ceilometercomputes,verbs=create;update,versions=v1beta1,name=vceilometercompute.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CeilometerCompute{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CeilometerCompute) ValidateCreate() error {
	ceilometercomputelog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CeilometerCompute) ValidateUpdate(old runtime.Object) error {
	ceilometercomputelog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CeilometerCompute) ValidateDelete() error {
	ceilometercomputelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
