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

DATAPLANE_TOTAL_NODES=2 make edpm_compute
```

3.- Deploy openstack-operator and openstack
```
cd ..
make crc_storage
make input

make openstack
make openstack_deploy
```

4.- Deploy dataplane operator
```
DATAPLANE_TOTAL_NODES=2 DATAPLANE_NTP_SERVER=clock.redhat.com make edpm_deploy
```
To know when dataplane-operator finishes, you have to keep looking at "*-edpm" pods that keep appearing to run ansible on the compute nodes. They will appear one after the other. When those stop appearing, it is finished and we have a default openstack environment.

You can also make your process wait until everything has run:
```
DATAPLANE_TOTAL_NODES=2 DATAPLANE_NTP_SERVER=clock.redhat.com make edpm_deploy_wait
```

Now, we proceed to run our own telemetry-operator instance:

5.- Remove Ceilometer deployment
```
oc patch openstackcontrolplane openstack-galera-network-isolation --type='json' -p='[{"op": "replace", "path": "/spec/ceilometer/enabled", "value":false}]'
```

6.- Remove telemetry-operator from the deployments
```
oc project openstack-operators
oc remove csv telemetry-operator.v0.0.1
```

7.- Deploy custom telemetry-operator version
```
cd telemetry-operator

oc delete -f config/crd/bases/
oc apply -f config/crd/bases/

make manifests generate
OPERATOR_TEMPLATES=$PWD/templates make run
```

8.- Deploy Telemetry:
```
oc apply -f config/samples/telemetry_v1beta1_telemetry.yaml
```

## Testing edpm-ansible changes

1.- Build your custom `openstack-ansibleee-runner` image using these [steps](https://github.com/openstack-k8s-operators/edpm-ansible/tree/main#build-and-push-the-openstack-ansibleee-runner-container-image) and push it to a registry

2.- Override `DATAPLANE_RUNNER_IMG` and `ANSIBLEEE_IMAGE_URL_DEFAULT` when running `edpm_deploy`
```
cd ~/install_yamls/
DATAPLANE_RUNNER_IMG=<url_to_custom_image> ANSIBLEEE_IMAGE_URL_DEFAULT=<url_to_custom_image> make edpm_deploy
```

3.- During deployment `dataplane-deployment-*` pods would get spawned with the custom image.

## Destroy the environment to start again
```
cd install_yamls/devsetup

# Delete the CRC node
make crc_cleanup

# Destroy edpm VMS
EDPM_TOTAL_NODES=2 make edpm_compute_cleanup
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
