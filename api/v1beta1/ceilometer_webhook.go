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
}

func (spec *CeilometerSpecCore) Default() {
	// NOTE: ApplicationCredentialSecret is NOT defaulted here.
	// AppCred is opt-in: only used when explicitly configured by the user.
}

var _ webhook.Validator = &Ceilometer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Ceilometer) ValidateCreate() (admission.Warnings, error) {
	ceilometerlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Ceilometer) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	ceilometerlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Ceilometer) ValidateDelete() (admission.Warnings, error) {
	ceilometerlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
