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

const openstackRabbitmqJSON = `{
  "apiVersion": "perses.dev/v1alpha2",
  "kind": "PersesDashboard",
  "metadata": {
    "name": "openstack-rabbitmq"
  },
  "spec": {
    "config": {
      "display": {
        "name": "OpenStack / RabbitMQ"
      },
      "variables": [
        {
          "kind": "ListVariable",
          "spec": {
            "display": {
              "name": "RabbitMQ Cluster",
              "hidden": false
            },
            "allowAllValue": false,
            "allowMultiple": false,
            "plugin": {
              "kind": "PrometheusLabelValuesVariable",
              "spec": {
                "labelName": "rabbitmq_cluster",
                "matchers": [
                  "rabbitmq_identity_info"
                ]
              }
            },
            "name": "cluster"
          }
        }
      ],
      "panels": {
        "0_0": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Queues",
              "description": "Number of message queues"
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
                      "query": "sum(rabbitmq_queues * on(instance) group_left(rabbitmq_cluster) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"})"
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
              "name": "Consumers",
              "description": "Number of message consumers"
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
                      "query": "sum(rabbitmq_consumers * on(instance) group_left(rabbitmq_cluster) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"})"
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
              "name": "Connections",
              "description": "Number of RabbitMQ connections"
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
                      "query": "sum(rabbitmq_connections * on(instance) group_left(rabbitmq_cluster) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"})"
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
              "name": "Channels",
              "description": "Number of RabbitMQ channels"
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
                      "query": "sum(rabbitmq_channels * on(instance) group_left(rabbitmq_cluster) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"})"
                    }
                  }
                }
              }
            ]
          }
        },
        "0_4": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Incoming messages / s",
              "description": "Number of incoming messages per second"
            },
            "plugin": {
              "kind": "StatChart",
              "spec": {
                "calculation": "last",
                "format": {
                  "unit": "decimal",
                  "decimalPlaces": 1
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
                      "query": "sum(rate(rabbitmq_global_messages_received_total[60s]) * on(instance) group_left(rabbitmq_cluster) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"})"
                    }
                  }
                }
              }
            ]
          }
        },
        "0_5": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Outgoing messages / s",
              "description": "Number of outgoing messages per second"
            },
            "plugin": {
              "kind": "StatChart",
              "spec": {
                "calculation": "last",
                "format": {
                  "unit": "decimal",
                  "decimalPlaces": 1
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
                      "query": "sum(rate(rabbitmq_global_messages_delivered_consume_auto_ack_total[60s]) * on(instance) group_left(rabbitmq_cluster) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) +sum(rate(rabbitmq_global_messages_delivered_consume_manual_ack_total[60s]) * on(instance) group_left(rabbitmq_cluster) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) +sum(rate(rabbitmq_global_messages_delivered_get_auto_ack_total[60s]) * on(instance) group_left(rabbitmq_cluster) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) +sum(rate(rabbitmq_global_messages_delivered_get_manual_ack_total[60s]) * on(instance) group_left(rabbitmq_cluster) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"})"
                    }
                  }
                }
              }
            ]
          }
        },
        "0_6": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Ready messages",
              "description": "Number of ready messages"
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
                      "query": "sum(rabbitmq_queue_messages_ready * on(instance) group_left(rabbitmq_cluster) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"})"
                    }
                  }
                }
              }
            ]
          }
        },
        "0_7": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Unacknowledged messages",
              "description": "Number of unacknowledged messages"
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
                      "query": "sum(rabbitmq_queue_messages_unacked * on(instance) group_left(rabbitmq_cluster) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"})"
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
              "name": "Memory available before publishers blocked"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "(rabbitmq_resident_memory_limit_bytes * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) -(rabbitmq_process_resident_memory_bytes * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"})"
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
              "name": "Disk space available before publishers blocked"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "(rabbitmq_disk_space_available_bytes - rabbitmq_disk_space_available_limit_bytes) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}"
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
              "name": "File descriptors available"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "(rabbitmq_process_max_fds * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) -(rabbitmq_process_open_fds * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"})",
                      "seriesNameFormat": "{{rabbitmq_node}}"
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
              "name": "TCP sockets available"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "(rabbitmq_process_max_tcp_sockets * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) -(rabbitmq_process_open_tcp_sockets * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"})",
                      "seriesNameFormat": "{{rabbitmq_node}}"
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
              "name": "Messages ready to be delivered"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rabbitmq_queue_messages_ready * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
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
              "name": "Unacknowledged messages"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rabbitmq_queue_messages_unacked * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
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
              "name": "Messages published / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_global_messages_received_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
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
              "name": "Messages confirmed / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_global_messages_confirmed_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
                    }
                  }
                }
              }
            ]
          }
        },
        "3_2": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Messages routed to queues / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_global_messages_routed_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
                    }
                  }
                }
              }
            ]
          }
        },
        "3_3": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Unconfirmed messages / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_global_messages_received_confirm_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"} - rate(rabbitmq_global_messages_confirmed_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
                    }
                  }
                }
              }
            ]
          }
        },
        "3_4": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Unroutable messages dropped / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_global_messages_unroutable_dropped_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
                    }
                  }
                }
              }
            ]
          }
        },
        "3_5": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Unroutable messages returned to publishers / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_global_messages_unroutable_returned_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
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
              "name": "Messages delivered / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(  (rate(rabbitmq_global_messages_delivered_consume_auto_ack_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) +  (rate(rabbitmq_global_messages_delivered_consume_manual_ack_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"})) by(rabbitmq_node)"
                    }
                  }
                }
              }
            ]
          }
        },
        "4_1": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Messages redelivered / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_global_messages_redelivered_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
                    }
                  }
                }
              }
            ]
          }
        },
        "4_2": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Messages delivered with manual acks / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_global_messages_delivered_consume_manual_ack_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
                    }
                  }
                }
              }
            ]
          }
        },
        "4_3": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Messages delivered with auto acks / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_global_messages_delivered_consume_auto_ack_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
                    }
                  }
                }
              }
            ]
          }
        },
        "4_4": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Messages acknowledged / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_global_messages_acknowledged_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
                    }
                  }
                }
              }
            ]
          }
        },
        "4_5": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Polling operations with auto ack / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_global_messages_delivered_get_auto_ack_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
                    }
                  }
                }
              }
            ]
          }
        },
        "4_6": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Polling operations that yield no result / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_global_messages_get_empty_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
                    }
                  }
                }
              }
            ]
          }
        },
        "4_7": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Polling operations with manual ack / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_global_messages_delivered_get_manual_ack_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
                    }
                  }
                }
              }
            ]
          }
        },
        "5_0": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Total queues"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "rabbitmq_queues * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}",
                      "seriesNameFormat": "{{rabbitmq_node}}"
                    }
                  }
                }
              }
            ]
          }
        },
        "5_1": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Queues declared / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_queues_declared_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
                    }
                  }
                }
              }
            ]
          }
        },
        "5_2": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Queues created / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_queues_created_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
                    }
                  }
                }
              }
            ]
          }
        },
        "5_3": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Queues deleted / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_queues_deleted_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
                    }
                  }
                }
              }
            ]
          }
        },
        "6_0": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Total channels"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "rabbitmq_channels * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}"
                    }
                  }
                }
              }
            ]
          }
        },
        "6_1": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Channels opened / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_channels_opened_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
                    }
                  }
                }
              }
            ]
          }
        },
        "6_2": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Channels closed / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_channels_closed_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
                    }
                  }
                }
              }
            ]
          }
        },
        "7_0": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Total connections"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "rabbitmq_connections * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}"
                    }
                  }
                }
              }
            ]
          }
        },
        "7_1": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Connections opened / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_connections_opened_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)"
                    }
                  }
                }
              }
            ]
          }
        },
        "7_2": {
          "kind": "Panel",
          "spec": {
            "display": {
              "name": "Connections closed / s"
            },
            "plugin": {
              "kind": "TimeSeriesChart",
              "spec": {
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
                      "query": "sum(rate(rabbitmq_connections_closed_total[60s]) * on(instance) group_left(rabbitmq_cluster, rabbitmq_node) rabbitmq_identity_info{rabbitmq_cluster=\"$cluster\"}) by(rabbitmq_node)",
                      "seriesNameFormat": "{{rabbitmq_node}}"
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
                "width": 6,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/0_0"
                }
              },
              {
                "x": 6,
                "y": 1,
                "width": 6,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/0_1"
                }
              },
              {
                "x": 12,
                "y": 1,
                "width": 6,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/0_2"
                }
              },
              {
                "x": 18,
                "y": 1,
                "width": 6,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/0_3"
                }
              },
              {
                "x": 0,
                "y": 8,
                "width": 6,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/0_4"
                }
              },
              {
                "x": 6,
                "y": 8,
                "width": 6,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/0_5"
                }
              },
              {
                "x": 12,
                "y": 8,
                "width": 6,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/0_6"
                }
              },
              {
                "x": 18,
                "y": 8,
                "width": 6,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/0_7"
                }
              }
            ]
          }
        },
        {
          "kind": "Grid",
          "spec": {
            "display": {
              "title": "Nodes",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 16,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/1_0"
                }
              },
              {
                "x": 12,
                "y": 16,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/1_1"
                }
              },
              {
                "x": 0,
                "y": 23,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/1_2"
                }
              },
              {
                "x": 12,
                "y": 23,
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
              "title": "Queued messages",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 31,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/2_0"
                }
              },
              {
                "x": 12,
                "y": 31,
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
              "title": "Incoming messages",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 39,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/3_0"
                }
              },
              {
                "x": 12,
                "y": 39,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/3_1"
                }
              },
              {
                "x": 0,
                "y": 46,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/3_2"
                }
              },
              {
                "x": 12,
                "y": 46,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/3_3"
                }
              },
              {
                "x": 0,
                "y": 53,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/3_4"
                }
              },
              {
                "x": 12,
                "y": 53,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/3_5"
                }
              }
            ]
          }
        },
        {
          "kind": "Grid",
          "spec": {
            "display": {
              "title": "Outgoing messages",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 61,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/4_0"
                }
              },
              {
                "x": 12,
                "y": 61,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/4_1"
                }
              },
              {
                "x": 0,
                "y": 68,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/4_2"
                }
              },
              {
                "x": 12,
                "y": 68,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/4_3"
                }
              },
              {
                "x": 0,
                "y": 75,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/4_4"
                }
              },
              {
                "x": 12,
                "y": 75,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/4_5"
                }
              },
              {
                "x": 0,
                "y": 82,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/4_6"
                }
              },
              {
                "x": 12,
                "y": 82,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/4_7"
                }
              }
            ]
          }
        },
        {
          "kind": "Grid",
          "spec": {
            "display": {
              "title": "Queues",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 90,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/5_0"
                }
              },
              {
                "x": 12,
                "y": 90,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/5_1"
                }
              },
              {
                "x": 0,
                "y": 97,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/5_2"
                }
              },
              {
                "x": 12,
                "y": 97,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/5_3"
                }
              }
            ]
          }
        },
        {
          "kind": "Grid",
          "spec": {
            "display": {
              "title": "Channels",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 105,
                "width": 24,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/6_0"
                }
              },
              {
                "x": 0,
                "y": 112,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/6_1"
                }
              },
              {
                "x": 12,
                "y": 112,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/6_2"
                }
              }
            ]
          }
        },
        {
          "kind": "Grid",
          "spec": {
            "display": {
              "title": "Connections",
              "collapse": {
                "open": true
              }
            },
            "items": [
              {
                "x": 0,
                "y": 120,
                "width": 24,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/7_0"
                }
              },
              {
                "x": 0,
                "y": 127,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/7_1"
                }
              },
              {
                "x": 12,
                "y": 127,
                "width": 12,
                "height": 7,
                "content": {
                  "$ref": "#/spec/panels/7_2"
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

// OpenstackRabbitmqPerses builds the "OpenStack / RabbitMQ" PersesDashboard
// from its JSON definition, in the given namespace.
func OpenstackRabbitmqPerses(_, namespace string) (*persesv1alpha1.PersesDashboard, error) {
	return dashboardFromJSON(OpenstackRabbitmqName, openstackRabbitmqJSON, namespace)
}
