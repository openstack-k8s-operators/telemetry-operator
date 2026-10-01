#!/bin/bash
set -euo pipefail

if oc get crd lokistacks.loki.grafana.com &>/dev/null; then
    echo "loki-operator CRDs already present, skipping installation"
    exit 0
fi

CHANNEL=""
for i in $(seq 1 30); do
    CHANNEL=$(oc get packagemanifest loki-operator -n openshift-marketplace \
        --request-timeout=30s \
        --output jsonpath='{range .status.channels[*]}{.name}{"\n"}{end}' \
        | grep '^stable' | sort -V | tail -1) && [ -n "${CHANNEL}" ] && break
    echo "Waiting for loki-operator packagemanifest (attempt ${i}/30)..."
    sleep 10
done

if [ -z "${CHANNEL}" ]; then
    echo "ERROR: Could not resolve loki-operator channel from packagemanifests" >&2
    exit 1
fi

echo "Resolved loki-operator channel: ${CHANNEL}"

cat <<EOF | oc apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: openshift-operators-redhat
  labels:
    name: openshift-operators-redhat
---
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: loki-operator
  namespace: openshift-operators-redhat
spec:
  upgradeStrategy: Default
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: loki-operator
  namespace: openshift-operators-redhat
spec:
  channel: ${CHANNEL}
  installPlanApproval: Automatic
  name: loki-operator
  source: redhat-operators
  sourceNamespace: openshift-marketplace
EOF

echo "Waiting for loki-operator CRDs to be registered..."
until oc get crd lokistacks.loki.grafana.com &>/dev/null; do sleep 5; done
echo "loki-operator CRDs are available"
