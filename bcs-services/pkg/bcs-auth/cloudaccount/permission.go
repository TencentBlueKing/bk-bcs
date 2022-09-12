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
 *
 */

package cloudaccount

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
)

// BCSCloudAccountPerm account perm client
type BCSCloudAccountPerm struct {
	iamClient iam.PermClient
}

// NewBCSAccountPermClient init account perm client
func NewBCSAccountPermClient(cli iam.PermClient) *BCSCloudAccountPerm {
	return &BCSCloudAccountPerm{iamClient: cli}
}

// CanManageCloudAccount check user manageAccount perm
func (bcp *BCSCloudAccountPerm) CanManageCloudAccount(user string, projectID string, accountID string) (bool, string,
	error) {
	if bcp == nil {
		return false, "", utils.ErrServerNotInited
	}

	// related actions
	resources := []utils.ResourceAction{
		{Resource: accountID, Action: AccountManage.String()},
		{Resource: projectID, Action: project.ProjectView.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{AccountManage.String(), project.ProjectView.String()}
	accountNode := AccountResourceNode{
		IsCreateAccount: false, SystemID: iam.SystemIDBKBCS,
		ProjectID: projectID, AccountID: accountID}.BuildResourceNodes()
	projectNode := project.ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.BuildResourceNodes()

	perms, err := bcp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, [][]iam.ResourceNode{accountNode,
		projectNode})
	if err != nil {
		return false, "", err
	}
	blog.V(4).Infof("BCSCloudAccountPerm CanManageCloudAccount user[%s] %+v", user, perms)

	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    BCSCloudAccountModule,
		Operation: CanManageCloudAccountOperation,
		User:      user,
	}, resources, perms)
	if err != nil {
		return false, "", err
	}

	if allow {
		return allow, "", nil
	}

	// generate apply url
	accountApps := BuildAccountApplicationInstance(AccountApplicationAction{
		IsCreateAccount: false,
		ActionID:        AccountManage.String(),
		Data: []ProjectAccountData{
			{
				Project: projectID,
				Account: accountID,
			},
		}},
	)
	projectApp := project.BuildProjectApplicationInstance(project.ProjectApplicationAction{
		IsCreateProject: false,
		ActionID:        project.ProjectView.String(),
		Data:            []string{projectID},
	})

	url, _ := bcp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, []iam.ApplicationAction{accountApps, projectApp})
	return allow, url, nil
}

// CanUseCloudAccount check user use cloudAccount perm
func (bcp *BCSCloudAccountPerm) CanUseCloudAccount(user string, projectID string, accountID string) (bool, string,
	error) {
	if bcp == nil {
		return false, "", utils.ErrServerNotInited
	}

	// related actions
	resources := []utils.ResourceAction{
		{Resource: accountID, Action: AccountUse.String()},
		{Resource: projectID, Action: project.ProjectView.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{AccountUse.String(), project.ProjectView.String()}
	accountNode := AccountResourceNode{
		IsCreateAccount: false, SystemID: iam.SystemIDBKBCS,
		ProjectID: projectID, AccountID: accountID}.BuildResourceNodes()
	projectNode := project.ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.BuildResourceNodes()

	perms, err := bcp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, [][]iam.ResourceNode{accountNode,
		projectNode})
	if err != nil {
		return false, "", err
	}
	blog.V(4).Infof("BCSCloudAccountPerm CanUseCloudAccount user[%s] %+v", user, perms)

	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    BCSCloudAccountModule,
		Operation: CanUseCloudAccountOperation,
		User:      user,
	}, resources, perms)
	if err != nil {
		return false, "", err
	}

	if allow {
		return allow, "", nil
	}

	// generate apply url
	accountApps := BuildAccountApplicationInstance(AccountApplicationAction{
		IsCreateAccount: false,
		ActionID:        AccountUse.String(),
		Data: []ProjectAccountData{
			{
				Project: projectID,
				Account: accountID,
			},
		}},
	)
	projectApp := project.BuildProjectApplicationInstance(project.ProjectApplicationAction{
		IsCreateProject: false,
		ActionID:        project.ProjectView.String(),
		Data:            []string{projectID},
	})

	url, _ := bcp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, []iam.ApplicationAction{accountApps, projectApp})
	return allow, url, nil
}

// GetMultiAccountMultiActionPermission only support same instanceSelection
func (bcp *BCSCloudAccountPerm) GetMultiAccountMultiActionPermission(user, projectID string, accountIDs []string,
	actionIDs []string) (map[string]map[string]bool, error) {
	if bcp == nil {
		return nil, utils.ErrServerNotInited
	}

	resourceNodes := make([][]iam.ResourceNode, 0)
	for i := range accountIDs {
		clusterNode := AccountResourceNode{
			IsCreateAccount: false, SystemID: iam.SystemIDBKBCS,
			ProjectID: projectID, AccountID: accountIDs[i]}.BuildResourceNodes()
		resourceNodes = append(resourceNodes, clusterNode)
	}

	return bcp.iamClient.BatchResourceMultiActionsAllowed(actionIDs, iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user}, resourceNodes)
}

// GenerateIAMApplicationURL build permission URL
func (bcp *BCSCloudAccountPerm) GenerateIAMApplicationURL(systemID string, applications []iam.ApplicationAction) (
	string, error) {
	if bcp == nil {
		return iam.IamAppURL, utils.ErrServerNotInited
	}

	url, err := bcp.iamClient.GetApplyURL(iam.ApplicationRequest{SystemID: systemID}, applications, iam.BkUser{
		BkUserName: iam.SystemUser,
	})
	if err != nil {
		return iam.IamAppURL, err
	}

	return url, nil
}
