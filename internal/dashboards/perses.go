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
	"errors"
	"fmt"
	"strings"

	persesv1alpha1 "github.com/perses/perses-operator/api/v1alpha1"
	persesv1 "github.com/perses/perses/pkg/model/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// persesDatasourceKind is the Perses plugin kind of the Prometheus datasource
// the dashboards query through, used both as the prefix identifying the
// Prometheus plugins to bind and as the kind of the datasource selector.
const persesDatasourceKind = "PrometheusDatasource"

// persesPrometheusPluginPrefix identifies the Perses plugins that query
// Prometheus, i.e. the ones that take a datasource selector:
// PrometheusTimeSeriesQuery in panel queries and PrometheusLabelValuesVariable
// (and friends) in list variables.
const persesPrometheusPluginPrefix = "Prometheus"

// Canonical dashboard names. These are the identifiers an administrator uses in
// MetricStorage.spec.disabledDashboards to disable an individual dashboard, and
// the names of the PersesDashboard objects created in the openstack namespace.
// A name is defined here alongside its builder as each dashboard is converted to
// Perses.
const (
	OpenstackCeilometerIpmiName = "openstack-ceilometer-ipmi"
	OpenstackCloudName          = "openstack-cloud"
	OpenstackLightspeedName     = "openstack-lightspeed"
	OpenstackNetworkName        = "openstack-network"
	OpenstackNetworkTrafficName = "openstack-network-traffic"
	OpenstackNodeName           = "openstack-node"
	OpenstackRabbitmqName       = "openstack-rabbitmq"
	OpenstackVMName             = "openstack-vm"
)

// ErrDashboardNoSpecConfig is returned when a dashboard JSON definition has no
// spec.config section, i.e. it is not in the perses.dev/v1alpha2 layout.
var ErrDashboardNoSpecConfig = errors.New("dashboard has no spec.config section")

// ErrDashboardNoPrometheusPlugin is returned when a dashboard JSON definition
// has no Prometheus plugin to bind to the datasource, which means none of its
// panels or variables would query the datasource the operator manages.
var ErrDashboardNoPrometheusPlugin = errors.New("dashboard has no Prometheus plugin to bind to the datasource")

// PersesDashboardBuilder builds a PersesDashboard object bound to the given
// datasource, in the given namespace. Returns an error if the dashboard
// definition cannot be constructed (e.g., unmarshal failure).
type PersesDashboardBuilder func(datasourceName, namespace string) (*persesv1alpha1.PersesDashboard, error)

// bindDatasource points every Prometheus plugin in the decoded dashboard
// content at the named datasource, by setting the datasource selector in the
// plugin spec, and returns how many plugins it bound.
//
// Without it the dashboards carry no datasource reference and Perses falls
// back to whichever datasource is the default in the project, so a default
// datasource the operator does not manage would silently serve the dashboards.
// Binding here rather than in the stored JSON keeps the definitions
// copy-pasteable between the Perses UI and this repository.
func bindDatasource(node interface{}, datasourceName string) int {
	bound := 0
	switch value := node.(type) {
	case map[string]interface{}:
		kind, isPlugin := value["kind"].(string)
		if isPlugin && kind != persesDatasourceKind && strings.HasPrefix(kind, persesPrometheusPluginPrefix) {
			if spec, ok := value["spec"].(map[string]interface{}); ok {
				spec["datasource"] = map[string]interface{}{
					"kind": persesDatasourceKind,
					"name": datasourceName,
				}
				bound++
			}
		}
		for _, child := range value {
			bound += bindDatasource(child, datasourceName)
		}
	case []interface{}:
		for _, child := range value {
			bound += bindDatasource(child, datasourceName)
		}
	}
	return bound
}

// dashboardFromJSON builds a perses.dev/v1alpha1 PersesDashboard named name in
// the given namespace from a perses.dev/v1alpha2 PersesDashboard JSON
// definition, with every Prometheus plugin bound to the named datasource.
//
// Dashboards are kept as JSON because that is what the Grafana-to-Perses
// conversion tooling emits and what the Perses UI exports, so edits stay
// ergonomic. The definitions use the v1alpha2 layout, where the dashboard
// content lives under spec.config, while the objects are created as v1alpha1
// (inline spec), which is the API served by the supported perses-operator.
// Only spec.config is read; metadata in the JSON is ignored.
func dashboardFromJSON(name, raw, datasourceName, namespace string) (*persesv1alpha1.PersesDashboard, error) {
	// Decode into a generic document first, so the datasource can be bound in
	// the schemaless plugin specs before the content is decoded into the typed
	// DashboardSpec, which validates it.
	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s dashboard: %w", name, err)
	}
	docSpec, _ := doc["spec"].(map[string]interface{})
	rawConfig, ok := docSpec["config"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%s: %w", name, ErrDashboardNoSpecConfig)
	}
	if bindDatasource(rawConfig, datasourceName) == 0 {
		return nil, fmt.Errorf("%s: %w", name, ErrDashboardNoPrometheusPlugin)
	}

	boundConfig, err := json.Marshal(rawConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal %s dashboard: %w", name, err)
	}
	config := &persesv1.DashboardSpec{}
	if err := json.Unmarshal(boundConfig, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s dashboard: %w", name, err)
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
			DashboardSpec: *config,
		},
	}, nil
}

// PersesDashboards returns the registry of PersesDashboard builders keyed by
// canonical dashboard name.
func PersesDashboards() map[string]PersesDashboardBuilder {
	return map[string]PersesDashboardBuilder{
		OpenstackCeilometerIpmiName: OpenstackCeilometerIpmiPerses,
		OpenstackCloudName:          OpenstackCloudPerses,
		OpenstackLightspeedName:     OpenstackLightspeedPerses,
		OpenstackNetworkName:        OpenstackNetworkPerses,
		OpenstackNetworkTrafficName: OpenstackNetworkTrafficPerses,
		OpenstackNodeName:           OpenstackNodePerses,
		OpenstackRabbitmqName:       OpenstackRabbitmqPerses,
		OpenstackVMName:             OpenstackVMPerses,
	}
}
