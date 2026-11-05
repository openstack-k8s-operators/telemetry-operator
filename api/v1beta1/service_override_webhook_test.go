package v1beta1

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
)

func routedOverride(url string, portName string, port int32) service.RoutedOverrideSpec {
	o := service.RoutedOverrideSpec{}
	if url != "" {
		o.EndpointURL = ptr.To(url)
	}
	if portName != "" {
		o.Spec = &service.OverrideServiceSpec{
			Ports: []service.OverrideServicePort{{Name: portName, Port: port}},
		}
	}
	return o
}

func TestServiceOverrideValidation(t *testing.T) {
	cases := []struct {
		name     string
		endpoint service.Endpoint
		override service.RoutedOverrideSpec
		wantErr  string // empty: accepted
	}{
		{"public bare name", service.EndpointPublic, routedOverride("", "public", 8042), ""},
		{"internal bare name", service.EndpointInternal, routedOverride("", "internal", 8042), ""},
		{"public wrong name", service.EndpointPublic, routedOverride("", "aodh-public", 8042), `must be "public"`},
		{"internal wrong name", service.EndpointInternal, routedOverride("", "aodh-internal", 8042), `must be "internal"`},
		{"public FQDN url", service.EndpointPublic, routedOverride("https://aodh.example.com:8042", "", 0), ""},
		{"internal FQDN url", service.EndpointInternal, routedOverride("https://aodh-int.example.com", "", 0), ""},
		{"public short name url", service.EndpointPublic, routedOverride("https://aodh", "", 0), "at least two segments"},
		{"public no scheme", service.EndpointPublic, routedOverride("aodh.example.com", "", 0), "endpointURL"},
		{"public IP url", service.EndpointPublic, routedOverride("https://192.168.1.10", "", 0), "endpointURL"},
		{"port mismatch", service.EndpointPublic, routedOverride("https://aodh.example.com:8042", "public", 9999), "endpointURL"},
		{"port match", service.EndpointPublic, routedOverride("https://aodh.example.com:8042", "public", 8042), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			ov := map[service.Endpoint]service.RoutedOverrideSpec{tc.endpoint: tc.override}
			base := field.NewPath("spec")

			aodh := &AodhCore{}
			aodh.Override.Service = ov
			_, aodhCreate := aodh.ValidateCreate(base, "ns")
			_, aodhUpdate := aodh.ValidateUpdate(*aodh, base, "ns")

			ck := &CloudKittySpecCore{}
			ck.CloudKittyAPI.Override.Service = ov
			_, ckCreate := ck.ValidateCreate(base, "ns")
			_, ckUpdate := ck.ValidateUpdate(*ck, base, "ns")

			for _, errs := range []field.ErrorList{aodhCreate, aodhUpdate, ckCreate, ckUpdate} {
				// unrelated required-field errors (e.g. cloudkitty s3 config) are out of scope
				var overrideErrs field.ErrorList
				for _, e := range errs {
					if strings.Contains(e.Field, "override") {
						overrideErrs = append(overrideErrs, e)
					}
				}
				errs = overrideErrs
				if tc.wantErr == "" {
					g.Expect(errs).To(BeEmpty())
				} else {
					g.Expect(errs.ToAggregate()).To(MatchError(ContainSubstring(tc.wantErr)))
				}
			}
		})
	}
}

// The Telemetry-level webhook must reach the aodh and cloudkitty override checks.
func TestTelemetryServiceOverrideValidation(t *testing.T) {
	g := NewWithT(t)
	bad := map[service.Endpoint]service.RoutedOverrideSpec{
		service.EndpointPublic: routedOverride("", "wrong", 8042),
	}
	spec := TelemetrySpecCore{}
	spec.Autoscaling.Enabled = ptr.To(true)
	spec.Autoscaling.Aodh.Override.Service = bad
	spec.CloudKitty.Enabled = ptr.To(true)
	spec.CloudKitty.CloudKittyAPI.Override.Service = bad

	_, createErrs := spec.ValidateCreate(field.NewPath("spec"), "ns")
	_, updateErrs := spec.ValidateUpdate(spec, field.NewPath("spec"), "ns")
	for _, errs := range []field.ErrorList{createErrs, updateErrs} {
		g.Expect(errs.ToAggregate()).To(MatchError(ContainSubstring("spec.autoscaling.aodh.override.service[public]")))
		g.Expect(errs.ToAggregate()).To(MatchError(ContainSubstring("spec.cloudkitty.cloudKittyAPI.override.service[public]")))
	}
}
