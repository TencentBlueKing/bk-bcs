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

package config

// StoreProvider :
type StoreProvider string

// StoreConf :
type StoreConf struct {
	Type   StoreProvider `yaml:"type"`
	Config interface{}   `yaml:"config,omitempty"`
}

const (
	// BKMONITOR 蓝鲸监控数据源
	BKMONITOR StoreProvider = "BK_MONITOR"
	// BCS_SYSTEM bcs 自定义 metrics
	BCS_SYSTEM StoreProvider = "BCS_SYSTEM"
	// PROMETHEUS prometheus 原生数据源
	PROMETHEUS StoreProvider = "PROMETHEUS"
	// SUANLI_CPU 算力 cpu 数据源
	SUANLI_CPU StoreProvider = "SUANLI_CPU"
	// SUANLI_GPU_NATIVE 算力 gpu 数据源
	SUANLI_GPU_NATIVE StoreProvider = "SUANLI_GPU_NATIVE"
)
