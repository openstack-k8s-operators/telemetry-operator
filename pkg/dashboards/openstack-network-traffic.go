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

func OpenstackNetworkTraffic(dsName string) *corev1.ConfigMap {
	dashboardCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "grafana-dashboard-openstack-network-traffic",
			Namespace: "openshift-config-managed",
			Labels: map[string]string{
				"console.openshift.io/dashboard": "true",
			},
		},
		Data: map[string]string{
			"openstack-network-traffic.json": `
			{
				"annotations": {
				  "list": [
					{
					  "builtIn": 1,
					  "datasource": {
						"type": "datasource",
						"uid": "grafana"
					  },
					  "enable": true,
					  "hide": true,
					  "iconColor": "rgba(0, 211, 255, 1)",
					  "name": "Annotations & Alerts",
					  "type": "dashboard"
					}
				  ]
				},
				"editable": true,
				"fiscalYearStartMonth": 0,
				"graphTooltip": 0,
				"id": 1,
				"links": [],
				"rows": [
				  {
					"collapsed": false,
					"gridPos": {
					  "h": 1,
					  "w": 24,
					  "x": 0,
					  "y": 0
					},
					"id": 18,
					"panels": [],
					"title": "Overview",
					"type": "row"
				  },
				  {
					"aliasColors": {},
					"bars": false,
					"dashLength": 10,
					"dashes": false,
					"datasource": { "name": "` + dsName + `", "type": "prometheus" },
					"fieldConfig": {
					  "defaults": {
						"links": []
					  },
					  "overrides": []
					},
					"fill": 10,
					"fillGradient": 0,
					"gridPos": {
					  "h": 8,
					  "w": 12,
					  "x": 0,
					  "y": 1
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
					  "alertThreshold": true
					},
					"percentage": false,
					"pluginVersion": "10.4.3",
					"pointradius": 2,
					"points": false,
					"renderer": "flot",
					"seriesOverrides": [],
					"spaceLength": 10,
					"stack": false,
					"steppedLine": false,
					"targets": [
					  {
						"datasource": { "name": "` + dsName + `", "type": "prometheus" },						,
						"expr": "vm:ceilometer_network_incoming_bytes:rate1m{project =~ \"$project\", vm_name =~ \"$VM\"}",
						"hide": false,
						"interval": "",
						"legendFormat": "{{vm_name}} in ({{device}})",
						"refId": "B"
					  },
					  {
						"datasource": {
						  "type": "prometheus",
						  "uid": "ce37wzjdfegw0e"
						},
						"expr": "vm:ceilometer_network_outgoing_bytes:rate1m{project =~ \"$project\", vm_name =~ \"$VM\"}",
						"hide": false,
						"interval": "",
						"legendFormat": "{{vm_name}} out ({{device}})",
						"refId": "A"
					  }
					],
					"thresholds": [],
					"timeRegions": [],
					"title": "Network Adapter",
					"tooltip": {
					  "shared": true,
					  "sort": 0,
					  "value_type": "individual"
					},
					"type": "graph",
					"xaxis": {
					  "mode": "time",
					  "show": true,
					  "values": []
					},
					"yaxes": [
					  {
						"format": "Bps",
						"logBase": 1,
						"show": true
					  },
					  {
						"format": "short",
						"logBase": 1,
						"show": true
					  }
					],
					"yaxis": {
					  "align": false
					}
				  },
				  {
					"collapsed": true,
					"gridPos": {
					  "h": 1,
					  "w": 24,
					  "x": 0,
					  "y": 9
					},
					"id": 15,
					"panels": [
					  {
						"aliasColors": {},
						"bars": false,
						"dashLength": 10,
						"dashes": false,
						"datasource": { "name": "` + dsName + `", "type": "prometheus" },						,
						"fieldConfig": {
						  "defaults": {
							"links": [],
							"unit": "p"
						  },
						  "overrides": []
						},
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
						  "alertThreshold": true
						},
						"percentage": false,
						"pluginVersion": "10.4.3",
						"pointradius": 2,
						"points": false,
						"renderer": "flot",
						"seriesOverrides": [],
						"spaceLength": 10,
						"stack": false,
						"steppedLine": false,
						"targets": [
						  {
							"datasource": { "name": "` + dsName + `", "type": "prometheus" },							,
							"editorMode": "code",
							"expr": "vm:ceilometer_network_incoming_bytes:rate1m{resource_name=~\"$VM:$In_adapter\", project=~\"$project\" } / 1000000",
							"hide": false,
							"interval": "",
							"legendFormat": "__auto",
							"range": true,
							"refId": "A"
						  }
						],
						"thresholds": [],
						"timeRegions": [],
						"title": "Network Incoming Packets",
						"tooltip": {
						  "shared": true,
						  "sort": 0,
						  "value_type": "individual"
						},
						"type": "graph",
						"xaxis": {
						  "mode": "time",
						  "show": true,
						  "values": []
						},
						"yaxes": [
						  {
							"$$hashKey": "object:892",
							"format": "p",
							"logBase": 1,
							"show": true
						  },
						  {
							"$$hashKey": "object:893",
							"format": "short",
							"logBase": 1,
							"show": true
						  }
						],
						"yaxis": {
						  "align": false
						}
					  },
					  {
						"datasource": { "name": "` + dsName + `", "type": "prometheus" },						,
						"fieldConfig": {
						  "defaults": {
							"color": {
							  "mode": "palette-classic"
							},
							"custom": {
							  "axisBorderShow": false,
							  "axisCenteredZero": false,
							  "axisColorMode": "text",
							  "axisGridShow": true,
							  "axisLabel": "",
							  "axisPlacement": "auto",
							  "barAlignment": 0,
							  "drawStyle": "line",
							  "fillOpacity": 100,
							  "gradientMode": "none",
							  "hideFrom": {
								"legend": false,
								"tooltip": false,
								"viz": false
							  },
							  "insertNulls": false,
							  "lineInterpolation": "linear",
							  "lineWidth": 1,
							  "pointSize": 5,
							  "scaleDistribution": {
								"type": "linear"
							  },
							  "showPoints": "never",
							  "spanNulls": false,
							  "stacking": {
								"group": "A",
								"mode": "none"
							  },
							  "thresholdsStyle": {
								"mode": "off"
							  }
							},
							"links": [],
							"mappings": [],
							"thresholds": {
							  "mode": "absolute",
							  "steps": [
								{
								  "color": "green"
								}
							  ]
							},
							"unit": "p"
						  },
						  "overrides": [
							{
							  "__systemRef": "hideSeriesFrom",
							  "matcher": {
								"id": "byNames",
								"options": {
								  "mode": "exclude",
								  "names": [
									"{__name__=\"vm:ceilometer_network_incoming_packets_drop:rate1m\", counter=\"network.incoming.packets.drop\", device=\"tap0cb7726a-da\", instance=\"ceilometer-internal.openstack.svc:3000\", network=\"packets\", project=\"539c3dc2361f4fd191aaa21c14360e35\", resource=\"instance-00000002-618ba30a-64d4-4795-9f6e-b85192d9305e-tap0cb7726a-da\", resource_name=\"vm1:tap0cb7726a-da\", unit=\"packet\", user=\"da73d63ddab141ab9fefbe03881fb6cd\", vm_instance=\"ee80218bf7db3248e7dd153f3340014d116f7af438ddeb08420677e9\", vm_name=\"vm1\"}"
								  ],
								  "prefix": "All except:",
								  "readOnly": true
								}
							  },
							  "properties": [
								{
								  "id": "custom.hideFrom",
								  "value": {
									"legend": false,
									"tooltip": false,
									"viz": true
								  }
								}
							  ]
							}
						  ]
						},
						"gridPos": {
						  "h": 9,
						  "w": 12,
						  "x": 12,
						  "y": 9
						},
						"id": 2,
						"options": {
						  "legend": {
							"calcs": [],
							"displayMode": "table",
							"placement": "bottom",
							"showLegend": true
						  },
						  "timezone": [
							""
						  ],
						  "tooltip": {
							"mode": "multi",
							"sort": "none"
						  }
						},
						"pluginVersion": "10.4.3",
						"targets": [
						  {
							"datasource": { "name": "` + dsName + `", "type": "prometheus" },							,
							"editorMode": "code",
							"expr": "vm:ceilometer_network_incoming_packets_drop:rate1m{project =~ \"$project\",vm_name =~ \"$VM\", device =~\"$In_adapter\" }",
							"hide": false,
							"interval": "",
							"legendFormat": "__auto",
							"range": true,
							"refId": "A"
						  }
						],
						"title": "Incoming Dropped Packets",
						"type": "timeseries"
					  },
					  {
						"aliasColors": {},
						"bars": false,
						"dashLength": 10,
						"dashes": false,
						"datasource": { "name": "` + dsName + `", "type": "prometheus" },						,
						"fieldConfig": {
						  "defaults": {
							"links": [],
							"unit": "percentunit"
						  },
						  "overrides": []
						},
						"fill": 10,
						"fillGradient": 0,
						"gridPos": {
						  "h": 9,
						  "w": 12,
						  "x": 0,
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
						  "alertThreshold": true
						},
						"percentage": false,
						"pluginVersion": "10.4.3",
						"pointradius": 2,
						"points": false,
						"renderer": "flot",
						"seriesOverrides": [],
						"spaceLength": 10,
						"stack": false,
						"steppedLine": false,
						"targets": [
						  {
							"datasource": { "name": "` + dsName + `", "type": "prometheus" },							,
							"editorMode": "code",
							"expr": "(rate(ceilometer_network_incoming_packets_drop{project=~\"$project\", resource_name=~\"$VM:$In_adapter\"}[1m]) / rate(ceilometer_network_incoming_packets{project=~\"$project\", resource_name=~\"$VM:In_adapter\"}[1m])) * 100\n",
							"hide": false,
							"interval": "",
							"legendFormat": "__auto",
							"range": true,
							"refId": "A"
						  }
						],
						"thresholds": [
						  {
							"$$hashKey": "object:151",
							"colorMode": "critical",
							"fill": true,
							"line": true,
							"op": "gt",
							"value": 30,
							"yaxis": "left"
						  }
						],
						"timeRegions": [],
						"title": "Incoming Packet Loss (%)",
						"tooltip": {
						  "shared": true,
						  "sort": 0,
						  "value_type": "individual"
						},
						"type": "graph",
						"xaxis": {
						  "mode": "time",
						  "show": true,
						  "values": []
						},
						"yaxes": [
						  {
							"$$hashKey": "object:946",
							"format": "percentunit",
							"logBase": 1,
							"show": true
						  },
						  {
							"$$hashKey": "object:947",
							"format": "short",
							"logBase": 1,
							"show": true
						  }
						],
						"yaxis": {
						  "align": false
						}
					  },
					  {
						"datasource": { "name": "` + dsName + `", "type": "prometheus" },						,
						"fieldConfig": {
						  "defaults": {
							"color": {
							  "mode": "palette-classic"
							},
							"custom": {
							  "axisBorderShow": false,
							  "axisCenteredZero": false,
							  "axisColorMode": "text",
							  "axisGridShow": true,
							  "axisLabel": "",
							  "axisPlacement": "auto",
							  "barAlignment": 0,
							  "drawStyle": "line",
							  "fillOpacity": 100,
							  "gradientMode": "none",
							  "hideFrom": {
								"legend": false,
								"tooltip": false,
								"viz": false
							  },
							  "insertNulls": false,
							  "lineInterpolation": "linear",
							  "lineWidth": 1,
							  "pointSize": 5,
							  "scaleDistribution": {
								"type": "linear"
							  },
							  "showPoints": "never",
							  "spanNulls": false,
							  "stacking": {
								"group": "A",
								"mode": "none"
							  },
							  "thresholdsStyle": {
								"mode": "off"
							  }
							},
							"links": [],
							"mappings": [],
							"thresholds": {
							  "mode": "absolute",
							  "steps": [
								{
								  "color": "green"
								}
							  ]
							},
							"unit": "p"
						  },
						  "overrides": [
							{
							  "__systemRef": "hideSeriesFrom",
							  "matcher": {
								"id": "byNames",
								"options": {
								  "mode": "exclude",
								  "names": [
									"{__name__=\"vm:ceilometer_network_incoming_packets_error:rate1m\", counter=\"network.incoming.packets.error\", device=\"tap0cb7726a-da\", instance=\"ceilometer-internal.openstack.svc:3000\", network=\"packets\", project=\"539c3dc2361f4fd191aaa21c14360e35\", resource=\"instance-00000002-618ba30a-64d4-4795-9f6e-b85192d9305e-tap0cb7726a-da\", resource_name=\"vm1:tap0cb7726a-da\", unit=\"packet\", user=\"da73d63ddab141ab9fefbe03881fb6cd\", vm_instance=\"ee80218bf7db3248e7dd153f3340014d116f7af438ddeb08420677e9\", vm_name=\"vm1\"}"
								  ],
								  "prefix": "All except:",
								  "readOnly": true
								}
							  },
							  "properties": [
								{
								  "id": "custom.hideFrom",
								  "value": {
									"legend": false,
									"tooltip": false,
									"viz": true
								  }
								}
							  ]
							}
						  ]
						},
						"gridPos": {
						  "h": 9,
						  "w": 12,
						  "x": 12,
						  "y": 18
						},
						"id": 16,
						"options": {
						  "legend": {
							"calcs": [],
							"displayMode": "table",
							"placement": "bottom",
							"showLegend": true
						  },
						  "timezone": [
							""
						  ],
						  "tooltip": {
							"mode": "multi",
							"sort": "none"
						  }
						},
						"pluginVersion": "10.4.3",
						"targets": [
						  {
							"datasource": { "name": "` + dsName + `", "type": "prometheus" },							,
							"editorMode": "code",
							"expr": "vm:ceilometer_network_incoming_packets_error:rate1m{project =~ \"$project\",vm_name =~ \"$VM\", device =~\"$In_adapter\"}",
							"hide": false,
							"interval": "",
							"legendFormat": "__auto",
							"range": true,
							"refId": "A"
						  }
						],
						"title": "Incoming Packets Error",
						"type": "timeseries"
					  }
					],
					"title": "Inbound Traffic",
					"type": "row"
				  },
				  {
					"collapsed": true,
					"gridPos": {
					  "h": 1,
					  "w": 24,
					  "x": 0,
					  "y": 10
					},
					"id": 14,
					"panels": [
					  {
						"aliasColors": {},
						"bars": false,
						"dashLength": 10,
						"dashes": false,
						"datasource": {
						  "type": "prometheus",
						  "uid": "ce37wzjdfegw0e"
						},
						"fieldConfig": {
						  "defaults": {
							"links": [],
							"unit": "p"
						  },
						  "overrides": []
						},
						"fill": 10,
						"fillGradient": 0,
						"gridPos": {
						  "h": 8,
						  "w": 12,
						  "x": 0,
						  "y": 11
						},
						"hiddenSeries": false,
						"id": 11,
						"legend": {
						  "alignAsTable": true,
						  "avg": false,
						  "current": false,
						  "max": false,
						  "min": false,
						  "rightSide": false,
						  "show": true,
						  "total": false,
						  "values": false
						},
						"lines": true,
						"linewidth": 1,
						"nullPointMode": "null",
						"options": {
						  "alertThreshold": true
						},
						"percentage": false,
						"pluginVersion": "10.4.3",
						"pointradius": 2,
						"points": false,
						"renderer": "flot",
						"seriesOverrides": [],
						"spaceLength": 10,
						"stack": false,
						"steppedLine": false,
						"targets": [
						  {
							"datasource": { "name": "` + dsName + `", "type": "prometheus" },							,
							"editorMode": "code",
							"expr": "vm:ceilometer_network_outgoing_bytes:rate1m{resource_name=~\"$VM:$out_adapter\", project=~\"$project\"} / 1000000",
							"hide": false,
							"interval": "",
							"legendFormat": "__auto",
							"range": true,
							"refId": "A"
						  }
						],
						"thresholds": [],
						"timeRegions": [],
						"title": "Network Outgoing Packets",
						"tooltip": {
						  "shared": true,
						  "sort": 0,
						  "value_type": "individual"
						},
						"type": "graph",
						"xaxis": {
						  "mode": "time",
						  "show": true,
						  "values": []
						},
						"yaxes": [
						  {
							"$$hashKey": "object:1127",
							"format": "p",
							"logBase": 1,
							"show": true
						  },
						  {
							"$$hashKey": "object:1128",
							"format": "short",
							"logBase": 1,
							"show": true
						  }
						],
						"yaxis": {
						  "align": false
						}
					  },
					  {
						"aliasColors": {},
						"bars": false,
						"dashLength": 10,
						"dashes": false,
						"datasource": { "name": "` + dsName + `", "type": "prometheus" },						,
						"fieldConfig": {
						  "defaults": {
							"links": [],
							"unit": "p"
						  },
						  "overrides": []
						},
						"fill": 10,
						"fillGradient": 0,
						"gridPos": {
						  "h": 9,
						  "w": 12,
						  "x": 12,
						  "y": 11
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
						  "alertThreshold": true
						},
						"percentage": false,
						"pluginVersion": "10.4.3",
						"pointradius": 2,
						"points": false,
						"renderer": "flot",
						"seriesOverrides": [],
						"spaceLength": 10,
						"stack": false,
						"steppedLine": false,
						"targets": [
						  {
							"datasource": { "name": "` + dsName + `", "type": "prometheus" },							,
							"editorMode": "code",
							"expr": "vm:ceilometer_network_outgoing_packets_drop:rate1m{project =~ \"$project\", vm_name =~ \"$VM\", device=~\"$out_adapter\"}",
							"hide": false,
							"interval": "",
							"legendFormat": "__auto",
							"range": true,
							"refId": "A"
						  }
						],
						"thresholds": [],
						"timeRegions": [],
						"title": "Outgoing Dropped Packets",
						"tooltip": {
						  "shared": true,
						  "sort": 0,
						  "value_type": "individual"
						},
						"type": "graph",
						"xaxis": {
						  "mode": "time",
						  "show": true,
						  "values": []
						},
						"yaxes": [
						  {
							"$$hashKey": "object:764",
							"format": "p",
							"logBase": 1,
							"show": true
						  },
						  {
							"$$hashKey": "object:765",
							"format": "short",
							"logBase": 1,
							"show": true
						  }
						],
						"yaxis": {
						  "align": false
						}
					  },
					  {
						"aliasColors": {},
						"bars": false,
						"dashLength": 10,
						"dashes": false,
						"datasource": { "name": "` + dsName + `", "type": "prometheus" },						,
						"fieldConfig": {
						  "defaults": {
							"links": [],
							"unit": "percentunit"
						  },
						  "overrides": []
						},
						"fill": 10,
						"fillGradient": 0,
						"gridPos": {
						  "h": 9,
						  "w": 12,
						  "x": 0,
						  "y": 19
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
						  "alertThreshold": true
						},
						"percentage": false,
						"pluginVersion": "10.4.3",
						"pointradius": 2,
						"points": false,
						"renderer": "flot",
						"seriesOverrides": [],
						"spaceLength": 10,
						"stack": false,
						"steppedLine": false,
						"targets": [
						  {
							"datasource": { "name": "` + dsName + `", "type": "prometheus" },							,
							"editorMode": "code",
							"expr": "(rate(ceilometer_network_outgoing_packets_drop{project=~\"$project\", resource_name=~\"$VM:$out_adapter\"}[1m]) / rate(ceilometer_network_outgoing_packets{project=~\"$project\", resource_name=~\"$VM:$out_adapter\"}[1m])) * 100\n",
							"hide": false,
							"interval": "",
							"legendFormat": "__auto",
							"range": true,
							"refId": "A"
						  }
						],
						"thresholds": [
						  {
							"$$hashKey": "object:216",
							"colorMode": "critical",
							"fill": true,
							"line": true,
							"op": "gt",
							"value": 80,
							"yaxis": "left"
						  }
						],
						"timeRegions": [],
						"title": "Ountgoing Packet Loss (%)",
						"tooltip": {
						  "shared": true,
						  "sort": 0,
						  "value_type": "individual"
						},
						"type": "graph",
						"xaxis": {
						  "mode": "time",
						  "show": true,
						  "values": []
						},
						"yaxes": [
						  {
							"$$hashKey": "object:1253",
							"format": "percentunit",
							"logBase": 1,
							"show": true
						  },
						  {
							"$$hashKey": "object:1254",
							"format": "short",
							"logBase": 1,
							"show": true
						  }
						],
						"yaxis": {
						  "align": false
						}
					  },
					  {
						"aliasColors": {},
						"bars": false,
						"dashLength": 10,
						"dashes": false,
						"datasource": { "name": "` + dsName + `", "type": "prometheus" },						,
						"fieldConfig": {
						  "defaults": {
							"links": [],
							"unit": "p"
						  },
						  "overrides": []
						},
						"fill": 10,
						"fillGradient": 0,
						"gridPos": {
						  "h": 9,
						  "w": 12,
						  "x": 12,
						  "y": 20
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
						  "alertThreshold": true
						},
						"percentage": false,
						"pluginVersion": "10.4.3",
						"pointradius": 2,
						"points": false,
						"renderer": "flot",
						"seriesOverrides": [],
						"spaceLength": 10,
						"stack": false,
						"steppedLine": false,
						"targets": [
						  {
							"datasource": { "name": "` + dsName + `", "type": "prometheus" },							,
							"editorMode": "code",
							"expr": "vm:ceilometer_network_outgoing_packets_error:rate1m{project =~ \"$project\", vm_name =~ \"$VM\", device=~\"$out_adapter\"}",
							"hide": false,
							"interval": "",
							"legendFormat": "__auto",
							"range": true,
							"refId": "A"
						  }
						],
						"thresholds": [],
						"timeRegions": [],
						"title": "Outgoing Packets Error",
						"tooltip": {
						  "shared": true,
						  "sort": 0,
						  "value_type": "individual"
						},
						"type": "graph",
						"xaxis": {
						  "mode": "time",
						  "show": true,
						  "values": []
						},
						"yaxes": [
						  {
							"$$hashKey": "object:764",
							"format": "p",
							"logBase": 1,
							"show": true
						  },
						  {
							"$$hashKey": "object:765",
							"format": "short",
							"logBase": 1,
							"show": true
						  }
						],
						"yaxis": {
						  "align": false
						}
					  }
					],
					"title": "Outbound Traffic",
					"type": "row"
				  }
				],
				"refresh": "10s",
				"schemaVersion": 39,
				"tags": [
				  "openstack-telemetry-operator"
				],
				"templating": {
				  "list": [
					{
					  "allValue": ".*",
					  "current": {
						"selected": false,
						"text": "ee80218bf7db3248e7dd153f3340014d116f7af438ddeb08420677e9",
						"value": "ee80218bf7db3248e7dd153f3340014d116f7af438ddeb08420677e9"
					  },
					  "datasource": { "name": "` + dsName + `", "type": "prometheus" },					  ,
					  "definition": "label_values(vm:ceilometer_network_incoming_bytes:rate1m,vm_instance)",
					  "hide": 0,
					  "includeAll": true,
					  "multi": true,
					  "name": "compute_node",
					  "options": [],
					  "query": {
						"qryType": 1,
						"query": "label_values(vm:ceilometer_network_incoming_bytes:rate1m,vm_instance)",
						"refId": "PrometheusVariableQueryEditor-VariableQuery"
					  },
					  "refresh": 1,
					  "regex": "",
					  "skipUrlSync": false,
					  "sort": 0,
					  "type": "query"
					},
					{
					  "allValue": ".*",
					  "current": {
						"selected": true,
						"text": [
						  "All"
						],
						"value": [
						  "$__all"
						]
					  },
					  "datasource":{ "name": "` + dsName + `", "type": "prometheus" },					  ,
					  "definition": "label_values(ceilometer_cpu{vm_instance=~\"$compute_node\"},project)",
					  "hide": 0,
					  "includeAll": true,
					  "multi": true,
					  "name": "project",
					  "options": [],
					  "query": {
						"qryType": 1,
						"query": "label_values(ceilometer_cpu{vm_instance=~\"$compute_node\"},project)",
						"refId": "PrometheusVariableQueryEditor-VariableQuery"
					  },
					  "refresh": 1,
					  "regex": "",
					  "skipUrlSync": false,
					  "sort": 1,
					  "tagValuesQuery": "",
					  "tagsQuery": "",
					  "type": "query",
					  "useTags": false
					},
					{
					  "allValue": ".*",
					  "current": {
						"selected": true,
						"text": [
						  "All"
						],
						"value": [
						  "$__all"
						]
					  },
					  "datasource": { "name": "` + dsName + `", "type": "prometheus" },					  ,
					  "definition": "label_values(ceilometer_cpu{project =~ \"$project\"}, vm_instance)",
					  "hide": 0,
					  "includeAll": true,
					  "multi": true,
					  "name": "VM",
					  "options": [],
					  "query": "label_values(vm:ceilometer_cpu:ratio1m{project =~ \"$project\"}, vm_name)",
					  "refresh": 1,
					  "regex": "",
					  "skipUrlSync": false,
					  "sort": 0,
					  "tagValuesQuery": "",
					  "tagsQuery": "",
					  "type": "query",
					  "useTags": false
					},
					{
					  "allValue": ".*",
					  "current": {
						"selected": true,
						"text": [
						  "All"
						],
						"value": [
						  "$__all"
						]
					  },
					  "datasource": { "name": "` + dsName + `", "type": "prometheus" },					  ,
					  "definition": "label_values(vm:ceilometer_network_incoming_bytes:rate1m{vm_name=~\"$VM\"},device)",
					  "hide": 0,
					  "includeAll": true,
					  "multi": true,
					  "name": "In_adapter",
					  "options": [],
					  "query": {
						"qryType": 1,
						"query": "label_values(vm:ceilometer_network_incoming_bytes:rate1m{vm_name=~\"$VM\"},device)",
						"refId": "PrometheusVariableQueryEditor-VariableQuery"
					  },
					  "refresh": 1,
					  "regex": "",
					  "skipUrlSync": false,
					  "sort": 0,
					  "type": "query"
					},
					{
					  "allValue": ".*",
					  "current": {
						"selected": false,
						"text": "tap0cb7726a-da",
						"value": "tap0cb7726a-da"
					  },
					  "datasource": { "name": "` + dsName + `", "type": "prometheus" },					  ,
					  "definition": "label_values(vm:ceilometer_network_outgoing_bytes:rate1m{vm_name=~\"$VM\"},device)",
					  "hide": 0,
					  "includeAll": true,
					  "multi": true,
					  "name": "out_adapter",
					  "options": [],
					  "query": {
						"qryType": 1,
						"query": "label_values(vm:ceilometer_network_outgoing_bytes:rate1m{vm_name=~\"$VM\"},device)",
						"refId": "PrometheusVariableQueryEditor-VariableQuery"
					  },
					  "refresh": 1,
					  "regex": "",
					  "skipUrlSync": false,
					  "sort": 0,
					  "type": "query"
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
				"title": "OpenStack / VMs Network Traffic",
				"version": 36,
				"weekStart": ""
			  }`,
		},
	}

	return dashboardCM
}
