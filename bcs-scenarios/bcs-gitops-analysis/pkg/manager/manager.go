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

// Package manager xx
package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	traceconst "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/constants"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/bkm"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/collect"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/manager/handler"
)

// AnalysisManager defines the manager of analysis
type AnalysisManager struct {
	op         *options.AnalysisOptions
	httpServer *http.Server

	metricCollector *collect.MetricCollect
	bkmClient       *bkm.BKMonitorClient

	internalAnalysisHandler handler.AnalysisInterface
	externalAnalysisHandler handler.AnalysisInterface
}

// NewAnalysisManager creat the analysis manager instance
func NewAnalysisManager() *AnalysisManager {
	return &AnalysisManager{
		op:                      options.GlobalOptions(),
		bkmClient:               bkm.NewBKMonitorClient(),
		internalAnalysisHandler: handler.NewAnalysisHandler(),
		externalAnalysisHandler: handler.NewAnalysisExternalHandler(),
	}
}

func (m *AnalysisManager) returnAnalysisHandler(request *http.Request) handler.AnalysisInterface {
	target := request.URL.Query().Get("target")
	if target == "gitops-external" {
		return m.externalAnalysisHandler
	}
	return m.internalAnalysisHandler
}

func (m *AnalysisManager) returnAnalysisTarget(request *http.Request) string {
	target := request.URL.Query().Get("target")
	if target == "external" {
		return externalTarget
	}
	return internalTarget
}

// Init the analysis manager
func (m *AnalysisManager) Init() error {
	db, err := dao.NewDriver()
	if err != nil {
		return errors.Wrapf(err, "create db driver failed")
	}
	if err = db.Init(); err != nil {
		return errors.Wrapf(err, "INIT DB DRIVER FAILED")
	}
	m.metricCollector = collect.NewMetricCollect()
	if err = m.metricCollector.Init(); err != nil {
		return errors.Wrapf(err, "init metric collector failed")
	}
	if err = m.internalAnalysisHandler.Init(); err != nil {
		return errors.Wrapf(err, "init analysis handler failed")
	}
	if err = m.externalAnalysisHandler.Init(); err != nil {
		return errors.Wrapf(err, "init external analysis handler failed")
	}
	m.initHTTPServer()
	return nil
}

func (m *AnalysisManager) initHTTPServer() {
	router := mux.NewRouter()
	router.UseEncodedPath()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(traceconst.RequestIDHeaderKey)
			if requestID == "" {
				requestID = uuid.New().String()
			}
			start := time.Now()
			blog.Infof("RequestID[%s] received request: %s, %s", requestID, r.Method, r.URL.String())
			next.ServeHTTP(w, r)
			elapsed := time.Since(start)
			blog.Infof("RequestID[%s] request cost: %v", requestID, elapsed)
		})
	})
	newSubRouter := router.PathPrefix("/api/v1/analysis_new").Subrouter()
	newSubRouter.Path("/raw_data").HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			result := m.returnAnalysisHandler(request).GetAnalysisProjects()
			m.httpJson(writer, result)
		})
	newSubRouter.Path("/overview").HandlerFunc(m.OverviewNew)
	newSubRouter.Path("/managed_resources").HandlerFunc(m.ManagedResourceNew)
	newSubRouter.Path("/projects/overview").HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			m.httpJson(writer, &AnalysisResponse{
				Code: 0,
				Data: []*AnalysisProjectOverview{m.handleProjectOverview(request)},
			})
		})
	newSubRouter.Path("/projects/group").HandlerFunc(m.ProjectGroups)
	newSubRouter.Path("/projects/rank").HandlerFunc(m.ProjectRank)
	newSubRouter.Path("/users/overview").HandlerFunc(m.UserOverview)
	newSubRouter.Path("/users/group/dept_outer").HandlerFunc(m.UserGroupDeptOuter)
	newSubRouter.Path("/users/group/dept_inner").HandlerFunc(m.UserGroupDeptInner)
	newSubRouter.Path("/users/group/dept_operate").HandlerFunc(m.UserGroupDeptOperate)
	newSubRouter.Path("/bkmonitor/common").Methods(http.MethodPost).HandlerFunc(m.BKMCommon)
	newSubRouter.Path("/bkmonitor/top").Methods(http.MethodPost).HandlerFunc(m.BKMTop)

	queryRouter := newSubRouter.PathPrefix("/query").Subrouter()
	queryRouter.Path("/projects").HandlerFunc(m.QueryProjects)
	queryRouter.Path("/users").HandlerFunc(m.QueryUsers)
	queryRouter.Path("/applications").HandlerFunc(m.QueryApplications)
	m.httpServer = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", m.op.Port),
		Handler: router,
	}
}

// Start the analysis manager
func (m *AnalysisManager) Start(ctx context.Context) error {
	go m.metricCollector.Start(ctx)
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		blog.Infof("analysis collector started")
		defer ticker.Stop()
		defer blog.Infof("analysis collection finished")
		for {
			select {
			case <-ticker.C:
				m.collectAnalysis()
			case <-ctx.Done():
				return
			}
		}
	}()
	blog.Infof("http server is listening on %s", m.httpServer.Addr)
	if err := m.httpServer.ListenAndServe(); err != nil {
		blog.Errorf("analysis http server listen failed: %s", err.Error())
	}
	blog.Infof("analysis http server finished")
	return nil
}

func (m *AnalysisManager) httpError(rw http.ResponseWriter, statusCode int, err error) {
	http.Error(rw, err.Error(), statusCode)
}

func (m *AnalysisManager) httpJson(rw http.ResponseWriter, obj interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	content, _ := json.Marshal(obj)
	rw.Write(content)
}
