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

package clientutil

const (
	// MinStepSeconds 最小步长, 单位秒
	MinStepSeconds = 60
	// SeriesStepDeltaSeconds 查询 Series 的回溯步长, 单位秒
	SeriesStepDeltaSeconds = 60 * 5
)

// MonitorSourceType monitor source type
type MonitorSourceType string

const (
	// MonitorSourceCompute 算力 metrics
	MonitorSourceCompute MonitorSourceType = "compute"
	// MonitorSourceComputeV2 算力 metrics v2，数据来源于 bk-monitor
	MonitorSourceComputeV2 MonitorSourceType = "compute_v2"
	// MonitorSourceFederation 联邦集群 metrics
	MonitorSourceFederation MonitorSourceType = "federation"
)

// DispatchConf xxx
type DispatchConf struct {
	ClusterID     string            `yaml:"cluster_id"`
	URL           string            `yaml:"url"`
	MetricsPrefix string            `yaml:"metrics_prefix"` // 某些集群 metrics 有特定的前缀
	SourceType    MonitorSourceType `yaml:"source_type"`
}
