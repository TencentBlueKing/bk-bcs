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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8sErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	clientcmdv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/yaml"
)

type StartupManager struct {
	ctx       context.Context
	clusterId string // 这里是 Pod 所在集群
	k8sClient *kubernetes.Clientset
}

func NewStartupManager(ctx context.Context, clusterId string) (*StartupManager, error) {
	mgr := &StartupManager{
		ctx:       ctx,
		clusterId: clusterId,
	}

	k8sClient, err := GetK8SClientByClusterId(clusterId)
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
	// TODO 大集群可能比较慢, 可以通过bcs的storage获取namespace优化
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
		if _, err := m.k8sClient.CoreV1().Namespaces().Create(m.ctx, namespace, metav1.CreateOptions{}); err != nil {
			// 创建失败
			logger.Errorf("create namespace %s failed, err: %s", name, err)
			return err
		}
		return nil
	}

	return err
}

// ensureConfigmap: 确保 configmap 配置正确
func (m *StartupManager) ensureConfigmap(namespace, name string, kubeConfig *clientcmdv1.Config) error {
	_, err := m.k8sClient.CoreV1().ConfigMaps(namespace).Get(m.ctx, name, metav1.GetOptions{})

	// 不存在，创建
	if k8sErr.IsNotFound(err) {
		kubeConfigYaml, err := yaml.Marshal(kubeConfig)
		if err != nil {
			return err
		}
		configMap := genConfigMap(name, string(kubeConfigYaml))

		if _, err := m.k8sClient.CoreV1().ConfigMaps(namespace).Create(m.ctx, configMap, metav1.CreateOptions{}); err != nil {
			// 创建失败
			logger.Errorf("create configmap failed, err :%s", err)
			return err
		}
		return nil
	}

	return err
}

