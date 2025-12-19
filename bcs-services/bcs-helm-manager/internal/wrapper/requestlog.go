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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
)

// RequestLogWarpper log request
func RequestLogWarpper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		// get metadata
		md, _ := metadata.FromContext(ctx)
		ctx = context.WithValue(ctx, contextx.LangContectKey, i18n.GetLangFromCookies(md))
		blog.V(6).Infof("receive %s, metadata: %v, req: %v", req.Method(), md, req.Body())
		return fn(ctx, req, rsp)
	}
}
