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

package utils

import (
	"net/http"
)

// ResponseWriterWrapper is a custom http.ResponseWriter that captures the status code.
type ResponseWriterWrapper struct {
	http.ResponseWriter

	statusCode int
	body       []byte
}

// WriteHeader overrides the WriteHeader method of the http.ResponseWriter to capture the status code.
func (rw *ResponseWriterWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write the bytes to response
func (rw *ResponseWriterWrapper) Write(bs []byte) (int, error) {
	if rw.statusCode == http.StatusOK {
		return rw.ResponseWriter.Write(bs)
	}
	rw.body = append(rw.body, bs...)
	return rw.ResponseWriter.Write(bs)
}

// Flush the bytes to response
func (rw *ResponseWriterWrapper) Flush() {
	rw.ResponseWriter.(http.Flusher).Flush()
}

// GetStatusCode returns the captured status code.
func (rw *ResponseWriterWrapper) GetStatusCode() int {
	return rw.statusCode
}

// GetResponseBody returns the response body
func (rw *ResponseWriterWrapper) GetResponseBody() []byte {
	return rw.body
}
