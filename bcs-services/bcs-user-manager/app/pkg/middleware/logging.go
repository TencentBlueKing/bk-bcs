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

// Package middleware xxx
package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/dustin/go-humanize"
	restful "github.com/emicklei/go-restful/v3"

	blog "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/log"
)

// LoggingFilter log request
func LoggingFilter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	inBody, err := io.ReadAll(request.Request.Body)
	if err != nil {
		_ = response.WriteError(400, err)
		return
	}
	blog.Log(request.Request.Context()).Infof("url: %s, method: %s, request: %s", request.Request.URL,
		request.Request.Method, string(inBody))
	request.Request.Body = io.NopCloser(bytes.NewReader(inBody))

	c := NewResponseCapture(response.ResponseWriter)
	response.ResponseWriter = c
	chain.ProcessFilter(request, response)

	body := string(c.Bytes())
	if len(body) > 1024 {
		body = fmt.Sprintf("%s...(Total %s)", body[:1024], humanize.Bytes(uint64(len(body))))
	}
	blog.Log(request.Request.Context()).Infof("response: %s", body)
}

// ResponseCapture is a wrapper for http response
type ResponseCapture struct {
	http.ResponseWriter
	wroteHeader bool
	status      int
	body        *bytes.Buffer
}

// NewResponseCapture creates a new ResponseCapture
func NewResponseCapture(w http.ResponseWriter) *ResponseCapture {
	return &ResponseCapture{
		ResponseWriter: w,
		wroteHeader:    false,
		body:           new(bytes.Buffer),
	}
}

// Header returns the header
func (c ResponseCapture) Header() http.Header {
	return c.ResponseWriter.Header()
}

// Write writes the data
func (c ResponseCapture) Write(data []byte) (int, error) {
	if !c.wroteHeader {
		c.WriteHeader(http.StatusOK)
	}
	c.body.Write(data)
	return c.ResponseWriter.Write(data)
}

// WriteHeader writes the header
func (c *ResponseCapture) WriteHeader(statusCode int) {
	c.status = statusCode
	c.wroteHeader = true
	c.ResponseWriter.WriteHeader(statusCode)
}

// Bytes returns the bytes
func (c ResponseCapture) Bytes() []byte {
	return c.body.Bytes()
}

// StatusCode returns the status code
func (c ResponseCapture) StatusCode() int {
	return c.status
}
