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
#    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
#    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
#    License for the specific language governing permissions and limitations
#    under the License.

import datetime
import json
import os
import psutil
import socket
import sys


HEARTBEAT_SOCKET = "/var/lib/ceilometer/ceilometer-central.socket"
POLL_SPREAD = datetime.timedelta(minutes=15)

CONN_SPREAD = datetime.timedelta(hours=1)
PROCESS_NAME = "ceilometer-polling"
PROCESS_CONN_CACHE = "/var/lib/ceilometer/tmp/connections"
# Keystosne 5000
# Cinder 8776
# Glance 9292
# Neutron 9696
PROCESS_PORTS = [5000, 8776, 9292, 9696]


def check_pollsters(socket_path: str) -> tuple[int, str]:
    """
    Returns 0 if socket_path content contains all records
    of polling no older than POLL_SPREAD, otherwise returns 1.
    """
    s = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
    s.connect(socket_path)

    report = ""
    try:
        while True:
            data = s.recv(2048)
            if data:
                report += data.decode('utf-8')
            else:
                break
    finally:
        s.close()

    limit = datetime.datetime.today() - POLL_SPREAD
    for line in report.split('\n'):
        if not line:
            # Skip empty line, which might occur for cases
            # when no pollster had polled yet. In such case
            # we assume agent healthy as at least the socket
            # has been successfully created.
            continue
        pollster, timestr = line.split()
        timestamp = datetime.datetime.fromisoformat(timestr)

        if timestamp < limit:
            return 1, f"{pollster}'s timestamp is out of limit"
    return 0, ""


def check_connection(cache_path: str,
                     process_name: str,
                     ports: list[int]) -> tuple[int, str]:
    """
    Returns 0 if given process is or was recently connected
    to the given port(s). Otherwise returns 1 or 2 with respective
    reason of failure.
    """
    # NOT(mmagr): process' connections can be empty if we hit the case
    #             when no pollster was polling from some API during running
    #             this script. The previous values will be checked from cache.
    #             Hence to avoid false negatives we need much higher spread
    #             than in case of heart beat check

    # load connection cache
    if os.path.isfile(cache_path):
        with open(cache_path, "r") as cch:
            conns = json.load(cch)
    else:
        conns = dict()

    # update connection cache values
    processes = [proc for proc in psutil.process_iter()
                    if process_name in proc.name()]
    if not processes:
        return 1, f"Given process {process_name} was not found"
    ports = set(ports)
    for p in processes:
        conn_method = getattr(p, "net_connections", p.connections)
        for c in conn_method():
            if c.raddr.port not in ports:
                continue
            key = f"{c.raddr.ip}/{c.raddr.port}"
            conns[key] = dict(ts=datetime.datetime.now().timestamp(),
                              ip=c.raddr.ip,
                              port=c.raddr.port)
    with open(cache_path, "w") as cch:
        json.dump(conns, cch)

    # check connection timestamps in the cache
    limit = datetime.datetime.today() - CONN_SPREAD
    for conn in conns.values():
        timestamp = datetime.datetime.fromtimestamp(conn["ts"])
        if timestamp < limit:
            msg = (f"Timestamp of connection to {conn['ip']}"
                   f" on port {conn['port']} is out of limit")
            return 1, msg

    return 0, ""


if __name__ == "__main__":
    try:
        if os.path.isfile(HEARTBEAT_SOCKET):
            # verify pollsters' heart beats
            rc, reason = check_pollsters(HEARTBEAT_SOCKET)
        else:
            # mimic heart beats check on process' connections
            # to various OpenStack APIs when the heart beat feature
            # is not available
            rc, reason = check_connection(PROCESS_CONN_CACHE,
                                          PROCESS_NAME,
                                          PROCESS_PORTS)
    except Exception as ex:
        rc, reason = 2, f"Unkown error: {ex}"

    if rc != 0:
        print(reason)
    sys.exit(rc)
