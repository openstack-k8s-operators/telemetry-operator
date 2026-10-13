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

const openstackLightspeedJSON = `{
  "apiVersion": "perses.dev/v1alpha2",
  "kind": "PersesDashboard",
  "metadata": {
    "name": "openstack-lightspeed"
  },
  "spec": {
    "config": {
      "display": {
        "name": "OpenStack / Lightspeed"
      },
      "panels": {
        "0_0": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Total API Calls",
              "description": "Total REST API calls"
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
                      "query": "round(sum(increase(ls_rest_api_calls_total[$__range])))"
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
              "name": "Total LLM Failures",
              "description": "Total LLM failures"
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
                      "query": "round(sum(increase(ls_llm_calls_failures_total[$__range])))"
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
              "name": "Model Configuration",
              "description": "Configured LLM provider and model combinations"
            },
            "plugin": {
              "kind": "Table",
              "spec": {
                "columnSettings": []
              }
            },
            "queries": [
              {
                "kind": "TimeSeriesQuery",
                "spec": {
                  "plugin": {
                    "kind": "PrometheusTimeSeriesQuery",
                    "spec": {
                      "query": "ls_provider_model_configuration"
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
              "name": "Service Status",
              "description": "Service health status"
            },
            "plugin": {
              "kind": "StatChart",
              "spec": {
                "calculation": "last",
                "format": {
                  "unit": "decimal"
                },
                "thresholds": {
                  "steps": [
                    {
                      "value": 0,
                      "color": "green",
                      "name": "OK"
                    },
                    {
                      "value": 1,
                      "color": "red",
                      "name": "Degraded"
                    }
                  ]
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
                      "query": "ls_started_in_degraded_mode"
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
              "name": "Response Time Percentiles",
              "description": "Response time percentiles"
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
                      "query": "histogram_quantile(0.50, sum(rate(ls_response_duration_seconds_bucket[5m])) by (le))",
                      "seriesNameFormat": "p50"
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
                      "query": "histogram_quantile(0.90, sum(rate(ls_response_duration_seconds_bucket[5m])) by (le))",
                      "seriesNameFormat": "p90"
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
                      "query": "histogram_quantile(0.95, sum(rate(ls_response_duration_seconds_bucket[5m])) by (le))",
                      "seriesNameFormat": "p95"
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
                      "query": "histogram_quantile(0.99, sum(rate(ls_response_duration_seconds_bucket[5m])) by (le))",
                      "seriesNameFormat": "p99"
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
              "name": "Average Response Time",
              "description": "Average response time"
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
                      "query": "rate(ls_response_duration_seconds_sum[5m]) / clamp_min(rate(ls_response_duration_seconds_count[5m]), 1)"
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
              "name": "LLM Calls per Minute",
              "description": "LLM calls per minute"
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
                      "query": "sum(rate(ls_llm_calls_total[5m])) * 60"
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
              "name": "Token per Second",
              "description": "Token usage rates"
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
                      "query": "sum(rate(ls_llm_token_sent_total[5m]))",
                      "seriesNameFormat": "sent"
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
                      "query": "sum(rate(ls_llm_token_received_total[5m]))",
                      "seriesNameFormat": "received"
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
              "name": "LLM Failures per Minute",
              "description": "LLM call failures per minute"
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
                      "query": "sum(rate(ls_llm_calls_failures_total[5m])) * 60",
                      "seriesNameFormat": "failures"
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
              "name": "Validation Errors per Minute",
              "description": "LLM validation errors per minute"
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
                      "query": "sum(rate(ls_llm_validation_errors_total[5m])) * 60"
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
              },
              {
                "x": 0,
                "y": 15,
                "width": 24,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/0_3"
                }
              }
            ]
          }
        },
        {
          "kind": "Grid",
          "spec": {
            "display": {
              "title": "Response Times",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 23,
                "width": 24,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/1_0"
                }
              },
              {
                "x": 0,
                "y": 30,
                "width": 24,
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
              "title": "LLM Metrics",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 38,
                "width": 24,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/2_0"
                }
              },
              {
                "x": 0,
                "y": 45,
                "width": 24,
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
              "title": "Errors",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 53,
                "width": 24,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/3_0"
                }
              },
              {
                "x": 0,
                "y": 60,
                "width": 24,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/3_1"
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

// OpenstackLightspeedPerses builds the "OpenStack / Lightspeed" PersesDashboard
// from its JSON definition, in the given namespace.
func OpenstackLightspeedPerses(_, namespace string) (*persesv1alpha1.PersesDashboard, error) {
	return dashboardFromJSON(OpenstackLightspeedName, openstackLightspeedJSON, namespace)
}
