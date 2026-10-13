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

const openstackNetworkTrafficJSON = `{
  "apiVersion": "perses.dev/v1alpha2",
  "kind": "PersesDashboard",
  "metadata": {
    "name": "openstack-network-traffic"
  },
  "spec": {
    "config": {
      "display": {
        "name": "OpenStack / VMs Network Traffic"
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
            "sort": "none",
            "plugin": {
              "kind": "PrometheusLabelValuesVariable",
              "spec": {
                "labelName": "fqdn",
                "matchers": [
                  "vm:ceilometer_cpu:ratio1m"
                ]
              }
            },
            "name": "compute_node"
          }
        },
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
                  "ceilometer_cpu{fqdn=~\"$compute_node\"}"
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
            "allowAllValue": true,
            "allowMultiple": true,
            "customAllValue": ".*",
            "sort": "none",
            "plugin": {
              "kind": "PrometheusLabelValuesVariable",
              "spec": {
                "labelName": "vm_name",
                "matchers": [
                  "vm:ceilometer_cpu:ratio1m{project_name =~ \"$project\", fqdn =~ \"$compute_node\"}"
                ]
              }
            },
            "name": "VM"
          }
        },
        {
          "kind": "ListVariable",
          "spec": {
            "display": {
              "hidden": false
            },
            "allowAllValue": true,
            "allowMultiple": true,
            "customAllValue": ".*",
            "sort": "none",
            "plugin": {
              "kind": "PrometheusLabelValuesVariable",
              "spec": {
                "labelName": "device",
                "matchers": [
                  "vm:ceilometer_network_incoming_bytes:rate1m{vm_name=~\"$VM\"}"
                ]
              }
            },
            "name": "In_adapter"
          }
        },
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
                "labelName": "device",
                "matchers": [
                  "vm:ceilometer_network_outgoing_bytes:rate1m{vm_name=~\"$VM\"}"
                ]
              }
            },
            "name": "Out_adapter"
          }
        }
      ],
      "panels": {
        "0_0": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Network Traffic (All Adapters)"
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
              }
            ]
          }
        },
        "1_0": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Incoming Bytes"
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
                      "query": "vm:ceilometer_network_incoming_bytes:rate1m{resource_name=~\"$VM:.*\", project_name=~\"$project\" }",
                      "seriesNameFormat": "{{vm_name}} in ({{device}})"
                    }
                  }
                }
              }
            ]
          }
        },
        "1_1": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Incoming Dropped"
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
                      "query": "vm:ceilometer_network_incoming_packets_drop:rate1m{project_name =~ \"$project\",vm_name =~ \"$VM\", device =~\"$In_adapter\" }",
                      "seriesNameFormat": "{{vm_name}} in ({{device}})"
                    }
                  }
                }
              }
            ]
          }
        },
        "1_2": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Incoming Packet Loss (%)"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
                "legend": {
                  "position": "bottom",
                  "size": "medium"
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
                      "query": "(vm:ceilometer_network_incoming_packets_drop:rate1m{project_name =~ \"$project\",vm_name =~ \"$VM\", device =~\"$In_adapter\"} + ignoring(counter) vm:ceilometer_network_incoming_packets_error:rate1m{project_name =~ \"$project\",vm_name =~ \"$VM\", device =~\"$In_adapter\"}) / clamp_min(vm:ceilometer_network_incoming_packets:rate1m{project_name =~ \"$project\",vm_name =~ \"$VM\", device =~\"$In_adapter\"} + ignoring(counter) vm:ceilometer_network_incoming_packets_drop:rate1m{project_name =~ \"$project\",vm_name =~ \"$VM\", device =~\"$In_adapter\"} + ignoring(counter) vm:ceilometer_network_incoming_packets_error:rate1m{project_name =~ \"$project\",vm_name =~ \"$VM\", device =~\"$In_adapter\"}, 1)",
                      "seriesNameFormat": "{{vm_name}} in ({{device}})"
                    }
                  }
                }
              }
            ]
          }
        },
        "1_3": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Incoming Errors"
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
                      "query": "vm:ceilometer_network_incoming_packets_error:rate1m{project_name =~ \"$project\",vm_name =~ \"$VM\", device =~\"$In_adapter\"}",
                      "seriesNameFormat": "{{vm_name}} in ({{device}})"
                    }
                  }
                }
              }
            ]
          }
        },
        "2_0": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Outgoing Bytes"
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
                      "query": "vm:ceilometer_network_outgoing_bytes:rate1m{resource_name=~\"$VM:.+\", project_name=~\"$project\"}",
                      "seriesNameFormat": "{{vm_name}} out ({{device}})"
                    }
                  }
                }
              }
            ]
          }
        },
        "2_1": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Outgoing Dropped"
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
                      "query": "vm:ceilometer_network_outgoing_packets_drop:rate1m{project_name =~ \"$project\",vm_name =~ \"$VM\", device =~\"$Out_adapter\" }",
                      "seriesNameFormat": "{{vm_name}} out ({{device}})"
                    }
                  }
                }
              }
            ]
          }
        },
        "2_2": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Outgoing Packet Loss (%)"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
                "legend": {
                  "position": "bottom",
                  "size": "medium"
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
                      "query": "(vm:ceilometer_network_outgoing_packets_drop:rate1m{project_name =~ \"$project\",vm_name =~ \"$VM\", device =~\"$Out_adapter\"} + ignoring(counter) vm:ceilometer_network_outgoing_packets_error:rate1m{project_name =~ \"$project\",vm_name =~ \"$VM\", device =~\"$Out_adapter\"}) / clamp_min(vm:ceilometer_network_outgoing_packets:rate1m{project_name =~ \"$project\",vm_name =~ \"$VM\", device =~\"$Out_adapter\"} + ignoring(counter) vm:ceilometer_network_outgoing_packets_drop:rate1m{project_name =~ \"$project\",vm_name =~ \"$VM\", device =~\"$Out_adapter\"} + ignoring(counter) vm:ceilometer_network_outgoing_packets_error:rate1m{project_name =~ \"$project\",vm_name =~ \"$VM\", device =~\"$Out_adapter\"}, 1)",
                      "seriesNameFormat": "{{vm_name}} out ({{device}})"
                    }
                  }
                }
              }
            ]
          }
        },
        "2_3": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Outgoing Errors"
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
                      "query": "vm:ceilometer_network_outgoing_packets_error:rate1m{project_name =~ \"$project\",vm_name =~ \"$VM\", device =~\"$Out_adapter\"}",
                      "seriesNameFormat": "{{vm_name}} out ({{device}})"
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
              "title": "Overview",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 0,
                "width": 24,
                "height": 8,
                "content": {
                  "$ref": "#/spec/panels/0_0"
                }
              }
            ]
          }
        },
        {
          "kind": "Grid",
          "spec": {
            "display": {
              "title": "Inbound Traffic",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 0,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/1_0"
                }
              },
              {
                "x": 12,
                "y": 0,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/1_1"
                }
              },
              {
                "x": 12,
                "y": 7,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/1_2"
                }
              },
              {
                "x": 0,
                "y": 7,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/1_3"
                }
              }
            ]
          }
        },
        {
          "kind": "Grid",
          "spec": {
            "display": {
              "title": "Outbound Traffic",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 0,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/2_0"
                }
              },
              {
                "x": 12,
                "y": 0,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/2_1"
                }
              },
              {
                "x": 12,
                "y": 7,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/2_2"
                }
              },
              {
                "x": 0,
                "y": 7,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/2_3"
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

// OpenstackNetworkTrafficPerses builds the "OpenStack / VMs Network Traffic"
// PersesDashboard from its JSON definition, in the given namespace.
func OpenstackNetworkTrafficPerses(_, namespace string) (*persesv1alpha1.PersesDashboard, error) {
	return dashboardFromJSON(OpenstackNetworkTrafficName, openstackNetworkTrafficJSON, namespace)
}
