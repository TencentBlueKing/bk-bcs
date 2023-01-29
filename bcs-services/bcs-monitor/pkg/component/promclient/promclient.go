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

// Package promclient xxx
package promclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql/parser"
	"github.com/thanos-io/thanos/pkg/store/storepb"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
)

var tracer = otel.Tracer("prom_client")

// PromStatus prometheus api status
type PromStatus string

const (
	// PromSuccess prometheus api success
	PromSuccess PromStatus = "success"
	// PromError prometheus api error
	PromError PromStatus = "error"
)

// BaseResponse prometheus api response
type BaseResponse struct {
	Status PromStatus `json:"status"`
	// Only set if status is "error".
	Error     string   `json:"error,omitempty"`
	ErrorType string   `json:"errorType,omitempty"`
	Warnings  []string `json:"warnings,omitempty"` // Extra field supported by Thanos Querier.
}

// IsSuccess check prometheus api is success
func (r BaseResponse) IsSuccess() bool {
	return r.Status == PromSuccess
}

// Result xxx
// Decode only ResultType and load Result only as RawJson since we don't know
// structure of the Result yet.
type Result struct {
	Data ResultData `json:"data"`
	BaseResponse
}

// ResultData :
type ResultData struct {
	ResultType string          `json:"resultType"`
	Result     json.RawMessage `json:"result"`
}

// LabelValuesResponse label values response
type LabelValuesResponse struct {
	Data []string `json:"data"`
	BaseResponse
}

