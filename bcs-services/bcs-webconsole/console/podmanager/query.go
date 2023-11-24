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

// Package podmanager xxx
package podmanager

import (
	"context"
	"net/url"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/google/shlex"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
)

// queryByContainerId 通过cluster_id, containerId 直连容器
func queryByContainerId(ctx context.Context,
	clusterId, containerId, shell string) (*types.PodContext, error) {
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
		Commands:       manager.ShCommand, // 直连容器默认使用 sh
	}
	switch shell {
	case manager.ShellSH:
		podCtx.Commands = manager.ShCommand
	case manager.ShellBash:
		podCtx.Commands = manager.BashCommand
	}
	return podCtx, nil
}

// queryByContainerName 通过cluster_id, namespace, podName, containerName 直连容器
func queryByContainerName(ctx context.Context,
	clusterId, namespace, podName, containerName, shell string) (*types.PodContext, error) {
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
		Commands:       manager.ShCommand, // 直连容器默认使用 sh
	}
	switch shell {
	case manager.ShellSH:
		podCtx.Commands = manager.ShCommand
	case manager.ShellBash:
		podCtx.Commands = manager.BashCommand
	}
	return podCtx, nil
}

// queryByClusterIdExternal 通过clusterId, 使用外部集群的方式访问 kubectl 容器
func queryByClusterIdExternal(ctx context.Context,
	clusterId, username, targetClusterId, shell string) (*types.PodContext,
	error) {
	startupMgr, err := NewStartupManager(ctx, clusterId)
	if err != nil {
		return nil, err
	}

	namespace := GetNamespace()
	if e := startupMgr.ensureNamespace(namespace); e != nil {
		return nil, e
	}

	kubeConfig, err := startupMgr.getExternalKubeConfig(targetClusterId, username)
	if err != nil {
		return nil, err
	}

	// kubeconfig cm 配置
	configmapName := getConfigMapName(targetClusterId, username)
	uid := getUid(targetClusterId, username)
	if err = startupMgr.ensureConfigmap(namespace, configmapName, uid, kubeConfig); err != nil {
		return nil, err
	}

	imageTag, err := GetKubectldVersion(targetClusterId)
	if err != nil {
		return nil, err
	}
	image := config.G.WebConsole.KubectldImage + ":" + imageTag

	// 确保 pod 配置正确
	podName := GetPodName(targetClusterId, username)
	// 外部集群, 默认 default 即可
	serviceAccountName := "default"
	podManifest := genPod(podName, image, configmapName, serviceAccountName, uid)

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
		Commands:       manager.BashCommand, // 进入 kubectld pod， 默认使用 bash
	}
	switch shell {
	case manager.ShellSH:
		podCtx.Commands = manager.ShCommand
	case manager.ShellBash:
		podCtx.Commands = manager.BashCommand
	}

	// 创建 .bash_history 文件
	if err := historyMgr.createBashHistory(podCtx); err != nil {
		logger.Warnf("create bash history fail: %s", err.Error())
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
	Shell         string `form:"shell,omitempty"`
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
	values.Set("shell", q.Shell)

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
func QueryAuthPodCtx(ctx context.Context, clusterId, username string, consoleQuery *ConsoleQuery) (*types.PodContext,
	error) {
	//  直连模式
	if consoleQuery.Namespace != "" && consoleQuery.PodName != "" && consoleQuery.ContainerName != "" {
		podCtx, err := queryByContainerName(
			ctx,
			clusterId,
			consoleQuery.Namespace,
			consoleQuery.PodName,
			consoleQuery.ContainerName,
			consoleQuery.Shell,
		)
		return podCtx, err
	}

	// 通过容器ID直连
	if consoleQuery.ContainerId != "" {
		podCtx, err := queryByContainerId(
			ctx,
			clusterId,
			consoleQuery.ContainerId,
			consoleQuery.Shell,
		)
		return podCtx, err
	}

	// 有任意参数, 使用直连模式
	if consoleQuery.IsContainerDirectMode() {
		return nil, errors.New("container_id或namespace/pod_name/container_name不能同时为空")
	}

	// 默认集群内 kubectl
	targetClusterID := clusterId

	// 集群外 kubectl
	if config.G.WebConsole.IsExternalMode() {
		targetClusterID = config.G.WebConsole.AdminClusterId

	}

	podCtx, err := queryByClusterIdExternal(
		ctx,
		targetClusterID,
		username,
		clusterId,
		consoleQuery.Shell,
	)
	return podCtx, err
}

// OpenQuery openapi 参数
type OpenQuery struct {
	Operator        string   `json:"operator" binding:"required"`
	Viewers         []string `json:"viewers"`           // 可共享查看
	SessionTimeout  int64    `json:"session_timeout"`   // session 过期时间, 单位分钟
	ConnIdleTimeout int64    `json:"conn_idle_timeout"` // 空闲时间, 单位分钟
	Command         string   `json:"command"`
	ContainerId     string   `json:"container_id"`
	Namespace       string   `json:"namespace"`
	PodName         string   `json:"pod_name"`
	ContainerName   string   `json:"container_name"`
	WSAcquire       bool     `json:"ws_acquire"` // 是否返回 websocket_url
}

// Validate 校验参数
func (q *OpenQuery) Validate() error {
	if q.ConnIdleTimeout < 0 || q.ConnIdleTimeout >= types.MaxConnIdleTimeout {
		return errors.Errorf("conn_idle_timeout 必须大于0, 小于%d", types.MaxConnIdleTimeout)
	}
	if q.SessionTimeout < 0 || q.SessionTimeout >= types.MaxSessionTimeout {
		return errors.Errorf("session_timeout 必须大于0, 小于%d", types.MaxSessionTimeout)
	}
	return nil
}

// SplitCommand 拆解命令行
func (q *OpenQuery) SplitCommand() ([]string, error) {
	if q.Command == "" {
		return []string{}, nil
	}
	return shlex.Split(q.Command)
}

// QueryOpenPodCtx openapi鉴权模式
func QueryOpenPodCtx(ctx context.Context, clusterId string, consoleQuery *OpenQuery) (*types.PodContext, error) {
	//  直连模式
	if consoleQuery.Namespace != "" && consoleQuery.PodName != "" && consoleQuery.ContainerName != "" {
		podCtx, err := queryByContainerName(
			ctx,
			clusterId,
			consoleQuery.Namespace,
			consoleQuery.PodName,
			consoleQuery.ContainerName,
			"",
		)
		return podCtx, err
	}

	// 通过容器ID直连
	if consoleQuery.ContainerId != "" {
		podCtx, err := queryByContainerId(
			ctx,
			clusterId,
			consoleQuery.ContainerId,
			"",
		)
		return podCtx, err
	}

	return nil, errors.New("container_id或namespace/pod_name/container_name不能同时为空")
}
