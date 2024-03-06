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
) (*appsv1.StatefulSet, error) {
	runAsUser := int64(0)

	livenessProbe := &corev1.Probe{
		// TODO might need tuning
		TimeoutSeconds:      5,
		PeriodSeconds:       3,
		InitialDelaySeconds: 3,
	}
	readinessProbe := &corev1.Probe{
		// TODO might need tuning
		TimeoutSeconds:      5,
		PeriodSeconds:       5,
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
	volumes := getVolumes(ServiceName)
	volumeMounts := getVolumeMounts("aodh-api")

	// add CA cert if defined
	if instance.Spec.Aodh.TLS.CaBundleSecretName != "" {
		volumes = append(volumes, instance.Spec.Aodh.TLS.CreateVolume())
		volumeMounts = append(volumeMounts, instance.Spec.Aodh.TLS.CreateVolumeMounts(nil)...)
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
			volumeMounts = append(volumeMounts, svc.CreateVolumeMounts(endpt.String())...)
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
		VolumeMounts: volumeMounts,
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
		VolumeMounts: getVolumeMounts("aodh-evaluator"),
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
		VolumeMounts: getVolumeMounts("aodh-notifier"),
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
		VolumeMounts: getVolumeMounts("aodh-listener"),
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

	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
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
