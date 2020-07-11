/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package controllers

import (
	"fmt"

	bkcmdbv1 "github.com/Tencent/bk-bcs/bcs-resources/bk-cmdb-operator/api/v1"

	batchV1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileJob reconciles bk-cmdb job
func (r *BkcmdbReconciler) reconcileJob(instance *bkcmdbv1.Bkcmdb) error {
	job := makeJob(instance)
	if err := controllerutil.SetControllerReference(instance, job, r.Scheme); err != nil {
		return fmt.Errorf("failed to set job owner reference: %s", err.Error())
	}

	err := r.Client.CreateOrUpdateJob(job)
	if err != nil {
		return fmt.Errorf("failed to create or update job: %s", err.Error())
	}

	return nil
}

// makeJob builds job object
func makeJob(z *bkcmdbv1.Bkcmdb) *batchV1.Job {
	backOffLimit := int32(20)
	return &batchV1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-bootstrap",
			Namespace: z.Namespace,
			Labels: map[string]string{
				"app":     "bk-cmdb",
				"release": z.GetName(),
			},
		},
		Spec: batchV1.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: z.GetName() + "-bootstrap",
					Labels: map[string]string{
						"app":     "bk-cmdb",
						"release": z.GetName(),
					},
				},
				Spec: v1.PodSpec{
					Containers:    makeJobContainers(z),
					RestartPolicy: v1.RestartPolicyOnFailure,
					Volumes: []v1.Volume{
						{
							Name: "configures",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: z.GetName() + "-configures",
									},
								},
							},
						},
					},
				},
			},
			BackoffLimit: &backOffLimit,
		},
	}
}

// makeJobContainers builds job containers object
func makeJobContainers(z *bkcmdbv1.Bkcmdb) []v1.Container {
	asSvc := z.GetName() + "-adminserver"
	asUrl := fmt.Sprintf("http://%s:80/migrate/v3/migrate/community/0", asSvc)
	contentType := "Content-Type:application/json"
	bkUser := "BK_USER:migrate"
	supplier := "HTTP_BLUEKING_SUPPLIER_ID:0"
	return []v1.Container{
		{
			Name:            "cmdb-migrate",
			Image:           z.Spec.Image,
			ImagePullPolicy: "IfNotPresent",
			Command:         []string{"curl", "-X", "POST", "-H", contentType, "-H", bkUser, "-H", supplier, asUrl},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      "configures",
					MountPath: "/etc/configures",
				},
			},
		},
	}
}
