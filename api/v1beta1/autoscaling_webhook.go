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
	common_webhook "github.com/openstack-k8s-operators/lib-common/modules/common/webhook"
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
	// NOTE: ApplicationCredentialSecret is NOT defaulted here.
	// AppCred is opt-in: only used when explicitly configured by the user.
}

// getDeprecatedFields returns the centralized list of deprecated fields for AodhCore
func (spec *AodhCore) getDeprecatedFields(old *AodhCore) []common_webhook.DeprecatedFieldUpdate {
	deprecatedFields := []common_webhook.DeprecatedFieldUpdate{
		{
			DeprecatedFieldName: "rabbitMqClusterName",
			NewFieldPath:        []string{"messagingBus", "cluster"},
			NewDeprecatedValue:  &spec.RabbitMqClusterName,
			NewValue:            &spec.MessagingBus.Cluster,
		},
	}

	// If old spec is provided (UPDATE operation), add old values
	if old != nil {
		deprecatedFields[0].OldDeprecatedValue = &old.RabbitMqClusterName
	}

	return deprecatedFields
}

// validateDeprecatedFieldsCreate validates deprecated fields during CREATE operations
func (spec *AodhCore) validateDeprecatedFieldsCreate(basePath *field.Path) ([]string, field.ErrorList) {
	// Get deprecated fields list (without old values for CREATE)
	deprecatedFieldsUpdate := spec.getDeprecatedFields(nil)

	// Convert to DeprecatedField list for CREATE validation
	deprecatedFields := make([]common_webhook.DeprecatedField, len(deprecatedFieldsUpdate))
	for i, df := range deprecatedFieldsUpdate {
		deprecatedFields[i] = common_webhook.DeprecatedField{
			DeprecatedFieldName: df.DeprecatedFieldName,
			NewFieldPath:        df.NewFieldPath,
			DeprecatedValue:     df.NewDeprecatedValue,
			NewValue:            df.NewValue,
		}
	}

	return common_webhook.ValidateDeprecatedFieldsCreate(deprecatedFields, basePath), nil
}

// validateDeprecatedFieldsUpdate validates deprecated fields during UPDATE operations
func (spec *AodhCore) validateDeprecatedFieldsUpdate(old AodhCore, basePath *field.Path) ([]string, field.ErrorList) {
	// Get deprecated fields list with old values
	deprecatedFields := spec.getDeprecatedFields(&old)
	return common_webhook.ValidateDeprecatedFieldsUpdate(deprecatedFields, basePath)
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

	// Validate deprecated fields using shared helper
	_, errs := spec.validateDeprecatedFieldsCreate(basePath)
	allErrs = append(allErrs, errs...)

	allErrs = append(allErrs, spec.ValidateTopology(basePath, namespace)...)

	return allErrs
}

// ValidateUpdate - Exported function wrapping non-exported validate functions,
// this function can be called externally to validate an aodh spec.
func (spec *AodhCore) ValidateUpdate(old AodhCore, basePath *field.Path, namespace string) field.ErrorList {
	var allErrs field.ErrorList

	// Validate deprecated fields using shared helper
	_, errs := spec.validateDeprecatedFieldsUpdate(old, basePath)
	allErrs = append(allErrs, errs...)

	allErrs = append(allErrs, spec.ValidateTopology(basePath, namespace)...)

	return allErrs
}
