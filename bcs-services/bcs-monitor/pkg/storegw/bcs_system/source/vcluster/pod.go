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

package vcluster

import (
	"context"
	"time"

	"github.com/prometheus/prometheus/prompb"

	bcsmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

// handlePodMetric xxx
func (m *VCluster) handlePodMetric(ctx context.Context, projectID, clusterID, namespace string, podNameList []string,
	promql string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterID":   clusterID,
		"namespace":   namespace,
		"podNameList": utils.StringJoinWithRegex(podNameList, "|", "$"),
		"provider":    PROVIDER,
	}

	matrix, _, err := bcsmonitor.QueryRangeMatrix(ctx, projectID, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetPodCPUUsage Pod CPU 使用率
func (m *VCluster) GetPodCPUUsage(ctx context.Context, projectID, clusterID, namespace string, podNameList []string,
	start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum by (pod_name) (rate(container_cpu_usage_seconds_total{cluster_id="%<clusterID>s", ` +
			`namespace="%<namespace>s", pod_name=~"%<podNameList>s", container_name!="", container_name!="POD", ` +
			`%<provider>s}[2m])) * 100`

	return m.handlePodMetric(ctx, projectID, clusterID, namespace, podNameList, promql, start, end, step)
}

// GetPodCPULimitUsage POD CPU Limit 使用率
func (m *VCluster) GetPodCPULimitUsage(ctx context.Context, projectID, clusterID, namespace string,
	podNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum by (pod_name) (rate(container_cpu_usage_seconds_total{cluster_id="%<clusterID>s", ` +
			`namespace="%<namespace>s", pod_name=~"%<podNameList>s", container_name!="", container_name!="POD", ` +
			`%<provider>s}[2m])) / ` +
			`sum by (pod_name) (kube_pod_container_resource_limits_cpu_cores{cluster_id="%<clusterID>s", ` +
			`job="kube-state-metrics", namespace="%<namespace>s", pod_name=~"%<podNameList>s", container_name!="", ` +
			`container_name!="POD",  %<provider>s}) * 100`

	return m.handlePodMetric(ctx, projectID, clusterID, namespace, podNameList, promql, start, end, step)
}

// GetPodCPURequestUsage POD CPU Request 使用率
func (m *VCluster) GetPodCPURequestUsage(ctx context.Context, projectID, clusterID, namespace string,
	podNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum by (pod_name) (rate(container_cpu_usage_seconds_total{cluster_id="%<clusterID>s", ` +
			`namespace="%<namespace>s", pod_name=~"%<podNameList>s", container_name!="", container_name!="POD", ` +
			`%<provider>s}[2m])) / ` +
			`sum by (pod_name) (kube_pod_container_resource_requests_cpu_cores{cluster_id="%<clusterID>s", ` +
			`job="kube-state-metrics", namespace="%<namespace>s", pod_name=~"%<podNameList>s", container_name!="", ` +
			`container_name!="POD",  %<provider>s}) * 100`

	return m.handlePodMetric(ctx, projectID, clusterID, namespace, podNameList, promql, start, end, step)
}

// GetPodMemoryUsed 内存使用量
func (m *VCluster) GetPodMemoryUsed(ctx context.Context, projectID, clusterID, namespace string, podNameList []string,
	start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum by (pod_name) (container_memory_working_set_bytes{cluster_id="%<clusterID>s", ` +
			`namespace="%<namespace>s", pod_name=~"%<podNameList>s", container_name!="", container_name!="POD", ` +
			`%<provider>s})`

	return m.handlePodMetric(ctx, projectID, clusterID, namespace, podNameList, promql, start, end, step)
}

// GetPodNetworkReceive 网络接收
func (m *VCluster) GetPodNetworkReceive(ctx context.Context, projectID, clusterID, namespace string,
	podNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum by (pod_name) (rate(container_network_receive_bytes_total{cluster_id="%<clusterID>s", ` +
			`namespace="%<namespace>s", pod_name=~"%<podNameList>s", %<provider>s}[2m]))`

	return m.handlePodMetric(ctx, projectID, clusterID, namespace, podNameList, promql, start, end, step)
}

// GetPodNetworkTransmit 网络发送
func (m *VCluster) GetPodNetworkTransmit(ctx context.Context, projectID, clusterID, namespace string,
	podNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum by (pod_name) (rate(container_network_transmit_bytes_total{cluster_id="%<clusterID>s", ` +
			`namespace="%<namespace>s", pod_name=~"%<podNameList>s", %<provider>s}[2m]))`

	return m.handlePodMetric(ctx, projectID, clusterID, namespace, podNameList, promql, start, end, step)
}
