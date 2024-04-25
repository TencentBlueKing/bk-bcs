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

package wrapper

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
)

// NewInjectContextWrapper 生成 request id, 用于操作审计等便于跟踪
func NewInjectContextWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		// generate uuid， e.g. 40a05290d67a4a39a04c705a0ee56add
		// Note: trace id by opentelemetry
		if ctx.Value(ctxkey.RequestIDKey) == nil {
			ctx = context.WithValue(ctx, ctxkey.RequestIDKey, stringx.GenUUID())
		}
		return fn(ctx, req, rsp)
	}
}

// HandleLanguageWrapper 从上下文获取语言
func HandleLanguageWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		md, _ := metadata.FromContext(ctx)
		ctx = i18n.WithLanguage(ctx, getLangFromCookies(md))
		return fn(ctx, req, rsp)
	}
}

// getLangFromCookies 从 Cookies 中获取语言版本
func getLangFromCookies(md metadata.Metadata) string {
	cookies, ok := md.Get(constant.MetadataCookiesKey)

	if !ok {
		return i18n.DefaultLanguage
	}
	for _, c := range stringx.SplitString(cookies) {
		k, v := stringx.Partition(c, "=")
		if k != constant.LangCookieName {
			continue
		}
		return v
	}
	return i18n.DefaultLanguage
}
