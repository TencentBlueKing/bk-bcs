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

// Package constants xxx
package constants

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
)

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

const (
	// RequestIDKey xxx
	RequestIDKey = contextKey("requestID")
	// ServerName server name
	ServerName = "bcs-ui"
	// TracerName tracer name
	TracerName = "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	// RequestIDHeaderKey request_id header key
	RequestIDHeaderKey = "X-Request-Id"
	// ClaimsKey claims key
	ClaimsKey = "bcs-claims"

	// BluekingLanguage switch cookies constant
	BluekingLanguage = "blueking_language"
)

// ServiceDomain service domain
var ServiceDomain = fmt.Sprintf("%s%s", config.G.Base.ModuleName, ".bkbcs.tencent.com")
