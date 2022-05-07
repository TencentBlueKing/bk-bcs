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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/errorx"
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

func NewPermClient() (*iamPerm.BCSProjectPerm, error) {
	opts := &bcsIAM.Options{
		SystemID:    bcsIAM.SystemIDBKBCS,
		AppCode:     config.GlobalConf.IAM.AppCode,
		AppSecret:   config.GlobalConf.IAM.AppSecret,
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

// CanCreateProject ...
func CanCreateProject(username string) error {
	// 判断是否校验权限
	if config.GlobalConf.ActionExemptPerm.Create {
		return nil
	}
	// 判断是否有创建权限
	permClient, err := NewPermClient()
	if err != nil {
		return errorx.NewIAMClientErr(err)
	}

	canCreate, applyUrl, err := permClient.CanCreateProject(username)
	if err != nil {
		return errorx.NewRequestIAMErr(applyUrl, "projectCreate", canCreate, err)
	}
	if !canCreate {
		return errorx.NewPermDeniedErr(applyUrl, "projectCreate", canCreate)
	}
	return nil
}

// CanViewProject ...
func CanViewProject(username string, projectID string) error {
	// 判断是否校验权限
	if config.GlobalConf.ActionExemptPerm.View {
		return nil
	}
	permClient, err := NewPermClient()
	if err != nil {
		return errorx.NewIAMClientErr(err)
	}

	canView, applyUrl, err := permClient.CanViewProject(username, projectID)
	if err != nil {
		return errorx.NewRequestIAMErr(applyUrl, "projectView", canView, err)
	}
	if !canView {
		return errorx.NewPermDeniedErr(applyUrl, "projectView", canView)
	}
	return nil
}

// CanEditProject ...
func CanEditProject(username, projectID string) error {
	// 判断是否校验权限
	if config.GlobalConf.ActionExemptPerm.Update {
		return nil
	}

	permClient, err := NewPermClient()
	if err != nil {
		return errorx.NewIAMClientErr(err)
	}
	// 校验是否有编辑权限
	canEdit, applyUrl, err := permClient.CanEditProject(username, projectID)
	if err != nil {
		return errorx.NewRequestIAMErr(applyUrl, "projectEdit", canEdit, err)
	}
	if !canEdit {
		return errorx.NewPermDeniedErr(applyUrl, "projectEdit", canEdit, err)
	}
	return nil
}

// CanEditProject ...
func CanDeleteProject(username string, projectID string) error {
	// 判断是否校验权限
	if config.GlobalConf.ActionExemptPerm.Delete {
		return nil
	}

	permClient, err := NewPermClient()
	if err != nil {
		return errorx.NewIAMClientErr(err)
	}
	// NOTE: 不校验集群
	canDelete, applyUrl, err := permClient.CanDeleteProject(username, projectID, "")
	if err != nil {
		return errorx.NewRequestIAMErr(applyUrl, "projectDelete", canDelete, err)
	}
	if !canDelete {
		return errorx.NewPermDeniedErr(applyUrl, "projectDelete", canDelete, err)
	}
	return nil
}
