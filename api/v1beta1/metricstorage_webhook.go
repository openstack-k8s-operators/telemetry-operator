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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// MetricStorageDefaults -
type MetricStorageDefaults struct {
	AetosContainerImageURL string
}

var metricStorageDefaults MetricStorageDefaults

// log is for logging in this package.
var metricstoragelog = logf.Log.WithName("metricstorage-resource")

// SetupMetricStorageDefaults - initialize MetricStorage spec defaults for use with either internal or external webhooks
func SetupMetricStorageDefaults(defaults MetricStorageDefaults) {
	metricStorageDefaults = defaults
	metricstoragelog.Info("MetricStorage defaults initialized", "defaults", defaults)
}

// Default sets default values for the MetricStorage spec
func (r *MetricStorage) Default() {
	r.Spec.Default()
}

// Default - set defaults for the MetricStorage spec
func (spec *MetricStorageSpec) Default() {
	if spec.Aetos.Image == "" {
		spec.Aetos.Image = metricStorageDefaults.AetosContainerImageURL
	}

	spec.MetricStorageSpecBase.Default()
}

// Default - set defaults for the MetricStorage spec base. NOTE: this version is used by the OpenStackControlplane webhook
func (spec *MetricStorageSpecBase) Default() {
	if spec.MonitoringStack == nil && spec.CustomMonitoringStack == nil {
		spec.MonitoringStack = &MonitoringStack{}
		// Set the AlertingEnabled to true here as the empty value means false
		// and we don't have a way of distinguishing if the value was left
		// empty or set to false by the user in the spec.MonitoringStack.Default()
		spec.MonitoringStack.AlertingEnabled = true

		// NOTE: If we want to enable dashboards in the future by default, set
		//       it here like Alerting above
	}

	if spec.DataplaneNetwork != nil && *spec.DataplaneNetwork == "" {
		*spec.DataplaneNetwork = "ctlplane"
	}

	if spec.MonitoringStack != nil {
		spec.MonitoringStack.Default()
	}

	// NOTE: ApplicationCredentialSecret is NOT defaulted here.
	// AppCred is opt-in: only used when explicitly configured by the user.
}

// SetDefaultRouteAnnotations sets HAProxy timeout values of the route
// NOTE: it is used by ctlplane webhook on openstack-operator
func (spec *AetosCore) SetDefaultRouteAnnotations(annotations map[string]string) {
	const haProxyAnno = "haproxy.router.openshift.io/timeout"
	const aetosAnno = "api.aetos.openstack.org/timeout"

	if spec.APITimeout <= 0 {
		return
	}

	valAetos, okAetos := annotations[aetosAnno]
	valHAProxy, okHAProxy := annotations[haProxyAnno]

	// Human operator set the HAProxy timeout manually
	if !okAetos && okHAProxy {
		return
	}

	// Human operator modified the HAProxy timeout manually without removing the Aetos flag
	if okAetos && okHAProxy && valAetos != valHAProxy {
		delete(annotations, aetosAnno)
		return
	}

	timeout := fmt.Sprintf("%ds", spec.APITimeout)
	annotations[aetosAnno] = timeout
	annotations[haProxyAnno] = timeout
}

// Default - set defaults for the MonitoringStack field
func (ms *MonitoringStack) Default() {
	if ms.ScrapeInterval == "" {
		ms.ScrapeInterval = "30s"
	}
	ms.Storage.Default()
}

// Default - set defaults for the Storage field
func (storage *Storage) Default() {
	if storage.Strategy == "" {
		storage.Strategy = "persistent"
	}
	if storage.Retention == "" {
		storage.Retention = "24h"
	}
	if storage.Strategy == "persistent" {
		if storage.Persistent == nil {
			storage.Persistent = &PersistentStorage{}
		}
		storage.Persistent.Default()
	} else {
		storage.Persistent = nil
	}
}

// Default - set defaults for the PersistentStorage field
func (ps *PersistentStorage) Default() {
	if ps.PvcStorageRequest == "" {
		ps.PvcStorageRequest = "20G"
	}
}

// ValidateCreate validates the MetricStorage resource on creation
func (r *MetricStorage) ValidateCreate() (admission.Warnings, error) {
	metricstoragelog.Info("validate create", "name", r.Name)

	basePath := field.NewPath("spec").Child("aetos")
	errs := r.Spec.Aetos.AetosCore.ValidateCreate(basePath)
	if len(errs) != 0 {
		return nil, apierrors.NewInvalid(
			schema.GroupKind{Group: "telemetry.openstack.org", Kind: "MetricStorage"},
			r.Name, errs)
	}
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *MetricStorage) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	metricstoragelog.Info("validate update", "name", r.Name)

	basePath := field.NewPath("spec").Child("aetos")
	errs := r.Spec.Aetos.AetosCore.ValidateCreate(basePath)
	if len(errs) != 0 {
		return nil, apierrors.NewInvalid(
			schema.GroupKind{Group: "telemetry.openstack.org", Kind: "MetricStorage"},
			r.Name, errs)
	}
	return nil, nil
}

// ValidateCreate validates the AetosCore override spec
func (spec *AetosCore) ValidateCreate(basePath *field.Path) field.ErrorList {
	return validateServicePortNames(basePath.Child("override").Child("service"), spec.Override.Service)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *MetricStorage) ValidateDelete() (admission.Warnings, error) {
	metricstoragelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
