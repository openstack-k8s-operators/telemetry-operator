#!/bin/bash
set -ex

oc delete validatingwebhookconfiguration/vtelemetry.kb.io --ignore-not-found
oc delete mutatingwebhookconfiguration/mtelemetry.kb.io --ignore-not-found
