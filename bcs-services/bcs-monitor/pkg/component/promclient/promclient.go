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

// Package promclient prom client
package promclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql/parser"
	"github.com/thanos-io/thanos/pkg/store/storepb"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

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

	data := map[string]string{
		"query": promql,
		"time":  t.Format(time.RFC3339Nano),
	}
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeaders(utils.GetLaneIDByCtx(ctx)). // 泳道特性
		SetFormData(data).
		SetHeaderMultiValues(header).
		Post(rawURL)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.Errorf("http code %d != 200", resp.StatusCode())
	}

	m := Result{}

	if err = json.Unmarshal(resp.Body(), &m); err != nil {
		return nil, errors.Wrap(err, "unmarshal query instant response")
	}

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

	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeaders(utils.GetLaneIDByCtx(ctx)). // 泳道特性
		SetFormData(data).
		SetHeaderMultiValues(header).
		Post(rawURL)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.Errorf("http code %d != 200", resp.StatusCode())
	}

	m := Result{}

	if err = json.Unmarshal(resp.Body(), &m); err != nil {
		return nil, errors.Wrap(err, "unmarshal query range response")
	}

	return &m, nil
}

// QueryInstantVector 查询实时数据
func QueryInstantVector(ctx context.Context, rawURL string, header http.Header, promql string,
	t time.Time) (model.Vector, []string, error) {
	m, err := QueryInstant(ctx, rawURL, header, promql, t)
	if err != nil {
		return nil, nil, err
	}

	var vectorResult model.Vector

	// Decode the Result depending on the ResultType
	// Currently only `vector` and `scalar` types are supported.
	switch m.Data.ResultType {
	case string(parser.ValueTypeVector):
		if err = json.Unmarshal(m.Data.Result, &vectorResult); err != nil {
			return nil, nil, errors.Wrap(err, "decode result into ValueTypeVector")
		}
	case string(parser.ValueTypeScalar):
		vectorResult, err = convertScalarJSONToVector(m.Data.Result)
		if err != nil {
			return nil, nil, errors.Wrap(err, "decode result into ValueTypeScalar")
		}
	default:
		if m.Warnings != nil {
			return nil, nil, errors.Errorf("error: %s, type: %s, warning: %s", m.Error, m.ErrorType, strings.Join(m.Warnings,
				", "))
		}
		if m.Error != "" {
			return nil, nil, errors.Errorf("error: %s, type: %s", m.Error, m.ErrorType)
		}
		return nil, nil, errors.Errorf("received status code: 200, unknown response type: '%q'", m.Data.ResultType)
	}

	return vectorResult, m.Warnings, nil
}

// QueryRangeMatrix 查询历史数据
func QueryRangeMatrix(ctx context.Context, rawURL string, header http.Header, promql string, start time.Time,
	end time.Time, step time.Duration) (model.Matrix, []string, error) {
	m, err := QueryRange(ctx, rawURL, header, promql, start, end, step)
	if err != nil {
		return nil, nil, err
	}

	var matrixResult model.Matrix

	// Decode the Result depending on the ResultType
	switch m.Data.ResultType {
	case string(parser.ValueTypeMatrix):
		if err = json.Unmarshal(m.Data.Result, &matrixResult); err != nil {
			return nil, nil, errors.Wrap(err, "decode result into ValueTypeMatrix")
		}
	default:
		if m.Warnings != nil {
			return nil, nil, errors.Errorf("error: %s, type: %s, warning: %s", m.Error, m.ErrorType, strings.Join(m.Warnings,
				", "))
		}
		if m.Error != "" {
			return nil, nil, errors.Errorf("error: %s, type: %s", m.Error, m.ErrorType)
		}

		return nil, nil, errors.Errorf("received status code: 200, unknown response type: '%q'", m.Data.ResultType)
	}

	return matrixResult, m.Warnings, nil

}

// QueryLabels query labels
func QueryLabels(ctx context.Context, rawURL string, header http.Header,
	r *storepb.LabelNamesRequest) ([]string, error) {
	rawURL = fmt.Sprintf("%s/api/v1/labels", strings.TrimSuffix(rawURL, "/"))

	query := make(map[string]string, 0)
	if r.Start != 0 {
		query["start"] = strconv.Itoa(int(r.Start))
	}
	if r.End != 0 {
		query["end"] = strconv.Itoa(int(r.End))
	}

	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeaders(utils.GetLaneIDByCtx(ctx)). // 泳道特性
		SetQueryParams(query).
		SetHeaderMultiValues(header).
		Get(rawURL)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.Errorf("http code %d != 200", resp.StatusCode())
	}

	m := LabelValuesResponse{}

	if err = json.Unmarshal(resp.Body(), &m); err != nil {
		return nil, errors.Wrap(err, "unmarshal query labels response")
	}

	if m.IsSuccess() {
		return m.Data, nil
	}

	return nil, errors.Errorf("errorType: %s, error: %s", m.ErrorType, m.Error)
}

// QueryLabelValues query label values
func QueryLabelValues(ctx context.Context, rawURL string, header http.Header,
	r *storepb.LabelValuesRequest) ([]string, error) {
	rawURL = fmt.Sprintf("%s/api/v1/label/%s/values", strings.TrimSuffix(rawURL, "/"), r.Label)

	query := make(map[string]string, 0)
	if r.Start != 0 {
		query["start"] = strconv.Itoa(int(r.Start))
	}
	if r.End != 0 {
		query["end"] = strconv.Itoa(int(r.End))
	}
	query["match[]"] = storepb.MatchersToString(r.Matchers...)

	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeaders(utils.GetLaneIDByCtx(ctx)). // 泳道特性
		SetQueryParams(query).
		SetHeaderMultiValues(header).
		Get(rawURL)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.Errorf("http code %d != 200", resp.StatusCode())
	}

	m := LabelValuesResponse{}

	if err = json.Unmarshal(resp.Body(), &m); err != nil {
		return nil, errors.Wrap(err, "unmarshal query label values response")
	}

	if m.IsSuccess() {
		return m.Data, nil
	}

	return nil, errors.Errorf("errorType: %s, error: %s", m.ErrorType, m.Error)
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
