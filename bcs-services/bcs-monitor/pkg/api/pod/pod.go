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

// Package pod pod operate
package pod

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

// GetPodContainers 获取 Pod 容器列表
// @Summary 获取 Pod 容器列表
// @Tags    Logs
// @Produce json
// @Success 200 {array} k8sclient.Container
// @Router  /namespaces/:namespace/pods/:pod/containers [get]
func GetPodContainers(c context.Context, req k8sclient.PodContainersReq) ([]*k8sclient.Container, error) {
	clusterId := req.ClusterId
	namespace := req.Namespace
	pod := req.Pod
	containers, err := k8sclient.GetPodContainers(c, clusterId, namespace, pod)
	if err != nil {
		return nil, err
	}

	return containers, nil
}

// GetPodLog 查询容器日志
// @Summary 查询容器日志
// @Tags    Logs
// @Param   container_name query string true "容器名称"
// @Param   previous       query string true "是否使用上一次日志, 异常退出使用"
// @Produce json
// @Success 200 {array} k8sclient.Log
// @Router  /namespaces/:namespace/pods/:pod/logs [get]
func GetPodLog(c context.Context, logQuery k8sclient.LogQuery) (*k8sclient.LogWithPreviousLink, error) {
	projectId := logQuery.ProjectId
	clusterId := logQuery.ClusterId
	namespace := logQuery.Namespace
	pod := logQuery.Pod
	logs, err := k8sclient.GetPodLog(c, clusterId, namespace, pod, &logQuery)
	if err != nil {
		return nil, err
	}

	if err := logs.MakePreviousLink(projectId, clusterId, namespace, pod, &logQuery); err != nil {
		return nil, err
	}

	return logs, nil
}

// DownloadPodLog 下载日志
// @Summary 下载日志
// @Tags    Logs
// @Param   container_name query string true "容器名称"
// @Param   previous       query string true "是否使用上一次日志, 异常退出使用"
// @Produce octet-stream
// @Success 200 {string} string
// @Router  /namespaces/:namespace/pods/:pod/logs/download [get]
func DownloadPodLog(logQuery k8sclient.LogQuery, ss rest.StreamingServer) error {
	clusterId := logQuery.ClusterId
	namespace := logQuery.Namespace
	pod := logQuery.Pod

	// 下载参数
	logQuery.TailLines = k8sclient.MAX_TAIL_LINES
	logQuery.LimitBytes = int64(k8sclient.MAX_LIMIT_BYTES)

	logs, err := k8sclient.GetPodLogByte(ss.Context(), clusterId, namespace, pod, &logQuery)
	if err != nil {
		return err
	}

	ts := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s-%s-%s.log", pod, logQuery.ContainerName, ts)

	rest.WriteAttachment(ss, logs, filename)
	return nil
}
