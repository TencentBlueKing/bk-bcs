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
	"net/http"

	"github.com/go-chi/render"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/component/aiagent"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/component/notice"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/i18n"
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
	Role  string `json:"role"`
	Input string `json:"input"`
}

// Bind request
func (a *AssistantRequest) Bind(r *http.Request) error {
	return nil
}

// Assistant ai assistant
func (s *WebServer) Assistant(w http.ResponseWriter, r *http.Request) {
	okResponse := &OKResponse{
		Message:   "success",
		RequestID: r.Header.Get("x-request-id"),
	}
	data := &AssistantRequest{}
	if err := render.Bind(r, data); err != nil {
		okResponse.Code = http.StatusBadRequest
		okResponse.Message = err.Error()
		render.JSON(w, r, okResponse)
		return
	}

	// 用户登录鉴权
	bk_ticket := GetBKTicketByRequest(r)
	if bk_ticket == "" {
		okResponse.Code = 401
		okResponse.Message = "user is invalid"
		render.JSON(w, r, okResponse)
		return
	}

	out, err := aiagent.Assistant(r.Context(), GetBKTicketByRequest(r), data.Role, data.Input)
	if err != nil {
		okResponse.Code = http.StatusBadRequest
		okResponse.Message = err.Error()
		render.JSON(w, r, okResponse)
		return
	}
	okResponse.Data = map[string]interface{}{
		"output": out,
	}
	render.JSON(w, r, okResponse)
}
