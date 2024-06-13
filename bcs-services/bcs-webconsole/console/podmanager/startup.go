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
 */

package podmanager

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/gosimple/slug"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8sErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clientcmdv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/yaml"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
)

// StartupManager xxx
type StartupManager struct {
	ctx       context.Context
	clusterId string // 这里是 Pod 所在集群
	k8sClient *kubernetes.Clientset
}

// NewStartupManager xxx
func NewStartupManager(ctx context.Context, clusterId string) (*StartupManager, error) {
	mgr := &StartupManager{
		ctx:       ctx,
		clusterId: clusterId,
	}

	k8sClient, err := k8sclient.GetK8SClientByClusterId(clusterId)
	if err != nil {
		return nil, err
	}
	mgr.k8sClient = k8sClient

	return mgr, nil
}

func matchContainerById(pod *v1.Pod, containerId string) (*types.Container, error) {
	for _, container := range pod.Status.ContainerStatuses {
		if container.ContainerID != "docker://"+containerId && container.ContainerID != "containerd://"+containerId {
			continue
		}

		reason, ok := IsContainerReady(&container)
		if !ok {
			return nil, errors.Errorf("Container %s not ready, %s", container.Name, reason)
		}

		container := &types.Container{
			Namespace:     pod.Namespace,
			PodName:       pod.Name,
			ContainerName: container.Name,
		}
		return container, nil
	}

	// 不返回错误, 到上层处理
	return nil, nil
}

// GetContainerById 通过 containerID 获取pod, namespace
func (m *StartupManager) GetContainerById(containerId string) (*types.Container, error) {
	// 大集群可能比较慢, 可以通过bcs的storage获取namespace优化
	pods, err := m.k8sClient.CoreV1().Pods("").List(m.ctx, metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		container, err := matchContainerById(&pod, containerId)
		if err != nil {
			return nil, err
		}
		if container != nil {
			return container, nil
		}
	}

	return nil, errors.New("Container not found")
}

