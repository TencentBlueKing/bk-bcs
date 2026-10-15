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

package perm

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cloudaccount"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
)

const (
	applyURL = "apply_url"
)

// UserActionPerm 用户权限上下文信息
type UserActionPerm struct {
	User     authutils.UserInfo
	ActionID string
	RelatedPermResource
}

// RelatedPermResource 通过不同的action读取不同的权限资源
type RelatedPermResource struct {
	ProjectID      string
	ClusterID      string
	Namespace      string
	TemplateID     string
	CloudAccountID string
}

// CheckUserPermByActionID check user actionID perm and return apply permURL when no permission
func CheckUserPermByActionID(ctx context.Context, iam iam.PermClient, permInfo UserActionPerm) (bool, string, error) {
	if iam == nil || permInfo.User.GetBKUserName() == "" || permInfo.ActionID == "" {
		return false, "", fmt.Errorf("CheckUserPermByActionID parameter error")
	}

	errFunc := func(action string) error {
		return fmt.Errorf("CheckUserPermByActionID project not support actionID[%s]", action)
	}

	switch {
	case actionPermInProject(permInfo.ActionID):
		projectIAM := project.NewBCSProjectPermClient(iam)
		switch permInfo.ActionID {
		case project.ProjectCreate.String():
			allow, url, _, err := projectIAM.CanCreateProject(permInfo.User.GetBKUserName())
			return allow, url, err
		case project.ProjectView.String():
			allow, url, _, err := projectIAM.CanViewProject(permInfo.User.GetBKUserName(), permInfo.ProjectID)
			return allow, url, err
		case project.ProjectEdit.String():
			allow, url, _, err := projectIAM.CanEditProject(permInfo.User.GetBKUserName(), permInfo.ProjectID)
			return allow, url, err
		case project.ProjectDelete.String():
			allow, url, _, err := projectIAM.CanDeleteProject(permInfo.User.GetBKUserName(), permInfo.ProjectID)
			return allow, url, err
		default:
			return false, "", errFunc(permInfo.ActionID)
		}
	case actionPermInCloudAccount(permInfo.ActionID):
		accountIAM := cloudaccount.NewBCSAccountPermClient(iam)
		switch permInfo.ActionID {
		case cloudaccount.AccountManage.String():
			return accountIAM.CanManageCloudAccount(permInfo.User.GetBKUserName(), permInfo.ProjectID, permInfo.CloudAccountID)
		case cloudaccount.AccountUse.String():
			return accountIAM.CanUseCloudAccount(permInfo.User.GetBKUserName(), permInfo.ProjectID, permInfo.CloudAccountID)
		case cloudaccount.AccountCreate.String():
			return accountIAM.CanCreateCloudAccount(permInfo.User.GetBKUserName(), permInfo.ProjectID)
		default:
			return false, "", errFunc(permInfo.ActionID)
		}
	default:
	}

	return false, "", errFunc(permInfo.ActionID)
}

// GetPermTypeByActionID xxx
type GetPermTypeByActionID func(actionID string) bool

func actionPermInProject(actionID string) bool {
	_, ok := project.ActionIDNameMap[iam.ActionID(actionID)]
	return ok
}

func actionPermInCloudAccount(actionID string) bool {
	_, ok := cloudaccount.ActionIDNameMap[iam.ActionID(actionID)]
	return ok
}

func actionPermInCluster(actionID string) bool { // nolint
	_, ok := cluster.ActionIDNameMap[iam.ActionID(actionID)]
	return ok
}
