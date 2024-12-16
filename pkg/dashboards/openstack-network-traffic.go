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
                "links": [],
                "rows": [
                    {
                        "collapse": false,
                        "collapsed": false,
                        "repeat": null,
                        "repeatIteration": null,
                        "repeatRowId": null,
                        "showTitle": true,
                        "title": "Overview",
                        "titleSize": "h6",
                        "type": "row",
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
                                    "alertThreshold": true
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
                                        "expr": "vm:ceilometer_network_incoming_bytes:rate1m{project =~ \"$project\", vm_name =~ \"$VM\"}",
                                        "legendFormat": "{{vm_name}} in ({{device}})",
                                        "refId": "B"
                                    },
                                    {
                                        "expr": "vm:ceilometer_network_outgoing_bytes:rate1m{project =~ \"$project\", vm_name =~ \"$VM\"}",
                                        "legendFormat": "{{vm_name}} out ({{device}})",
                                        "refId": "A"
                                    }
                                ],
                                "thresholds": [],
                                "timeFrom": null,
                                "timeRegions": [],
                                "timeShift": null,
                                "title": "Network Adapter",
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
                        "title": "Inbound Traffic",
                        "titleSize": "h6",
                        "type": "row",
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
                                    "y": 0
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
                                        "expr": "vm:ceilometer_network_incoming_bytes:rate1m{resource_name=~\"$VM:$In_adapter\", project=~\"$project\" } / 1000000",
                                        "legendFormat": "__auto",
                                        "editorMode": "code",
                                        "range": true,
                                        "refId": "A"
                                    }
                                ],
                                "thresholds": [],
                                "timeFrom": null,
                                "timeRegions": [],
                                "timeShift": null,
                                "title": "Network Incoming Packets",
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
                                    "y": 0
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
                                        "expr": "vm:ceilometer_network_incoming_packets_drop:rate1m{project =~ \"$project\",vm_name =~ \"$VM\", device =~\"$In_adapter\" }",
                                        "legendFormat": "__auto",
                                        "editorMode": "code",
                                        "range": true,
                                        "refId": "A"
                                    }
                                ],
                                "thresholds": [],
                                "timeFrom": null,
                                "timeRegions": [],
                                "timeShift": null,
                                "title": "Incoming Dropped Packets",
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
                                    "y": 0
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
                                        "expr": "(rate(ceilometer_network_incoming_packets_drop{project=~\"$project\", resource_name=~\"$VM:$In_adapter\"}[1m]) / rate(ceilometer_network_incoming_packets{project=~\"$project\", resource_name=~\"$VM:In_adapter\"}[1m])) * 100\n",
                                        "legendFormat": "__auto",
                                        "editorMode": "code",
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
                                "timeFrom": null,
                                "timeRegions": [],
                                "timeShift": null,
                                "title": "Incoming Packet Loss (%)",
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
                                    "y": 0
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
                                        "expr": "vm:ceilometer_network_incoming_packets_error:rate1m{project =~ \"$project\",vm_name =~ \"$VM\", device =~\"$In_adapter\"}",
                                        "legendFormat": "__auto",
                                        "editorMode": "code",
                                        "range": true,
                                        "refId": "A"
                                    }
                                ],
                                "thresholds": [],
                                "timeFrom": null,
                                "timeRegions": [],
                                "timeShift": null,
                                "title": "Incoming Packets Error",
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
                        "title": "Outbound Traffic",
                        "titleSize": "h6",
                        "type": "row",
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
                                    "y": 0
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
                                        "expr": "vm:ceilometer_network_outgoing_bytes:rate1m{resource_name=~\"$VM:$In_adapter\", project=~\"$project\" } / 1000000",
                                        "legendFormat": "__auto",
                                        "editorMode": "code",
                                        "range": true,
                                        "refId": "A"
                                    }
                                ],
                                "thresholds": [],
                                "timeFrom": null,
                                "timeRegions": [],
                                "timeShift": null,
                                "title": "Network Outgoing Packets",
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
                                    "y": 0
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
                                        "expr": "vm:ceilometer_network_outgoing_packets_drop:rate1m{project =~ \"$project\",vm_name =~ \"$VM\", device =~\"$In_adapter\" }",
                                        "legendFormat": "__auto",
                                        "editorMode": "code",
                                        "range": true,
                                        "refId": "A"
                                    }
                                ],
                                "thresholds": [],
                                "timeFrom": null,
                                "timeRegions": [],
                                "timeShift": null,
                                "title": "Outgoing Dropped Packets",
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
                                    "y": 0
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
                                        "expr": "(rate(ceilometer_network_outgoing_packets_drop{project=~\"$project\", resource_name=~\"$VM:$In_adapter\"}[1m]) / rate(ceilometer_network_incoming_packets{project=~\"$project\", resource_name=~\"$VM:In_adapter\"}[1m])) * 100\n",
                                        "legendFormat": "__auto",
                                        "editorMode": "code",
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
                                "timeFrom": null,
                                "timeRegions": [],
                                "timeShift": null,
                                "title": "Outgoing Packet Loss (%)",
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
                                    "y": 0
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
                                        "expr": "vm:ceilometer_network_outgoing_packets_error:rate1m{project =~ \"$project\",vm_name =~ \"$VM\", device =~\"$In_adapter\"}",
                                        "legendFormat": "__auto",
                                        "editorMode": "code",
                                        "range": true,
                                        "refId": "A"
                                    }
                                ],
                                "thresholds": [],
                                "timeFrom": null,
                                "timeRegions": [],
                                "timeShift": null,
                                "title": "Outgoing Packets Error",
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
                            "allValue": ".*",
                            "current": {
                                "text": "",
                                "value": ""
                            },
                            "datasource": {
                                "name": "` + dsName + `",
                                "type": "prometheus"
                            },
                            "definition": "label_values(ceilometer_cpu, vm_instance)",
                            "hide": 0,
                            "includeAll": true,
                            "label": null,
                            "multi": true,
                            "name": "Compute_node",
                            "options": [],
                            "query": "label_values(vm:ceilometer_cpu:ratio1m, vm_instance)",
                            "refresh": 0,
                            "regex": "",
                            "skipUrlSync": false,
                            "sort": 0,
                            "tagValuesQuery": "",
                            "tags": [],
                            "tagsQuery": "",
                            "type": "query",
                            "useTags": false
                        },
                        {
                            "allValue": ".*",
                            "current": {
                            "tags": [],
                            "text": "539c3dc2361f4fd191aaa21c14360e35",
                            "value": [
                                "539c3dc2361f4fd191aaa21c14360e35"
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
                            "options": [],
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
                                "text": "",
                                "value": ""
                            },
                            "datasource": {
                                "name": "` + dsName + `",
                                "type": "prometheus"
                            },
                            "definition": "label_values(ceilometer_cpu{project =~ \"$project\"}, vm_instance)",
                            "hide": 0,
                            "includeAll": true,
                            "label": null,
                            "multi": true,
                            "name": "VM",
                            "options": [],
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
                        },
                        {
                            "allValue": ".*",
                            "current": {
                                "text": "",
                                "value": ""
                            },
                            "datasource": {
                                "name": "` + dsName + `",
                                "type": "prometheus"
                            },
                            "definition": "label_values(ceilometer_network_incoming_bytes{vm_name=~\"$VM\"}, device)",
                            "hide": 0,
                            "includeAll": true,
                            "label": null,
                            "multi": true,
                            "name": "In_adapter",
                            "options": [],
                            "query": "label_values(ceilometer_network_incoming_bytes{vm_name=~\"$VM\"}, device)",
                            "refresh": 0,
                            "regex": "",
                            "skipUrlSync": false,
                            "sort": 0,
                            "tagValuesQuery": "",
                            "tags": [],
                            "tagsQuery": "",
                            "type": "query",
                            "useTags": false
                        },
                        {
                            "allValue": ".*",
                            "current": {
                                "text": "",
                                "value": ""
                            },
                            "datasource": {
                                "name": "` + dsName + `",
                                "type": "prometheus"
                            },
                            "definition": "label_values(ceilometer_network_outgoing_bytes{vm_name=~\"$VM\"}, device)",
                            "hide": 0,
                            "includeAll": true,
                            "label": null,
                            "multi": true,
                            "name": "Out_adapter",
                            "options": [],
                            "query": "label_values(ceilometer_network_outgoing_bytes{vm_name=~\"$VM\"}, device)",
                            "refresh": 1,
                            "regex": "",
                            "skipUrlSync": false,
                            "sort": 1,
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
                "title": "OpenStack / VMs Network Traffic",
                "version": 7
            }`,
		},
	}

	return dashboardCM
}