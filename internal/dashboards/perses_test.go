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
	"strings"
	"testing"

	"github.com/onsi/gomega"
	persesv1alpha1 "github.com/perses/perses-operator/api/v1alpha1"
	persesdashboard "github.com/perses/perses/pkg/model/api/v1/dashboard"
)

// panelRefPrefix is the JSON reference prefix a layout item uses to point at a
// panel declared in the dashboard's spec.panels map.
const panelRefPrefix = "#/spec/panels/"

// TestPersesDashboardsBuild verifies that every registered dashboard builder
// parses its JSON definition and produces a well-formed perses.dev/v1alpha1
// PersesDashboard, with the v1alpha2 spec.config content, in the requested
// namespace.
//
// The content assertions matter because the definitions are decoded with a
// lenient json.Unmarshal, which drops anything it does not recognise: a
// definition in the wrong layout, or one with a mistyped key, would otherwise
// yield a dashboard that renders nothing and still pass.
func TestPersesDashboardsBuild(t *testing.T) {
	g := gomega.NewWithT(t)

	const namespace = "openstack"
	labels := map[string]string{"service": "metricStorage"}
	for name, builder := range PersesDashboards() {
		dashboard, err := builder("prometheus", namespace, labels)
		g.Expect(err).NotTo(gomega.HaveOccurred(), name)

		g.Expect(dashboard.Name).To(gomega.Equal(name))
		g.Expect(dashboard.Namespace).To(gomega.Equal(namespace))
		// The dashboards carry the service labels, so they can be selected
		// like the other resources the MetricStorage manages. Each gets its
		// own map, so one dashboard cannot relabel the others.
		g.Expect(dashboard.Labels).To(gomega.Equal(labels), name)
		g.Expect(dashboard.Labels).NotTo(gomega.BeIdenticalTo(labels), name)
		g.Expect(dashboard.Kind).To(gomega.Equal("PersesDashboard"))
		g.Expect(dashboard.APIVersion).To(gomega.Equal(persesv1alpha1.GroupVersion.String()))
		g.Expect(dashboard.Spec.Display).NotTo(gomega.BeNil(), name)
		g.Expect(dashboard.Spec.Display.Name).NotTo(gomega.BeEmpty(), name)

		g.Expect(dashboard.Spec.Panels).NotTo(gomega.BeEmpty(), name)
		g.Expect(dashboard.Spec.Layouts).NotTo(gomega.BeEmpty(), name)
		for panelName, panel := range dashboard.Spec.Panels {
			g.Expect(panel.Spec.Display.Name).NotTo(gomega.BeEmpty(), "%s/%s", name, panelName)
			g.Expect(panel.Spec.Plugin.Kind).NotTo(gomega.BeEmpty(), "%s/%s", name, panelName)
		}

		// Every panel placed by a layout must exist, and every declared panel
		// must be placed, so a panel dropped or renamed on one side only does
		// not leave the dashboard with an empty slot or an invisible panel.
		placed := map[string]bool{}
		for _, layout := range dashboard.Spec.Layouts {
			grid, ok := layout.Spec.(*persesdashboard.GridLayoutSpec)
			g.Expect(ok).To(gomega.BeTrue(), name)
			for _, item := range grid.Items {
				g.Expect(item.Content).NotTo(gomega.BeNil(), name)
				panelName := strings.TrimPrefix(item.Content.Ref, panelRefPrefix)
				g.Expect(panelName).NotTo(gomega.Equal(item.Content.Ref),
					"%s: unexpected panel reference %q", name, item.Content.Ref)
				g.Expect(dashboard.Spec.Panels).To(gomega.HaveKey(panelName), name)
				placed[panelName] = true
			}
		}
		for panelName := range dashboard.Spec.Panels {
			g.Expect(placed).To(gomega.HaveKey(panelName),
				"%s: panel %s is not placed by any layout", name, panelName)
		}
	}
}

