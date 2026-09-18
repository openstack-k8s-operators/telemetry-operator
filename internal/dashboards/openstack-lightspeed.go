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
								"description": "Total REST API calls",
								"gridPos": {
									"h": 6,
									"w": 12,
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
								"format": "none",
								"decimals": 0,
								"targets": [
									{
										"expr": "round(sum(increase(ls_rest_api_calls_total[$__range])))",
										"interval": "",
										"legendFormat": "",
										"refId": "A"
									}
								],
								"timeFrom": null,
								"timeShift": null,
								"title": "Total API Calls",
								"span": 6,
								"type": "singlestat"
							},
							{
								"cacheTimeout": null,
								"datasource": {
									"name": "` + dsName + `",
									"type": "prometheus"
								},
								"description": "Total LLM failures",
								"gridPos": {
									"h": 6,
									"w": 12,
									"x": 12,
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
														"value": 5
													},
													{
														"color": "red",
														"value": 10
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
								"format": "none",
								"decimals": 0,
								"targets": [
									{
										"expr": "round(sum(increase(ls_llm_calls_failures_total[$__range])))",
										"interval": "",
										"legendFormat": "",
										"refId": "A"
									}
								],
								"timeFrom": null,
								"timeShift": null,
								"title": "Total LLM Failures",
								"span": 6,
								"type": "singlestat"
							},
							{
								"cacheTimeout": null,
								"datasource": {
									"name": "` + dsName + `",
									"type": "prometheus"
								},
								"description": "Configured LLM provider and model combinations",
								"gridPos": {
									"h": 6,
									"w": 24,
									"x": 0,
									"y": 7
								},
								"id": 4,
								"links": [],
								"options": {
									"showHeader": true
								},
								"pluginVersion": "6.7.6",
								"targets": [
									{
										"expr": "ls_provider_model_configuration",
										"instant": true,
										"format": "table",
										"legendFormat": "{{provider}} / {{model}}",
										"refId": "A"
									}
								],
								"timeFrom": null,
								"timeShift": null,
								"title": "Model Configuration",
								"span": 24,
								"type": "table",
								"transform": "table",
								"styles": [
									{
										"pattern": "Time",
										"type": "hidden"
									},
									{
										"pattern": "Value",
										"alias": "Is default (1=yes)",
										"type": "number",
										"decimals": 0
									},
									{
										"pattern": "provider",
										"alias": "Provider",
										"type": "string"
									},
									{
										"pattern": "model",
										"alias": "Model",
										"type": "string"
									}
								]
							},
							{
								"cacheTimeout": null,
								"datasource": {
									"name": "` + dsName + `",
									"type": "prometheus"
								},
								"description": "Service health status",
								"gridPos": {
									"h": 6,
									"w": 24,
									"x": 0,
									"y": 13
								},
								"id": 3,
								"links": [],
								"type": "singlestat",
								"valueName": "current",
								"mappingType": 1,
								"valueMaps": [
									{
										"op": "=",
										"text": "Healthy",
										"value": "0"
									},
									{
										"op": "=",
										"text": "Degraded",
										"value": "1"
									}
								],
								"rangeMaps": [
									{
										"from": "null",
										"to": "null",
										"text": "N/A"
									}
								],
								"thresholds": "0.5,1",
								"colors": [
									"rgba(50, 172, 45, 0.97)",
									"rgba(237, 129, 40, 0.89)",
									"rgba(245, 54, 54, 0.9)"
								],
								"colorBackground": true,
								"colorValue": false,
								"format": "none",
								"decimals": 0,
								"gauge": {
									"show": false
								},
								"sparkline": {
									"show": false
								},
								"targets": [
									{
										"expr": "ls_started_in_degraded_mode",
										"refId": "A"
									}
								],
								"timeFrom": null,
								"timeShift": null,
								"title": "Service Status",
								"span": 24
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
								"description": "LLM calls per minute",
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
										"expr": "sum(rate(ls_llm_calls_total[5m])) * 60",
										"interval": "",
										"legendFormat": "total",
										"refId": "A"
									}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "LLM Calls per Minute",
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
										"label": "calls/min",
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
								"title": "Token per Second",
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
								"description": "LLM call failures per minute",
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
										"expr": "sum(rate(ls_llm_calls_failures_total[5m])) * 60",
										"interval": "",
										"legendFormat": "failures",
										"refId": "A"
									}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "LLM Failures per Minute",
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
										"label": "failures/min",
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
								"description": "LLM validation errors per minute",
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
										"expr": "sum(rate(ls_llm_validation_errors_total[5m])) * 60",
										"interval": "",
										"legendFormat": "validation errors",
										"refId": "A"
									}
								],
								"thresholds": [],
								"timeFrom": null,
								"timeRegions": [],
								"timeShift": null,
								"title": "Validation Errors per Minute",
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
										"label": "errors/min",
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
