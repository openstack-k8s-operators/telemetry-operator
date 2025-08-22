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
	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	topologyv1 "github.com/openstack-k8s-operators/infra-operator/apis/topology/v1beta1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
)

const (
	// ServiceCommand -
	ServiceCommand = "/usr/local/bin/kolla_set_configs && /usr/local/bin/kolla_start"
	// CentralHCScript is the path to the central health check script
	CentralHCScript = "/var/lib/openstack/bin/centralhealth.py"
	// NotificationHCScript is the path to the notification health check script
	NotificationHCScript = "/var/lib/openstack/bin/notificationhealth.py"
)

// StatefulSet func
func StatefulSet(
	instance *telemetryv1.Ceilometer,
	configHash string,
	labels map[string]string,
	topology *topologyv1.Topology,
) (*appsv1.StatefulSet, error) {
	runAsUser := int64(0)

	// container probes
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

	args := []string{"-c"}
	args = append(args, ServiceCommand)

	envVarsCentral := map[string]env.Setter{}
	envVarsCentral["KOLLA_CONFIG_STRATEGY"] = env.SetValue("COPY_ALWAYS")
	envVarsCentral["CONFIG_HASH"] = env.SetValue(configHash)

	envVarsNotification := map[string]env.Setter{}
	envVarsNotification["KOLLA_CONFIG_STRATEGY"] = env.SetValue("COPY_ALWAYS")
	envVarsNotification["CONFIG_HASH"] = env.SetValue(configHash)

	var replicas int32 = 1

	volumes := getVolumes()
	centralVolumeMounts := getVolumeMounts("ceilometer-central")
	notificationVolumeMounts := getVolumeMounts("ceilometer-notification")
	httpdVolumeMounts := getHttpdVolumeMounts()

	if instance.Spec.TLS.Enabled() {
		svc, err := instance.Spec.TLS.ToService()
		if err != nil {
			return nil, err
		}
		// httpd container is not using kolla, mount the certs to its dst
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
		ImagePullPolicy: corev1.PullAlways,
		Command: []string{
			"/bin/bash",
		},
		Args:  args,
		Image: instance.Spec.CentralImage,
		Name:  "ceilometer-central-agent",
		Env:   env.MergeEnvs([]corev1.EnvVar{}, envVarsCentral),
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &runAsUser,
		},
		VolumeMounts:  centralVolumeMounts,
		LivenessProbe: centralLivenessProbe,
	}
	notificationAgentContainer := corev1.Container{
		ImagePullPolicy: corev1.PullAlways,
		Command: []string{
			"/bin/bash",
		},
		Args:  args,
		Image: instance.Spec.NotificationImage,
		Name:  "ceilometer-notification-agent",
		Env:   env.MergeEnvs([]corev1.EnvVar{}, envVarsNotification),
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &runAsUser,
		},
		VolumeMounts:  notificationVolumeMounts,
		LivenessProbe: notificationLivenessProbe,
	}
	sgCoreContainer := corev1.Container{
		ImagePullPolicy: corev1.PullAlways,
		Image:           instance.Spec.SgCoreImage,
		Name:            "sg-core",
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &runAsUser,
		},
		VolumeMounts: getSgCoreVolumeMounts(),
	}
	proxyContainer := corev1.Container{
		ImagePullPolicy: corev1.PullAlways,
		Image:           instance.Spec.ProxyImage,
		Name:            "proxy-httpd",
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &runAsUser,
		},
		Ports: []corev1.ContainerPort{{
			ContainerPort: int32(CeilometerPrometheusPort),
			Name:          "proxy-httpd",
		}},
		VolumeMounts:   httpdVolumeMounts,
		ReadinessProbe: sgReadinessProbe,
		LivenessProbe:  sgLivenessProbe,
		Command:        []string{"/usr/sbin/httpd"},
		Args:           []string{"-DFOREGROUND"},
	}

	pod := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: instance.RbacResourceName(),
			Containers: []corev1.Container{
				centralAgentContainer,
				notificationAgentContainer,
				sgCoreContainer,
				proxyContainer,
			},
		},
	}

	if instance.Spec.NodeSelector != nil {
		pod.Spec.NodeSelector = *instance.Spec.NodeSelector
	}
	if topology != nil {
		topology.ApplyTo(&pod)
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
			Template: pod,
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
