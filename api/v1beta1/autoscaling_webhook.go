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

// AutoscalingDefaults -
type AutoscalingDefaults struct {
	AodhAPIContainerImageURL       string
	AodhEvaluatorContainerImageURL string
	AodhNotifierContainerImageURL  string
	AodhListenerContainerImageURL  string
}

var autoscalingDefaults AutoscalingDefaults

// log is for logging in this package.
var autoscalinglog = logf.Log.WithName("autoscaling-resource")

// SetupAutoscalingDefaults - initialize Autoscaling spec defaults for use with either internal or external webhooks
func SetupAutoscalingDefaults(defaults AutoscalingDefaults) {
	autoscalingDefaults = defaults
	autoscalinglog.Info("Autoscaling defaults initialized", "defaults", defaults)
}

// SetupWebhookWithManager - setups webhook with the adequate manager
func (r *Autoscaling) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-telemetry-openstack-org-v1beta1-autoscaling,mutating=true,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=autoscalings,verbs=create;update,versions=v1beta1,name=mautoscaling.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Autoscaling{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Autoscaling) Default() {
	autoscalinglog.Info("default", "name", r.Name)

	r.Spec.Default()
}

// Default - set defaults for this Autoscaling spec
func (spec *AutoscalingSpec) Default() {
	if spec.Aodh.APIImage == "" {
		spec.Aodh.APIImage = autoscalingDefaults.AodhAPIContainerImageURL
	}
	if spec.Aodh.EvaluatorImage == "" {
		spec.Aodh.EvaluatorImage = autoscalingDefaults.AodhEvaluatorContainerImageURL
	}
	if spec.Aodh.NotifierImage == "" {
		spec.Aodh.NotifierImage = autoscalingDefaults.AodhNotifierContainerImageURL
	}
	if spec.Aodh.ListenerImage == "" {
		spec.Aodh.ListenerImage = autoscalingDefaults.AodhListenerContainerImageURL
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-telemetry-openstack-org-v1beta1-autoscaling,mutating=false,failurePolicy=fail,sideEffects=None,groups=telemetry.openstack.org,resources=autoscalings,verbs=create;update,versions=v1beta1,name=vautoscaling.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Autoscaling{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Autoscaling) ValidateCreate() (admission.Warnings, error) {
	autoscalinglog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Autoscaling) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	autoscalinglog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Autoscaling) ValidateDelete() (admission.Warnings, error) {
	autoscalinglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
