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

// Package bcsmonitor monitor query
package bcsmonitor

import (
	"context"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/chonla/format"
	"github.com/prometheus/common/model"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/promclient"
)

// QueryInstant 查询实时数据, 带格式化
func QueryInstant(ctx context.Context, projectId string, promql string, params map[string]interface{},
	t time.Time) (*promclient.Result, error) {
	var rawQL string
	if params == nil {
		rawQL = promql
	} else {
		rawQL = format.Sprintf(promql, params)
	}

	queryURL, header := getQueryURL()
	return promclient.QueryInstant(ctx, queryURL, header, rawQL, t)
}

// QueryInstantVector 查询实时数据, 带格式化
func QueryInstantVector(ctx context.Context, projectId string, promql string, params map[string]interface{},
	t time.Time) (model.Vector, []string, error) {
	var rawQL string
	if params == nil {
		rawQL = promql
	} else {
		rawQL = format.Sprintf(promql, params)
	}

	queryURL, header := getQueryURL()
	return promclient.QueryInstantVector(ctx, queryURL, header, rawQL, t)
}

// QueryRange 查询历史数据 带格式的查询
func QueryRange(ctx context.Context, projectId string, promql string, params map[string]interface{}, start time.Time,
	end time.Time, step time.Duration) (*promclient.Result, error) {
	var rawQL string
	if params == nil {
		rawQL = promql
	} else {
		rawQL = format.Sprintf(promql, params)
	}

	queryURL, header := getQueryURL()
	return promclient.QueryRange(ctx, queryURL, header, rawQL, start, end, step)
}

// QueryRangeMatrix 查询历史数据, 包含租户等信息
func QueryRangeMatrix(ctx context.Context, projectId string, promql string, params map[string]interface{},
	start time.Time, end time.Time, step time.Duration) (model.Matrix, []string, error) {
	var rawQL string
	if params == nil {
		rawQL = promql
	} else {
		rawQL = format.Sprintf(promql, params)
	}

	queryURL, header := getQueryURL()
	return promclient.QueryRangeMatrix(ctx, queryURL, header, rawQL, start, end, step)
}

// QueryValue 查询第一个值 format 格式 %<var>s
func QueryValue(ctx context.Context, projectId string, promql string, params map[string]interface{},
	t time.Time) (string, error) {
	vector, _, err := QueryInstantVector(ctx, projectId, promql, params, t)
	if err != nil {
		return "", err
	}
	return GetFirstValue(vector), nil
}

// QueryMultiValues 查询第一个值 format 格式 %<var>s
func QueryMultiValues(ctx context.Context, projectId string, promqlMap map[string]string, params map[string]interface{},
	t time.Time) (map[string]string, error) {
	var (
		wg  sync.WaitGroup
		mtx sync.Mutex
	)

	defaultValue := ""

	resultMap := map[string]string{}

	// promql 数量已知, 不控制并发数量
	for k, v := range promqlMap {
		wg.Add(1)
		go func(key, promql string) {
			defer wg.Done()

			vector, _, err := QueryInstantVector(ctx, projectId, promql, params, t)
			mtx.Lock()
			defer mtx.Unlock()

			// 多个查询不报错, 有默认值
			if err != nil {
				blog.Warnf("query_multi_values %s error, %s", promql, err)
				resultMap[key] = defaultValue
			} else {
				resultMap[key] = GetFirstValue(vector)
			}
		}(k, v)
	}

	wg.Wait()

	return resultMap, nil
}

// QueryLabelSet 查询
func QueryLabelSet(ctx context.Context, projectId string, promql string, params map[string]interface{},
	t time.Time) (map[string]string, error) {
	vector, _, err := QueryInstantVector(ctx, projectId, promql, params, t)
	if err != nil {
		return nil, err
	}
	return GetLabelSet(vector), nil
}
