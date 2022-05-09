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
	"net/url"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"github.com/pkg/errors"
)

// queryByContainerId 通过cluster_id, containerId 直连容器
func queryByContainerId(ctx context.Context, clusterId, username, containerId string) (*types.PodContext, error) {
	startupMgr, err := NewStartupManager(ctx, clusterId)
	if err != nil {
		return nil, err
	}

	container, err := startupMgr.GetContainerById(containerId)
	if err != nil {
		return nil, err
	}

	podCtx := &types.PodContext{
		Mode:           types.ContainerDirectMode,
		AdminClusterId: clusterId,
		ClusterId:      clusterId,
		Namespace:      container.Namespace,
		PodName:        container.PodName,
		ContainerName:  container.ContainerName,
		Commands:       manager.DefaultCommand,
	}
	return podCtx, nil
}

// queryContainerName 通过cluster_id, namespace, podName, containerName 直连容器
func queryByContainerName(ctx context.Context, clusterId, username, namespace, podName, containerName string) (*types.PodContext, error) {
	startupMgr, err := NewStartupManager(ctx, clusterId)
	if err != nil {
		return nil, err
	}

	container, err := startupMgr.GetContainerByName(namespace, podName, containerName)
	if err != nil {
		return nil, err
	}

	podCtx := &types.PodContext{
		Mode:           types.ContainerDirectMode,
		AdminClusterId: clusterId,
		ClusterId:      clusterId,
		Namespace:      container.Namespace,
		PodName:        container.PodName,
		ContainerName:  container.ContainerName,
		Commands:       manager.DefaultCommand,
	}
	return podCtx, nil
}

// queryByClusterIdExternal 通过clusterId, 使用外部集群的方式访问 kubectl 容器
func queryByClusterIdExternal(ctx context.Context, clusterId, username, targetClusterId string) (*types.PodContext, error) {
	startupMgr, err := NewStartupManager(ctx, clusterId)
	if err != nil {
		return nil, err
	}

	namespace := GetNamespace()
	if err := startupMgr.ensureNamespace(namespace); err != nil {
		return nil, err
	}

	kubeConfig, err := startupMgr.getExternalKubeConfig(targetClusterId, username)
	if err != nil {
		return nil, err
	}

	// kubeconfig cm 配置
	configmapName := getConfigMapName(targetClusterId, username)
	startupMgr.ensureConfigmap(namespace, configmapName, kubeConfig)

	imageTag, err := GetKubectldVersion(targetClusterId)
	if err != nil {
		return nil, err
	}
	image := config.G.WebConsole.KubectldImage + ":" + imageTag

	// 确保 pod 配置正确
	podName := GetPodName(targetClusterId, username)
	// 外部集群, 默认 default 即可
	serviceAccountName := "default"
	podManifest := genPod(podName, namespace, image, configmapName, serviceAccountName)

	if err := startupMgr.ensurePod(namespace, podName, podManifest); err != nil {
		return nil, err
	}

	podCtx := &types.PodContext{
		Mode:           types.ClusterExternalMode,
		AdminClusterId: clusterId,
		ClusterId:      targetClusterId,
		PodName:        podName,
		Namespace:      namespace,
		ContainerName:  KubectlContainerName,
		Commands:       []string{"/bin/bash"}, // 进入 kubectld pod， 固定使用bash
	}
	return podCtx, nil
}

// queryByClusterIdInternal 通过clusterId, 使用inCluster的方式访问 kubectl 容器
func queryByClusterIdInternal(ctx context.Context, clusterId, username string) (*types.PodContext, error) {
	startupMgr, err := NewStartupManager(ctx, clusterId)
	if err != nil {
		return nil, err
	}

	namespace := GetNamespace()
	if err := startupMgr.ensureNamespace(namespace); err != nil {
		return nil, err
	}

	// kubeconfig cm 配置
	kubeConfig, err := startupMgr.getInternalKubeConfig(namespace, username)
	if err != nil {
		return nil, err
	}

	configmapName := getConfigMapName(clusterId, username)
	if err := startupMgr.ensureConfigmap(namespace, configmapName, kubeConfig); err != nil {
		return nil, err
	}

	// 确保 pod 配置正确
	imageTag, err := GetKubectldVersion(clusterId)
	if err != nil {
		return nil, err
	}
	image := config.G.WebConsole.KubectldImage + ":" + imageTag

	podName := GetPodName(clusterId, username)
	serviceAccountName := namespace
	podManifest := genPod(podName, namespace, image, configmapName, serviceAccountName)

	if err := startupMgr.ensurePod(namespace, podName, podManifest); err != nil {
		return nil, err
	}

	podCtx := &types.PodContext{
		Mode:           types.ClusterInternalMode,
		AdminClusterId: clusterId,
		ClusterId:      clusterId,
		PodName:        podName,
		Namespace:      namespace,
		ContainerName:  KubectlContainerName,
		Commands:       []string{"/bin/bash"}, // 进入 kubectld pod， 固定使用bash
	}
	return podCtx, nil
}

