/*
Copyright 2024.

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

package availability

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"

	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
)

const (
	tlsConfPath         = "/etc/ksm"
	tlsConfKey          = "tls_config.yaml"
	tlsBaseConfTemplate = `---
tls_server_config:`
	tlsClntConfTemplate = `
  cert_file: %s
  key_file: %s`
	tlsCaConfTemplate = `
  client_ca_file: %s
  client_auth_type: VerifyClientCertIfGiven
`
)

var (
	// TLSCertPath is the path to the TLS certificate file
	TLSCertPath = fmt.Sprintf("/etc/pki/tls/certs/%s", tls.CertKey)
	// TLSKeyPath is the path to the TLS private key file
	TLSKeyPath = fmt.Sprintf("/etc/pki/tls/private/%s", tls.PrivateKey)
)

// KSMTLSConfig creates TLS configuration for kube-state-metrics
func KSMTLSConfig(
	instance *telemetryv1.Ceilometer,
	labels map[string]string,
	clntCert bool,
) *corev1.Secret {
	content := tlsBaseConfTemplate
	if clntCert {
		content = fmt.Sprintf("%s%s", content, fmt.Sprintf(tlsClntConfTemplate, TLSCertPath, TLSKeyPath))
	}
	//if instance.Spec.KSMTLS.CaBundleSecretName != "" {
	//	content = fmt.Sprintf("%s%s", content, fmt.Sprintf(tlsCaConfTemplate, tls.DownstreamTLSCABundlePath))
	//}

	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-tls-config", KSMServiceName),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			tlsConfKey: []byte(content),
		},
	}
	return sec
}
