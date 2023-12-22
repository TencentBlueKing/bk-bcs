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

package service

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

var (
	allowMethod = map[string]struct{}{
		"/pbfs.Upstream/ListApps": {},
	}
)

func (s *Service) authorize(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Aborted, "not have valid metadata")
	}
	p, _ := peer.FromContext(ctx)
	fmt.Println("lejioamin", md, p.Addr.String())
	return ctx, nil
}

// FeedUnaryAuthInterceptor feed 鉴权中间件
func FeedUnaryAuthInterceptor(
	ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 兼容老的请求
	if _, ok := allowMethod[info.FullMethod]; !ok {
		return handler(ctx, req)
	}

	svr := info.Server.(*Service)
	ctx, err := svr.authorize(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := handler(ctx, req)
	return resp, err
}

// wrappedStream stream 封装, 可自定义 context 传值
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *wrappedStream) Context() context.Context {
	return s.ctx
}

// FeedStreamAuthInterceptor feed 鉴权中间件
func FeedStreamAuthInterceptor(
	srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// 兼容老的请求
	if _, ok := allowMethod[info.FullMethod]; !ok {
		return handler(srv, ss)
	}

	ctx, err := srv.(*Service).authorize(ss.Context())
	if err != nil {
		return err
	}

	w := &wrappedStream{ServerStream: ss, ctx: ctx}
	return handler(srv, w)
}
