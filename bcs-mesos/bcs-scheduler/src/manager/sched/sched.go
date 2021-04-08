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
 *
 */

package sched

import (
	"net/http"
	"net/http/pprof"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/api"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/backend"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/scheduler"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/schedcontext"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/util"

	restful "github.com/emicklei/go-restful"
)

type Sched struct {
	config    util.Scheduler
	scheduler *scheduler.Scheduler
	scontext  *schedcontext.SchedContext
}

func New(config util.Scheduler, scontext *schedcontext.SchedContext) *Sched {
	s := &Sched{
		config:    config,
		scontext:  scontext,
		scheduler: scheduler.NewScheduler(config, scontext.Store, scontext.AlertManager),
	}

	backend := backend.NewBackend(s.scheduler, s.scontext.Store)

	r := api.NewRouter(backend)
	apiActions := r.GetActions()
	s.scontext.ApiServer2.RegisterWebServer("/v1", nil, apiActions)
	//use pprof
	if s.config.DebugMode {
		s.initDebug()
	}

	return s
}

func (s *Sched) Start() error {
	if err := s.scheduler.Start(); err != nil {
		return err
	}

	return nil
}
func (s *Sched) initDebug() {
	blog.Infof("init debug pprof")
	action := []*httpserver.Action{
		httpserver.NewAction("GET", "/debug/pprof/", nil, getRouteFunc(pprof.Index)),
		httpserver.NewAction("GET", "/debug/pprof/{uri:*}", nil, getRouteFunc(pprof.Index)),
		httpserver.NewAction("GET", "/debug/pprof/cmdline", nil, getRouteFunc(pprof.Cmdline)),
		httpserver.NewAction("GET", "/debug/pprof/profile", nil, getRouteFunc(pprof.Profile)),
		httpserver.NewAction("GET", "/debug/pprof/symbol", nil, getRouteFunc(pprof.Symbol)),
		httpserver.NewAction("GET", "/debug/pprof/trace", nil, getRouteFunc(pprof.Trace)),
	}
	s.scontext.ApiServer2.RegisterWebServer("", nil, action)
}

func getRouteFunc(f http.HandlerFunc) restful.RouteFunction {
	return restful.RouteFunction(func(req *restful.Request, resp *restful.Response) {
		f(resp, req.Request)
	})
}
