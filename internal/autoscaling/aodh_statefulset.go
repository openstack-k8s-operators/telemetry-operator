/*
Copyright 2022.

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

// Package autoscaling provides functionality for managing OpenStack telemetry autoscaling components
package autoscaling

import (
	"fmt"

	"github.com/openstack-k8s-operators/lib-common/modules/common/annotations"
	"github.com/openstack-k8s-operators/lib-common/modules/common/env"
	"github.com/openstack-k8s-operators/lib-common/modules/common/pod"
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	"github.com/openstack-k8s-operators/lib-common/modules/users"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	memcachedv1 "github.com/openstack-k8s-operators/infra-operator/apis/memcached/v1beta1"
	topologyv1 "github.com/openstack-k8s-operators/infra-operator/apis/topology/v1beta1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
)

// AodhStatefulSet func
func AodhStatefulSet(
	instance *telemetryv1.Autoscaling,
	configHash string,
	labels map[string]string,
	topology *topologyv1.Topology,
	memcached *memcachedv1.Memcached,
	customConfigKeys []string,
) (*appsv1.StatefulSet, error) {

	// TODO might need tuning
	livenessProbe := &corev1.Probe{
		TimeoutSeconds:      30,
		PeriodSeconds:       30,
		InitialDelaySeconds: 5,
	}
	readinessProbe := &corev1.Probe{
		TimeoutSeconds:      30,
		PeriodSeconds:       30,
		InitialDelaySeconds: 5,
	}

	livenessProbe.HTTPGet = &corev1.HTTPGetAction{
		Path: "/",
		Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(AodhAPIPort)},
	}
	readinessProbe.HTTPGet = &corev1.HTTPGetAction{
		Path: "/",
		Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(AodhAPIPort)},
	}

	if instance.Spec.Aodh.TLS.API.Enabled(service.EndpointPublic) {
		livenessProbe.HTTPGet.Scheme = corev1.URISchemeHTTPS
		readinessProbe.HTTPGet.Scheme = corev1.URISchemeHTTPS
	}

	// create Volume and VolumeMounts
	volumes := getVolumes(instance)
	apiVolumeMounts := getAPIVolumeMounts(customConfigKeys)
	evaluatorVolumeMounts := getEvaluatorVolumeMounts(customConfigKeys)
	notifierVolumeMounts := getWorkerVolumeMounts(customConfigKeys)
	listenerVolumeMounts := getWorkerVolumeMounts(customConfigKeys)

	// add openstack CA cert if defined
	if instance.Spec.Aodh.TLS.CaBundleSecretName != "" {
		volumes = append(volumes, instance.Spec.Aodh.TLS.CreateVolume())
		apiVolumeMounts = append(apiVolumeMounts, instance.Spec.Aodh.TLS.CreateVolumeMounts(nil)...)
		evaluatorVolumeMounts = append(evaluatorVolumeMounts, instance.Spec.Aodh.TLS.CreateVolumeMounts(nil)...)
		notifierVolumeMounts = append(notifierVolumeMounts, instance.Spec.Aodh.TLS.CreateVolumeMounts(nil)...)
		listenerVolumeMounts = append(listenerVolumeMounts, instance.Spec.Aodh.TLS.CreateVolumeMounts(nil)...)
	}

	// add prometheus CA cert if defined
	if instance.Spec.PrometheusTLSCaCertSecret != nil {
		volumes = append(volumes, getCustomPrometheusCaVolume(instance.Spec.PrometheusTLSCaCertSecret.Name))
		evaluatorVolumeMounts = append(evaluatorVolumeMounts, getCustomPrometheusCaVolumeMount(instance.Spec.PrometheusTLSCaCertSecret.Key))
	}

	// add MTLS cert if defined
	if memcached.GetMemcachedMTLSSecret() != "" {
		volumes = append(volumes, memcached.CreateMTLSVolume())
		apiVolumeMounts = append(apiVolumeMounts, memcached.CreateMTLSVolumeMounts(nil, nil)...)
		evaluatorVolumeMounts = append(evaluatorVolumeMounts, memcached.CreateMTLSVolumeMounts(nil, nil)...)
		notifierVolumeMounts = append(notifierVolumeMounts, memcached.CreateMTLSVolumeMounts(nil, nil)...)
		listenerVolumeMounts = append(listenerVolumeMounts, memcached.CreateMTLSVolumeMounts(nil, nil)...)
	}

	for _, endpt := range []service.Endpoint{service.EndpointInternal, service.EndpointPublic} {
		if instance.Spec.Aodh.TLS.API.Enabled(endpt) {
			var tlsEndptCfg tls.GenericService
			switch endpt {
			case service.EndpointPublic:
				tlsEndptCfg = instance.Spec.Aodh.TLS.API.Public
			case service.EndpointInternal:
				tlsEndptCfg = instance.Spec.Aodh.TLS.API.Internal
			}

			svc, err := tlsEndptCfg.ToService()
			if err != nil {
				return nil, err
			}
			certMount := fmt.Sprintf("/etc/pki/tls/certs/%s.crt", endpt.String())
			keyMount := fmt.Sprintf("/etc/pki/tls/private/%s.key", endpt.String())
			svc.CertMount = &certMount
			svc.KeyMount = &keyMount
			volumes = append(volumes, svc.CreateVolume(endpt.String()))
			apiVolumeMounts = append(apiVolumeMounts, svc.CreateVolumeMounts(endpt.String())...)
		}
	}

	envVarsAodh := map[string]env.Setter{}
	envVarsAodh["CONFIG_HASH"] = env.SetValue(configHash)

	var replicas int32 = 1

	apiContainer := corev1.Container{
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/usr/sbin/httpd"},
		Args:            []string{"-DFOREGROUND"},
		Image:           instance.Spec.Aodh.APIImage,
		Name:            "aodh-api",
		SecurityContext: pod.RestrictiveSecurityContext(users.AodhUID, users.AodhGID),
		Env:             env.MergeEnvs([]corev1.EnvVar{}, envVarsAodh),
		VolumeMounts:    apiVolumeMounts,
		ReadinessProbe:  readinessProbe,
		LivenessProbe:   livenessProbe,
	}

	evaluatorContainer := corev1.Container{
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/usr/bin/aodh-evaluator"},
		Args:            []string{"--logfile", "/dev/stdout"},
		Image:           instance.Spec.Aodh.EvaluatorImage,
		Name:            "aodh-evaluator",
		SecurityContext: pod.RestrictiveSecurityContext(users.AodhUID, users.AodhGID),
		Env:             env.MergeEnvs([]corev1.EnvVar{}, envVarsAodh),
		VolumeMounts:    evaluatorVolumeMounts,
	}

	notifierContainer := corev1.Container{
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/usr/bin/aodh-notifier"},
		Args:            []string{"--logfile", "/dev/stdout"},
		Image:           instance.Spec.Aodh.NotifierImage,
		Name:            "aodh-notifier",
		SecurityContext: pod.RestrictiveSecurityContext(users.AodhUID, users.AodhGID),
		Env:             env.MergeEnvs([]corev1.EnvVar{}, envVarsAodh),
		VolumeMounts:    notifierVolumeMounts,
	}

	listenerContainer := corev1.Container{
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/usr/bin/aodh-listener"},
		Args:            []string{"--logfile", "/dev/stdout"},
		Image:           instance.Spec.Aodh.ListenerImage,
		Name:            "aodh-listener",
		SecurityContext: pod.RestrictiveSecurityContext(users.AodhUID, users.AodhGID),
		Env:             env.MergeEnvs([]corev1.EnvVar{}, envVarsAodh),
		VolumeMounts:    listenerVolumeMounts,
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
			SecurityContext:              pod.RestrictivePodSecurityContext(users.AodhUID, users.AodhGID, users.ApacheGID),
			Containers: []corev1.Container{
				apiContainer,
				evaluatorContainer,
				notifierContainer,
				listenerContainer,
			},
		},
	}

	if instance.Spec.Aodh.NodeSelector != nil {
		podSpec.Spec.NodeSelector = *instance.Spec.Aodh.NodeSelector
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
	nwAnnotation, err := annotations.GetNADAnnotation(instance.Namespace, instance.Spec.Aodh.NetworkAttachmentDefinitions)
	if err != nil {
		return nil, fmt.Errorf("failed create network annotation from %s: %w",
			instance.Spec.Aodh.NetworkAttachmentDefinitions, err)
	}
	statefulset.Spec.Template.Annotations = util.MergeStringMaps(statefulset.Spec.Template.Annotations, nwAnnotation)

	return statefulset, nil
}
