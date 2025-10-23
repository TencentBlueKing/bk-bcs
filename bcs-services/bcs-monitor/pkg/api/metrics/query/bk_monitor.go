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

package query

import (
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/chonla/format"
	"gopkg.in/yaml.v2"

	bkmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/promclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

// BKMonitorHandler metric handler
type BKMonitorHandler struct {
	bkBizID string
	url     string
}

// handleBKMonitorClusterMetric Cluster 处理公共函数
func (h BKMonitorHandler) handleBKMonitorClusterMetric(
	c *rest.Context, promql string, query *UsageQuery) (*promclient.ResultData, error) {

	queryTime, err := query.GetQueryTime()
	if err != nil {
		return nil, err
	}

	_, nodeNameMatch, ok := GetMasterNodeMatchIgnoreErr(c.Request.Context(), c.ClusterId)
	if !ok {
		return nil, nil
	}

	nodeFormat := ""
	if nodeNameMatch != "" {
		nodeFormat = fmt.Sprintf(`, node!~"%s"`, nodeNameMatch)
	}
	rawQL := format.Sprintf(promql, map[string]interface{}{
		"clusterID":  c.ClusterId,
		"node":       nodeFormat,
		"fstype":     utils.FSType,
		"mountpoint": config.G.BKMonitor.MountPoint,
	})

	result, err := bkmonitor.QueryByPromQLRaw(c.Request.Context(), h.url, h.bkBizID, c.TenantId, queryTime.Start.Unix(),
		queryTime.End.Unix(), int64(queryTime.Step.Seconds()), nil, rawQL)
	if err != nil {
		return nil, err
	}

	raw, err := json.Marshal(result.Series)
	if err != nil {
		return nil, err
	}
	return &promclient.ResultData{Result: raw}, nil
}

// GetClusterOverview 获取集群概览
// nolint
func (h BKMonitorHandler) GetClusterOverview(c *rest.Context) (*ClusterOverviewMetric, error) {
	_, nodeNameMatch, ok := GetMasterNodeMatchIgnoreErr(c.Request.Context(), c.ClusterId)
	if !ok {
		return nil, nil
	}

	nodeFormat := ""
	if nodeNameMatch != "" {
		nodeFormat = fmt.Sprintf(`, node!~"%s"`, nodeNameMatch)
	}
	params := map[string]interface{}{
		"clusterID":  c.ClusterId,
		"node":       nodeFormat,
		"fstype":     utils.FSType,
		"mountpoint": config.G.BKMonitor.MountPoint,
	}

	promqlMap := map[string]string{
		"cpu_used":       `sum(rate(bkmonitor:container_cpu_usage_seconds_total{pod_name!="", bcs_cluster_id="%<clusterID>s", node!=""%<node>s}[2m]))`,
		"cpu_request":    `sum(avg_over_time(bkmonitor:kube_pod_container_resource_requests_cpu_cores{bcs_cluster_id="%<clusterID>s", node!=""%<node>s}[2m]))`,
		"cpu_total":      `sum(avg_over_time(bkmonitor:kube_node_status_allocatable_cpu_cores{bcs_cluster_id="%<clusterID>s", node!=""%<node>s}[2m]))`,
		"memory_used":    `sum(avg_over_time(bkmonitor:container_memory_usage_bytes{pod_name!="", bcs_cluster_id="%<clusterID>s", node!=""%<node>s}[2m]))`,
		"memory_request": `sum(avg_over_time(bkmonitor:kube_pod_container_resource_requests_memory_bytes{bcs_cluster_id="%<clusterID>s", node!=""%<node>s}[2m]))`,
		"memory_total":   `sum(avg_over_time(bkmonitor:kube_node_status_allocatable_memory_bytes{bcs_cluster_id="%<clusterID>s", node!=""%<node>s}[2m]))`,
		"disk_used": `(sum(avg_over_time(bkmonitor:node_filesystem_size_bytes{bcs_cluster_id="%<clusterID>s", fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s"}[2m])) ` +
			`- sum(avg_over_time(bkmonitor:node_filesystem_free_bytes{bcs_cluster_id="%<clusterID>s", fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s"}[2m])))`,
		"disk_total":   `sum(avg_over_time(bkmonitor:node_filesystem_size_bytes{bcs_cluster_id="%<clusterID>s", fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s"}[2m]))`,
		"diskio_used":  `sum(max by(instance) (rate(bkmonitor:node_disk_io_time_seconds_total{bcs_cluster_id="%<clusterID>s"}[2m])))`,
		"diskio_total": `count(max by(instance) (rate(bkmonitor:node_disk_io_time_seconds_total{bcs_cluster_id="%<clusterID>s"}[2m])))`,
		"pod_used":     `sum(avg_over_time(bkmonitor:kubelet_running_pods{bcs_cluster_id="%<clusterID>s", node!=""%<node>s}[2m]))`,
		"pod_total":    `sum(avg_over_time(bkmonitor:kube_node_status_capacity_pods{bcs_cluster_id="%<clusterID>s", node!=""%<node>s}[2m]))`,
	}

	result, err := bkmonitor.QueryMultiValues(c.Request.Context(), h.url, h.bkBizID, c.TenantId,
		utils.GetNowQueryTime().Unix(), promqlMap, params)
	if err != nil {
		return nil, err
	}

	m := ClusterOverviewMetric{
		CPUUsage: &Usage{
			Used:    result["cpu_used"],
			Request: result["cpu_request"],
			Total:   result["cpu_total"],
		},
		MemoryUsage: &UsageByte{
			UsedByte:    result["memory_used"],
			RequestByte: result["memory_request"],
			TotalByte:   result["memory_total"],
		},
		DiskUsage: &UsageByte{
			UsedByte:  result["disk_used"],
			TotalByte: result["disk_total"],
		},
		DiskIOUsage: &Usage{
			Used:  result["diskio_used"],
			Total: result["diskio_total"],
		},
		PodUsage: &Usage{
			Used:  result["pod_used"],
			Total: result["pod_total"],
		},
	}

	return &m, nil
}

// ClusterPodUsage implements Handler.
// nolint
func (h BKMonitorHandler) ClusterPodUsage(c *rest.Context, query *UsageQuery) (*promclient.ResultData, error) {
	promql := `sum(avg_over_time(bkmonitor:kubelet_running_pods{bcs_cluster_id="%<clusterID>s", node!=""%<node>s}[2m])) / ` +
		`sum(avg_over_time(bkmonitor:kube_node_status_capacity_pods{bcs_cluster_id="%<clusterID>s", node!=""%<node>s}[2m])) * 100`

	return h.handleBKMonitorClusterMetric(c, promql, query)
}

// ClusterCPUUsage implements Handler.
// nolint
func (h BKMonitorHandler) ClusterCPUUsage(c *rest.Context, query *UsageQuery) (*promclient.ResultData, error) {
	promql := `sum(rate(bkmonitor:container_cpu_usage_seconds_total{pod_name!="", bcs_cluster_id="%<clusterID>s", node!=""%<node>s}[2m])) / ` +
		`sum(avg_over_time(bkmonitor:kube_node_status_allocatable_cpu_cores{bcs_cluster_id="%<clusterID>s", node!=""%<node>s}[2m])) * 100`

	return h.handleBKMonitorClusterMetric(c, promql, query)
}

// ClusterCPURequestUsage implements Handler.
// nolint
func (h BKMonitorHandler) ClusterCPURequestUsage(c *rest.Context, query *UsageQuery) (*promclient.ResultData, error) {
	promql := `sum(avg_over_time(bkmonitor:kube_pod_container_resource_requests_cpu_cores{bcs_cluster_id="%<clusterID>s", node!=""%<node>s}[2m])) / ` +
		`sum(avg_over_time(bkmonitor:kube_node_status_allocatable_cpu_cores{bcs_cluster_id="%<clusterID>s", node!=""%<node>s}[2m])) * 100`

	return h.handleBKMonitorClusterMetric(c, promql, query)
}

// ClusterMemoryUsage implements Handler.
// nolint
func (h BKMonitorHandler) ClusterMemoryUsage(c *rest.Context, query *UsageQuery) (*promclient.ResultData, error) {
	promql := `sum(avg_over_time(bkmonitor:container_memory_usage_bytes{pod_name!="", bcs_cluster_id="%<clusterID>s", node!=""%<node>s}[2m])) / ` +
		`sum(avg_over_time(bkmonitor:kube_node_status_allocatable_memory_bytes{bcs_cluster_id="%<clusterID>s", node!=""%<node>s}[2m])) * 100`

	return h.handleBKMonitorClusterMetric(c, promql, query)
}

// ClusterMemoryRequestUsage implements Handler.
// nolint
func (h BKMonitorHandler) ClusterMemoryRequestUsage(c *rest.Context, query *UsageQuery) (*promclient.ResultData, error) {
	promql := `sum(avg_over_time(bkmonitor:kube_pod_container_resource_requests_memory_bytes{bcs_cluster_id="%<clusterID>s", node!=""%<node>s}[2m])) / ` +
		`sum(avg_over_time(bkmonitor:kube_node_status_allocatable_memory_bytes{bcs_cluster_id="%<clusterID>s", node!=""%<node>s}[2m])) * 100`

	return h.handleBKMonitorClusterMetric(c, promql, query)
}

// ClusterDiskUsage implements Handler.
// nolint
func (h BKMonitorHandler) ClusterDiskUsage(c *rest.Context, query *UsageQuery) (*promclient.ResultData, error) {
	promql := `(sum(avg_over_time(bkmonitor:node_filesystem_size_bytes{bcs_cluster_id="%<clusterID>s", fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s"}[2m])) - ` +
		`sum(avg_over_time(bkmonitor:node_filesystem_free_bytes{bcs_cluster_id="%<clusterID>s", fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s"}[2m]))) / ` +
		`sum(avg_over_time(bkmonitor:node_filesystem_size_bytes{bcs_cluster_id="%<clusterID>s", fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s"}[2m])) * 100`

	return h.handleBKMonitorClusterMetric(c, promql, query)
}

// ClusterDiskioUsage implements Handler.
// nolint
func (h BKMonitorHandler) ClusterDiskioUsage(c *rest.Context, query *UsageQuery) (*promclient.ResultData, error) {
	promql := `sum(max by(instance) (rate(bkmonitor:node_disk_io_time_seconds_total{bcs_cluster_id="%<clusterID>s"}[2m]))) / ` +
		`count(max by(instance) (rate(bkmonitor:node_disk_io_time_seconds_total{bcs_cluster_id="%<clusterID>s"}[2m]))) * 100`

	return h.handleBKMonitorClusterMetric(c, promql, query)
}

// NewBKMonitorHandler new handler
func NewBKMonitorHandler(bkBizID, clusterID string) *BKMonitorHandler {
	url := config.G.BKMonitor.URL
	ll := config.G.StoreGWList
	for _, v := range ll {
		if v.Type == config.BKMONITOR {
			c, err := yaml.Marshal(v.Config)
			if err != nil {
				blog.Errorf("marshal content of store configuration error, %s", err.Error())
				continue
			}
			var conf Config
			if err := yaml.UnmarshalStrict(c, &conf); err != nil {
				blog.Errorf("parsing bk_monitor store config error, %s", err.Error())
				continue
			}
			for _, dis := range conf.Dispatch {
				if clusterID == dis.ClusterID {
					url = dis.URL
					break
				}
			}
		}
	}
	return &BKMonitorHandler{bkBizID: bkBizID, url: url}
}

var _ Handler = BKMonitorHandler{}
