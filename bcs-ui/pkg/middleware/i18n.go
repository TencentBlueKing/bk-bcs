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
	"strings"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/constants"
)

// DefaultLanguage middleware set default language cookie
func DefaultLanguage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie(constants.BluekingLanguage)
		if cookie != nil {
			next.ServeHTTP(w, r)
			return
		}

		cookie = &http.Cookie{
			Name:   constants.BluekingLanguage,
			Value:  config.G.Base.LanguageCode,
			Domain: config.G.Base.Domain,
			Path:   "/",
		}

		// set secure if bcs api is https schema
		if strings.HasPrefix("https://", config.G.BCS.Host) { // nolint
			cookie.Secure = true
			cookie.SameSite = http.SameSiteNoneMode
		}

		// set cookie message
		if config.G.Base.Domain != "" {
			http.SetCookie(w, cookie)
		}
		next.ServeHTTP(w, r)
	})
}
