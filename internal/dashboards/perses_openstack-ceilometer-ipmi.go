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

const openstackCeilometerIpmiJSON = `{
  "apiVersion": "perses.dev/v1alpha2",
  "kind": "PersesDashboard",
  "metadata": {
    "name": "openstack-ceilometer-ipmi"
  },
  "spec": {
    "config": {
      "display": {
        "name": "OpenStack / Compute Power Monitoring / Ceilometer IPMI"
      },
      "variables": [
        {
          "kind": "ListVariable",
          "spec": {
            "display": {
              "hidden": false
            },
            "allowAllValue": false,
            "allowMultiple": false,
            "sort": "none",
            "plugin": {
              "kind": "PrometheusLabelValuesVariable",
              "spec": {
                "labelName": "fqdn",
                "matchers": [
                  "ceilometer_hardware_ipmi_voltage"
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
              "name": "Voltage (Volts)"
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
                      "query": "label_replace(ceilometer_hardware_ipmi_voltage{unit=\"V\", resource=~\"$fqdn.*\"}, \"resource_name\", \"$1\", \"resource\", \".*-(.*)\")",
                      "seriesNameFormat": "{{ resource_name }}"
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
              "name": "Current (Amps)"
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
                      "query": "label_replace(ceilometer_hardware_ipmi_current{unit=\"Amps\", resource=~\"$fqdn.*\"}, \"resource_name\", \"$1\", \"resource\", \".*-(current.*)\")",
                      "seriesNameFormat": "{{ resource_name }}"
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
              "name": "Power (Watts)"
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
                      "query": "label_replace(ceilometer_hardware_ipmi_current{unit=\"W\", resource=~\"$fqdn.*\"}, \"resource_name\", \"$1\", \"resource\", \".*-(pwr.*)\")",
                      "seriesNameFormat": "{{ resource_name }}"
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
              "name": "Fan (RPM)"
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
                      "query": "label_replace(ceilometer_hardware_ipmi_fan{unit=\"RPM\", resource=~\"$fqdn.*\"}, \"resource_name\", \"$1\", \"resource\", \".*-(fan.*)\")",
                      "seriesNameFormat": "{{ resource_name }}"
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
              "name": "Temperature (Celsius)"
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
                      "query": "label_replace(ceilometer_hardware_ipmi_temperature{unit=\"C\", resource=~\"$fqdn.*\"}, \"resource_name\", \"$2\", \"resource\", \"^(.*)\\\\-(exhaust_temp_.*|inlet_temp_.*|temp_.*)$\")",
                      "seriesNameFormat": "{{ resource_name }}"
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
              "title": "Power Supply Unit",
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
              },
              {
                "x": 0,
                "y": 8,
                "width": 24,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/0_2"
                }
              }
            ]
          }
        },
        {
          "kind": "Grid",
          "spec": {
            "display": {
              "title": "Cooling Overview",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 16,
                "width": 24,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/1_0"
                }
              },
              {
                "x": 0,
                "y": 23,
                "width": 24,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/1_1"
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

// OpenstackCeilometerIpmiPerses builds the "OpenStack / Compute Power
// Monitoring / Ceilometer IPMI" PersesDashboard from its JSON definition, in
// the given namespace.
func OpenstackCeilometerIpmiPerses(_, namespace string) (*persesv1alpha1.PersesDashboard, error) {
	return dashboardFromJSON(OpenstackCeilometerIpmiName, openstackCeilometerIpmiJSON, namespace)
}
