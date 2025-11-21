/*
Copyright 2025.

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

package cloudkitty

import (
	"fmt"

	certmgrv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmgrmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/openstack-k8s-operators/lib-common/modules/certmanager"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
)

// Certificate defines a client certificate for communication between cloudkitty and loki
func Certificate(
	instance *telemetryv1.CloudKitty,
	labels map[string]string,
	issuer *certmgrv1.Issuer,
) *certmgrv1.Certificate {
	cert := certmanager.Cert(
		ClientCertSecretName,
		instance.Namespace,
		labels,
		certmgrv1.CertificateSpec{
			CommonName: fmt.Sprintf("%s.%s.svc", instance.Name, instance.Namespace),
			DNSNames: []string{
				fmt.Sprintf("%s.%s.svc", instance.Name, instance.Namespace),
				fmt.Sprintf("%s.%s.svc.cluster.local", instance.Name, instance.Namespace),
			},
			SecretName: ClientCertSecretName,
			Subject: &certmgrv1.X509Subject{
				OrganizationalUnits: []string{
					instance.Name,
				},
			},
			PrivateKey: &certmgrv1.CertificatePrivateKey{
				Algorithm: "RSA",
				Size:      3072,
			},
			Usages: []certmgrv1.KeyUsage{
				certmgrv1.UsageDigitalSignature,
				certmgrv1.UsageKeyEncipherment,
				certmgrv1.UsageClientAuth,
			},
			IssuerRef: certmgrmetav1.ObjectReference{
				Name:  issuer.Name,
				Kind:  issuer.Kind,
				Group: issuer.GroupVersionKind().Group,
			},
		},
	)
	return cert
}
