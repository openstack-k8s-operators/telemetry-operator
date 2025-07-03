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

package mysqldexporter

import (
	"fmt"

	"github.com/openstack-k8s-operators/lib-common/modules/common"
	"github.com/openstack-k8s-operators/lib-common/modules/common/affinity"
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

// StatefulSet func
func StatefulSet(
	instance *telemetryv1.Ceilometer,
	configHash string,
	labels map[string]string,
	topology *topologyv1.Topology,
) (*appsv1.StatefulSet, error) {
	runAsUser := int64(0)

	envVars := map[string]env.Setter{}
	envVars["CONFIG_HASH"] = env.SetValue(configHash)

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

	args := []string{}
	args = append(args, fmt.Sprintf("%s=%s", configFlag, configPath))
	args = append(args, collectorArgs...)

	livenessProbe.HTTPGet = &corev1.HTTPGetAction{
		Path: "/",
		Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(MysqldExporterPort)},
	}
	readinessProbe.HTTPGet = &corev1.HTTPGetAction{
		Path: "/",
		Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(MysqldExporterPort)},
	}

	var replicas int32 = 1

	volumes := getVolumes()
	volumeMounts := getVolumeMounts()

	if instance.Spec.MysqldExporterTLS.Enabled() {
		svc, err := instance.Spec.MysqldExporterTLS.GenericService.ToService()
		if err != nil {
			return nil, err
		}
		svc.CertMount = ptr.To(fmt.Sprintf("/etc/pki/tls/certs/%s", tls.CertKey))
		svc.KeyMount = ptr.To(fmt.Sprintf("/etc/pki/tls/private/%s", tls.PrivateKey))

		livenessProbe.HTTPGet.Scheme = corev1.URISchemeHTTPS
		readinessProbe.HTTPGet.Scheme = corev1.URISchemeHTTPS

		volumes = append(volumes, svc.CreateVolume(ServiceName))
		volumeMounts = append(volumeMounts, svc.CreateVolumeMounts(ServiceName)...)
		args = append(args, fmt.Sprintf("%s=%s", webConfigFlag, webConfigPath))
	}

	// add CA cert if defined
	if instance.Spec.MysqldExporterTLS.CaBundleSecretName != "" {
		ca := instance.Spec.MysqldExporterTLS.Ca
		volumes = append(volumes, ca.CreateVolume())
		volumeMounts = append(volumeMounts, ca.CreateVolumeMounts(nil)...)
	}

	mysqldExporterContainer := corev1.Container{
		ImagePullPolicy: corev1.PullAlways,
		Args:            args,
		Image:           instance.Spec.MysqldExporterImage,
		Name:            ServiceName,
		Env:             env.MergeEnvs([]corev1.EnvVar{}, envVars),
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &runAsUser,
		},
		VolumeMounts: volumeMounts,
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
				mysqldExporterContainer,
			},
			Volumes: volumes,
		},
	}

	if topology != nil {
		topology.ApplyTo(&pod)
	} else {
		// If possible two pods of the same service should not
		// run on the same worker node. If this is not possible
		// the get still created on the same worker node.
		pod.Spec.Affinity = affinity.DistributePods(
			common.AppSelector,
			[]string{
				ServiceName,
			},
			corev1.LabelHostname,
		)
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

	// networks to attach to
	nwAnnotation, err := annotations.GetNADAnnotation(instance.Namespace, instance.Spec.NetworkAttachmentDefinitions)
	if err != nil {
		return nil, fmt.Errorf("failed create network annotation from %s: %w",
			instance.Spec.NetworkAttachmentDefinitions, err)
	}
	statefulset.Spec.Template.Annotations = util.MergeStringMaps(statefulset.Spec.Template.Annotations, nwAnnotation)

	return statefulset, nil
}
