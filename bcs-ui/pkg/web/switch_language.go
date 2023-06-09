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

package web

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/constants"
)

// ErrorRsp error response
type ErrorRsp struct {
	ErrorMessage string `json:"error_message"`
}

// CookieSwitchLanguage switch cookie language
func (s *WebServer) CookieSwitchLanguage(w http.ResponseWriter, r *http.Request) {
	okResponse := &OKResponse{Message: "OK", RequestID: r.Header.Get(constants.RequestIDHeaderKey)}
	// response message
	defer render.JSON(w, r, okResponse)
	// get cookie key blueking_language
	cookie, err := r.Cookie(constants.BluekingLanguage)
	if err != nil {
		// failure return
		okResponse.Code = http.StatusBadRequest
		okResponse.Message = err.Error()
		return
	}
	// get language
	lang := cookie.Value

	// set cookie message
	if config.G.Base.Domain != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     constants.BluekingLanguage,
			Value:    lang,
			Domain:   config.G.Base.Domain,
			Path:     "/",
			SameSite: http.SameSiteNoneMode,
		})
	}
}
