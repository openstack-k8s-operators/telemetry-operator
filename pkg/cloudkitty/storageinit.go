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
package cloudkitty

import (
	"github.com/openstack-k8s-operators/lib-common/modules/common/env"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// storageInitCommand -
	// TODO: Once we work on update/upgrades revisit the command in the
	//       the cloudkitty-storageinit-config.json file.
	//       If we stop all services during the update/upgrade then we can keep
	//       the --bump-versions flag.
	//       If we are doing rolling upgrades we'll need to use the flag
	//       conditionally (only for adoption) and do the restart cycle of
	//       services as described in the upstream rolling upgrades process.
	storageInitCommand = "/usr/local/bin/kolla_set_configs && /usr/local/bin/kolla_start"
)

// StorageInitJob func
func StorageInitJob(instance *telemetryv1.CloudKitty, labels map[string]string) *batchv1.Job {
	args := []string{"-c", storageInitCommand}

	// create Volume and VolumeMounts
	volumes := GetVolumes("cloudkitty")
	volumeMounts := GetVolumeMounts("cloudkitty-storageinit")
	// add CA cert if defined
	if instance.Spec.CloudKittyAPI.TLS.CaBundleSecretName != "" {
		volumes = append(volumes, instance.Spec.CloudKittyAPI.TLS.CreateVolume())
		volumeMounts = append(volumeMounts, instance.Spec.CloudKittyAPI.TLS.CreateVolumeMounts(nil)...)
	}

	runAsUser := int64(0)
	envVars := map[string]env.Setter{}
	envVars["KOLLA_CONFIG_STRATEGY"] = env.SetValue("COPY_ALWAYS")
	envVars["KOLLA_BOOTSTRAP"] = env.SetValue("TRUE")
	cloudKittyPassword := []corev1.EnvVar{
		{
			Name: "CloudKittyPassword",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: instance.Spec.Secret,
					},
					Key: instance.Spec.PasswordSelectors.CloudKittyService,
				},
			},
		},
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName + "-storageinit",
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
							Name: ServiceName + "-storageinit",
							Command: []string{
								"/bin/bash",
							},
							Args:  args,
							Image: instance.Spec.CloudKittyAPI.ContainerImage,
							SecurityContext: &corev1.SecurityContext{
								RunAsUser: &runAsUser,
							},
							Env:          env.MergeEnvs(cloudKittyPassword, envVars),
							VolumeMounts: volumeMounts,
						},
					},
					Volumes: volumes,
				},
			},
		},
	}

	if instance.Spec.NodeSelector != nil {
		job.Spec.Template.Spec.NodeSelector = *instance.Spec.NodeSelector
	}

	return job
}
