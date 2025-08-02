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

package query

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"github.com/thanos-io/thanos/pkg/api"
	extpromhttp "github.com/thanos-io/thanos/pkg/extprom/http"
	"github.com/thanos-io/thanos/pkg/store"
	"google.golang.org/grpc/metadata"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest/tracing"
)

const (
	// LabelMatcherParam labelMatch
	LabelMatcherParam = "labelMatch[]"
)

// TenantAuthMiddleware tenant auth middleware
type TenantAuthMiddleware struct {
	ins extpromhttp.InstrumentationMiddleware
	ctx context.Context
}

// NewTenantAuthMiddleware 租户鉴权中间件
func NewTenantAuthMiddleware(ctx context.Context, ins extpromhttp.InstrumentationMiddleware) (*TenantAuthMiddleware,
	error) {
	return &TenantAuthMiddleware{ctx: ctx, ins: ins}, nil
}

// parseLabelMatchersParam 解析 labelMatch selector
func parseLabelMatchersParam(r *http.Request) ([][]*labels.Matcher, *api.ApiError) {
	var labelMatchers [][]*labels.Matcher
	if err := r.ParseForm(); err != nil {
		return nil, &api.ApiError{Typ: api.ErrorInternal, Err: errors.Wrap(err, "parse form")}
	}

	for _, s := range r.Form[LabelMatcherParam] {
		matchers, err := parser.ParseMetricSelector(s)
		if err != nil {
			return nil, &api.ApiError{Typ: api.ErrorBadData, Err: err}
		}
		labelMatchers = append(labelMatchers, matchers)
	}

	return labelMatchers, nil
}

// NewHandler 处理函数
func (t *TenantAuthMiddleware) NewHandler(handlerName string, handler http.Handler) http.HandlerFunc {
	handleFunc := t.ins.NewHandler(handlerName, handler)

	return func(w http.ResponseWriter, r *http.Request) {
		if config.G.Web.QueryAuth {
			// 仅内部调用
			if r.Header.Get("Authorization") != "Bearer admin" {
				api.RespondError(w, &api.ApiError{Typ: api.ErrorInternal, Err: errors.New("forbidden")}, nil)
				return
			}
		}
		labelMatchers, err := parseLabelMatchersParam(r)
		if err != nil {
			api.RespondError(w, err, nil)
			return
		}

		scopeClusteID := r.Header.Get("X-Scope-ClusterId")
		partialResponse := r.Header.Get("X-Partial-Response")
		requestID := tracing.RequestIDValue(r, true)
		blog.Infow("handle request",
			"request_id", requestID,
			"handler_name", handlerName,
			"label_matchers", fmt.Sprintf("%s", labelMatchers),
			"X-Scope-ClusterId", scopeClusteID,
			"X-Partial-Response", partialResponse,
			"req", fmt.Sprintf("%s %s", r.Method, r.URL),
			"query", r.Form.Get("query"),
			"start", r.Form.Get("start"),
			"end", r.Form.Get("end"),
			"step", r.Form.Get("step"),
		)

		// 返回的 header 写入 request_id
		w.Header().Set(store.RequestIdHeaderKey(), requestID)

		ctx := store.WithLabelMatchValue(r.Context(), labelMatchers)
		ctx = store.WithScopeClusterIDValue(ctx, scopeClusteID)
		ctx = store.WithPartialResponseValue(ctx, partialResponse)
		ctx = store.WithRequestIDValue(ctx, requestID)
		// Traceparent 透传给grpc
		ctx = metadata.AppendToOutgoingContext(ctx, "Traceparent", r.Header.Get("traceparent"))
		r = r.WithContext(ctx)
		handleFunc(w, r)
	}
}
