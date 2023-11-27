#!/bin/bash

# taken from https://github.com/openstack-k8s-operators/nova-operator/blob/main/hack/install-krew.sh
(
    set -x
    tdir="$(mktemp -d)"
    pushd $tdir &&
    OS="$(uname | tr '[:upper:]' '[:lower:]')" &&
    ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/\(arm\)\(64\)\?.*/\1\2/' -e 's/aarch64$/arm64/')" &&
    KREW="krew-${OS}_${ARCH}" &&
    curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/${KREW}.tar.gz" &&
    tar zxvf "${KREW}.tar.gz" &&
    ./"${KREW}" install krew
    popd && rm -fr $tdir
)