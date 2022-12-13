/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cluster

import (
	"context"
	"crypto/tls"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	microMetadata "go-micro.dev/v4/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// NewGrpcConn 新建 Grpc 连接
func NewGrpcConn(ctx context.Context, address string, tlsConf *tls.Config) (conn *grpc.ClientConn, err error) {
	// 组装配置信息
	md := metadata.New(map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
	})
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	if tlsConf != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConf)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	parentSpanContext := ctx.Value("ParentSpanContext")
	tracer := ctx.Value("Tracer")
	// 客户端调用追踪
	opts = append(opts, grpc.WithUnaryInterceptor(NewClientWrapper(tracer.(opentracing.Tracer), parentSpanContext.(opentracing.SpanContext))))

	// 尝试建立 grpc 连接
	return grpc.Dial(address, opts...)
}

// SetMD4CTX 为调用 Grpc 的 Context 设置 Metadata
func SetMD4CTX(ctx context.Context) context.Context {
	// 若存在 jwtToken 则透传到依赖服务
	rawMetadata, ok := microMetadata.FromContext(ctx)
	if ok {
		authorization, exists := rawMetadata.Get("Authorization")
		if exists {
			md := metadata.New(map[string]string{
				"Authorization": authorization,
			})
			return metadata.NewOutgoingContext(ctx, md)
		}
	}
	return ctx
}

type MDReaderWriter struct {
	metadata.MD
}

func (c MDReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vs := range c.MD {
		for _, v := range vs {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c MDReaderWriter) Set(key, val string) {
	key = strings.ToLower(key)
	c.MD[key] = append(c.MD[key], val)
}

func NewClientWrapper(tracer opentracing.Tracer, spanContext opentracing.SpanContext) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string,
		req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		span := opentracing.StartSpan(
			"call gRPC",
			opentracing.ChildOf(spanContext),
			opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
			ext.SpanKindRPCClient,
		)

		defer span.Finish()

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}
		mdWriter := MDReaderWriter{md}
		err := tracer.Inject(span.Context(), opentracing.TextMap, mdWriter)
		if err != nil {
			span.LogFields(log.String("inject-error", err.Error()))
		}

		newCtx := metadata.NewOutgoingContext(ctx, md)
		err = invoker(newCtx, method, req, reply, cc, opts...)
		if err != nil {
			span.LogFields(log.String("call-error", err.Error()))
		}
		return err
	}
}
