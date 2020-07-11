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
	policyV1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	defaultZkClientPort = 2181
)

var zkClientPort int32

// reconcileZookeeper reconciles bk-cmdb zookeeper
func (r *BkcmdbReconciler) reconcileZookeeper(instance *bkcmdbv1.Bkcmdb) error {
	sts := makeZookeeperSts(instance)
	if err := controllerutil.SetControllerReference(instance, sts, r.Scheme); err != nil {
		return fmt.Errorf("failed to set zookeeper statefulset owner reference: %s", err.Error())
	}
	err := r.Client.CreateOrUpdateSts(sts)
	if err != nil {
		return fmt.Errorf("failed to create or update zookeeper statefulset: %s", err.Error())
	}

	pdb := makeZookeeperPdb(instance)
	if err := controllerutil.SetControllerReference(instance, pdb, r.Scheme); err != nil {
		return fmt.Errorf("failed to set zookeeper PodDisruptionBudget owner reference: %s", err.Error())
	}
	err = r.Client.CreateOrUpdatePdb(pdb)
	if err != nil {
		return fmt.Errorf("failed to create or update zookeeper PodDisruptionBudget: %s", err.Error())
	}

	svc := makeZkService(instance)
	if err := controllerutil.SetControllerReference(instance, svc, r.Scheme); err != nil {
		return fmt.Errorf("failed to set zookeeper service owner reference: %s", err.Error())
	}
	err = r.Client.CreateOrUpdateService(svc)
	if err != nil {
		return fmt.Errorf("failed to create or update zookeeper service: %s", err.Error())
	}

	headlessSvc := makeZkHeadlessService(instance)
	if err := controllerutil.SetControllerReference(instance, headlessSvc, r.Scheme); err != nil {
		return fmt.Errorf("failed to set zookeeper service owner reference: %s", err.Error())
	}
	err = r.Client.CreateOrUpdateService(headlessSvc)
	if err != nil {
		return fmt.Errorf("failed to create or update zookeeper service: %s", err.Error())
	}

	return nil
}

// makeZookeeperSts builds zookeeper statefulset object
func makeZookeeperSts(z *bkcmdbv1.Bkcmdb) *appsv1.StatefulSet {
	replicas := int32(1)
	terminationPeriod := int64(1800)
	fsGroup := int64(1000)
	runAsUser := int64(1000)

	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-zookeeper",
			Namespace: z.Namespace,
			Labels: map[string]string{
				"app":     "zookeeper",
				"release": z.GetName(),
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":       "zookeeper",
					"release":   z.GetName(),
					"component": "server",
				},
			},
			Replicas:    &replicas,
			ServiceName: z.GetName() + "-zookeeper-headless",
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.OnDeleteStatefulSetStrategyType,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":       "zookeeper",
						"release":   z.GetName(),
						"component": "server",
					},
				},
				Spec: v1.PodSpec{
					TerminationGracePeriodSeconds: &terminationPeriod,
					SecurityContext: &v1.PodSecurityContext{
						RunAsUser: &runAsUser,
						FSGroup:   &fsGroup,
					},
					Containers: makeZookeeperContainers(z),
					Volumes: []v1.Volume{
						{
							Name: "data",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
}

// makeZookeeperContainers build zookeeper containers object
func makeZookeeperContainers(z *bkcmdbv1.Bkcmdb) []v1.Container {
	return []v1.Container{
		{
			Name:            z.GetName() + "-zookeeper",
			Image:           "gcr.io/google_samples/k8szk:v3",
			ImagePullPolicy: "IfNotPresent",
			Command:         []string{"/bin/bash", "-xec", "zkGenConfig.sh && exec zkServer.sh start-foreground"},
			Env: []v1.EnvVar{
				{
					Name:  "ZK_REPLICAS",
					Value: "1",
				},
				{
					Name:  "JMXAUTH",
					Value: "false",
				},
				{
					Name:  "JMXDISABLE",
					Value: "false",
				},
				{
					Name:  "JMXPORT",
					Value: "1099",
				},
				{
					Name:  "JMXSSL",
					Value: "false",
				},
				{
					Name:  "ZK_CLIENT_PORT",
					Value: "2181",
				},
				{
					Name:  "ZK_ELECTION_PORT",
					Value: "3888",
				},
				{
					Name:  "ZK_HEAP_SIZE",
					Value: "2G",
				},
				{
					Name:  "ZK_INIT_LIMIT",
					Value: "5",
				},
				{
					Name:  "ZK_LOG_LEVEL",
					Value: "INFO",
				},
				{
					Name:  "ZK_MAX_CLIENT_CNXNS",
					Value: "60",
				},
				{
					Name:  "ZK_MAX_SESSION_TIMEOUT",
					Value: "40000",
				},
				{
					Name:  "ZK_MIN_SESSION_TIMEOUT",
					Value: "4000",
				},
				{
					Name:  "ZK_PURGE_INTERVAL",
					Value: "0",
				},
				{
					Name:  "ZK_SERVER_PORT",
					Value: "2888",
				},
				{
					Name:  "ZK_SNAP_RETAIN_COUNT",
					Value: "3",
				},
				{
					Name:  "ZK_SYNC_LIMIT",
					Value: "10",
				},
				{
					Name:  "ZK_TICK_TIME",
					Value: "2000",
				},
			},
			Ports: []v1.ContainerPort{
				{
					Name:          "client",
					ContainerPort: 2181,
					Protocol:      v1.ProtocolTCP,
				},
				{
					Name:          "election",
					ContainerPort: 3888,
					Protocol:      v1.ProtocolTCP,
				},
				{
					Name:          "server",
					ContainerPort: 2888,
					Protocol:      v1.ProtocolTCP,
				},
			},
			LivenessProbe: &v1.Probe{
				Handler: v1.Handler{
					Exec: &v1.ExecAction{
						Command: []string{"zkOk.sh"},
					},
				},
				InitialDelaySeconds: 20,
			},
			ReadinessProbe: &v1.Probe{
				Handler: v1.Handler{
					Exec: &v1.ExecAction{
						Command: []string{"zkOk.sh"},
					},
				},
				InitialDelaySeconds: 20,
			},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      "data",
					MountPath: "/var/lib/zookeeper",
				},
			},
		},
	}
}

