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

// Package ceilometer provides functionality for managing OpenStack Ceilometer telemetry components
package ceilometer

import (
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
)

const (
	// ServiceName -
	ServiceName = "ceilometer"
	// ComputeServiceName -
	ComputeServiceName = "ceilometer-compute"
	// IpmiServiceName -
	IpmiServiceName = "ceilometer-ipmi"
	// ServiceType -
	ServiceType = "Ceilometer"

	// CeilometerPrometheusPort -
	CeilometerPrometheusPort int = 3000

	// KollaConfigCentral -
	KollaConfigCentral = "/var/lib/config-data/merged/config-central.json"

	// KollaConfigNotification -
	KollaConfigNotification = "/var/lib/config-data/merged/config-notification.json"

	// CeilometerUserID -
	CeilometerUserID = 42405

	// CeilometerNotificationBusReadyCondition Status=True condition indicates if the RabbitMQ NotificationsBus TransportURL is configured
	CeilometerNotificationBusReadyCondition condition.Type = "CeilometerNotificationBusReady"

	// CeilometerNotificationBusReadyMessage
	CeilometerNotificationBusReadyMessage = "NotificationsBus TransportURL successfully created"

	// CeilometerNotificationBusReadyRunningMessage
	CeilometerNotificationBusReadyRunningMessage = "NotificationsBus TransportURL creation in progress"

	// CeilometerNotificationBusReadyErrorMessage
	CeilometerNotificationBusReadyErrorMessage = "NotificationsBus TransportURL error occured %s"
)
