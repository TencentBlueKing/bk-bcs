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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	WebConsoleHeartbeatKey         = "bcs::web_console::heartbeat"
	NAMESPACE                      = "web-console"
	LabelWebConsoleCreateTimestamp = "io.tencent.web_console.create_timestamp"

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
func (m *manager) GetK8sContext(r http.ResponseWriter, req *http.Request, username, clusterID string) (string, error) {
	// namespace存在
	err := m.ensureNamespace()
	if err != nil {
		return "", err
	}

	configMapName := getConfigMapName(clusterID, username)
	podName := getPodName(clusterID, username)
	serviceAccountToken, err := m.getServiceAccountToken()
	if err != nil {

	}
	conf := types.UserPodConfig{
		ServiceAccountToken: serviceAccountToken,
		SourceClusterID:     clusterID,
		HttpsServerAddress:  "",
		Username:            clusterID,
		UserToken:           "",
		PodName:             podName,
		ConfigMapName:       configMapName,
	}

	err = m.ensureConfigmap(conf)
	if err != nil {
		return "", err
	}
	pod, err := m.ensurePod(conf)
	if err != nil {
		return "", err
	}

	return pod.GetName(), nil
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
		blog.Errorf("create namespaces failed, err : %v", err)
		return err
	}

	err = m.createServiceAccountRbac()
	if err != nil {
		blog.Errorf("create ServiceAccountRbac failed, err : %v", err)
		return err
	}

	return nil
}

// 创建configMap
func (m *manager) ensureConfigmap(conf types.UserPodConfig) error {

	configMap, err := m.k8sClient.CoreV1().ConfigMaps(NAMESPACE).Get(conf.ConfigMapName, metav1.GetOptions{})
	if err == nil {
		// 存在，直接返回
		return nil
	}
	// 不存在，创建
	configMap = m.genConfigMap(conf)
	_, err = m.k8sClient.CoreV1().ConfigMaps(NAMESPACE).Create(configMap)
	if err != nil {
		// 创建失败
		blog.Errorf("crate config failed, err :%v", err)
		return err
	}

	return nil
}

// 确保pod存在
func (m *manager) ensurePod(conf types.UserPodConfig) (*v1.Pod, error) {
	// k8s 客户端
	pod, err := m.k8sClient.CoreV1().Pods(NAMESPACE).Get(conf.PodName, metav1.GetOptions{})
	if err == nil {
		if pod.Status.Phase != "Running" {
			// pod不是Running状态，请稍后再试{}
			return nil, err
		}
		return pod, nil
	}
	// 不存在则创建
	pod = m.genPod(conf)
	_, err = m.k8sClient.CoreV1().Pods(NAMESPACE).Create(pod)
	if err != nil {
		return nil, err
	}
	// 等待pod启动成功
	err = m.waitUserPodReady(conf.PodName)
	if err != nil {
		return nil, err
	}

	return pod, nil
}

// 获取pod
func (m *manager) genPod(conf types.UserPodConfig) *v1.Pod {

	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: conf.PodName,
			Labels: map[string]string{
				LabelWebConsoleCreateTimestamp: time.Unix(time.Now().Unix(), 0).Format("20060102150405"),
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{
					Name:            conf.PodName,
					ImagePullPolicy: "Always",
					Image:           m.conf.WebConsoleImage,
					VolumeMounts: []v1.VolumeMount{
						v1.VolumeMount{Name: "kube-config",
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
								Name: conf.ConfigMapName,
							},
						}},
				},
			},
		},
	}

	if len(conf.ServiceAccountToken) > 0 {
		pod.Spec.ServiceAccountName = NAMESPACE
	}

	return pod

}

// 获取configMap
func (m *manager) genConfigMap(conf types.UserPodConfig) *v1.ConfigMap {

	cmData := m.genConfigMapData(conf)

	cm := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: conf.ConfigMapName,
		},
		Data: map[string]string{
			"config": cmData,
		},
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

// 清理用户下的相关集群pod
func (m *manager) cleanUserPodByCluster() {

}

// 等待pod启动成功
func (m *manager) waitUserPodReady(podName string) error {
	// 错误次数
	errorCount := 0
	// 最多等待1分钟
	waitTimeout := 60
	// 异常情况最多7次
	allowableNumberOfErrors := 7

	for {
		select {
		default:
			pod, err := m.k8sClient.CoreV1().Pods(NAMESPACE).Get(podName, metav1.GetOptions{})
			if err != nil {
				fmt.Println("查询pod失败，errorCount:", errorCount)
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

// 获取web-console token
func (m *manager) getServiceAccountToken() (string, error) {
	secrets, err := m.k8sClient.CoreV1().Secrets("default").List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	for _, item := range secrets.Items {
		if _, ok := item.Data["token"]; ok {
			return string(item.Data["token"]), nil
		}
	}

	return "", fmt.Errorf("not found ServiceAccountToken")
}

// 创建serviceAccount, 绑定Role
func (m *manager) createServiceAccountRbac() error {
	serviceAccount := genServiceAccount()
	_, err := m.k8sClient.CoreV1().ServiceAccounts(NAMESPACE).Create(serviceAccount)
	if err != nil {
		return err
	}
	clusterRoleBinding := genServiceAccountRoleBind()
	_, err = m.k8sClient.RbacV1().ClusterRoleBindings().Create(clusterRoleBinding)
	if err != nil {
		return err
	}
	return nil
}

// 获取 ServiceAccountRoleBind
func genServiceAccountRoleBind() *rbacv1.ClusterRoleBinding {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "bcs:" + NAMESPACE,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      NAMESPACE,
				Namespace: NAMESPACE,
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

// 获取ServiceAccount
func genServiceAccount() *v1.ServiceAccount {
	serviceAccount := &v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      NAMESPACE,
			Namespace: NAMESPACE,
		},
	}
	return serviceAccount
}

// 获取configMapData
func (m *manager) genConfigMapData(conf types.UserPodConfig) string {

	clusters := make([]types.PodCmClusters, 1)
	if len(conf.ServiceAccountToken) > 0 {
		clusters = []types.PodCmClusters{
			{
				Name: conf.SourceClusterID,
				Cluster: types.PodCmCluster{
					CertificateAuthority: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
					Server:               "https://kubernetes.default.svc",
				},
			},
		}
	} else {
		clusters = []types.PodCmClusters{
			{
				Name: conf.SourceClusterID,
				Cluster: types.PodCmCluster{
					InsecureSkipTlsVerify: true,
					Server:                conf.HttpsServerAddress,
				},
			},
		}
	}

	contexts := []types.PodCmContexts{
		{
			Name: conf.SourceClusterID,
			Context: types.PodCmContext{
				Cluster:   conf.SourceClusterID,
				User:      conf.Username,
				Namespace: "default",
			},
		},
	}
	users := make([]types.PodCmUsers, 1)
	if len(conf.ServiceAccountToken) > 0 {
		users = []types.PodCmUsers{
			{
				Name: conf.Username,
				User: types.PodCmUser{
					Token: conf.ServiceAccountToken,
				},
			},
		}
	} else {
		users = []types.PodCmUsers{
			{
				Name: conf.Username,
				User: types.PodCmUser{
					Token: conf.UserToken,
				},
			},
		}
	}

	data := types.PodCmData{
		ApiVersion:     "v1",
		CurrentContext: conf.SourceClusterID,
		Kind:           "Config",
		Clusters:       clusters,
		Contexts:       contexts,
		Users:          users,
	}

	dataYaml, _ := yaml.Marshal(data)

	return string(dataYaml)
}