// ensurePod 确保 pod 配置正确
func (m *StartupManager) ensurePod(namespace, name string, podManifest *v1.Pod) error {
	_, err := m.k8sClient.CoreV1().Pods(namespace).Get(m.ctx, name, metav1.GetOptions{})

	if k8sErr.IsNotFound(err) {
		start := time.Now()
		if _, createErr := m.k8sClient.CoreV1().Pods(namespace).Create(m.ctx, podManifest, metav1.CreateOptions{}); createErr != nil {
			metrics.CollectPodCreateDurations(namespace, name, metrics.ErrStatus, start)
			return createErr
		}
		metrics.CollectPodCreateDurations(namespace, name, metrics.SucStatus, start)

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
	bcsConf := GetBCSConfByClusterId(targetClusterId)
	tokenObj, err := bcs.CreateTempToken(m.ctx, bcsConf, username)
	if err != nil {
		return nil, err
	}

	authInfo := &clusterAuth{
		Token: tokenObj.Token,
		Cluster: clientcmdv1.Cluster{
			Server:                fmt.Sprintf("%s/clusters/%s", bcsConf.Host, targetClusterId),
			InsecureSkipTLSVerify: true,
		},
	}

	kubeConfig := genKubeConfig(targetClusterId, username, authInfo)

	return kubeConfig, nil
}

// getInternalKubeConfig 集群内鉴权
func (m *StartupManager) getInternalKubeConfig(namespace, username string) (*clientcmdv1.Config, error) {
	// serviceAccount 名称和 namespace 保持一致
	if err := m.ensureServiceAccountRBAC(namespace); err != nil {
		logger.Errorf("create ServiceAccountRbac failed, err : %s", err)
		return nil, err
	}

	token, err := m.getServiceAccountToken(namespace)
	if err != nil {
		return nil, err
	}

	authInfo := &clusterAuth{
		Token: token,
		Cluster: clientcmdv1.Cluster{
			Server:               "https://kubernetes.default.svc",
			CertificateAuthority: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
		},
	}

	kubeConfig := genKubeConfig(m.clusterId, username, authInfo)

	return kubeConfig, nil
}

// ensureServiceAccountRBAC 创建serviceAccount, 绑定Role
func (m *StartupManager) ensureServiceAccountRBAC(name string) error {
	// ensure serviceAccount
	serviceAccount := genServiceAccount(name)
	if _, err := m.k8sClient.CoreV1().ServiceAccounts(name).Get(m.ctx, serviceAccount.Name, metav1.GetOptions{}); err != nil {
		if !k8sErr.IsNotFound(err) {
			return err
		}

		if _, err := m.k8sClient.CoreV1().ServiceAccounts(name).Create(m.ctx, serviceAccount, metav1.CreateOptions{}); err != nil {
			return err
		}
	}

	// ensure rolebind
	clusterRoleBinding := genClusterRoleBinding(name)
	if _, err := m.k8sClient.RbacV1().ClusterRoleBindings().Get(m.ctx, clusterRoleBinding.Name, metav1.GetOptions{}); err != nil {
		if !k8sErr.IsNotFound(err) {
			return err
		}

		if _, err = m.k8sClient.RbacV1().ClusterRoleBindings().Create(m.ctx, clusterRoleBinding, metav1.CreateOptions{}); err != nil {
			return err
		}
	}

	return nil
}

// getServiceAccountToken 获取web-console token
func (m *StartupManager) getServiceAccountToken(namespace string) (string, error) {
	secrets, err := m.k8sClient.CoreV1().Secrets(namespace).List(m.ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	for _, item := range secrets.Items {
		if !strings.HasPrefix(item.Name, namespace) {
			continue
		}

		if item.Type != "kubernetes.io/service-account-token" {
			continue
		}

		if _, ok := item.Data["token"]; !ok {
			continue
		}

		return string(item.Data["token"]), nil
	}

	return "", errors.New("not found ServiceAccountToken")
}

// 等待pod启动成功
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

// GetNamespace
func GetNamespace() string {
	// 正式环境使用 web-console 命名空间
	if config.G.Base.RunEnv == config.ProdEnv {
		return Namespace
	}
	// 其他环境, 使用 web-console-dev
	return fmt.Sprintf("%s-%s", Namespace, config.G.Base.RunEnv)
}

// 获取configMap名称
func getConfigMapName(clusterID, username string) string {
	cmName := fmt.Sprintf("kube-config-%s-u%s", clusterID, username)
	cmName = strings.ToLower(cmName)

	return cmName
}

// 获取pod名称
func getPodName(clusterID, username string) string {
	podName := fmt.Sprintf("kubectld-%s-u%s", clusterID, username)
	podName = strings.ToLower(podName)

	return podName
}

// GetK8SClientByClusterId 通过集群 ID 获取 k8s client 对象
func GetK8SClientByClusterId(clusterId string) (*kubernetes.Clientset, error) {
	bcsConf := GetBCSConfByClusterId(clusterId)
	host := fmt.Sprintf("%s/clusters/%s", bcsConf.Host, clusterId)
	config := &rest.Config{
		Host:        host,
		BearerToken: bcsConf.Token,
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return k8sClient, nil
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

func hasPodReadyCondition(conditions []v1.PodCondition) bool {
	for _, condition := range conditions {
		if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

// GetEnvByClusterId 获取集群所属环境, 目前通过集群ID前缀判断
func GetEnvByClusterId(clusterId string) config.BCSClusterEnv {
	if strings.HasPrefix(clusterId, "BCS-K8S-1") {
		return config.UatCluster
	}
	if strings.HasPrefix(clusterId, "BCS-K8S-2") {
		return config.DebugCLuster
	}
	if strings.HasPrefix(clusterId, "BCS-K8S-4") {
		return config.ProdEnv
	}
	return config.ProdEnv
}

// GetBCSConfByClusterId 通过集群ID, 获取不同admin token 信息
func GetBCSConfByClusterId(clusterId string) *config.BCSConf {
	env := GetEnvByClusterId(clusterId)
	conf, ok := config.G.BCSEnvMap[env]
	if ok {
		return conf
	}
	// 默认返回bcs配置
	return config.G.BCS
}

// GetKubectldVersion 获取服务端 Kubectld 版本
func GetKubectldVersion(clusterId string) (string, error) {
	k8sClient, err := GetK8SClientByClusterId(clusterId)
	if err != nil {
		return "", err
	}

	info, err := k8sClient.ServerVersion()
	if err != nil {
		return "", err
	}

	v, err := config.G.WebConsole.MatchTag(info.GitVersion)
	return v, err
}
