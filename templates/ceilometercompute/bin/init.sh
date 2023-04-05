#!/bin/bash
#
# Copyright 2023 Red Hat Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License"); you may
# not use this file except in compliance with the License. You may obtain
# a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
# WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
# License for the specific language governing permissions and limitations
# under the License.
set -ex

# Secrets are obtained from ENV variables.
export RABBITMQ_HOST=${RabbitMQHost:?"Please specify a RabbitMQHost variable."}
export RABBITMQ_USER=${RabbitMQUsername:?"Please specify a RabbitMQUsername variable."}
export RABBITMQ_PASS=${RabbitMQPassword:?"Please specify a RabbitMQPassword variable."}
export CEILOMETER_PASS=${CeilometerPassword:?"Please specify a CeilometerPassword variable."}

SVC_CFG=/etc/ceilometer/ceilometer.conf
SVC_CFG_MERGED=/var/lib/config-data/merged/ceilometer.conf

# expect that the common.sh is in the same dir as the calling script
SCRIPTPATH="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
. ${SCRIPTPATH}/common.sh --source-only

# Copy default service config from container image as base
cp -a ${SVC_CFG} ${SVC_CFG_MERGED}

# Merge all templates from config CM
for dir in /var/lib/config-data/default; do
    merge_config_dir ${dir}
done

# set secrets
crudini --set ${SVC_CFG_MERGED} DEFAULT transport_url rabbit://${RABBITMQ_USER}:${RABBITMQ_PASS}@${RABBITMQ_HOST}:5672/?ssl=0
crudini --set ${SVC_CFG_MERGED} oslo_messaging_notifications transport_url rabbit://${RABBITMQ_USER}:${RABBITMQ_PASS}@${RABBITMQ_HOST}:5672/?ssl=0
crudini --set ${SVC_CFG_MERGED} service_credentials password ${CEILOMETER_PASS}
