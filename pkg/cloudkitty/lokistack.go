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

	lokistackv1 "github.com/grafana/loki/operator/api/loki/v1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LokiStack defines a lokistack for cloudkitty
func LokiStack(
	instance *telemetryv1.CloudKitty,
	labels map[string]string,
) *lokistackv1.LokiStack {
	lokiStack := &lokistackv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-lokistack", instance.Name),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: lokistackv1.LokiStackSpec{
			// TODO: What size do we even want? I assume something
			//       smallish since only rating interact with this
			Size:             lokistackv1.LokiStackSizeType("1x.demo"),
			Storage:          instance.Spec.S3StorageConfig,
			StorageClassName: instance.Spec.StorageClass,
			Tenants: &lokistackv1.TenantsSpec{
				Mode: lokistackv1.Static,
				Authentication: []lokistackv1.AuthenticationSpec{
					{
						TenantName: instance.Name,
						TenantID:   instance.Name,
						MTLS: &lokistackv1.MTLSSpec{
							CA: &lokistackv1.CASpec{
								CAKey: CaConfigmapKey,
								CA:    CaConfigmapName,
							},
						},
					},
				},
				Authorization: &lokistackv1.AuthorizationSpec{
					// TODO: Determine what exactly this does and what's needed here
					Roles: []lokistackv1.RoleSpec{
						{
							Name: "cloudkitty-logs",
							Resources: []string{
								"logs",
							},
							Tenants: []string{
								"cloudkitty",
							},
							Permissions: []lokistackv1.PermissionType{
								lokistackv1.Write,
								lokistackv1.Read,
							},
						},
						{
							Name: "cluster-reader",
							Resources: []string{
								"logs",
							},
							Tenants: []string{
								"cloudkitty",
							},
							Permissions: []lokistackv1.PermissionType{
								lokistackv1.Read,
							},
						},
					},
					RoleBindings: []lokistackv1.RoleBindingsSpec{
						{
							Name: "cloudkitty-logs",
							Subjects: []lokistackv1.Subject{
								{
									Name: "cloudkitty",
									Kind: lokistackv1.Group,
								},
							},
							Roles: []string{
								"cloudkitty-logs",
							},
						},
						{
							Name: "cluster-reader",
							Subjects: []lokistackv1.Subject{
								{
									Name: "cloudkitty-logs-admin",
									Kind: lokistackv1.Group,
								},
							},
							Roles: []string{
								"cluster-reader",
							},
						},
					},
				},
			},
		},
	}
	return lokiStack
}
