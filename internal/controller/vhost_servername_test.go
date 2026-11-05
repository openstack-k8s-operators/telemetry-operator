package controller

import (
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/onsi/gomega"
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"k8s.io/utils/ptr"
)

// renderVhosts mirrors the controllers' vhost params loop (ServerName and
// ServerAliases from the endpoint override) and renders the given template.
func renderVhosts(t *testing.T, tmplPath, svcName string, overrides map[service.Endpoint]service.RoutedOverrideSpec) string {
	t.Helper()
	vhosts := map[string]any{}
	for _, endpt := range []service.Endpoint{service.EndpointInternal, service.EndpointPublic} {
		name, aliases, err := service.GetVHostServerName(overrides[endpt], svcName+"-"+endpt.String()+".ns.svc")
		if err != nil {
			t.Fatal(err)
		}
		vhosts[endpt.String()] = map[string]any{"ServerName": name, "ServerAliases": aliases, "TLS": false}
	}
	content, err := os.ReadFile(tmplPath) //nolint:gosec // test-only, path is a fixed template in the repo
	if err != nil {
		t.Fatal(err)
	}
	tmpl, err := template.New("vhost").Parse(string(content))
	if err != nil {
		t.Fatal(err)
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, map[string]any{"VHosts": vhosts, "Timeout": 60}); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestVhostServerNameAlias(t *testing.T) {
	for _, tc := range []struct{ tmpl, svc string }{
		{"../../templates/autoscaling/config/wsgi-aodh.conf", "aodh"},
		{"../../templates/cloudkitty/config/wsgi-cloudkitty.conf", "cloudkitty"},
	} {
		t.Run(tc.svc, func(t *testing.T) {
			g := gomega.NewWithT(t)
			conf := renderVhosts(t, tc.tmpl, tc.svc, map[service.Endpoint]service.RoutedOverrideSpec{
				service.EndpointPublic:   {EndpointURL: ptr.To("https://" + tc.svc + ".example.com:13000")},
				service.EndpointInternal: {EndpointURL: ptr.To("https://" + tc.svc + "-int.example.com")},
			})
			g.Expect(conf).To(gomega.ContainSubstring("ServerName " + tc.svc + ".example.com\n"))
			g.Expect(conf).To(gomega.ContainSubstring("ServerAlias " + tc.svc + "-public.ns.svc\n"))
			g.Expect(conf).To(gomega.ContainSubstring("ServerName " + tc.svc + "-int.example.com\n"))
			g.Expect(conf).To(gomega.ContainSubstring("ServerAlias " + tc.svc + "-internal.ns.svc\n"))

			// without an endpointURL the in-cluster name stays the ServerName
			conf = renderVhosts(t, tc.tmpl, tc.svc, nil)
			g.Expect(conf).To(gomega.ContainSubstring("ServerName " + tc.svc + "-public.ns.svc\n"))
			g.Expect(conf).To(gomega.ContainSubstring("ServerName " + tc.svc + "-internal.ns.svc\n"))
		})
	}
}
