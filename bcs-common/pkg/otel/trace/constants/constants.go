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

import "go.opentelemetry.io/otel/attribute"

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

const (
	// RequestIDKey xxx
	RequestIDKey = contextKey("requestID")
	// Traceparent 上游TraceID
	Traceparent = "Traceparent"
	// TracerKey xxx
	TracerKey = "otel-go-contrib-tracer"
	// TracerName xxx
	TracerName = "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	// RequestIDHeaderKey xxx
	RequestIDHeaderKey = "X-Request-Id"
	// GRPCStatusCodeKey grpc响应码
	GRPCStatusCodeKey = attribute.Key("rpc.grpc.status_code")
	// GrpcTraceparent grpc 上游TraceID
	GrpcTraceparent = "traceparent"
)

// service name
const (
	// BCSClusterManager bcs-cluster-manager service name
	BCSClusterManager = "bcs-cluster-manager"
	// BCSHelmManager bcs-helm-manager service name
	BCSHelmManager = "bcs-helm-manager"
	// BCSProjectManager bcs-project-manager service name
	BCSProjectManager = "bcs-project-manager"
)
