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

package metricstorage

import (
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	monv1 "github.com/rhobs/obo-prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func DashboardPrometheusRule(
	instance *telemetryv1.MetricStorage,
	labels map[string]string,
) *monv1.PrometheusRule {

	prometheusRule := &monv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: monv1.PrometheusRuleSpec{
			Groups: []monv1.RuleGroup{
				{
					Name: "osp-node-exporter-dashboard.rules",
					Rules: []monv1.Rule{
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `count without (cpu, mode) (node_cpu_seconds_total{mode="idle"})`,
							},
							Record: "instance:node_num_cpu:sum",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `1 - avg without (cpu) (sum without (mode) (rate(node_cpu_seconds_total{mode=~"idle|iowait|steal"}[1m])))`,
							},
							Record: "instance:node_cpu_utilisation:rate1m",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `(node_load1 / instance:node_num_cpu:sum)`,
							},
							Record: "instance:node_load1_per_cpu:ratio",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `1 - ( ( node_memory_MemAvailable_bytes or ( node_memory_Buffers_bytes + node_memory_Cached_bytes + node_memory_MemFree_bytes + node_memory_Slab_bytes ) ) / node_memory_MemTotal_bytes )`,
							},
							Record: "instance:node_memory_utilisation:ratio",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `rate(node_vmstat_pgmajfault[1m])`,
							},
							Record: "instance:node_vmstat_pgmajfault:rate1m",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `rate(node_disk_io_time_seconds_total{device=~"mmcblk.p.+|nvme.+|sd.+|vd.+|xvd.+|dm-.+|dasd.+"}[1m])`,
							},
							Record: "instance_device:node_disk_io_time_seconds:rate1m",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `rate(node_disk_io_time_weighted_seconds_total{device=~"mmcblk.p.+|nvme.+|sd.+|vd.+|xvd.+|dm-.+|dasd.+"}[1m])`,
							},
							Record: "instance_device:node_disk_io_time_weighted_seconds:rate1m",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `sum without (device) ( rate(node_network_receive_bytes_total{device!="lo"}[1m]) )`,
							},
							Record: "instance:node_network_receive_bytes_excluding_lo:rate1m",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `sum without (device) ( rate(node_network_transmit_bytes_total{device!="lo"}[1m]) )`,
							},
							Record: "instance:node_network_transmit_bytes_excluding_lo:rate1m",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `sum without (device) ( rate(node_network_receive_drop_total{device!="lo"}[1m]) )`,
							},
							Record: "instance:node_network_receive_drop_excluding_lo:rate1m",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `sum without (device) ( rate(node_network_transmit_drop_total{device!="lo"}[1m]) )`,
							},
							Record: "instance:node_network_transmit_drop_excluding_lo:rate1m",
						},
					},
				},
				{
					Name: "osp-ceilometer-dashboard.rules",
					Rules: []monv1.Rule{
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `sum without(unit, type) (label_replace(rate(ceilometer_cpu[1m]), "vm_name", "$1", "resource_name", "(.*):.*") / 1000000000)`,
							},
							Record: "vm:ceilometer_cpu:ratio1m",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `sum without(type) (label_replace(ceilometer_memory_usage, "vm_name", "$1", "resource_name", "(.*):.*") * 1024 * 1024)`,
							},
							Record: "vm:ceilometer_memory_usage:total",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `sum without(type) (label_replace(label_replace(ceilometer_disk_device_usage, "device", "$1", "resource", ".*-.*-.*-.*-.*-(.*)"), "vm_name", "$1", "resource_name", "(.*):.*"))`,
							},
							Record: "vm:ceilometer_disk_device_usage:total",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `sum without(type) (label_replace(label_replace(rate(ceilometer_disk_device_read_bytes[1m]), "device", "$1", "resource", ".*-.*-.*-.*-.*-(.*)"), "vm_name", "$1", "resource_name", "(.*):.*"))`,
							},
							Record: "vm:ceilometer_disk_device_read_bytes:rate1m",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `sum without(type) (label_replace(label_replace(rate(ceilometer_disk_device_write_bytes[1m]), "device", "$1", "resource", ".*-.*-.*-.*-.*-(.*)"), "vm_name", "$1", "resource_name", "(.*):.*"))`,
							},
							Record: "vm:ceilometer_disk_device_write_bytes:rate1m",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `sum without(type) (label_replace(label_replace(rate(ceilometer_network_outgoing_bytes[1m]), "device", "$1", "resource_name", ".*:(.*)"), "vm_name", "$1", "resource_name", "(.*):.*"))`,
							},
							Record: "vm:ceilometer_network_outgoing_bytes:rate1m",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `sum without(type) (label_replace(label_replace(rate(ceilometer_network_incoming_bytes[1m]), "device", "$1", "resource_name", ".*:(.*)"), "vm_name", "$1", "resource_name", "(.*):.*"))`,
							},
							Record: "vm:ceilometer_network_incoming_bytes:rate1m",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `sum without(type) (label_replace(label_replace(rate(ceilometer_network_outgoing_packets_drop[1m]), "device", "$1", "resource_name", ".*:(.*)"), "vm_name", "$1", "resource_name", "(.*):.*"))`,
							},
							Record: "vm:ceilometer_network_outgoing_packets_drop:rate1m",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `sum without(type) (label_replace(label_replace(rate(ceilometer_network_incoming_packets_drop[1m]), "device", "$1", "resource_name", ".*:(.*)"), "vm_name", "$1", "resource_name", "(.*):.*"))`,
							},
							Record: "vm:ceilometer_network_incoming_packets_drop:rate1m",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `sum without(type) (label_replace(label_replace(rate(ceilometer_network_incoming_packets_error[1m]), "device", "$1", "resource_name", ".*:(.*)"), "vm_name", "$1", "resource_name", "(.*):.*"))`,
							},
							Record: "vm:ceilometer_network_incoming_packets_error:rate1m",
						},
						{
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: `sum without(type) (label_replace(label_replace(rate(ceilometer_network_outgoing_packets_error[1m]), "device", "$1", "resource_name", ".*:(.*)"), "vm_name", "$1", "resource_name", "(.*):.*"))`,
							},
							Record: "vm:ceilometer_network_outgoing_packets_error:rate1m",
						},
					},
				},
			},
		},
	}
	return prometheusRule
}
