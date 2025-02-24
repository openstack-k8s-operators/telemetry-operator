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

package autoscaling

import (
	"fmt"

	"github.com/openstack-k8s-operators/lib-common/modules/common/annotations"
	"github.com/openstack-k8s-operators/lib-common/modules/common/env"
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	topologyv1 "github.com/openstack-k8s-operators/infra-operator/apis/topology/v1beta1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
)

const (
	// ServiceCommand -
	ServiceCommand = "/usr/local/bin/kolla_set_configs && /usr/local/bin/kolla_start"
)

// AodhStatefulSet func
func AodhStatefulSet(
	instance *telemetryv1.Autoscaling,
	configHash string,
	labels map[string]string,
	topology *topologyv1.Topology,
) (*appsv1.StatefulSet, error) {
	runAsUser := int64(0)

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
	volumes := getVolumes()
	apiVolumeMounts := getVolumeMounts("aodh-api")
	evaluatorVolumeMounts := getVolumeMounts("aodh-evaluator")
	notifierVolumeMounts := getVolumeMounts("aodh-notifier")
	listenerVolumeMounts := getVolumeMounts("aodh-listener")

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
		volumes = append(volumes, getCustomPrometheusCaVolume(instance.Spec.PrometheusTLSCaCertSecret.LocalObjectReference.Name))
		evaluatorVolumeMounts = append(evaluatorVolumeMounts, getCustomPrometheusCaVolumeMount(instance.Spec.PrometheusTLSCaCertSecret.Key))
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
			volumes = append(volumes, svc.CreateVolume(endpt.String()))
			apiVolumeMounts = append(apiVolumeMounts, svc.CreateVolumeMounts(endpt.String())...)
		}
	}

	envVarsAodh := map[string]env.Setter{}
	envVarsAodh["KOLLA_CONFIG_STRATEGY"] = env.SetValue("COPY_ALWAYS")
	envVarsAodh["CONFIG_HASH"] = env.SetValue(configHash)

	var replicas int32 = 1

	apiContainer := corev1.Container{
		ImagePullPolicy: corev1.PullAlways,
		Command: []string{
			"/bin/bash",
		},
		Args:  args,
		Image: instance.Spec.Aodh.APIImage,
		Name:  "aodh-api",
		Env:   env.MergeEnvs([]corev1.EnvVar{}, envVarsAodh),
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &runAsUser,
		},
		VolumeMounts: apiVolumeMounts,
	}

	evaluatorContainer := corev1.Container{
		ImagePullPolicy: corev1.PullAlways,
		Command: []string{
			"/bin/bash",
		},
		Args:  args,
		Image: instance.Spec.Aodh.EvaluatorImage,
		Name:  "aodh-evaluator",
		Env:   env.MergeEnvs([]corev1.EnvVar{}, envVarsAodh),
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &runAsUser,
		},
		VolumeMounts: evaluatorVolumeMounts,
	}

	notifierContainer := corev1.Container{
		ImagePullPolicy: corev1.PullAlways,
		Command: []string{
			"/bin/bash",
		},
		Args:  args,
		Image: instance.Spec.Aodh.NotifierImage,
		Name:  "aodh-notifier",
		Env:   env.MergeEnvs([]corev1.EnvVar{}, envVarsAodh),
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &runAsUser,
		},
		VolumeMounts: notifierVolumeMounts,
	}

	listenerContainer := corev1.Container{
		ImagePullPolicy: corev1.PullAlways,
		Command: []string{
			"/bin/bash",
		},
		Args:  args,
		Image: instance.Spec.Aodh.ListenerImage,
		Name:  "aodh-listener",
		Env:   env.MergeEnvs([]corev1.EnvVar{}, envVarsAodh),
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &runAsUser,
		},
		VolumeMounts: listenerVolumeMounts,
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
				apiContainer,
				evaluatorContainer,
				notifierContainer,
				listenerContainer,
			},
		},
	}

	if instance.Spec.Aodh.NodeSelector != nil {
		pod.Spec.NodeSelector = *instance.Spec.Aodh.NodeSelector
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
	nwAnnotation, err := annotations.GetNADAnnotation(instance.Namespace, instance.Spec.Aodh.NetworkAttachmentDefinitions)
	if err != nil {
		return nil, fmt.Errorf("failed create network annotation from %s: %w",
			instance.Spec.Aodh.NetworkAttachmentDefinitions, err)
	}
	statefulset.Spec.Template.Annotations = util.MergeStringMaps(statefulset.Spec.Template.Annotations, nwAnnotation)

	return statefulset, nil
}
