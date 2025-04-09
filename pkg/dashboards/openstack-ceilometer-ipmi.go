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

func OpenstackCeilometerIpmi(dsName string) *corev1.ConfigMap {
	dashboardCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "grafana-dashboard-openstack-ceilometer-ipmi",
			Namespace: "openshift-config-managed",
			Labels: map[string]string{
				"console.openshift.io/dashboard": "true",
			},
		},
		Data: map[string]string{
			"openstack-ceilometer-ipmi.json": `
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
						"title": "Power Supply Unit",
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
									"expr": "label_replace(ceilometer_hardware_ipmi_voltage{unit=\"V\", resource=~\"$compute.*\"}, \"resource_name\", \"$1\", \"resource\", \".*-(.*)\")",
									"legendFormat": "{{ resource_name }}",
									"refId": "A"
									}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "Voltage (Volts)",
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
									"format": "volts",
									"label": null,
									"logBase": 1,
									"max": null,
									"min": null,
									"show": true
									},
									{
									"format": "volts",
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
								"stack": false,
								"steppedLine": false,
								"targets": [
									{
									"expr": "label_replace(ceilometer_hardware_ipmi_current{unit=\"Amps\", resource=~\"$compute.*\"}, \"resource_name\", \"$1\", \"resource\", \".*-(current.*)\")",
									"legendFormat": "{{ resource_name }}",
									"refId": "A"
									}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "Current(Amps)",
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
									"format": "amps",
									"label": null,
									"logBase": 1,
									"max": null,
									"min": null,
									"show": true
									},
									{
									"format": "amps",
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
									"expr": "label_replace(ceilometer_hardware_ipmi_current{unit=\"W\", resource=~\"$compute.*\"}, \"resource_name\", \"$1\", \"resource\", \".*-(pwr.*)\")",
									"legendFormat": "{{ resource_name }}",
									"refId": "A"
									}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "Power (Watts)",
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
									"format": "watts",
									"label": null,
									"logBase": 1,
									"max": null,
									"min": null,
									"show": true
									},
									{
									"format": "watts",
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
						"title": "Cooling Overview",
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
									"expr": "label_replace(ceilometer_hardware_ipmi_fan{unit=\"RPM\", resource=~\"$compute.*\"}, \"resource_name\", \"$1\", \"resource\", \".*-(fan.*)\")",
									"legendFormat": "{{ resource_name }}",
									"refId": "A"
									}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "Fan (RPM)",
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
									"format": "rpm",
									"label": null,
									"logBase": 1,
									"max": null,
									"min": null,
									"show": true
									},
									{
									"format": "rpm",
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
									"expr": "label_replace(ceilometer_hardware_ipmi_temperature{unit=\"C\", resource=~\"$compute.*\"}, \"resource_name\", \"$2\", \"resource\", \"^(.*)\\\\-(exhaust_temp_.*|inlet_temp_.*|temp_.*)$\")",
									"legendFormat": "{{ resource_name }}",
									"refId": "A"
									}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "Temperature (Celsius)",
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
									"format": "celsius",
									"label": null,
									"logBase": 1,
									"max": null,
									"min": null,
									"show": true
									},
									{
									"format": "celsius",
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
						"definition": "label_values(label_replace(ceilometer_hardware_ipmi_voltage, \"node\", \"$1\", \"resource\", \"^(.*)-voltage.*\"), node)",
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
						"query": "label_values(label_replace(ceilometer_hardware_ipmi_voltage, \"node\", \"$1\", \"resource\", \"^(.*)-voltage.*\"), node)",
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
				"title": "OpenStack / Compute Power Monitoring / Ceilometer IPMI",
				"version": 7
			}`,
		},
	}

	return dashboardCM
}
