# ceilometer kuttl test

The "ceilometer" kuttl test covers the deployment and configuration of the Ceilometer CR, which manages the OpenStack Metering service for collecting and processing measurements from OpenStack services.

## Topology

The target topology for this test is:

1. one Ceilometer object
2. Ceilometer StatefulSet with 1 replica containing 4 containers:
   - ceilometer-central-agent (collects metrics via polling)
   - ceilometer-notification-agent (processes notifications)
   - sg-core (Smart Gateway core for service graph)
   - proxy-httpd (HTTP proxy for metrics exposure)
3. kube-state-metrics StatefulSet (when ksmEnabled: true)
4. mysqld-exporter StatefulSet (when mysqldExporterEnabled: true)
5. Keystone integration (KeystoneService)

The ceilometer test steps cover:
- Basic deployment with exporters enabled
- Custom configuration mounting
- Application credential authentication
- Application credential rotation

### Prerequisites

As a prerequisite for this test, we assume:

1. an existing `Keystone` deployed via keystone-operator
2. an existing `RabbitMQ` cluster in the target namespace for notifications
3. a running `telemetry-operator` deployed in the cluster
4. access to MariaDB for mysqld-exporter (if enabled)

These resources are deployed via the `deps/` directory using kustomize.

### Run the ceilometer tests

Once the kuttl requirements are satisfied, the actual tests can be executed from the telemetry-operator root directory via:

```bash
cd telemetry-operator/
kubectl-kuttl test --test ceilometer --namespace telemetry-kuttl-ceilometer
```

Same tests can be executed from a different location via:

```bash
kubectl-kuttl test --config <KUTTL_CONFIG> <KUTTL_TESTS_DIR> --test ceilometer --namespace telemetry-kuttl-ceilometer
```

> **NOTE**: <KUTTL_CONFIG> is the path where kuttl-test.yaml for the repo is located and <KUTTL_TESTS_DIR> is where the tests are located.
