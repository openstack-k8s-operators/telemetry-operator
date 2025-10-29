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

package metricstorage

const (
	// RabbitMQPrometheusPortTLS is the port number for RabbitMQ
	// Prometheus metrics when TLS is enabled
	RabbitMQPrometheusPortTLS = "15691"

	// RabbitMQPrometheusPortNoTLS is the port number for RabbitMQ
	// Prometheus metrics when TLS is disabled
	RabbitMQPrometheusPortNoTLS = "15692"

	// PrometheusHost is the key for Prometheus host configuration
	PrometheusHost = "host"
	// PrometheusPort is the key for Prometheus port configuration
	PrometheusPort = "port"
	// PrometheusCaCertSecret is the key for Prometheus CA certificate secret
	PrometheusCaCertSecret = "ca_secret"
	// PrometheusCaCertKey is the key for Prometheus CA certificate key
	PrometheusCaCertKey = "ca_key"
)
