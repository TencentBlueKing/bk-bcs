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

package metrics

import (
	bcsmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

// ContainerUsageQuery 容器查询参数
type ContainerUsageQuery struct {
	UsageQuery    `json:",inline"`
	Namespace     string `json:"namespace"`
	PodName       string `json:"pod_name"`
	ContainerName string `json:"container_name"`
}

// handleContainerMetric Container 处理公共函数
func handleContainerMetric(c *rest.Context, promql string) (interface{}, error) {
	query := &UsageQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		return nil, err
	}

	queryTime, err := query.GetQueryTime()
	if err != nil {
		return nil, err
	}

	params := map[string]interface{}{
		"clusterId":     c.ClusterId,
		"namespace":     c.Param("namespace"),
		"podName":       c.Param("pod"),
		"containerName": c.Param("container"),
		"provider":      PROVIDER,
	}

	result, err := bcsmonitor.QueryRange(c.Request.Context(), c.ProjectCode, promql, params, queryTime.Start, queryTime.End,
		queryTime.Step)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// ContainerCPUUsage 容器 CPU 使用率
// @Summary 容器 CPU 使用率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/namespace/pods/:pod/containers/:container/cpu_usage [GET]
func ContainerCPUUsage(c *rest.Context) (interface{}, error) {
	promql :=
		`bcs:container:cpu_usage{cluster_id="%<clusterId>s", namespace="%<namespace>s", pod_name=~"%<podName>s", container_name=~"%<containerName>s", %<provider>s}`

	return handleContainerMetric(c, promql)

}

// ContainerMemoryUsed 容器内存使用量
// @Summary 容器内存使用量
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/namespace/pods/:pod/containers/:container/memory_used [GET]
func ContainerMemoryUsed(c *rest.Context) (interface{}, error) {
	promql :=
		`bcs:container:memory_used{cluster_id="%<clusterId>s", namespace="%<namespace>s", pod_name=~"%<podName>s", container_name=~"%<containerName>s", %<provider>s}`

	return handleContainerMetric(c, promql)
}

// ContainerCPULimit 容器 CPU 限制
// @Summary 容器 CPU 限制
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/namespace/pods/:pod/containers/:container/cpu_limit [GET]
func ContainerCPULimit(c *rest.Context) (interface{}, error) {
	promql :=
		`bcs:container:cpu_limit{cluster_id="%<clusterId>s", namespace="%<namespace>s", pod_name=~"%<podName>s", container_name=~"%<containerName>s", %<provider>s}`

	return handleContainerMetric(c, promql)
}

// ContainerMemoryLimit 容器内存限制
// @Summary 容器内存限制
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/namespace/pods/:pod/containers/:container/memory_limit [GET]
func ContainerMemoryLimit(c *rest.Context) (interface{}, error) {
	promql :=
		`bcs:container:memory_limit{cluster_id="%<clusterId>s", namespace="%<namespace>s", pod_name=~"%<podName>s", container_name=~"%<containerName>s", %<provider>s}`

	return handleContainerMetric(c, promql)

}

// ContainerDiskReadTotal 容器磁盘读总量
// @Summary 容器磁盘读总量
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/namespace/pods/:pod/containers/:container/disk_read_total [GET]
func ContainerDiskReadTotal(c *rest.Context) (interface{}, error) {
	promql :=
		`bcs:container:disk_read_total{cluster_id="%<clusterId>s", namespace="%<namespace>s", pod_name=~"%<podName>s", container_name=~"%<containerName>s", %<provider>s}`

	return handleContainerMetric(c, promql)

}

// ContainerDiskWriteTotal 容器磁盘写总量
// @Summary 容器磁盘写总量
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/namespace/pods/:pod/containers/:container/disk_write_total [GET]
func ContainerDiskWriteTotal(c *rest.Context) (interface{}, error) {
	promql :=
		`bcs:container:disk_write_total{cluster_id="%<clusterId>s", namespace="%<namespace>s", pod_name=~"%<podName>s", container_name=~"%<containerName>s", %<provider>s}`

	return handleContainerMetric(c, promql)
}
