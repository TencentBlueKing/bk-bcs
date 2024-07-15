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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

// handleGroupMetric group 处理公共函数
func handleGroupMetric(c *rest.Context, promql string) (interface{}, error) {
	query := &UsageQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		return nil, err
	}

	queryTime, err := query.GetQueryTime()
	if err != nil {
		return nil, err
	}

	params := map[string]interface{}{
		"clusterId": config.G.BKMonitor.ClusterID,
		"group":     c.Param("nodegroup"),
		"provider":  PROVIDER,
	}

	result, err := bcsmonitor.QueryRange(c.Request.Context(), c.ProjectCode, promql, params, queryTime.Start,
		queryTime.End, queryTime.Step)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// ClusterGroupNodeNum 集群节点池数目
// @Summary 集群节点池数目
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /group/:group/node_num [get]
func ClusterGroupNodeNum(c *rest.Context) (interface{}, error) {
	promql := `bcs:cluster:group:node_num{cluster_id="%<clusterId>s", group="%<group>s", %<provider>s}`

	return handleGroupMetric(c, promql)
}

// ClusterGroupMaxNodeNum 集群最大节点池数目
// @Summary 集群最大节点池数目
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /group/:group/max_node_num [get]
func ClusterGroupMaxNodeNum(c *rest.Context) (interface{}, error) {
	promql := `bcs:cluster:group:max_node_num{cluster_id="%<clusterId>s", group="%<group>s", %<provider>s}`

	return handleGroupMetric(c, promql)
}
