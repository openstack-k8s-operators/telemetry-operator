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

func OpenstackCloud(dsName string) *corev1.ConfigMap {

	dashboardCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "grafana-dashboard-openstack-cloud",
			Namespace: "openshift-config-managed",
			Labels: map[string]string{
				"console.openshift.io/dashboard": "true",
			},
		},
		Data: map[string]string{
			"openstack-cloud.json": `
			{
				"__inputs": [

				],
				"__requires": [

				],
				"annotations": {
					"list": [

					]
				},
				"editable": false,
				"gnetId": null,
				"graphTooltip": 1,
				"hideControls": false,
				"id": null,
				"links": [

				],
				"refresh": "30s",
				"rows": [
					{
						"collapse": false,
						"collapsed": false,
						"panels": [
							{
								"aliasColors": {

								},
								"bars": false,
								"dashLength": 10,
								"dashes": false,
								"datasource":  {
									"name": "` + dsName + `",
									"type": "prometheus"
								},
								"fill": 10,
								"fillGradient": 0,
								"gridPos": {

								},
								"id": 2,
								"legend": {
									"alignAsTable": false,
									"avg": false,
									"current": false,
									"max": false,
									"min": false,
									"rightSide": false,
									"show": false,
									"sideWidth": null,
									"total": false,
									"values": false
								},
								"lines": true,
								"linewidth": 1,
								"links": [

								],
								"nullPointMode": "null",
								"percentage": false,
								"pointradius": 5,
								"points": false,
								"renderer": "flot",
								"repeat": null,
								"seriesOverrides": [

								],
								"spaceLength": 10,
								"span": 6,
								"stack": true,
								"steppedLine": false,
								"targets": [
									{
										"expr": "((\n  instance:node_cpu_utilisation:rate1m\n  *\n  instance:node_num_cpu:sum\n) )\n/ scalar(sum(instance:node_num_cpu:sum))\n",
										"format": "time_series",
										"intervalFactor": 2,
										"legendFormat": "{{ instance }}",
										"refId": "A"
									}
								],
								"thresholds": [

								],
								"timeFrom": null,
								"timeShift": null,
								"title": "CPU Utilisation",
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
									"values": [

									]
								},
								"yaxes": [
									{
										"format": "percentunit",
										"label": null,
										"logBase": 1,
										"max": null,
										"min": null,
										"show": true
									},
									{
										"format": "percentunit",
										"label": null,
										"logBase": 1,
										"max": null,
										"min": null,
										"show": true
									}
								]
							},
							{
								"aliasColors": {

								},
								"bars": false,
								"dashLength": 10,
								"dashes": false,
								"datasource":  {
									"name": "` + dsName + `",
									"type": "prometheus"
								},
								"fill": 10,
								"fillGradient": 0,
								"gridPos": {

								},
								"id": 3,
								"legend": {
									"alignAsTable": false,
									"avg": false,
									"current": false,
									"max": false,
									"min": false,
									"rightSide": false,
									"show": false,
									"sideWidth": null,
									"total": false,
									"values": false
								},
								"lines": true,
								"linewidth": 1,
								"links": [

								],
								"nullPointMode": "null",
								"percentage": false,
								"pointradius": 5,
								"points": false,
								"renderer": "flot",
								"repeat": null,
								"seriesOverrides": [

								],
								"spaceLength": 10,
								"span": 6,
								"stack": true,
								"steppedLine": false,
								"targets": [
									{
										"expr": "(\n  instance:node_load1_per_cpu:ratio\n  / scalar(count(instance:node_load1_per_cpu:ratio))\n) \n",
										"format": "time_series",
										"intervalFactor": 2,
										"legendFormat": "{{instance}}",
										"refId": "A"
									}
								],
								"thresholds": [

								],
								"timeFrom": null,
								"timeShift": null,
								"title": "CPU Saturation (Load1 per CPU)",
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
									"values": [

									]
								},
								"yaxes": [
									{
										"format": "percentunit",
										"label": null,
										"logBase": 1,
										"max": null,
										"min": null,
										"show": true
									},
									{
										"format": "percentunit",
										"label": null,
										"logBase": 1,
										"max": null,
										"min": null,
										"show": true
									}
								]
							}
						],
						"repeat": null,
						"repeatIteration": null,
						"repeatRowId": null,
						"showTitle": true,
						"title": "CPU",
						"titleSize": "h6",
						"type": "row"
					},
					{
						"collapse": false,
						"collapsed": false,
						"panels": [
							{
								"aliasColors": {

								},
								"bars": false,
								"dashLength": 10,
								"dashes": false,
								"datasource":  {
									"name": "` + dsName + `",
									"type": "prometheus"
								},
								"fill": 10,
								"fillGradient": 0,
								"gridPos": {

								},
								"id": 4,
								"legend": {
									"alignAsTable": false,
									"avg": false,
									"current": false,
									"max": false,
									"min": false,
									"rightSide": false,
									"show": false,
									"sideWidth": null,
									"total": false,
									"values": false
								},
								"lines": true,
								"linewidth": 1,
								"links": [

								],
								"nullPointMode": "null",
								"percentage": false,
								"pointradius": 5,
								"points": false,
								"renderer": "flot",
								"repeat": null,
								"seriesOverrides": [

								],
								"spaceLength": 10,
								"span": 6,
								"stack": true,
								"steppedLine": false,
								"targets": [
									{
										"expr": "(\n  instance:node_memory_utilisation:ratio\n  / scalar(count(instance:node_memory_utilisation:ratio))\n)\n",
										"format": "time_series",
										"intervalFactor": 2,
										"legendFormat": "{{instance}}",
										"refId": "A"
									}
								],
								"thresholds": [

								],
								"timeFrom": null,
								"timeShift": null,
								"title": "Memory Utilisation",
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
									"values": [

									]
								},
								"yaxes": [
									{
										"format": "percentunit",
										"label": null,
										"logBase": 1,
										"max": null,
										"min": null,
										"show": true
									},
									{
										"format": "percentunit",
										"label": null,
										"logBase": 1,
										"max": null,
										"min": null,
										"show": true
									}
								]
							},
							{
								"aliasColors": {

								},
								"bars": false,
								"dashLength": 10,
								"dashes": false,
								"datasource":  {
									"name": "` + dsName + `",
									"type": "prometheus"
								},
								"fill": 10,
								"fillGradient": 0,
								"gridPos": {

								},
								"id": 5,
								"legend": {
									"alignAsTable": false,
									"avg": false,
									"current": false,
									"max": false,
									"min": false,
									"rightSide": false,
									"show": false,
									"sideWidth": null,
									"total": false,
									"values": false
								},
								"lines": true,
								"linewidth": 1,
								"links": [

								],
								"nullPointMode": "null",
								"percentage": false,
								"pointradius": 5,
								"points": false,
								"renderer": "flot",
								"repeat": null,
								"seriesOverrides": [

								],
								"spaceLength": 10,
								"span": 6,
								"stack": true,
								"steppedLine": false,
								"targets": [
									{
										"expr": "instance:node_vmstat_pgmajfault:rate1m",
										"format": "time_series",
										"intervalFactor": 2,
										"legendFormat": "{{instance}}",
										"refId": "A"
									}
								],
								"thresholds": [

								],
								"timeFrom": null,
								"timeShift": null,
								"title": "Memory Saturation (Major Page Faults)",
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
									"values": [

									]
								},
								"yaxes": [
									{
										"format": "rds",
										"label": null,
										"logBase": 1,
										"max": null,
										"min": null,
										"show": true
									},
									{
										"format": "rds",
										"label": null,
										"logBase": 1,
										"max": null,
										"min": null,
										"show": true
									}
								]
							}
						],
						"repeat": null,
						"repeatIteration": null,
						"repeatRowId": null,
						"showTitle": true,
						"title": "Memory",
						"titleSize": "h6",
						"type": "row"
					},
					{
						"collapse": false,
						"collapsed": false,
						"panels": [
							{
								"aliasColors": {

								},
								"bars": false,
								"dashLength": 10,
								"dashes": false,
								"datasource":  {
									"name": "` + dsName + `",
									"type": "prometheus"
								},
								"fill": 10,
								"fillGradient": 0,
								"gridPos": {

								},
								"id": 6,
								"legend": {
									"alignAsTable": false,
									"avg": false,
									"current": false,
									"max": false,
									"min": false,
									"rightSide": false,
									"show": false,
									"sideWidth": null,
									"total": false,
									"values": false
								},
								"lines": true,
								"linewidth": 1,
								"links": [

								],
								"nullPointMode": "null",
								"percentage": false,
								"pointradius": 5,
								"points": false,
								"renderer": "flot",
								"repeat": null,
								"seriesOverrides": [
									{
										"alias": "/Receive/",
										"stack": "A"
									},
									{
										"alias": "/Transmit/",
										"stack": "B",
										"transform": "negative-Y"
									}
								],
								"spaceLength": 10,
								"span": 6,
								"stack": true,
								"steppedLine": false,
								"targets": [
									{
										"expr": "instance:node_network_receive_bytes_excluding_lo:rate1m",
										"format": "time_series",
										"intervalFactor": 2,
										"legendFormat": "{{instance}} Receive",
										"refId": "A"
									},
									{
										"expr": "instance:node_network_transmit_bytes_excluding_lo:rate1m",
										"format": "time_series",
										"intervalFactor": 2,
										"legendFormat": "{{instance}} Transmit",
										"refId": "B"
									}
								],
								"thresholds": [

								],
								"timeFrom": null,
								"timeShift": null,
								"title": "Network Utilisation (Bytes Receive/Transmit)",
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
									"values": [

									]
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
										"format": "Bps",
										"label": null,
										"logBase": 1,
										"max": null,
										"min": null,
										"show": true
									}
								]
							},
							{
								"aliasColors": {

								},
								"bars": false,
								"dashLength": 10,
								"dashes": false,
								"datasource":  {
									"name": "` + dsName + `",
									"type": "prometheus"
								},
								"fill": 10,
								"fillGradient": 0,
								"gridPos": {

								},
								"id": 7,
								"legend": {
									"alignAsTable": false,
									"avg": false,
									"current": false,
									"max": false,
									"min": false,
									"rightSide": false,
									"show": false,
									"sideWidth": null,
									"total": false,
									"values": false
								},
								"lines": true,
								"linewidth": 1,
								"links": [

								],
								"nullPointMode": "null",
								"percentage": false,
								"pointradius": 5,
								"points": false,
								"renderer": "flot",
								"repeat": null,
								"seriesOverrides": [
									{
										"alias": "/ Receive/",
										"stack": "A"
									},
									{
										"alias": "/ Transmit/",
										"stack": "B",
										"transform": "negative-Y"
									}
								],
								"spaceLength": 10,
								"span": 6,
								"stack": true,
								"steppedLine": false,
								"targets": [
									{
										"expr": "instance:node_network_receive_drop_excluding_lo:rate1m",
										"format": "time_series",
										"intervalFactor": 2,
										"legendFormat": "{{instance}} Receive",
										"refId": "A"
									},
									{
										"expr": "instance:node_network_transmit_drop_excluding_lo:rate1m",
										"format": "time_series",
										"intervalFactor": 2,
										"legendFormat": "{{instance}} Transmit",
										"refId": "B"
									}
								],
								"thresholds": [

								],
								"timeFrom": null,
								"timeShift": null,
								"title": "Network Saturation (Drops Receive/Transmit)",
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
									"values": [

									]
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
										"format": "Bps",
										"label": null,
										"logBase": 1,
										"max": null,
										"min": null,
										"show": true
									}
								]
							}
						],
						"repeat": null,
						"repeatIteration": null,
						"repeatRowId": null,
						"showTitle": true,
						"title": "Network",
						"titleSize": "h6",
						"type": "row"
					},
					{
						"collapse": false,
						"collapsed": false,
						"panels": [
							{
								"aliasColors": {

								},
								"bars": false,
								"dashLength": 10,
								"dashes": false,
								"datasource":  {
									"name": "` + dsName + `",
									"type": "prometheus"
								},
								"fill": 10,
								"fillGradient": 0,
								"gridPos": {

								},
								"id": 8,
								"legend": {
									"alignAsTable": false,
									"avg": false,
									"current": false,
									"max": false,
									"min": false,
									"rightSide": false,
									"show": false,
									"sideWidth": null,
									"total": false,
									"values": false
								},
								"lines": true,
								"linewidth": 1,
								"links": [

								],
								"nullPointMode": "null",
								"percentage": false,
								"pointradius": 5,
								"points": false,
								"renderer": "flot",
								"repeat": null,
								"seriesOverrides": [

								],
								"spaceLength": 10,
								"span": 6,
								"stack": true,
								"steppedLine": false,
								"targets": [
									{
										"expr": "(\n  instance_device:node_disk_io_time_seconds:rate1m\n  / scalar(count(instance_device:node_disk_io_time_seconds:rate1m))\n)\n",
										"format": "time_series",
										"intervalFactor": 2,
										"legendFormat": "{{instance}} {{device}}",
										"refId": "A"
									}
								],
								"thresholds": [

								],
								"timeFrom": null,
								"timeShift": null,
								"title": "Disk IO Utilisation",
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
									"values": [

									]
								},
								"yaxes": [
									{
										"format": "percentunit",
										"label": null,
										"logBase": 1,
										"max": null,
										"min": null,
										"show": true
									},
									{
										"format": "percentunit",
										"label": null,
										"logBase": 1,
										"max": null,
										"min": null,
										"show": true
									}
								]
							},
							{
								"aliasColors": {

								},
								"bars": false,
								"dashLength": 10,
								"dashes": false,
								"datasource":  {
									"name": "` + dsName + `",
									"type": "prometheus"
								},
								"fill": 10,
								"fillGradient": 0,
								"gridPos": {

								},
								"id": 9,
								"legend": {
									"alignAsTable": false,
									"avg": false,
									"current": false,
									"max": false,
									"min": false,
									"rightSide": false,
									"show": false,
									"sideWidth": null,
									"total": false,
									"values": false
								},
								"lines": true,
								"linewidth": 1,
								"links": [

								],
								"nullPointMode": "null",
								"percentage": false,
								"pointradius": 5,
								"points": false,
								"renderer": "flot",
								"repeat": null,
								"seriesOverrides": [

								],
								"spaceLength": 10,
								"span": 6,
								"stack": true,
								"steppedLine": false,
								"targets": [
									{
										"expr": "(\n  instance_device:node_disk_io_time_weighted_seconds:rate1m\n  / scalar(count(instance_device:node_disk_io_time_weighted_seconds:rate1m))\n)\n",
										"format": "time_series",
										"intervalFactor": 2,
										"legendFormat": "{{instance}} {{device}}",
										"refId": "A"
									}
								],
								"thresholds": [

								],
								"timeFrom": null,
								"timeShift": null,
								"title": "Disk IO Saturation",
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
									"values": [

									]
								},
								"yaxes": [
									{
										"format": "percentunit",
										"label": null,
										"logBase": 1,
										"max": null,
										"min": null,
										"show": true
									},
									{
										"format": "percentunit",
										"label": null,
										"logBase": 1,
										"max": null,
										"min": null,
										"show": true
									}
								]
							}
						],
						"repeat": null,
						"repeatIteration": null,
						"repeatRowId": null,
						"showTitle": true,
						"title": "Disk IO",
						"titleSize": "h6",
						"type": "row"
					},
					{
						"collapse": false,
						"collapsed": false,
						"panels": [
							{
								"aliasColors": {

								},
								"bars": false,
								"dashLength": 10,
								"dashes": false,
								"datasource":  {
									"name": "` + dsName + `",
									"type": "prometheus"
								},
								"fill": 10,
								"fillGradient": 0,
								"gridPos": {

								},
								"id": 10,
								"legend": {
									"alignAsTable": false,
									"avg": false,
									"current": false,
									"max": false,
									"min": false,
									"rightSide": false,
									"show": false,
									"sideWidth": null,
									"total": false,
									"values": false
								},
								"lines": true,
								"linewidth": 1,
								"links": [

								],
								"nullPointMode": "null",
								"percentage": false,
								"pointradius": 5,
								"points": false,
								"renderer": "flot",
								"repeat": null,
								"seriesOverrides": [

								],
								"spaceLength": 10,
								"span": 12,
								"stack": true,
								"steppedLine": false,
								"targets": [
									{
										"expr": "sum without (device) (\n  max without (fstype, mountpoint) ((\n    node_filesystem_size_bytes{fstype!=\"\", mountpoint!~\"/var/lib/ibmc-s3fs.*\"}\n    -\n    node_filesystem_avail_bytes{fstype!=\"\", mountpoint!~\"/var/lib/ibmc-s3fs.*\"}\n  ))\n)\n/ scalar(sum(max without (fstype, mountpoint) (node_filesystem_size_bytes{fstype!=\"\", mountpoint!~\"/var/lib/ibmc-s3fs.*\"})))\n",
										"format": "time_series",
										"intervalFactor": 2,
										"legendFormat": "{{instance}}",
										"refId": "A"
									}
								],
								"thresholds": [

								],
								"timeFrom": null,
								"timeShift": null,
								"title": "Disk Space Utilisation",
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
									"values": [

									]
								},
								"yaxes": [
									{
										"format": "percentunit",
										"label": null,
										"logBase": 1,
										"max": null,
										"min": null,
										"show": true
									},
									{
										"format": "percentunit",
										"label": null,
										"logBase": 1,
										"max": null,
										"min": null,
										"show": true
									}
								]
							}
						],
						"repeat": null,
						"repeatIteration": null,
						"repeatRowId": null,
						"showTitle": true,
						"title": "Disk Space",
						"titleSize": "h6",
						"type": "row"
					}
				],
				"schemaVersion": 14,
				"style": "dark",
				"tags": [
					"openstack-telemetry-operator"
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
				"title": "OpenStack / Node Exporter / USE Method / Cluster",
				"version": 0
			}
			`,
		},
	}

	return dashboardCM
}
