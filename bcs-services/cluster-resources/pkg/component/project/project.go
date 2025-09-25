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

// Package project xxx
package project

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runmode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// Project BCS 项目
type Project struct {
	ID         string `json:"projectID"`
	Code       string `json:"projectCode"`
	BusinessID string `json:"businessID"`
}

// Namespace BCS 命名空间
type Namespace struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// IsActive 判断命名空间是否为激活状态
func (n Namespace) IsActive() bool {
	return n.Status == "Active"
}

// VariableValue 变量值
type VariableValue struct {
	Id          string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Key         string `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
	Name        string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	ClusterID   string `protobuf:"bytes,4,opt,name=clusterID,proto3" json:"clusterID,omitempty"`
	ClusterName string `protobuf:"bytes,5,opt,name=clusterName,proto3" json:"clusterName,omitempty"`
	Namespace   string `protobuf:"bytes,6,opt,name=namespace,proto3" json:"namespace,omitempty"`
	Value       string `protobuf:"bytes,7,opt,name=value,proto3" json:"value,omitempty"`
	Scope       string `protobuf:"bytes,8,opt,name=scope,proto3" json:"scope,omitempty"`
}

// GetProjectInfo 获取项目信息（bcsProject）
func GetProjectInfo(ctx context.Context, projectID string) (*Project, error) {
	if runtime.RunMode == runmode.Dev || runtime.RunMode == runmode.UnitTest {
		return fetchMockProjectInfo(projectID)
	}
	return projMgrCli.fetchProjInfoWithCache(ctx, projectID)
}

// GetProjectNamespace 获取项目命名空间（bcsProjectNamespace）
func GetProjectNamespace(ctx context.Context, projectID, clusterID string) ([]Namespace, error) {
	return projMgrCli.fetchSharedClusterProjNsWitchCache(ctx, projectID, clusterID)
}

// GetVariable get project from project code
func GetVariable(ctx context.Context, projectCode, clusterID, namespace string) ([]VariableValue, error) {
	return getVariable(ctx, projectCode, clusterID, namespace)
}

// FromContext 通过 Context 获取项目信息
func FromContext(ctx context.Context) (*Project, error) {
	p := ctx.Value(ctxkey.ProjKey)
	if p == nil {
		return nil, errorx.New(errcode.General, "project info not exists in context")
	}
	return p.(*Project), nil
}
