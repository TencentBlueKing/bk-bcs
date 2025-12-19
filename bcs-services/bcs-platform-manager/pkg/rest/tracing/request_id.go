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

// Package tracing request id
package tracing

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const (
	requestIDHeaderKey = "X-Request-ID"
)

// RequestIdGenerator :
func RequestIdGenerator() string {
	uid := uuid.New().String()
	requestId := strings.ReplaceAll(uid, "-", "")
	return requestId
}

// SetRequestIDValue :
func SetRequestIDValue(req *http.Request, id string) {
	req.Header.Set(requestIDHeaderKey, id)
}

// RequestIDValue :
func RequestIDValue(req *http.Request, autoGen bool) string {
	id := req.Header.Get(requestIDHeaderKey)
	if id == "" && autoGen {
		id = RequestIdGenerator()
	}
	return id
}

// RequestIdMiddleware middleware request id
func RequestIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get id from request
		rid := r.Header.Get(requestIDHeaderKey)
		if rid == "" {
			rid = RequestIdGenerator()
		}
		// Set the id to ensure that the requestid is in the response
		w.Header().Set(requestIDHeaderKey, rid)
		next.ServeHTTP(w, r)
	})
}

// GetRequestIDResp get request id by response
func GetRequestIDResp(w http.ResponseWriter) string {
	return w.Header().Get(requestIDHeaderKey)
}
