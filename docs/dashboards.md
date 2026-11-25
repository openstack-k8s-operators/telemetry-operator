# Dashboards

When `dashboardsEnabled: true` is set in the MetricStorage, the telemetry-operator delivers the RHOSO observability dashboards as [Perses](https://perses.dev/) resources, created in the MetricStorage namespace and owned by the MetricStorage:

- a `PersesDatasource` named `prometheus`, pointing to the Prometheus deployed for the MetricStorage (`<metricstorage-name>-prometheus.<namespace>.svc:9090`)
- one `PersesDashboard` per dashboard listed below. Every panel query and variable references the `prometheus` datasource by name, so the dashboards do not depend on which datasource is the default in the Perses project
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

## Upgrading from the ConfigMap-based dashboards

Earlier versions of the operator delivered the dashboards as ConfigMaps in the
shared `openshift-config-managed` namespace, labelled
`console.openshift.io/dashboard: "true"`, which the OpenShift Console rendered
under **Observe > Dashboards**. They are now Perses resources in the
MetricStorage namespace, rendered by the Perses UI that the Monitoring UIPlugin
provides.

**Create the UIPlugin before upgrading.** The dashboards disappear from
**Observe > Dashboards** as soon as the operator is upgraded, and without the
UIPlugin nothing renders the Perses resources that replace them, so the
dashboards are unavailable until it is created.

On the first reconcile after the upgrade, the operator removes the ConfigMaps
it created in `openshift-config-managed`:

- `grafana-dashboard-openstack-cloud`
- `grafana-dashboard-openstack-node`
- `grafana-dashboard-openstack-openstack-network`
- `grafana-dashboard-openstack-vm`
- `grafana-dashboard-openstack-rabbitmq`
- `grafana-dashboard-openstack-network-traffic`
- `grafana-dashboard-openstack-ceilometer-ipmi`
- `grafana-dashboard-openstack-lightspeed`
- `grafana-dashboard-openstack-kepler`
- `<namespace>-<metricstorage-name>-datasource`

This cleanup runs on every reconcile and on MetricStorage deletion, regardless
of `dashboardsEnabled`, so it also converges on deployments that have the
dashboards turned off. Local edits to those ConfigMaps are lost; to keep a
customised dashboard, export it before upgrading.

The `DashboardPluginReady` condition is no longer reported in the MetricStorage
status, as the operator no longer manages the UIPlugin. Anything watching that
condition needs updating. `DashboardDatasourceReady` and
`DashboardDefinitionReady` are unchanged and now track the Perses resources.

## Dashboards and their data sources

A dashboard is only populated when the services it gets its metrics from are deployed. Individual dashboards can be disabled by name with `disabledDashboards` in the MetricStorage, for example `disabledDashboards: [openstack-rabbitmq]`. Only the names in the table below are recognised; anything else disables nothing and is reported in the `DashboardDefinitionReady` condition message, together with the names that can be used.

| Name | Title | Requires |
|------|-------|----------|
| `openstack-cloud` | Openstack / Node Exporter / USE Method / Cluster | `telemetry` service on the dataplane nodesets (node_exporter) |
| `openstack-node` | Openstack / Node Exporter / USE Method / Node | `telemetry` service on the dataplane nodesets (node_exporter) |
| `openstack-network` | Openstack / Openstack Network Exporter | `telemetry` service on the dataplane nodesets (openstack-network-exporter). Currently limited to the OVN/OVS panels, see the note below. |
| `openstack-vm` | OpenStack / Ceilometer / VMs | Ceilometer enabled, and the `telemetry` service on the compute nodesets |
| `openstack-network-traffic` | OpenStack / VMs Network Traffic | Ceilometer enabled, and the `telemetry` service on the compute nodesets |
| `openstack-rabbitmq` | OpenStack / RabbitMQ | RabbitMQ clusters in the control plane |
| `openstack-ceilometer-ipmi` | OpenStack / Compute Power Monitoring / Ceilometer IPMI | `telemetry-power-monitoring` service on at least one dataplane nodeset. The dashboard is only created when it is deployed. |
| `openstack-lightspeed` | OpenStack / Lightspeed | OpenStack Lightspeed deployed in `openStackLightspeedNamespace` |

### Known limitation: DPDK and SR-IOV panels

The `openstack-network` dashboard currently ships only the **OVN/OVS** row
(Flow Count of Integration Bridge, Logical Flow Translations, Netlink and
Recirculation Rate). The **PMD** row (uS per Iteration, Upcalls, Non-Voluntary
Context Switches, Utilization, RX/TX Packets), the **DPDK** row (RX/TX Packets)
and the **SR-IOV** row (VF Configuration, VF Traffic, VF Packets, VF Dropped,
VF Broadcast / Multicast) of the ConfigMap-based dashboard have not been
converted to Perses yet and are not available.

Deployments that rely on those panels should export the previous
`grafana-dashboard-openstack-openstack-network` ConfigMap before upgrading, as
described above. The panels used to be rendered only when the
`configure-ovs-dpdk` or `neutron-sriov` service was enabled on at least one
nodeset; that conditional rendering will come back with them.

### Dropped panel: RabbitMQ TCP sockets available

The **Nodes** row of the `openstack-rabbitmq` dashboard no longer carries the
**TCP sockets available** panel. It was dropped on purpose: the panel queried
`rabbitmq_process_max_tcp_sockets` and `rabbitmq_process_open_tcp_sockets`,
which the RabbitMQ Prometheus plugin does not expose, so it rendered nothing in
the ConfigMap-based dashboard either. The **File descriptors available**,
**Memory available before publishers blocked** and **Disk space available
before publishers blocked** panels of that row are unaffected and carried over
as they were.
