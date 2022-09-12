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

package util

import "github.com/miekg/dns"

// ResponseInterceptor xxx
type ResponseInterceptor struct {
	dns.ResponseWriter
	Msg *dns.Msg
}

// NewResponseInterceptor returns a pointer to a new ResponseReverter.
func NewResponseInterceptor(w dns.ResponseWriter) *ResponseInterceptor {
	return &ResponseInterceptor{ResponseWriter: w}
}

// WriteMsg records the status code and calls the
// underlying ResponseWriter's WriteMsg method.
func (r *ResponseInterceptor) WriteMsg(res *dns.Msg) error {
	r.Msg = res
	return nil
}

// Write is a wrapper that records the size of the message that gets written.
func (r *ResponseInterceptor) Write(buf []byte) (int, error) {
	return len(buf), nil
}

// Hijack implements dns.Hijacker. It simply wraps the underlying
// ResponseWriter's Hijack method if there is one, or returns an error.
func (r *ResponseInterceptor) Hijack() {
	return
}
