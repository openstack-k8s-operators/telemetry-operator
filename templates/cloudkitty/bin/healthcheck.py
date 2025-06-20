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

# Trivial HTTP server to check health of scheduler, backup and volume services.
# Cinder-API hast its own health check endpoint and does not need this.
#
# The only check this server currently does is using the heartbeat in the
# database service table, accessing the DB directly here using cinder's
# configuration options.
#
# The benefit of accessing the DB directly is that it doesn't depend on the
# Cinder-API service being up and we can also differentiate between the
# container not having a connection to the DB and the cinder service not doing
# the heartbeats.
#
# For volume services all enabled backends must be up to return 200, so it is
# recommended to use a different pod for each backend to avoid one backend
# affecting others.
#
# Requires the name of the service as the first argument (volume, backup,
# scheduler) and optionally a second argument with the location of the
# configuration directory (defaults to /etc/cinder/cinder.conf.d)

from http import server
import signal
import socket
import sys
import time
import threading

from oslo_config import cfg

SERVER_PORT = 8080
CONF = cfg.CONF

class HTTPServerV6(server.HTTPServer):
  address_family = socket.AF_INET6

class HeartbeatServer(server.BaseHTTPRequestHandler):
    def do_GET(self):
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
    hostname = socket.gethostname()
    ipv6_address = socket.getaddrinfo(hostname, None, socket.AF_INET6)
    if ipv6_address:
        webServer = HTTPServerV6(("::",SERVER_PORT), HeartbeatServer)
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
