/*

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

package ceilometer

import (
	"fmt"

	"github.com/openstack-k8s-operators/lib-common/modules/common/annotations"
	"github.com/openstack-k8s-operators/lib-common/modules/common/env"
	"github.com/openstack-k8s-operators/lib-common/modules/common/pod"
	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	"github.com/openstack-k8s-operators/lib-common/modules/users"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	topologyv1 "github.com/openstack-k8s-operators/infra-operator/apis/topology/v1beta1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
)

// StatefulSet func
func StatefulSet(
	instance *telemetryv1.Ceilometer,
	configHash string,
	labels map[string]string,
	topology *topologyv1.Topology,
	customConfigKeys []string,
) (*appsv1.StatefulSet, error) {

	sgRootEndpointCurl := corev1.HTTPGetAction{
		Path: "/",
		Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(CeilometerPrometheusPort)},
	}
	sgLivenessProbe := &corev1.Probe{
		TimeoutSeconds:      30,
		PeriodSeconds:       30,
		InitialDelaySeconds: 300,
	}
	sgLivenessProbe.HTTPGet = &sgRootEndpointCurl

	sgReadinessProbe := &corev1.Probe{
		TimeoutSeconds:      30,
		PeriodSeconds:       30,
		InitialDelaySeconds: 10,
	}
	sgReadinessProbe.HTTPGet = &sgRootEndpointCurl

	//NOTE(mmagr): Once we will be sure (OSP19 timeframe) that we have Ceilometer
	//             running with heartbeat feature, we can make below probes run much
	//             less often (poll interval is 5 minutes currently). Right now we need
	//             to execute HC as often as possible to hit times when pollers connect
	//             to OpenStack API nodes
	centralLivenessProbe := &corev1.Probe{
		TimeoutSeconds:      5,
		PeriodSeconds:       5,
		InitialDelaySeconds: 300,
	}
	centralLivenessProbe.Exec = &corev1.ExecAction{
		Command: []string{"/usr/bin/python3", CentralHCScript},
	}

	notificationLivenessProbe := &corev1.Probe{
		TimeoutSeconds:      5,
		PeriodSeconds:       30,
		InitialDelaySeconds: 300,
	}
	notificationLivenessProbe.Exec = &corev1.ExecAction{
		Command: []string{"/usr/bin/python3", NotificationHCScript},
	}

	envVars := map[string]env.Setter{}
	envVars["CONFIG_HASH"] = env.SetValue(configHash)

	var replicas int32 = 1

	volumes := getVolumes(instance)
	centralVolumeMounts := getCentralVolumeMounts(customConfigKeys)
	notificationVolumeMounts := getNotificationVolumeMounts(customConfigKeys)
	httpdVolumeMounts := getHttpdVolumeMounts()

	centralVolumeMounts = append(centralVolumeMounts, getHealthCheckVolumeMounts()...)
	notificationVolumeMounts = append(notificationVolumeMounts, getHealthCheckVolumeMounts()...)

	// add TLS cert if defined
	if instance.Spec.TLS.Enabled() {
		svc, err := instance.Spec.TLS.ToService()
		if err != nil {
			return nil, err
		}
		svc.CertMount = ptr.To(fmt.Sprintf("/etc/pki/tls/certs/%s", tls.CertKey))
		svc.KeyMount = ptr.To(fmt.Sprintf("/etc/pki/tls/private/%s", tls.PrivateKey))

		sgLivenessProbe.HTTPGet.Scheme = corev1.URISchemeHTTPS
		sgReadinessProbe.HTTPGet.Scheme = corev1.URISchemeHTTPS

		volumes = append(volumes, svc.CreateVolume(ServiceName))
		httpdVolumeMounts = append(httpdVolumeMounts, svc.CreateVolumeMounts(ServiceName)...)
	}

	// add CA cert if defined
	if instance.Spec.TLS.CaBundleSecretName != "" {
		ca := instance.Spec.TLS.Ca
		volumes = append(volumes, ca.CreateVolume())
		httpdVolumeMounts = append(httpdVolumeMounts, ca.CreateVolumeMounts(nil)...)
		centralVolumeMounts = append(centralVolumeMounts, ca.CreateVolumeMounts(nil)...)
		notificationVolumeMounts = append(notificationVolumeMounts, ca.CreateVolumeMounts(nil)...)
	}

	centralAgentContainer := corev1.Container{
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/usr/bin/ceilometer-polling"},
		Args:            []string{"--polling-namespaces", "central", "--logfile", "/dev/stdout"},
		Image:           instance.Spec.CentralImage,
		Name:            "ceilometer-central-agent",
		SecurityContext: pod.RestrictiveSecurityContext(users.CeilometerUID, users.CeilometerGID),
		Env:             env.MergeEnvs([]corev1.EnvVar{}, envVars),
		VolumeMounts:    centralVolumeMounts,
		LivenessProbe:   centralLivenessProbe,
	}
	notificationAgentContainer := corev1.Container{
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/usr/bin/ceilometer-agent-notification"},
		Args:            []string{"--logfile", "/dev/stdout"},
		Image:           instance.Spec.NotificationImage,
		Name:            "ceilometer-notification-agent",
		SecurityContext: pod.RestrictiveSecurityContext(users.CeilometerUID, users.CeilometerGID),
		Env:             env.MergeEnvs([]corev1.EnvVar{}, envVars),
		VolumeMounts:    notificationVolumeMounts,
		LivenessProbe:   notificationLivenessProbe,
	}
	sgCoreContainer := corev1.Container{
		ImagePullPolicy: corev1.PullIfNotPresent,
		Image:           instance.Spec.SgCoreImage,
		Name:            "sg-core",
		SecurityContext: pod.RestrictiveSecurityContext(users.CeilometerUID, users.CeilometerGID),
		VolumeMounts:    getSgCoreVolumeMounts(),
	}
	proxyContainer := corev1.Container{
		ImagePullPolicy: corev1.PullIfNotPresent,
		Image:           instance.Spec.ProxyImage,
		Name:            "proxy-httpd",
		Ports: []corev1.ContainerPort{{
			ContainerPort: int32(CeilometerPrometheusPort),
			Name:          "proxy-httpd",
		}},
		SecurityContext: pod.RestrictiveSecurityContext(users.CeilometerUID, users.CeilometerGID),
		VolumeMounts:    httpdVolumeMounts,
		ReadinessProbe:  sgReadinessProbe,
		LivenessProbe:   sgLivenessProbe,
		Command:         []string{"/usr/sbin/httpd"},
		Args:            []string{"-DFOREGROUND"},
	}

	podSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName:           instance.RbacResourceName(),
			AutomountServiceAccountToken: ptr.To(false),
			SecurityContext:              pod.RestrictivePodSecurityContext(users.CeilometerUID, users.CeilometerGID, users.ApacheGID),
			Containers: []corev1.Container{
				centralAgentContainer,
				notificationAgentContainer,
				sgCoreContainer,
				proxyContainer,
			},
		},
	}

	if instance.Spec.NodeSelector != nil {
		podSpec.Spec.NodeSelector = *instance.Spec.NodeSelector
	}
	if topology != nil {
		topology.ApplyTo(&podSpec)
	}
	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Replicas:            &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: podSpec,
		},
	}

	statefulset.Spec.Template.Spec.Volumes = volumes

	// networks to attach to
	nwAnnotation, err := annotations.GetNADAnnotation(instance.Namespace, instance.Spec.NetworkAttachmentDefinitions)
	if err != nil {
		return nil, fmt.Errorf("failed create network annotation from %s: %w",
			instance.Spec.NetworkAttachmentDefinitions, err)
	}
	statefulset.Spec.Template.Annotations = util.MergeStringMaps(statefulset.Spec.Template.Annotations, nwAnnotation)

	return statefulset, nil
}
