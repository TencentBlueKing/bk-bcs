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

package bcsmonitor

import (
	"context"
	"time"

	"github.com/chonla/format"
	"github.com/prometheus/common/model"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/promclient"
)

// QueryInstant 查询实时数据, 包含租户等信息
func QueryInstant(ctx context.Context, projectId string, promql string, t time.Time) (model.Vector, []string, error) {
	return promclient.QueryInstant(ctx, getQueryURL(), promql, t)
}

// QueryInstantF 查询实时数据, 带格式化
func QueryInstantF(ctx context.Context, projectId string, promql string, params map[string]interface{}, t time.Time) (model.Vector, []string, error) {
	var rawQL string
	if params == nil {
		rawQL = promql
	} else {
		rawQL = format.Sprintf(promql, params)
	}

	return promclient.QueryInstant(ctx, getQueryURL(), rawQL, t)
}

// MultiQueryInstant 并发查询多个
// func MultiQueryInstant(ctx context.Context, projectId string, promqlList []string, t time.Time) (map[string]model.Vector, []string, error) {

// 	return promclient.QueryInstant(ctx, getQueryURL(), promql, t)
// }

// QueryRange 查询历史数据, 包含租户等信息
func QueryRange(ctx context.Context, projectId string, promql string, start time.Time, end time.Time, step time.Duration) (model.Matrix, []string, error) {
	return promclient.QueryRange(ctx, getQueryURL(), promql, start, end, step)
}

// QueryRangeF 查询历史数据 带格式的查询
func QueryRangeF(ctx context.Context, projectId string, promql string, params map[string]interface{}, start time.Time, end time.Time, step time.Duration) (model.Matrix, []string, error) {
	var rawQL string
	if params == nil {
		rawQL = promql
	} else {
		rawQL = format.Sprintf(promql, params)
	}

	return promclient.QueryRange(ctx, getQueryURL(), rawQL, start, end, step)
}

// QueryValue 查询第一个值 format 格式 %<var>s
func QueryValue(ctx context.Context, projectId string, promql string, params map[string]interface{}, t time.Time) (string, error) {
	vector, _, err := QueryInstantF(ctx, projectId, promql, params, t)
	if err != nil {
		return "", err
	}
	return GetFirstValue(vector), nil
}

// QueryLabelSet 查询
func QueryLabelSet(ctx context.Context, projectId string, promql string, params map[string]interface{}, t time.Time) (map[string]string, error) {
	vector, _, err := QueryInstantF(ctx, projectId, promql, params, t)
	if err != nil {
		return nil, err
	}
	return GetLabelSet(vector), nil
}
