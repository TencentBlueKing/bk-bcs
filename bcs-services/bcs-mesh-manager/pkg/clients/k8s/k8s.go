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
