# topology kuttl test

The "topology" kuttl test covers the deployment of telemetry services with Kubernetes topology spread constraints, ensuring proper pod distribution across nodes for high availability.

## Topology

The target topology for this test is:

1. one Topology object defining spread constraints:
   - maxSkew: 1
   - topologyKey: `topology.kubernetes.io/hostname`
   - whenUnsatisfiable: ScheduleAnyway
2. Complete telemetry stack with topology constraints applied:
   - Ceilometer with topology constraints
   - mysqld-exporter with topology constraints
   - kube-state-metrics with topology constraints
   - Aodh with topology constraints
   - CloudKitty API with topology constraints
   - CloudKitty Proc with topology constraints
3. Each StatefulSet configured with appropriate labelSelector

The topology test validates:
- Topology CR creation and propagation
- StatefulSet configuration with topologySpreadConstraints
- Status tracking of applied topology in service CRs

### Prerequisites

As a prerequisite for this test, we assume:

1. A Kubernetes cluster with multiple nodes (for meaningful topology spread)
2. topology.openstack.org/v1beta1 API available
3. All standard telemetry prerequisites:
   - MariaDB/Galera
   - Keystone
   - Memcached
   - RabbitMQ
   - Heat
   - Loki Operator
   - RHOBS
4. S3-compatible storage (MinIO) for CloudKitty
5. a running `telemetry-operator` deployed in the cluster

These resources are deployed via the `deps/` directory using kustomize.

### Run the topology tests

Once the kuttl requirements are satisfied, the actual tests can be executed from the telemetry-operator root directory via:

```bash
cd telemetry-operator/
kubectl-kuttl test --test topology --namespace telemetry-kuttl-topology
```

Same tests can be executed from a different location via:

```bash
kubectl-kuttl test --config <KUTTL_CONFIG> <KUTTL_TESTS_DIR> --test topology --namespace telemetry-kuttl-topology
```

> **NOTE**: <KUTTL_CONFIG> is the path where kuttl-test.yaml for the repo is located and <KUTTL_TESTS_DIR> is where the tests are located.
