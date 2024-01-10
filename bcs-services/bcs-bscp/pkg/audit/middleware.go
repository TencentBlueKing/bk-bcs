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
	"io"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"k8s.io/klog/v2"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
)

// Audit is a middleware that add audit.
func Audit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pattern, err := GetRouteMatcher().Match(r.Method, r.URL.Path)
		fn, ok := auditFuncMap[r.Method+"."+pattern]
		if err != nil || !ok {
			klog.Warningf("no audit for method:%s, path: %s, pattern:%s", r.Method, r.URL.Path, pattern)
			next.ServeHTTP(w, r)
			return
		}
		res, act := fn()
		st := time.Now()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		resp := bytes.NewBuffer(nil)
		ww.Tee(resp)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			render.Render(w, r, rest.BadRequest(err))
			return
		}
		// Unmarshal the request body into a map[string]interface{}
		input := make(map[string]any)
		// Get URL parameters
		params := r.URL.Query()
		for key, values := range params {
			if len(values) > 0 {
				input[key] = values[0]
			}
		}

		if len(body) != 0 {
			err = json.Unmarshal(body, &input)
			if err != nil {
				render.Render(w, r, rest.BadRequest(err))
				return
			}
		}

		defer func() {
			status := ww.Status()
			msg := "OK"
			if status != http.StatusOK {
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

			addAudit(r, res, act, st, input, status, msg)
		}()

		r.Body = io.NopCloser(bytes.NewBuffer(body))

		next.ServeHTTP(ww, r)
	})
}

const ignoredField string = "ignored"

func addAudit(r *http.Request, res audit.Resource, act audit.Action, st time.Time, input map[string]any, status int,
	msg string) {
	user := r.Header.Get(constant.UserKey)
	// if no auth for the api
	if user == "" {
		user = "no-user"
	}
	auditCtx := audit.RecorderContext{
		Username:  user,
		SourceIP:  r.RemoteAddr,
		UserAgent: r.UserAgent(),
		RequestID: r.Header.Get(constant.RidKey),
		StartTime: st,
		EndTime:   time.Now(),
	}
	resource := audit.Resource{
		ProjectCode:  ignoredField,
		ResourceID:   ignoredField,
		ResourceName: ignoredField,
		ResourceType: res.ResourceType,
		ResourceData: input,
	}
	action := audit.Action{
		ActionID:     act.ActionID,
		ActivityType: act.ActivityType,
	}

	result := audit.ActionResult{
		Status:        audit.ActivityStatusSuccess,
		ResultCode:    status,
		ResultContent: msg,
	}

	if err := GetAuditClient().R().DisableActivity().SetContext(auditCtx).SetResource(resource).SetAction(action).
		SetResult(result).Do(); err != nil {
		klog.Errorf("add audit err: %v", err)
	}
}
