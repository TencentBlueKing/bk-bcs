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

// Package query metric query
package query

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/thanos-io/thanos/pkg/store"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs"
	bkmonitor_client "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/promclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

// Handler metric handler
type Handler interface {
	GetClusterOverview(c *rest.Context) (*ClusterOverviewMetric, error)
	ClusterPodUsage(c *rest.Context, query *UsageQuery) (*promclient.ResultData, error)
	ClusterCPUUsage(c *rest.Context, query *UsageQuery) (*promclient.ResultData, error)
	ClusterCPURequestUsage(c *rest.Context, query *UsageQuery) (*promclient.ResultData, error)
	ClusterMemoryUsage(c *rest.Context, query *UsageQuery) (*promclient.ResultData, error)
	ClusterMemoryRequestUsage(c *rest.Context, query *UsageQuery) (*promclient.ResultData, error)
	ClusterDiskUsage(c *rest.Context, query *UsageQuery) (*promclient.ResultData, error)
	ClusterDiskioUsage(c *rest.Context, query *UsageQuery) (*promclient.ResultData, error)
}

// HandlerFactory 自动切换Prometheus/蓝鲸监控
func HandlerFactory(ctx context.Context, clusterID string) (Handler, error) {
	cls, err := bcs.GetCluster(clusterID)
	if err != nil {
		return nil, err
	}
	if bkmonitor_client.IsBKMonitorEnabled(clusterID) && !cls.IsVirtual() {
		return NewBKMonitorHandler(cls.BKBizID, clusterID), nil
	}
	return NewBCSMonitorHandler(), nil
}

// GetMasterNodeMatch 按集群node节点正则匹配
func GetMasterNodeMatch(ctx context.Context, clusterID string) (string, string, error) {
	nodeList, nodeNameList, err := k8sclient.GetMasterNodeList(ctx, clusterID)
	if err != nil {
		return "", "", err
	}
	return utils.StringJoinWithRegex(nodeList, "|", ".*"), utils.StringJoinWithRegex(nodeNameList, "|", "$"), nil
}

// GetMasterNodeMatchIgnoreErr 按集群node节点正则匹配
func GetMasterNodeMatchIgnoreErr(ctx context.Context, clusterID string) (string, string, bool) {
	nodeList, nodeNameList, err := GetMasterNodeMatch(ctx, clusterID)
	if err != nil {
		blog.Infow("get node error", "request_id", store.RequestIDValue(ctx), "cluster_id", clusterID,
			"err", err.Error())
		return "", "", false
	}
	return nodeList, nodeNameList, true
}