// GetContainerByName 通过 namespace, podName, containerName 校验后获取容器信息
func (m *StartupManager) GetContainerByName(namespace, podName, containerName string) (*types.Container, error) {
	pod, err := m.k8sClient.CoreV1().Pods(namespace).Get(m.ctx, podName, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	for _, container := range pod.Status.ContainerStatuses {
		if container.Name != containerName {
			continue
		}

		reason, ok := IsContainerReady(&container)
		if !ok {
			return nil, errors.Errorf("Container %s not ready, %s", containerName, reason)
		}

		container := &types.Container{
			Namespace:     pod.Namespace,
			PodName:       pod.Name,
			ContainerName: container.Name,
		}
		return container, nil
	}

	return nil, errors.Errorf("Container %s not found", containerName)
}

// ensureNamespace 确保 web-console 命名空间配置正确
func (m *StartupManager) ensureNamespace(name string) error {
	namespace := genNamespace(name)
	_, err := m.k8sClient.CoreV1().Namespaces().Get(m.ctx, name, metav1.GetOptions{})

	if k8sErr.IsNotFound(err) {
		// 命名空间不存在，创建命名空间
		if _, e := m.k8sClient.CoreV1().Namespaces().Create(m.ctx, namespace, metav1.CreateOptions{}); e != nil {
			// 创建失败
			logger.Errorf("create namespace %s failed, err: %s", name, e)
			return e
		}
		return nil
	}

	return err
}

// ensureConfigmap : 确保 configmap 配置正确
func (m *StartupManager) ensureConfigmap(namespace, name, uid string, kubeConfig *clientcmdv1.Config) error {
	kubeConfigYaml, err := yaml.Marshal(kubeConfig)
	if err != nil {
		return err
	}
	configMap := genConfigMap(name, string(kubeConfigYaml), uid)

	_, err = m.k8sClient.CoreV1().ConfigMaps(namespace).Get(m.ctx, name, metav1.GetOptions{})
	// 不存在，创建
	if k8sErr.IsNotFound(err) {
		if _, e := m.k8sClient.CoreV1().ConfigMaps(namespace).Create(m.ctx, configMap, metav1.CreateOptions{}); e != nil {
			// 创建失败
			logger.Errorf("create configmap failed, err :%s", e)
			return e
		}
		return nil
	}

	if err != nil {
		return err
	}

	// 每次刷新配置
	if _, err := m.k8sClient.CoreV1().ConfigMaps(namespace).Update(m.ctx, configMap, metav1.UpdateOptions{}); err != nil {
		logger.Errorf("update configmap failed, err :%s", err)
		return err
	}

	return nil
}

// ensurePod 确保 pod 配置正确
func (m *StartupManager) ensurePod(namespace, name string, podManifest *v1.Pod) error {
	_, err := m.k8sClient.CoreV1().Pods(namespace).Get(m.ctx, name, metav1.GetOptions{})

	if k8sErr.IsNotFound(err) {
		if _, createErr := m.k8sClient.CoreV1().Pods(namespace).Create(m.ctx, podManifest,
			metav1.CreateOptions{}); createErr != nil {
			return createErr
		}

		// 等待pod启动成功
		return m.waitPodReady(namespace, name)
	}

	if err != nil {
		return err
	}

	return m.waitPodReady(namespace, name)
}

// getExternalKubeConfig 外部集群鉴权
func (m *StartupManager) getExternalKubeConfig(targetClusterId, username string) (*clientcmdv1.Config, error) {
	tokenObj, err := bcs.CreateTempToken(m.ctx, username, targetClusterId)
	if err != nil {
		return nil, err
	}

	// 获取kubeconfig server
	server := m.getServer(targetClusterId)

	authInfo := &clusterAuth{
		Token: tokenObj.Token,
		Cluster: clientcmdv1.Cluster{
			Server:                server,
			InsecureSkipTLSVerify: config.G.BCS.InsecureSkipVerify,
		},
	}

	kubeConfig := genKubeConfig(targetClusterId, username, authInfo)

	return kubeConfig, nil
}

// waitPodReady 等待pod启动成功
func (m *StartupManager) waitPodReady(namespace, name string) error {
	// 错误次数
	errorCount := 0
	// 最多等待1分钟
	waitTimeout := 60
	// 异常情况最多7次
	allowableNumberOfErrors := 7

	// context.WithDeadline()
	interval := time.NewTicker(time.Second)
	defer interval.Stop()

	for {
		select {
		case <-interval.C:
			pod, err := m.k8sClient.CoreV1().Pods(namespace).Get(m.ctx, name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			reason, ready := IsPodReady(pod)
			if ready {
				return nil
			}

			errorCount++

			if errorCount > allowableNumberOfErrors {
				return errors.New(reason)
			}

		case <-time.After(time.Second * time.Duration(waitTimeout)):
			// 超时退出
			return fmt.Errorf("申请pod资源超时，请稍后再试")
		}
	}

}

// GetAdminClusterId xxx
func GetAdminClusterId(clusterId string) string {
	if config.G.WebConsole.IsExternalMode() {
		return config.G.WebConsole.AdminClusterId
	}
	return clusterId
}

// GetNamespace xxx
func GetNamespace() string {
	// 正式环境使用 web-console 命名空间
	if config.G.Base.RunEnv == config.ProdEnv {
		return Namespace
	}
	// 其他环境, 使用 web-console-dev
	return fmt.Sprintf("%s-%s", Namespace, config.G.Base.RunEnv)
}

// makeSlugName 规范化名称
// 符合k8s规则 https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#rfc-1035-label-names
// 返回格式{hash[8]}-{slugname[20]}
func makeSlugName(username string) string {
	h := sha1.New() // nolint
	h.Write([]byte(username))
	hashId := hex.EncodeToString(h.Sum(nil))[:8]

	// 有些命名非k8s规范, slug 处理
	slug.CustomSub = map[string]string{
		"_": "-",
	}
	slug.Lowercase = true

	username = slug.Make(username)
	if len(username) > 20 {
		username = username[:20]
	}
	return fmt.Sprintf("%s-%s", hashId, username)
}

// getUid
func getUid(clusterID, username string) string {
	username = makeSlugName(username)
	uid := fmt.Sprintf("%s-u%s", clusterID, username)
	uid = strings.ToLower(uid)

	return uid
}

// getConfigMapName 获取configMap名称
func getConfigMapName(clusterID, username string) string {
	username = makeSlugName(username)
	cmName := fmt.Sprintf("kube-config-%s-u%s", clusterID, username)
	cmName = strings.ToLower(cmName)

	return cmName
}

// GetPodName 获取pod名称
func GetPodName(clusterID, username string) string {
	username = makeSlugName(username)
	podName := fmt.Sprintf("kubectld-%s-u%s", clusterID, username)
	podName = strings.ToLower(podName)

	return podName
}

// IsPodReady returns status string calculated based on the same logic as kubectl
// Base code: https://github.com/kubernetes/dashboard/blob/master/src/app/backend/resource/pod/common.go#L40
func IsPodReady(pod *v1.Pod) (string, bool) {
	if pod.DeletionTimestamp != nil && pod.Status.Reason == "NodeLost" {
		return string(v1.PodUnknown), false
	}

	if pod.DeletionTimestamp != nil {
		return "Terminating", false
	}

	if pod.Status.Phase != v1.PodRunning {
		return string(pod.Status.Phase), false
	}

	// 检查内部容器状态
	for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
		reason, ok := IsContainerReady(&pod.Status.ContainerStatuses[i])
		if !ok {
			return reason, false
		}
	}

	return "", true
}

// IsContainerReady 容器是否 Ready 检查
func IsContainerReady(container *v1.ContainerStatus) (string, bool) {
	if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
		reason := container.State.Waiting.Reason
		if container.State.Waiting.Message != "" {
			reason = reason + ": " + container.State.Waiting.Message
		}
		return reason, false
	}

	if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
		reason := container.State.Terminated.Reason
		if container.State.Terminated.Message != "" {
			reason = reason + ": " + container.State.Terminated.Message
		}
		return reason, false
	}

	if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
		if container.State.Terminated.Signal != 0 {
			return fmt.Sprintf("Signal: %d", container.State.Terminated.Signal), false
		}
		return fmt.Sprintf("ExitCode: %d", container.State.Terminated.Signal), false
	}
	return "", true
}

