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

// Package source client
package source

import (
	"context"

	bkmonitor_client "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
	bkmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/bk_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/compute"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/computev2"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/federation"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/prometheus"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/vcluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/clientutil"
)

// ClientFactory 自动切换Prometheus/蓝鲸监控
func ClientFactory(ctx context.Context, clusterId string, source clientutil.MonitorSourceType,
	dispatch map[string]clientutil.DispatchConf, isVCluster bool) (base.MetricHandler, error) {
	switch source {
	case clientutil.MonitorSourceCompute:
		return compute.NewCompute(), nil
	case clientutil.MonitorSourceFederation:
		return federation.NewFederation(dispatch), nil
	case clientutil.MonitorSourceComputeV2:
		return computev2.NewCompute(dispatch[clusterId].MetricsPrefix), nil
	default:
		ok, err := bkmonitor_client.IsBKMonitorEnabled(ctx, clusterId)
		if err != nil {
			return nil, err
		}
		if ok {
			if isVCluster {
				return vcluster.NewVCluster(), nil
			}
			return bkmonitor.NewBKMonitor(), nil
		}
		return prometheus.NewPrometheus(), nil
	}
}
