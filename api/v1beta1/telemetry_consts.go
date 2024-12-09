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

package v1beta1

import "time"

const (
	// DefaultServiceName -
	DefaultServiceName = "metric-storage"
	// DefaultPvcStorageRequest -
	DefaultPvcStorageRequest = "20G"
	// DefaultPrometheusPort -
	DefaultPrometheusPort = 9090
	// DefaultNodeExporterPort -
	DefaultNodeExporterPort = 9100
	// DefaultScrapeInterval -
	DefaultScrapeInterval = "30s"
	// PauseBetweenWatchAttempts -
	PauseBetweenWatchAttempts = time.Duration(60) * time.Second
	// DefaultKeplerPort -
	DefaultKeplerPort = 8888
	// Dataplane power monitoring service name
	TelemetryPowerMonitoring = "telemetry-power-monitoring"
)

// PrometheusReplicas -
var PrometheusReplicas int32 = 1
