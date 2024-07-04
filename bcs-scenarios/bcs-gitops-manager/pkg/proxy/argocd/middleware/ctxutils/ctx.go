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

package ctxutils

import (
	"context"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	traceconst "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/tracing"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
)

type contextKey string

const (
	ctxKeyUser contextKey = "user"
)

// RequestID return the requestID of context
func RequestID(ctx context.Context) string {
	return ctx.Value(traceconst.RequestIDHeaderKey).(string)
}

// User return user info of context
func User(ctx context.Context) *proxy.UserInfo {
	return ctx.Value(ctxKeyUser).(*proxy.UserInfo)
}

// SetContext set the user-info and trace info
func SetContext(rw http.ResponseWriter, r *http.Request, jwtDecoder *jwt.JWTClient) (context.Context, string) {
	// 获取 RequestID 信息，并重新存入上下文
	var requestID string
	requestIDHeader := r.Context().Value(traceconst.RequestIDHeaderKey)
	if v, ok := requestIDHeader.(string); ok && v != "" {
		requestID = v
	} else {
		requestID = uuid.New().String()
	}
	// nolint
	ctx := context.WithValue(r.Context(), traceconst.RequestIDHeaderKey, requestID)
	// nolint
	ctx = tracing.ContextWithRequestID(ctx, requestID)
	rw.Header().Set(traceconst.RequestIDHeaderKey, requestID)

	// 统一获取 User 信息，并存入上下文
	user, err := proxy.GetJWTInfo(r, jwtDecoder)
	if err != nil || user == nil {
		http.Error(rw, errors.Wrapf(err, "get user info failed").Error(), http.StatusUnauthorized)
		return nil, requestID
	}
	if user.ClientID != "" {
		blog.Infof("RequestID[%s] manager received user '%s' with client '%s' serve [%s/%s]",
			requestID, user.GetUser(), user.ClientID, r.Method, r.URL.RequestURI())
	} else {
		blog.Infof("RequestID[%s] manager received user '%s' serve [%s/%s]",
			requestID, user.GetUser(), r.Method, r.URL.RequestURI())
	}
	ctx = context.WithValue(ctx, ctxKeyUser, user)
	return ctx, requestID
}
