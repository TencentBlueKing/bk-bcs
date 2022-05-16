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

package perm

import (
	bcsIAM "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	iamPerm "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/stringx"
)

const (
	// ProjectCreate 创建项目
	ProjectCreate string = "project_create"
	// ProjectView 查看项目
	ProjectView string = "project_view"
	// ProjectEdit 编辑项目
	ProjectEdit string = "project_edit"
	// ProjectDelete 删除项目
	ProjectDelete string = "project_delete"
)

// NewPermClient new a perm client
func NewPermClient() (*iamPerm.BCSProjectPerm, error) {
	opts := &bcsIAM.Options{
		SystemID:    bcsIAM.SystemIDBKBCS,
		AppCode:     config.GlobalConf.App.Code,
		AppSecret:   config.GlobalConf.App.Secret,
		External:    !config.GlobalConf.IAM.UseGWHost,
		GateWayHost: config.GlobalConf.IAM.GatewayHost,
		Metric:      false,
		Debug:       config.GlobalConf.IAM.Debug,
	}
	cli, err := bcsIAM.NewIamClient(opts)
	if err != nil {
		return nil, err
	}
	return iamPerm.NewBCSProjectPermClient(cli), nil
}

const (
	CreateAction = "create"
	ViewAction   = "view"
	UpdateAction = "update"
	DeleteAction = "delete"
)

// 是否豁免权限
func canExemptClientPerm(clientID string, action string) bool {
	clientActions := config.GlobalConf.ClientActionExemptPerm.ClientActions
	for _, ca := range clientActions {
		if ca.ClientID == clientID && stringx.StringInSlice(action, ca.Actions) {
			return true
		}
	}
	return false
}

// CanCreateProject ...
func CanCreateProject(authUser auth.AuthUser) error {
	// 判断是否校验权限
	if canExemptClientPerm(authUser.ClientID, CreateAction) {
		return nil
	}

	// 判断是否有创建权限
	permClient, err := NewPermClient()
	if err != nil {
		return errorx.NewIAMClientErr(err)
	}

	canCreate, applyUrl, err := permClient.CanCreateProject(authUser.Username)
	if err != nil {
		return errorx.NewRequestIAMErr(applyUrl, "projectCreate", canCreate, err)
	}
	if !canCreate {
		return errorx.NewPermDeniedErr(applyUrl, "projectCreate", canCreate)
	}
	return nil
}

// CanViewProject ...
func CanViewProject(authUser auth.AuthUser, projectID string) error {
	// 判断是否校验权限
	if canExemptClientPerm(authUser.ClientID, ViewAction) {
		return nil
	}

	permClient, err := NewPermClient()
	if err != nil {
		return errorx.NewIAMClientErr(err)
	}

	canView, applyUrl, err := permClient.CanViewProject(authUser.Username, projectID)
	if err != nil {
		return errorx.NewRequestIAMErr(applyUrl, "projectView", canView, err)
	}
	if !canView {
		return errorx.NewPermDeniedErr(applyUrl, "projectView", canView)
	}
	return nil
}

// CanEditProject ...
func CanEditProject(authUser auth.AuthUser, projectID string) error {
	// 判断是否校验权限
	if canExemptClientPerm(authUser.ClientID, UpdateAction) {
		return nil
	}

	permClient, err := NewPermClient()
	if err != nil {
		return errorx.NewIAMClientErr(err)
	}
	// 校验是否有编辑权限
	canEdit, applyUrl, err := permClient.CanEditProject(authUser.Username, projectID)
	if err != nil {
		return errorx.NewRequestIAMErr(applyUrl, "projectEdit", canEdit, err)
	}
	if !canEdit {
		return errorx.NewPermDeniedErr(applyUrl, "projectEdit", canEdit, err.Error())
	}
	return nil
}

// CanEditProject ...
func CanDeleteProject(authUser auth.AuthUser, projectID string) error {
	// 判断是否校验权限
	if canExemptClientPerm(authUser.ClientID, DeleteAction) {
		return nil
	}

	permClient, err := NewPermClient()
	if err != nil {
		return errorx.NewIAMClientErr(err)
	}
	// NOTE: 不校验集群
	canDelete, applyUrl, err := permClient.CanDeleteProject(authUser.Username, projectID, "")
	if err != nil {
		return errorx.NewRequestIAMErr(applyUrl, "projectDelete", canDelete, err)
	}
	if !canDelete {
		return errorx.NewPermDeniedErr(applyUrl, "projectDelete", canDelete, err.Error())
	}
	return nil
}
