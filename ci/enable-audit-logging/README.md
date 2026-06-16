enable-audit-logging
=========

A role for enabling CADF audit logging on OpenStack services deployed by openstack-k8s-operators. The role creates the necessary audit configuration secrets and applies kustomizations to the OpenStackControlPlane CR.

Currently supported services:
- Barbican
- Cinder
- Glance
- Keystone
- Neutron
- Nova

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
