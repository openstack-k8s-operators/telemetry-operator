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
	"time"

	persesv1alpha1 "github.com/perses/perses-operator/api/v1alpha1"
	persesv1 "github.com/perses/perses/pkg/model/api/v1"
	persescommon "github.com/perses/perses/pkg/model/api/v1/common"
	persesdashboard "github.com/perses/perses/pkg/model/api/v1/dashboard"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OpenstackCloudPerses builds the "OpenStack / Cloud" PersesDashboard.
//
// This is a minimal placeholder that establishes the object shape, ownership and
// datasource wiring. The full panel/layout content is converted from the legacy
// Grafana dashboard and delivered separately.
func OpenstackCloudPerses(_, namespace string) *persesv1alpha1.PersesDashboard {
	return &persesv1alpha1.PersesDashboard{
		TypeMeta: metav1.TypeMeta{
			APIVersion: persesv1alpha1.GroupVersion.String(),
			Kind:       "PersesDashboard",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      OpenstackCloudName,
			Namespace: namespace,
		},
		Spec: persesv1alpha1.Dashboard{
			DashboardSpec: persesv1.DashboardSpec{
				Display: &persescommon.Display{
					Name: "OpenStack / Cloud",
				},
				Duration: persescommon.Duration(30 * time.Minute),
				Panels:   map[string]*persesv1.Panel{},
				Layouts:  []persesdashboard.Layout{},
			},
		},
	}
}
