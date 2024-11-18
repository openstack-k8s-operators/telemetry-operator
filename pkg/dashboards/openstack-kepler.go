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

package dashboards

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func OpenstackKepler(dsName string) *corev1.ConfigMap {
	dashboardCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "grafana-dashboard-openstack-kepler",
			Namespace: "openshift-config-managed",
			Labels: map[string]string{
				"console.openshift.io/dashboard": "true",
			},
		},
		Data: map[string]string{
			"openstack-kepler.json": `
			{
				"annotations": {
					"list": [
					{
						"builtIn": 1,
						"datasource": "-- Grafana --",
						"enable": true,
						"hide": true,
						"iconColor": "rgba(0, 211, 255, 1)",
						"name": "Annotations & Alerts",
						"type": "dashboard"
					}
					]
				},
				"editable": true,
				"gnetId": null,
				"graphTooltip": 0,
				"id": 1,
				"iteration": 1728039706090,
				"links": [],
				"rows": [
					{
						"collapse": false,
						"collapsed": false,
						"repeat": null,
						"repeatIteration": null,
						"repeatRowId": null,
						"showTitle": true,
						"title": "Virtual Machines",
						"titleSize": "h6",
						"type": "row",
						"panels": [
							{
								"aliasColors": {},
								"bars": false,
								"dashLength": 10,
								"dashes": false,
								"datasource": { "name": "` + dsName + `", "type": "prometheus" },
								"fill": 10,
								"fillGradient": 0,
								"gridPos": {
									"h": 8,
									"w": 12,
									"x": 0,
									"y": 0
								},
								"hiddenSeries": false,
								"id": 17,
								"legend": {
									"avg": false,
									"current": false,
									"max": false,
									"min": false,
									"show": true,
									"total": false,
									"values": false
								},
								"lines": true,
								"linewidth": 1,
								"nullPointMode": "null",
								"options": {
									"dataLinks": []
								},
								"percentage": false,
								"pointradius": 5,
								"points": false,
								"renderer": "flot",
								"seriesOverrides": [],
								"spaceLength": 10,
								"span": 6,
								"stack": false,
								"steppedLine": false,
								"targets": [
									{
									"expr": "sum(rate(kepler_vm_platform_joules_total{fqdn=\"$compute\"}[1m])) by (vm_id,fqdn)",
									"legendFormat": "{{ vm_id }}",
									"refId": "A"
									}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "Power Consumption (W)",
								"tooltip": {
									"shared": true,
									"sort": 2,
									"value_type": "individual"
								},
								"type": "graph",
								"xaxis": {
									"buckets": null,
									"mode": "time",
									"name": null,
									"show": true,
									"values": []
								},
								"yaxes": [
									{
									"format": "watt",
									"label": null,
									"logBase": 1,
									"max": null,
									"min": null,
									"show": true
									},
									{
									"format": "watt",
									"label": null,
									"logBase": 1,
									"max": null,
									"min": null,
									"show": true
									}
								],
								"yaxis": {
									"align": false,
									"alignLevel": null
								}
							},
							{
								"aliasColors": {},
								"bars": true,
								"cacheTimeout": null,
								"dashLength": 10,
								"dashes": false,
								"datasource": { "name": "` + dsName + `", "type": "prometheus" },
								"fill": 1,
								"fillGradient": 0,
								"gridPos": {
								"h": 8,
								"w": 12,
								"x": 12,
								"y": 1
								},
								"hiddenSeries": false,
								"id": 14,
								"legend": {
								"avg": false,
								"current": false,
								"max": false,
								"min": false,
								"show": true,
								"total": false,
								"values": false
								},
								"lines": false,
								"linewidth": 1,
								"links": [],
								"nullPointMode": "null",
								"options": {
								"dataLinks": []
								},
								"percentage": false,
								"pluginVersion": "6.5.3",
								"pointradius": 2,
								"points": false,
								"renderer": "flot",
								"seriesOverrides": [],
								"spaceLength": 10,
								"span": 6,
								"stack": false,
								"steppedLine": false,
								"targets": [
								{
									"expr": "sum by (vm_id, fqdn) (increase(kepler_vm_platform_joules_total{fqdn=\"$compute\"}[24h:1m])) * 0.000000277777777777778",
									"legendFormat": "{{ vm_id }}",
									"refId": "A"
								}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "Power Consumed (kWh) - Last 24 hours",
								"tooltip": {
								"shared": true,
								"sort": 0,
								"value_type": "individual"
								},
								"type": "graph",
								"xaxis": {
								"buckets": null,
								"mode": "series",
								"name": null,
								"show": true,
								"values": [
									"current"
								]
								},
								"yaxes": [
								{
									"format": "kwatth",
									"label": null,
									"logBase": 1,
									"max": null,
									"min": null,
									"show": true
								},
								{
									"format": "kwatth",
									"label": null,
									"logBase": 1,
									"max": null,
									"min": null,
									"show": true
								}
								],
								"yaxis": {
									"align": false,
									"alignLevel": null
								}
							}
						]
					},
					{
						"collapse": false,
						"collapsed": false,
						"repeat": null,
						"repeatIteration": null,
						"repeatRowId": null,
						"showTitle": true,
						"title": "Compute Host",
						"titleSize": "h6",
						"type": "row",
						"panels": [
							{
								"aliasColors": {},
								"bars": false,
								"dashLength": 10,
								"dashes": false,
								"datasource": { "name": "` + dsName + `", "type": "prometheus" },
								"fill": 10,
								"fillGradient": 0,
								"gridPos": {
									"h": 8,
									"w": 12,
									"x": 0,
									"y": 10
								},
								"hiddenSeries": false,
								"id": 13,
								"legend": {
									"avg": false,
									"current": false,
									"max": false,
									"min": false,
									"show": true,
									"total": false,
									"values": false
								},
								"lines": true,
								"linewidth": 1,
								"nullPointMode": "null",
								"options": {
									"dataLinks": []
								},
								"percentage": false,
								"pointradius": 2,
								"points": false,
								"renderer": "flot",
								"seriesOverrides": [],
								"spaceLength": 10,
								"span": 6,
								"stack": true,
								"steppedLine": false,
								"targets": [
									{
									"expr": "sum by (fqdn) (rate(kepler_node_platform_joules_total{fqdn=\"$compute\"}[1m]))",
									"legendFormat": "{{ fqdn }}",
									"refId": "A"
									}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "Compute Node Power Consumption (W)",
								"tooltip": {
									"shared": true,
									"sort": 0,
									"value_type": "individual"
								},
								"type": "graph",
								"xaxis": {
									"buckets": null,
									"mode": "time",
									"name": null,
									"show": true,
									"values": []
								},
								"yaxes": [
									{
									"format": "watt",
									"label": null,
									"logBase": 1,
									"max": null,
									"min": null,
									"show": true
									},
									{
									"format": "watt",
									"label": null,
									"logBase": 1,
									"max": null,
									"min": null,
									"show": true
									}
								],
								"yaxis": {
									"align": false,
									"alignLevel": null
								}
							},
							{
								"aliasColors": {},
								"bars": true,
								"cacheTimeout": null,
								"dashLength": 10,
								"dashes": false,
								"datasource": { "name": "` + dsName + `", "type": "prometheus" },
								"fill": 1,
								"fillGradient": 0,
								"gridPos": {
								"h": 8,
								"w": 12,
								"x": 12,
								"y": 10
								},
								"hiddenSeries": false,
								"id": 18,
								"legend": {
								"avg": false,
								"current": false,
								"max": false,
								"min": false,
								"show": true,
								"total": false,
								"values": false
								},
								"lines": false,
								"linewidth": 1,
								"links": [],
								"nullPointMode": "null",
								"options": {
								"dataLinks": []
								},
								"percentage": false,
								"pluginVersion": "6.5.3",
								"pointradius": 2,
								"points": false,
								"renderer": "flot",
								"seriesOverrides": [],
								"spaceLength": 10,
								"span": 6,
								"stack": true,
								"steppedLine": false,
								"targets": [
								{
									"expr": "sum by (fqdn) (increase((kepler_node_platform_joules_total{fqdn=\"$compute\"}[24h:1m]))) * 0.000000277777777777778",
									"legendFormat": "{{ fqdn }}",
									"refId": "A"
								}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "Power Consumed (kWh) - Last 24 hours",
								"tooltip": {
								"shared": true,
								"sort": 0,
								"value_type": "individual"
								},
								"type": "graph",
								"xaxis": {
								"buckets": null,
								"mode": "time",
								"name": null,
								"show": true,
								"values": []
								},
								"yaxes": [
								{
									"format": "kwatth",
									"label": null,
									"logBase": 1,
									"max": null,
									"min": null,
									"show": true
								},
								{
									"format": "kwatth",
									"label": null,
									"logBase": 1,
									"max": null,
									"min": null,
									"show": true
								}
								],
								"yaxis": {
								"align": false,
								"alignLevel": null
								}
							}
						]
					},
					{
						"collapse": false,
						"collapsed": false,
						"repeat": null,
						"repeatIteration": null,
						"repeatRowId": null,
						"showTitle": true,
						"title": "Dataplane Services",
						"titleSize": "h6",
						"type": "row",
						"panels": [
							{
								"aliasColors": {},
								"bars": false,
								"dashLength": 10,
								"dashes": false,
								"datasource": { "name": "` + dsName + `", "type": "prometheus" },
								"fill": 10,
								"fillGradient": 0,
								"gridPos": {
									"h": 8,
									"w": 12,
									"x": 0,
									"y": 19
								},
								"hiddenSeries": false,
								"id": 9,
								"legend": {
									"avg": false,
									"current": false,
									"max": false,
									"min": false,
									"show": true,
									"total": false,
									"values": false
								},
								"lines": true,
								"linewidth": 1,
								"nullPointMode": "null",
								"options": {
									"dataLinks": []
								},
								"percentage": false,
								"pointradius": 2,
								"points": false,
								"renderer": "flot",
								"seriesOverrides": [],
								"spaceLength": 10,
								"stack": false,
								"steppedLine": false,
								"targets": [
									{
									"expr": "sum(rate(kepler_process_joules_total{vm_id=\"\",container_id=\"system_processes\",command=~\"virt.+|ovn.+|nova.+|ceil.+|neut.+|node.+|ovs.+|kepler|multi.+|haproxy|ceph.+|swift.+|chronyd\", fqdn=\"$compute\"}[1m])) by (command,fqdn)",
									"legendFormat": "{{ command }}",
									"refId": "A"
									}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "Power Consumption (W)",
								"tooltip": {
									"shared": true,
									"sort": 0,
									"value_type": "individual"
								},
								"type": "graph",
								"xaxis": {
									"buckets": null,
									"mode": "time",
									"name": null,
									"show": true,
									"values": []
								},
								"yaxes": [
									{
									"format": "Watt",
									"label": null,
									"logBase": 1,
									"max": null,
									"min": null,
									"show": true
									},
									{
									"format": "Watt",
									"label": null,
									"logBase": 1,
									"max": null,
									"min": null,
									"show": true
									}
								],
								"yaxis": {
									"align": false,
									"alignLevel": null
								}
							}
						]
					},
					{
						"collapse": false,
						"collapsed": false,
						"repeat": null,
						"repeatIteration": null,
						"repeatRowId": null,
						"showTitle": true,
						"title": "Container Namespaces",
						"titleSize": "h6",
						"type": "row",
						"panels": [
							{
								"aliasColors": {},
								"bars": false,
								"dashLength": 10,
								"dashes": false,
								"datasource": { "name": "` + dsName + `", "type": "prometheus" },
								"fill": 10,
								"fillGradient": 0,
								"gridPos": {
									"h": 9,
									"w": 12,
									"x": 0,
									"y": 28
								},
								"hiddenSeries": false,
								"id": 2,
								"legend": {
									"avg": false,
									"current": false,
									"max": false,
									"min": false,
									"show": true,
									"total": false,
									"values": false
								},
								"lines": true,
								"linewidth": 1,
								"nullPointMode": "null",
								"options": {
									"dataLinks": []
								},
								"percentage": false,
								"pointradius": 2,
								"points": false,
								"renderer": "flot",
								"seriesOverrides": [],
								"spaceLength": 10,
								"span": 6,
								"stack": false,
								"steppedLine": false,
								"targets": [
									{
									"expr": "sum by (container_namespace,fqdn) (rate(kepler_container_joules_total{fqdn=\"$compute\"}[1m]))",
									"legendFormat": "{{ container_namespace }}",
									"refId": "A"
									}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "Namespace Power Consumption (W)",
								"tooltip": {
									"shared": true,
									"sort": 0,
									"value_type": "individual"
								},
								"type": "graph",
								"xaxis": {
									"buckets": null,
									"mode": "time",
									"name": null,
									"show": true,
									"values": []
								},
								"yaxes": [
									{
									"format": "Watts",
									"label": null,
									"logBase": 1,
									"max": null,
									"min": null,
									"show": true
									},
									{
									"format": "Watts",
									"label": null,
									"logBase": 1,
									"max": null,
									"min": null,
									"show": true
									}
								],
								"yaxis": {
									"align": false,
									"alignLevel": null
								}
							}
						]
					}
				],
				"schemaVersion": 21,
				"style": "dark",
				"tags": [
					"openstack-telemetry-operator"
				],
				"templating": {
					"list": [
					{
						"allValue": null,
						"current": {
						"text": "",
						"value": ""
						},
						"datasource": { "name": "` + dsName + `", "type": "prometheus" },
						"definition": "label_values(kepler_node_info, fqdn)",
						"hide": 0,
						"includeAll": false,
						"label": null,
						"multi": false,
						"name": "compute",
						"options": [
						{
							"selected": true,
							"text": "",
							"value": ""
						}
						],
						"query": "label_values(kepler_node_info, fqdn)",
						"refresh": 0,
						"regex": "",
						"skipUrlSync": false,
						"sort": 0,
						"tagValuesQuery": "",
						"tags": [],
						"tagsQuery": "",
						"type": "query",
						"useTags": false
					}
					]
				},
				"time": {
					"from": "now-5m",
					"to": "now"
				},
				"timepicker": {
					"refresh_intervals": [
					"5s",
					"10s",
					"30s",
					"1m",
					"5m",
					"15m",
					"30m",
					"1h",
					"2h",
					"1d"
					]
				},
				"timezone": "",
				"title": "OpenStack / Compute Power Monitoring / Kepler",
				"version": 7
			}`,
		},
	}

	return dashboardCM
}
