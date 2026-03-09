# metricstorage kuttl test

The "metricstorage" kuttl test covers the deployment and configuration of the MetricStorage CR, which manages Prometheus-based metrics collection and storage for OpenStack services.

## Topology

The target topology for this test is:

1. one MetricStorage object managing:
   - MonitoringStack (Prometheus deployment)
   - ScrapeConfigs for metrics collection
   - PrometheusRules for alerting
   - Dashboard configurations (optional)
2. Prometheus StatefulSet with persistent storage
3. Integration with Red Hat Observability Stack (RHOBS)

The metricstorage test steps cover:
- Basic deployment with default monitoring stack
- Dashboard configuration management
- Custom monitoring stack integration
- Network attachments for multi-network setups

### Prerequisites

As a prerequisite for this test, we assume:

1. Red Hat Observability Stack (RHOBS) operators installed:
   - monitoring.rhobs/v1 API group available
   - MonitoringStack CRD
   - ScrapeConfig CRD
   - PrometheusRule CRD
2. Storage provisioner for persistent volumes
3. OpenShift infrastructure for dashboard integration (optional)
4. a running `telemetry-operator` deployed in the cluster

These resources are deployed via the `deps/` directory using kustomize.

### Run the metricstorage tests

Once the kuttl requirements are satisfied, the actual tests can be executed from the telemetry-operator root directory via:

```bash
cd telemetry-operator/
kubectl-kuttl test --test metricstorage --namespace telemetry-kuttl-metricstorage
```

Same tests can be executed from a different location via:

```bash
kubectl-kuttl test --config <KUTTL_CONFIG> <KUTTL_TESTS_DIR> --test metricstorage --namespace telemetry-kuttl-metricstorage
```

> **NOTE**: <KUTTL_CONFIG> is the path where kuttl-test.yaml for the repo is located and <KUTTL_TESTS_DIR> is where the tests are located.
