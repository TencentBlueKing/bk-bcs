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
	ID   string `json:"projectID"`
	Code string `json:"projectCode"`
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

// FromContext 通过 Context 获取项目信息
func FromContext(ctx context.Context) (*Project, error) {
	p := ctx.Value(ctxkey.ProjKey)
	if p == nil {
		return nil, errorx.New(errcode.General, "project info not exists in context")
	}
	return p.(*Project), nil
}
