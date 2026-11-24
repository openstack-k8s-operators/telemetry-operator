# Dashboards

When `dashboardsEnabled: true` is set in the MetricStorage, the telemetry-operator delivers the RHOSO observability dashboards as [Perses](https://perses.dev/) resources, created in the MetricStorage namespace and owned by the MetricStorage:

- a `PersesDatasource` named `prometheus`, pointing to the Prometheus deployed for the MetricStorage (`<metricstorage-name>-prometheus.<namespace>.svc:9090`)
- one `PersesDashboard` per dashboard listed below
- a `PrometheusRule` with the recording rules used by the dashboards

## Requirements

- **Cluster Observability Operator (COO) 1.5 or later.** COO provides the MonitoringStack used by the MetricStorage and the `perses.dev` CRDs (`PersesDatasource`, `PersesDashboard`). If the CRDs are not available, the operator reports it in the MetricStorage `DashboardDatasourceReady` and `DashboardDefinitionReady` conditions and retries.
- **The COO Monitoring UIPlugin with Perses enabled.** It provides the Perses backend that renders the dashboards in the OpenShift Console. The telemetry-operator does not manage it, as it is cluster-wide and can be shared with other workloads, so it must be created by the administrator. Without it, the dashboards only exist as CRs in the cluster:

    ```bash
    oc apply -f - <<EOF
    apiVersion: observability.openshift.io/v1alpha1
    kind: UIPlugin
    metadata:
      name: monitoring
    spec:
      type: Monitoring
      monitoring:
        perses:
          enabled: true
    EOF
    ```

    In development environments, step 4.e of the [development setup](../README.md#development-setup-with-install_yamls) creates it.

## Dashboards and their data sources

A dashboard is only populated when the services it gets its metrics from are deployed. Individual dashboards can be disabled by name with `disabledDashboards` in the MetricStorage, for example `disabledDashboards: [openstack-rabbitmq]`.

| Name | Title | Requires |
|------|-------|----------|
| `openstack-cloud` | Openstack / Node Exporter / USE Method / Cluster | `telemetry` service on the dataplane nodesets (node_exporter) |
| `openstack-node` | Openstack / Node Exporter / USE Method / Node | `telemetry` service on the dataplane nodesets (node_exporter) |
| `openstack-network` | Openstack / Openstack Network Exporter | `telemetry` service on the dataplane nodesets (openstack-network-exporter) |
| `openstack-vm` | OpenStack / Ceilometer / VMs | Ceilometer enabled, and the `telemetry` service on the compute nodesets |
| `openstack-network-traffic` | OpenStack / VMs Network Traffic | Ceilometer enabled, and the `telemetry` service on the compute nodesets |
| `openstack-rabbitmq` | OpenStack / RabbitMQ | RabbitMQ clusters in the control plane |
| `openstack-ceilometer-ipmi` | OpenStack / Compute Power Monitoring / Ceilometer IPMI | `telemetry-power-monitoring` service on at least one dataplane nodeset. The dashboard is only created when it is deployed. |
| `openstack-lightspeed` | OpenStack / Lightspeed | OpenStack Lightspeed deployed in `openStackLightspeedNamespace` |
