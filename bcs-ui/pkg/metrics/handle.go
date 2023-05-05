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

package metrics

import (
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"strconv"
	"time"
)

// RequestCollect 统计请求计数、耗时
func RequestCollect(handler string, hr http.HandlerFunc) func(w http.ResponseWriter,
	r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		hr(w, r)
		statusCode := 200
		if value, ok := w.(middleware.WrapResponseWriter); ok {
			statusCode = value.Status()
		}
		code := strconv.Itoa(statusCode)
		requestDuration := time.Since(start)
		collectHTTPRequestMetric(handler, r.Method, code, requestDuration)
	}
}
