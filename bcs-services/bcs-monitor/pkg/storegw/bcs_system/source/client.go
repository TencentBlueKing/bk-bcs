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

package source

import (
	"context"

	bkmonitor_client "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
	bkmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/bk_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/compute"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/prometheus"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/clientutil"
)

// IsBKMonitorEnabled 集群是否接入到蓝鲸监控
func IsBKMonitorEnabled(ctx context.Context, clusterId string) (bool, error) {
	// 不配置则全量接入
	if len(config.G.BKMonitor.MetadataURL) == 0 {
		return true, nil
	}
	grayClusterMap, err := bkmonitor_client.QueryGrayClusterMap(ctx, config.G.BKMonitor.MetadataURL)
	if err != nil {
		return false, err
	}

	_, ok := grayClusterMap[clusterId]
	return ok, nil
}

// ClientFactory 自动切换Prometheus/蓝鲸监控
func ClientFactory(ctx context.Context, clusterId string, source clientutil.MonitorSourceType) (base.MetricHandler, error) {
	ok, err := IsBKMonitorEnabled(ctx, clusterId)
	if err != nil {
		return nil, err
	}

	if ok {
		return bkmonitor.NewBKMonitor(), nil
	}

	switch source {
	case clientutil.MonitorSourceCompute:
		return compute.NewCompute(), nil
	default:
		return prometheus.NewPrometheus(), nil
	}
}
