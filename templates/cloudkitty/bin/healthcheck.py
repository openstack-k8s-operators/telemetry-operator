#!/usr/bin/env python3
#
# Copyright 2022 Red Hat Inc.
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


from http import server
import signal
import socket
import sys
import time
import threading
import requests

from oslo_config import cfg


SERVER_PORT = 8080
CONF = cfg.CONF


class HTTPServerV6(server.HTTPServer):
    address_family = socket.AF_INET6


class HeartbeatServer(server.BaseHTTPRequestHandler):

    @staticmethod
    def check_services():
        print("Starting health checks")
        results = {}

        # Todo Database Endpoint Reachability
        # Keystone Endpoint Reachability
        try:
            keystone_uri = CONF.keystone_authtoken.auth_url
            response = requests.get(keystone_uri, timeout=5)
            response.raise_for_status()
            server_header = response.headers.get('Server', '').lower()
            if 'keystone' in server_header:
                results['keystone_endpoint'] = 'OK'
                print("Keystone endpoint reachable and responsive.")
            else:
                results['keystone_endpoint'] = 'WARN'
                print(f"Keystone endpoint reachable, but not a valid Keystone service: {keystone_uri}")
        except requests.exceptions.RequestException as e:
            results['keystone_endpoint'] = 'FAIL'
            print(f"ERROR: Keystone endpoint check failed: {e}")
            raise Exception('ERROR: Keystone check failed', e)

        # Prometheus Collector Endpoint Reachability
        try:
            prometheus_url = CONF.collector_prometheus.prometheus_url
            insecure = CONF.collector_prometheus.insecure
            cafile = CONF.collector_prometheus.cafile
            verify_ssl = cafile if cafile and not insecure else not insecure

            response = requests.get(prometheus_url, timeout=5, verify=verify_ssl)
            response.raise_for_status()
            results['collector_endpoint'] = 'OK'
            print("Prometheus collector endpoint reachable.")
        except requests.exceptions.RequestException as e:
            results['collector_endpoint'] = 'FAIL'
            print(f"ERROR: Prometheus collector check failed: {e}")
            raise Exception('ERROR: Prometheus collector check failed', e)

    def do_GET(self):
        try:
            self.check_services()
        except Exception as exc:
            self.send_error(500, exc.args[0], exc.args[1])
            return

        self.send_response(200)
        self.send_header("Content-type", "text/html")
        self.end_headers()
        self.wfile.write('<html><body>OK</body></html>'.encode('utf-8'))


def get_stopper(server):
    def stopper(signal_number=None, frame=None):
        print("Stopping server.")
        server.shutdown()
        server.server_close()
        print("Server stopped.")
        sys.exit(0)
    return stopper


if __name__ == "__main__":
    # Register config options
    cfg.CONF.register_group(cfg.OptGroup(name='database', title='Database connection options'))
    cfg.CONF.register_opt(cfg.StrOpt('connection', default=None), group='database')

    cfg.CONF.register_group(cfg.OptGroup(name='keystone_authtoken', title='Keystone Auth Token Options'))
    cfg.CONF.register_opt(cfg.StrOpt('auth_url',
                                    default='https://keystone-internal.openstack.svc:5000'),
                         group='keystone_authtoken')

    cfg.CONF.register_group(cfg.OptGroup(name='collector_prometheus', title='Prometheus Collector Options'))
    cfg.CONF.register_opt(cfg.StrOpt('prometheus_url',
                                    default='http://metric-storage-prometheus.openstack.svc:9090'),
                         group='collector_prometheus')
    cfg.CONF.register_opt(cfg.BoolOpt('insecure', default=False), group='collector_prometheus')
    cfg.CONF.register_opt(cfg.StrOpt('cafile', default=None), group='collector_prometheus')

    # Load configuration from file
    try:
        cfg.CONF(sys.argv[1:], default_config_files=['/etc/cloudkitty/cloudkitty.conf.d/cloudkitty.conf'])
    except cfg.ConfigFilesNotFoundError as e:
        print(f"Health check failed: {e}", file=sys.stderr)
        sys.exit(1)

    # Detect IPv6 support for binding
    hostname = socket.gethostname()
    try:
        ipv6_address = socket.getaddrinfo(hostname, None, socket.AF_INET6)
    except socket.gaierror:
        ipv6_address = None

    if ipv6_address:
        webServer = HTTPServerV6(("::", SERVER_PORT), HeartbeatServer)
    else:
        webServer = server.HTTPServer(("0.0.0.0", SERVER_PORT), HeartbeatServer)

    stop = get_stopper(webServer)

    # Need to run the server on a different thread because its shutdown method
    # will block if called from the same thread, and the signal handler must be
    # on the main thread in Python.
    thread = threading.Thread(target=webServer.serve_forever)
    thread.daemon = True
    thread.start()
    print(f"CloudKitty Healthcheck Server started http://{hostname}:{SERVER_PORT}")
    signal.signal(signal.SIGTERM, stop)

    try:
        while True:
            time.sleep(60)
    except KeyboardInterrupt:
        pass
    finally:
        stop()