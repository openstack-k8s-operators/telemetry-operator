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
	"testing"

	. "github.com/onsi/gomega"
	persesv1alpha1 "github.com/perses/perses-operator/api/v1alpha1"
)

// TestPersesDashboardsBuild verifies that every registered dashboard builder
// parses its JSON definition and produces a well-formed perses.dev/v1alpha1
// PersesDashboard, with the v1alpha2 spec.config content, in the requested
// namespace.
func TestPersesDashboardsBuild(t *testing.T) {
	g := NewWithT(t)

	const namespace = "openstack"
	for name, builder := range PersesDashboards() {
		dashboard, err := builder("prometheus", namespace)
		g.Expect(err).NotTo(HaveOccurred(), name)

		g.Expect(dashboard.Name).To(Equal(name))
		g.Expect(dashboard.Namespace).To(Equal(namespace))
		g.Expect(dashboard.Kind).To(Equal("PersesDashboard"))
		g.Expect(dashboard.APIVersion).To(Equal(persesv1alpha1.GroupVersion.String()))
		g.Expect(dashboard.Spec.Display).NotTo(BeNil(), name)
		g.Expect(dashboard.Spec.Display.Name).NotTo(BeEmpty(), name)
	}
}

// TestOpenstackCloudPerses checks the openstack-cloud dashboard content is read
// from the v1alpha2 spec.config section into the v1alpha1 inline spec.
func TestOpenstackCloudPerses(t *testing.T) {
	g := NewWithT(t)

	dashboard, err := OpenstackCloudPerses("prometheus", "openstack")
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(dashboard.Name).To(Equal(OpenstackCloudName))
	g.Expect(dashboard.Spec.Display).NotTo(BeNil())
	g.Expect(dashboard.Spec.Display.Name).To(Equal("OpenStack / Cloud"))
	g.Expect(dashboard.Spec.Duration.String()).To(Equal("30m"))
}

// TestDashboardFromJSONErrors checks malformed JSON and definitions without a
// spec.config section (e.g., the v1alpha1 inline layout, whose content would
// otherwise be silently dropped) are rejected.
func TestDashboardFromJSONErrors(t *testing.T) {
	g := NewWithT(t)

	_, err := dashboardFromJSON("bad", `{"spec": `, "openstack")
	g.Expect(err).To(HaveOccurred())

	_, err = dashboardFromJSON("inline", `{"spec": {"display": {"name": "Inline"}}}`, "openstack")
	g.Expect(err).To(MatchError(ContainSubstring("no spec.config")))
}
