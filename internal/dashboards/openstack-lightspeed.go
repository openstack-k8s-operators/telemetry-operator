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

// OpenstackLightspeed creates a ConfigMap containing the OpenStack Lightspeed monitoring dashboard
func OpenstackLightspeed(dsName string) *corev1.ConfigMap {
	dashboardCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "grafana-dashboard-openstack-lightspeed",
			Namespace: "openshift-config-managed",
			Labels: map[string]string{
				"console.openshift.io/dashboard": "true",
			},
		},
		Data: map[string]string{
			"openstack-lightspeed.json": `
			{
				"annotations": {
					"list": []
				},
				"editable": false,
				"gnetId": null,
				"graphTooltip": 1,
				"id": null,
				"links": [],
				"rows": [
					{
						"collapsed": false,
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
						"id": 1,
						"panels": [
							{
								"cacheTimeout": null,
								"datasource": {
									"name": "` + dsName + `",
									"type": "prometheus"
								},
								"description": "Total REST API calls per second",
								"gridPos": {
									"h": 6,
									"w": 6,
									"x": 0,
									"y": 1
								},
								"id": 2,
								"links": [],
								"options": {
									"colorMode": "value",
									"fieldOptions": {
										"calcs": ["lastNotNull"],
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
										"expr": "sum(rate(ls_rest_api_calls_total[5m]))",
										"interval": "",
										"legendFormat": "",
										"refId": "A"
									}
								],
								"timeFrom": null,
								"timeShift": null,
								"title": "API Calls / s",
								"span": 6,
								"type": "singlestat"
							},
							{
								"cacheTimeout": null,
								"datasource": {
									"name": "` + dsName + `",
									"type": "prometheus"
								},
								"description": "Service degraded mode status",
								"gridPos": {
									"h": 6,
									"w": 6,
									"x": 6,
									"y": 1
								},
								"id": 3,
								"links": [],
								"options": {
									"colorMode": "background",
									"fieldOptions": {
										"calcs": ["lastNotNull"],
										"defaults": {
											"mappings": [
												{
													"text": "Normal",
													"value": "0"
												},
												{
													"text": "Degraded",
													"value": "1"
												}
											],
											"thresholds": {
												"mode": "absolute",
												"steps": [
													{
														"color": "green",
														"value": null
													},
													{
														"color": "red",
														"value": 1
													}
												]
											}
										},
										"overrides": [],
										"values": false
									},
									"graphMode": "none",
									"justifyMode": "auto",
									"orientation": "auto"
								},
								"pluginVersion": "6.7.6",
								"targets": [
									{
										"expr": "ls_started_in_degraded_mode",
										"interval": "",
										"legendFormat": "",
										"refId": "A"
									}
								],
								"timeFrom": null,
								"timeShift": null,
								"title": "Service Status",
								"span": 6,
								"type": "singlestat"
							},
							{
								"cacheTimeout": null,
								"datasource": {
									"name": "` + dsName + `",
									"type": "prometheus"
								},
								"description": "Active provider model configuration",
								"gridPos": {
									"h": 6,
									"w": 6,
									"x": 12,
									"y": 1
								},
								"id": 4,
								"links": [],
								"options": {
									"colorMode": "value",
									"fieldOptions": {
										"calcs": ["lastNotNull"],
										"defaults": {
											"mappings": [],
											"thresholds": {
												"mode": "absolute",
												"steps": [
													{
														"color": "blue",
														"value": null
													}
												]
											}
										},
										"overrides": [],
										"values": false
									},
									"graphMode": "none",
									"justifyMode": "auto",
									"orientation": "auto"
								},
								"pluginVersion": "6.7.6",
								"targets": [
									{
										"expr": "ls_provider_model_configuration",
										"interval": "",
										"legendFormat": "",
										"refId": "A"
									}
								],
								"timeFrom": null,
								"timeShift": null,
								"title": "Model Configuration",
								"span": 6,
								"type": "singlestat"
							},
							{
								"cacheTimeout": null,
								"datasource": {
									"name": "` + dsName + `",
									"type": "prometheus"
								},
								"description": "LLM call failure rate",
								"gridPos": {
									"h": 6,
									"w": 6,
									"x": 18,
									"y": 1
								},
								"id": 5,
								"links": [],
								"options": {
									"colorMode": "value",
									"fieldOptions": {
										"calcs": ["lastNotNull"],
										"defaults": {
											"mappings": [],
											"thresholds": {
												"mode": "absolute",
												"steps": [
													{
														"color": "green",
														"value": null
													},
													{
														"color": "yellow",
														"value": 0.01
													},
													{
														"color": "red",
														"value": 0.1
													}
												]
											},
											"unit": "percentunit"
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
										"expr": "sum(rate(ls_llm_calls_failures_total[5m])) / sum(rate(ls_llm_calls_total[5m]))",
										"interval": "",
										"legendFormat": "",
										"refId": "A"
									}
								],
								"timeFrom": null,
								"timeShift": null,
								"title": "LLM Failure Rate",
								"span": 6,
								"type": "singlestat"
							}
						],
						"repeat": null,
						"repeatIteration": null,
						"repeatRowId": null,
						"showTitle": true,
						"title": "Overview",
						"titleSize": "h6",
						"type": "row"
					},
					{
						"collapsed": false,
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
						"id": 6,
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
								"description": "Response time percentiles",
								"fill": 1,
								"fillGradient": 0,
								"gridPos": {
									"h": 8,
									"w": 12,
									"x": 0,
									"y": 2
								},
								"hiddenSeries": false,
								"id": 7,
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
										"expr": "histogram_quantile(0.50, sum(rate(ls_response_duration_seconds_bucket[5m])) by (le))",
										"interval": "",
										"legendFormat": "p50",
										"refId": "A"
									},
									{
										"expr": "histogram_quantile(0.90, sum(rate(ls_response_duration_seconds_bucket[5m])) by (le))",
										"interval": "",
										"legendFormat": "p90",
										"refId": "B"
									},
									{
										"expr": "histogram_quantile(0.95, sum(rate(ls_response_duration_seconds_bucket[5m])) by (le))",
										"interval": "",
										"legendFormat": "p95",
										"refId": "C"
									},
									{
										"expr": "histogram_quantile(0.99, sum(rate(ls_response_duration_seconds_bucket[5m])) by (le))",
										"interval": "",
										"legendFormat": "p99",
										"refId": "D"
									}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "Response Time Percentiles",
								"span": 12,
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
										"format": "s",
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
								"description": "Average response time",
								"fill": 10,
								"fillGradient": 0,
								"gridPos": {
									"h": 8,
									"w": 12,
									"x": 12,
									"y": 2
								},
								"hiddenSeries": false,
								"id": 8,
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
										"expr": "rate(ls_response_duration_seconds_sum[5m]) / rate(ls_response_duration_seconds_count[5m])",
										"interval": "",
										"legendFormat": "avg response time",
										"refId": "A"
									}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "Average Response Time",
								"span": 12,
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
										"format": "s",
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
						"repeat": null,
						"repeatIteration": null,
						"repeatRowId": null,
						"showTitle": true,
						"title": "Response Times",
						"titleSize": "h6",
						"type": "row"
					},
					{
						"collapsed": false,
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
						"id": 9,
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
								"description": "LLM calls per second",
								"fill": 10,
								"fillGradient": 0,
								"gridPos": {
									"h": 8,
									"w": 12,
									"x": 0,
									"y": 3
								},
								"hiddenSeries": false,
								"id": 10,
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
										"expr": "sum(rate(ls_llm_calls_total[5m]))",
										"interval": "",
										"legendFormat": "total",
										"refId": "A"
									}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "LLM Calls Rate",
								"span": 12,
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
										"label": "calls/s",
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
								"description": "Token usage rates",
								"fill": 1,
								"fillGradient": 0,
								"gridPos": {
									"h": 8,
									"w": 12,
									"x": 12,
									"y": 3
								},
								"hiddenSeries": false,
								"id": 11,
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
										"expr": "sum(rate(ls_llm_token_sent_total[5m]))",
										"interval": "",
										"legendFormat": "sent",
										"refId": "A"
									},
									{
										"expr": "sum(rate(ls_llm_token_received_total[5m]))",
										"interval": "",
										"legendFormat": "received",
										"refId": "B"
									}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "Token Rate",
								"span": 12,
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
										"label": "tokens/s",
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
						"repeat": null,
						"repeatIteration": null,
						"repeatRowId": null,
						"showTitle": true,
						"title": "LLM Metrics",
						"titleSize": "h6",
						"type": "row"
					},
					{
						"collapsed": false,
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
						"id": 12,
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
								"description": "LLM call failures per second",
								"fill": 10,
								"fillGradient": 0,
								"gridPos": {
									"h": 8,
									"w": 12,
									"x": 0,
									"y": 4
								},
								"hiddenSeries": false,
								"id": 13,
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
										"expr": "sum(rate(ls_llm_calls_failures_total[5m]))",
										"interval": "",
										"legendFormat": "failures",
										"refId": "A"
									}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "LLM Failures Rate",
								"span": 12,
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
										"label": "failures/s",
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
								"description": "LLM validation errors per second",
								"fill": 10,
								"fillGradient": 0,
								"gridPos": {
									"h": 8,
									"w": 12,
									"x": 12,
									"y": 4
								},
								"hiddenSeries": false,
								"id": 14,
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
										"expr": "sum(rate(ls_llm_validation_errors_total[5m]))",
										"interval": "",
										"legendFormat": "validation errors",
										"refId": "A"
									}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "Validation Errors Rate",
								"span": 12,
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
										"label": "errors/s",
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
						"repeat": null,
						"repeatIteration": null,
						"repeatRowId": null,
						"showTitle": true,
						"title": "Errors",
						"titleSize": "h6",
						"type": "row"
					}
				],
				"refresh": "30s",
				"schemaVersion": 22,
				"style": "dark",
				"tags": [
					"openstack-telemetry-operator",
					"lightspeed"
				],
				"time": {
					"from": "now-1h",
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
				"timezone": "utc",
				"title": "OpenStack / Lightspeed",
				"uid": "lightspeed-monitoring",
				"version": 1
			}`,
		},
	}

	return dashboardCM
}
