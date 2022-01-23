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
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	WebConsoleHeartbeatKey = "bcs::web_console::heartbeat"
	NAMESPACE              = "web-console"

	// DefaultCols DefaultRows 1080p页面测试得来
	DefaultCols = 211
	DefaultRows = 25

	// WebsocketPingInterval ping/pong时间间隔
	WebsocketPingInterval = 10
	// CleanUserPodInterval pod清理时间间隔
	CleanUserPodInterval = 60
	// LockShift 锁偏差时间常量
	LockShift = -2

	// TickTimeout 链接自动断开时间, 30分钟
	TickTimeout = 60 * 30
	// LoginTimeout 自动登出时间
	LoginTimeout = 60 * 60 * 24
	// UserPodExpireTime 清理POD，4个小时
	UserPodExpireTime = 3600 * 4
	// UserCtxExpireTime Context 过期时间, 12个小时
	UserCtxExpireTime = 3600 * 12

	//InterNel 用户自己集群
	InterNel = "internel"
	//EXTERNAL 平台集群
	EXTERNAL = "external"
)

//GetK8sContext 调用k8s上下文关系
func (m *manager) GetK8sContext(r http.ResponseWriter, req *http.Request, ctx context.Context, username, clusterID string) (string, error) {
	// 确保 web-console 命名空间配置正确
	if err := m.ensureNamespace(ctx, NAMESPACE); err != nil {
		return "", err
	}

	// 确保 configmap 配置正确
	if err := m.ensureConfigmap(ctx, NAMESPACE, clusterID, username); err != nil {
		return "", err
	}

	// 确保 pod 配置正确
	image := m.Config.Get("webconsole", "image").String("")
	podName, err := m.ensurePod(ctx, NAMESPACE, clusterID, username, image)
	if err != nil {
		return "", err
	}

	return podName, nil
}

// ensureNamespace 确保 web-console 命名空间配置正确
func (m *manager) ensureNamespace(ctx context.Context, name string) error {
	namespace := genNamespace(name)
	if _, err := m.k8sClient.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{}); err != nil {
		// 命名空间不存在，创建命名空间
		if _, err = m.k8sClient.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{}); err != nil {
			// 创建失败
			blog.Errorf("create namespaces failed, err : %v", err)
			return err
		}
	}

	// serviceAccount 名称和 namespace 保持一致
	if err := m.ensureServiceAccountRBAC(ctx, name); err != nil {
		blog.Errorf("create ServiceAccountRbac failed, err : %v", err)
		return err
	}

	return nil
}

// ensureServiceAccountRBAC 创建serviceAccount, 绑定Role
func (m *manager) ensureServiceAccountRBAC(ctx context.Context, name string) error {
	// ensure serviceAccount
	serviceAccount := genServiceAccount(name)
	if _, err := m.k8sClient.CoreV1().ServiceAccounts(NAMESPACE).Get(ctx, serviceAccount.Name, metav1.GetOptions{}); err != nil {
		if _, err := m.k8sClient.CoreV1().ServiceAccounts(NAMESPACE).Create(ctx, serviceAccount, metav1.CreateOptions{}); err != nil {
			return err
		}
	}

	// ensure rolebind
	clusterRoleBinding := genClusterRoleBinding(name)
	if _, err := m.k8sClient.RbacV1().ClusterRoleBindings().Get(ctx, clusterRoleBinding.Name, metav1.GetOptions{}); err != nil {
		if _, err = m.k8sClient.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBinding, metav1.CreateOptions{}); err != nil {
			return err
		}
	}

	return nil
}

// ensureConfigmap: 确保 configmap 配置正确
func (m *manager) ensureConfigmap(ctx context.Context, namespace, clusterId, username string) error {
	configmapName := getConfigMapName(clusterId, username)
	if _, err := m.k8sClient.CoreV1().ConfigMaps(namespace).Get(ctx, configmapName, metav1.GetOptions{}); err == nil {
		return nil
	}

	serviceAccountToken, err := m.getServiceAccountToken(ctx, namespace)
	if err != nil {
		return err
	}

	kubeConfig := genKubeConfig(clusterId, namespace, serviceAccountToken, username)
	kubeConfigYaml, err := yaml.Marshal(kubeConfig)
	if err != nil {
		return err
	}

	configMap := genConfigMap(configmapName, string(kubeConfigYaml))

	// 不存在，创建
	if _, err = m.k8sClient.CoreV1().ConfigMaps(namespace).Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
		// 创建失败
		blog.Errorf("create configmap failed, err :%v", err)
		return err
	}

	return nil
}

// ensurePod 确保 pod 配置正确
func (m *manager) ensurePod(ctx context.Context, namespace, clusterId, username, image string) (string, error) {
	podName := getPodName(clusterId, username)
	configmapName := getConfigMapName(clusterId, username)

	// k8s 客户端
	pod, err := m.k8sClient.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err == nil {
		if pod.Status.Phase == "Running" {
			return podName, nil
		}

		if pod.Status.Phase == "Pending" {
			// 等待pod启动成功
			if err := m.waitUserPodReady(ctx, namespace, podName); err != nil {
				return "", err
			}
			// 已经正常启动
			return podName, nil
		}

		return "", errors.New("Pod not Running or Pending")
	}

	// 不存在则创建
	podManifest := genPod(podName, namespace, image, configmapName)
	if _, err := m.k8sClient.CoreV1().Pods(namespace).Create(ctx, podManifest, metav1.CreateOptions{}); err != nil {
		return "", err
	}

	// 等待pod启动成功
	if err := m.waitUserPodReady(ctx, namespace, podName); err != nil {
		return "", err
	}

	return podName, nil
}

// getServiceAccountToken 获取web-console token
func (m *manager) getServiceAccountToken(ctx context.Context, namespace string) (string, error) {
	secrets, err := m.k8sClient.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
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
func (m *manager) waitUserPodReady(ctx context.Context, namespace, name string) error {
	// 错误次数
	errorCount := 0
	// 最多等待1分钟
	waitTimeout := 60
	// 异常情况最多7次
	allowableNumberOfErrors := 7

	for {
		select {
		default:
			pod, err := m.k8sClient.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				blog.Errorf("查询pod失败，errorCount: %d", errorCount)
				// 获取不到pod信息，最多等待7秒
				// 记录查询次数，超过七次退出
				errorCount++
				if errorCount > allowableNumberOfErrors {
					return fmt.Errorf("申请pod资源失败，请稍后再试")
				}
			} else {
				if pod.Status.Phase == "Running" {
					return nil
				}
			}
			time.Sleep(time.Second)
		case <-time.After(time.Second * time.Duration(waitTimeout)):
			// 超时退出
			return fmt.Errorf("申请pod资源超时，请稍后再试")
		}
	}

}

// 获取pod名称
func getPodName(clusterID, username string) string {
	podName := fmt.Sprintf("kubectld-%s-u%s", clusterID, username)
	podName = strings.ToLower(podName)

	return podName
}

// 获取configMap名称
func getConfigMapName(clusterID, username string) string {
	cmName := fmt.Sprintf("kube-config-%s-u%s", clusterID, username)
	cmName = strings.ToLower(cmName)

	return cmName
}