// QueryInstant 查询实时数据
func QueryInstant(ctx context.Context, rawURL string, header http.Header, promql string, t time.Time) (*Result, error) {
	rawURL = strings.TrimSuffix(rawURL, "/") + "/api/v1/query"
	commonAttrs := []attribute.KeyValue{
		attribute.String("rawURL", rawURL),
	}
	ctx, span := tracer.Start(ctx, "QueryInstant", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()
	data := map[string]string{
		"query": promql,
		"time":  t.Format(time.RFC3339Nano),
	}
	dataStr, _ := json.Marshal(data)
	span.SetAttributes(attribute.String("data", string(dataStr)))
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetFormData(data).
		SetHeaderMultiValues(header).
		Post(rawURL)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if !resp.IsSuccess() {
		err = errors.Errorf("http code %d != 200", resp.StatusCode())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	m := Result{}

	if err = json.Unmarshal(resp.Body(), &m); err != nil {
		err = errors.Wrap(err, "unmarshal query instant response")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	respBody := string(resp.Body())
	if len(respBody) > 1024 {
		respBody = fmt.Sprintf("%s...(Total %s)", respBody[:1024], humanize.Bytes(uint64(len(respBody))))
	}
	span.SetAttributes(attribute.String("time", t.String()))
	span.SetAttributes(attribute.Key("rsp").String(respBody))

	return &m, nil
}

// QueryRange 查询历史数据
func QueryRange(ctx context.Context, rawURL string, header http.Header, promql string, start time.Time, end time.Time,
	step time.Duration) (*Result, error) {
	data := map[string]string{
		"query": promql,
		"start": start.Format(time.RFC3339Nano),
		"end":   end.Format(time.RFC3339Nano),
		"step":  strconv.FormatInt(int64(step.Seconds()), 10) + "s",
	}
	rawURL = strings.TrimSuffix(rawURL, "/") + "/api/v1/query_range"

	dataStr, _ := json.Marshal(data)
	commonAttrs := []attribute.KeyValue{
		attribute.String("data", string(dataStr)),
		attribute.String("rawURL", rawURL),
	}
	ctx, span := tracer.Start(ctx, "QueryRange", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetFormData(data).
		SetHeaderMultiValues(header).
		Post(rawURL)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if !resp.IsSuccess() {
		err = errors.Errorf("http code %d != 200", resp.StatusCode())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	m := Result{}

	if err = json.Unmarshal(resp.Body(), &m); err != nil {
		err = errors.Wrap(err, "unmarshal query range response")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	respBody := string(resp.Body())
	if len(respBody) > 1024 {
		respBody = fmt.Sprintf("%s...(Total %s)", respBody[:1024], humanize.Bytes(uint64(len(respBody))))
	}
	span.SetAttributes(attribute.Key("rsp").String(respBody))
	return &m, nil
}

// QueryInstantVector 查询实时数据
func QueryInstantVector(ctx context.Context, rawURL string, header http.Header, promql string,
	t time.Time) (model.Vector, []string, error) {
	commonAttrs := []attribute.KeyValue{
		attribute.String("rawURL", rawURL),
	}
	ctx, span := tracer.Start(ctx, "QueryInstantVector", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()
	m, err := QueryInstant(ctx, rawURL, header, promql, t)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, err
	}

	var vectorResult model.Vector

	// Decode the Result depending on the ResultType
	// Currently only `vector` and `scalar` types are supported.
	switch m.Data.ResultType {
	case string(parser.ValueTypeVector):
		if err = json.Unmarshal(m.Data.Result, &vectorResult); err != nil {
			err = errors.Wrap(err, "decode result into ValueTypeVector")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, nil, err
		}
	case string(parser.ValueTypeScalar):
		vectorResult, err = convertScalarJSONToVector(m.Data.Result)
		if err != nil {
			err = errors.Wrap(err, "decode result into ValueTypeScalar")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, nil, err
		}
	default:
		if m.Warnings != nil {
			err = errors.Errorf("error: %s, type: %s, warning: %s", m.Error, m.ErrorType, strings.Join(m.Warnings,
				", "))
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, nil, err
		}
		if m.Error != "" {
			err = errors.Errorf("error: %s, type: %s", m.Error, m.ErrorType)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, nil, err
		}
		err = errors.Errorf("received status code: 200, unknown response type: '%q'", m.Data.ResultType)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, err
	}
	vectorResultStr, _ := json.Marshal(vectorResult)
	respBody := string(vectorResultStr)
	if len(vectorResultStr) > 1024 {
		respBody = fmt.Sprintf("%s...(Total %s)", respBody[:1024], humanize.Bytes(uint64(len(respBody))))
	}
	span.SetAttributes(attribute.Key("vectorResult").String(respBody))
	span.SetAttributes(attribute.Key("warnings").StringSlice(m.Warnings))
	return vectorResult, m.Warnings, nil
}

// QueryRangeMatrix 查询历史数据
func QueryRangeMatrix(ctx context.Context, rawURL string, header http.Header, promql string, start time.Time,
	end time.Time, step time.Duration) (model.Matrix, []string, error) {
	commonAttrs := []attribute.KeyValue{
		attribute.String("rawURL", rawURL),
	}
	ctx, span := tracer.Start(ctx, "QueryRangeMatrix", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()
	m, err := QueryRange(ctx, rawURL, header, promql, start, end, step)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, err
	}

	var matrixResult model.Matrix

	// Decode the Result depending on the ResultType
	switch m.Data.ResultType {
	case string(parser.ValueTypeMatrix):
		if err = json.Unmarshal(m.Data.Result, &matrixResult); err != nil {
			err = errors.Wrap(err, "decode result into ValueTypeMatrix")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, nil, err
		}
	default:
		if m.Warnings != nil {
			err = errors.Errorf("error: %s, type: %s, warning: %s", m.Error, m.ErrorType, strings.Join(m.Warnings,
				", "))
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, nil, err
		}
		if m.Error != "" {
			err = errors.Errorf("error: %s, type: %s", m.Error, m.ErrorType)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, nil, err
		}

		err = errors.Errorf("received status code: 200, unknown response type: '%q'", m.Data.ResultType)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, err
	}
	matrixResultStr, _ := json.Marshal(matrixResult)
	respBody := string(matrixResultStr)
	if len(matrixResultStr) > 1024 {
		respBody = fmt.Sprintf("%s...(Total %s)", respBody[:1024], humanize.Bytes(uint64(len(respBody))))
	}
	span.SetAttributes(attribute.Key("matrixResult").String(respBody))
	span.SetAttributes(attribute.Key("warnings").StringSlice(m.Warnings))
	return matrixResult, m.Warnings, nil

}

// QueryLabels query labels
func QueryLabels(ctx context.Context, rawURL string, header http.Header, r *storepb.LabelNamesRequest) ([]string, error) {
	rawURL = fmt.Sprintf("%s/api/v1/labels", strings.TrimSuffix(rawURL, "/"))

	query := make(map[string]string, 0)
	if r.Start != 0 {
		query["start"] = strconv.Itoa(int(r.Start))
	}
	if r.End != 0 {
		query["end"] = strconv.Itoa(int(r.End))
	}

	queryStr, _ := json.Marshal(query)
	commonAttrs := []attribute.KeyValue{
		attribute.String("query", string(queryStr)),
		attribute.String("rawURL", rawURL),
	}
	ctx, span := tracer.Start(ctx, "QueryLabels", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()

	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetQueryParams(query).
		SetHeaderMultiValues(header).
		Get(rawURL)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if !resp.IsSuccess() {
		err = errors.Errorf("http code %d != 200", resp.StatusCode())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	m := LabelValuesResponse{}

	if err = json.Unmarshal(resp.Body(), &m); err != nil {
		err = errors.Wrap(err, "unmarshal query labels response")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if m.IsSuccess() {
		span.SetAttributes(attribute.StringSlice("data", m.Data))
		return m.Data, nil
	}

	err = errors.Errorf("errorType: %s, error: %s", m.ErrorType, m.Error)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return nil, err
}

// QueryLabelValues query label values
func QueryLabelValues(ctx context.Context, rawURL string, header http.Header, r *storepb.LabelValuesRequest) ([]string, error) {
	rawURL = fmt.Sprintf("%s/api/v1/label/%s/values", strings.TrimSuffix(rawURL, "/"), r.Label)

	query := make(map[string]string, 0)
	if r.Start != 0 {
		query["start"] = strconv.Itoa(int(r.Start))
	}
	if r.End != 0 {
		query["end"] = strconv.Itoa(int(r.End))
	}
	queryStr, _ := json.Marshal(query)
	commonAttrs := []attribute.KeyValue{
		attribute.String("query", string(queryStr)),
		attribute.String("rawURL", rawURL),
	}
	ctx, span := tracer.Start(ctx, "QueryLabels", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetPathParams(query).
		SetHeaderMultiValues(header).
		Get(rawURL)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if !resp.IsSuccess() {
		err = errors.Errorf("http code %d != 200", resp.StatusCode())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	m := LabelValuesResponse{}

	if err = json.Unmarshal(resp.Body(), &m); err != nil {
		err = errors.Wrap(err, "unmarshal query label values response")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if m.IsSuccess() {
		span.SetAttributes(attribute.StringSlice("data", m.Data))
		return m.Data, nil
	}

	err = errors.Errorf("errorType: %s, error: %s", m.ErrorType, m.Error)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return nil, err
}

// convertScalarJSONToVector xxx
// Scalar response consists of array with mixed types so it needs to be
// unmarshaled separately.
func convertScalarJSONToVector(scalarJSONResult json.RawMessage) (model.Vector, error) {
	var (
		// Do not specify exact length of the expected slice since JSON unmarshaling
		// would make the length fit the size and we won't be able to check the length afterwards.
		resultPointSlice []json.RawMessage
		resultTime       model.Time
		resultValue      model.SampleValue
	)
	if err := json.Unmarshal(scalarJSONResult, &resultPointSlice); err != nil {
		return nil, err
	}
	if len(resultPointSlice) != 2 {
		return nil, errors.Errorf("invalid scalar result format %v, expected timestamp -> value tuple", resultPointSlice)
	}
	if err := json.Unmarshal(resultPointSlice[0], &resultTime); err != nil {
		return nil, errors.Wrapf(err, "unmarshaling scalar time from %v", resultPointSlice)
	}
	if err := json.Unmarshal(resultPointSlice[1], &resultValue); err != nil {
		return nil, errors.Wrapf(err, "unmarshaling scalar value from %v", resultPointSlice)
	}
	return model.Vector{&model.Sample{
		Metric:    model.Metric{},
		Value:     resultValue,
		Timestamp: resultTime}}, nil
}
