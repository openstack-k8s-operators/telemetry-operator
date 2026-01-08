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
	"fmt"

	rabbitmqv1 "github.com/openstack-k8s-operators/infra-operator/apis/rabbitmq/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
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

var _ webhook.Defaulter = &Autoscaling{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Autoscaling) Default() {
	autoscalinglog.Info("default", "name", r.Name)

	r.Spec.Default()
	r.Spec.Aodh.AodhCore.Default()
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

// Default - note only *Core* versions like this will have validations that are called from the
// Controlplane webhook
func (spec *AodhCore) Default() {
	if spec.RabbitMqClusterName == "" {
		spec.RabbitMqClusterName = "rabbitmq"
	}
	rabbitmqv1.DefaultRabbitMqConfig(&spec.MessagingBus, spec.RabbitMqClusterName)

	if spec.MemcachedInstance == "" {
		spec.MemcachedInstance = "memcached"
	}
}

// SetDefaultRouteAnnotations sets HAProxy timeout values of the route
// NOTE: it is used by ctlplane webhook on openstack-operator
func (spec *AutoscalingSpecCore) SetDefaultRouteAnnotations(annotations map[string]string) {
	const haProxyAnno = "haproxy.router.openshift.io/timeout"
	// Use a custom annotation to flag when the operator has set the default HAProxy timeout
	// With the annotation func determines when to overwrite existing HAProxy timeout with the APITimeout
	const aodhAnno = "api.aodh.openstack.org/timeout"

	valAodh, okAodh := annotations[aodhAnno]
	valHAProxy, okHAProxy := annotations[haProxyAnno]

	// Human operator set the HAProxy timeout manually
	if !okAodh && okHAProxy {
		return
	}

	// Human operator modified the HAProxy timeout manually without removing the Aodh flag
	if okAodh && okHAProxy && valAodh != valHAProxy {
		delete(annotations, aodhAnno)
		return
	}

	timeout := fmt.Sprintf("%ds", spec.Aodh.APITimeout)
	annotations[aodhAnno] = timeout
	annotations[haProxyAnno] = timeout
}

var _ webhook.Validator = &Autoscaling{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Autoscaling) ValidateCreate() (admission.Warnings, error) {
	autoscalinglog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList
	basePath := field.NewPath("spec")

	if err := r.Spec.Aodh.ValidateCreate(basePath.Child("aodh"), r.Namespace); err != nil {
		allErrs = append(allErrs, err...)
	}

	if len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(
			schema.GroupKind{Group: "telemetry.openstack.org", Kind: "Autoscaling"},
			r.Name, allErrs)
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Autoscaling) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	autoscalinglog.Info("validate update", "name", r.Name)

	oldAutoscaling, ok := old.(*Autoscaling)
	if !ok || oldAutoscaling == nil {
		return nil, apierrors.NewInternalError(nil)
	}

	var allErrs field.ErrorList
	basePath := field.NewPath("spec")

	if err := r.Spec.Aodh.ValidateUpdate(oldAutoscaling.Spec.Aodh.AodhCore, basePath.Child("aodh"), r.Namespace); err != nil {
		allErrs = append(allErrs, err...)
	}

	if len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(
			schema.GroupKind{Group: "telemetry.openstack.org", Kind: "Autoscaling"},
			r.Name, allErrs)
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Autoscaling) ValidateDelete() (admission.Warnings, error) {
	autoscalinglog.Info("validate delete", "name", r.Name)

	return nil, nil
}

// ValidateCreate - Exported function wrapping non-exported validate functions,
// this function can be called externally to validate an aodh spec.
func (spec *AodhCore) ValidateCreate(basePath *field.Path, namespace string) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, spec.ValidateTopology(basePath, namespace)...)

	return allErrs
}

// ValidateUpdate - Exported function wrapping non-exported validate functions,
// this function can be called externally to validate an aodh spec.
func (spec *AodhCore) ValidateUpdate(old AodhCore, basePath *field.Path, namespace string) field.ErrorList {
	var allErrs field.ErrorList

	// Reject changes to deprecated RabbitMqClusterName field
	if spec.RabbitMqClusterName != old.RabbitMqClusterName {
		allErrs = append(allErrs, field.Forbidden(
			basePath.Child("rabbitMqClusterName"),
			"rabbitMqClusterName is deprecated and cannot be changed. Please use messagingBus.cluster instead"))
	}

	allErrs = append(allErrs, spec.ValidateTopology(basePath, namespace)...)

	return allErrs
}
