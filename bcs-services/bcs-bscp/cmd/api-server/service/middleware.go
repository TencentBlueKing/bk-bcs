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
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/metrics"
	pbcs "bscp.io/pkg/protocol/config-server"
	"bscp.io/pkg/rest"
)

// CheckDefaultTmplSpace create default template space if not existent
func (p *proxy) CheckDefaultTmplSpace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bizIDStr := chi.URLParam(r, "biz_id")
		bizIDInt, err := strconv.Atoi(bizIDStr)
		if err != nil {
			render.Render(w, r, rest.BadRequest(err))
			return
		}
		bizID := uint32(bizIDInt)
		if bizsOfTS.Has(bizID) {
			next.ServeHTTP(w, r)
			return
		}

		// use system user to create default template space
		kt := kit.MustGetKit(r.Context())
		kt.User = constant.BKSystemUser

		// create default template space when not existent
		in := &pbcs.CreateDefaultTmplSpaceReq{BizId: bizID}
		if _, err := p.cfgClient.CreateDefaultTmplSpace(kt.RpcCtx(), in); err != nil {
			render.Render(w, r, rest.BadRequest(err))
			return
		}
		bizsOfTS.Set(bizID)

		next.ServeHTTP(w, r)
	})
}

// HttpServerHandledTotal count http operands
func (p *proxy) HttpServerHandledTotal(serviceName, handler string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			kt := kit.MustGetKit(r.Context())
			var status string
			if serviceName == "" {
				serviceName = chi.RouteContext(r.Context()).RoutePattern()
			}
			if handler == "" {
				handler = r.URL.String()
			}
			defer func() {
				metrics.BSCPServerHandledTotal.
					WithLabelValues(serviceName, handler, status, strconv.Itoa(int(kt.BizID)), kt.User).
					Inc()
			}()
			next.ServeHTTP(w, r)
			status = strconv.Itoa(w.(interface {
				http.ResponseWriter
				Status() int
			}).Status())
		}
		return http.HandlerFunc(fn)
	}
}
