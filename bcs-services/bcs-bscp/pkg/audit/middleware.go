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

package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
)

// Audit is a http middleware that add audit.
func Audit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pattern, err := GetRouteMatcher().Match(r.Method, r.URL.Path)
		fn, ok := auditHttpMap[r.Method+"."+pattern]
		if err != nil || !ok {
			// klog.Warningf("no audit for method:%s, path: %s, pattern:%s", r.Method, r.URL.Path, pattern)
			next.ServeHTTP(w, r)
			return
		}
		res, act := fn()
		st := time.Now()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		resp := bytes.NewBuffer(nil)
		ww.Tee(resp)

		defer func() {
			status := ww.Status()
			msg := "Success"
			if status >= http.StatusBadRequest {
				rs := struct {
					Error struct {
						Message string `json:"message"`
					} `json:"error"`
				}{}
				err = json.Unmarshal(resp.Bytes(), &rs)
				if err != nil {
					render.Render(w, r, rest.BadRequest(err))
					return
				}
				msg = rs.Error.Message
			}

			user := r.Header.Get(constant.UserKey)
			// if no auth for the api
			if user == "" {
				user = "no-user"
			}

			p := auditParam{
				Username:     user,
				SourceIP:     r.RemoteAddr,
				UserAgent:    r.UserAgent(),
				Rid:          r.Header.Get(constant.RidKey),
				Resource:     res,
				Action:       act,
				StartTime:    st,
				ResultStatus: status,
				ResultMsg:    msg,
			}
			addAudit(p)
		}()

		next.ServeHTTP(ww, r)
	})
}
