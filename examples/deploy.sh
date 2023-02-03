#/bin/bash
set -e

make docker-build docker-push IMG="quay.io/jlarriba/ceilometer-operator"
rm -rf /home/jlarriba/Development/ceilometer-operator/bin
make deploy IMG="quay.io/jlarriba/ceilometer-operator"
