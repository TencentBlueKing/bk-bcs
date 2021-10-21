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
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileAdminServer reconciles bk-cmdb adminserver
func (r *BkcmdbReconciler) reconcileAdminServer(instance *bkcmdbv1.Bkcmdb) error {
	deploy := makeAdminServerDeploy(instance)
	if err := controllerutil.SetControllerReference(instance, deploy, r.Scheme); err != nil {
		return fmt.Errorf("failed to set adminserver deployment owner reference: %s", err.Error())
	}

	err := r.Client.CreateOrUpdateDeploy(deploy)
	if err != nil {
		return fmt.Errorf("failed to create or update adminserver deploy: %s", err.Error())
	}

	service := makeAdminServerService(instance)
	if err := controllerutil.SetControllerReference(instance, service, r.Scheme); err != nil {
		return fmt.Errorf("failed to set adminserver service owner reference: %s", err.Error())
	}

	err = r.Client.CreateOrUpdateService(service)
	if err != nil {
		return fmt.Errorf("failed to create or update adminserver service: %s", err.Error())
	}

	return nil
}

// makeAdminServerDeploy builds bk-cmdb adminserver deployment object
func makeAdminServerDeploy(z *bkcmdbv1.Bkcmdb) *appsv1.Deployment {
	replicas := int32(z.Spec.AdminServer.Replicas)
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-adminserver",
			Namespace: z.Namespace,
			Labels: map[string]string{
				"app":       "bk-cmdb",
				"release":   z.GetName(),
				"component": "adminserver",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":       "bk-cmdb",
					"release":   z.GetName(),
					"component": "adminserver",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":       "bk-cmdb",
						"release":   z.GetName(),
						"component": "adminserver",
					},
				},
				Spec: v1.PodSpec{
					Containers: makeAdminServerContainers(z),
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
		},
	}
}

// makeAdminServerContainers builds adminserver containers object
func makeAdminServerContainers(z *bkcmdbv1.Bkcmdb) []v1.Container {
	return []v1.Container{
		{
			Name:            "adminserver",
			Image:           z.Spec.Image,
			ImagePullPolicy: "IfNotPresent",
			WorkingDir:      "/data/bin/bk-cmdb/cmdb_adminserver/",
			Command:         []string{"./cmdb_adminserver"},
			Args:            []string{"--addrport=$(POD_IP):80", "--config=/etc/configures/migrate.conf", "--log-dir", "./logs", "--v", "3", "--enable-auth", "false"},
			LivenessProbe: &v1.Probe{
				Handler: v1.Handler{
					HTTPGet: &v1.HTTPGetAction{
						Path: "/healthz",
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 80,
						},
					},
				},
				InitialDelaySeconds: 30,
				PeriodSeconds:       10,
			},
			ReadinessProbe: &v1.Probe{
				Handler: v1.Handler{
					HTTPGet: &v1.HTTPGetAction{
						Path: "/healthz",
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 80,
						},
					},
				},
				InitialDelaySeconds: 30,
				PeriodSeconds:       10,
			},
			Env: []v1.EnvVar{
				{
					Name: "POD_IP",
					ValueFrom: &v1.EnvVarSource{
						FieldRef: &v1.ObjectFieldSelector{
							FieldPath: "status.podIP",
						},
					},
				},
			},
			Ports: []v1.ContainerPort{
				{
					ContainerPort: 80,
				},
			},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      "configures",
					MountPath: "/etc/configures",
				},
			},
			Resources: z.Spec.AdminServer.Resources,
		},
	}
}

// makeAdminServerService build adminserver service object
func makeAdminServerService(z *bkcmdbv1.Bkcmdb) *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-adminserver",
			Namespace: z.Namespace,
			Labels: map[string]string{
				"app":     "bk-cmdb",
				"release": z.GetName(),
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port: 80,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 80,
					},
				},
			},
			Selector: map[string]string{
				"app":       "bk-cmdb",
				"release":   z.GetName(),
				"component": "adminserver",
			},
		},
	}
}
