#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025 Inc.
# All Rights Reserved.
#
#    Licensed under the Apache License, Version 2.0 (the "License"); you may
#    not use this file except in compliance with the License. You may obtain
#    a copy of the License at
#
#         http://www.apache.org/licenses/LICENSE-2.0
#
#    Unless required by applicable law or agreed to in writing, software
#    distributed under the License is distributed on an "AS IS" BASIS,
#    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#    See the License for the specific language governing permissions and
#    limitations under the License.

import sys
import requests
from oslo_config import cfg
import psutil


CONF = cfg.CONF
PROCESS_NAME = "cloudkitty-proc"

def check_process() -> tuple[int, str]:
    # Return 0 if process with given name exists, else 1 with reason.
    for proc in psutil.process_iter(attrs=["name"]):
        try:
            if proc.info["name"] == PROCESS_NAME:
                return 0, ""
        except (psutil.NoSuchProcess, psutil.AccessDenied):
            continue
    return 1, f"Process {PROCESS_NAME} not found"

def check_keystone() -> tuple[int, str]:
   # Check Keystone endpoint reachability
    try:
        keystone_uri = CONF.keystone_authtoken.auth_url
        response = requests.get(keystone_uri, timeout=5)
        response.raise_for_status()
        server_header = response.headers.get("Server", "").lower()
        if "keystone" in server_header:
            return 0, ""
        return 1, f"Keystone reachable but not identified as Keystone: {keystone_uri}"
    except requests.exceptions.RequestException as e:
        return 1, f"Keystone endpoint check failed: {e}"


def run_checks() -> tuple[int, str]:
    # Run all health checks and return aggregated result
    checks = [check_process, check_keystone]
    for check in checks:
        rc, reason = check()
        if rc != 0:
            return rc, reason
    return 0, ""


if __name__ == "__main__":
    cfg.CONF.register_group(cfg.OptGroup(name="keystone_authtoken", title="Keystone Auth Token Options"))
    cfg.CONF.register_opt(
        cfg.StrOpt("auth_url", default="https://keystone-internal.openstack.svc:5000"),
        group="keystone_authtoken",
    )

    try:
        cfg.CONF(sys.argv[1:])
    except cfg.ConfigFilesNotFoundError as e:
        print(f"Config load failed: {e}")
        sys.exit(2)

    try:
        rc, reason = run_checks()
    except Exception as ex:
        rc, reason = 2, f"Unknown error: {ex}"

    if rc != 0:
        print(reason)
    sys.exit(rc)