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
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql/parser"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
)

// QueryInstant 查询实时数据
func QueryInstant(ctx context.Context, projectId string, promql string, t time.Time) (model.Vector, []string, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/monitor/query/api/v1/query", config.G.BCS.Host)
	data := map[string]string{
		"query": promql,
		"time":  t.Format(time.RFC3339Nano),
	}
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetAuthToken(config.G.BCS.Token).
		SetFormData(data).
		Post(url)

	if err != nil {
		return nil, nil, err
	}

	if !resp.IsSuccess() {
		return nil, nil, errors.Errorf("http code %d != 200", resp.StatusCode())
	}

	// Decode only ResultType and load Result only as RawJson since we don't know
	// structure of the Result yet.
	var m struct {
		Data struct {
			ResultType string          `json:"resultType"`
			Result     json.RawMessage `json:"result"`
		} `json:"data"`

		Error     string `json:"error,omitempty"`
		ErrorType string `json:"errorType,omitempty"`
		// Extra field supported by Thanos Querier.
		Warnings []string `json:"warnings"`
	}

	if err = json.Unmarshal(resp.Body(), &m); err != nil {
		return nil, nil, errors.Wrap(err, "unmarshal query instant response")
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
			return nil, nil, errors.Errorf("error: %s, type: %s, warning: %s", m.Error, m.ErrorType, strings.Join(m.Warnings, ", "))
		}
		if m.Error != "" {
			return nil, nil, errors.Errorf("error: %s, type: %s", m.Error, m.ErrorType)
		}
		return nil, nil, errors.Errorf("received status code: 200, unknown response type: '%q'", m.Data.ResultType)
	}

	return vectorResult, m.Warnings, nil
}

// QueryRange 查询历史数据
func QueryRange(ctx context.Context, projectId string, promql string, start time.Time, end time.Time, step time.Duration) (model.Matrix, []string, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/monitor/query/api/v1/query_range", config.G.BCS.Host)
	data := map[string]string{
		"query": promql,
		"start": start.Format(time.RFC3339Nano),
		"end":   end.Format(time.RFC3339Nano),
		"step":  strconv.FormatInt(int64(step.Seconds()), 10) + "s",
	}
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetAuthToken(config.G.BCS.Token).
		SetFormData(data).
		Post(url)

	if err != nil {
		return nil, nil, err
	}

	// Decode only ResultType and load Result only as RawJson since we don't know
	// structure of the Result yet.
	var m struct {
		Data struct {
			ResultType string          `json:"resultType"`
			Result     json.RawMessage `json:"result"`
		} `json:"data"`

		Error     string `json:"error,omitempty"`
		ErrorType string `json:"errorType,omitempty"`
		// Extra field supported by Thanos Querier.
		Warnings []string `json:"warnings"`
	}

	if err = json.Unmarshal(resp.Body(), &m); err != nil {
		return nil, nil, errors.Wrap(err, "unmarshal query range response")
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
			return nil, nil, errors.Errorf("error: %s, type: %s, warning: %s", m.Error, m.ErrorType, strings.Join(m.Warnings, ", "))
		}
		if m.Error != "" {
			return nil, nil, errors.Errorf("error: %s, type: %s", m.Error, m.ErrorType)
		}

		return nil, nil, errors.Errorf("received status code: 200, unknown response type: '%q'", m.Data.ResultType)
	}

	return matrixResult, m.Warnings, nil

}

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
