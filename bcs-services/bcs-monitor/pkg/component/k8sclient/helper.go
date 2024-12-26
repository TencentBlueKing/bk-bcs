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

package k8sclient

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bcsclientset "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubebkbcs/generated/clientset/versioned"
	clusternet "github.com/clusternet/clusternet/pkg/generated/clientset/versioned"
	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	k8sVersion "k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage"
)

// GetBCSConf 返回BCS配置
func GetBCSConf() *config.BCSConf {
	// 默认返回bcs配置
	return config.G.BCS
}

// GetK8SClientByClusterId 通过集群 ID 获取 k8s client 对象
func GetK8SClientByClusterId(clusterId string) (*kubernetes.Clientset, error) {
	bcsConf := GetBCSConf()
	host := fmt.Sprintf("%s/clusters/%s", bcsConf.Host, clusterId)
	config := &rest.Config{
		Host:            host,
		BearerToken:     bcsConf.Token,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return k8sClient, nil
}

// GetDynamicClientByClusterId 通过集群 ID 获取 k8s dynamic client 对象
func GetDynamicClientByClusterId(clusterId string) (dynamic.Interface, error) {
	bcsConf := GetBCSConf()
	host := fmt.Sprintf("%s/clusters/%s", bcsConf.Host, clusterId)
	config := &rest.Config{
		Host:            host,
		BearerToken:     bcsConf.Token,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}
	k8sClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return k8sClient, nil
}

// GetClusterNetClientByClusterId 通过集群 ID 获取 clusternet client 对象
func GetClusterNetClientByClusterId(clusterId string) (*clusternet.Clientset, error) {
	bcsConf := GetBCSConf()
	host := fmt.Sprintf("%s/clusters/%s", bcsConf.Host, clusterId)
	config := &rest.Config{
		Host:            host,
		BearerToken:     bcsConf.Token,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}
	k8sClient, err := clusternet.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return k8sClient, nil
}

// GetKubebkbcsClientByClusterID 通过集群 ID 获取 kube bcs 对象
func GetKubebkbcsClientByClusterID(clusterID string) (*bcsclientset.Clientset, error) {
	bcsConf := GetBCSConf()
	host := fmt.Sprintf("%s/clusters/%s", bcsConf.Host, clusterID)
	config := &rest.Config{
		Host:            host,
		BearerToken:     bcsConf.Token,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}
	k8sClient, err := bcsclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return k8sClient, nil
}

// parseVersion 解析版本, 优先使用gitVersion, 回退到 v{Major}.{Minor}.0
func parseVersion(versionInfo *k8sVersion.Info) (*version.Version, error) {
	v, err := version.NewVersion(versionInfo.GitVersion)
	if err == nil {
		return v, nil
	}

	// 回退使用 v{Major}.{Minor}.0
	majorVersion := fmt.Sprintf("v%s.%s.0", versionInfo.Major, versionInfo.Minor)
	v, err = version.NewVersion(majorVersion)
	if err == nil {
		return v, nil
	}

	return nil, errors.Errorf("Malformed version: GitVersion %s, MajorVersion %s", versionInfo.GitVersion, majorVersion)
}

// GetK8SVersion 获取 k8s 版本
func GetK8SVersion(ctx context.Context, clusterId string) (*version.Version, error) {
	cacheKey := fmt.Sprintf("k8sclient.GetK8SVersion:%s", clusterId)
	if cacheResult, ok := storage.LocalCache.Slot.Get(cacheKey); ok {
		return cacheResult.(*version.Version), nil
	}

	k8sClient, err := GetK8SClientByClusterId(clusterId)
	if err != nil {
		return nil, err
	}

	info, err := k8sClient.ServerVersion()
	if err != nil {
		return nil, err
	}

	v, err := parseVersion(info)
	if err != nil {
		return nil, err
	}

	// 缓存2个小时
	storage.LocalCache.Slot.Set(cacheKey, v, time.Hour*2)

	return v, nil
}

// K8SLessThan 对比版本, 如果异常, 按向下兼容
func K8SLessThan(ctx context.Context, clusterId, ver string) bool {
	k8sV, err := GetK8SVersion(ctx, clusterId)
	if err != nil {
		blog.Warnf("GetK8SVersion error, %s, %s, %s", clusterId, err, err == nil)
		return false
	}

	rawV, err := version.NewVersion(ver)
	if err != nil {
		blog.Warnf("parse raw version, %s, %s, %s", clusterId, ver, err)
		return false
	}

	return k8sV.LessThan(rawV)
}
