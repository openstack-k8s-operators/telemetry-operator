/*
Copyright 2024.

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

package metricstorage

import (
	"context"

	persesv1alpha1 "github.com/perses/perses-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	"github.com/openstack-k8s-operators/telemetry-operator/internal/dashboards"
	utils "github.com/openstack-k8s-operators/telemetry-operator/internal/utils"
	monv1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1"
)

// DashboardArtifactsNamespace is the namespace where the legacy ConfigMap-based
// dashboard artifacts were stored by earlier versions of the operator.
const DashboardArtifactsNamespace = "openshift-config-managed"

// legacyDashboardConfigMapNames are the ConfigMap-based dashboards created by
// earlier versions of the operator in DashboardArtifactsNamespace. They are
// removed on upgrade in favour of the PersesDashboard objects.
var legacyDashboardConfigMapNames = []string{
	"grafana-dashboard-openstack-cloud",
	"grafana-dashboard-openstack-node",
	"grafana-dashboard-openstack-openstack-network",
	"grafana-dashboard-openstack-vm",
	"grafana-dashboard-openstack-rabbitmq",
	"grafana-dashboard-openstack-network-traffic",
	"grafana-dashboard-openstack-ceilometer-ipmi",
	"grafana-dashboard-openstack-lightspeed",
	"grafana-dashboard-openstack-kepler",
}

// DeleteDashboardObjects removes the dashboard-related objects owned by the
// MetricStorage instance. It is used both on the disable path
// (DashboardsEnabled=false) and on CR deletion. Perses object deletions
// tolerate the perses.dev CRDs not being installed.
func DeleteDashboardObjects(ctx context.Context, instance *telemetryv1.MetricStorage, helper *helper.Helper) (ctrl.Result, error) {
	promRule := &monv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}
	if res, err := utils.EnsureDeleted(ctx, helper, promRule); err != nil {
		return res, err
	}

	datasource := &persesv1alpha1.PersesDatasource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PrometheusDatasourceName,
			Namespace: instance.Namespace,
		},
	}
	if res, err := ensureDeletedTolerateNoCRD(ctx, helper, datasource); err != nil {
		return res, err
	}

	for name := range dashboards.PersesDashboards() {
		dashboard := &persesv1alpha1.PersesDashboard{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: instance.Namespace,
			},
		}
		if res, err := ensureDeletedTolerateNoCRD(ctx, helper, dashboard); err != nil {
			return res, err
		}
	}

	return ctrl.Result{}, nil
}

// CleanupLegacyDashboardConfigMaps removes the ConfigMap-based dashboard
// artifacts created by earlier versions of the operator in the shared
// openshift-config-managed namespace. It is idempotent and safe to call on
// every reconcile regardless of whether dashboards are enabled.
func CleanupLegacyDashboardConfigMaps(ctx context.Context, instance *telemetryv1.MetricStorage, helper *helper.Helper) (ctrl.Result, error) {
	datasourceCMName := instance.Namespace + "-" + instance.Name + "-datasource"
	names := append([]string{datasourceCMName}, legacyDashboardConfigMapNames...)
	for _, name := range names {
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: DashboardArtifactsNamespace,
			},
		}
		if res, err := utils.EnsureDeleted(ctx, helper, configMap); err != nil {
			return res, err
		}
	}

	return ctrl.Result{}, nil
}

// ensureDeletedTolerateNoCRD deletes an object but treats a missing CRD (no REST
// mapping) as success, so the operator does not error on clusters where the
// perses.dev CRDs are not installed.
func ensureDeletedTolerateNoCRD(ctx context.Context, helper *helper.Helper, obj client.Object) (ctrl.Result, error) {
	res, err := utils.EnsureDeleted(ctx, helper, obj)
	if err != nil && meta.IsNoMatchError(err) {
		return ctrl.Result{}, nil
	}
	return res, err
}
