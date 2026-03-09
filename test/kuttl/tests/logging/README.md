# logging kuttl test

The "logging" kuttl test covers the deployment and configuration of the Logging CR, which manages log forwarding from OpenStack compute nodes to a centralized logging infrastructure.

## Topology

The target topology for this test is:

1. one Logging object managing log collection configuration
2. LoadBalancer service for receiving logs from compute nodes
3. Integration with Cluster Logging Operator (CLO) via Vector
4. Configuration secrets for compute nodes

The logging test validates:
- Service endpoint creation for log ingestion
- Proper MetalLB configuration for load balancing
- Compute node configuration secret generation

### Prerequisites

As a prerequisite for this test, we assume:

1. MetalLB or equivalent LoadBalancer provider in the cluster
2. Cluster Logging Operator (CLO) deployed
3. Vector collector pods managed by CLO
4. a running `telemetry-operator` deployed in the cluster

These resources are deployed via the `deps/` directory using kustomize.

### Run the logging tests

Once the kuttl requirements are satisfied, the actual tests can be executed from the telemetry-operator root directory via:

```bash
cd telemetry-operator/
kubectl-kuttl test --test logging --namespace telemetry-kuttl-logging
```

Same tests can be executed from a different location via:

```bash
kubectl-kuttl test --config <KUTTL_CONFIG> <KUTTL_TESTS_DIR> --test logging --namespace telemetry-kuttl-logging
```

> **NOTE**: <KUTTL_CONFIG> is the path where kuttl-test.yaml for the repo is located and <KUTTL_TESTS_DIR> is where the tests are located.
