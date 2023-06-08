# telemetry-operator
The Telemetry Operator handles the deployment of all the needed agents for gathering telemetry to assess the full state of a running Openstack cluster.

## Description
This operator deploys a multiple telemetry agents, both in the control plane and in the dataplane nodes.

## Dev setup
1.- Deploy crc:
```
cd install_yamls/devsetup
CPUS=12 MEMORY=25600 DISK=100 make crc
```

2.- Create edpm nodes
```
make crc_attach_default_interface

EDPM_COMPUTE_SUFFIX=0 make edpm_compute
EDPM_COMPUTE_SUFFIX=1 make edpm_compute
EDPM_COMPUTE_SUFFIX=0 make edpm_compute_repos
EDPM_COMPUTE_SUFFIX=1 make edpm_compute_repos
```

3.- Deploy openstack-operator
```
cd ..
make crc_storage
make input

make openstack
```

4.- Disable stock telemetry-operator
```
oc project openstack-operators

export OPERATOR_NAME="telemetry-operator"
OPERATOR_INDEX=(oc get csv openstack-operator.v0.0.1 -o json | jq -r '.spec.install.spec.deployments | map(.name == $ENV.OPERATOR_NAME + "-controller-manager") | index(true)'

oc patch csv openstack-operator.v0.0.1 --type json \
  -p="[{"op": "remove", "path": "/spec/install/spec/deployments/${OPERATOR_INDEX}"}]"

oc delete deployment telemetry-operator-controller-manager
```

5.- Deploy openstack and disable telemetry
```
make openstack_deploy

oc edit openstackcontrolplane
      Search for "telemetry"
	  Change "enabled" to false
```

6.- Deploy dataplane operator
```
DATAPLANE_SINGLE_NODE=false DATAPLANE_CHRONY_NTP_SERVER=clock.redhat.com make edpm_deploy
```

7.- Deploy custom telemetry-operator version
```
cd telemetry-operator

oc delete -f config/crd/bases/
oc apply -f config/crd/bases/

make manifests generate
OPERATOR_TEMPLATES=$PWD/templates OPERATOR_PLAYBOOKS=$PWD/playbooks make run
```

8.- Deploy a telemetry in the cluster
```
oc apply -f config/samples/telemetry_v1beta1_telemetry.yaml
```

## Emergency rescue access to CRC VM
If you need to connect directly to the CRC VM just use
```
ssh -i ~/.crc/machines/crc/id_ecdsa core@"192.168.130.11"
```

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
