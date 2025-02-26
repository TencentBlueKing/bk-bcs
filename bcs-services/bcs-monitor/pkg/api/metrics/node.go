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
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	bcsmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/promclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/clientutil"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

// GetNodeInfoReq xxx
type GetNodeInfoReq struct {
	ProjectCode string `json:"projectCode" in:"path=projectCode" validate:"required"`
	ClusterId   string `json:"clusterId" in:"path=clusterId" validate:"required"`
	Node        string `json:"node" in:"path=node" validate:"required"`
}

// NodeOveriewMetric 节点概览
type NodeOveriewMetric struct {
	ContainerCount     string `json:"container_count"`
	PodTotal           string `json:"pod_total"`
	PodCount           string `json:"pod_count"`
	CPUUsed            string `json:"cpu_used"`
	CPURequest         string `json:"cpu_request"`
	CPUTotal           string `json:"cpu_total"`
	CPUUsage           string `json:"cpu_usage"`
	CPURequestUsage    string `json:"cpu_request_usage"`
	DiskUsed           string `json:"disk_used"`
	DiskTotal          string `json:"disk_total"`
	DiskUsage          string `json:"disk_usage"`
	DiskioUsage        string `json:"diskio_usage"`
	MemoryUsed         string `json:"memory_used"`
	MemoryRequest      string `json:"memory_request"`
	MemoryTotal        string `json:"memory_total"`
	MemoryUsage        string `json:"memory_usage"`
	MemoryRequestUsage string `json:"memory_request_usage"`
}

// UsageQuery 节点查询
type UsageQuery struct {
	ProjectCode string `json:"projectCode" in:"path=projectCode" validate:"required"`
	ClusterId   string `json:"clusterId" in:"path=clusterId" validate:"required"`
	Nodegroup   string `json:"nodegroup" in:"path=nodegroup"`
	Node        string `json:"node" in:"path=node"`
	Namespace   string `json:"namespace" in:"path=namespace"`
	Pod         string `json:"pod" in:"path=pod"`
	Container   string `json:"container" in:"path=container"`
	StartAt     string `json:"start_at" in:"query=start_at" form:"start_at"` // 必填参数`
	EndAt       string `json:"end_at" in:"query=end_at" form:"end_at"`
}

// Nodes 列表
type Nodes struct {
	ProjectCode string   `json:"projectCode" in:"path=projectCode" validate:"required"`
	ClusterId   string   `json:"clusterId" in:"path=clusterId" validate:"required"`
	Node        []string `json:"node"`
}

const (
	// 限制并发为 20，防止对数据源造成压力
	defaultQueryConcurrent = 20
)

