default-telemetry
=========

A scenario for testing telemetry pieces in openstack-k8s-operators. The scenario does the following:
- Checks the operator isn't restarting and isn't logging any error
- Enables MetricStorage and Autoscaling, while COO isn't installed
- Checks the operator is reporting the expected states and logging the correct errors and it checks that the operator isn't restarting on its own
- Afterwards COO is installed
- Then it checks that the telemetry-operator doesn't output errors, that Autoscaling and MetricStorage are ready and that the operator still didn't restart
- At the end it tries to use a customMonitoringStack and configure an external Prometheus for Autoscaling

Requirements
------------
- An openstack deployed with the openstack-k8s-operators
- Heat is enabled
- Autoscaling and MetricStorage are disabled
- Telemetry is enabled
- COO isn't installed

Side effects
------------
After the scenario ends, the environment is in a modified state compared to the beginning. COO is installed during the scenario and Autoscaling and MetricStorage are enabled.

Role Variables
--------------
- default\_telemetry\_no\_coo\_expected\_logs: A list of logs to expect in the telemetry-operator's logs when MetricStorage is enabled, but COO isn't installed
- default\_telemetry\_no\_coo\_expected\_metricstorage\_conditions: A dictionary of {<condition-type>: <message>} with all MetricStorage Error conditions when MetricStorage is enabled, but COO isn't installed
- default\_telemetry\_no\_coo\_expected\_autoscaling\_conditions: A dictionary of {<condition-type>: <message>} with all Autoscaling Error conditions when Autoscaling is enabled, but COO isn't installed
- default\_telemetry\_control\_plane\_name: The name of the openstackcontrolplane CR used.

Example Playbook
----------------

License
-------

Apache 2
