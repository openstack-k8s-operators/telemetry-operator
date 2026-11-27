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

package controller

import (
	"testing"

	"github.com/onsi/gomega"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	"github.com/openstack-k8s-operators/telemetry-operator/internal/dashboards"
)

// TestUnknownDisabledDashboards checks the names in spec.disabledDashboards
// that disable no dashboard are reported, sorted, and that the names of the
// dashboards the operator manages are not.
func TestUnknownDisabledDashboards(t *testing.T) {
	builders := dashboards.PersesDashboards()
	for _, tc := range []struct {
		name     string
		disabled []string
		want     []string
	}{
		{
			name:     "none disabled",
			disabled: nil,
			want:     nil,
		},
		{
			name:     "all known",
			disabled: []string{dashboards.OpenstackCloudName, dashboards.OpenstackRabbitmqName},
			want:     nil,
		},
		{
			name: "typo and dropped legacy name",
			// openstack-kepler was delivered by an earlier version of the
			// operator and no longer exists, openstack-clod is a typo.
			disabled: []string{"openstack-kepler", dashboards.OpenstackNodeName, "openstack-clod"},
			want:     []string{"openstack-clod", "openstack-kepler"},
		},
		{
			name: "legacy ConfigMap name",
			// The ConfigMap names the dashboards used to be delivered under
			// are not the names they are disabled by.
			disabled: []string{"grafana-dashboard-openstack-cloud"},
			want:     []string{"grafana-dashboard-openstack-cloud"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			instance := &telemetryv1.MetricStorage{}
			instance.Spec.DisabledDashboards = tc.disabled

			g.Expect(unknownDisabledDashboards(instance, builders)).To(gomega.Equal(tc.want))
		})
	}
}
