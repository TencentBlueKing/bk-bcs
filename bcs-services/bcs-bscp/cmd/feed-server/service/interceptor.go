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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbfs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/feed-server"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
	sfs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/sf-share"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

var (
	// 老的请求,不使用中间件
	disabledMethod = map[string]struct{}{
		"/pbfs.Upstream/Handshake":       {},
		"/pbfs.Upstream/Messaging":       {},
		"/pbfs.Upstream/Watch":           {},
		"/pbfs.Upstream/PullAppFileMeta": {},
		"/pbfs.Upstream/GetDownloadURL":  {},
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
		return nil, status.Errorf(codes.PermissionDenied, "credential is disabled")
	}

	// 获取scope，到下一步处理
	ctx = withCredential(ctx, cred)
	return ctx, nil
}

// FeedUnaryAuthInterceptor feed 鉴权中间件
func FeedUnaryAuthInterceptor(
	ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 兼容老的请求
	if _, ok := disabledMethod[info.FullMethod]; ok {
		return handler(ctx, req)
	}

	var bizID uint32
	switch r := req.(type) {
	case interface{ GetBizId() uint32 }: // 请求都必须有 uint32 biz_id 参数
		bizID = r.GetBizId()
	default:
		return nil, status.Error(codes.Aborted, "missing bizId in request")
	}

	ctx = context.WithValue(ctx, constant.BizIDKey, bizID) //nolint:staticcheck

	svc, ok := info.Server.(*Service)
	// 处理非业务 Service 时不鉴权，如 GRPC Reflection
	if !ok {
		return handler(ctx, req)
	}

	ctx, err := svc.authorize(ctx, bizID)
	if err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

// FeedUnaryUpdateLastConsumedTimeInterceptor feed 更新拉取时间中间件
func FeedUnaryUpdateLastConsumedTimeInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	svc, ok := info.Server.(*Service)
	// 跳过非业务 Service，如 GRPC Reflection
	if !ok {
		return handler(ctx, req)
	}

	type lastConsumedTime struct {
		BizID    uint32
		AppNames []string
		AppIDs   []uint32
	}

	param := lastConsumedTime{}

	switch info.FullMethod {
	case pbfs.Upstream_GetKvValue_FullMethodName:
		request := req.(*pbfs.GetKvValueReq)
		param.BizID = request.BizId
		param.AppNames = append(param.AppNames, request.GetAppMeta().App)
	case pbfs.Upstream_PullKvMeta_FullMethodName:
		request := req.(*pbfs.PullKvMetaReq)
		param.BizID = request.BizId
		param.AppNames = append(param.AppNames, request.GetAppMeta().App)
	case pbfs.Upstream_Messaging_FullMethodName:
		request := req.(*pbfs.MessagingMeta)
		if sfs.MessagingType(request.Type) == sfs.VersionChangeMessage {
			vc := new(sfs.VersionChangePayload)
			if err := vc.Decode(request.Payload); err != nil {
				logs.Errorf("version change message decoding failed, %s", err.Error())
				return handler(ctx, req)
			}
			param.BizID = vc.BasicData.BizID
			param.AppNames = append(param.AppNames, vc.Application.App)
		}
	case pbfs.Upstream_Watch_FullMethodName:
		request := req.(*pbfs.SideWatchMeta)
		payload := new(sfs.SideWatchPayload)
		if err := jsoni.Unmarshal(request.Payload, payload); err != nil {
			logs.Errorf("parse request payload failed, %s", err.Error())
			return handler(ctx, req)
		}
		param.BizID = payload.BizID
		for _, v := range payload.Applications {
			param.AppNames = append(param.AppNames, v.App)
		}
	case pbfs.Upstream_PullAppFileMeta_FullMethodName:
		request := req.(*pbfs.PullAppFileMetaReq)
		param.BizID = request.BizId
		param.AppNames = append(param.AppNames, request.GetAppMeta().App)
	case pbfs.Upstream_GetDownloadURL_FullMethodName:
		request := req.(*pbfs.GetDownloadURLReq)
		param.BizID = request.BizId
		param.AppIDs = append(param.AppIDs, request.GetFileMeta().GetConfigItemAttachment().AppId)
	case pbfs.Upstream_GetSingleKvValue_FullMethodName, pbfs.Upstream_GetSingleKvMeta_FullMethodName:
		request := req.(*pbfs.GetSingleKvValueReq)
		param.BizID = request.BizId
		param.AppNames = append(param.AppNames, request.GetAppMeta().App)
	default:
		return handler(ctx, req)
	}

	if param.BizID != 0 {
		ctx = context.WithValue(ctx, constant.BizIDKey, param.BizID) //nolint:staticcheck

		if len(param.AppIDs) == 0 {
			for _, appName := range param.AppNames {
				appID, err := svc.bll.AppCache().GetAppID(kit.FromGrpcContext(ctx), param.BizID, appName)
				if err != nil {
					logs.Errorf("get app id failed, err: %v", err)
					return handler(ctx, req)
				}
				param.AppIDs = append(param.AppIDs, appID)
			}
		}

		if err := svc.bll.AppCache().BatchUpdateLastConsumedTime(kit.FromGrpcContext(ctx),
			param.BizID, param.AppIDs); err != nil {
			logs.Errorf("batch update app last consumed failed, err: %v", err)
			return handler(ctx, req)
		}
		logs.Infof("batch update app last consumed time success")
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
	if _, ok := disabledMethod[info.FullMethod]; ok {
		return handler(srv, ss)
	}

	var bizID uint32
	svc, ok := srv.(*Service)
	// 处理非业务 Service 时不鉴权，如 GRPC Reflection
	if !ok {
		return handler(srv, ss)
	}
	ctx, err := svc.authorize(ss.Context(), bizID)
	if err != nil {
		return err
	}

	w := &wrappedStream{ServerStream: ss, ctx: ctx}
	return handler(srv, w)
}
