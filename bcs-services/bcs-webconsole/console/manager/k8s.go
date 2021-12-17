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

package manager

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NAMESPACE                      = "web-console"
	LabelWebConsoleCreateTimestamp = "io.tencent.web_console.create_timestamp"
)

//GetK8sContext 调用k8s上下文关系
func (m *manager) GetK8sContext(r http.ResponseWriter, req *http.Request, clusterID, username string) (string, error) {
	// namespace存在
	err := m.ensureNamespace()
	if err != nil {
		return "", err
	}
	err = m.ensureConfigmap(clusterID, username)
	if err != nil {
		return "", err
	}
	pod, err := m.ensurePod(clusterID, username)
	if err != nil {
		return "", err
	}

	return pod.GetName(), nil
}

// GetActiveUserPodContainerID 获取存活节点的ContainerID
func (m *manager) GetActiveUserPodContainerID(podName string) (string, error) {
	pod, err := m.k8sClient.CoreV1().Pods(NAMESPACE).Get(podName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if pod.Status.String() != "Running" {
		return "", fmt.Errorf("the state of the pod(%s) is not running", podName)
	}
	for _, status := range pod.Status.ContainerStatuses {
		containerID := status.ContainerID
		if len(containerID) > 10 {
			return containerID[9:], nil
		}
	}
	return "", fmt.Errorf("no running pod(%s) were found", podName)

}

// 创建命名空间
func (m *manager) ensureNamespace() error {

	_, err := m.k8sClient.CoreV1().Namespaces().Get(NAMESPACE, metav1.GetOptions{})
	if err == nil {
		// 命名空间存在,直接返回
		return nil
	}

	// 命名空间不存在，创建命名空间
	namespace := m.getNamespace()
	_, err = m.k8sClient.CoreV1().Namespaces().Create(namespace)
	if err != nil {
		// 创建失败
		blog.Errorf("create namespaces failed, err :%v", err)
		return err
	}
	return nil
}

// 创建configMap
func (m *manager) ensureConfigmap(clusterID, username string) error {
	configMapName := getConfigMapName(clusterID, username)

	configMap, err := m.k8sClient.CoreV1().ConfigMaps(NAMESPACE).Get(configMapName, metav1.GetOptions{})
	if err == nil {
		// 存在，直接返回
		return nil
	}
	// 不存在，创建
	configMap = m.getConfigMap(configMapName)
	_, err = m.k8sClient.CoreV1().ConfigMaps(NAMESPACE).Create(configMap)
	if err != nil {
		// 创建失败
		blog.Errorf("crate config failed, err :%v", err)
		return err
	}

	return nil
}

// 确保pod存在
func (m *manager) ensurePod(clusterID, username string) (*v1.Pod, error) {
	podName := getPodName(clusterID, username)
	// k8s 客户端
	pod, err := m.k8sClient.CoreV1().Pods(NAMESPACE).Get(podName, metav1.GetOptions{})
	if err == nil {
		if pod.Status.Phase != "Running" {
			// pod不是Running状态，请稍后再试{}
			return nil, err
		}
		return pod, nil
	}
	// 不存在则创建
	pod = m.genPod(clusterID, username, "")
	_, err = m.k8sClient.CoreV1().Pods(NAMESPACE).Create(pod)
	if err != nil {
		return nil, err
	}
	// sink already exists errors
	if apierrors.IsAlreadyExists(err) {
		return nil, err
	}

	return pod, nil
}

// 确保configMap存在

// 获取pod
func (m *manager) genPod(clusterID, username, serviceAccountToken string) *v1.Pod {
	podName := getPodName(clusterID, username)
	configMapName := getConfigMapName(clusterID, username)

	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: NAMESPACE,
			Labels: map[string]string{
				LabelWebConsoleCreateTimestamp: time.Unix(time.Now().Unix(), 0).Format("20060102150405"),
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{
					Name:            podName,
					ImagePullPolicy: "Always",
					Image:           m.conf.WebConsoleImage,
					VolumeMounts: []v1.VolumeMount{
						v1.VolumeMount{Name: "kube-config",
							//MountPath: "C:\\Users\\lin\\.kube\\config",
							MountPath: "/root/.kube/config",
							SubPath:   "config",
						},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyAlways,
			Volumes: []v1.Volume{
				v1.Volume{
					Name: "kube-config",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: configMapName,
							},
						}},
				},
			},
		},
	}

	if len(serviceAccountToken) > 0 {
		pod.Spec.ServiceAccountName = NAMESPACE
	}

	return pod

}

// 获取configMap
func (m *manager) getConfigMap(name string) *v1.ConfigMap {
	cm := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   NAMESPACE,
			Annotations: map[string]string{},
		},
		Data:       map[string]string{},
		BinaryData: nil,
	}
	return cm
}

// 获取pod名称
func getPodName(clusterID, username string) string {
	podName := fmt.Sprintf("kubectld-%s-u%s", clusterID, username)
	podName = strings.ToLower(podName)

	return podName
}

// 获取configMap名称
func getConfigMapName(clusterID, username string) string {
	podName := fmt.Sprintf("kube-config-%s-u%s", clusterID, username)
	podName = strings.ToLower(podName)

	return podName
}

// 获取namespace
func (m *manager) getNamespace() *v1.Namespace {
	namespace := &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: NAMESPACE,
		},
	}

	return namespace
}
