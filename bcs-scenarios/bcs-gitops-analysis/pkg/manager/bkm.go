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

package manager

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/bkm"
)

// AnalysisResponse defines the response of analysis data
type AnalysisResponse struct {
	Code      int32       `json:"code"`
	Message   string      `json:"message"`
	RequestID string      `json:"requestID"`
	Data      interface{} `json:"data"`
}

// BKMCommon common query the bkmonitor, and trans data to normal json data
func (m *AnalysisManager) BKMCommon(writer http.ResponseWriter, request *http.Request) {
	bkmMessage, err := m.buildBKMRequest(request)
	if err != nil {
		m.httpError(writer, http.StatusInternalServerError, errors.Wrapf(err, "build bkmonitor request failed"))
		return
	}
	if len(bkmMessage.PromQLs) == 0 {
		series, err := m.bkmCommon(bkmMessage)
		if err != nil {
			m.httpError(writer, http.StatusInternalServerError, err)
			return
		}
		m.httpJson(writer, &AnalysisResponse{
			Code: 0,
			Data: series,
		})
		return
	}

	result := make([]*bkm.BKMonitorSeriesItem, 0)
	for name, promQL := range bkmMessage.PromQLs {
		bkmMessage.PromQL = promQL
		series, err := m.bkmCommon(bkmMessage)
		if err != nil {
			m.httpError(writer, http.StatusInternalServerError, err)
			return
		}
		for _, item := range series {
			item.Name = name
			result = append(result, item)
		}
	}
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: result,
	})
	return
}

// BKMTop top query with bkmonitor with top promQL.
func (m *AnalysisManager) BKMTop(writer http.ResponseWriter, request *http.Request) {
	bkmMessage, err := m.buildBKMRequest(request)
	if err != nil {
		m.httpError(writer, http.StatusInternalServerError, errors.Wrapf(err, "build bkmonitor request failed"))
		return
	}
	result, err := m.calculateBKMTop(bkmMessage)
	if err != nil {
		m.httpError(writer, http.StatusInternalServerError, err)
		return
	}
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: result,
	})
}

// bkmCommon bkmonitor common query
func (m *AnalysisManager) bkmCommon(message *bkm.BKMonitorGetMessage) ([]*bkm.BKMonitorSeriesItem, error) {
	result, err := m.bkmCommonAll(message)
	if err != nil {
		return nil, errors.Wrapf(err, "bkmonitor query failed")
	}
	if len(result) != 1 {
		return nil, errors.Errorf("bkmonitor return series not 1 but '%d'", len(result))
	}
	return result[0], nil
}

func (m *AnalysisManager) bkmCommonAll(message *bkm.BKMonitorGetMessage) ([][]*bkm.BKMonitorSeriesItem, error) {
	resp, err := m.bkmClient.Get(message)
	if err != nil {
		return nil, errors.Wrapf(err, "bkmonitor get failed")
	}
	result := make([][]*bkm.BKMonitorSeriesItem, 0, len(resp.Series))
	for i := range resp.Series {
		seriesItem := &(resp.Series[i])
		resultItem := make([]*bkm.BKMonitorSeriesItem, 0, len(seriesItem.Values))
		for j := range seriesItem.Values {
			v := seriesItem.Values[j]
			if len(v) != 2 {
				return nil, errors.Errorf("bkmonitor return series values not 2 but '%d': %v", len(v), v)
			}
			resultItem = append(resultItem, &bkm.BKMonitorSeriesItem{
				Timestamp: time.UnixMilli(int64(v[0])).Add(8 * time.Hour).UnixNano(),
				Value:     v[1],
			})
		}
		result = append(result, resultItem)
	}
	return result, nil
}

func (m *AnalysisManager) buildBKMRequest(r *http.Request) (*bkm.BKMonitorGetMessage, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read body failed")
	}
	req := &bkm.BKMonitorGetMessage{}
	if err = json.Unmarshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "unmarshal request body failed")
	}
	return req, nil
}

// BKMTopObject defines the bkmonitor top object
type BKMTopObject struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

// calculateBKMTop will change the data format from bkmonitor platform
func (m *AnalysisManager) calculateBKMTop(message *bkm.BKMonitorGetMessage) ([]*BKMTopObject, error) {
	message.Complete()
	resp, err := m.bkmClient.Get(message)
	if err != nil {
		return nil, errors.Wrapf(err, "bkmonitor get failed with promql '%s'", message.PromQL)
	}
	if len(resp.Series) == 0 {
		return nil, nil
	}
	var lastTime int64
	result := make([]*BKMTopObject, 0)
	for i := range resp.Series {
		series := resp.Series[i]
		if len(series.GroupValues) != 1 {
			return nil, errors.Errorf("bkmonitor top projects '%d' group values length not 1: %v",
				i, series.GroupValues)
		}
		if len(series.Values) == 0 {
			continue
		}
		if len(series.Values[len(series.Values)-1]) != 2 {
			return nil, errors.Errorf("bkmonitor top projects '%d' values last item not 2: %v",
				i, series.Values[len(series.Values)-1])
		}
		timeStamp := int64(series.Values[len(series.Values)-1][0])
		if i == 0 {
			lastTime = timeStamp
		}
		if timeStamp != lastTime {
			continue
		}
		result = append(result, &BKMTopObject{
			Name:  series.GroupValues[0],
			Value: int64(series.Values[len(series.Values)-1][1]),
		})
	}
	return result, nil
}
