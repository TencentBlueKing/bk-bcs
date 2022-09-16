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

package auth

import (
	bcsIAM "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	iamPerm "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
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

// ProjectIamClient iam client for project
var ProjectIamClient *project.BCSProjectPerm

// InitPermClient init perm client
func InitPermClient() error {
	opts := &bcsIAM.Options{
		SystemID:    bcsIAM.SystemIDBKBCS,
		AppCode:     config.GlobalConf.App.Code,
		AppSecret:   config.GlobalConf.App.Secret,
		External:    !config.GlobalConf.IAM.UseGWHost,
		GateWayHost: config.GlobalConf.IAM.GatewayHost,
		IAMHost:     config.GlobalConf.IAM.IAMHost,
		BkiIAMHost:  config.GlobalConf.IAM.BKPaaSHost,
		Metric:      false,
		Debug:       config.GlobalConf.IAM.Debug,
	}
	cli, err := bcsIAM.NewIamClient(opts)
	if err != nil {
		return err
	}
	ProjectIamClient = iamPerm.NewBCSProjectPermClient(cli)
	return nil
}

const (
	// CreateAction xxx
	CreateAction = "create"
	// ViewAction xxx
	ViewAction = "view"
	// UpdateAction xxx
	UpdateAction = "update"
	// DeleteAction xxx
	DeleteAction = "delete"
)
