# default kuttl test

The "default" kuttl test covers the integrated deployment of all major telemetry components together, validating that they can coexist and interoperate in a complete telemetry stack.

## Topology

The target topology for this test is:

1. Complete telemetry stack including:
   - **Ceilometer** with kube-state-metrics
   - **MetricStorage** with Prometheus monitoring stack
   - **Autoscaling** with Aodh service
   - **CloudKitty** with LokiStack for billing/rating
2. All components deployed and ready simultaneously
3. Full integration with:
   - Keystone for authentication
   - MariaDB for database backend
   - RabbitMQ for messaging
   - Memcached for caching
   - S3 storage (MinIO) for CloudKitty/Loki

This test validates the complete "batteries-included" telemetry deployment scenario.

### Prerequisites

As a prerequisite for this test, we assume:

1. an existing `MariaDB/Galera` entity in the target namespace
2. an existing `Keystone` deployed via keystone-operator
3. an existing `Memcached` instance in the target namespace
4. an existing `RabbitMQ` cluster in the target namespace
5. an existing `Heat` instance in the target namespace
6. a running `Loki Operator` in the cluster
7. Red Hat Observability Stack (RHOBS) components for metrics
8. S3-compatible storage (MinIO) for LokiStack backend
9. Certificate infrastructure for TLS
10. a running `telemetry-operator` deployed in the cluster

These resources are deployed via the `deps/` directory using kustomize.

### Run the default tests

Once the kuttl requirements are satisfied, the actual tests can be executed from the telemetry-operator root directory via:

```bash
cd telemetry-operator/
kubectl-kuttl test --test ceilometer --namespace telemetry-kuttl-default
```

Same tests can be executed from a different location via:

```bash
kubectl-kuttl test --config <KUTTL_CONFIG> <KUTTL_TESTS_DIR> --test ceilometer --namespace telemetry-kuttl-default
```

> **NOTE**: <KUTTL_CONFIG> is the path where kuttl-test.yaml for the repo is located and <KUTTL_TESTS_DIR> is where the tests are located.
