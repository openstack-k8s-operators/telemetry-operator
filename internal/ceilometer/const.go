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

	// CentralHCScript is the path to the central health check script
	CentralHCScript = "/var/lib/openstack/bin/centralhealth.py"
	// NotificationHCScript is the path to the notification health check script
	NotificationHCScript = "/var/lib/openstack/bin/notificationhealth.py"

	// ACConsumerFinalizer is added to AC secrets that Ceilometer is actively consuming
	ACConsumerFinalizer = "openstack.org/ceilometer-ac-consumer"
)
