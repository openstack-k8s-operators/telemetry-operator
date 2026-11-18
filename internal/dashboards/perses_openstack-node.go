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

const openstackNodeJSON = `{
  "apiVersion": "perses.dev/v1alpha2",
  "kind": "PersesDashboard",
  "metadata": {
    "name": "openstack-node"
  },
  "spec": {
    "config": {
      "display": {
        "name": "Openstack / Node Exporter / USE Method / Node"
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
            "sort": "alphabetical-asc",
            "plugin": {
              "kind": "PrometheusLabelValuesVariable",
              "spec": {
                "labelName": "fqdn",
                "matchers": [
                  "node_load1"
                ]
              }
            },
            "name": "fqdn"
          }
        }
      ],
      "panels": {
        "0_0": {
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
                      "query": "instance:node_cpu_utilisation:rate1m{fqdn=~\"$fqdn\"}",
                      "seriesNameFormat": "{{fqdn}}"
                    }
                  }
                }
              }
            ]
          }
        },
        "0_1": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "CPU Saturation (Load1 per CPU)"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
                "legend": {
                  "position": "bottom"
                },
                "yAxis": {
                  "format": {
                    "unit": "decimal"
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
                      "query": "instance:node_load1_per_cpu:ratio{fqdn=~\"$fqdn\"}",
                      "seriesNameFormat": "{{fqdn}}"
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
                      "query": "instance:node_memory_utilisation:ratio{fqdn=~\"$fqdn\"}",
                      "seriesNameFormat": "{{fqdn}}"
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
              "name": "Memory Saturation (Major Page Faults)"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
                "legend": {
                  "position": "bottom"
                },
                "yAxis": {
                  "format": {
                    "unit": "ops/sec"
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
                      "query": "instance:node_vmstat_pgmajfault:rate1m{fqdn=~\"$fqdn\"}",
                      "seriesNameFormat": "{{fqdn}}"
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
              "name": "Network Utilisation (Bytes Receive/Transmit)"
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
                      "query": "instance:node_network_receive_bytes_excluding_lo:rate1m{fqdn=~\"$fqdn\"}",
                      "seriesNameFormat": "{{fqdn}} Receive"
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
                      "query": "instance:node_network_transmit_bytes_excluding_lo:rate1m{fqdn=~\"$fqdn\"}",
                      "seriesNameFormat": "{{fqdn}} Transmit"
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
              "name": "Network Saturation (Drops Receive/Transmit)"
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
                      "query": "instance:node_network_receive_drop_excluding_lo:rate1m{fqdn=~\"$fqdn\"}",
                      "seriesNameFormat": "{{fqdn}} Receive"
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
                      "query": "instance:node_network_transmit_drop_excluding_lo:rate1m{fqdn=~\"$fqdn\"}",
                      "seriesNameFormat": "{{fqdn}} Transmit"
                    }
                  }
                }
              }
            ]
          }
        },
        "3_0": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Disk IO Utilisation"
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
                      "query": "instance_device:node_disk_io_time_seconds:rate1m{fqdn=~\"$fqdn\"}",
                      "seriesNameFormat": "{{fqdn}} {{device}}"
                    }
                  }
                }
              }
            ]
          }
        },
        "3_1": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Disk IO Saturation (Queue Time)"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
                "legend": {
                  "position": "bottom"
                },
                "yAxis": {
                  "format": {
                    "unit": "seconds"
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
                      "query": "instance_device:node_disk_io_time_weighted_seconds:rate1m{fqdn=~\"$fqdn\"}",
                      "seriesNameFormat": "{{fqdn}} {{device}}"
                    }
                  }
                }
              }
            ]
          }
        },
        "4_0": {
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
                      "query": "1 -\n  (\n   max without (mountpoint, fstype) (node_filesystem_avail_bytes{fstype!=\"\", fqdn=~\"$fqdn\"})\n   /\n   max without (mountpoint, fstype) (node_filesystem_size_bytes{fstype!=\"\", fqdn=~\"$fqdn\"})\n  )\n",
                      "seriesNameFormat": "{{fqdn}} {{device}}"
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
              "title": "CPU",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 1,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/0_0"
                }
              },
              {
                "x": 12,
                "y": 1,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/0_1"
                }
              }
            ]
          }
        },
        {
          "kind": "Grid",
          "spec": {
            "display": {
              "title": "Memory",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 9,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/1_0"
                }
              },
              {
                "x": 12,
                "y": 9,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/1_1"
                }
              }
            ]
          }
        },
        {
          "kind": "Grid",
          "spec": {
            "display": {
              "title": "Network",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 17,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/2_0"
                }
              },
              {
                "x": 12,
                "y": 17,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/2_1"
                }
              }
            ]
          }
        },
        {
          "kind": "Grid",
          "spec": {
            "display": {
              "title": "Disk IO",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 25,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/3_0"
                }
              },
              {
                "x": 12,
                "y": 25,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/3_1"
                }
              }
            ]
          }
        },
        {
          "kind": "Grid",
          "spec": {
            "display": {
              "title": "Disk Space",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 33,
                "width": 24,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/4_0"
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

// OpenstackNodePerses builds the "Openstack / Node Exporter / USE Method /
// Node" PersesDashboard from its JSON definition, in the given namespace.
func OpenstackNodePerses(_, namespace string) (*persesv1alpha1.PersesDashboard, error) {
	return dashboardFromJSON(OpenstackNodeName, openstackNodeJSON, namespace)
}
