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

func OpenstackOpenstackNetwork(dsName string, hasDpdk bool) *corev1.ConfigMap {
	dashBoardOVN :=
		`{
        "__inputs": [],
        "__requires": [],
        "annotations": {
            "list": []
        },
        "editable": true,
        "gnetId": null,
        "graphTooltip": 0,
        "hideControls": false,
        "id": null,
        "links": [],
        "refresh": "30s",
        "rows":
        [
            {
                "collapse": false,
                "collapsed": false,
                "height": "300px",
                "panels":
                [
                    {
                        "cacheTimeout": null,
                        "datasource": {
                            "name": "` + dsName + `",
                            "type": "prometheus"
                        },
                        "description": "Integration Bridge flow count",
                        "id": 1,
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
                                "expr": "ovs_bridge_flow_count{bridge=\"br-int\",instance=\"$instance\"}",
                                "interval": "",
                                "legendFormat": "{{bridge}}",
                                "refId": "A"
                            }
                        ],
                        "timeFrom": null,
                        "timeShift": null,
                        "title": "Flow Count of Integration Bridge",
                        "span": 3,
                        "type": "singlestat"
                    },
                    {
                        "cacheTimeout": null,
                        "datasource": {
                            "name": "` + dsName + `",
                            "type": "prometheus"
                        },
                        "description": "Number of times ovn-controller has translated the Logical_Flow table in the OVN SB database into OpenFlow flow in the past time period.",
                        "id": 1,
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
                                "expr": "increase(ovnc_lflow_run{instance=\"$instance\"}[$__rate_interval])",
                                "interval": "",
                                "legendFormat": "",
                                "refId": "A"
                            }
                        ],
                        "timeFrom": null,
                        "timeShift": null,
                        "title": "Logical Flow Translations",
                        "span": 3,
                        "type": "singlestat"
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
                        },
                        "hiddenSeries": false,
                        "id": 20,
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
                        "stack": false,
                        "steppedLine": false,
                        "targets": [
                            {
                                "expr": "rate(ovs_coverage_netlink_sent_total{instance=\"$instance\"}[$__rate_interval])",
                                "legendFormat": "TX Pkts/s",
                                "editorMode": "code",
                                "range": true,
                                "refId": "A"
                            },
                            {
                                "expr": "rate(ovs_coverage_netlink_received_total{instance=\"$instance\"}[$__rate_interval])",
                                "legendFormat": "RX Pkts/s",
                                "editorMode": "code",
                                "range": true,
                                "refId": "A"
                            }
                        ],
                        "thresholds": [],
                        "timeFrom": null,
                        "timeRegions": [],
                        "timeShift": null,
                        "title": "Netlink",
                        "tooltip": {
                            "shared": true,
                            "sort": 0,
                            "value_type": "individual"
                        },
                        "type": "graph",
                        "xaxis": {
                            "mode": "time",
                            "name": null,
                            "show": true,
                            "values": []
                        },
                        "yaxes": [
                            {
                                "format": "pps",
                                "logBase": 1,
                                "show": true
                            },
                            {
                                "format": "pps",
                                "logBase": 1,
                                "show": true
                            }
                        ],
                        "yaxis": {
                            "align": false,
                            "alignLevel": null
                        },
                        "span": 3
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
                        },
                        "hiddenSeries": false,
                        "id": 21,
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
                        "stack": false,
                        "steppedLine": false,
                        "targets": [
                            {
                                "expr": "rate(ovs_pmd_datapath_passes{instance=\"$instance\"}[$__rate_interval]) / rate(ovs_pmd_rx_packets{instance=\"$instance\"}[$__rate_interval])",
                                "legendFormat": "cpu(numa) {{cpu}}({{numa}})",
                                "editorMode": "code",
                                "range": true,
                                "refId": "A"
                            }
                        ],
                        "thresholds": [],
                        "timeFrom": null,
                        "timeRegions": [],
                        "timeShift": null,
                        "title": "Recirculation Rate",
                        "tooltip": {
                            "shared": true,
                            "sort": 0,
                            "value_type": "individual"
                        },
                        "type": "graph",
                        "xaxis": {
                            "mode": "time",
                            "name": null,
                            "show": true,
                            "values": []
                        },
                        "yaxes": [
                            {
                                "format": "short",
                                "logBase": 1,
                                "show": true
                            },
                            {
                                "format": "short",
                                "logBase": 1,
                                "show": false
                            }
                        ],
                        "yaxis": {
                            "align": false,
                            "alignLevel": null
                        },
                        "span": 3
                    }
                ],
                "repeat": null,
                "repeatIteration": null,
                "repeatRowId": null,
                "showTitle": true,
                "title": "OVN/OVS",
                "titleSize": "h6",
                "type": "row"
            }`
	dashBoardDpdk := `
            {
                "collapse": false,
                "collapsed": false,
                "panels":
                [
                    {
                        "cacheTimeout": null,
                        "datasource": {
                            "name": "` + dsName + `",
                            "type": "prometheus"
                        },
                        "description": "Short term average microseconds / iterations",
                        "gridPos": {
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
                                "expr": "1000000 / irate(ovs_pmd_total_iterations{instance=\"$instance\"}[$__rate_interval])",
                                "interval": "",
                                "legendFormat": "cpu(numa) {{cpu}}({{numa}})",
                                "refId": "A"
                            }
                        ],
                        "timeFrom": null,
                        "timeShift": null,
                        "title": "uS per Iteration",
                        "span": 3,
                        "type": "singlestat"
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
                            "alertThreshold": true
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
                                "expr": "rate(ovs_pmd_total_upcalls{instance=\"$instance\"}[$__rate_interval])",
                                "legendFormat": "cpu(numa) {{cpu}}({{numa}})",
                                "editorMode": "code",
                                "range": true,
                                "refId": "A"
                            }
                        ],
                        "thresholds": [],
                        "timeFrom": null,
                        "timeRegions": [],
                        "timeShift": null,
                        "title": "Upcalls",
                        "tooltip": {
                            "shared": true,
                            "sort": 0,
                            "value_type": "individual"
                        },
                        "type": "graph",
                        "xaxis": {
                            "mode": "time",
                            "name": null,
                            "show": true,
                            "values": []
                        },
                        "yaxes": [
                            {
                                "format": "short",
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
                            "align": false,
                            "alignLevel": null
                        },
                        "span": 3
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
                            "alertThreshold": true
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
                                "expr": "ovs_pmd_nonvol_context_switches{instance=\"$instance\"}",
                                "legendFormat": "cpu(numa) {{cpu}}({{numa}})",
                                "editorMode": "code",
                                "range": true,
                                "refId": "A"
                            }
                        ],
                        "thresholds": [],
                        "timeFrom": null,
                        "timeRegions": [],
                        "timeShift": null,
                        "title": "Non-Voluntary Context Switches",
                        "tooltip": {
                            "shared": true,
                            "sort": 0,
                            "value_type": "individual"
                        },
                        "type": "graph",
                        "xaxis": {
                            "mode": "time",
                            "name": null,
                            "show": true,
                            "values": []
                        },
                        "yaxes": [
                            {
                                "format": "short",
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
                            "align": false,
                            "alignLevel": null
                        },
                        "span": 3
                    },
                    {
                      "datasource": {
                          "name": "` + dsName + `",
                          "type": "prometheus"
                      },
                      "description": "Ratio of busy iterations to idle iterations",
                      "fontSize": "100%",
                      "span": 3,
                      "id": 20,
                      "pageSize": null,
                      "showHeader": true,
                      "sort": {
                        "col": 0,
                        "desc": true
                      },
                      "styles": [
                        {
                          "alias": "Time",
                          "align": "auto",
                          "dateFormat": "YYYY-MM-DD HH:mm:ss",
                          "pattern": "Time",
                          "type": "hidden"
                        },
                        {
                          "alias": "CPU",
                          "align": "",
                          "colorMode": null,
                          "colors": [
                          ],
                          "dateFormat": "YYYY-MM-DD HH:mm:ss",
                          "decimals": 0,
                          "mappingType": 1,
                          "pattern": "cpu",
                          "thresholds": [],
                          "type": "number",
                          "unit": "short"
                        },
                        {
                          "alias": "Numa",
                          "align": "auto",
                          "colorMode": null,
                          "colors": [
                          ],
                          "dateFormat": "YYYY-MM-DD HH:mm:ss",
                          "decimals": 2,
                          "mappingType": 1,
                          "pattern": "numa",
                          "thresholds": [],
                          "type": "number",
                          "unit": "short"
                        },
                        {
                          "alias": "fqdn",
                          "align": "right",
                          "colorMode": null,
                          "colors": [
                          ],
                          "decimals": 2,
                          "pattern": "fqdn",
                          "thresholds": [],
                          "type": "hidden",
                          "unit": "short"
                        },
                        {
                          "alias": "instance",
                          "align": "auto",
                          "colorMode": null,
                          "colors": [
                          ],
                          "dateFormat": "YYYY-MM-DD HH:mm:ss",
                          "decimals": 2,
                          "mappingType": 1,
                          "pattern": "instance",
                          "thresholds": [],
                          "type": "hidden",
                          "unit": "short"
                        },
                        {
                          "alias": "% Busy",
                          "align": "auto",
                          "colorMode": null,
                          "colors": [
                          ],
                          "dateFormat": "YYYY-MM-DD HH:mm:ss",
                          "decimals": 2,
                          "mappingType": 1,
                          "pattern": "Value",
                          "thresholds": [],
                          "type": "number",
                          "unit": "percent"
                        }
                      ],
                      "targets": [
                        {
                          "expr": "(ovs_pmd_busy_iterations{instance=\"$instance\"} / (ovs_pmd_busy_iterations{instance=\"$instance\"} + ovs_pmd_idle_iterations{instance=\"$instance\"}) * 100)",
                          "format": "table",
                          "instant": true,
                          "interval": "",
                          "legendFormat": "",
                          "refId": "A"
                        }
                      ],
                      "timeFrom": null,
                      "timeShift": null,
                      "title": "Utilization",
                      "transform": "table",
                      "type": "table"
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
                        },
                        "hiddenSeries": false,
                        "id": 3,
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
                        "stack": false,
                        "steppedLine": false,
                        "targets": [
                            {
                                "expr": "rate(ovs_pmd_tx_packets{instance=\"$instance\"}[$__rate_interval])",
                                "legendFormat": "cpu(numa) {{cpu}}({{numa}})",
                                "editorMode": "code",
                                "range": true,
                                "refId": "A"
                            }
                        ],
                        "thresholds": [],
                        "timeFrom": null,
                        "timeRegions": [],
                        "timeShift": null,
                        "title": "TX Packets",
                        "tooltip": {
                            "shared": true,
                            "sort": 0,
                            "value_type": "individual"
                        },
                        "type": "graph",
                        "xaxis": {
                            "mode": "time",
                            "name": null,
                            "show": true,
                            "values": []
                        },
                        "yaxes": [
                            {
                                "format": "pps",
                                "logBase": 1,
                                "show": true
                            },
                            {
                                "format": "pps",
                                "logBase": 1,
                                "show": true
                            }
                        ],
                        "yaxis": {
                            "align": false,
                            "alignLevel": null
                        },
                        "span": 12
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
                        "pointradius": 2,
                        "points": false,
                        "renderer": "flot",
                        "seriesOverrides": [],
                        "spaceLength": 10,
                        "stack": false,
                        "steppedLine": false,
                        "targets": [
                            {
                                "expr": "rate(ovs_pmd_rx_packets{instance=\"$instance\"}[$__rate_interval])",
                                "legendFormat": "cpu(numa) {{cpu}}({{numa}})",
                                "editorMode": "code",
                                "range": true,
                                "refId": "A"
                            }
                        ],
                        "thresholds": [],
                        "timeFrom": null,
                        "timeRegions": [],
                        "timeShift": null,
                        "title": "RX Packets",
                        "tooltip": {
                            "shared": true,
                            "sort": 0,
                            "value_type": "individual"
                        },
                        "type": "graph",
                        "xaxis": {
                            "mode": "time",
                            "name": null,
                            "show": true,
                            "values": []
                        },
                        "yaxes": [
                            {
                                "format": "pps",
                                "logBase": 1,
                                "show": true
                            },
                            {
                                "format": "pps",
                                "logBase": 1,
                                "show": true
                            }
                        ],
                        "yaxis": {
                            "align": false,
                            "alignLevel": null
                        },
                        "span": 12
                    }
                ],
                "repeat": null,
                "repeatIteration": null,
                "repeatRowId": null,
                "showTitle": true,
                "title": "PMD",
                "titleSize": "h6",
                "type": "row"
            },
            {
                "collapse": false,
                "collapsed": false,
                "panels":
                [
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
                        },
                        "hiddenSeries": false,
                        "id": 5,
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
                        "stack": false,
                        "steppedLine": false,
                        "targets": [
                            {
                                "expr": "rate(ovs_interface_tx_packets{type=\"dpdk\",instance=\"$instance\"}[$__rate_interval])",
                                "legendFormat": "{{bridge}}({{interface}})",
                                "editorMode": "code",
                                "range": true,
                                "refId": "A"
                            }
                        ],
                        "thresholds": [],
                        "timeFrom": null,
                        "timeRegions": [],
                        "timeShift": null,
                        "title": "TX Packets",
                        "tooltip": {
                            "shared": true,
                            "sort": 0,
                            "value_type": "individual"
                        },
                        "type": "graph",
                        "xaxis": {
                            "mode": "time",
                            "name": null,
                            "show": true,
                            "values": []
                        },
                        "yaxes": [
                            {
                                "format": "pps",
                                "logBase": 1,
                                "show": true
                            },
                            {
                                "format": "pps",
                                "logBase": 1,
                                "show": true
                            }
                        ],
                        "yaxis": {
                            "align": false,
                            "alignLevel": null
                        },
                        "span": 3
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
                        "stack": false,
                        "steppedLine": false,
                        "targets": [
                            {
                                "expr": "rate(ovs_interface_rx_packets{type=\"dpdk\",instance=\"$instance\"}[$__rate_interval])",
                                "legendFormat": "{{bridge}}({{interface}})",
                                "editorMode": "code",
                                "range": true,
                                "refId": "A"
                            }
                        ],
                        "thresholds": [],
                        "timeFrom": null,
                        "timeRegions": [],
                        "timeShift": null,
                        "title": "RX Packets",
                        "tooltip": {
                            "shared": true,
                            "sort": 0,
                            "value_type": "individual"
                        },
                        "type": "graph",
                        "xaxis": {
                            "mode": "time",
                            "name": null,
                            "show": true,
                            "values": []
                        },
                        "yaxes": [
                            {
                                "format": "pps",
                                "logBase": 1,
                                "show": true
                            },
                            {
                                "format": "pps",
                                "logBase": 1,
                                "show": true
                            }
                        ],
                        "yaxis": {
                            "align": false,
                            "alignLevel": null
                        },
                        "span": 3
                    }
                ],
                "repeat": null,
                "repeatIteration": null,
                "repeatRowId": null,
                "showTitle": true,
                "title": "DPDK",
                "titleSize": "h6",
                "type": "row"
            }`
	dashBaordFooter := `],
        "schemaVersion": 14,
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
                    "datasource": {
                        "name": "` + dsName + `",
                        "type": "prometheus"
                    },
                    "hide": 0,
                    "includeAll": false,
                    "label": null,
                    "multi": true,
                    "name": "instance",
                    "options": [],
                    "query": "label_values(ovs_build_info, instance)",
                    "refresh": 2,
                    "regex": "",
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
            ],
            "time_options": [
                "5m",
                "15m",
                "1h",
                "6h",
                "12h",
                "24h",
                "2d",
                "7d",
                "30d"
            ]
        },
        "timezone": "utc",
        "title": "Openstack / Openstack Network Exporter",
        "version": 0
    }`

	dashBoard := dashBoardOVN
	if hasDpdk {
		dashBoard += "," + dashBoardDpdk
	}
	dashBoard += dashBaordFooter

	dashboardCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "grafana-dashboard-openstack-openstack-network",
			Namespace: "openshift-config-managed",
			Labels: map[string]string{
				"console.openshift.io/dashboard": "true",
			},
		},
		Data: map[string]string{
			"openstack-openstack-network.json": dashBoard,
		},
	}

	return dashboardCM
}
