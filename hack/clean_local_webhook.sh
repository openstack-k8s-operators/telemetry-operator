#!/bin/bash
set -ex

oc delete validatingwebhookconfiguration/vautoscaling.kb.io --ignore-not-found
oc delete mutatingwebhookconfiguration/mautoscaling.kb.io --ignore-not-found
oc delete validatingwebhookconfiguration/vceilometer.kb.io --ignore-not-found
oc delete mutatingwebhookconfiguration/mceilometer.kb.io --ignore-not-found
oc delete validatingwebhookconfiguration/vtelemetry.kb.io --ignore-not-found
oc delete mutatingwebhookconfiguration/mtelemetry.kb.io --ignore-not-found
