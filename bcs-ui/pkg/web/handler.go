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

// Package web xxx
package web

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/component/aiagent"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/component/notice"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/middleware"
)

// GetCurrentAnnouncements 获取当前公告
func (s *WebServer) GetCurrentAnnouncements(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLangByRequest(r, config.G.Base.LanguageCode)
	announcements, err := notice.GetCurrentAnnouncements(r.Context(), lang)
	if err != nil {
		klog.Warningf("get current announcements failed: %s", err.Error())
	}
	okResponse := &OKResponse{
		Message:   "success",
		Data:      announcements,
		RequestID: r.Header.Get("x-request-id")}
	render.JSON(w, r, okResponse)
}

// AssistantRequest assistant request
type AssistantRequest struct {
	Role   string `json:"role"`
	Input  string `json:"input"`
	Stream bool   `json:"stream"`
}

// Bind request
func (a *AssistantRequest) Bind(r *http.Request) error {
	return nil
}

// Assistant ai assistant
func (s *WebServer) Assistant(w http.ResponseWriter, r *http.Request) {
	resp := &OKResponse{
		Message:   "success",
		RequestID: r.Header.Get("x-request-id"),
	}
	req := &AssistantRequest{}
	if err := render.Bind(r, req); err != nil {
		resp.Code = http.StatusBadRequest
		resp.Message = err.Error()
		render.JSON(w, r, resp)
		return
	}

	// 用户登录鉴权
	bk_ticket := middleware.MustGetBKTicketFromContext(r.Context())
	user := middleware.MustGetUserFromContext(r.Context())

	out, err := aiagent.Assistant(r.Context(), bk_ticket, req.Role, req.Input, user.UserName, req.Stream)
	if err != nil {
		resp.Code = http.StatusBadRequest
		resp.Message = err.Error()
		render.JSON(w, r, resp)
		return
	}

	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Transfer-Encoding", "chunked")
		buf := make([]byte, 4096)
		data := out.(io.ReadCloser)
		defer data.Close()
		for {
			n, err := data.Read(buf)
			if n > 0 {
				_, writeErr := w.Write(buf[:n])
				if writeErr != nil {
					http.Error(w, "Error writing response", http.StatusInternalServerError)
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					http.Error(w, fmt.Sprintf("Error reading from API, err %s", err), http.StatusInternalServerError)
					return
				}
				break
			}
		}
		// r = r.WithContext(context.WithValue(r.Context(), render.ContentTypeCtxKey, render.ContentTypeEventStream))
		// render.DefaultResponder(w, r, out)
		return
	}
	resp.Data = map[string]interface{}{
		"output": out,
	}
	render.JSON(w, r, resp)
}
