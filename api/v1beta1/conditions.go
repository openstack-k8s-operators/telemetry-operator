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
	// CeilometerReadyCondition Status=True condition which indicates if the Ceilometer is configured and operational
	CeilometerReadyCondition condition.Type = "CeilometerReady"

	// AutoscalingReadyCondition Status=True condition which indicates if the Autoscaling is configured and operational
	AutoscalingReadyCondition condition.Type = "AutoscalingReady"

	// HeatReadyCondition Status=True condition which indicates if the Heat is configured and operational
	HeatReadyCondition condition.Type = "HeatReady"

	// PrometheusReadyCondition Status=True condition which indicates if the Prometheus is configured and operational
	PrometheusReadyCondition condition.Type = "PrometheusReady"
)

// Telemetry Reasons used by API objects.
const ()

// Common Messages used by API objects.
const (
	//
	// CeilometerReady condition messages
	//
	// CeilometerReadyInitMessage
	CeilometerReadyInitMessage = "Ceilometer not started"

	// CeilometerReadyMessage
	CeilometerReadyMessage = "Ceilometer completed"

	// CeilometerReadyErrorMessage
	CeilometerReadyErrorMessage = "Ceilometer error occured %s"

	// CeilometerReadyRunningMessage
	CeilometerReadyRunningMessage = "Ceilometer in progress"

	//
	// AutoscalingReady condition messages
	//
	// AutoscalingReadyInitMessage
	AutoscalingReadyInitMessage = "Autoscaling not started"

	// AutoscalingReadyMessage
	AutoscalingReadyMessage = "Autoscaling completed"

	// AutoscalingReadyErrorMessage
	AutoscalingReadyErrorMessage = "Autoscaling error occured %s"

	// AutoscalingReadyRunningMessage
	AutoscalingReadyRunningMessage = "Autoscaling in progress"

	// HeatReadyInitMessage
	HeatReadyInitMessage = "Heat not started"

	// HeatReadyErrorMessage
	HeatReadyErrorMessage = "Heat error occured %s"

	// HeatReadyNotFoundMessage
	HeatReadyNotFoundMessage = "Heat has not been found"

	// HeatReadyUnreadyMessage
	HeatReadyUnreadyMessage = "Heat isn't ready yet"

	// PrometheusReadyInitMessage
	PrometheusReadyInitMessage = "Prometheus not started"

	// PrometheusReadyErrorMessage
	PrometheusReadyErrorMessage = "Prometheus error occured %s"

	// PrometheusReadyConfigurationMissingMessage
	PrometheusReadyConfigurationMissingMessage = "deployPrometheus is false and either port or host isn't set"
)
