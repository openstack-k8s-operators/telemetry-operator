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

import (
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
)

// Telemetry Condition Types used by API objects.
const (
	// TelemetryRabbitMqTransportURLReadyCondition Status=True condition which indicates if the RabbitMQ TransportURLUrl is ready
	TelemetryRabbitMqTransportURLReadyCondition condition.Type = "TelemetryRabbitMqTransportURLReady"

	// CeilometerCentralReadyCondition Status=True condition which indicates if the CeilometerCentral is configured and operational
	CeilometerCentralReadyCondition condition.Type = "CeilometerCentralReady"

	// CeilometerComputeReadyCondition Status=True condition which indicates if the CeilometerCompute is configured and operational
	CeilometerComputeReadyCondition condition.Type = "CeilometerComputeReady"
)

// Telemetry Reasons used by API objects.
const ()

// Common Messages used by API objects.
const (
	//
	// TelemetryRabbitMqTransportURLReady condition messages
	//
	// TelemetryRabbitMqTransportURLReadyInitMessage
	TelemetryRabbitMqTransportURLReadyInitMessage = "TelemetryRabbitMqTransportURL not started"

	// TelemetryRabbitMqTransportURLReadyRunningMessage
	TelemetryRabbitMqTransportURLReadyRunningMessage = "TelemetryRabbitMqTransportURL creation in progress"

	// TelemetryRabbitMqTransportURLReadyMessage
	TelemetryRabbitMqTransportURLReadyMessage = "TelemetryRabbitMqTransportURL successfully created"

	// TelemetryRabbitMqTransportURLReadyErrorMessage
	TelemetryRabbitMqTransportURLReadyErrorMessage = "TelemetryRabbitMqTransportURL error occured %s"

	//
	// CeilometerCentralReady condition messages
	//
	// CeilometerCentralReadyInitMessage
	CeilometerCentralReadyInitMessage = "CeilometerCentral not started"

	// CeilometerCentralReadyErrorMessage
	CeilometerCentralReadyErrorMessage = "CeilometerCentral error occured %s"

	//
	// CeilometerComputeReady condition messages
	//
	// CeilometerComputeReadyInitMessage
	CeilometerComputeReadyInitMessage = "CeilometerCompute not started"

	// CeilometerComputeReadyErrorMessage
	CeilometerComputeReadyErrorMessage = "CeilometerCompute error occured %s"
)
