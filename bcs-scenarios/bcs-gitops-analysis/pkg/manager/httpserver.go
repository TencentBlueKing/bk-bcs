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
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/analyze"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/bkm"
)

func (m *AnalysisManager) httpError(rw http.ResponseWriter, statusCode int, err error) {
	http.Error(rw, err.Error(), statusCode)
}

func (m *AnalysisManager) httpJson(rw http.ResponseWriter, obj interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	content, _ := json.Marshal(obj)
	rw.Write(content)
}

// AnalysisResponse defines the response of analysis data
type AnalysisResponse struct {
	Code      int32       `json:"code"`
	Message   string      `json:"message"`
	RequestID string      `json:"requestID"`
	Data      interface{} `json:"data"`
}

// Analysis 按照查询参数 projects 查找项目运营数据
func (m *AnalysisManager) Analysis(writer http.ResponseWriter, req *http.Request) {
	projects := req.URL.Query()["projects"]
	if len(projects) == 0 {
		m.httpJson(writer, &AnalysisResponse{
			Code: 0,
			Data: m.alysisHandler.AnalysisProjectsAll(),
		})
		return
	}
	argoProjects, err := m.alysisHandler.QueryArgoProjects(projects)
	if err != nil {
		m.httpError(writer, http.StatusInternalServerError, err)
		return
	}
	analysis, err := m.alysisHandler.AnalysisProject(req.Context(), argoProjects)
	if err != nil {
		m.httpError(writer, http.StatusInternalServerError, err)
		return
	}
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: analysis,
	})
}

// Overview 运营总览数据
func (m *AnalysisManager) Overview(writer http.ResponseWriter, request *http.Request) {
	overview, _ := m.alysisHandler.AnalysisOverview()
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: []*analyze.AnalysisOverviewAll{overview},
	})
}

// OverviewCompare 运营总览数据对比
func (m *AnalysisManager) OverviewCompare(writer http.ResponseWriter, request *http.Request) {
	if m.alysisHandlerExternal == nil {
		m.httpError(writer, http.StatusNotFound, errors.Errorf("not found"))
		return
	}
	internal, _ := m.alysisHandler.AnalysisOverview()
	external, err := m.alysisHandlerExternal.AnalysisOverview()
	if err != nil {
		m.httpError(writer, http.StatusInternalServerError, errors.Wrapf(err, "query analysis external failed"))
		return
	}
	internal.Type = "国内"
	external.Type = "海外"
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: []*analyze.AnalysisOverviewAll{internal, external},
	})
}

// TopProjects 排行靠前项目
func (m *AnalysisManager) TopProjects(writer http.ResponseWriter, request *http.Request) {
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: m.alysisHandler.TopProjects(),
	})
}

// ManagedResources 管理的资源信息
func (m *AnalysisManager) ManagedResources(writer http.ResponseWriter, request *http.Request) {
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: m.alysisHandler.ResourceInfosAll(),
	})
}

// BKMCommon 监控常规查询
func (m *AnalysisManager) BKMCommon(writer http.ResponseWriter, request *http.Request) {
	bkmMessage, err := m.buildBKMRequest(request)
	if err != nil {
		m.httpError(writer, http.StatusInternalServerError, errors.Wrapf(err, "build bkmonitor request failed"))
		return
	}
	series, err := m.bkmCommon(bkmMessage)
	if err != nil {
		m.httpError(writer, http.StatusInternalServerError, err)
		return
	}
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: series,
	})
}

// BKMActivityProjects 活跃项目
func (m *AnalysisManager) BKMActivityProjects(writer http.ResponseWriter, request *http.Request) {
	target := request.URL.Query().Get("target")
	if target != "internal" && target != "external" {
		target = "internal"
	}
	message := &bkm.BKMonitorGetMessage{
		// nolint
		PromQL: fmt.Sprintf(`topk(10, max by (project) (increase(custom:GitOpsOperationData:project_sync{target="%s"}[1440m])) != 0)`, target),
		Step:   "3600s",
	}
	projects, err := m.calculateBKMonitorActivityProjects(message)
	if err != nil {
		m.httpError(writer, http.StatusInternalServerError, err)
		return
	}
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: projects,
	})
}

