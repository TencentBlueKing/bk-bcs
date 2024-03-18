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

package suanlicpu

const (
	concurrency = 32           // 并发数
	provider    = "SUANLI_CPU" // 数据源名称
)

// MetricNames 可选的 metrics
var MetricNames = []string{
	"k8s_container_bs_fs_write_times",                         // 磁盘写IOPS
	"k8s_container_bs_fs_read_times",                          // 磁盘读IOPS
	"k8s_container_bs_fs_write_bytes",                         // 磁盘写流量
	"k8s_container_bs_fs_read_bytes",                          // 磁盘读流量
	"k8s_container_bs_network_transmit_packets",               // 网络出包量
	"k8s_container_bs_network_receive_packets",                // 网络入包量
	"k8s_container_bs_network_transmit_bytes_bw",              // 网络出流量
	"k8s_container_bs_network_receive_bytes_bw",               // 网络入流量
	"k8s_container_bs_mem_no_cache_bytes",                     // 内存（不含cache）实际使用量
	"k8s_container_bs_mem_usage_bytes",                        // 内存实际使用量
	"k8s_container_bs_cpu_core_used",                          // CPU实际使用量
	"k8s_container_bs_rate_mem_no_cache_request",              // 内存（no cache）实际利用率（request)
	"k8s_container_bs_rate_mem_usage_request",                 // 内存实际利用率（request)
	"k8s_container_bs_rate_cpu_core_used_request",             // CPU利用率（request）
	"k8s_container_bs_network_transmit_bit_bw",                // 网络入带宽
	"k8s_container_bs_network_receive_bit_bw",                 // 网络出带宽
	"k8s_container_bs_network_transmit_packets_dropped_total", // 出丢包数
	"k8s_container_bs_network_receive_packets_dropped_total",  // 入丢包数
}

// IgnoreGPULabels 需要过滤的 labels
var IgnoreGPULabels = map[string]string{
	"cluster_id":           "",
	"cluster_display_name": "",
	"pod_type":             "",
	"namespace":            "",
	"pod_name":             "",
	"vm_id":                "", // vm_id 对应 lowerPodID
}
