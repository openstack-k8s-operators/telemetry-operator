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
	"github.com/openstack-k8s-operators/lib-common/modules/common/env"
	autoscalingv1beta1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	dbSyncCommand = "/usr/local/bin/kolla_set_configs && /usr/local/bin/kolla_start"
)

// DbSyncJob func
func DbSyncJob(instance *autoscalingv1beta1.Autoscaling, labels map[string]string) *batchv1.Job {
	args := []string{"-c"}
	args = append(args, dbSyncCommand)

	// create Volume and VolumeMounts
	volumes := getVolumes(ServiceName)
	volumeMounts := getVolumeMounts("aodh-dbsync")
	// add CA cert if defined
	if instance.Spec.Aodh.TLS.CaBundleSecretName != "" {
		volumes = append(volumes, instance.Spec.Aodh.TLS.CreateVolume())
		volumeMounts = append(volumeMounts, instance.Spec.Aodh.TLS.CreateVolumeMounts(nil)...)
	}

	runAsUser := int64(0)
	envVars := map[string]env.Setter{}
	envVars["KOLLA_CONFIG_STRATEGY"] = env.SetValue("COPY_ALWAYS")
	envVars["KOLLA_BOOTSTRAP"] = env.SetValue("TRUE")
	aodhPassword := []corev1.EnvVar{
		{
			Name: "AodhPassword",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: instance.Spec.Aodh.Secret,
					},
					Key: instance.Spec.Aodh.PasswordSelectors.Service,
				},
			},
		},
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName + "-db-sync",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					ServiceAccountName: instance.RbacResourceName(),
					Containers: []corev1.Container{
						{
							Name: ServiceName + "-db-sync",
							Command: []string{
								"/bin/bash",
							},
							Args:  args,
							Image: instance.Spec.Aodh.APIImage,
							SecurityContext: &corev1.SecurityContext{
								RunAsUser: &runAsUser,
							},
							Env:          env.MergeEnvs(aodhPassword, envVars),
							VolumeMounts: volumeMounts,
						},
					},
					Volumes: volumes,
				},
			},
		},
	}

	return job
}
