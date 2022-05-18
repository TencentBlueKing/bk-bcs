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
	"github.com/spf13/viper"
)

const (
	QueryMaxConCurrentQueriesConfKey     = "query.max_concurrent_queries"   // 最大并行查询数
	QueryMaxConCurrentSelectsConfKey     = "query.max_concurrent_selects"   // 一次查询最多的并行数
	QueryDefaultRangeQueryStepConfKey    = "query.default_query_step"       // 默认查询步长
	QueryStoreTimeoutConfKey             = "query.query_store_timeout"      // 查询store的超时时间
	QueryMaxLookBackDeltaConfKey         = "query.max_lookback_delta"       // 最大回溯时间，当步长太短，回溯去找上一个点的最大时间
	QueryDynamicLookbackDeltaConfKey     = "query.dynamic_lookup_delata"    // 允许具有解析的查询，具有更大的回溯时间
	QueryEnableAutoDownsamplingConfKey   = "query.enable_auto_downsampling" // 自动降采样，（如果max_source_resolution没配置的话）
	QueryEnableQueryPartialConfKey       = "query.enable_partial"           // query模块的 部分响应参数
	QueryMaxSourceResolutionConfKey      = "query.max_source_resolution"    // 即使查询的最大分辨率
	QueryDefaultMetadataTimeRangeConfKey = "query.default_metadata_range"   // metadata的默认检索时间范围。0代表从头到尾全部
	QueryUnhealthyStoreTimeoutKey        = "query.unhealthy_store_timeout"  // 健康检查的时间
	QueryStoreListKey                    = "query.store"                    // store list
	QueryStoreRespTimeoutKey             = "query.store_resp_timeout"       // store list
)

var queryReplicaLabels = []string{"bcs_monitor_replica"}

func init() {
	viper.SetDefault(QueryMaxConCurrentQueriesConfKey, 20)
	viper.SetDefault(QueryMaxConCurrentSelectsConfKey, 4)
	viper.SetDefault(QueryDefaultRangeQueryStepConfKey, 0)
	viper.SetDefault(QueryStoreTimeoutConfKey, "2m")
	viper.SetDefault(QueryMaxLookBackDeltaConfKey, "5m")
	viper.SetDefault(QueryDynamicLookbackDeltaConfKey, true)
	viper.SetDefault(QueryEnableAutoDownsamplingConfKey, false)
	viper.SetDefault(QueryEnableQueryPartialConfKey, true)
	viper.SetDefault(QueryMaxSourceResolutionConfKey, 0)
	viper.SetDefault(QueryDefaultMetadataTimeRangeConfKey, 0)
	viper.SetDefault(QueryUnhealthyStoreTimeoutKey, "5m")
	viper.SetDefault(QueryStoreListKey, nil)
	viper.SetDefault(QueryStoreRespTimeoutKey, 0)
}
