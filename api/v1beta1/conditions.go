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

	// MetricStorageReadyCondition Status=True condition which indicates if the MetricStorage is configured and operational
	MetricStorageReadyCondition condition.Type = "MetricStorageReady"

	// HeatReadyCondition Status=True condition which indicates if the Heat is configured and operational
	HeatReadyCondition condition.Type = "HeatReady"

	// MonitoringStackReadyCondition Status=True condition which indicates if the MonitoringStack is configured and operational
	MonitoringStackReadyCondition condition.Type = "MonitoringStackReady"

	// ScrapeConfigReadyCondition Status=True condition which indicates if the ScrapeConfig is configured and operational
	ScrapeConfigReadyCondition condition.Type = "ScrapeConfigReady"

	// PrometheusReadyCondition Status=True condition which indicates if the Prometheus watch is operational
	PrometheusReadyCondition condition.Type = "PrometheusReady"

	// LoggingReadyCondition Status=True condition which indicates if the Logging is configured and operational
	LoggingReadyCondition condition.Type = "LoggingReady"

	// LoggingCLONamespaceReadyCondition Status=True condition which indicates if the cluster-logging-operator namespace is created
	LoggingCLONamespaceReadyCondition condition.Type = "LoggingCLONamespaceReady"

	DashboardPrometheusRuleReadyCondition condition.Type = "DashboardPrometheusRuleReady"

	DashboardPluginReadyCondition condition.Type = "DashboardPluginReady"

	DashboardDatasourceReadyCondition condition.Type = "DashboardDatasourceReady"

	DashboardDefinitionReadyCondition condition.Type = "DashboardDefinitionReady"

	// KSMReadyCondition Status=True condition which indicates if the KSM is configured and operational
	KSMReadyCondition condition.Type = "KSMReady"
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

	//
	// MetricStorageReady condition messages
	//
	// MetricStorageReadyInitMessage
	MetricStorageReadyInitMessage = "MetricStorage not started"

	// MetricStorageReadyMessage
	MetricStorageReadyMessage = "MetricStorage completed"

	// MetricStorageReadyErrorMessage
	MetricStorageReadyErrorMessage = "MetricStorage error occured %s"

	// MetricStorageReadyRunningMessage
	MetricStorageReadyRunningMessage = "MetricStorage in progress"

	//
	// HeatReady condition messages
	//
	// HeatReadyInitMessage
	HeatReadyInitMessage = "Heat not started"

	// HeatReadyErrorMessage
	HeatReadyErrorMessage = "Heat error occured %s"

	// HeatReadyNotFoundMessage
	HeatReadyNotFoundMessage = "Heat has not been found"

	// HeatReadyUnreadyMessage
	HeatReadyUnreadyMessage = "Heat isn't ready yet"

	//
	// MonitoringStackReady condition messages
	//
	// MonitoringStackReadyInitMessage
	MonitoringStackReadyInitMessage = "MonitoringStack not started"

	// MonitoringStackUnableToOwnMessage
	MonitoringStackUnableToOwnMessage = "Error occured when trying to own: %s"

	// MonitoringStackReadyErrorMessage
	MonitoringStackReadyErrorMessage = "MonitoringStack error occured %s"

	// MonitoringStackReadyMisconfiguredMessage
	MonitoringStackReadyMisconfiguredMessage = "MonitoringStack isn't configured properly: %s"

	//
	// ScrapeConfigReady condition messages
	//
	// ScrapeConfigReadyInitMessage
	ScrapeConfigReadyInitMessage = "ScrapeConfig not started"

	// ScrapeConfigUnableToOwnMessage
	ScrapeConfigUnableToOwnMessage = "Error occured when trying to own %s"

	//
	// PrometheusReady condition messages
	//
	// PrometheusReadyInitMessage
	PrometheusReadyInitMessage = "Prometheus not ready"

	// PrometheusUnableToWatchMessage
	PrometheusUnableToWatchMessage = "Error occured when trying to watch %s"

	// PrometheusUnableToRemoveTLSMessage
	PrometheusUnableToRemoveTLSMessage = "Error occured when trying to remove TLS config: %s"

	//
	// LoggingReady condition messages
	//
	// LoggingReadyInitMessage
	LoggingReadyInitMessage = "Logging not started"

	// LoggingReadyMessage
	LoggingReadyMessage = "Logging completed"

	// LoggingReadyErrorMessage
	LoggingReadyErrorMessage = "Logging error occured %s"

	// LoggingReadyRunningMessage
	LoggingReadyRunningMessage = "Logging in progress"

	// LoggingCLONamespaceFailedMessage
	LoggingCLONamespaceFailedMessage = "CLO Namespace %s does not exist"

	DashboardsNotEnabledMessage = "Dashboarding was not enabled, so no actions required"

	DashboardPrometheusRuleReadyInitMessage   = "Dashboard PrometheusRule not started"
	DashboardPrometheusRuleUnableToOwnMessage = "Error occured when trying to own %s"

	DashboardPluginReadyInitMessage = "Dashboard Plugin not started"
	DashboardPluginFailedMessage    = "Error occured when trying to install the dashboard plugin: %s"

	DashboardDatasourceReadyInitMessage = "Dashboard Datasource not started"
	DashboardDatasourceFailedMessage    = "Error occured when trying to install the dashboard datasource: %s"

	DashboardDefinitionReadyInitMessage = "Dashboard Definition not started"
	DashboardDefinitionFailedMessage    = "Error occured when trying to install the dashboard definitions: %s"

	//
	// KSMReady condition messages
	//
	// KSMReadyInitMessage
	KSMReadyInitMessage = "KSM not started"

	// KSMReadyMessage
	KSMReadyMessage = "KSM completed"

	// KSMReadyErrorMessage
	KSMReadyErrorMessage = "KSM error occured %s"

	// KSMReadyRunningMessage
	KSMReadyRunningMessage = "KSM in progress"
)
