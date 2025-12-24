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

package autoscaling

import (
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
)

const (
	// ServiceName -
	ServiceName = "aodh"

	// ServiceType -
	ServiceType = "alarming"

	// PrometheusRetention -
	PrometheusRetention = "5h"

	// DatabaseName - Name of the database used in CREATE DATABASE statement
	DatabaseName = "aodh"

	// DatabaseCRName - Name of the MariaDBDatabase CR
	DatabaseCRName = "aodh"

	// DatabaseUsernamePrefix - used by EnsureMariaDBAccount when a new username
	// is to be generated, e.g. "aodh_e5a4", "aodh_78bc", etc
	DatabaseUsernamePrefix = "aodh"

	// AodhAPIPort -
	AodhAPIPort = 8042

	// CustomPrometheusCaCertFolderPath -
	CustomPrometheusCaCertFolderPath = "/etc/pki/ca-trust/extracted/pem/"

	// PrometheusEndpointSecret is the name of the secret containing Prometheus endpoint configuration
	PrometheusEndpointSecret = "metric-storage-prometheus-endpoint"

	// AodhUserID -
	AodhUserID = 42402

	// AutoscalingNotificationBusReadyCondition Status=True condition indicates if the RabbitMQ NotificationsBus TransportURL is configured
	AutoscalingNotificationBusReadyCondition condition.Type = "AutoscalingNotificationBusReady"

	// AutoscalingNotificationBusReadyMessage
	AutoscalingNotificationBusReadyMessage = "NotificationsBus TransportURL successfully created"

	// AutoscalingNotificationBusReadyRunningMessage
	AutoscalingNotificationBusReadyRunningMessage = "NotificationsBus TransportURL creation in progress"

	// AutoscalingNotificationBusReadyErrorMessage
	AutoscalingNotificationBusReadyErrorMessage = "NotificationsBus TransportURL error occured %s"
)

// PrometheusReplicas -
var PrometheusReplicas int32 = 1
