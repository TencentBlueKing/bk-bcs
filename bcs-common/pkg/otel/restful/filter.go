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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/utils"
	"github.com/emicklei/go-restful"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
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
func NewOTFilter(options ...FilterOption) restful.FilterFunction {
	opts := filterOptions{
		operationNameFunc: DefaultOperationNameFunc,
		componentName:     DefaultComponentName,
	}
	for _, opt := range options {
		opt(&opts)
	}

	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		ctx, span := utils.Tracer(opts.operationNameFunc(req)).Start(req.Request.Context(), "Processing Request")
		setHTTPSpanAttributes(span, req.Request)
		//requestID := "requestID"
		//hc := propagation.HeaderCarrier(req.Request.Header)
		//
		//hc.Set(requestID, "0000001")
		//otel.GetTextMapPropagator().Inject(req.Request.Context(), propagation.HeaderCarrier(req.Request.Header))
		//ctx = otel.GetTextMapPropagator().Extract(req.Request.Context(), propagation.HeaderCarrier(req.Request.Header))
		req.Request = req.Request.WithContext(utils.ContextWithSpan(ctx, span))
		//span.SetAttributes(attribute.Key(requestID).StringSlice(req.Request.Header[requestID]))
		span.SetAttributes(attribute.Key("component").String(opts.componentName))

		defer func() {
			// record HTTP status code
			span.SetAttributes(utils.HTTPStatusCodeKey.Int(resp.StatusCode()))

			span.SetAttributes(utils.HTTPResponseContentLengthKey.Int(resp.ContentLength()))
			if resp.Error() != nil {
				span.RecordError(resp.Error())
			}
			span.End()
		}()
		chain.ProcessFilter(req, resp)
	}
}

func setHTTPSpanAttributes(span trace.Span, request *http.Request) {
	attrs := []attribute.KeyValue{}

	if request.Method != "" {
		attrs = append(attrs, utils.HTTPMethodKey.String(request.Method))
	} else {
		attrs = append(attrs, utils.HTTPMethodKey.String(http.MethodGet))
	}

	// remove any username/password info that may be in the URL
	// before adding it to the attributes
	userinfo := request.URL.User
	request.URL.User = nil

	attrs = append(attrs, utils.HTTPURLKey.String(request.URL.String()))

	// restore any username/password info that was removed
	request.URL.User = userinfo

	if request.TLS != nil {
		attrs = append(attrs, utils.HTTPSchemeKey.String("https"))
	} else {
		attrs = append(attrs, utils.HTTPSchemeKey.String("http"))
	}

	if request.Host != "" {
		attrs = append(attrs, utils.HTTPHostKey.String(request.Host))
	}

	span.SetAttributes(attrs...)
}
