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

func OpenstackRabbitmq(dsName string) *corev1.ConfigMap {
	dashboardCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "grafana-dashboard-openstack-rabbitmq",
			Namespace: "openshift-config-managed",
			Labels: map[string]string{
				"console.openshift.io/dashboard": "true",
			},
		},
		Data: map[string]string{
			"openstack-rabbitmq.json": `
			{
				"annotations": {
					"list": [
					]
				},
				"editable": false,
				"gnetId": null,
				"graphTooltip": 0,
				"id": null,
				"iteration": 1718714973825,
				"links": [],
				"rows": [
				{
					"collapsed": true,
					"datasource": {
						"name": "` + dsName + `",
						"type": "prometheus"
					},
					"gridPos": {
						"h": 1,
						"w": 24,
						"x": 0,
						"y": 0
					},
					"id": 15,
					"panels": [
					{
						"cacheTimeout": null,
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"description": "Number of message queues",
						"gridPos": {
							"h": 6,
							"w": 3,
							"x": 0,
							"y": 1
						},
						"id": 6,
						"links": [],
						"options": {
							"colorMode": "value",
							"fieldOptions": {
								"calcs": [
								"mean"
								],
								"defaults": {
									"mappings": [],
									"thresholds": {
										"mode": "absolute",
										"steps": [
										{
											"color": "green",
											"value": null
										}
										]
									}
								},
								"overrides": [],
								"values": false
							},
							"graphMode": "area",
							"justifyMode": "auto",
							"orientation": "auto"
						},
						"pluginVersion": "6.7.6",
						"targets": [
						{
							"expr": "sum(rabbitmq_queues{instance=\"$cluster\"})",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"timeFrom": null,
						"timeShift": null,
						"title": "Queues",
						"span": 3,
						"type": "singlestat"
					},
					{
						"cacheTimeout": null,
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"description": "Number of message consumers",
						"gridPos": {
							"h": 6,
							"w": 3,
							"x": 3,
							"y": 1
						},
						"id": 7,
						"links": [],
						"options": {
							"colorMode": "value",
							"fieldOptions": {
								"calcs": [
								"mean"
								],
								"defaults": {
									"mappings": [],
									"thresholds": {
										"mode": "absolute",
										"steps": [
										{
											"color": "green",
											"value": null
										}
										]
									}
								},
								"overrides": [],
								"values": false
							},
							"graphMode": "area",
							"justifyMode": "auto",
							"orientation": "auto"
						},
						"pluginVersion": "6.7.6",
						"targets": [
						{
							"expr": "sum(rabbitmq_consumers{instance=\"$cluster\"})",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"timeFrom": null,
						"timeShift": null,
						"title": "Consumers",
						"span": 3,
						"type": "singlestat"
					},
					{
						"cacheTimeout": null,
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"description": "Number of RabbitMQ connections",
						"gridPos": {
							"h": 6,
							"w": 3,
							"x": 6,
							"y": 1
						},
						"id": 8,
						"links": [],
						"options": {
							"colorMode": "value",
							"fieldOptions": {
								"calcs": [
								"mean"
								],
								"defaults": {
									"mappings": [],
									"thresholds": {
										"mode": "absolute",
										"steps": [
										{
											"color": "green",
											"value": null
										}
										]
									}
								},
								"overrides": [],
								"values": false
							},
							"graphMode": "area",
							"justifyMode": "auto",
							"orientation": "auto"
						},
						"pluginVersion": "6.7.6",
						"targets": [
						{
							"expr": "sum(rabbitmq_connections{instance=\"$cluster\"})",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"timeFrom": null,
						"timeShift": null,
						"title": "Connections",
						"span": 3,
						"type": "singlestat"
					},
					{
						"cacheTimeout": null,
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"description": "Number of RabbitMQ channels",
						"gridPos": {
							"h": 6,
							"w": 3,
							"x": 9,
							"y": 1
						},
						"id": 9,
						"links": [],
						"options": {
							"colorMode": "value",
							"fieldOptions": {
								"calcs": [
								"mean"
								],
								"defaults": {
									"mappings": [],
									"thresholds": {
										"mode": "absolute",
										"steps": [
										{
											"color": "green",
											"value": null
										}
										]
									}
								},
								"overrides": [],
								"values": false
							},
							"graphMode": "area",
							"justifyMode": "auto",
							"orientation": "auto"
						},
						"pluginVersion": "6.7.6",
						"targets": [
						{
							"expr": "sum(rabbitmq_channels{instance=\"$cluster\"})",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"timeFrom": null,
						"timeShift": null,
						"title": "Channels",
						"span": 3,
						"type": "singlestat"
					},
					{
						"cacheTimeout": null,
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"description": "Number of incomming messages per second",
						"gridPos": {
							"h": 6,
							"w": 3,
							"x": 12,
							"y": 1
						},
						"id": 10,
						"links": [],
						"options": {
							"colorMode": "value",
							"fieldOptions": {
								"calcs": [
								"mean"
								],
								"defaults": {
									"mappings": [],
									"thresholds": {
										"mode": "absolute",
										"steps": [
										{
											"color": "green",
											"value": null
										}
										]
									}
								},
								"overrides": [],
								"values": false
							},
							"graphMode": "area",
							"justifyMode": "auto",
							"orientation": "auto"
						},
						"pluginVersion": "6.7.6",
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_global_messages_received_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"timeFrom": null,
						"timeShift": null,
						"title": "Incoming messages / s",
						"span": 3,
						"type": "singlestat"
					},
					{
						"cacheTimeout": null,
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"description": "Number of outgoing messages per second",
						"gridPos": {
							"h": 6,
							"w": 3,
							"x": 15,
							"y": 1
						},
						"id": 11,
						"links": [],
						"options": {
							"colorMode": "value",
							"fieldOptions": {
								"calcs": [
								"mean"
								],
								"defaults": {
									"mappings": [],
									"thresholds": {
										"mode": "absolute",
										"steps": [
										{
											"color": "green",
											"value": null
										}
										]
									}
								},
								"overrides": [],
								"values": false
							},
							"graphMode": "area",
							"justifyMode": "auto",
							"orientation": "auto"
						},
						"pluginVersion": "6.7.6",
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_global_messages_redelivered_total{instance=\"$cluster\"}[60s])) + sum(rate(rabbitmq_global_messages_delivered_consume_auto_ack_total{instance=\"$cluster\"}[60s])) + sum(rate(rabbitmq_global_messages_delivered_consume_manual_ack_total{instance=\"$cluster\"}[60s])) + sum(rate(rabbitmq_global_messages_delivered_get_auto_ack_total{instance=\"$cluster\"}[60s])) + sum(rate(rabbitmq_global_messages_delivered_get_manual_ack_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"timeFrom": null,
						"timeShift": null,
						"title": "Outgoing messages / s",
						"span": 3,
						"type": "singlestat"
					},
					{
						"cacheTimeout": null,
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"description": "Number of ready messages",
						"gridPos": {
							"h": 6,
							"w": 3,
							"x": 18,
							"y": 1
						},
						"id": 12,
						"links": [],
						"options": {
							"colorMode": "value",
							"fieldOptions": {
								"calcs": [
								"mean"
								],
								"defaults": {
									"mappings": [],
									"thresholds": {
										"mode": "absolute",
										"steps": [
										{
											"color": "green",
											"value": null
										}
										]
									}
								},
								"overrides": [],
								"values": false
							},
							"graphMode": "area",
							"justifyMode": "auto",
							"orientation": "auto"
						},
						"pluginVersion": "6.7.6",
						"targets": [
						{
							"expr": "sum(rabbitmq_queue_messages_ready{instance=\"$cluster\"})",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"timeFrom": null,
						"timeShift": null,
						"title": "Ready messages",
						"span": 3,
						"type": "singlestat"
					},
					{
						"cacheTimeout": null,
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"description": "Number of unacknowledged messages",
						"gridPos": {
							"h": 6,
							"w": 3,
							"x": 21,
							"y": 1
						},
						"id": 13,
						"links": [],
						"options": {
							"colorMode": "value",
							"fieldOptions": {
								"calcs": [
								"mean"
								],
								"defaults": {
									"mappings": [],
									"thresholds": {
										"mode": "absolute",
										"steps": [
										{
											"color": "green",
											"value": null
										}
										]
									}
								},
								"overrides": [],
								"values": false
							},
							"graphMode": "area",
							"justifyMode": "auto",
							"orientation": "auto"
						},
						"pluginVersion": "6.7.6",
						"targets": [
						{
							"expr": "sum(rabbitmq_queue_messages_unacked{instance=\"$cluster\"})",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"timeFrom": null,
						"timeShift": null,
						"title": "Unacknowledged messages",
						"span": 3,
						"type": "singlestat"
					}
					],
					"title": "Overview",
					"showTitle": "true",
					"type": "row"
				},
				{
					"collapsed": true,
					"datasource": {
						"name": "` + dsName + `",
						"type": "prometheus"
					},
					"gridPos": {
						"h": 1,
						"w": 24,
						"x": 0,
						"y": 1
					},
					"id": 55,
					"panels": [
					{
						"aliasColors": {},
						"bars": false,
						"dashLength": 10,
						"dashes": false,
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 10,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 12,
							"x": 0,
							"y": 2
						},
						"hiddenSeries": false,
						"id": 57,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "(rabbitmq_resident_memory_limit_bytes{instance=\"$cluster\"}) -\n(rabbitmq_process_resident_memory_bytes{instance=\"$cluster\"})",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Memory available before publishers blocked",
						"span": 6,
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 8,
							"x": 12,
							"y": 2
						},
						"hiddenSeries": false,
						"id": 58,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "rabbitmq_disk_space_available_bytes{instance=\"$cluster\"}",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Disk space available before publishers blocked",
						"span": 6,
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 4,
							"w": 4,
							"x": 20,
							"y": 2
						},
						"hiddenSeries": false,
						"id": 59,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "(rabbitmq_process_max_fds{instance=\"$cluster\"}) -\n(rabbitmq_process_open_fds{instance=\"$cluster\"})",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "File descriptors available",
						"span": 6,
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
							"format": "none",
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 4,
							"w": 4,
							"x": 20,
							"y": 6
						},
						"hiddenSeries": false,
						"id": 60,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "(rabbitmq_process_max_tcp_sockets{instance=\"$cluster\"}) -\n(rabbitmq_process_open_tcp_sockets{instance=\"$cluster\"})",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "TCP sockets available",
						"span": 6,
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
							"format": "none",
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
					"title": "Nodes",
					"showTitle": "true",
					"type": "row"
				},
				{
					"collapsed": true,
					"datasource": {
						"name": "` + dsName + `",
						"type": "prometheus"
					},
					"gridPos": {
						"h": 1,
						"w": 24,
						"x": 0,
						"y": 2
					},
					"id": 4,
					"panels": [
					{
						"aliasColors": {},
						"bars": false,
						"dashLength": 10,
						"dashes": false,
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"description": "",
						"fill": 0,
						"fillGradient": 0,
						"gridPos": {
							"h": 9,
							"w": 12,
							"x": 0,
							"y": 3
						},
						"hiddenSeries": false,
						"id": 2,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rabbitmq_queue_messages_ready{instance=\"$cluster\"})",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Messages ready to be delivered",
						"span": 6,
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
							"min": "0",
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"description": "",
						"fill": 0,
						"fillGradient": 0,
						"gridPos": {
							"h": 9,
							"w": 12,
							"x": 12,
							"y": 3
						},
						"hiddenSeries": false,
						"id": 16,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rabbitmq_queue_messages_unacked{instance=\"$cluster\"})",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Unacknowledged messages",
						"span": 6,
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
							"min": "0",
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
					"title": "Queued messages",
					"showTitle": "true",
					"type": "row"
				},
				{
					"collapsed": true,
					"datasource": {
						"name": "` + dsName + `",
						"type": "prometheus"
					},
					"gridPos": {
						"h": 1,
						"w": 24,
						"x": 0,
						"y": 3
					},
					"id": 18,
					"panels": [
					{
						"aliasColors": {},
						"bars": false,
						"dashLength": 10,
						"dashes": false,
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 12,
							"x": 0,
							"y": 4
						},
						"hiddenSeries": false,
						"id": 20,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_global_messages_received_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Messages published / s",
						"span": 6,
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 12,
							"x": 12,
							"y": 4
						},
						"hiddenSeries": false,
						"id": 21,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_global_messages_confirmed_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Messages confirmed / s",
						"span": 6,
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 12,
							"x": 0,
							"y": 12
						},
						"hiddenSeries": false,
						"id": 23,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_global_messages_routed_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Messages routed to queues / s",
						"span": 6,
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 12,
							"x": 12,
							"y": 12
						},
						"hiddenSeries": false,
						"id": 22,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_global_messages_received_confirm_total{instance=\"$cluster\"}[60s]) - \nrate(rabbitmq_global_messages_confirmed_total{instance=\"$cluster\"}[60s])\n)",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Unconfirmed messages / s",
						"span": 6,
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
							"min": "0",
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 12,
							"x": 0,
							"y": 20
						},
						"hiddenSeries": false,
						"id": 24,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_global_messages_unroutable_dropped_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Unroutable messages dropped / s",
						"span": 6,
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
							"min": "0",
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 12,
							"x": 12,
							"y": 20
						},
						"hiddenSeries": false,
						"id": 25,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_global_messages_unroutable_returned_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Unroutable messages returned to publishers / s",
						"span": 6,
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
							"min": "0",
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
					"title": "Incoming messages",
					"showTitle": "true",
					"type": "row"
				},
				{
					"collapsed": true,
					"datasource": {
						"name": "` + dsName + `",
						"type": "prometheus"
					},
					"gridPos": {
						"h": 1,
						"w": 24,
						"x": 0,
						"y": 4
					},
					"id": 27,
					"panels": [
					{
						"aliasColors": {},
						"bars": false,
						"dashLength": 10,
						"dashes": false,
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 12,
							"x": 0,
							"y": 5
						},
						"hiddenSeries": false,
						"id": 28,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(\n  rate(rabbitmq_global_messages_delivered_consume_auto_ack_total{instance=\"$cluster\"}[60s])+\n  rate(rabbitmq_global_messages_delivered_consume_manual_ack_total{instance=\"$cluster\"}[60s])\n)",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Messages delivered / s",
						"span": 6,
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 12,
							"x": 12,
							"y": 5
						},
						"hiddenSeries": false,
						"id": 29,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_global_messages_redelivered_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Messages redelivered / s",
						"span": 6,
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
							"min": "0",
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 12,
							"x": 0,
							"y": 13
						},
						"hiddenSeries": false,
						"id": 30,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_global_messages_delivered_consume_manual_ack_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Messages delivered with manual acks / s",
						"span": 6,
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 12,
							"x": 12,
							"y": 13
						},
						"hiddenSeries": false,
						"id": 31,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_global_messages_delivered_consume_auto_ack_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Messages delivered with auto acks / s",
						"span": 6,
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
							"min": "0",
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 12,
							"x": 0,
							"y": 21
						},
						"hiddenSeries": false,
						"id": 32,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_global_messages_acknowledged_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Messages acknowledged / s",
						"span": 6,
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
							"min": "0",
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 12,
							"x": 12,
							"y": 21
						},
						"hiddenSeries": false,
						"id": 33,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_global_messages_delivered_get_auto_ack_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Polling operations with auto ack / s",
						"span": 6,
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
							"min": "0",
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 12,
							"x": 0,
							"y": 29
						},
						"hiddenSeries": false,
						"id": 34,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_global_messages_get_empty_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Polling operations that yield no result / s",
						"span": 6,
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
							"min": "0",
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 12,
							"x": 12,
							"y": 29
						},
						"hiddenSeries": false,
						"id": 35,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_global_messages_delivered_get_manual_ack_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Polling operations with manual ack / s",
						"span": 6,
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
							"min": "0",
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
					"title": "Outgoing messages",
					"showTitle": "true",
					"type": "row"
				},
				{
					"collapsed": true,
					"datasource": {
						"name": "` + dsName + `",
						"type": "prometheus"
					},
					"gridPos": {
						"h": 1,
						"w": 24,
						"x": 0,
						"y": 5
					},
					"id": 37,
					"panels": [
					{
						"aliasColors": {},
						"bars": false,
						"dashLength": 10,
						"dashes": false,
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 12,
							"x": 0,
							"y": 6
						},
						"hiddenSeries": false,
						"id": 39,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "rabbitmq_queues{instance=\"$cluster\"}",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Total queues",
						"span": 6,
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
							"min": "0",
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 4,
							"x": 12,
							"y": 6
						},
						"hiddenSeries": false,
						"id": 40,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_queues_declared_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Queues declared / s",
						"span": 6,
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
							"min": "0",
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 4,
							"x": 16,
							"y": 6
						},
						"hiddenSeries": false,
						"id": 41,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_queues_created_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Queues created / s",
						"span": 6,
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
							"min": "0",
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 4,
							"x": 20,
							"y": 6
						},
						"hiddenSeries": false,
						"id": 42,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_queues_deleted_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Queues deleted / s",
						"span": 6,
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
							"min": "0",
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
					"title": "Queues",
					"showTitle": "true",
					"type": "row"
				},
				{
					"collapsed": true,
					"datasource": {
						"name": "` + dsName + `",
						"type": "prometheus"
					},
					"gridPos": {
						"h": 1,
						"w": 24,
						"x": 0,
						"y": 6
					},
					"id": 44,
					"panels": [
					{
						"aliasColors": {},
						"bars": false,
						"dashLength": 10,
						"dashes": false,
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 12,
							"x": 0,
							"y": 7
						},
						"hiddenSeries": false,
						"id": 46,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "rabbitmq_channels{instance=\"$cluster\"}",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Total channels",
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
							"min": "0",
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 6,
							"x": 12,
							"y": 7
						},
						"hiddenSeries": false,
						"id": 47,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_channels_opened_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Channels opened / s",
						"span": 6,
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
							"min": "0",
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 6,
							"x": 18,
							"y": 7
						},
						"hiddenSeries": false,
						"id": 48,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_channels_closed_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Channels closed / s",
						"span": 6,
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
							"min": "0",
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
					"title": "Channels",
					"showTitle": "true",
					"type": "row"
				},
				{
					"collapsed": true,
					"datasource": {
						"name": "` + dsName + `",
						"type": "prometheus"
					},
					"gridPos": {
						"h": 1,
						"w": 24,
						"x": 0,
						"y": 7
					},
					"id": 50,
					"panels": [
					{
						"aliasColors": {},
						"bars": false,
						"dashLength": 10,
						"dashes": false,
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 12,
							"x": 0,
							"y": 8
						},
						"hiddenSeries": false,
						"id": 51,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "rabbitmq_connections{instance=\"$cluster\"}",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Total connections",
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
							"min": "0",
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 6,
							"x": 12,
							"y": 8
						},
						"hiddenSeries": false,
						"id": 52,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_connections_opened_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Connections opened / s",
						"span": 6,
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
							"min": "0",
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
						"datasource": {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"fill": 1,
						"fillGradient": 0,
						"gridPos": {
							"h": 8,
							"w": 6,
							"x": 18,
							"y": 8
						},
						"hiddenSeries": false,
						"id": 53,
						"legend": {
							"avg": false,
							"current": false,
							"max": false,
							"min": false,
							"show": false,
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
						"stack": true,
						"steppedLine": false,
						"targets": [
						{
							"expr": "sum(rate(rabbitmq_connections_closed_total{instance=\"$cluster\"}[60s]))",
							"interval": "",
							"legendFormat": "",
							"refId": "A"
						}
						],
						"thresholds": [],
						"timeFrom": null,
						"timeRegions": [],
						"timeShift": null,
						"title": "Connections closed / s",
						"span": 6,
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
							"min": "0",
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
					"title": "Connections",
					"showTitle": "true",
					"type": "row"
				}
				],
				"refresh": false,
				"schemaVersion": 22,
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
						"datasource":  {
							"name": "` + dsName + `",
							"type": "prometheus"
						},
						"hide": 0,
						"includeAll": false,
						"label": "RabbitMQ Cluster",
						"multi": false,
						"name": "cluster",
						"options": [
						],
						"query": "label_values(rabbitmq_identity_info, instance)",
						"skipUrlSync": false,
						"type": "query"
					}
					]
				},
				"time": {
					"from": "now-30m",
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
				"title": "OpenStack / RabbitMQ",
				"uid": "r7U03y8Sk",
				"variables": {
					"list": []
				},
				"version": 5
			}`,
		},
	}

	return dashboardCM
}
