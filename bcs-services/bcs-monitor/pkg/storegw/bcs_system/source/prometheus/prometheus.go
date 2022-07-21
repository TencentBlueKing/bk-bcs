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

package prometheus

import (
	"context"
	"fmt"
	"time"

	bcsmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
	"github.com/prometheus/prometheus/prompb"
)

// Prometheus
type Prometheus struct {
}

// NewPrometheus
func NewPrometheus() *Prometheus {
	return &Prometheus{}
}

// GetClusterCPUTotal 获取集群CPU核心总量
func (p *Prometheus) GetClusterCPUTotal(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	instanceMatch, err := base.GetNodeMatch(ctx, clusterId)
	if err != nil {
		return nil, err
	}

	promql := fmt.Sprintf(
		`sum(count without(cpu, mode) (node_cpu_seconds_total{cluster_id="%s", job="node-exporter", mode="idle", instance=~"%s"}))`,
		clusterId,
		instanceMatch,
	)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, promql, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterCPUUsed 获取CPU核心使用量
func (p *Prometheus) GetClusterCPUUsed(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	instanceMatch, err := base.GetNodeMatch(ctx, clusterId)
	if err != nil {
		return nil, err
	}

	promql := fmt.Sprintf(
		`sum(irate(node_cpu_seconds_total{cluster_id="%s", job="node-exporter", mode!="idle", instance=~"%s"}[2m]))`,
		clusterId,
		instanceMatch,
	)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, promql, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}
