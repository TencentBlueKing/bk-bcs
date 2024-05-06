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
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/analyze"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/analyze/external"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/bkm"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/collect"
)

// AnalysisManager defines the manager of analysis
type AnalysisManager struct {
	op         *options.AnalysisOptions
	httpServer *http.Server

	alysisHandler         analyze.AnalysisInterface
	alysisHandlerExternal analyze.AnalysisInterface
	alysisCollector       *collect.AnalysisCollect
	metricCollector       *collect.MetricCollect
	bkmClient             *bkm.BKMonitorClient
}

// NewAnalysisManager creat the analysis manager instance
func NewAnalysisManager() *AnalysisManager {
	return &AnalysisManager{
		op:        options.GlobalOptions(),
		bkmClient: bkm.NewBKMonitorClient(),
	}
}

// Init init the analysis manager
func (m *AnalysisManager) Init() error {
	db, err := dao.NewDriver()
	if err != nil {
		return errors.Wrapf(err, "create db driver failed")
	}
	if err = db.Init(); err != nil {
		return errors.Wrapf(err, "INIT DB DRIVER FAILED")
	}
	m.alysisHandler = analyze.NewAnalysisHandler()
	if err := m.alysisHandler.Init(); err != nil {
		return errors.Wrapf(err, "init analysis handler failed")
	}
	if m.op.ExternalAnalysisUrl != "" && m.op.ExternalAnalysisToken != "" {
		m.alysisHandlerExternal = external.NewExternalAnalysisHandler()
	}
	m.alysisCollector = collect.NewAnalysisCollector()
	m.metricCollector = collect.NewMetricCollect()
	if err := m.metricCollector.Init(); err != nil {
		return errors.Wrapf(err, "init metric collector failed")
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
	subRouter := router.PathPrefix("/api/v1/analysis").Subrouter()
	subRouter.Path("").Methods(http.MethodGet).HandlerFunc(m.Analysis)
	subRouter.Path("/overview").Methods(http.MethodGet).HandlerFunc(m.Overview)
	subRouter.Path("/overview/compare").Methods(http.MethodGet).HandlerFunc(m.OverviewCompare)
	subRouter.Path("/top_projects").Methods(http.MethodGet).HandlerFunc(m.TopProjects)
	subRouter.Path("/managed_resources").Methods(http.MethodGet).HandlerFunc(m.ManagedResources)

	subRouter.Path("/bkmonitor/common").Methods(http.MethodPost).HandlerFunc(m.BKMCommon)
	subRouter.Path("/bkmonitor/activity_projects").Methods(http.MethodGet).Queries("target", "{target}").
		HandlerFunc(m.BKMActivityProjects)
	subRouter.Path("/bkmonitor/slo").Methods(http.MethodPost).HandlerFunc(m.BKMQuerySLO)
	subRouter.Path("/bkmonitor/slo_unavailable").Methods(http.MethodPost).HandlerFunc(m.BKMQuerySLOUnavailable)

	subRouter.Path("/query/projects").Methods(http.MethodGet).HandlerFunc(m.QueryProjects)
	subRouter.Path("/query/applications").Methods(http.MethodGet).HandlerFunc(m.QueryApplications)
	m.httpServer = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", m.op.Port),
		Handler: router,
	}
}

// Start the analysis manager
func (m *AnalysisManager) Start(ctx context.Context) error {
	go m.alysisCollector.Start(ctx)
	go m.metricCollector.Start(ctx)
	blog.Infof("http server is listening on %s", m.httpServer.Addr)
	if err := m.httpServer.ListenAndServe(); err != nil {
		blog.Errorf("analysis http server listen failed: %s", err.Error())
	}
	blog.Infof("analysis http server finished")
	return nil
}
