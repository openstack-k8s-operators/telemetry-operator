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

// InfraComputeDefaults -
type InfraComputeDefaults struct {
	NodeExporterContainerImageURL string
}

var infracomputeDefaults InfraComputeDefaults

// log is for logging in this package.
var infracomputelog = logf.Log.WithName("infracompute-resource")

// SetupInfraComputeDefaults - initialize InfraCompute spec defaults for use with either internal or external webhooks
func SetupInfraComputeDefaults(defaults InfraComputeDefaults) {
	infracomputeDefaults = defaults
	infracomputelog.Info("InfraCompute defaults initialized", "defaults", defaults)
}

// SetupWebhookWithManager - setups webhook with the adequate manager
func (r *InfraCompute) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-telemetry-openstack-org-v1beta1-infracompute,mutating=true,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=infracomputes,verbs=create;update,versions=v1beta1,name=minfracompute.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &InfraCompute{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *InfraCompute) Default() {
	infracomputelog.Info("default", "name", r.Name)

	r.Spec.Default()
}

// Default - set defaults for this InfraCompute spec
func (spec *InfraComputeSpec) Default() {
	if spec.NodeExporterImage == "" {
		spec.NodeExporterImage = infracomputeDefaults.NodeExporterContainerImageURL
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-telemetry-openstack-org-v1beta1-infracompute,mutating=false,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=infracomputes,verbs=create;update,versions=v1beta1,name=vinfracompute.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &InfraCompute{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *InfraCompute) ValidateCreate() error {
	infracomputelog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *InfraCompute) ValidateUpdate(old runtime.Object) error {
	infracomputelog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *InfraCompute) ValidateDelete() error {
	infracomputelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
