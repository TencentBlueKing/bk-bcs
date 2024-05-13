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

package utils

import (
	"context"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	"github.com/go-micro/plugins/v4/server/grpc"
	microSvc "go-micro.dev/v4"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/server"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

const (
	// NoPermissionErr auth failed
	NoPermissionErr = 40403
)

var (
	// MaxBodySize define maximum message size that grpc server can send or receive. Default value is 50MB.
	MaxBodySize = 1024 * 1024 * 50
)

// MaxMsgSize of the max msg size
func MaxMsgSize(s int) microSvc.Option {
	return func(o *microSvc.Options) {
		_ = o.Server.Init(grpc.MaxMsgSize(s))
	}
}

// RequestLogWarpper log request
func RequestLogWarpper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		md, _ := metadata.FromContext(ctx)
		blog.Infof("receive %s, metadata: %v, req: %v", req.Method(), md, req.Body())
		return fn(ctx, req, rsp)
	}
}

// ResponseWrapper 处理返回
func ResponseWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		requestID := getRequestID(ctx)
		ctx = context.WithValue(ctx, RequestIDContextKey, requestID)
		err = fn(ctx, req, rsp)
		return renderResponse(rsp, requestID, err)
	}
}

// NewAuditWrapper 审计
func NewAuditWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		startTime := time.Now()
		err := fn(ctx, req, rsp)
		endTime := time.Now()
		// async add audit
		go addAudit(ctx, req, rsp, startTime, endTime)
		return err
	}
}

func renderResponse(rsp interface{}, requestID string, err error) error {
	v := reflect.ValueOf(rsp)

	if v.Elem().FieldByName("RequestID").IsValid() {
		v.Elem().FieldByName("RequestID").SetString(requestID)
	}

	if err == nil {
		return nil
	}
	switch e := err.(type) {
	case *authutils.PermDeniedError:
		errCode := uint32(NoPermissionErr)
		errMsg := err.(*authutils.PermDeniedError).Error()
		if v.Elem().FieldByName("Code").IsValid() {
			v.Elem().FieldByName("Code").SetUint(uint64(errCode))
		}
		if v.Elem().FieldByName("Message").IsValid() {
			v.Elem().FieldByName("Message").SetString(errMsg)
		}

		if v.Elem().FieldByName("WebAnnotations").IsValid() {
			perms := &proto.WebAnnotationsV2{}
			permsMap := map[string]interface{}{}
			permsMap["apply_url"] = e.Perms.ApplyURL
			actionList := []map[string]string{}
			for _, actions := range e.Perms.ActionList {
				actionList = append(actionList, map[string]string{
					"action_id":     actions.Action,
					"resource_type": actions.Type,
				})
			}
			permsMap["action_list"] = actionList
			perms.Perms = Map2pbStruct(permsMap)
			v.Elem().FieldByName("WebAnnotations").Set(reflect.ValueOf(perms))
			return nil
		}
		return err
	default:
		return err
	}
}

// getRequestID 获取 request id
func getRequestID(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return GenUUID()
	}
	// 当request id不存在或者为空时，生成id
	requestID, ok := md.Get(RequestIDHeaderKey)
	if !ok || requestID == "" {
		return GenUUID()
	}

	return requestID
}

// HandleLanguageWrapper 从上下文获取语言
func HandleLanguageWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		md, _ := metadata.FromContext(ctx)
		ctx = i18n.WithLanguage(ctx, getLangFromCookies(md))
		return fn(ctx, req, rsp)
	}
}

// getLangFromCookies 从 Cookies 中获取语言版本
func getLangFromCookies(md metadata.Metadata) string {
	cookies, ok := md.Get(common.MetadataCookiesKey)

	if !ok {
		return i18n.DefaultLanguage
	}
	for _, c := range Split(cookies) {
		k, v := Partition(c, "=")
		if k != common.LangCookieName {
			continue
		}
		return v
	}
	return i18n.DefaultLanguage
}