// GetKubectldVersion 获取服务端 Kubectld 版本
func GetKubectldVersion(clusterId string) (string, error) {
	k8sClient, err := k8sclient.GetK8SClientByClusterId(clusterId)
	if err != nil {
		return "", err
	}

	info, err := k8sClient.ServerVersion()
	if err != nil {
		return "", err
	}

	v, err := config.G.WebConsole.MatchTag(info)
	return v, err
}

// ProjectIdContext projectId context
type ProjectIdContext string

// ProjectIdContextKey projectId context key
const ProjectIdContextKey ProjectIdContext = "projectIdContext"

// 获取kubeconfig server
func (m *StartupManager) getServer(clusterId string) string {
	server := fmt.Sprintf("%s/clusters/%s", config.G.BCS.Host, clusterId)
	projectId, ok := m.ctx.Value(ProjectIdContextKey).(string)
	// 取不出来及空的情况下返回默认地址
	if !ok || projectId == "" {
		return server
	}

	cluster, err := bcs.GetCluster(m.ctx, projectId, clusterId)
	if err != nil {
		// 报错情况下返回默认的
		return server
	}

	// 如果是共享集群则加上/projects/%s
	if cluster.IsShared {
		return fmt.Sprintf("%s/projects/%s/clusters/%s", config.G.BCS.Host, projectId, clusterId)
	}
	return server
}
