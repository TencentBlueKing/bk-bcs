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

package service

import (
	"net/http"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/shutdown"
)

// HealthyHandler livenessProbe 健康检查
func (g *gateway) HealthyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// ReadyHandler ReadinessProbe 健康检查
func (g *gateway) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	g.Healthz(w, r)
}

// Healthz service health check.
func (g *gateway) Healthz(w http.ResponseWriter, r *http.Request) {
	if shutdown.IsShuttingDown() {
		logs.Errorf("service healthz check failed, current service is shutting down")
		w.WriteHeader(http.StatusServiceUnavailable)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealth, "current service is shutting down"))
		return
	}

	if err := g.state.Healthz(); err != nil {
		logs.Errorf("etcd healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealth, "etcd healthz error, "+err.Error()))
		return
	}

	if err := g.dao.Healthz(); err != nil {
		logs.Errorf("mysql healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealth, "mysql healthz error, "+err.Error()))
		return
	}

	if err := g.bs.Healthz(); err != nil {
		logs.Errorf("redis healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealth, "redis healthz error, "+err.Error()))
		return
	}

	rest.WriteResp(w, rest.NewBaseResp(errf.OK, "healthy"))
}
