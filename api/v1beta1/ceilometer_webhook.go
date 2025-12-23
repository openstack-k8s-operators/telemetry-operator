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
	rabbitmqv1 "github.com/openstack-k8s-operators/infra-operator/apis/rabbitmq/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// CeilometerDefaults -
type CeilometerDefaults struct {
	CentralContainerImageURL        string
	NotificationContainerImageURL   string
	SgCoreContainerImageURL         string
	ComputeContainerImageURL        string
	IpmiContainerImageURL           string
	ProxyContainerImageURL          string
	KSMContainerImageURL            string
	MysqldExporterContainerImageURL string
}

var ceilometerDefaults CeilometerDefaults

// log is for logging in this package.
var ceilometerlog = logf.Log.WithName("ceilometer-resource")

// SetupCeilometerDefaults - initialize Ceilometer spec defaults for use with either internal or external webhooks
func SetupCeilometerDefaults(defaults CeilometerDefaults) {
	ceilometerDefaults = defaults
	ceilometerlog.Info("Ceilometer defaults initialized", "defaults", defaults)
}

var _ webhook.Defaulter = &Ceilometer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Ceilometer) Default() {
	ceilometerlog.Info("default", "name", r.Name)

	r.Spec.Default()
}

// Default - set defaults for this Ceilometer spec
func (spec *CeilometerSpec) Default() {
	if spec.CentralImage == "" {
		spec.CentralImage = ceilometerDefaults.CentralContainerImageURL
	}
	if spec.NotificationImage == "" {
		spec.NotificationImage = ceilometerDefaults.NotificationContainerImageURL
	}
	if spec.SgCoreImage == "" {
		spec.SgCoreImage = ceilometerDefaults.SgCoreContainerImageURL
	}
	if spec.ComputeImage == "" {
		spec.ComputeImage = ceilometerDefaults.ComputeContainerImageURL
	}
	if spec.IpmiImage == "" {
		spec.IpmiImage = ceilometerDefaults.IpmiContainerImageURL
	}
	if spec.ProxyImage == "" {
		spec.ProxyImage = ceilometerDefaults.ProxyContainerImageURL
	}
	if spec.KSMImage == "" {
		spec.KSMImage = ceilometerDefaults.KSMContainerImageURL
	}
	if spec.MysqldExporterImage == "" {
		spec.MysqldExporterImage = ceilometerDefaults.MysqldExporterContainerImageURL
	}

	spec.CeilometerSpecCore.Default()
}

// Default - set defaults for this CeilometerSpecCore. NOTE: this version is used by the OpenStackControlplane webhook
func (spec *CeilometerSpecCore) Default() {
	if spec.RabbitMqClusterName == "" {
		spec.RabbitMqClusterName = "rabbitmq"
	}
	rabbitmqv1.DefaultRabbitMqConfig(&spec.MessagingBus, spec.RabbitMqClusterName)
}

var _ webhook.Validator = &Ceilometer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Ceilometer) ValidateCreate() (admission.Warnings, error) {
	ceilometerlog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList
	basePath := field.NewPath("spec")

	if err := r.Spec.CeilometerSpecCore.ValidateCreate(basePath, r.Namespace); err != nil {
		allErrs = append(allErrs, err...)
	}

	if len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(
			schema.GroupKind{Group: "telemetry.openstack.org", Kind: "Ceilometer"},
			r.Name, allErrs)
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Ceilometer) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	ceilometerlog.Info("validate update", "name", r.Name)

	oldCeilometer, ok := old.(*Ceilometer)
	if !ok || oldCeilometer == nil {
		return nil, apierrors.NewInternalError(nil)
	}

	var allErrs field.ErrorList
	basePath := field.NewPath("spec")

	if err := r.Spec.CeilometerSpecCore.ValidateUpdate(oldCeilometer.Spec.CeilometerSpecCore, basePath, r.Namespace); err != nil {
		allErrs = append(allErrs, err...)
	}

	if len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(
			schema.GroupKind{Group: "telemetry.openstack.org", Kind: "Ceilometer"},
			r.Name, allErrs)
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Ceilometer) ValidateDelete() (admission.Warnings, error) {
	ceilometerlog.Info("validate delete", "name", r.Name)

	return nil, nil
}

// ValidateCreate - Exported function wrapping non-exported validate functions,
// this function can be called externally to validate a ceilometer spec.
func (spec *CeilometerSpecCore) ValidateCreate(basePath *field.Path, namespace string) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, spec.ValidateTopology(basePath, namespace)...)

	return allErrs
}

// ValidateUpdate - Exported function wrapping non-exported validate functions,
// this function can be called externally to validate a ceilometer spec.
func (spec *CeilometerSpecCore) ValidateUpdate(old CeilometerSpecCore, basePath *field.Path, namespace string) field.ErrorList {
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
