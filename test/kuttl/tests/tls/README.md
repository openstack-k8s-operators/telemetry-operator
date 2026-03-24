# tls kuttl test

The "tls" kuttl test covers the deployment and configuration of TLS certificates for all telemetry services, ensuring secure communication between components and with external clients.

## Topology

The target topology for this test is:

1. Complete telemetry stack with TLS enabled:
   - Aodh with internal and public TLS endpoints
   - Ceilometer with TLS for HTTPS proxy
   - Prometheus with TLS for web interface
   - CloudKitty with internal and public TLS endpoints
2. Certificate infrastructure:
   - Combined CA bundle for trust
   - Individual certificates for each service
   - Certificates for internal and public endpoints
3. TLS volume mounts in all service pods
4. ScrapeConfigs with TLS configuration

### Prerequisites

As a prerequisite for this test, we assume:

1. cert-manager or OpenShift cert infrastructure
2. Certificate issuer for generating service certificates
3. All standard telemetry prerequisites:
   - MariaDB/Galera
   - Keystone
   - Memcached
   - RabbitMQ
   - Heat (for Aodh)
   - Loki Operator (for CloudKitty)
   - RHOBS (for MetricStorage)
4. S3-compatible storage (MinIO) for CloudKitty
5. a running `telemetry-operator` deployed in the cluster

These resources are deployed via the `deps/` directory using kustomize.

### Run the tls tests

Once the kuttl requirements are satisfied, the actual tests can be executed from the telemetry-operator root directory via:

```bash
cd telemetry-operator/
oc kuttl test --test tls --namespace telemetry-kuttl-tls
```

Same tests can be executed from a different location via:

```bash
oc kuttl test --config <KUTTL_CONFIG> <KUTTL_TESTS_DIR> --test tls --namespace telemetry-kuttl-tls
```

> **NOTE**: <KUTTL_CONFIG> is the path where kuttl-test.yaml for the repo is located and <KUTTL_TESTS_DIR> is where the tests are located.
