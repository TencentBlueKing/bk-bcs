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

// Package utils 提供mesh manager的工具函数
package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/server"

	projectClient "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

const (
	// NoPermissionErr auth failed
	NoPermissionErr = 40403
)

type (
	// stringValueMeshID 解析 google.protobuf.StringValue 类型的meshID请求结构
	stringValueMeshID struct {
		MeshID *struct {
			Value string `json:"value"`
		} `json:"meshID"`
	}

	// stringMeshID 解析字符串格式的meshID请求结构
	stringMeshID struct {
		MeshID string `json:"meshID"`
	}

	// projectStringValue 解析 google.protobuf.StringValue 类型的项目请求结构
	projectStringValue struct {
		ProjectCode *struct {
			Value string `json:"value"`
		} `json:"projectCode,omitempty"`
		ProjectID *struct {
			Value string `json:"value"`
		} `json:"projectID,omitempty"`
	}

	// projectString 解析字符串格式的项目请求结构
	projectString struct {
		ProjectCode string `json:"projectCode,omitempty"`
		ProjectID   string `json:"projectID,omitempty"`
	}
)

// RequestLogWarpper log request
func RequestLogWarpper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		// get metadata
		md, _ := metadata.FromContext(ctx)
		requestID := ctx.Value(RequestIDContextKey)
		blog.Infof("receive %s, metadata: %v, req: %v, requestID: %s", req.Method(), md, req.Body(), requestID)
		return fn(ctx, req, rsp)
	}
}

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
			// code in mesh manager is *uint32 type instead of int32 type
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

// ParseProjectIDWrapper parse projectID from req
func ParseProjectIDWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		projectCode, err := extractProjectCode(req)
		if err != nil {
			return fmt.Errorf("failed to extract project code: %s", err)
		}

		if projectCode == "" {
			blog.Warn("projectCode is empty")
			return fn(ctx, req, rsp)
		}

		pj, err := projectClient.GetProjectByCode(ctx, projectCode)
		if err != nil {
			return fmt.Errorf("failed to get project by code %s: %s", projectCode, err.Error())
		}

		ctx = context.WithValue(ctx, ProjectIDContextKey, pj.ProjectID)
		ctx = context.WithValue(ctx, ProjectCodeContextKey, pj.ProjectCode)
		return fn(ctx, req, rsp)
	}
}

// extractProjectCode 从请求中提取项目代码
func extractProjectCode(req server.Request) (string, error) {
	body := req.Body()
	b, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %s", err)
	}

	// 尝试解析 google.protobuf.StringValue 格式
	if projectCode := extractFromStringValue(b); projectCode != "" {
		return projectCode, nil
	}

	// 尝试解析直接字符串格式
	if projectCode := extractFromString(b); projectCode != "" {
		return projectCode, nil
	}

	return "", nil
}

// extractFromStringValue 从 StringValue 格式中提取项目代码
func extractFromStringValue(data []byte) string {
	var req projectStringValue
	if err := json.Unmarshal(data, &req); err != nil {
		return ""
	}

	projectCode := ""
	if req.ProjectCode != nil {
		projectCode = req.ProjectCode.Value
	}

	// 如果 projectCode 为空，使用 projectID
	if projectCode == "" && req.ProjectID != nil {
		projectCode = req.ProjectID.Value
	}

	return projectCode
}

// extractFromString 从字符串格式中提取项目代码
func extractFromString(data []byte) string {
	var req projectString
	if err := json.Unmarshal(data, &req); err != nil {
		return ""
	}

	projectCode := req.ProjectCode
	// 如果 projectCode 为空，使用 projectID
	if projectCode == "" {
		projectCode = req.ProjectID
	}

	return projectCode
}

// ParseMeshIDWrapper parse meshID from req
func ParseMeshIDWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		// 安装接口不获取网格ID
		if req.Method() == common.MeshManagerInstallIstio {
			return fn(ctx, req, rsp)
		}

		meshID, err := getMeshIDFromRequest(req)
		if err != nil {
			blog.Warnf("ParseMeshIDWrapper error: failed to get meshID for method %s, err: %s", req.Method(), err.Error())
			return fn(ctx, req, rsp)
		}

		if meshID == "" {
			blog.Warnf("ParseMeshIDWrapper error: meshID is empty for method %s", req.Method())
			return fn(ctx, req, rsp)
		}

		ctx = context.WithValue(ctx, MeshIDContextKey, meshID)
		return fn(ctx, req, rsp)
	}
}

// getMeshIDFromRequest 从请求中获取网格ID
func getMeshIDFromRequest(req server.Request) (string, error) {
	switch req.Method() {
	case common.MeshManagerUpdateIstio, common.MeshManagerDeleteIstio, common.MeshManagerGetIstioDetail:
		// meshID通过URL路径参数传递，grpc-gateway会将其设置到proto消息中
		body := req.Body()
		b, err := json.Marshal(body)
		if err != nil {
			return "", err
		}

		// 解析 google.protobuf.StringValue 格式的 meshID
		var meshRequest stringValueMeshID
		if err = json.Unmarshal(b, &meshRequest); err == nil && meshRequest.MeshID != nil && meshRequest.MeshID.Value != "" {
			return meshRequest.MeshID.Value, nil
		}

		// 解析字符串格式的 meshID
		var meshStringRequest stringMeshID
		if err = json.Unmarshal(b, &meshStringRequest); err == nil && meshStringRequest.MeshID != "" {
			return meshStringRequest.MeshID, nil
		}

		return "", fmt.Errorf("meshID not found in request for method %s", req.Method())
	default:
		// 其他接口不需要处理meshID
		return "", nil
	}
}

// ParseInstallClustersWrapper 解析安装接口中的集群信息
func ParseInstallClustersWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		// 只处理安装接口
		if req.Method() != common.MeshManagerInstallIstio {
			return fn(ctx, req, rsp)
		}

		primaryClusters, remoteClusters, err := getClustersFromInstallRequest(req)
		if err != nil {
			blog.Warnf("failed to get clusters for method %s, err: %s", req.Method(), err.Error())
			return fn(ctx, req, rsp)
		}

		// 将集群信息存储到context中
		ctx = context.WithValue(ctx, PrimaryClustersContextKey, primaryClusters)
		ctx = context.WithValue(ctx, RemoteClustersContextKey, remoteClusters)

		return fn(ctx, req, rsp)
	}
}

type clusters struct {
	PrimaryClusters []string `json:"primaryClusters,omitempty"`
	RemoteClusters  []string `json:"remoteClusters,omitempty"`
}

// getClustersFromInstallRequest 从安装请求中获取集群信息
func getClustersFromInstallRequest(req server.Request) ([]string, []string, error) {
	body := req.Body()
	b, err := json.Marshal(body)
	if err != nil {
		return nil, nil, err
	}

	installReq := &clusters{}
	if err := json.Unmarshal(b, installReq); err != nil {
		return nil, nil, err
	}

	return installReq.PrimaryClusters, installReq.RemoteClusters, nil
}
