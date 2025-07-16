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

package middleware

import (
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n"
	restful "github.com/emicklei/go-restful/v3"
)

const defaultLang = "zh-cn"

// LanguageFilter set language to context
func LanguageFilter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	lang := getLangByRequest(request.Request)
	ctx := request.Request.Context()
	ctx = i18n.WithLanguage(ctx, lang)
	request.Request = request.Request.WithContext(ctx)
	chain.ProcessFilter(request, response)
}

// getLangByRequest get the language by request
func getLangByRequest(r *http.Request) string {

	// lang参数 -> cookie -> accept-language -> 默认
	lng := r.FormValue("lang")
	if lng != "" {
		return lng
	}

	cookie, err := r.Cookie("blueking_language")
	if err == nil && cookie != nil {
		return cookie.Value
	}

	acceptLanguage := r.Header.Get("accept-language")
	if acceptLanguage != "" {
		return acceptLanguage
	}

	// default config
	return defaultLang
}
