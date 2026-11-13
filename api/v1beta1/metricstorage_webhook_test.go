/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"testing"

	. "github.com/onsi/gomega"
)

const (
	testAetosHAProxyTimeoutAnno = "haproxy.router.openshift.io/timeout"
	testAetosTimeoutAnno   = "api.aetos.openstack.org/timeout"
)

func newAetosCoreWithAPITimeout(apiTimeout int) *AetosCore {
	spec := &AetosCore{}
	spec.APITimeout = apiTimeout
	return spec
}

func TestAetosSetDefaultRouteAnnotations(t *testing.T) {
	g := NewWithT(t)

	annotations := map[string]string{}
	newAetosCoreWithAPITimeout(300).SetDefaultRouteAnnotations(annotations)

	g.Expect(annotations).To(HaveKeyWithValue(testAetosHAProxyTimeoutAnno, "300s"))
	g.Expect(annotations).To(HaveKeyWithValue(testAetosTimeoutAnno, "300s"))
}

func TestAetosSetDefaultRouteAnnotationsUpdatesDefaultedTimeout(t *testing.T) {
	g := NewWithT(t)

	annotations := map[string]string{
		testAetosHAProxyTimeoutAnno: "60s",
		testAetosTimeoutAnno:   "60s",
	}
	newAetosCoreWithAPITimeout(120).SetDefaultRouteAnnotations(annotations)

	g.Expect(annotations).To(HaveKeyWithValue(testAetosHAProxyTimeoutAnno, "120s"))
	g.Expect(annotations).To(HaveKeyWithValue(testAetosTimeoutAnno, "120s"))
}

func TestAetosSetDefaultRouteAnnotationsKeepsManualTimeout(t *testing.T) {
	g := NewWithT(t)

	annotations := map[string]string{
		testAetosHAProxyTimeoutAnno: "600s",
	}
	newAetosCoreWithAPITimeout(60).SetDefaultRouteAnnotations(annotations)

	g.Expect(annotations).To(HaveKeyWithValue(testAetosHAProxyTimeoutAnno, "600s"))
	g.Expect(annotations).NotTo(HaveKey(testAetosTimeoutAnno))
}

func TestAetosSetDefaultRouteAnnotationsDropsFlagOnManualChange(t *testing.T) {
	g := NewWithT(t)

	annotations := map[string]string{
		testAetosHAProxyTimeoutAnno: "600s",
		testAetosTimeoutAnno:   "60s",
	}
	newAetosCoreWithAPITimeout(60).SetDefaultRouteAnnotations(annotations)

	g.Expect(annotations).To(HaveKeyWithValue(testAetosHAProxyTimeoutAnno, "600s"))
	g.Expect(annotations).NotTo(HaveKey(testAetosTimeoutAnno))
}

func TestAetosSetDefaultRouteAnnotationsNoAPITimeout(t *testing.T) {
	g := NewWithT(t)

	annotations := map[string]string{}
	newAetosCoreWithAPITimeout(0).SetDefaultRouteAnnotations(annotations)

	g.Expect(annotations).To(BeEmpty())
}
