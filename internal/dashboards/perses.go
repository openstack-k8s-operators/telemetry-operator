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
	"encoding/json"
	"fmt"

	persesv1alpha1 "github.com/perses/perses-operator/api/v1alpha1"
	persesv1 "github.com/perses/perses/pkg/model/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Canonical dashboard names. These are the identifiers an administrator uses in
// MetricStorage.spec.disabledDashboards to disable an individual dashboard, and
// the names of the PersesDashboard objects created in the openstack namespace.
// A name is defined here alongside its builder as each dashboard is converted to
// Perses.
const (
	OpenstackCloudName = "openstack-cloud"
)

// PersesDashboardBuilder builds a PersesDashboard object bound to the given
// datasource, in the given namespace. Returns an error if the dashboard
// definition cannot be constructed (e.g., unmarshal failure).
type PersesDashboardBuilder func(datasourceName, namespace string) (*persesv1alpha1.PersesDashboard, error)

// dashboardFromJSON builds a perses.dev/v1alpha1 PersesDashboard named name in
// the given namespace from a perses.dev/v1alpha2 PersesDashboard JSON
// definition.
//
// Dashboards are kept as JSON because that is what the Grafana-to-Perses
// conversion tooling emits and what the Perses UI exports, so edits stay
// ergonomic. The definitions use the v1alpha2 layout, where the dashboard
// content lives under spec.config, while the objects are created as v1alpha1
// (inline spec), which is the API served by the supported perses-operator.
// Only spec.config is read; metadata in the JSON is ignored.
func dashboardFromJSON(name, raw, namespace string) (*persesv1alpha1.PersesDashboard, error) {
	var doc struct {
		Spec struct {
			Config *persesv1.DashboardSpec `json:"config"`
		} `json:"spec"`
	}
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s dashboard: %w", name, err)
	}
	if doc.Spec.Config == nil {
		return nil, fmt.Errorf("%s dashboard has no spec.config section", name)
	}
	return &persesv1alpha1.PersesDashboard{
		TypeMeta: metav1.TypeMeta{
			APIVersion: persesv1alpha1.GroupVersion.String(),
			Kind:       "PersesDashboard",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: persesv1alpha1.Dashboard{
			// v1alpha2 spec.config -> v1alpha1 inline spec
			DashboardSpec: *doc.Spec.Config,
		},
	}, nil
}

// PersesDashboards returns the registry of PersesDashboard builders keyed by
// canonical dashboard name.
func PersesDashboards() map[string]PersesDashboardBuilder {
	return map[string]PersesDashboardBuilder{
		OpenstackCloudName: OpenstackCloudPerses,
	}
}
