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

package dashboards

import (
	persesv1alpha1 "github.com/perses/perses-operator/api/v1alpha1"
)

// openstackCeilometerIpmiJSON is the "OpenStack / Compute Power Monitoring /
// Ceilometer IPMI" PersesDashboard definition, in the perses.dev/v1alpha2
// layout (content under spec.config). It is only deployed when at least one
// dataplane nodeset runs the telemetry-power-monitoring service.
//
// This is a minimal placeholder that establishes the object shape, ownership and
// datasource wiring. The full panel/layout content is converted from the legacy
// Grafana dashboard and delivered separately.
const openstackCeilometerIpmiJSON = `{
  "apiVersion": "perses.dev/v1alpha2",
  "kind": "PersesDashboard",
  "metadata": {
    "name": "openstack-ceilometer-ipmi"
  },
  "spec": {
    "config": {
      "display": {
        "name": "OpenStack / Compute Power Monitoring / Ceilometer IPMI"
      },
      "duration": "30m",
      "panels": {},
      "layouts": []
    }
  }
}`

// OpenstackCeilometerIpmiPerses builds the "OpenStack / Compute Power
// Monitoring / Ceilometer IPMI" PersesDashboard from its JSON definition, in the
// given namespace.
func OpenstackCeilometerIpmiPerses(_, namespace string) (*persesv1alpha1.PersesDashboard, error) {
	return dashboardFromJSON(OpenstackCeilometerIpmiName, openstackCeilometerIpmiJSON, namespace)
}
