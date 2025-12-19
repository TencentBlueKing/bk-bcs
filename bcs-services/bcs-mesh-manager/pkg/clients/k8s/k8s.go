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

// Package k8s 提供k8s client功能
package k8s

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	aggregatorclientset "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
)

var (
	bcsToken    string
	bcsEndpoint string
)

// InitClient 初始化k8s client
func InitClient(endpoint, token string) error {
	bcsEndpoint = endpoint
	bcsToken = token
	return nil
}

// GetConfig 获取k8s config
func GetConfig(clusterID string) (*rest.Config, error) {
	host := fmt.Sprintf("%s/clusters/%s", bcsEndpoint, clusterID)
	return &rest.Config{
		Host:        host,
		BearerToken: bcsToken,
	}, nil
}

// GetClient 获取k8s client
func GetClient(clusterID string) (*kubernetes.Clientset, error) {
	host := fmt.Sprintf("%s/clusters/%s", bcsEndpoint, clusterID)
	config := &rest.Config{
		Host:        host,
		BearerToken: bcsToken,
	}

	return kubernetes.NewForConfig(config)
}

// CheckNamespaceExist 检查namespace是否存在
func CheckNamespaceExist(ctx context.Context, clusterID, namespace string) (bool, error) {
	_, err := GetNamespace(ctx, clusterID, namespace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetNamespace 获取namespace
func GetNamespace(ctx context.Context, clusterID, namespace string) (*corev1.Namespace, error) {
	client, err := GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	return client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
}

// CreateNamespace 创建namespace
func CreateNamespace(ctx context.Context, clusterID, namespace string) error {
	client, err := GetClient(clusterID)
	if err != nil {
		return err
	}

	_, err = client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}, metav1.CreateOptions{})
	return err
}

// GetClusterVersion 获取集群版本
func GetClusterVersion(ctx context.Context, clusterID string) (string, error) {
	client, err := GetClient(clusterID)
	if err != nil {
		return "", err
	}

	version, err := client.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}

	return version.String(), nil
}

// CheckIstioInstalled 检查k8s集群是否安装了istio
func CheckIstioInstalled(ctx context.Context, clusterID string) (bool, error) {
	// 检查apiservices
	config := &rest.Config{
		Host:        fmt.Sprintf("%s/clusters/%s", bcsEndpoint, clusterID),
		BearerToken: bcsToken,
	}
	aggregatorClient, err := aggregatorclientset.NewForConfig(config)
	if err != nil {
		return false, err
	}
	apiservices, err := aggregatorClient.ApiregistrationV1().APIServices().List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	// 包含"istio.io"
	for _, apiservice := range apiservices.Items {
		if strings.Contains(apiservice.Name, "istio.io") {
			return true, nil
		}
	}
	return false, nil
}

// GetDynamicClient returns a dynamic client for the given cluster
func GetDynamicClient(clusterID string) (dynamic.Interface, error) {
	host := fmt.Sprintf("%s/clusters/%s", bcsEndpoint, clusterID)
	config := &rest.Config{
		Host:        host,
		BearerToken: bcsToken,
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("create dynamic client failed: %v", err)
	}

	return client, nil
}

// GetDiscoveryClient returns a discovery client for the given cluster
func GetDiscoveryClient(clusterID string) (*discovery.DiscoveryClient, error) {
	host := fmt.Sprintf("%s/clusters/%s", bcsEndpoint, clusterID)
	config := &rest.Config{
		Host:        host,
		BearerToken: bcsToken,
	}

	client, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("create discovery client failed: %v", err)
	}

	return client, nil
}

// AnnotateNamespace 为指定集群的命名空间添加注解
func AnnotateNamespace(ctx context.Context, clusterID, namespace string, annotations map[string]string) error {
	client, err := GetClient(clusterID)
	if err != nil {
		return fmt.Errorf("failed to get k8s client for cluster %s: %v", clusterID, err)
	}

	ns, err := client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get namespace %s in cluster %s: %v", namespace, clusterID, err)
	}

	if ns.Annotations == nil {
		ns.Annotations = make(map[string]string)
	}

	// 添加新的注解
	for key, value := range annotations {
		ns.Annotations[key] = value
	}

	// 更新 namespace
	_, err = client.CoreV1().Namespaces().Update(ctx, ns, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update namespace %s annotations in cluster %s: %v", namespace, clusterID, err)
	}

	return nil
}