// ConsoleQuery 支持的请求参数
type ConsoleQuery struct {
	ContainerId   string `form:"container_id,omitempty"`
	Namespace     string `form:"namespace,omitempty"`
	PodName       string `form:"pod_name,omitempty"`
	ContainerName string `form:"container_name,omitempty"`
	Source        string `form:"source,omitempty"`
	Lang          string `form:"lang,omitempty"`
}

// MakeEncodedQuery 去掉空值后组装url
func (q *ConsoleQuery) MakeEncodedQuery() string {
	values := url.Values{}

	values.Set("container_id", q.ContainerId)
	values.Set("namespace", q.Namespace)
	values.Set("pod_name", q.PodName)
	values.Set("container_name", q.ContainerName)
	values.Set("source", q.Source)
	values.Set("lang", q.Lang)

	// 去掉空值
	for k := range values {
		if values.Get(k) == "" {
			values.Del(k)
		}
	}

	return values.Encode()
}

// IsContainerDirectMode 是否是直连容器请求
func (q *ConsoleQuery) IsContainerDirectMode() bool {
	if q.ContainerId != "" || q.Namespace != "" || q.PodName != "" || q.ContainerName != "" {
		return true
	}
	return false
}

// QueryAuthPodCtx web鉴权模式
func QueryAuthPodCtx(ctx context.Context, clusterId, username string, consoleQuery *ConsoleQuery) (*types.PodContext, error) {
	//  直连模式
	if consoleQuery.Namespace != "" && consoleQuery.PodName != "" && consoleQuery.ContainerName != "" {
		podCtx, err := queryByContainerName(
			ctx,
			clusterId,
			username,
			consoleQuery.Namespace,
			consoleQuery.PodName,
			consoleQuery.ContainerName,
		)
		return podCtx, err
	}

	// 通过容器ID直连
	if consoleQuery.ContainerId != "" {
		podCtx, err := queryByContainerId(
			ctx,
			clusterId,
			username,
			consoleQuery.ContainerId,
		)
		return podCtx, err
	}

	// 有任意参数, 使用直连模式
	if consoleQuery.IsContainerDirectMode() {
		return nil, errors.New("container_id或namespace/pod_name/container_name不能同时为空")
	}

	// 集群外 kubectl
	if config.G.WebConsole.IsExternalMode() {
		podCtx, err := queryByClusterIdExternal(
			ctx,
			config.G.WebConsole.AdminClusterId,
			username,
			clusterId)
		return podCtx, err
	}

	// 集群内 kubectl
	podCtx, err := queryByClusterIdInternal(
		ctx,
		clusterId,
		username,
	)
	return podCtx, err
}

type OpenQuery struct {
	ContainerId   string `json:"container_id"`
	Operator      string `json:"operator" binding:"required"`
	Command       string `json:"command"`
	Namespace     string `json:"namespace"`
	PodName       string `json:"pod_name"`
	ContainerName string `json:"container_name"`
}

// QueryOpenPodCtx openapi鉴权模式
func QueryOpenPodCtx(ctx context.Context, clusterId string, consoleQuery *OpenQuery) (*types.PodContext, error) {
	//  直连模式
	if consoleQuery.Namespace != "" && consoleQuery.PodName != "" && consoleQuery.ContainerName != "" {
		podCtx, err := queryByContainerName(
			ctx,
			clusterId,
			consoleQuery.Operator,
			consoleQuery.Namespace,
			consoleQuery.PodName,
			consoleQuery.ContainerName,
		)
		return podCtx, err
	}

	// 通过容器ID直连
	if consoleQuery.ContainerId != "" {
		podCtx, err := queryByContainerId(
			ctx,
			clusterId,
			consoleQuery.Operator,
			consoleQuery.ContainerId,
		)
		return podCtx, err
	}

	return nil, errors.New("container_id或namespace/pod_name/container_name不能同时为空")
}
