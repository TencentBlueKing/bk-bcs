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
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

var (
	// gRPCComponentTag component tag: grpc
	gRPCComponentTag = opentracing.Tag{string(ext.Component), "gRPC"}
)

// OpenTracingServerInterceptor returns a grpc.UnaryServerInterceptor suitable for use in a grpc.NewServer call.
//
// For example:
//
//     s := grpc.NewServer(
//         ...,  // (existing ServerOptions)
//         grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(tracer)))
//
// All gRPC server spans will look for an OpenTracing SpanContext in the gRPC
// metadata; if found, the server span will act as the ChildOf that RPC SpanContext.
//
// Root or not, the server Span will be embedded in the context.Context for the
// application-specific gRPC handler(s) to access.
func OpenTracingServerInterceptor(tracer opentracing.Tracer, optFuncs ...Option) grpc.UnaryServerInterceptor {
	otgrpcOpts := newOptions()
	otgrpcOpts.apply(optFuncs...)
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		spanContext, err := extractSpanContext(ctx, tracer)
		if err != nil && err != opentracing.ErrSpanContextNotFound {
			blog.V(4).Infof("grpc server OpenTracingServerInterceptor get rootSpan failed: %v", err)
		}

		// inclusionFunc show whether or not trace server
		if otgrpcOpts.inclusionFunc != nil &&
			!otgrpcOpts.inclusionFunc(spanContext, info.FullMethod, req, nil) {
			return handler(ctx, req)
		}

		serverSpan := tracer.StartSpan(
			info.FullMethod,
			ext.RPCServerOption(spanContext),
			gRPCComponentTag,
		)
		defer serverSpan.Finish()

		// delivery serverSpan by context
		ctx = opentracing.ContextWithSpan(ctx, serverSpan)
		if otgrpcOpts.logPayloads {
			serverSpan.LogFields(log.Object("gRPC request", req))
		}
		resp, err = handler(ctx, req)
		if err == nil {
			if otgrpcOpts.logPayloads {
				serverSpan.LogFields(log.Object("gRPC response", resp))
			}
			SetSpanTags(serverSpan, err, false)
		} else {
			SetSpanTags(serverSpan, err, false)
			serverSpan.LogFields(log.String("event", "error"), log.String("message", err.Error()))
		}
		if otgrpcOpts.decorator != nil {
			otgrpcOpts.decorator(ctx, serverSpan, info.FullMethod, req, resp, err)
		}
		return resp, err
	}
}

// OpenTracingStreamServerInterceptor returns a grpc.StreamServerInterceptor suitable
// for use in a grpc.NewServer call. The interceptor instruments streaming RPCs by
// creating a single span to correspond to the lifetime of the RPC's stream.
//
// For example:
//
//     s := grpc.NewServer(
//         ...,  // (existing ServerOptions)
//         grpc.StreamInterceptor(otgrpc.OpenTracingStreamServerInterceptor(tracer)))
//
// All gRPC server spans will look for an OpenTracing SpanContext in the gRPC
// metadata; if found, the server span will act as the ChildOf that RPC SpanContext.
//
// Root or not, the server Span will be embedded in the context.Context for the
// application-specific gRPC handler(s) to access.
func OpenTracingStreamServerInterceptor(tracer opentracing.Tracer, optFuncs ...Option) grpc.StreamServerInterceptor {
	otgrpcOpts := newOptions()
	otgrpcOpts.apply(optFuncs...)
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		spanContext, err := extractSpanContext(ss.Context(), tracer)
		if err != nil && err != opentracing.ErrSpanContextNotFound {
			blog.V(4).Infof("grpc server OpenTracingStreamServerInterceptor get rootSpan failed: %v", err)
		}

		// inclusionFunc show whether or not trace server
		if otgrpcOpts.inclusionFunc != nil &&
			!otgrpcOpts.inclusionFunc(spanContext, info.FullMethod, nil, nil) {
			return handler(srv, ss)
		}

		serverSpan := tracer.StartSpan(
			info.FullMethod,
			ext.RPCServerOption(spanContext),
			gRPCComponentTag,
		)
		defer serverSpan.Finish()

		// delivery serverSpan by context
		ss = &openTracingServerStream{
			ServerStream: ss,
			ctx:          opentracing.ContextWithSpan(ss.Context(), serverSpan),
		}
		err = handler(srv, ss)
		SetSpanTags(serverSpan, err, false)

		if err != nil {
			serverSpan.LogFields(log.String("event", "error"), log.String("message", err.Error()))
		}

		if otgrpcOpts.decorator != nil {
			otgrpcOpts.decorator(ss.Context(), serverSpan, info.FullMethod, nil, nil, err)
		}
		return err
	}
}

type openTracingServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context return ctx from openTracingServerStream
func (ss *openTracingServerStream) Context() context.Context {
	return ss.ctx
}

func extractSpanContext(ctx context.Context, tracer opentracing.Tracer) (opentracing.SpanContext, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	return tracer.Extract(opentracing.HTTPHeaders, metadataReaderWriter{md})
}

// metadataReaderWriter satisfies both the opentracing.TextMapReader and
// opentracing.TextMapWriter interfaces.
type metadataReaderWriter struct {
	metadata.MD
}

// Set set key,val for metadataReaderWriter
func (w metadataReaderWriter) Set(key, val string) {
	// The GRPC HPACK implementation rejects any uppercase keys here.
	//
	// As such, since the HTTP_HEADERS format is case-insensitive anyway, we
	// blindly lowercase the key (which is guaranteed to work in the
	// Inject/Extract sense per the OpenTracing spec).
	key = strings.ToLower(key)
	w.MD[key] = append(w.MD[key], val)
}

// ForeachKey handler key,val from metadataReaderWriter
func (w metadataReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range w.MD {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}

	return nil
}
