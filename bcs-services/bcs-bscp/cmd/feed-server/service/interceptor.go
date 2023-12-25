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
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

var (
	// 兼容老的请求, 只部分方法使用中间件
	allowMethod = map[string]struct{}{
		"/pbfs.Upstream/ListApps":   {},
		"/pbfs.Upstream/PullKvMeta": {},
		"/pbfs.Upstream/GetKvValue": {},
	}
)

// ctxKey context key
type ctxKey int

const (
	credentialKey ctxKey = iota
)

func withCredential(ctx context.Context, value *types.CredentialCache) context.Context {
	return context.WithValue(ctx, credentialKey, value)
}

// getCredential 包内私有方法断言, 认为一直可用
func getCredential(ctx context.Context) *types.CredentialCache {
	return ctx.Value(credentialKey).(*types.CredentialCache)
}

func getBearerToken(md metadata.MD) (string, error) {
	values := md.Get("authorization")
	if len(values) < 1 {
		return "", fmt.Errorf("missing authorization header")
	}

	authorizationHeader := values[0]
	authHeaderParts := strings.Split(authorizationHeader, " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return authHeaderParts[1], nil
}

func (s *Service) authorize(ctx context.Context, bizID uint32) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Aborted, "missing grpc metadata")
	}

	token, err := getBearerToken(md)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	cred, err := s.bll.Auth().GetCred(kit.FromGrpcContext(ctx), bizID, token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}
	if !cred.Enabled {
		return nil, status.Errorf(codes.PermissionDenied, err.Error())
	}

	// 获取scope，到下一步处理
	ctx = withCredential(ctx, cred)
	return ctx, nil
}

// FeedUnaryAuthInterceptor feed 鉴权中间件
func FeedUnaryAuthInterceptor(
	ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 兼容老的请求
	if _, ok := allowMethod[info.FullMethod]; !ok {
		return handler(ctx, req)
	}

	var bizID uint32
	switch r := req.(type) {
	case interface{ GetBizId() uint32 }: // 请求都必须有 uint32 biz_id 参数
		bizID = r.GetBizId()
	default:
		return nil, status.Error(codes.Aborted, "missing bizId in request")
	}

	svr := info.Server.(*Service)
	ctx, err := svr.authorize(ctx, bizID)
	if err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

// wrappedStream stream 封装, 可自定义 context 传值
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context 覆盖 context
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

	var bizID uint32
	ctx, err := srv.(*Service).authorize(ss.Context(), bizID)
	if err != nil {
		return err
	}

	w := &wrappedStream{ServerStream: ss, ctx: ctx}
	return handler(srv, w)
}