// GetService 获取指定集群中的service
func GetService(ctx context.Context, clusterID, namespace, serviceName string) (*corev1.Service, error) {
	client, err := GetClient(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get k8s client for cluster %s: %v", clusterID, err)
	}

	service, err := client.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service %s in namespace %s cluster %s: %v",
			serviceName, namespace, clusterID, err)
	}

	return service, nil
}

const (
	// ServiceTypeLoadBalancer service type loadbalancer
	ServiceTypeLoadBalancer = "LoadBalancer"
)

// GetCLBIP 获取东西向网关的CLB IP
func GetCLBIP(ctx context.Context, clusterID, serviceName string) (string, error) {
	// 获取eastwestgateway的service
	service, err := GetService(ctx, clusterID, common.IstioNamespace, serviceName)
	if err != nil {
		return "", fmt.Errorf("get eastwestgateway service failed: %s", err)
	}

	// 检查service类型是否为LoadBalancer
	if service.Spec.Type != ServiceTypeLoadBalancer {
		return "", fmt.Errorf("eastwestgateway service type is not LoadBalancer: %s", service.Spec.Type)
	}

	// 从LoadBalancer Ingress中获取IP地址
	if len(service.Status.LoadBalancer.Ingress) == 0 {
		return "", fmt.Errorf("eastwestgateway service has no LoadBalancer ingress")
	}

	ingress := service.Status.LoadBalancer.Ingress[0]
	if ingress.IP != "" {
		return ingress.IP, nil
	}

	return "", fmt.Errorf("eastwestgateway service LoadBalancer ingress has no IP")
}

// GetBCSEndpoint 获取BCS endpoint
func GetBCSEndpoint() string {
	return bcsEndpoint
}

// GetBCSToken 获取BCS token
func GetBCSToken() string {
	return bcsToken
}

// CreateIstioNamespace 创建 istio-system 命名空间,如果已经存在则忽略
func CreateIstioNamespace(ctx context.Context, clusterID string) error {
	exist, err := CheckNamespaceExist(ctx, clusterID, common.IstioNamespace)
	if err != nil {
		return fmt.Errorf("check namespace exist failed: %s", err)
	}
	// 不存在则创建
	if !exist {
		if createErr := CreateNamespace(ctx, clusterID, common.IstioNamespace); createErr != nil {
			return fmt.Errorf("create namespace failed: %s", createErr)
		}
	}
	return nil
}

// CreateRemoteClusterSecret 为远程集群创建Secret
func CreateRemoteClusterSecret(
	ctx context.Context,
	primaryClusterID,
	remoteClusterID string,
) error {
	// 使用common包生成Secret YAML
	bcsEndpoint := GetBCSEndpoint()
	bcsToken := GetBCSToken()
	secretYAML := common.GetRemoteClusterSecretYAML(remoteClusterID, bcsEndpoint, bcsToken)

	// 通过YAML部署Secret资源，注意secret名称需要使用小写
	lowerRemoteClusterID := strings.ToLower(remoteClusterID)
	secretName := fmt.Sprintf("%s%s", common.RemoteClusterSecretNamePrefix, lowerRemoteClusterID)
	if err := DeployResourceByYAML(ctx, primaryClusterID, secretYAML, common.SecretKind, secretName); err != nil {
		return fmt.Errorf("deploy remote cluster secret failed: %s", err)
	}

	return nil
}

// DeleteRemoteClusterSecret 删除远程集群Secret
func DeleteRemoteClusterSecret(
	ctx context.Context,
	primaryClusterID,
	remoteClusterID string,
) error {
	lowerRemoteClusterID := strings.ToLower(remoteClusterID)
	secretName := fmt.Sprintf("%s%s", common.RemoteClusterSecretNamePrefix, lowerRemoteClusterID)
	if err := DeleteResource(ctx, primaryClusterID, common.SecretKind, secretName); err != nil {
		return fmt.Errorf("delete remote cluster secret failed: %s", err)
	}
	return nil
}
