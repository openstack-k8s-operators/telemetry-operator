enable-audit-logging
=========

A role for enabling CADF audit logging on OpenStack services deployed by openstack-k8s-operators. The role creates the necessary audit configuration secrets and applies kustomizations to the OpenStackControlPlane CR.

Currently supported services:
- Cinder
- Glance
- Keystone
- Neutron

Services not yet supported (config files included for future use):
- Barbican (blocked by [OSPRH-25781](https://issues.redhat.com/browse/OSPRH-25781))
- Nova (blocked by a nova-operator issue with audit map file naming)

Audit map files
----------------
The `*_api_audit_map.conf` files in `files/` are copies of the upstream pycadf defaults from https://github.com/openstack/pycadf/tree/master/etc/pycadf. If the upstream files change, the local copies should be updated to match.

Requirements
------------
- An OpenStack deployed with the openstack-k8s-operators
- `oc` CLI configured with cluster access

Side effects
------------
After the role runs, the following secrets are created in the `openstack` namespace:
- `cinder-audit-config-secret`
- `glance-audit-config-secret`
- `neutron-audit-config-secret`

The OpenStackControlPlane CR is patched (via kustomization) to mount these secrets and enable the audit middleware notification driver for each service.

Role Variables
--------------
- enable\_audit\_logging\_local\_apply: When `true`, the role fetches the live OSCP CR and applies the kustomization directly. When `false` (default), the kustomization is copied to the ci-framework kustomizations directory for pre-deploy use.

Example Playbook
----------------

    ansible-playbook ci/enable-audit-logging.yml -e enable_audit_logging_local_apply=true

License
-------

Apache 2