// BKMQuerySLO query the slo
func (m *AnalysisManager) BKMQuerySLO(writer http.ResponseWriter, request *http.Request) {
	tid := request.URL.Query().Get("task")
	taskID, err := strconv.Atoi(tid)
	if err != nil {
		m.httpError(writer, http.StatusBadRequest, errors.Wrapf(err, "convert task id '%s' failed", tid))
		return
	}
	start, _ := time.Parse("2006-01-02 15:04:05", m.op.BKMonitorSLOStartDay+" 00:00:00")
	end := time.Now()
	bkmMessage := &bkm.BKMonitorGetMessage{
		PromQL: fmt.Sprintf(`count_over_time(bkmonitor:uptimecheck:http:available{task_id="%d"}[1m]) != 1`, taskID),
		Start:  fmt.Sprintf("%d", start.Unix()),
		End:    fmt.Sprintf("%d", end.Unix()),
		Step:   "60s",
	}
	result, err := m.bkmCommonAll(bkmMessage)
	if err != nil {
		m.httpError(writer, http.StatusBadRequest, errors.Wrapf(err, "bkmonitor query failed"))
		return
	}
	all := end.Sub(start).Minutes()
	unavailable := len(result)
	slo := (float64(unavailable) * 1.0) / (float64(all) * 1.0) * 100
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: []*bkm.BKMonitorSeriesItem{
			{
				Timestamp: end.Add(8 * time.Hour).UnixNano(),
				Value:     100.0 - slo,
			},
		},
	})
}

// BKMQuerySLOUnavailable query the slo unavailable
func (m *AnalysisManager) BKMQuerySLOUnavailable(writer http.ResponseWriter, request *http.Request) {
	tid := request.URL.Query().Get("task")
	taskID, err := strconv.Atoi(tid)
	if err != nil {
		m.httpError(writer, http.StatusBadRequest, errors.Wrapf(err, "convert task id '%s' failed", tid))
		return
	}
	start, _ := time.Parse("2006-01-02 15:04:05", m.op.BKMonitorSLOStartDay+" 00:00:00")
	end := time.Now()
	thirtyDayAgo := end.Add(-720 * time.Hour)
	if thirtyDayAgo.After(start) {
		start = thirtyDayAgo
	}
	bkmMessage := &bkm.BKMonitorGetMessage{
		PromQL: fmt.Sprintf(`count_over_time(bkmonitor:uptimecheck:http:available{task_id="%d"}[1m]) != 1`, taskID),
		Start:  fmt.Sprintf("%d", start.Unix()),
		End:    fmt.Sprintf("%d", end.Unix()),
		Step:   "60s",
	}
	resp, err := m.bkmCommonAll(bkmMessage)
	if err != nil {
		m.httpError(writer, http.StatusBadRequest, errors.Wrapf(err, "bkmonitor query failed"))
		return
	}
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: []*bkm.BKMonitorSeriesItem{
			{
				Timestamp: start.Add(8 * time.Hour).UnixNano(),
				Value:     0,
			},
			{
				Timestamp: end.Add(8 * time.Hour).UnixNano(),
				Value:     float64(len(resp)),
			},
		},
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

// BKMActivityProject defines the project message
type BKMActivityProject struct {
	Project string `json:"project"`
	Value   int64  `json:"value"`
}

func (m *AnalysisManager) calculateBKMonitorActivityProjects(message *bkm.BKMonitorGetMessage) (
	[]*BKMActivityProject, error) {
	message.Complete()
	resp, err := m.bkmClient.Get(message)
	if err != nil {
		return nil, errors.Wrapf(err, "bkmonitor get failed with promql '%s'", message.PromQL)
	}
	if len(resp.Series) == 0 {
		return nil, nil
	}
	var lastTime int64
	result := make([]*BKMActivityProject, 0)
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
		result = append(result, &BKMActivityProject{
			Project: series.GroupValues[0],
			Value:   int64(series.Values[len(series.Values)-1][1]),
		})
	}
	return result, nil
}

// QueryProjects query projects
func (m *AnalysisManager) QueryProjects(writer http.ResponseWriter, request *http.Request) {
	projects := m.alysisHandler.AnalysisProjectsAll()
	for _, proj := range projects {
		proj.ClusterNum = len(proj.Clusters)
		proj.ApplicationNum = len(proj.Applications)
		proj.SecretNum = len(proj.Secrets)
		proj.RepoNum = len(proj.Repos)
		for _, user := range proj.ActivityUsers {
			if user.LastActivityTime.After(time.Now().Add(-168 * time.Hour)) {
				proj.ActivityUserNum++
			}
		}
		for _, sync := range proj.Syncs {
			proj.SyncTotal += sync.SyncTotal
		}
		proj.BizName = m.alysisHandler.GetBusinessName(int(proj.BizID))
	}
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: projects,
	})
}

// QueryApplications query applications
func (m *AnalysisManager) QueryApplications(writer http.ResponseWriter, request *http.Request) {
	result, err := m.alysisHandler.Applications()
	if err != nil {
		m.httpError(writer, http.StatusInternalServerError, err)
		return
	}
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: result,
	})
}
