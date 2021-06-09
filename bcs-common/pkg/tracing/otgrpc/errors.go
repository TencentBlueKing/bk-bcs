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

package grpc

import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// A Class is a set of types of outcomes (including errors) that will often
// be handled in the same way.
type Class string

const (
	// Unknown represents unknown errors
	Unknown Class = "0xx"
	// Success represents outcomes that achieved the desired results.
	Success Class = "2xx"
	// ClientError represents errors that were the client's fault.
	ClientError Class = "4xx"
	// ServerError represents errors that were the server's fault.
	ServerError Class = "5xx"
)

// ErrorClass returns the class of the given error
func ErrorClass(err error) Class {
	if s, ok := status.FromError(err); ok {
		switch s.Code() {
		// Success or "success"
		case codes.OK, codes.Canceled:
			return Success

		// Client errors
		case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists,
			codes.PermissionDenied, codes.Unauthenticated, codes.FailedPrecondition,
			codes.OutOfRange:
			return ClientError

		// Server errors
		case codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted,
			codes.Unimplemented, codes.Internal, codes.Unavailable, codes.DataLoss:
			return ServerError

		// Not sure
		case codes.Unknown:
			fallthrough
		default:
			return Unknown
		}
	}
	return Unknown
}

// SetSpanTags sets one or more tags on the given span according to the
// error.
func SetSpanTags(span opentracing.Span, err error, client bool) {
	c := ErrorClass(err)
	code := codes.Unknown
	if s, ok := status.FromError(err); ok {
		code = s.Code()
	}
	span.SetTag("response_code", code)
	span.SetTag("response_class", c)
	if err == nil {
		return
	}
	if c != Success && (client || c == ServerError) {
		ext.Error.Set(span, true)
	}
}