// parseTime 兼容前端多个格式
func parseTime(rawTime string) (time.Time, error) {
	// 和前端约定, 只支持这种带时区的格式
	formats := []string{
		"2006-01-02T15:04:05Z",
	}

	for _, format := range formats {
		t, err := time.Parse(format, rawTime)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, errors.Errorf("invalid datetime %s", rawTime)
}

// GetQueryTime 转换为 promql 查询时间
func (q *UsageQuery) GetQueryTime() (*clientutil.PromQueryTime, error) {
	queryTime := &clientutil.PromQueryTime{}

	if q.EndAt == "" {
		queryTime.End = utils.GetNowQueryTime()
	} else {
		t, err := parseTime(q.EndAt)
		if err != nil {
			return nil, err
		}
		queryTime.End = t
	}

	if q.StartAt == "" {
		queryTime.Start = queryTime.End.Add(-time.Hour)
	} else {
		t, err := parseTime(q.StartAt)
		if err != nil {
			return nil, err
		}
		queryTime.Start = t
	}

	// 默认只返回 60 个点
	queryTime.Step = queryTime.End.Sub(queryTime.Start) / 60

	return queryTime, nil
}

// handleNodeMetric Node 处理公共函数
func handleNodeMetric(c *rest.Context, promql string, query *UsageQuery) (*promclient.ResultData, error) {

	queryTime, err := query.GetQueryTime()
	if err != nil {
		return nil, err
	}

	params := map[string]interface{}{
		"clusterId": c.ClusterId,
		"node":      query.Node,
		"provider":  PROVIDER,
	}

	result, err := bcsmonitor.QueryRange(c.Request.Context(), c.ProjectCode, promql, params, queryTime.Start,
		queryTime.End, queryTime.Step)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// GetNodeInfo 节点信息
// @Summary 节点信息
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /nodes/:node/info [get]
func GetNodeInfo(c context.Context, req *GetNodeInfoReq) (*map[string]string, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	params := map[string]interface{}{
		"clusterId": rctx.ClusterId,
		"node":      req.Node,
		"provider":  PROVIDER,
	}

	promql := `bcs:node:info{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`
	labelSet, err := bcsmonitor.QueryLabelSet(c, rctx.ProjectId, promql, params, utils.GetNowQueryTime())
	if err != nil {
		return nil, err
	}
	return &labelSet, nil
}

// GetNodeOverview 查询节点概览
// @Summary 查询节点概览
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /nodes/:node/overview [get]
func GetNodeOverview(c context.Context, req *GetNodeInfoReq) (*NodeOveriewMetric, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	params := map[string]interface{}{
		"clusterId": rctx.ClusterId,
		"node":      req.Node,
		"provider":  PROVIDER,
	}

	promqlMap := map[string]string{
		"container_count":      `bcs:node:container_count{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"pod_total":            `bcs:node:pod_total{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"pod_count":            `bcs:node:pod_count{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"cpu_used":             `bcs:node:cpu:used{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"cpu_request":          `bcs:node:cpu:request{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"cpu_total":            `bcs:node:cpu:total{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"cpu_usage":            `bcs:node:cpu:usage{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"cpu_request_usage":    `bcs:node:cpu_request:usage{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"memory_used":          `bcs:node:memory:used{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"memory_request":       `bcs:node:memory:request{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"memory_total":         `bcs:node:memory:total{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"memory_usage":         `bcs:node:memory:usage{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"memory_request_usage": `bcs:node:memory_request:usage{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"disk_usage":           `bcs:node:disk:usage{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"disk_used":            `bcs:node:disk:used{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"disk_total":           `bcs:node:disk:total{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"diskio_usage":         `bcs:node:diskio:usage{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
	}

	result, err := bcsmonitor.QueryMultiValues(c, rctx.ProjectId, promqlMap, params,
		utils.GetNowQueryTime())
	if err != nil {
		return nil, err
	}

	overview := &NodeOveriewMetric{
		ContainerCount:     result["container_count"],
		PodTotal:           result["pod_total"],
		PodCount:           result["pod_count"],
		CPUUsed:            result["cpu_used"],
		CPURequest:         result["cpu_request"],
		CPUTotal:           result["cpu_total"],
		CPUUsage:           result["cpu_usage"],
		CPURequestUsage:    result["cpu_request_usage"],
		DiskUsed:           result["disk_used"],
		DiskTotal:          result["disk_total"],
		DiskUsage:          result["disk_usage"],
		DiskioUsage:        result["diskio_usage"],
		MemoryUsed:         result["memory_used"],
		MemoryRequest:      result["memory_request"],
		MemoryTotal:        result["memory_total"],
		MemoryUsage:        result["memory_usage"],
		MemoryRequestUsage: result["memory_request_usage"],
	}

	return overview, nil
}

// ListNodeOverviews 查询节点列表概览
// @Summary 查询节点列表概览
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /nodes/overviews [post]
func ListNodeOverviews(c context.Context, nodes *Nodes) (*map[string]*NodeOveriewMetric, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	nodeOveriewMetrics := make(map[string]*NodeOveriewMetric, len(nodes.Node))

	var mtx sync.Mutex
	var wg errgroup.Group
	wg.SetLimit(defaultQueryConcurrent)

	promqlMap := map[string]string{
		"container_count":      `bcs:node:container_count{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"pod_total":            `bcs:node:pod_total{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"pod_count":            `bcs:node:pod_count{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"cpu_used":             `bcs:node:cpu:used{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"cpu_request":          `bcs:node:cpu:request{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"cpu_total":            `bcs:node:cpu:total{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"cpu_usage":            `bcs:node:cpu:usage{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"cpu_request_usage":    `bcs:node:cpu_request:usage{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"memory_used":          `bcs:node:memory:used{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"memory_request":       `bcs:node:memory:request{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"memory_total":         `bcs:node:memory:total{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"memory_usage":         `bcs:node:memory:usage{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"memory_request_usage": `bcs:node:memory_request:usage{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"disk_usage":           `bcs:node:disk:usage{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"disk_used":            `bcs:node:disk:used{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"disk_total":           `bcs:node:disk:total{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
		"diskio_usage":         `bcs:node:diskio:usage{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`,
	}
	for _, node := range nodes.Node {
		node := node
		wg.Go(func() error {
			params := map[string]interface{}{
				"clusterId": rctx.ClusterId,
				"node":      node,
				"provider":  PROVIDER,
			}

			// 设计如此，忽略报错原因如下两点：
			// 1、忽略报错信息，底层函数始终也是返回err nil的情况；
			// 2、如果报错的情况下，result始终也是有默认值的，而且结果需要返回空的结构体
			result, _ := bcsmonitor.QueryMultiValues(c, rctx.ProjectId, promqlMap, params,
				utils.GetNowQueryTime())

			overview := &NodeOveriewMetric{
				ContainerCount:     result["container_count"],
				PodTotal:           result["pod_total"],
				PodCount:           result["pod_count"],
				CPUUsed:            result["cpu_used"],
				CPURequest:         result["cpu_request"],
				CPUTotal:           result["cpu_total"],
				CPUUsage:           result["cpu_usage"],
				CPURequestUsage:    result["cpu_request_usage"],
				DiskUsed:           result["disk_used"],
				DiskTotal:          result["disk_total"],
				DiskUsage:          result["disk_usage"],
				DiskioUsage:        result["diskio_usage"],
				MemoryUsed:         result["memory_used"],
				MemoryRequest:      result["memory_request"],
				MemoryTotal:        result["memory_total"],
				MemoryUsage:        result["memory_usage"],
				MemoryRequestUsage: result["memory_request_usage"],
			}
			mtx.Lock()
			nodeOveriewMetrics[node] = overview
			mtx.Unlock()
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return nil, err
	}
	return &nodeOveriewMetrics, nil
}

// GetNodeCPUUsage 查询 CPU 使用率
// @Summary 查询 CPU 使用率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /nodes/:node/cpu_usage [get]
func GetNodeCPUUsage(c context.Context, req *UsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	promql := `bcs:node:cpu:usage{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`

	return handleNodeMetric(rctx, promql, req)
}

// GetNodeCPURequestUsage 查询 CPU 装箱率
// @Summary 查询 CPU 装箱率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /nodes/:node/cpu_request_usage [get]
func GetNodeCPURequestUsage(c context.Context, req *UsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	promql := `bcs:node:cpu_request:usage{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`

	return handleNodeMetric(rctx, promql, req)
}

// GetNodeMemoryUsage 节点内存使用率
// @Summary 节点内存使用率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /nodes/:node/memory_usage [get]
func GetNodeMemoryUsage(c context.Context, req *UsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	promql := `bcs:node:memory:usage{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`

	return handleNodeMetric(rctx, promql, req)
}

// GetNodeMemoryRequestUsage 节点内存装箱率
// @Summary 节点内存装箱率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /nodes/:node/memory_request_usage [get]
func GetNodeMemoryRequestUsage(c context.Context, req *UsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	promql := `bcs:node:memory_request:usage{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`

	return handleNodeMetric(rctx, promql, req)
}

// GetNodeNetworkTransmitUsage 节点网络发送
// @Summary 节点网络发送
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /nodes/:node/network_receive [get]
func GetNodeNetworkTransmitUsage(c context.Context, req *UsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	promql := `bcs:node:network_transmit{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`

	return handleNodeMetric(rctx, promql, req)
}

// GetNodeNetworkReceiveUsage 节点网络接收
// @Summary 节点网络接收
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /nodes/:node/network_transmit [get]
func GetNodeNetworkReceiveUsage(c context.Context, req *UsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	promql := `bcs:node:network_receive{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`

	return handleNodeMetric(rctx, promql, req)
}

// GetNodeDiskUsage 节点磁盘使用率
// @Summary 节点磁盘使用率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /nodes/:node/disk_usage [get]
func GetNodeDiskUsage(c context.Context, req *UsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	promql := `bcs:node:disk:usage{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`

	return handleNodeMetric(rctx, promql, req)
}

// GetNodeDiskioUsage 节点磁盘IO
// @Summary 节点磁盘IO
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /nodes/:node/diskio_usage [get]
func GetNodeDiskioUsage(c context.Context, req *UsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	promql := `bcs:node:diskio:usage{cluster_id="%<clusterId>s", node="%<node>s", %<provider>s}`

	return handleNodeMetric(rctx, promql, req)
}
