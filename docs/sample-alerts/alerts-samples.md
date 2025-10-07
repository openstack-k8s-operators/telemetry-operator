# Alertmanager sample rules

This document outlines the custom Alertmanager alerting rules for monitoring an OpenStack deployment running on top of an OpenShift cluster. The sample alerts are divided into two main groups: services status alerts and nodes status alerts.

> **NOTE:** The samples provided in this document are intended as examples for guidance only. You should review and adapt them to fit the specific metrics, labels, and operational context of your environment. Thresholds for resource utilization, in particular, may need significant tuning based on your workload patterns and capacity planning.

## OpenStack Observability Services Status Alerts

This group of alerts monitors the availability of core OpenStack services. These alerts are critical as they indicate a direct impact on the functionality of the OpenStack control plane and its APIs.

## OpenStack Observability Nodes Status Alerts

This group of alerts monitors the fundamental compute and resources managed by the OpenStack deployment. These alerts help prevent service degradation by providing early warnings about resource exhaustion.
