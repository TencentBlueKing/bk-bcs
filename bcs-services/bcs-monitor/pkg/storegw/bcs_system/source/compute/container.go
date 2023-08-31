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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

const (
	// PROVIDER xxx
	PROVIDER = `provider="PROMETHEUS"`
)

// handleContainerMetric 处理公共函数
func (m *Compute) handleContainerMetric(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, promql string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {

	params := map[string]interface{}{
		"clusterID":     clusterID,
		"namespace":     namespace,
		"podname":       podname,
		"containerName": utils.StringJoinWithRegex(containerNameList, "|", "$"),
		"provider":      PROVIDER,
	}

	matrix, _, err := bcsmonitor.QueryRangeMatrix(ctx, projectID, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetContainerCPUUsage 容器CPU使用率
func (m *Compute) GetContainerCPUUsage(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum by(container_name) (rate(container_cpu_usage_seconds_total_value{cluster_id="%<clusterID>s", ` +
			`namespace="%<namespace>s", pod_name="%<podname>s", container_name=~"%<containerName>s", ` +
			`%<provider>s}[2m])) * 100`

	return m.handleContainerMetric(ctx, projectID, clusterID, namespace, podname, containerNameList, promql, start, end,
		step)
}

// GetContainerMemoryUsed 容器内存使用率
func (m *Compute) GetContainerMemoryUsed(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`max by(container_name) (container_memory_working_set_bytes_value{cluster_id="%<clusterID>s", ` +
			`namespace="%<namespace>s", pod_name="%<podname>s", container_name=~"%<containerName>s", %<provider>s})`

	return m.handleContainerMetric(ctx, projectID, clusterID, namespace, podname, containerNameList, promql, start, end,
		step)
}

// GetContainerCPULimit 容器CPU限制
func (m *Compute) GetContainerCPULimit(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	// 过滤掉0
	// 名字替换
	// "container_name", "$0 cpu_limit", "container_name", ".*" 带名称
	promql := `label_replace(max by(container_name) (container_spec_cpu_quota_value{cluster_id="%<clusterID>s", ` +
		`namespace="%<namespace>s", pod_name="%<podname>s", container_name=~"%<containerName>s", ` +
		`%<provider>s}) / 1000 > 0, "container_name", "cpu_limit", "container_name", ".*")`

	return m.handleContainerMetric(ctx, projectID, clusterID, namespace, podname, containerNameList, promql, start, end,
		step)
}

// GetContainerMemoryLimit 容器内存限制
func (m *Compute) GetContainerMemoryLimit(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	// 过滤掉0
	// 名字替换
	promql := `
		label_replace(max by(container_name) (container_spec_memory_limit_bytes_value{cluster_id="%<clusterID>s",` +
		` namespace="%<namespace>s", pod_name="%<podname>s", container_name=~"%<containerName>s", ` +
		`%<provider>s}) > 0, "container_name", "memory_limit", "container_name", ".*")`

	return m.handleContainerMetric(ctx, projectID, clusterID, namespace, podname, containerNameList, promql, start, end,
		step)
}

// GetContainerGPUMemoryUsage 容器GPU显卡使用率
func (m *Compute) GetContainerGPUMemoryUsage(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`max by(container_name) (k8s_container_gpu_mem_copy_util_value{cluster_id="%<clusterID>s", ` +
			`namespace="%<namespace>s", pod_name="%<podname>s", container_name=~"%<containerName>s", %<provider>s})`

	return m.handleContainerMetric(ctx, projectID, clusterID, namespace, podname, containerNameList, promql, start, end,
		step)
}

// GetContainerGPUUsed 容器GPU使用量
func (m *Compute) GetContainerGPUUsed(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`max by(container_name) (k8s_container_gpu_used_value{cluster_id="%<clusterID>s", ` +
			`namespace="%<namespace>s", pod_name="%<podname>s", container_name=~"%<containerName>s", %<provider>s})`

	return m.handleContainerMetric(ctx, projectID, clusterID, namespace, podname, containerNameList, promql, start, end,
		step)
}

// GetContainerGPUUsage 容器GPU使用率
func (m *Compute) GetContainerGPUUsage(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`max by(container_name) (k8s_container_gpu_util_value{cluster_id="%<clusterID>s", ` +
			`namespace="%<namespace>s", pod_name="%<podname>s", container_name=~"%<containerName>s", %<provider>s})`

	return m.handleContainerMetric(ctx, projectID, clusterID, namespace, podname, containerNameList, promql, start, end,
		step)
}

// GetContainerDiskReadTotal 容器磁盘读
func (m *Compute) GetContainerDiskReadTotal(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum by(container_name) (container_fs_reads_bytes_total_value{cluster_id="%<clusterID>s", ` +
			`namespace="%<namespace>s", pod_name="%<podname>s", container_name=~"%<containerName>s", %<provider>s})`

	return m.handleContainerMetric(ctx, projectID, clusterID, namespace, podname, containerNameList, promql, start, end,
		step)
}

// GetContainerDiskWriteTotal 容器磁盘写
func (m *Compute) GetContainerDiskWriteTotal(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum by(container_name) (container_fs_writes_bytes_total_value{cluster_id="%<clusterID>s", ` +
			`namespace="%<namespace>s", pod_name="%<podname>s", container_name=~"%<containerName>s", %<provider>s})`

	return m.handleContainerMetric(ctx, projectID, clusterID, namespace, podname, containerNameList, promql, start, end,
		step)
}
