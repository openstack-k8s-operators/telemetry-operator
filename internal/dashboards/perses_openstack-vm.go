/*
Copyright 2026.

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
	persesv1alpha1 "github.com/perses/perses-operator/api/v1alpha1"
)

const openstackVmJSON = `{
  "apiVersion": "perses.dev/v1alpha2",
  "kind": "PersesDashboard",
  "metadata": {
    "name": "openstack-vm"
  },
  "spec": {
    "config": {
      "display": {
        "name": "OpenStack / Ceilometer / VMs"
      },
      "variables": [
        {
          "kind": "ListVariable",
          "spec": {
            "display": {
              "hidden": false
            },
            "allowAllValue": true,
            "allowMultiple": true,
            "customAllValue": ".*",
            "sort": "alphabetical-asc",
            "plugin": {
              "kind": "PrometheusLabelValuesVariable",
              "spec": {
                "labelName": "project_name",
                "matchers": [
                  "ceilometer_cpu"
                ]
              }
            },
            "name": "project"
          }
        },
        {
          "kind": "ListVariable",
          "spec": {
            "display": {
              "hidden": false
            },
            "allowAllValue": false,
            "allowMultiple": true,
            "sort": "none",
            "plugin": {
              "kind": "PrometheusLabelValuesVariable",
              "spec": {
                "labelName": "vm_name",
                "matchers": [
                  "vm:ceilometer_cpu:ratio1m{project_name =~ \"$project\"}"
                ]
              }
            },
            "name": "VM"
          }
        }
      ],
      "panels": {
        "0": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Power State"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
                "legend": {
                  "position": "bottom"
                }
              }
            },
            "queries": [
              {
                "kind": "TimeSeriesQuery",
                "spec": {
                  "plugin": {
                    "kind": "PrometheusTimeSeriesQuery",
                    "spec": {
                      "query": "label_replace(\n  ceilometer_power_state{resource_name=~\"$VM.*\"} == 7, \n  \"state\", \"suspended\", \"__name__\", \".*\"\n) or\nlabel_replace(\n  ceilometer_power_state{resource_name=~\"$VM.*\"} == 6, \n  \"state\", \"crashed\", \"__name__\", \".*\"\n) or\nlabel_replace(\n  ceilometer_power_state{resource_name=~\"$VM.*\"} == 5, \n  \"state\", \"shut off\", \"__name__\", \".*\"\n) or\nlabel_replace(\n  ceilometer_power_state{resource_name=~\"$VM.*\"} == 4, \n  \"state\", \"shut down\", \"__name__\", \".*\"\n) or\nlabel_replace(\n  ceilometer_power_state{resource_name=~\"$VM.*\"} == 3, \n  \"state\", \"paused\", \"__name__\", \".*\"\n) or\nlabel_replace(\n  ceilometer_power_state{resource_name=~\"$VM.*\"} == 2, \n  \"state\", \"blocked\", \"__name__\", \".*\"\n) or\nlabel_replace(\n  ceilometer_power_state{resource_name=~\"$VM.*\"} == 1, \n  \"state\", \"running\", \"__name__\", \".*\"\n)",
                      "seriesNameFormat": "{{resource_name}} is {{state}}"
                    }
                  }
                }
              }
            ]
          }
        },
        "1": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "CPU Utilisation"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
                "legend": {
                  "position": "bottom"
                },
                "yAxis": {
                  "format": {
                    "unit": "percent-decimal"
                  }
                }
              }
            },
            "queries": [
              {
                "kind": "TimeSeriesQuery",
                "spec": {
                  "plugin": {
                    "kind": "PrometheusTimeSeriesQuery",
                    "spec": {
                      "query": "vm:ceilometer_cpu:ratio1m{project_name =~ \"$project\", vm_name =~ \"$VM\"}",
                      "seriesNameFormat": "{{vm_name}}"
                    }
                  }
                }
              }
            ]
          }
        },
        "2": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Memory Utilisation"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
                "legend": {
                  "position": "bottom"
                },
                "yAxis": {
                  "format": {
                    "unit": "bytes"
                  }
                }
              }
            },
            "queries": [
              {
                "kind": "TimeSeriesQuery",
                "spec": {
                  "plugin": {
                    "kind": "PrometheusTimeSeriesQuery",
                    "spec": {
                      "query": "vm:ceilometer_memory_usage:total{project_name =~ \"$project\", vm_name =~ \"$VM\"}",
                      "seriesNameFormat": "{{vm_name}}"
                    }
                  }
                }
              }
            ]
          }
        },
        "3": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Disk Space Utilisation"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
                "legend": {
                  "position": "bottom"
                },
                "yAxis": {
                  "format": {
                    "unit": "bytes"
                  }
                }
              }
            },
            "queries": [
              {
                "kind": "TimeSeriesQuery",
                "spec": {
                  "plugin": {
                    "kind": "PrometheusTimeSeriesQuery",
                    "spec": {
                      "query": "vm:ceilometer_disk_device_usage:total{project_name =~ \"$project\", vm_name =~ \"$VM\"}",
                      "seriesNameFormat": "{{vm_name}}({{device}})"
                    }
                  }
                }
              }
            ]
          }
        },
        "4": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Disk R/W Utilisation"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
                "legend": {
                  "position": "bottom"
                },
                "yAxis": {
                  "format": {
                    "unit": "bytes/sec"
                  }
                }
              }
            },
            "queries": [
              {
                "kind": "TimeSeriesQuery",
                "spec": {
                  "plugin": {
                    "kind": "PrometheusTimeSeriesQuery",
                    "spec": {
                      "query": "vm:ceilometer_disk_device_read_bytes:rate1m{project_name =~ \"$project\", vm_name =~ \"$VM\"}",
                      "seriesNameFormat": "{{vm_name}} read ({{device}})"
                    }
                  }
                }
              },
              {
                "kind": "TimeSeriesQuery",
                "spec": {
                  "plugin": {
                    "kind": "PrometheusTimeSeriesQuery",
                    "spec": {
                      "query": "vm:ceilometer_disk_device_write_bytes:rate1m{project_name =~ \"$project\", vm_name =~ \"$VM\"}",
                      "seriesNameFormat": "{{vm_name}} write ({{device}})"
                    }
                  }
                }
              }
            ]
          }
        },
        "5": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Network Utilisation"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
                "legend": {
                  "position": "bottom"
                },
                "yAxis": {
                  "format": {
                    "unit": "bytes/sec"
                  }
                }
              }
            },
            "queries": [
              {
                "kind": "TimeSeriesQuery",
                "spec": {
                  "plugin": {
                    "kind": "PrometheusTimeSeriesQuery",
                    "spec": {
                      "query": "vm:ceilometer_network_incoming_bytes:rate1m{project_name =~ \"$project\", vm_name =~ \"$VM\"}",
                      "seriesNameFormat": "{{vm_name}} in ({{device}})"
                    }
                  }
                }
              },
              {
                "kind": "TimeSeriesQuery",
                "spec": {
                  "plugin": {
                    "kind": "PrometheusTimeSeriesQuery",
                    "spec": {
                      "query": "vm:ceilometer_network_outgoing_bytes:rate1m{project_name =~ \"$project\", vm_name =~ \"$VM\"}",
                      "seriesNameFormat": "{{vm_name}} out ({{device}})"
                    }
                  }
                }
              },
              {
                "kind": "TimeSeriesQuery",
                "spec": {
                  "plugin": {
                    "kind": "PrometheusTimeSeriesQuery",
                    "spec": {
                      "query": "label_replace(sum by(resource_name) (label_replace(rate(net_vf_receive_bytes_total[5m]) * on(device, vf, pci_address, fqdn) group_left(vm_id) net_vf_info{mac!=\"00:00:00:00:00:00\"}, \"resource\", \"$1\", \"vm_id\", \"(.*)\") * on(resource) group_left(resource_name) (ceilometer_cpu * 0 + 1)), \"vm_name\", \"$1\", \"resource_name\", \"(.*):.*\")",
                      "seriesNameFormat": "{{vm_name}} in (SR-IOV)"
                    }
                  }
                }
              },
              {
                "kind": "TimeSeriesQuery",
                "spec": {
                  "plugin": {
                    "kind": "PrometheusTimeSeriesQuery",
                    "spec": {
                      "query": "label_replace(sum by(resource_name) (label_replace(rate(net_vf_transmit_bytes_total[5m]) * on(device, vf, pci_address, fqdn) group_left(vm_id) net_vf_info{mac!=\"00:00:00:00:00:00\"}, \"resource\", \"$1\", \"vm_id\", \"(.*)\") * on(resource) group_left(resource_name) (ceilometer_cpu * 0 + 1)), \"vm_name\", \"$1\", \"resource_name\", \"(.*):.*\")",
                      "seriesNameFormat": "{{vm_name}} out (SR-IOV)"
                    }
                  }
                }
              }
            ]
          }
        },
        "6": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Network Saturation (Drop Rate)"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
                "legend": {
                  "position": "bottom"
                },
                "yAxis": {
                  "format": {
                    "unit": "packets/sec"
                  }
                }
              }
            },
            "queries": [
              {
                "kind": "TimeSeriesQuery",
                "spec": {
                  "plugin": {
                    "kind": "PrometheusTimeSeriesQuery",
                    "spec": {
                      "query": "vm:ceilometer_network_incoming_packets_drop:rate1m{project_name =~ \"$project\", vm_name =~ \"$VM\"}",
                      "seriesNameFormat": "{{vm_name}} in drop ({{device}})"
                    }
                  }
                }
              },
              {
                "kind": "TimeSeriesQuery",
                "spec": {
                  "plugin": {
                    "kind": "PrometheusTimeSeriesQuery",
                    "spec": {
                      "query": "vm:ceilometer_network_outgoing_packets_drop:rate1m{project_name =~ \"$project\", vm_name =~ \"$VM\"}",
                      "seriesNameFormat": "{{vm_name}} out drop ({{device}})"
                    }
                  }
                }
              },
              {
                "kind": "TimeSeriesQuery",
                "spec": {
                  "plugin": {
                    "kind": "PrometheusTimeSeriesQuery",
                    "spec": {
                      "query": "label_replace(sum by(resource_name) (label_replace(rate(net_vf_receive_dropped_total[5m]) * on(device, vf, pci_address, fqdn) group_left(vm_id) net_vf_info{mac!=\"00:00:00:00:00:00\"}, \"resource\", \"$1\", \"vm_id\", \"(.*)\") * on(resource) group_left(resource_name) (ceilometer_cpu * 0 + 1)), \"vm_name\", \"$1\", \"resource_name\", \"(.*):.*\")",
                      "seriesNameFormat": "{{vm_name}} in (SR-IOV)"
                    }
                  }
                }
              },
              {
                "kind": "TimeSeriesQuery",
                "spec": {
                  "plugin": {
                    "kind": "PrometheusTimeSeriesQuery",
                    "spec": {
                      "query": "label_replace(sum by(resource_name) (label_replace(rate(net_vf_transmit_dropped_total[5m]) * on(device, vf, pci_address, fqdn) group_left(vm_id) net_vf_info{mac!=\"00:00:00:00:00:00\"}, \"resource\", \"$1\", \"vm_id\", \"(.*)\") * on(resource) group_left(resource_name) (ceilometer_cpu * 0 + 1)), \"vm_name\", \"$1\", \"resource_name\", \"(.*):.*\")",
                      "seriesNameFormat": "{{vm_name}} out (SR-IOV)"
                    }
                  }
                }
              }
            ]
          }
        },
        "7": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Network Error Rate"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
                "legend": {
                  "position": "bottom"
                },
                "yAxis": {
                  "format": {
                    "unit": "packets/sec"
                  }
                }
              }
            },
            "queries": [
              {
                "kind": "TimeSeriesQuery",
                "spec": {
                  "plugin": {
                    "kind": "PrometheusTimeSeriesQuery",
                    "spec": {
                      "query": "vm:ceilometer_network_incoming_packets_error:rate1m{project_name =~ \"$project\", vm_name =~ \"$VM\"}",
                      "seriesNameFormat": "{{vm_name}} in error ({{device}})"
                    }
                  }
                }
              },
              {
                "kind": "TimeSeriesQuery",
                "spec": {
                  "plugin": {
                    "kind": "PrometheusTimeSeriesQuery",
                    "spec": {
                      "query": "vm:ceilometer_network_outgoing_packets_error:rate1m{project_name =~ \"$project\", vm_name =~ \"$VM\"}",
                      "seriesNameFormat": "{{vm_name}} out error ({{device}})"
                    }
                  }
                }
              }
            ]
          }
        }
      },
      "layouts": [
        {
          "kind": "Grid",
          "spec": {
            "display": {
              "title": "Panel Group 1"
            },
            "items": [
              {
                "x": 0,
                "y": 0,
                "width": 24,
                "height": 8,
                "content": {
                  "$ref": "#/spec/panels/0"
                }
              },
              {
                "x": 0,
                "y": 8,
                "width": 24,
                "height": 8,
                "content": {
                  "$ref": "#/spec/panels/1"
                }
              },
              {
                "x": 0,
                "y": 16,
                "width": 24,
                "height": 8,
                "content": {
                  "$ref": "#/spec/panels/2"
                }
              },
              {
                "x": 0,
                "y": 24,
                "width": 24,
                "height": 8,
                "content": {
                  "$ref": "#/spec/panels/3"
                }
              },
              {
                "x": 0,
                "y": 32,
                "width": 24,
                "height": 8,
                "content": {
                  "$ref": "#/spec/panels/4"
                }
              },
              {
                "x": 0,
                "y": 40,
                "width": 24,
                "height": 8,
                "content": {
                  "$ref": "#/spec/panels/5"
                }
              },
              {
                "x": 0,
                "y": 48,
                "width": 24,
                "height": 8,
                "content": {
                  "$ref": "#/spec/panels/6"
                }
              },
              {
                "x": 0,
                "y": 56,
                "width": 24,
                "height": 8,
                "content": {
                  "$ref": "#/spec/panels/7"
                }
              }
            ]
          }
        }
      ],
      "duration": "1h"
    }
  }
}`

// OpenstackVmPerses builds the "OpenStack / Ceilometer / VMs" PersesDashboard
// from its JSON definition, in the given namespace.
func OpenstackVmPerses(_, namespace string) (*persesv1alpha1.PersesDashboard, error) {
	return dashboardFromJSON(OpenstackVmName, openstackVmJSON, namespace)
}
