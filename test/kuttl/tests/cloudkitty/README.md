# cloudkitty kuttl test

The "cloudkitty" kuttl test covers the deployment and configuration of the CloudKitty CR, which manages the OpenStack Rating service for cloud resources.

## Topology

The target topology for this test is:

1. one CloudKitty object managing two sub-components:
   - CloudKittyAPI (API service with 1 replica)
   - CloudKittyProc (rating processor with 1 replica)
2. LokiStack for metrics storage backend
3. CloudKittyAPI pod containing:
   - cloudkitty-api container
   - log sidecar container
4. CloudKittyProc pod containing:
   - cloudkitty-proc container
5. S3 storage (MinIO) for LokiStack backend
6. TLS certificates for Loki communication
7. Keystone integration (KeystoneService and KeystoneEndpoint)

The cloudkitty test steps cover:
- Basic deployment with LokiStack integration
- Custom configuration mounting for both API and Proc
- Application credential authentication
- Application credential rotation

### Prerequisites

As a prerequisite for this test, we assume:

1. an existing `MariaDB/Galera` entity in the target namespace
2. an existing `Keystone` deployed via keystone-operator
3. an existing `Memcached` instance in the target namespace
4. an existing `RabbitMQ` cluster in the target namespace
5. a running `Loki Operator` in the cluster
6. S3-compatible storage (MinIO) for LokiStack backend
7. a running `telemetry-operator` deployed in the cluster

These resources are deployed via the `deps/` directory using kustomize.

**Note:** This test has a longer timeout (600s) because LokiStack can take significant time to become ready.

### Run the cloudkitty tests

Once the kuttl requirements are satisfied, the actual tests can be executed from the telemetry-operator root directory via:

```bash
cd telemetry-operator/
oc kuttl test --test cloudkitty --namespace telemetry-kuttl-cloudkitty
```

Same tests can be executed from a different location via:

```bash
oc kuttl test --config <KUTTL_CONFIG> <KUTTL_TESTS_DIR> --test cloudkitty --namespace telemetry-kuttl-cloudkitty
```

> **NOTE**: <KUTTL_CONFIG> is the path where kuttl-test.yaml for the repo is located and <KUTTL_TESTS_DIR> is where the tests are located.