// makeZookeeperPdb builds zookeeper PodDisruptionBudget object
func makeZookeeperPdb(z *bkcmdbv1.Bkcmdb) *policyV1beta1.PodDisruptionBudget {
	return &policyV1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-zookeeper",
			Namespace: z.Namespace,
			Labels: map[string]string{
				"app":      "zookeeper",
				"release":  z.GetName(),
				"componet": "server",
			},
		},
		Spec: policyV1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":       "zookeeper",
					"release":   z.GetName(),
					"component": "server",
				},
			},
			MaxUnavailable: &intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 1,
			},
		},
	}
}

// makeZkService builds zookeeper service object
func makeZkService(z *bkcmdbv1.Bkcmdb) *v1.Service {
	zkClientPort = defaultZkClientPort
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-zookeeper",
			Namespace: z.Namespace,
			Labels: map[string]string{
				"app":     "zookeeper",
				"release": z.GetName(),
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{
				{
					Name:     "client",
					Port:     zkClientPort,
					Protocol: v1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "client",
					},
				},
			},
			Selector: map[string]string{
				"app":     "zookeeper",
				"release": z.GetName(),
			},
		},
	}
}

// makeZkHeadlessService builds zookeeper headless service object
func makeZkHeadlessService(z *bkcmdbv1.Bkcmdb) *v1.Service {
	zkClientPort = defaultZkClientPort
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName() + "-zookeeper-headless",
			Namespace: z.Namespace,
			Labels: map[string]string{
				"app":     "zookeeper",
				"release": z.GetName(),
			},
		},
		Spec: v1.ServiceSpec{
			ClusterIP: v1.ClusterIPNone,
			Ports: []v1.ServicePort{
				{
					Name:     "client",
					Port:     zkClientPort,
					Protocol: v1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "client",
					},
				},
				{
					Name:     "election",
					Port:     3888,
					Protocol: v1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "election",
					},
				},
				{
					Name:     "server",
					Port:     2888,
					Protocol: v1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "server",
					},
				},
			},
			Selector: map[string]string{
				"app":     "zookeeper",
				"release": z.GetName(),
			},
		},
	}
}
