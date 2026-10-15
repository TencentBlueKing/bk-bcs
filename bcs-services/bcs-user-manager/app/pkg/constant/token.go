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

// Package constant xxx
package constant

const (
	// CurrentUserAttr user header
	CurrentUserAttr = "current-user"
	// CurrentTenantID user tenant id header
	CurrentTenantID = "current-tenant-id"
	// ProjectAttr project header
	ProjectAttr = "project"

	// DefaultTokenLength user token default length
	// token is consisted of digital and alphabet(case sensetive)
	// we can refer to http://coolaf.com/tool/rd when testing
	DefaultTokenLength = 32
	// TokenKeyPrefix is the redis key for token
	// nolint
	TokenKeyPrefix = "bcs_auth:token:" // NOCC:gas/crypto(误报)
	// TokenLimits for token
	TokenLimits = 1

	// RequestIDHeaderKey X-Request-Id
	RequestIDHeaderKey = "X-Request-Id"
	// RequestIDKey requestID
	RequestIDKey = "requestID"
	// Traceparent Traceparent
	Traceparent = "Traceparent"
	// TracerName go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
	TracerName = "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)
