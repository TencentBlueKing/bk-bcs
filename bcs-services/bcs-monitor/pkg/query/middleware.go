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

package query

import (
	"context"
	"net/http"

	extpromhttp "github.com/thanos-io/thanos/pkg/extprom/http"

	"github.com/TencentBlueKing/bkmonitor-kits/logger"
)

type tenantAuthMiddleware struct {
	ins extpromhttp.InstrumentationMiddleware
	ctx context.Context
}

// NewTenantAuthMiddleware 租户鉴权中间件
func NewTenantAuthMiddleware(ctx context.Context, ins extpromhttp.InstrumentationMiddleware) (*tenantAuthMiddleware, error) {
	return &tenantAuthMiddleware{ctx: ctx, ins: ins}, nil
}

// NewHandler 处理函数
func (t *tenantAuthMiddleware) NewHandler(handlerName string, handler http.Handler) http.HandlerFunc {
	handleFunc := t.ins.NewHandler(handlerName, handler)

	return func(w http.ResponseWriter, r *http.Request) {
		logger.Infow("handle request", "handler_name", handlerName, "url", r.URL)

		// ctx := store.WithLabelMatchValue(r.Context(), labelMatches)
		// r = r.WithContext(ctx)
		handleFunc(w, r)
	}
}
