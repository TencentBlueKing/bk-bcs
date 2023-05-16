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
 *
 */

package query

import (
	"time"
)

// 减少配置, 默认使用 thanos 的参数配置
const (
	maxConcurrentQueries              = 20 * 1000        // 最大并行查询数, 提供并发查询, bcs-system storegw 也发并发, 如果 20 会blocks自己
	maxConcurrentSelects              = 4                // 一次查询最多的并行数
	defaultRangeQueryStep             = time.Second * 30 // 默认查询步长
	queryTimeout                      = time.Second * 20 // 查询超时时间
	lookbackDelta                     = time.Minute * 5  // 最大回溯时间，当步长太短，回溯去找上一个点的最大时间
	dynamicLookbackDelta              = true             // 允许具有解析的查询，具有更大的回溯时间
	enableAutodownsampling            = false            // 自动降采样，（如果max_source_resolution没配置的话）
	enableQueryPartialResponse        = true             // query模块的 部分响应参数
	instantDefaultMaxSourceResolution = 0                // metadata的默认检索时间范围。0代表从头到尾全部
	defaultMetadataTimeRange          = 0                // 即使查询的最大分辨率
	unhealthyStoreTimeout             = time.Minute * 5  // 健康检查的时间
	storeResponseTimeout              = time.Second * 20 // 查询store的超时时间
	defaultEvaluationInterval         = time.Minute * 1  // 自查询的默认处理间隔。这里用不到
)

var queryReplicaLabels = []string{"prometheus_replica"}
