# AGENTS.md - telemetry-operator

## Project overview

telemetry-operator is a Kubernetes operator that manages
[OpenStack Telemetry](https://docs.openstack.org/ceilometer/latest/) services
(Ceilometer metering, Aodh alarming, CloudKitty rating, autoscaling, logging,
and monitoring) on OpenShift/Kubernetes. It is part of the
[openstack-k8s-operators](https://github.com/openstack-k8s-operators) project.

Key Telemetry domain concepts: **Ceilometer** (metering: meters, samples,
pipelines, polling agents), **Aodh** (alarming: alarms, evaluators,
notifiers), **CloudKitty** (rating/billing: modules, hashmap, processors),
**autoscaling** (Aodh + Heat integration), **logging** (centralized log
collection), **monitoring stack** (metrics storage and dashboards).

Go module: `github.com/openstack-k8s-operators/telemetry-operator`
API group: `telemetry.openstack.org`
API version: `v1beta1`

## Tech stack

| Layer | Technology |
|-------|------------|
| Language | Go (modules, multi-module workspace via `go.work`) |
| Scaffolding | [Kubebuilder v4](https://book.kubebuilder.io/) + [Operator SDK](https://sdk.operatorframework.io/) |
| CRD generation | controller-gen (DeepCopy, CRDs, RBAC, webhooks) |
| Config management | Kustomize |
| Packaging | OLM bundle |
| Testing | KUTTL (integration) |
| Linting | golangci-lint (`.golangci.yaml`) |
| CI | Zuul (`zuul.d/`), Prow (`.ci-operator.yaml`), GitHub Actions |

## Custom Resources

| Kind | Purpose |
|------|---------|
| `Telemetry` | Top-level CR. Orchestrates all telemetry sub-components (Ceilometer, Autoscaling, Logging, CloudKitty). |
| `Ceilometer` | Manages Ceilometer metering services (central agent, compute agent, IPMI agent, notification agent). |
| `Autoscaling` | Manages Aodh alarming services (API, evaluator, notifier, listener) for autoscaling. |
| `Logging` | Manages centralized logging infrastructure. |
| `CloudKitty` | Manages CloudKitty rating/billing services. |
| `CloudKittyAPI` | Manages the CloudKitty API deployment. |
| `CloudKittyProc` | Manages the CloudKitty processor deployment. |

The `Telemetry` CR has defaulting and validating admission webhooks.
Sub-CRs are created and owned by the `Telemetry` controller -- not intended to
be created directly by users.

## Directory structure

| Directory | Contents |
|-----------|----------|
| `api/v1beta1/` | CRD types (`telemetry_types.go`, `ceilometer_types.go`, `autoscaling_types.go`, `logging_types.go`, `cloudkitty_types.go`, `cloudkittyapi_types.go`, `cloudkittyproc_types.go`), conditions, webhook markers |
| `cmd/` | `main.go` entry point |
| `internal/controller/` | Reconcilers for all CR types |
| `internal/telemetry/` | Telemetry-level resource builders |
| `internal/ceilometer/` | Ceilometer resource builders |
| `internal/autoscaling/` | Autoscaling/Aodh resource builders |
| `internal/cloudkitty/` | CloudKitty-level resource builders |
| `internal/cloudkittyapi/` | CloudKittyAPI resource builders |
| `internal/cloudkittyproc/` | CloudKittyProc resource builders |
| `internal/logging/` | Logging resource builders |
| `internal/availability/` | Availability helpers |
| `internal/dashboards/` | Dashboard resource builders |
| `internal/metricstorage/` | Metric storage resource builders |
| `internal/mysqldexporter/` | MySQL exporter resource builders |
| `internal/utils/` | Shared utility functions |
| `internal/webhook/` | Webhook implementation |
| `templates/` | Config files and scripts mounted into pods via `OPERATOR_TEMPLATES` env var |
| `config/crd,rbac,manager,webhook/` | Generated Kubernetes manifests (CRDs, RBAC, deployment, webhooks) |
| `config/samples/` | Example CRs (Kustomize overlays). Includes TLS variants. |
| `test/kuttl/` | KUTTL integration tests |
| `hack/` | Helper scripts (CRD schema checker, local webhook runner) |
| `docs/` | Documentation |
| `ci/` | CI job definitions |

## Build commands

After modifying Go code, always run: `make generate manifests fmt vet`.

## Code style guidelines

- Follow standard openstack-k8s-operators conventions and lib-common patterns.
- Use `lib-common` modules for conditions, endpoints, TLS, storage, and other
  cross-cutting concerns rather than re-implementing them.
- CRD types go in `api/v1beta1/`. Controller logic goes in
  `internal/controller/`. Resource-building helpers go in `internal/<service>`
  packages matching the CR they support.
- Config templates are plain files in `templates/` -- they are mounted at
  runtime via the `OPERATOR_TEMPLATES` environment variable.
- Webhook logic is split between the kubebuilder markers in `api/v1beta1/` and
  the implementation in `internal/webhook/`.

## Testing

- KUTTL integration tests live in `test/kuttl/`.
- Run all tests: `make test`.
- When adding a new field or feature, add corresponding test cases and update
  fixture data accordingly.

## Key dependencies

- [lib-common](https://github.com/openstack-k8s-operators/lib-common): shared modules for conditions, endpoints, database, TLS, secrets, Ansible, cert-manager, etc.
- [infra-operator](https://github.com/openstack-k8s-operators/infra-operator): RabbitMQ and topology APIs.
- [mariadb-operator](https://github.com/openstack-k8s-operators/mariadb-operator): database provisioning.
- [keystone-operator](https://github.com/openstack-k8s-operators/keystone-operator): identity service registration.
- [heat-operator](https://github.com/openstack-k8s-operators/heat-operator): Heat API for autoscaling integration.
- [ovn-operator](https://github.com/openstack-k8s-operators/ovn-operator): OVN integration.
