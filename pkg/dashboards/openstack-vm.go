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

func OpenstackVM(dsName string) *corev1.ConfigMap {
	dashboardCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "grafana-dashboard-openstack-vm",
			Namespace: "openshift-config-managed",
			Labels: map[string]string{
				"console.openshift.io/dashboard": "true",
			},
		},
		Data: map[string]string{
			"openstack-vm.json": `
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
				"iteration": 1711654297335,
				"links": [],
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
						"y": 0
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
					"stack": false,
					"steppedLine": false,
					"targets": [
						{
						"expr": "vm:ceilometer_cpu:ratio1m{project =~ \"$project\", vm_name =~ \"$VM\"}",
						"hide": false,
						"interval": "",
						"legendFormat": "{{vm_name}}",
						"refId": "A"
						}
					],
					"thresholds": [],
					"timeFrom": null,
					"timeRegions": [],
					"timeShift": null,
					"title": "CPU Utilisation",
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
						"decimals": null,
						"format": "percentunit",
						"label": "",
						"logBase": 1,
						"max": null,
						"min": null,
						"show": true
						},
						{
						"format": "short",
						"label": "",
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
						"x": 12,
						"y": 0
					},
					"hiddenSeries": false,
					"id": 4,
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
						"expr": "vm:ceilometer_memory_usage:total{project =~ \"$project\", vm_name =~ \"$VM\"}",
						"hide": false,
						"interval": "",
						"legendFormat": "{{vm_name}}",
						"refId": "A"
						}
					],
					"thresholds": [],
					"timeFrom": null,
					"timeRegions": [],
					"timeShift": null,
					"title": "Memory Utilisation",
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
						"format": "bytes",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": true
						},
						{
						"format": "short",
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
						"y": 9
					},
					"hiddenSeries": false,
					"id": 6,
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
						"expr": "vm:ceilometer_disk_device_usage:total{project =~ \"$project\", vm_name =~ \"$VM\"}",
						"hide": false,
						"interval": "",
						"legendFormat": "{{vm_name}}({{device}})",
						"refId": "A"
						}
					],
					"thresholds": [],
					"timeFrom": null,
					"timeRegions": [],
					"timeShift": null,
					"title": "Disk Space Utilisation",
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
						"format": "bytes",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": true
						},
						{
						"format": "short",
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
						"x": 12,
						"y": 9
					},
					"hiddenSeries": false,
					"id": 8,
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
						"expr": "vm:ceilometer_disk_device_read_bytes:rate1m{project =~ \"$project\", vm_name =~ \"$VM\"}",
						"interval": "",
						"legendFormat": "{{vm_name}} read ({{device}})",
						"refId": "A"
						},
						{
						"expr": "vm:ceilometer_disk_device_write_bytes:rate1m{project =~ \"$project\", vm_name =~ \"$VM\"}",
						"hide": false,
						"interval": "",
						"legendFormat": "{{vm_name}} write ({{device}})",
						"refId": "B"
						}
					],
					"thresholds": [],
					"timeFrom": null,
					"timeRegions": [],
					"timeShift": null,
					"title": "Disk R/W Utilisation",
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
						"format": "Bps",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": true
						},
						{
						"format": "short",
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
						"y": 17
					},
					"hiddenSeries": false,
					"id": 10,
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
							"expr": "vm:ceilometer_network_incoming_bytes:rate1m{project =~ \"$project\", vm_name =~ \"$VM\"}",
							"hide": false,
							"interval": "",
							"legendFormat": "{{vm_name}} in ({{device}})",
							"refId": "B"
						},
						{
						"expr": "vm:ceilometer_network_outgoing_bytes:rate1m{project =~ \"$project\", vm_name =~ \"$VM\"}",
						"hide": false,
						"interval": "",
						"legendFormat": "{{vm_name}} out ({{device}})",
						"refId": "A"
						}
					],
					"thresholds": [],
					"timeFrom": null,
					"timeRegions": [],
					"timeShift": null,
					"title": "Network Utilisation",
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
						"format": "Bps",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": true
						},
						{
						"format": "short",
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
						"x": 12,
						"y": 17
					},
					"hiddenSeries": false,
					"id": 12,
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
						"expr": "vm:ceilometer_network_incoming_packets_drop:rate1m{project =~ \"$project\", vm_name =~ \"$VM\"}",
						"hide": false,
						"interval": "",
						"legendFormat": "{{vm_name}} in ({{device}})",
						"refId": "A"
						},
						{
						"expr": "vm:ceilometer_network_outgoing_packets_drop:rate1m{project =~ \"$project\", vm_name =~ \"$VM\"}",
						"hide": false,
						"interval": "",
						"legendFormat": "{{vm_name}} out ({{device}})",
						"refId": "B"
						}
					],
					"thresholds": [],
					"timeFrom": null,
					"timeRegions": [],
					"timeShift": null,
					"title": "Network Saturation (Drop Rate)",
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
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": true
						},
						{
						"format": "short",
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
						"x": 12,
						"y": 17
					},
					"hiddenSeries": false,
					"id": 12,
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
						"expr": "vm:ceilometer_network_incoming_packets_error:rate1m{project =~ \"$project\", vm_name =~ \"$VM\"}",
						"hide": false,
						"interval": "",
						"legendFormat": "{{vm_name}} in ({{device}})",
						"refId": "A"
						},
						{
						"expr": "vm:ceilometer_network_outgoing_packets_error:rate1m{project =~ \"$project\", vm_name =~ \"$VM\"}",
						"hide": false,
						"interval": "",
						"legendFormat": "{{vm_name}} out ({{device}})",
						"refId": "B"
						}
					],
					"thresholds": [],
					"timeFrom": null,
					"timeRegions": [],
					"timeShift": null,
					"title": "Network Error Rate",
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
						"format": "short",
						"label": null,
						"logBase": 1,
						"max": null,
						"min": null,
						"show": true
						},
						{
						"format": "short",
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
				],
				"refresh": "10s",
				"schemaVersion": 22,
				"style": "dark",
				"tags": [
					"openstack-telemetry-operator"
				],
				"templating": {
					"list": [
					{
						"allValue": ".*",
						"current": {
						"tags": [],
						"text": "66ad788dfacb4560b4c4e27e7ebc3b35",
						"value": [
							"66ad788dfacb4560b4c4e27e7ebc3b35"
						]
						},
						"datasource": { "name": "` + dsName + `", "type": "prometheus" },
						"definition": "label_values(ceilometer_cpu, project)",
						"hide": 0,
						"includeAll": true,
						"index": -1,
						"label": null,
						"multi": true,
						"name": "project",
						"options": [
						{
							"selected": false,
							"text": "All",
							"value": "$__all"
						},
						{
							"selected": true,
							"text": "66ad788dfacb4560b4c4e27e7ebc3b35",
							"value": "66ad788dfacb4560b4c4e27e7ebc3b35"
						},
						{
							"selected": false,
							"text": "ad3f0a004afc41fc80bb2a6e69ec4e6e",
							"value": "ad3f0a004afc41fc80bb2a6e69ec4e6e"
						}
						],
						"query": "label_values(ceilometer_cpu, project)",
						"refresh": 0,
						"regex": "",
						"skipUrlSync": false,
						"sort": 1,
						"tagValuesQuery": "",
						"tags": [],
						"tagsQuery": "",
						"type": "query",
						"useTags": false
					},
					{
						"allValue": ".*",
						"current": {
						"selected": false,
						"tags": [],
						"text": "d2a5f62671fcb969312fab2fef4bbfea78847225bb4686dc16d4f6dd",
						"value": [
							"d2a5f62671fcb969312fab2fef4bbfea78847225bb4686dc16d4f6dd"
						]
						},
						"datasource": { "name": "` + dsName + `", "type": "prometheus" },
						"definition": "label_values(ceilometer_cpu{project =~ \"$project\"}, vm_instance)",
						"hide": 0,
						"includeAll": true,
						"index": -1,
						"label": null,
						"multi": true,
						"name": "VM",
						"options": [
						{
							"selected": false,
							"text": "All",
							"value": "$__all"
						},
						{
							"selected": true,
							"text": "d2a5f62671fcb969312fab2fef4bbfea78847225bb4686dc16d4f6dd",
							"value": "d2a5f62671fcb969312fab2fef4bbfea78847225bb4686dc16d4f6dd"
						},
						{
							"selected": false,
							"text": "1d0808c7d062b9b20bfb2979b0d940b55c10312b1a56894a4b7c3238",
							"value": "1d0808c7d062b9b20bfb2979b0d940b55c10312b1a56894a4b7c3238"
						}
						],
						"query": "label_values(vm:ceilometer_cpu:ratio1m{project =~ \"$project\"}, vm_name)",
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
					"from": "now-6h",
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
				"title": "OpenStack / Ceilometer / VMs",
				"variables": {
					"list": []
				},
				"version": 18
				}`,
		},
	}

	return dashboardCM
}
