/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"net/http"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/rest"
	"bscp.io/pkg/runtime/handler"
	"bscp.io/pkg/runtime/shutdown"
)

// setupFilter setups all api filters here. All request would cross here, and we filter request base on URL.
func (g *gateway) setupFilter(mux *http.ServeMux) http.Handler {
	var httpHandler http.Handler

	httpHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// healthz.
		if r.RequestURI == "/healthz" {
			g.Healthz(w)
			return
		}

		// handle request.
		mux.ServeHTTP(w, r)
	})

	// add common handler
	httpHandler = handler.HTTPMiddleware(httpHandler)

	return httpHandler
}

// Healthz service health check.
func (g *gateway) Healthz(w http.ResponseWriter) {
	if shutdown.IsShuttingDown() {
		logs.Errorf("service healthz check failed, current service is shutting down")
		w.WriteHeader(http.StatusServiceUnavailable)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealth, "current service is shutting down"))
		return
	}

	if err := g.state.Healthz(cc.DataService().Service.Etcd); err != nil {
		logs.Errorf("etcd healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealth, "etcd healthz error, "+err.Error()))
		return
	}

	if err := g.dao.Healthz(); err != nil {
		logs.Errorf("mysql healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealth, "mysql healthz error, "+err.Error()))
		return
	}

	rest.WriteResp(w, rest.NewBaseResp(errf.OK, "healthy"))
	return
}
