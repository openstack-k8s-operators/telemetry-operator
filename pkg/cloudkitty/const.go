/*

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

package cloudkitty

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// ServiceName -
	ServiceName = "cloudkitty"
	// ServiceType -
	ServiceType = "rating"
	// DatabaseName -
	DatabaseName = "cloudkitty"

	// DefaultsConfigFileName -
	DefaultsConfigFileName = "cloudkitty.conf"
	// ServiceConfigFileName -
	ServiceConfigFileName = "01-service-defaults.conf"
	// CustomConfigFileName -
	CustomConfigFileName = "02-global-custom.conf"
	// CustomServiceConfigFileName -
	CustomServiceConfigFileName = "03-service-custom.conf"
	// CustomServiceConfigSecretsFileName -
	CustomServiceConfigSecretsFileName = "04-service-custom-secrets.conf"
	// MyCnfFileName -
	MyCnfFileName = "my.cnf"

	// CloudKittyPublicPort -
	CloudKittyPublicPort int32 = 8889
	// CloudKittyInternalPort -
	CloudKittyInternalPort int32 = 8889

	ShortDuration  = time.Duration(5) * time.Second
	NormalDuration = time.Duration(10) * time.Second

	// PrometheusEndpointSecret - The name of the secret that contains the Prometheus endpoint configuration.
	PrometheusEndpointSecret = "metric-storage-prometheus-endpoint"
)

var ResultRequeue = ctrl.Result{RequeueAfter: NormalDuration}
