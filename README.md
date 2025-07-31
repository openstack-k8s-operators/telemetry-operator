# telemetry-operator

The Telemetry Operator handles the deployment of all the needed agents for gathering telemetry to assess the full state of a running Openstack cluster.

## Description

This operator deploys a multiple telemetry agents, both in the control plane and in the dataplane nodes.

## Development setup with install_yamls

1. Deploy crc:

    ```bash
    cd install_yamls/devsetup
    CPUS=12 MEMORY=25600 DISK=100 make crc
    ```

2. Create edpm nodes

    ```bash
    make crc_attach_default_interface

    EDPM_TOTAL_NODES=2 make edpm_compute
    ```

3. Deploy openstack-operator and openstack

    ```bash
    cd ..
    make crc_storage
    make input

    make openstack
    make openstack_init
    make openstack_deploy
    ```

4. Deploy dataplane operator

    ```bash
    DATAPLANE_TOTAL_NODES=2 DATAPLANE_NTP_SERVER=clock.redhat.com make edpm_deploy
    ```

    To know when dataplane-operator finishes, you have to keep looking at "*-edpm" pods that keep appearing to run ansible on the compute nodes. They will appear one after the other. When those stop appearing, it is finished.

    You can also make your process wait until everything finishes:

    ```bash
    DATAPLANE_TOTAL_NODES=2 DATAPLANE_NTP_SERVER=clock.redhat.com make edpm_wait_deploy
    ```

5. Refresh Nova discover hosts

    ```bash
    make edpm_nova_discover_hosts
    ```

    Now, we proceed to run our own telemetry-operator instance:

6. Remove Telemetry deployment

    ```bash
    oc patch openstackcontrolplane openstack-galera-network-isolation --type='json' -p='[{"op": "replace", "path": "/spec/telemetry/enabled", "value":false}]'
    ```

7. Scale down telemetry-operator and openstack-operator

    ```bash
    oc scale deployments/openstack-operator-controller-operator --replicas 0 -n openstack-operators

    oc scale deployments/openstack-operator-controller-manager --replicas 0 -n openstack-operators

    oc scale deployments/telemetry-operator-controller-manager --replicas 0 -n openstack-operators
    ```

8. Deploy custom telemetry-operator version

    NOTE: If you intend to deploy a custom telemetry object *with pre-populated image URLs*, you can use `make run` instead of `make run-with-webhook`, because the webhooks will not be required.

    ```bash
    cd telemetry-operator

    # Only needed if API changed
    oc delete -f config/crd/bases/
    oc apply -f config/crd/bases/

    make manifests generate
    OPERATOR_TEMPLATES=$PWD/templates make run-with-webhook
    ```

9. Deploy Telemetry

    There are two options, either use the existing OpenStackControlPlane to manage the telemetry object, or manage it yourself.

    9.a. To use the existing OpenStackControlPlane to manage the telemetry object, re-enable telemetry (preferred way, unless necessary for some reason):

    ```bash
    oc patch openstackcontrolplane openstack-galera-network-isolation --type='json' -p='[{"op": "replace", "path": "/spec/telemetry/enabled", "value":true}]'
    ```

    9.b. To use a custom telemetry object

    >NOTE: Make sure to use a different name other than "telemetry" in your custom object
    >
    >Otherwise the openstack-operator will remove the object automatically

    ```bash
    oc apply -f config/samples/telemetry_v1beta1_telemetry.yaml
    ```

## Run custom telemetry-operator bundle as part of openstack-operator

There are times where deploying a dev environment with the existing telemetry-operator and then replace it is not enough. For example, when changing the API its always good to be able to check if the OpenStackControlPlane can still be applied or there is something wrong.

For this, follow the procedure:

1. Create three repositories for telemetry-operator in your quay.io personal account: **telemetry-operator**, **telemetry-operator-bundle** and **telemetry-operator-index**. The three of them must be public.

2. Create three repositories for openstack-operator in your quay.io personal account: **openstack-operator**, **openstack-operator-bundle** and **openstack-operator-index**. The three of them must be public.

3. Commit your changes in telemetry-operator and push the commit to the fork in your personal repository.

4. Introduce a replace rule in openstack-operator/go.mod and openstack-operator/apis/go.mod like this:

    ```go
    replace github.com/openstack-k8s-operators/telemetry-operator/api => github.com/<github_user>/telemetry-operator/api <commit_id>
    ```

    And run

    ```bash
    make tidy
    ```

    This would make the line look something like this:

    ```bash
    replace github.com/openstack-k8s-operators/telemetry-operator/api => github.com/<github_user>/telemetry-operator/api v0.1.1-0.20240715084507-c8fd68f4cc2c
    ```

5. Build telemetry-operator bundle, push it to your quay.io account and create a tag for the bundle with your commit id pointing to the version you have just built. This is used to find the exact sha of the image that is being to be used:

    ```bash
    IMAGE_TAG_BASE=quay.io/<quay_user>/telemetry-operator VERSION=0.0.2 IMG=$IMAGE_TAG_BASE:v$VERSION make manifests build docker-build docker-push bundle bundle-build bundle-push catalog-build catalog-push

    podman tag quay.io/<quay_user>/telemetry-operator-bundle:v0.0.2 quay.io/<quay_user>/telemetry-operator-bundle:<commit_id>
    podman push quay.io/<quay_user>/telemetry-operator-bundle:<commit_id>
    ```

6. Build openstack-operator bundle and push it to your quay.io account:

    ```bash
    IMAGE_TAG_BASE=quay.io/<quay_user>/openstack-operator VERSION=0.0.2 IMG=$IMAGE_TAG_BASE:v$VERSION make manifests build docker-build docker-push bundle bundle-build bundle-push catalog-build catalog-push
    ```

