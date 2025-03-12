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

	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/server"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/proto/bcs-federation-manager"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
)

const (
	// NoPermissionErr auth failed
	NoPermissionErr = 40403
)

// ResponseWrapper handler response
func ResponseWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		requestID := getRequestID(ctx)
		ctx = context.WithValue(ctx, RequestIDContextKey, requestID)
		err = fn(ctx, req, rsp)
		return renderResponse(rsp, requestID, err)
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
			// code in federation manager is *uint32 type instead of int32 type
			codePtr := &errCode
			v.Elem().FieldByName("Code").Set(reflect.ValueOf(codePtr))
		}
		if v.Elem().FieldByName("Message").IsValid() {
			v.Elem().FieldByName("Message").SetString(errMsg)
		}

		if v.Elem().FieldByName("WebAnnotations").IsValid() {
			perms := &proto.WebAnnotations{}
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

// getRequestID get request id
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
