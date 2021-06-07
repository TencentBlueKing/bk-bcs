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
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/emicklei/go-restful"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	// DefaultComponentName show default component
	DefaultComponentName = "go-restful"
)

var (
	// DefaultOperationNameFunc get default operation name
	DefaultOperationNameFunc = func(r *restful.Request) string {
		// extract the route that the request maps to and use it as the operation name.
		return r.SelectedRoutePath()
	}

	responseSizeKey = "http.response_size"
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

// NewOTFilter returns a go-restful filter which add OpenTracing instrument
func NewOTFilter(tracer opentracing.Tracer, options ...FilterOption) restful.FilterFunction {
	opts := filterOptions{
		operationNameFunc: DefaultOperationNameFunc,
		componentName:     DefaultComponentName,
	}
	for _, opt := range options {
		opt(&opts)
	}

	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		spanCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Request.Header))
		if err != nil {
			blog.V(4).Infof("NewOTFilter tracer extract failed: %v",  err)
		}

		// record operation name
		span := tracer.StartSpan(opts.operationNameFunc(req), ext.RPCServerOption(spanCtx))
		req.Request = req.Request.WithContext(opentracing.ContextWithSpan(req.Request.Context(), span))

		defer func() {
			// record HTTP status code
			ext.HTTPStatusCode.Set(span, uint16(resp.StatusCode()))
			span.SetTag(responseSizeKey, resp.ContentLength())
			if resp.Error() != nil {
				ext.Error.Set(span, true)
			}
			span.Finish()
		}()

		// record component name
		ext.Component.Set(span, opts.componentName)
		// record HTTP method
		ext.HTTPMethod.Set(span, req.Request.Method)
		// record HTTP url
		ext.HTTPUrl.Set(span, req.Request.URL.String())

		chain.ProcessFilter(req, resp)
	}
}

