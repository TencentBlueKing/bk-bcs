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

package suanligpunative

const (
	concurrency = 32                  // 并发数
	provider    = "SUANLI_GPU_NATIVE" // 数据源名称
)

// MetricNames 可选的 metrics
var MetricNames = []string{
	"k8s_container_bs_rate_mem_usage_request",       // 容器内存使用率
	"k8s_container_bs_rate_mem_working_set_request", // 容器内存使用率（working set）
	"k8s_container_bs_cpu_core_used",                // 容器CPU使用量
	"k8s_container_bs_mem_no_cache_bytes",           // 容器内存使用量（no cache）
	"k8s_container_resource_request_gpu",            // 容器GPU申请量（request）
	"k8s_container_bs_mem_working_set_bytes",        // 容器内存使用量（working set）
	"k8s_container_rate_gpu_used_request",           // 容器GPU使用率
	"k8s_container_bs_resource_request_mem",         // 容器内存申请量
	"k8s_container_cpu_core_used",                   // 容器CPU实际使用量
	"k8s_container_bs_mem_usage_bytes",              // 容器内存使用量
	"k8s_container_gpu_used",                        // 容器GPU使用量
	"k8s_container_bs_resource_request_cpu",         // 容器CPU申请量（request）
	"k8s_container_bs_rate_mem_no_cache_request",    // 容器内存使用率（no cache)
}
