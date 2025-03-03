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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	utils "github.com/openstack-k8s-operators/telemetry-operator/pkg/utils"
	monv1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1"
)

const DashboardArtifactsNamespace = "openshift-config-managed"

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

	datasourceName := instance.Namespace + "-" + instance.Name + "-datasource"
	datasourceCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      datasourceName,
			Namespace: DashboardArtifactsNamespace,
		},
	}
	if res, err := utils.EnsureDeleted(ctx, helper, datasourceCM); err != nil {
		return res, err
	}

	var dashboards = []string{
		"grafana-dashboard-openstack-cloud",
		"grafana-dashboard-openstack-node",
		"grafana-dashboard-openstack-openstack-network",
		"grafana-dashboard-openstack-vm",
		"grafana-dashboard-openstack-rabbitmq",
		"grafana-dashboard-openstack-kepler",
		"grafana-dashboard-openstack-network-traffic",
		"grafana-dashboard-openstack-ceilometer-ipmi",
	}
	for _, name := range dashboards {
		dashboardCM := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: DashboardArtifactsNamespace,
			},
		}
		if res, err := utils.EnsureDeleted(ctx, helper, dashboardCM); err != nil {
			return res, err
		}
	}

	return ctrl.Result{}, nil
}
