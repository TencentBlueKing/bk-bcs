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

package project

import (
	blog "k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/audit"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/utils"
)

// BCSProjectPerm project perm client
type BCSProjectPerm struct {
	iamClient iam.PermClient
}

// NewBCSProjectPermClient init project perm client
func NewBCSProjectPermClient(cli iam.PermClient) *BCSProjectPerm {
	return &BCSProjectPerm{iamClient: cli}
}

// CanCreateProject check user createProject perm
func (bpp *BCSProjectPerm) CanCreateProject(user string) (bool, string, []utils.ResourceAction, error) {
	if bpp == nil {
		return false, "", nil, utils.ErrServerNotInited
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}

	allow, err := bpp.iamClient.IsAllowedWithoutResource(ProjectCreate.String(), req, false)
	if err != nil {
		return false, "", nil, err
	}
	blog.V(4).Infof("BCSProjectPerm CanCreateProject user[%s] %+v", user, allow)

	if allow {
		return allow, "", nil, nil
	}

	// generate apply url if cluster perm notAllow
	projectApp := BuildProjectApplicationInstance(ProjectApplicationAction{
		IsCreateProject: true,
		ActionID:        ProjectCreate.String(),
	})

	url, _ := bpp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, []iam.ApplicationAction{projectApp})
	return allow, url, nil, nil
}

// CanEditProject check user manageCluster perm
func (bpp *BCSProjectPerm) CanEditProject(user string, projectID string) (bool, string, []utils.ResourceAction, error) {
	if bpp == nil {
		return false, "", nil, utils.ErrServerNotInited
	}

	// related actions
	resources := []utils.ResourceAction{
		{Type: string(SysProject), Resource: projectID, Action: ProjectView.String()},
		{Type: string(SysProject), Resource: projectID, Action: ProjectEdit.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{ProjectEdit.String(), ProjectView.String()}
	projectNode := ProjectResourceNode{
		IsCreateProject: false, SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.BuildResourceNodes()

	perms, err := bpp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, [][]iam.ResourceNode{projectNode})
	if err != nil {
		return false, "", nil, err
	}
	blog.V(4).Infof("BCSProjectPerm CanEditProject user[%s] %+v", user, perms)

	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    BCSProjectModule,
		Operation: CanEditProjectOperation,
		User:      user,
	}, resources, perms)
	defer audit.AddEvent(ProjectEdit.String(), string(SysProject), projectID, user, allow, nil)
	if err != nil {
		return false, "", nil, err
	}

	if allow {
		return allow, "", nil, nil
	}

	// generate apply url
	projectApp := BuildProjectSameInstanceApplication(false,
		[]string{ProjectEdit.String(), ProjectView.String()}, []string{projectID})

	url, _ := bpp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, append(projectApp))
	return allow, url, resources, nil
}

// CanDeleteProject check user deleteProject perm
func (bpp *BCSProjectPerm) CanDeleteProject(user string, projectID string) (bool, string,
	[]utils.ResourceAction, error) {
	if bpp == nil {
		return false, "", nil, utils.ErrServerNotInited
	}

	// related actions
	resources := []utils.ResourceAction{
		{Type: string(SysProject), Resource: projectID, Action: ProjectDelete.String()},
		{Type: string(SysProject), Resource: projectID, Action: ProjectView.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{ProjectDelete.String(), ProjectView.String()}
	projectNode := ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.BuildResourceNodes()

	perms, err := bpp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, [][]iam.ResourceNode{projectNode})
	if err != nil {
		return false, "", nil, err
	}
	blog.V(4).Infof("BCSProjectPerm CanDeleteProject user[%s] %+v", user, perms)

	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    BCSProjectModule,
		Operation: CanDeleteProjectOperation,
		User:      user,
	}, resources, perms)
	defer audit.AddEvent(ProjectDelete.String(), string(SysProject), projectID, user, allow, nil)

	if err != nil {
		return false, "", nil, err
	}

	if allow {
		return allow, "", nil, nil
	}

	// generate apply url
	projectApp := BuildProjectSameInstanceApplication(false,
		[]string{ProjectDelete.String(), ProjectView.String()}, []string{projectID})
	url, _ := bpp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, append(projectApp))
	return allow, url, resources, nil
}

// CanViewProject check user viewProject perm
func (bpp *BCSProjectPerm) CanViewProject(user string, projectID string) (bool, string, []utils.ResourceAction, error) {
	if bpp == nil {
		return false, "", nil, utils.ErrServerNotInited
	}

	// related actions
	resources := []utils.ResourceAction{
		{Type: string(SysProject), Resource: projectID, Action: ProjectView.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{ProjectView.String()}
	projectNode := ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.BuildResourceNodes()

	perms, err := bpp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, [][]iam.ResourceNode{projectNode})
	if err != nil {
		return false, "", nil, err
	}
	blog.V(4).Infof("BCSProjectPerm CanViewProject user[%s] %+v", user, perms)

	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    BCSProjectModule,
		Operation: CanViewProjectOperation,
		User:      user,
	}, resources, perms)
	defer audit.AddEvent(ProjectView.String(), string(SysProject), projectID, user, allow, nil)

	if err != nil {
		return false, "", nil, err
	}

	if allow {
		return allow, "", nil, nil
	}

	// generate apply url
	projectApp := BuildProjectApplicationInstance(ProjectApplicationAction{
		IsCreateProject: false,
		ActionID:        ProjectView.String(),
		Data:            []string{projectID},
	})

	url, _ := bpp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, []iam.ApplicationAction{projectApp})
	return allow, url, resources, nil
}

// GenerateIAMApplicationURL build permission URL
func (bpp *BCSProjectPerm) GenerateIAMApplicationURL(systemID string, applications []iam.ApplicationAction) (string,
	error) {
	if bpp == nil {
		return iam.IamAppURL, utils.ErrServerNotInited
	}

	url, err := bpp.iamClient.GetApplyURL(iam.ApplicationRequest{SystemID: systemID}, applications, iam.BkUser{
		BkUserName: iam.SystemUser,
	})
	if err != nil {
		return iam.IamAppURL, err
	}

	return url, nil
}

// GetProjectMultiActionPermission only support same instanceSelection
func (bpp *BCSProjectPerm) GetProjectMultiActionPermission(user, projectID string, actionIDs []string) (map[string]bool,
	error) {
	if bpp == nil {
		return nil, utils.ErrServerNotInited
	}

	projectNode := ProjectResourceNode{
		IsCreateProject: false, SystemID: iam.SystemIDBKBCS,
		ProjectID: projectID}.BuildResourceNodes()

	return bpp.iamClient.ResourceMultiActionsAllowed(actionIDs, iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user}, projectNode)
}

// GetMultiProjectMultiActionPerm only support same instanceSelection
func (bpp *BCSProjectPerm) GetMultiProjectMultiActionPerm(user string, projectIDs,
	actionIDs []string) (map[string]map[string]bool, error) {
	if bpp == nil {
		return nil, utils.ErrServerNotInited
	}

	resourceNodes := make([][]iam.ResourceNode, 0)
	for i := range projectIDs {
		clusterNode := ProjectResourceNode{
			IsCreateProject: false, SystemID: iam.SystemIDBKBCS,
			ProjectID: projectIDs[i]}.BuildResourceNodes()
		resourceNodes = append(resourceNodes, clusterNode)
	}

	return bpp.iamClient.BatchResourceMultiActionsAllowed(actionIDs, iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user}, resourceNodes)
}