7. Deploy openstack with the recently build openstack-operator image and then deploy:

    ```bash
    OPENSTACK_IMG=quay.io/<quay_user>/openstack-operator-index:v0.0.2 make openstack

    make openstack_deploy
    ```

> NOTE: This is if your quay.io account and your github account are identical.
>
> If you are using different names, you must use *IMAGENAMESPACE=<quay_user>* in the bundle building commands, both for telemetry and openstack, like this:
>
```bash
IMAGENAMESPACE=<quay_user> IMAGE_TAG_BASE=quay.io/<quay_user>/openstack-operator VERSION=0.0.2 IMG=$IMAGE_TAG_BASE:v$VERSION make manifests build docker-build docker-push bundle bundle-build bundle-push catalog-build catalog-push
```

## Connect to Dataplane nodes

You can connect directly to the compute nodes using password 12345678:

```bash
ssh root@192.168.122.100

ssh root@192.168.122.101
```

## Testing edpm-ansible changes

1. Build your custom `openstack-ansibleee-runner` image using these [steps](https://github.com/openstack-k8s-operators/edpm-ansible/tree/main#build-and-push-the-openstack-ansibleee-runner-container-image) and push it to a registry

2. Override `DATAPLANE_RUNNER_IMG` and `ANSIBLEEE_IMAGE_URL_DEFAULT` when running `edpm_deploy`

    ```bash
    cd ~/install_yamls/
    DATAPLANE_RUNNER_IMG=<url_to_custom_image> ANSIBLEEE_IMAGE_URL_DEFAULT=<url_to_custom_image> make edpm_deploy
    ```

3. During deployment `dataplane-deployment-*` pods would get spawned with the custom image.

## Testing edpm-ansible changes using a volume mount backed by NFS server (tested)

The following procedure is to be performed after the

1. Follow all these [steps](https://openstack-k8s-operators.github.io/edpm-ansible/testing_with_ansibleee.html#provide-nfs-access-to-your-edpm-ansible-directory) to add a PVC to OpenStackDataPlaneNodeSet’s CR.

2. Deploy a debug pod with the PVC to it to verify whether your local repository was mounted into it

    ```bash
    oc apply -f - <<EOF
    ---
    apiVersion: v1
    kind: Pod
    metadata:
      name: debug-pod
      labels:
        app: debug
    spec:
      containers:
        - name: debug-container
          image: busybox  # Replace with your preferred image for debugging
          command: ["/bin/sh", "-c", "sleep 3600"]  # Keeps the pod alive for an hour
          volumeMounts:
            - mountPath: /mnt
              name: pvc-volume
      volumes:
        - name: pvc-volume
          persistentVolumeClaim:
            claimName: edpm-ansible-dev  # Name of the PVC
      restartPolicy: Never  # Ensures the pod doesn't restart automatically
    EOF
    ```

    Verify whether repository exists

    ```bash
    $ oc exec -it debug-pod sh -- ls /mnt
    CHANGELOG.md               OWNERS_ALIASES             contribute                 molecule                   plugins                    tests
    LICENSE                    README.md                  docs                       molecule-requirements.txt  requirements.yml           zuul.d
    Makefile                   app-root                   galaxy.yml                 openstack_ansibleee        roles
    OWNERS                     bindep.txt                 meta                       playbooks                  scripts
    ```

    Delete the pod once verified

3. Once the [step](https://openstack-k8s-operators.github.io/edpm-ansible/testing_with_ansibleee.html#add-extramount-to-your-openstackdataplanenodeset-cr) to add `extraMount` to `OpenStackDataPlaneNodeSet CR` is executed, the deployment is reported as `Deployment not started` which is expected.

4. Create a new EDPM deployment using the existing `nodeSets`

    ```bash
    oc apply -f - <<EOF
    apiVersion: dataplane.openstack.org/v1beta1
    kind: OpenStackDataPlaneDeployment
    metadata:
      name: edpm-deployment-debug
    spec:
      nodeSets:
        - openstack-edpm-ipam
    EOF
    ```

    Deployment progresses using the `edpm-ansible` repository from the NFS mount

    ```bash
    oc get osdpns
    NAME                  STATUS   MESSAGE
    openstack-edpm-ipam   False    Deployment in progress
    ```

## Running kuttl tests locally

For the default suite, simply run `make kuttl-test`.

For standalone suites, you must follow these steps:

1. Set up the testing namespace using install_yamls: `cd install_yamls && make kuttl_common_prep heat heat_deploy NAMESPACE=telemetry-kuttl-tests`
2. (Optionally) Edit the list of suites to run: `cd telemetry-operator && vi kuttl-test.yaml` (Comment out any suites you don't need from the testDirs list)
3. Run kuttl specifying that config file and namespace: `kubectl-kuttl test --config ./kuttl-test.yaml --namespace telemetry-kuttl-tests`

> NOTE: (May 2024) These tests appear very reliable when running a single suite, but occasional flakiness (~ 25% failure) has been observed when they all run serially. This problem appears to be order/timing related and may or may not affect the automated CI.

## Destroy the environment to start again

```bash
cd install_yamls/devsetup

# Delete the CRC node
make crc_cleanup

# Destroy edpm VMS
EDPM_TOTAL_NODES=2 make edpm_compute_cleanup
```

## Emergency rescue access to CRC VM

If you need to connect directly to the CRC VM just use

```bash
ssh -i ~/.crc/machines/crc/id_ecdsa core@"192.168.130.11"
```

## License

```text
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
```
