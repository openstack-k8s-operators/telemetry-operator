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
	"errors"
	"fmt"
	"slices"

	lokistackv1 "github.com/grafana/loki/operator/api/loki/v1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// ErrEffectiveDateRequired is returned when schema effectiveDate is missing
	ErrEffectiveDateRequired = errors.New("invalid CloudKitty spec. Field .spec.s3StorageConfig.schema.effectiveDate is required")
	// ErrSchemaVersionRequired is returned when schema version is missing
	ErrSchemaVersionRequired = errors.New("invalid CloudKitty spec. Field .spec.s3StorageConfig.schema.version is required")
	// ErrSecretNameRequired is returned when secret name is missing
	ErrSecretNameRequired = errors.New("invalid CloudKitty spec. Field .spec.s3StorageConfig.secret.name is required")
	// ErrSecretTypeRequired is returned when secret type is missing
	ErrSecretTypeRequired = errors.New("invalid CloudKitty spec. Field .spec.s3StorageConfig.secret.type is required")
	// ErrInvalidSecretType is returned when secret type is not valid
	ErrInvalidSecretType = errors.New("invalid CloudKitty spec. Field .spec.s3StorageConfig.secret.type needs to be one of: azure, gcs, s3, swift, alibabacloud")
	// ErrCANameRequired is returned when TLS CA name is missing
	ErrCANameRequired = errors.New("invalid CloudKitty spec. Field .spec.s3StorageConfig.tls.caName is required")
)

func validateObjectStorageSpec(spec telemetryv1.ObjectStorageSpec) error {
	for _, schema := range spec.Schemas {
		if schema.EffectiveDate == "" {
			return ErrEffectiveDateRequired
		}
		if schema.Version == "" {
			return ErrSchemaVersionRequired
		}
	}

	if spec.Secret.Name == "" {
		return ErrSecretNameRequired
	}

	if spec.Secret.Type == "" {
		return ErrSecretTypeRequired
	}
	validTypes := []string{"azure", "gcs", "s3", "swift", "alibabacloud"}
	if !slices.Contains(validTypes, spec.Secret.Type) {
		return ErrInvalidSecretType
	}

	if spec.TLS != nil && spec.TLS.CA == "" {
		return ErrCANameRequired
	}

	return nil
}

func getLokiStackObjectStorageSpec(telemetryObjectStorageSpec telemetryv1.ObjectStorageSpec) lokistackv1.ObjectStorageSpec {
	var result lokistackv1.ObjectStorageSpec

	if len(telemetryObjectStorageSpec.Schemas) == 0 {
		// NOTE: if no schema is defined, use the same as defined in loki-operator.
		result.Schemas = []lokistackv1.ObjectStorageSchema{
			{
				Version:       lokistackv1.ObjectStorageSchemaVersion("v11"),
				EffectiveDate: lokistackv1.StorageSchemaEffectiveDate("2020-10-11"),
			},
		}
	} else {
		for _, schema := range telemetryObjectStorageSpec.Schemas {
			result.Schemas = append(result.Schemas, lokistackv1.ObjectStorageSchema{
				Version:       lokistackv1.ObjectStorageSchemaVersion(schema.Version),
				EffectiveDate: lokistackv1.StorageSchemaEffectiveDate(schema.EffectiveDate),
			})
		}
	}

	result.Secret.Type = lokistackv1.ObjectStorageSecretType(telemetryObjectStorageSpec.Secret.Type)
	result.Secret.Name = telemetryObjectStorageSpec.Secret.Name
	result.Secret.CredentialMode = lokistackv1.CredentialMode(telemetryObjectStorageSpec.Secret.CredentialMode)

	if telemetryObjectStorageSpec.TLS != nil {
		result.TLS = &lokistackv1.ObjectStorageTLSSpec{
			CASpec: lokistackv1.CASpec{
				CAKey: telemetryObjectStorageSpec.TLS.CAKey,
				CA:    telemetryObjectStorageSpec.TLS.CA,
			},
		}
		if result.TLS.CAKey == "" {
			// NOTE: if no CAKey is defined, use the same as defined in loki-operator
			result.TLS.CAKey = "service-ca.crt"
		}
	}

	return result
}

// LokiStack defines a lokistack for cloudkitty
func LokiStack(
	instance *telemetryv1.CloudKitty,
	labels map[string]string,
) (*lokistackv1.LokiStack, error) {
	err := validateObjectStorageSpec(instance.Spec.S3StorageConfig)
	if err != nil {
		return nil, err
	}
	size := "1x.demo"
	if instance.Spec.LokiStackSize != "" {
		size = instance.Spec.LokiStackSize
	}

	// Build limits spec only if retention is enabled (lokiRetentionDays > 0)
	var limitsSpec *lokistackv1.LimitsSpec
	if instance.Spec.LokiRetentionDays > 0 {
		limitsSpec = &lokistackv1.LimitsSpec{
			Global: &lokistackv1.LimitsTemplateSpec{
				Retention: &lokistackv1.RetentionLimitSpec{
					Days: instance.Spec.LokiRetentionDays,
				},
			},
		}
	}

	lokiStack := &lokistackv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-lokistack", instance.Name),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: lokistackv1.LokiStackSpec{
			// TODO: What size do we even want? I assume something
			//       smallish since only rating interact with this
			Size:             lokistackv1.LokiStackSizeType(size),
			Storage:          getLokiStackObjectStorageSpec(instance.Spec.S3StorageConfig),
			StorageClassName: instance.Spec.StorageClass,
			Limits:           limitsSpec,
			Tenants: &lokistackv1.TenantsSpec{
				Mode: lokistackv1.Static,
				Authentication: []lokistackv1.AuthenticationSpec{
					{
						TenantName: instance.Name,
						TenantID:   instance.Name,
						MTLS: &lokistackv1.MTLSSpec{
							CA: &lokistackv1.CASpec{
								CAKey: CaConfigmapKey,
								CA:    fmt.Sprintf("%s-%s", instance.Name, CaConfigmapName),
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
	return lokiStack, nil
}
