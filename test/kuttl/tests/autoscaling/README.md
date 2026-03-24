# autoscaling kuttl test

The "autoscaling" kuttl test covers the deployment and configuration of the Autoscaling CR, which manages Aodh, the OpenStack Alarming service used for autoscaling capabilities.

## Topology

The target topology for this test is:

1. one Autoscaling object managing Aodh service
2. Aodh StatefulSet with 1 replica containing 4 containers:
   - aodh-api
   - aodh-evaluator
   - aodh-notifier
   - aodh-listener
3. Keystone integration (KeystoneService and KeystoneEndpoint)
4. Internal and public service endpoints

The autoscaling test steps cover:
- Basic deployment with password-based authentication
- Custom configuration mounting
- Application credential authentication
- Application credential rotation

### Prerequisites

As a prerequisite for this test, we assume:

1. an existing `MariaDB/Galera` entity in the target namespace
2. an existing `Keystone` deployed via keystone-operator
3. an existing `Memcached` instance in the target namespace
4. an existing `RabbitMQ` cluster in the target namespace
5. an existing `Heat` instance in the target namespace
6. a running `telemetry-operator` deployed in the cluster

These resources are deployed via the `deps/` directory using kustomize.

### Run the autoscaling tests

Once the kuttl requirements are satisfied, the actual tests can be executed from the telemetry-operator root directory via:

```bash
cd telemetry-operator/
oc kuttl test --test autoscaling --namespace telemetry-kuttl-autoscaling
```

Same tests can be executed from a different location via:

```bash
oc kuttl test --config <KUTTL_CONFIG> <KUTTL_TESTS_DIR> --test autoscaling --namespace telemetry-kuttl-autoscaling
```

> **NOTE**: <KUTTL_CONFIG> is the path where kuttl-test.yaml for the repo is located and <KUTTL_TESTS_DIR> is where the tests are located.
