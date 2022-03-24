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

package podmanager

import (
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientcmdv1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

// genNamespace 生成 namespace 配置
func genNamespace(name string) *v1.Namespace {
	namespace := &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	return namespace
}

type clusterAuth struct {
	ClusterId string
	Username  string
	Token     string
	Cluster   clientcmdv1.Cluster
}

// genKubeConfig 生成 kubeconfig 配置
func genKubeConfig(clusterId, username string, authInfo *clusterAuth) *clientcmdv1.Config {
	kubeConfig := &clientcmdv1.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: []clientcmdv1.NamedCluster{
			{
				Name:    clusterId,
				Cluster: authInfo.Cluster,
			},
		},
		Contexts: []clientcmdv1.NamedContext{
			{
				Name: clusterId,
				Context: clientcmdv1.Context{
					Cluster:   clusterId,
					Namespace: "default",
					AuthInfo:  username,
				},
			},
		},
		AuthInfos: []clientcmdv1.NamedAuthInfo{
			{
				Name:     username,
				AuthInfo: clientcmdv1.AuthInfo{Token: authInfo.Token},
			},
		},
		CurrentContext: clusterId,
	}

	return kubeConfig
}

// genConfigMap 生成 configmap 配置
func genConfigMap(name, config string) *v1.ConfigMap {
	cm := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string]string{
			"config": config,
		},
	}

	return cm
}

// genPod 生成 Pod 配置
func genPod(name, namespace, image, configmapName string) *v1.Pod {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "bcs-webconsole",
				"app.kubernetes.io/name":       "bcs-webconsole",
				"app.kubernetes.io/instance":   name,
			},
		},
		Spec: v1.PodSpec{
			ServiceAccountName: namespace,
			Containers: []v1.Container{
				{
					Name:            KubectlContainerName,
					ImagePullPolicy: "Always",
					Image:           image,
					VolumeMounts: []v1.VolumeMount{
						{Name: "kube-config",
							MountPath: "/root/.kube/config",
							SubPath:   "config",
						},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyAlways,
			Volumes: []v1.Volume{
				{
					Name: "kube-config",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: configmapName,
							},
						}},
				},
			},
		},
	}

	return pod
}

// genServiceAccount 生成 serviceAccount 配置
func genServiceAccount(name string) *v1.ServiceAccount {
	serviceAccount := &v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: name,
		},
	}
	return serviceAccount
}

// genClusterRoleBinding 生成 RoleBind 配置
func genClusterRoleBinding(name string) *rbacv1.ClusterRoleBinding {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "bcs:" + name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      name,
				Namespace: name,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
	}

	return clusterRoleBinding
}
