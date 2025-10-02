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
import psutil


def check_process() -> tuple[int, str]:
    # Return 0 if cloudkitty-processor process with given cmdline exists, else 1 with reason.
    for proc in psutil.process_iter(attrs=["name", "cmdline"]):
        try:
            cmdline = proc.info.get("cmdline", [])
            if cmdline and any("cloudkitty-processor" in arg for arg in cmdline):
                return 0, ""
        except (psutil.NoSuchProcess, psutil.AccessDenied):
            continue
    return 1, "CloudKitty processor process not found"


def run_checks() -> tuple[int, str]:
    # Run all health checks and return aggregated result
    checks = [check_process]
    for check in checks:
        rc, reason = check()
        if rc != 0:
            return rc, reason
    return 0, ""


if __name__ == "__main__":
    try:
        rc, reason = run_checks()
    except Exception as ex:
        rc, reason = 2, f"Unknown error: {ex}"

    if rc != 0:
        print(reason)
    sys.exit(rc)