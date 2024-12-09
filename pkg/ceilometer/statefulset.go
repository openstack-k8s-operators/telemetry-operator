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

	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
)

const (
	// ServiceCommand -
	ServiceCommand = "/usr/local/bin/kolla_set_configs && /usr/local/bin/kolla_start"
)

// StatefulSet func
func StatefulSet(
	instance *telemetryv1.Ceilometer,
	configHash string,
	labels map[string]string,
) (*appsv1.StatefulSet, error) {
	runAsUser := int64(0)

	// TO-DO Probes
	livenessProbe := &corev1.Probe{
		// TODO might need tuning
		TimeoutSeconds:      30,
		PeriodSeconds:       30,
		InitialDelaySeconds: 5,
	}
	readinessProbe := &corev1.Probe{
		// TODO might need tuning
		TimeoutSeconds:      30,
		PeriodSeconds:       30,
		InitialDelaySeconds: 5,
	}

	args := []string{"-c"}
	args = append(args, ServiceCommand)

	livenessProbe.HTTPGet = &corev1.HTTPGetAction{
		Path: "/v3",
		Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(CeilometerPrometheusPort)},
	}
	readinessProbe.HTTPGet = &corev1.HTTPGetAction{
		Path: "/v3",
		Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(CeilometerPrometheusPort)},
	}

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
		svc, err := instance.Spec.TLS.GenericService.ToService()
		if err != nil {
			return nil, err
		}
		// httpd container is not using kolla, mount the certs to its dst
		svc.CertMount = ptr.To(fmt.Sprintf("/etc/pki/tls/certs/%s", tls.CertKey))
		svc.KeyMount = ptr.To(fmt.Sprintf("/etc/pki/tls/private/%s", tls.PrivateKey))

		livenessProbe.HTTPGet.Scheme = corev1.URISchemeHTTPS
		readinessProbe.HTTPGet.Scheme = corev1.URISchemeHTTPS

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
		VolumeMounts: centralVolumeMounts,
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
		VolumeMounts: notificationVolumeMounts,
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
			ContainerPort: 3000,
			Name:          "proxy-httpd",
		}},
		VolumeMounts:   httpdVolumeMounts,
		ReadinessProbe: readinessProbe,
		LivenessProbe:  livenessProbe,
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
