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

package restful

import (
	"context"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/emicklei/go-restful"
	"net/http"
	"strings"
)

const (
	// DefaultComponentName show default component
	DefaultComponentName     = "go-restful"
	tracerLogHandlerID   key = 32702 // random key
	realIPValueID        key = 16221
)

type key int

var (
	// DefaultOperationNameFunc get default operation name
	DefaultOperationNameFunc = func(r *restful.Request) string {
		// extract the route that the request maps to and use it as the operation name.
		return r.SelectedRoutePath()
	}
)

type filterOptions struct {
	operationNameFunc func(r *restful.Request) string
	componentName     string
}

// FilterOption controls the behavior of the Filter.
type FilterOption func(*filterOptions)

// OperationNameFunc returns a FilterOption that uses given function f
// to generate operation name for each server-side span.
func OperationNameFunc(f func(r *restful.Request) string) FilterOption {
	return func(options *filterOptions) {
		options.operationNameFunc = f
	}
}

// ComponentName returns a FilterOption that sets the component name
// name for the server-side span.
func ComponentName(componentName string) FilterOption {
	return func(options *filterOptions) {
		options.componentName = componentName
	}
}

// NewLTFilter returns a go-restful filter which add OpenTracing instrument
func NewLTFilter(options ...FilterOption) restful.FilterFunction {
	opts := filterOptions{
		operationNameFunc: DefaultOperationNameFunc,
	}
	for _, opt := range options {
		opt(&opts)
	}

	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		var ctx context.Context
		var logTracer blog.Trace
		ri := req.Request.Header.Get("X-Request-Id")

		if len(ri) > 0 {
			logTracer = blog.WithID(opts.operationNameFunc(req), ri)
		} else {
			logTracer = blog.New(opts.operationNameFunc(req))
		}

		lastRoute, ip := func(r *http.Request) (string, string) {
			lastRoute := strings.Split(r.RemoteAddr, ":")[0]
			if ip, exists := r.Header["X-Real-IP"]; exists && len(ip) > 0 {
				return lastRoute, ip[0]
			}
			if ips, exists := r.Header["X-Forwarded-For"]; exists && len(ips) > 0 {
				return lastRoute, ips[0]
			}
			return lastRoute, lastRoute
		}(req.Request)

		logTracer.Infof("traceID=[%s] event=[request-in] remote=[%s] route=[%s] method=[%s] url=[%s]",
			logTracer.ID(), ip, lastRoute, req.Request.Method, req.Request.URL.String())
		defer logTracer.Info("traceID=[%s] event=[request-out]", logTracer.ID())

		ctx = context.WithValue(ctx, tracerLogHandlerID, logTracer)
		ctx = context.WithValue(ctx, realIPValueID, ip)

		req.Request = req.Request.WithContext(ctx)

		chain.ProcessFilter(req, resp)
	}
}
