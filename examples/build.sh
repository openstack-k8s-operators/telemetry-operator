#!/bin/bash
set -e

echo "UNDEPLOY"
#oc delete -f config/crd/bases/ceilometer.openstack.org_ceilometers.yaml

echo "BUILD"
make generate
make manifests
make build

echo "DEPLOY"
oc apply -f config/crd/bases/ceilometer.openstack.org_ceilometers.yaml
./bin/manager