// TestOpenstackCloudPerses checks the openstack-cloud dashboard content is read
// from the v1alpha2 spec.config section into the v1alpha1 inline spec.
func TestOpenstackCloudPerses(t *testing.T) {
	g := gomega.NewWithT(t)

	dashboard, err := OpenstackCloudPerses("prometheus", "openstack", nil)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(dashboard.Name).To(gomega.Equal(OpenstackCloudName))
	g.Expect(dashboard.Spec.Display).NotTo(gomega.BeNil())
	g.Expect(dashboard.Spec.Display.Name).To(gomega.Equal("Openstack / Node Exporter / USE Method / Cluster"))
	g.Expect(dashboard.Spec.Duration.String()).To(gomega.Equal("1h"))
}

// TestDashboardFromJSONErrors checks malformed JSON and definitions without a
// spec.config section (e.g., the v1alpha1 inline layout, whose content would
// otherwise be silently dropped) are rejected.
func TestDashboardFromJSONErrors(t *testing.T) {
	g := gomega.NewWithT(t)

	_, err := dashboardFromJSON("bad", `{"spec": `, "prometheus", "openstack", nil)
	g.Expect(err).To(gomega.HaveOccurred())

	_, err = dashboardFromJSON("inline", `{"spec": {"display": {"name": "Inline"}}}`, "prometheus", "openstack", nil)
	g.Expect(err).To(gomega.MatchError(ErrDashboardNoSpecConfig))
}

// TestDashboardDatasourceBinding checks every Prometheus plugin in the stored
// definitions, both panel queries and list variables, is pointed at the
// datasource the operator manages, so the dashboards do not silently fall back
// to whichever datasource is the default in the Perses project.
func TestDashboardDatasourceBinding(t *testing.T) {
	g := gomega.NewWithT(t)

	wantSelector := map[string]interface{}{
		"kind": persesDatasourceKind,
		"name": "my-datasource",
	}
	for name, builder := range PersesDashboards() {
		dashboard, err := builder("my-datasource", "openstack", nil)
		g.Expect(err).NotTo(gomega.HaveOccurred(), name)

		bound := 0
		for panelName, panel := range dashboard.Spec.Panels {
			for _, query := range panel.Spec.Queries {
				spec, ok := query.Spec.Plugin.Spec.(map[string]interface{})
				g.Expect(ok).To(gomega.BeTrue(), "%s/%s", name, panelName)
				g.Expect(spec).To(gomega.HaveKeyWithValue("datasource", wantSelector),
					"%s/%s", name, panelName)
				bound++
			}
		}
		for _, variable := range dashboard.Spec.Variables {
			listVariable, ok := variable.Spec.(*persesdashboard.ListVariableSpec)
			if !ok {
				continue
			}
			spec, ok := listVariable.Plugin.Spec.(map[string]interface{})
			g.Expect(ok).To(gomega.BeTrue(), name)
			g.Expect(spec).To(gomega.HaveKeyWithValue("datasource", wantSelector), name)
			bound++
		}
		g.Expect(bound).To(gomega.BeNumerically(">", 0), name)
	}
}

// TestDashboardFromJSONUnboundDatasource checks a definition whose panels would
// not be bound to the managed datasource, and would therefore silently query
// the project's default datasource, is rejected.
func TestDashboardFromJSONUnboundDatasource(t *testing.T) {
	g := gomega.NewWithT(t)

	_, err := dashboardFromJSON("no-prometheus", `{
	  "spec": {
	    "config": {
	      "display": {"name": "No Prometheus"},
	      "panels": {
	        "0_0": {
	          "kind": "Panel",
	          "spec": {
	            "display": {"name": "Static"},
	            "plugin": {"kind": "Markdown", "spec": {"text": "hello"}}
	          }
	        }
	      },
	      "layouts": []
	    }
	  }
	}`, "prometheus", "openstack", nil)
	g.Expect(err).To(gomega.MatchError(ErrDashboardNoPrometheusPlugin))
}
