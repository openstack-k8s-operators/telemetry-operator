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

const openstackOpenstackNetworkJSON = `{
  "apiVersion": "perses.dev/v1alpha2",
  "kind": "PersesDashboard",
  "metadata": {
    "name": "openstack-network"
  },
  "spec": {
    "config": {
      "display": {
        "name": "Openstack / Openstack Network Exporter"
      },
      "variables": [
        {
          "kind": "ListVariable",
          "spec": {
            "display": {
              "hidden": false
            },
            "allowAllValue": false,
            "allowMultiple": true,
            "sort": "alphabetical-asc",
            "plugin": {
              "kind": "PrometheusLabelValuesVariable",
              "spec": {
                "labelName": "fqdn",
                "matchers": [
                  "ovs_build_info"
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
              "name": "Flow Count of Integration Bridge",
              "description": "Integration Bridge flow count"
            },
            "plugin": {
              "kind": "StatChart",
              "spec": {
                "calculation": "last",
                "format": {
                  "unit": "decimal"
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
                      "query": "ovs_bridge_flow_count{bridge=\"br-int\",fqdn=~\"$fqdn\"}",
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
              "name": "Logical Flow Translations",
              "description": "Number of times ovn-controller has translated the Logical_Flow table in the OVN SB database into OpenFlow flow in the past time period."
            },
            "plugin": {
              "kind": "StatChart",
              "spec": {
                "calculation": "last",
                "format": {
                  "unit": "decimal"
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
                      "query": "increase(ovnc_lflow_run{fqdn=~\"$fqdn\"}[5m])",
                      "seriesNameFormat": "{{fqdn}}"
                    }
                  }
                }
              }
            ]
          }
        },
        "0_2": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Netlink"
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
                      "query": "rate(ovs_coverage_netlink_sent_total{fqdn=~\"$fqdn\"}[5m])",
                      "seriesNameFormat": "{{fqdn}} Sent"
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
                      "query": "rate(ovs_coverage_netlink_received_total{fqdn=~\"$fqdn\"}[5m])",
                      "seriesNameFormat": "{{fqdn}} Received"
                    }
                  }
                }
              }
            ]
          }
        },
        "0_3": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Recirculation Rate"
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
                      "query": "rate(ovs_pmd_datapath_passes{fqdn=~\"$fqdn\"}[5m]) / clamp_min(rate(ovs_pmd_rx_packets{fqdn=~\"$fqdn\"}[5m]), 1)",
                      "seriesNameFormat": "{{fqdn}}"
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
              "title": "OVN/OVS",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 1,
                "width": 12,
                "height": 5,
                "content": {
                  "$ref": "#/spec/panels/0_0"
                }
              },
              {
                "x": 12,
                "y": 1,
                "width": 12,
                "height": 5,
                "content": {
                  "$ref": "#/spec/panels/0_1"
                }
              },
              {
                "x": 0,
                "y": 7,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/0_2"
                }
              },
              {
                "x": 12,
                "y": 7,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/0_3"
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

// OpenstackOpenstackNetworkPerses builds the "Openstack / Openstack Network
// Exporter" PersesDashboard from its JSON definition, in the given namespace.
func OpenstackOpenstackNetworkPerses(_, namespace string) (*persesv1alpha1.PersesDashboard, error) {
	return dashboardFromJSON(OpenstackOpenstackNetworkName, openstackOpenstackNetworkJSON, namespace)
}
