package controller

import (
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/onsi/gomega"
)

// renderAodhConf renders templates/autoscaling/config/aodh.conf with the given
// template parameters and returns the resulting configuration.
func renderAodhConf(t *testing.T, params map[string]any) string {
	t.Helper()

	content, err := os.ReadFile("../../templates/autoscaling/config/aodh.conf")
	if err != nil {
		t.Fatalf("failed to read aodh.conf template: %v", err)
	}

	tmpl, err := template.New("aodh.conf").Parse(string(content))
	if err != nil {
		t.Fatalf("failed to parse aodh.conf template: %v", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, params); err != nil {
		t.Fatalf("failed to execute aodh.conf template: %v", err)
	}

	return buf.String()
}

// baseAodhParams returns the minimal set of template parameters needed to render
// the [keystone_authtoken] memcached configuration.
func baseAodhParams() map[string]any {
	return map[string]any{
		"MemcachedServers":         "aodh-memcached-0.svc:11212,aodh-memcached-1.svc:11212",
		"MemcachedServersWithInet": "inet6:[aodh-memcached-0.svc]:11211,inet6:[aodh-memcached-1.svc]:11211",
	}
}

// TestAodhConfKeystoneAuthtokenMemcachedNoTLS verifies that when memcached does
// not run with TLS, [keystone_authtoken] uses the inet-prefixed server list and
// does not emit any memcache TLS settings.
func TestAodhConfKeystoneAuthtokenMemcachedNoTLS(t *testing.T) {
	g := gomega.NewWithT(t)

	params := baseAodhParams()
	params["MemcachedTLS"] = false

	config := renderAodhConf(t, params)

	g.Expect(config).To(gomega.ContainSubstring(
		"memcached_servers=" + params["MemcachedServersWithInet"].(string)))
	g.Expect(config).NotTo(gomega.ContainSubstring("memcache_tls_enabled"))
	g.Expect(config).NotTo(gomega.ContainSubstring("memcache_tls_certfile"))
	g.Expect(config).NotTo(gomega.ContainSubstring("memcache_tls_keyfile"))
	g.Expect(config).NotTo(gomega.ContainSubstring("memcache_tls_cafile"))
}

// TestAodhConfKeystoneAuthtokenMemcachedMTLS verifies that when memcached runs
// with TLS and MTLS is configured, [keystone_authtoken] uses the TLS server
// list, enables memcache TLS and emits the MTLS cert files.
func TestAodhConfKeystoneAuthtokenMemcachedMTLS(t *testing.T) {
	g := gomega.NewWithT(t)

	params := baseAodhParams()
	params["MemcachedTLS"] = true
	params["MemcachedAuthCert"] = "/etc/pki/tls/certs/mtls.crt"
	params["MemcachedAuthKey"] = "/etc/pki/tls/private/mtls.key"
	params["MemcachedAuthCa"] = "/etc/pki/tls/certs/mtls-ca.crt"

	config := renderAodhConf(t, params)

	g.Expect(config).To(gomega.ContainSubstring(
		"memcached_servers=" + params["MemcachedServers"].(string)))
	g.Expect(config).NotTo(gomega.ContainSubstring(params["MemcachedServersWithInet"].(string)))
	g.Expect(config).To(gomega.ContainSubstring("memcache_tls_enabled = true"))
	g.Expect(config).To(gomega.ContainSubstring("memcache_tls_certfile = /etc/pki/tls/certs/mtls.crt"))
	g.Expect(config).To(gomega.ContainSubstring("memcache_tls_keyfile = /etc/pki/tls/private/mtls.key"))
	g.Expect(config).To(gomega.ContainSubstring("memcache_tls_cafile = /etc/pki/tls/certs/mtls-ca.crt"))
}

// TestAodhConfKeystoneAuthtokenMemcachedTLSNoMTLS verifies that when memcached
// runs with TLS but no MTLS client cert is configured, memcache TLS is still
// enabled (gated on base TLS) while no cert files are emitted.
func TestAodhConfKeystoneAuthtokenMemcachedTLSNoMTLS(t *testing.T) {
	g := gomega.NewWithT(t)

	params := baseAodhParams()
	params["MemcachedTLS"] = true

	config := renderAodhConf(t, params)

	g.Expect(config).To(gomega.ContainSubstring(
		"memcached_servers=" + params["MemcachedServers"].(string)))
	g.Expect(config).To(gomega.ContainSubstring("memcache_tls_enabled = true"))
	g.Expect(config).NotTo(gomega.ContainSubstring("memcache_tls_certfile"))
	g.Expect(config).NotTo(gomega.ContainSubstring("memcache_tls_keyfile"))
	g.Expect(config).NotTo(gomega.ContainSubstring("memcache_tls_cafile"))
}
