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
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/chonla/format"
	"github.com/dustin/go-humanize"
	"github.com/prometheus/common/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/promclient"
)

var tracer = otel.Tracer("bcs_monitor_client")

// QueryInstant 查询实时数据, 带格式化
func QueryInstant(ctx context.Context, projectId string, promql string, params map[string]interface{},
	t time.Time) (*promclient.Result, error) {
	var rawQL string
	if params == nil {
		rawQL = promql
	} else {
		rawQL = format.Sprintf(promql, params)
	}

	commonAttrs := []attribute.KeyValue{
		attribute.String("projectId", projectId),
		attribute.String("rawQL", rawQL),
		attribute.String("params", MapToJson(params)),
	}
	ctx, span := tracer.Start(ctx, "QueryInstant", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()

	queryURL, header := getQueryURL()
	instant, err := promclient.QueryInstant(ctx, queryURL, header, rawQL, t)
	rspData, _ := json.Marshal(instant.Data.Result)
	respBody := string(rspData)
	if len(respBody) > 1024 {
		respBody = fmt.Sprintf("%s...(Total %s)", respBody[:1024], humanize.Bytes(uint64(len(respBody))))
	}
	// 设置额外标签
	span.SetAttributes(attribute.String("queryURL", queryURL))
	span.SetAttributes(attribute.String("time", t.String()))
	span.SetAttributes(attribute.Key("rsp").String(respBody))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return instant, err
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

	commonAttrs := []attribute.KeyValue{
		attribute.String("rawQL", rawQL),
		attribute.String("projectId", projectId),
		attribute.String("params", MapToJson(params)),
	}
	ctx, span := tracer.Start(ctx, "QueryInstantVector", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()
	queryURL, header := getQueryURL()
	vectorResult, warnings, err := promclient.QueryInstantVector(ctx, queryURL, header, rawQL, t)
	rspData, _ := json.Marshal(vectorResult.String())
	respBody := string(rspData)
	if len(respBody) > 1024 {
		respBody = fmt.Sprintf("%s...(Total %s)", respBody[:1024], humanize.Bytes(uint64(len(respBody))))
	}
	// 设置额外标签
	span.SetAttributes(attribute.String("queryURL", queryURL))
	span.SetAttributes(attribute.String("time", t.String()))
	span.SetAttributes(attribute.Key("rsp").String(respBody))
	span.SetAttributes(attribute.Key("warnings").StringSlice(warnings))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return vectorResult, warnings, err
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
	commonAttrs := []attribute.KeyValue{
		attribute.String("projectId", projectId),
		attribute.String("rawQL", rawQL),
		attribute.String("params", MapToJson(params)),
	}
	ctx, span := tracer.Start(ctx, "QueryRange", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()
	queryURL, header := getQueryURL()
	queryRangeResult, err := promclient.QueryRange(ctx, queryURL, header, rawQL, start, end, step)

	rspData, _ := json.Marshal(queryRangeResult.Data.Result)
	respBody := string(rspData)
	if len(respBody) > 1024 {
		respBody = fmt.Sprintf("%s...(Total %s)", respBody[:1024], humanize.Bytes(uint64(len(respBody))))
	}
	// 设置额外标签
	span.SetAttributes(attribute.String("queryURL", queryURL))
	span.SetAttributes(attribute.String("start", start.String()))
	span.SetAttributes(attribute.String("end", end.String()))
	span.SetAttributes(attribute.String("step", step.String()))
	span.SetAttributes(attribute.Key("rsp").String(respBody))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return queryRangeResult, err
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
	commonAttrs := []attribute.KeyValue{
		attribute.String("projectId", projectId),
		attribute.String("rawQL", rawQL),
		attribute.String("params", MapToJson(params)),
	}
	ctx, span := tracer.Start(ctx, "QueryRangeMatrix", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()
	queryURL, header := getQueryURL()

	matrixResult, warnings, err := promclient.QueryRangeMatrix(ctx, queryURL, header, rawQL, start, end, step)
	rspData, _ := json.Marshal(matrixResult.String())
	respBody := string(rspData)
	if len(respBody) > 1024 {
		respBody = fmt.Sprintf("%s...(Total %s)", respBody[:1024], humanize.Bytes(uint64(len(respBody))))
	}
	// 设置额外标签
	span.SetAttributes(attribute.String("queryURL", queryURL))
	span.SetAttributes(attribute.String("start", start.String()))
	span.SetAttributes(attribute.String("end", end.String()))
	span.SetAttributes(attribute.String("step", step.String()))
	span.SetAttributes(attribute.Key("rsp").String(respBody))
	span.SetAttributes(attribute.Key("warnings").StringSlice(warnings))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return matrixResult, warnings, err
}

// QueryValue 查询第一个值 format 格式 %<var>s
func QueryValue(ctx context.Context, projectId string, promql string, params map[string]interface{},
	t time.Time) (string, error) {
	commonAttrs := []attribute.KeyValue{
		attribute.String("promql", promql),
		attribute.String("params", MapToJson(params)),
	}
	ctx, span := tracer.Start(ctx, "QueryValue", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()
	vector, warnings, err := QueryInstantVector(ctx, projectId, promql, params, t)
	rspData, _ := json.Marshal(vector.String())
	respBody := string(rspData)
	if len(respBody) > 1024 {
		respBody = fmt.Sprintf("%s...(Total %s)", respBody[:1024], humanize.Bytes(uint64(len(respBody))))
	}
	// 设置额外标签
	span.SetAttributes(attribute.String("time", t.String()))
	span.SetAttributes(attribute.Key("rsp").String(respBody))
	span.SetAttributes(attribute.Key("warnings").StringSlice(warnings))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	return GetFirstValue(vector), nil
}

// QueryMultiValues 查询第一个值 format 格式 %<var>s
func QueryMultiValues(ctx context.Context, projectId string, promqlMap map[string]string, params map[string]interface{},
	t time.Time) (map[string]string, error) {
	commonAttrs := []attribute.KeyValue{
		attribute.String("projectId", projectId),
		attribute.String("params", MapToJson(params)),
	}
	ctx, span := tracer.Start(ctx, "QueryMultiValues", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()
	var (
		wg  sync.WaitGroup
		mtx sync.Mutex
	)

	defaultValue := "0"

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
				klog.Warningf("query_multi_values %s error, %s", promql, err)
				resultMap[key] = defaultValue
			} else {
				resultMap[key] = GetFirstValue(vector)
			}
		}(k, v)
	}

	wg.Wait()

	promqlMapStr, _ := json.Marshal(promqlMap)
	resultMapStr, _ := json.Marshal(resultMap)
	span.SetAttributes(attribute.String("promql", string(promqlMapStr)))
	span.SetAttributes(attribute.String("resultMap", string(resultMapStr)))
	return resultMap, nil
}

// QueryLabelSet 查询
func QueryLabelSet(ctx context.Context, projectId string, promql string, params map[string]interface{},
	t time.Time) (map[string]string, error) {
	commonAttrs := []attribute.KeyValue{
		attribute.String("promql", promql),
		attribute.String("params", MapToJson(params)),
	}
	ctx, span := tracer.Start(ctx, "QueryLabelSet", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()
	vector, _, err := QueryInstantVector(ctx, projectId, promql, params, t)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	result := GetLabelSet(vector)
	resultStr, _ := json.Marshal(result)
	span.SetAttributes(attribute.Key("labelSet").String(string(resultStr)))
	return result, nil
}
