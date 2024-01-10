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
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/sys"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/handler"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
)

// gateway auth server's grpc-gateway.
type gateway struct {
	iamSys *sys.Sys
	state  serviced.State
}

// newGateway create new auth server's grpc-gateway.
// nolint: unparam
func newGateway(st serviced.State, iamSys *sys.Sys) (*gateway, error) {
	g := &gateway{
		state: st,

		iamSys: iamSys,
	}

	return g, nil
}

// handler return gateway handler.
func (g *gateway) handler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/-/healthy", g.HealthyHandler)
	r.Get("/-/ready", g.ReadyHandler)
	r.Get("/healthz", g.Healthz)

	r.Mount("/", handler.RegisterCommonToolHandler())

	return r
}
