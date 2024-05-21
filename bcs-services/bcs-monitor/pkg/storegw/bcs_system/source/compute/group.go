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

package compute

import (
	"context"
	"time"

	"github.com/prometheus/prometheus/prompb"

	bcsmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
)

func (m *Compute) handleGroupMetric(ctx context.Context, projectID, clusterID, group string,
	promql string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterID": clusterID,
		"group":     group,
		"provider":  PROVIDER,
	}

	matrix, _, err := bcsmonitor.QueryRangeMatrix(ctx, projectID, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterGroupNodeNum 集群节点池数目
func (m *Compute) GetClusterGroupNodeNum(ctx context.Context, projectID, clusterID, group string, start,
	end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`bkbcs_clustermanager_cluster_group_nodeNum{cluster_id="%<clusterID>s", group="%<group>s", ` +
			`%<provider>s}`

	return m.handleGroupMetric(ctx, projectID, clusterID, group, promql, start, end, step)
}

// GetClusterGroupMaxNodeNum 集群最大节点池数目
func (m *Compute) GetClusterGroupMaxNodeNum(ctx context.Context, projectID, clusterID, group string, start,
	end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`bkbcs_clustermanager_cluster_group_maxNodeNum{cluster_id="%<clusterID>s", group="%<group>s", ` +
			`%<provider>s}`
	return m.handleGroupMetric(ctx, projectID, clusterID, group, promql, start, end, step)
}
