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

// Package metrics api metric
package metrics

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/api/metrics/query"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/promclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

const (
	// PROVIDER provider
	PROVIDER = `provider="BCS_SYSTEM"`
)

// GetClusterOverview 集群概览数据
// @Summary 集群概览数据
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /overview [get]
func GetClusterOverview(c context.Context, req *query.GetClusterOverviewReq) (*query.ClusterOverviewMetric, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	handler, err := query.HandlerFactory(c, req.ClusterId)
	if err != nil {
		return nil, err
	}
	return handler.GetClusterOverview(rctx)
}

// ClusterPodUsage 集群 POD 使用率
// @Summary 集群 POD 使用率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /pod_usage [get]
func ClusterPodUsage(c context.Context, req *query.UsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	handler, err := query.HandlerFactory(c, rctx.ClusterId)
	if err != nil {
		return nil, err
	}
	return handler.ClusterPodUsage(rctx, req)
}

// ClusterCPUUsage 集群 CPU 使用率
// @Summary 集群 CPU 使用率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /cpu_usage [get]
func ClusterCPUUsage(c context.Context, req *query.UsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	handler, err := query.HandlerFactory(c, rctx.ClusterId)
	if err != nil {
		return nil, err
	}
	return handler.ClusterCPUUsage(rctx, req)
}

// ClusterCPURequestUsage 集群 CPU 装箱率
// @Summary 集群 CPU 装箱率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /cpu_request_usage [get]
func ClusterCPURequestUsage(c context.Context, req *query.UsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	handler, err := query.HandlerFactory(c, rctx.ClusterId)
	if err != nil {
		return nil, err
	}
	return handler.ClusterCPURequestUsage(rctx, req)
}

// ClusterMemoryUsage 集群内存使用率
// @Summary 集群内存使用率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /memory_usage [get]
func ClusterMemoryUsage(c context.Context, req *query.UsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	handler, err := query.HandlerFactory(c, rctx.ClusterId)
	if err != nil {
		return nil, err
	}
	return handler.ClusterMemoryUsage(rctx, req)
}

// ClusterMemoryRequestUsage 集群内存装箱率
// @Summary 集群内存装箱率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /memory_request_usage [get]
func ClusterMemoryRequestUsage(c context.Context, req *query.UsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	handler, err := query.HandlerFactory(c, rctx.ClusterId)
	if err != nil {
		return nil, err
	}
	return handler.ClusterMemoryRequestUsage(rctx, req)
}

// ClusterDiskUsage 集群磁盘使用率
// @Summary 集群磁盘使用率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /disk_usage [get]
func ClusterDiskUsage(c context.Context, req *query.UsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	handler, err := query.HandlerFactory(c, rctx.ClusterId)
	if err != nil {
		return nil, err
	}
	return handler.ClusterDiskUsage(rctx, req)
}

// ClusterDiskioUsage 集群磁盘IO使用率
// @Summary 集群磁盘IO使用率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /diskio_usage [get]
func ClusterDiskioUsage(c context.Context, req *query.UsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	handler, err := query.HandlerFactory(c, rctx.ClusterId)
	if err != nil {
		return nil, err
	}
	return handler.ClusterDiskioUsage(rctx, req)
}
