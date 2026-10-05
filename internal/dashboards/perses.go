/*
Copyright 2026.

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

// Package dashboards contains the builders for the observability dashboards
// managed by the telemetry-operator. Dashboards are delivered as
// perses.dev PersesDashboard objects in the same namespace as the
// MetricStorage CR that owns them.
package dashboards

import (
	persesv1alpha1 "github.com/perses/perses-operator/api/v1alpha1"
)

// Canonical dashboard names. These are the identifiers an administrator uses in
// MetricStorage.spec.disabledDashboards to disable an individual dashboard, and
// the names of the PersesDashboard objects created in the openstack namespace.
const (
	OpenstackCloudName          = "openstack-cloud"
	OpenstackNodeName           = "openstack-node"
	OpenstackNetworkName        = "openstack-openstack-network"
	OpenstackVMName             = "openstack-vm"
	OpenstackRabbitmqName       = "openstack-rabbitmq"
	OpenstackNetworkTrafficName = "openstack-network-traffic"
	OpenstackCeilometerIpmiName = "openstack-ceilometer-ipmi"
	OpenstackLightspeedName     = "openstack-lightspeed"
)

// PersesDashboardBuilder builds a PersesDashboard object bound to the given
// datasource, in the given namespace.
type PersesDashboardBuilder func(datasourceName, namespace string) *persesv1alpha1.PersesDashboard

// AllDashboardNames returns every dashboard name the operator knows about.
// It is the vocabulary accepted by MetricStorage.spec.disabledDashboards.
func AllDashboardNames() []string {
	return []string{
		OpenstackCloudName,
		OpenstackNodeName,
		OpenstackNetworkName,
		OpenstackVMName,
		OpenstackRabbitmqName,
		OpenstackNetworkTrafficName,
		OpenstackCeilometerIpmiName,
		OpenstackLightspeedName,
	}
}

// PersesDashboards returns the registry of PersesDashboard builders keyed by
// canonical dashboard name. Only dashboards whose Perses content has been
// implemented appear here; the remaining Grafana dashboards are converted and
// added separately.
func PersesDashboards() map[string]PersesDashboardBuilder {
	return map[string]PersesDashboardBuilder{
		OpenstackCloudName: OpenstackCloudPerses,
	}
}
