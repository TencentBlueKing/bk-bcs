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

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/component/notice"
)

// GetCurrentAnnouncements 获取当前公告
func (s *WebServer) GetCurrentAnnouncements(w http.ResponseWriter, r *http.Request) {
	announcements, err := notice.GetCurrentAnnouncements(r.Context())
	if err != nil {
		klog.Warningf("get current announcements failed: %s", err.Error())
	}
	okResponse := &OKResponse{
		Message:   "success",
		Data:      announcements,
		RequestID: r.Header.Get("x-request-id")}
	render.JSON(w, r, okResponse)
}
