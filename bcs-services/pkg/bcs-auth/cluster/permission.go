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

package cluster

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
)

// BCSClusterPerm cluster perm client
type BCSClusterPerm struct {
	iamClient iam.PermClient
}

// NewBCSClusterPermClient init cluster perm client
func NewBCSClusterPermClient(cli iam.PermClient) *BCSClusterPerm {
	return &BCSClusterPerm{iamClient: cli}
}

// CanCreateCluster check user createCluster perm
func (bcp *BCSClusterPerm) CanCreateCluster(user string, projectID string) (bool, string, error) {
	if bcp == nil {
		return false, "", utils.ErrServerNotInited
	}

	// related actions
	resources := []utils.ResourceAction{
		{Resource: projectID, Action: ClusterCreate.String()},
		{Resource: projectID, Action: project.ProjectView.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{ClusterCreate.String(), project.ProjectView.String()}
	clusterNode := ClusterResourceNode{
		IsCreateCluster: true, SystemID: iam.SystemIDBKBCS,
		ProjectID: projectID, ClusterID: ""}.BuildResourceNodes()
	projectNode := project.ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.BuildResourceNodes()

	// get cluster permission by iam
	perms, err := bcp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, [][]iam.ResourceNode{clusterNode, projectNode})
	if err != nil {
		return false, "", err
	}
	blog.V(4).Infof("BCSClusterPerm CanCreateCluster user[%s] %+v", user, perms)

	// check cluster resource perms
	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    BCSClusterModule,
		Operation: CanCreateClusterOperation,
		User:      user,
	}, resources, perms)
	if err != nil {
		return false, "", err
	}

	if allow {
		return allow, "", nil
	}

	// generate apply url if cluster perm notAllow
	clusterApp := BuildClusterApplicationInstance(ClusterApplicationAction{
		IsCreateCluster: true,
		ActionID:        ClusterCreate.String(),
		Data: []ProjectClusterData{
			{
				Project: projectID,
				Cluster: "",
			},
		},
	})
	projectApp := project.BuildProjectApplicationInstance(project.ProjectApplicationAction{
		IsCreateProject: false,
		ActionID:        project.ProjectView.String(),
		Data:            []string{projectID},
	})

	url, _ := bcp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, []iam.ApplicationAction{clusterApp, projectApp})
	return allow, url, nil
}

// CanManageCluster check user manageCluster perm
func (bcp *BCSClusterPerm) CanManageCluster(user string, projectID string, clusterID string) (bool, string, error) {
	if bcp == nil {
		return false, "", utils.ErrServerNotInited
	}

	// related actions
	resources := []utils.ResourceAction{
		{Resource: clusterID, Action: ClusterManage.String()},
		{Resource: projectID, Action: project.ProjectView.String()},
		{Resource: clusterID, Action: ClusterView.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{ClusterManage.String(), project.ProjectView.String(), ClusterView.String()}
	clusterNode := ClusterResourceNode{
		IsCreateCluster: false, SystemID: iam.SystemIDBKBCS,
		ProjectID: projectID, ClusterID: clusterID}.BuildResourceNodes()
	projectNode := project.ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.BuildResourceNodes()

	perms, err := bcp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, [][]iam.ResourceNode{clusterNode, projectNode})
	if err != nil {
		return false, "", err
	}
	blog.V(4).Infof("BCSClusterPerm CanManageCluster user[%s] %+v", user, perms)

	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    BCSClusterModule,
		Operation: CanManageClusterOperation,
		User:      user,
	}, resources, perms)
	if err != nil {
		return false, "", err
	}

	if allow {
		return allow, "", nil
	}

	// generate apply url
	clusterApps := BuildClusterSameInstanceApplication(false,
		[]string{ClusterManage.String(), ClusterView.String()}, []ProjectClusterData{
			{
				Project: projectID,
				Cluster: clusterID,
			},
		})
	projectApp := project.BuildProjectApplicationInstance(project.ProjectApplicationAction{
		IsCreateProject: false,
		ActionID:        project.ProjectView.String(),
		Data:            []string{projectID},
	})

	url, _ := bcp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, append(clusterApps, projectApp))
	return allow, url, nil
}

// CanDeleteCluster check user deleteCluster perm
func (bcp *BCSClusterPerm) CanDeleteCluster(user string, projectID string, clusterID string) (bool, string, error) {
	if bcp == nil {
		return false, "", utils.ErrServerNotInited
	}

	// related actions
	resources := []utils.ResourceAction{
		{Resource: clusterID, Action: ClusterDelete.String()},
		{Resource: projectID, Action: project.ProjectView.String()},
		{Resource: clusterID, Action: ClusterView.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{ClusterDelete.String(), project.ProjectView.String(), ClusterView.String()}
	clusterNode := ClusterResourceNode{
		IsCreateCluster: false, SystemID: iam.SystemIDBKBCS,
		ProjectID: projectID, ClusterID: clusterID}.BuildResourceNodes()
	projectNode := project.ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.BuildResourceNodes()

	perms, err := bcp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, [][]iam.ResourceNode{clusterNode, projectNode})
	if err != nil {
		return false, "", err
	}
	blog.V(4).Infof("BCSClusterPerm CanDeleteCluster user[%s] %+v", user, perms)

	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    BCSClusterModule,
		Operation: CanDeleteClusterOperation,
		User:      user,
	}, resources, perms)

	if err != nil {
		return false, "", err
	}

	if allow {
		return allow, "", nil
	}

	// generate apply url
	clusterApps := BuildClusterSameInstanceApplication(false,
		[]string{ClusterDelete.String(), ClusterView.String()}, []ProjectClusterData{
			{
				Project: projectID,
				Cluster: clusterID,
			},
		})
	projectApp := project.BuildProjectApplicationInstance(project.ProjectApplicationAction{
		IsCreateProject: false,
		ActionID:        project.ProjectView.String(),
		Data:            []string{projectID},
	})

	url, _ := bcp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, append(clusterApps, projectApp))
	return allow, url, nil
}

// CanViewCluster check user viewCluster perm
func (bcp *BCSClusterPerm) CanViewCluster(user string, projectID string, clusterID string) (bool, string, error) {
	if bcp == nil {
		return false, "", utils.ErrServerNotInited
	}

	// related actions
	resources := []utils.ResourceAction{
		{Resource: clusterID, Action: ClusterView.String()},
		{Resource: projectID, Action: project.ProjectView.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{ClusterView.String(), project.ProjectView.String()}
	clusterNode := ClusterResourceNode{
		IsCreateCluster: false, SystemID: iam.SystemIDBKBCS,
		ProjectID: projectID, ClusterID: clusterID}.BuildResourceNodes()
	projectNode := project.ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.BuildResourceNodes()

	perms, err := bcp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, [][]iam.ResourceNode{clusterNode, projectNode})
	if err != nil {
		return false, "", err
	}
	blog.V(4).Infof("BCSClusterPerm CanViewCluster user[%s] %+v", user, perms)

	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    BCSClusterModule,
		Operation: CanViewClusterOperation,
		User:      user,
	}, resources, perms)

	if err != nil {
		return false, "", err
	}

	if allow {
		return allow, "", nil
	}

	// generate apply url
	clusterApps := BuildClusterSameInstanceApplication(false,
		[]string{ClusterView.String()}, []ProjectClusterData{
			{
				Project: projectID,
				Cluster: clusterID,
			},
		})
	projectApp := project.BuildProjectApplicationInstance(project.ProjectApplicationAction{
		IsCreateProject: false,
		ActionID:        project.ProjectView.String(),
		Data:            []string{projectID},
	})

	url, _ := bcp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, append(clusterApps, projectApp))
	return allow, url, nil
}

// GenerateIAMApplicationURL build permission URL
func (bcp *BCSClusterPerm) GenerateIAMApplicationURL(systemID string, applications []iam.ApplicationAction) (string, error) {
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

// GetClusterMultiActionPermission only support same instanceSelection
func (bcp *BCSClusterPerm) GetClusterMultiActionPermission(user, projectID, clusterID string, actionIDs []string) (map[string]bool, error) {
	if bcp == nil {
		return nil, utils.ErrServerNotInited
	}

	clusterNode := ClusterResourceNode{
		IsCreateCluster: false, SystemID: iam.SystemIDBKBCS,
		ProjectID: projectID, ClusterID: clusterID}.BuildResourceNodes()

	return bcp.iamClient.ResourceMultiActionsAllowed(actionIDs, iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user}, clusterNode)
}

// GetClusterMultiActionPermission only support same instanceSelection
func (bcp *BCSClusterPerm) GetMultiClusterMultiActionPermission(user, projectID string, clusterIDs []string, actionIDs []string) (map[string]map[string]bool, error) {
	if bcp == nil {
		return nil, utils.ErrServerNotInited
	}

	resourceNodes := make([][]iam.ResourceNode, 0)
	for i := range clusterIDs {
		clusterNode := ClusterResourceNode{
			IsCreateCluster: false, SystemID: iam.SystemIDBKBCS,
			ProjectID: projectID, ClusterID: clusterIDs[i]}.BuildResourceNodes()
		resourceNodes = append(resourceNodes, clusterNode)
	}

	return bcp.iamClient.BatchResourceMultiActionsAllowed(actionIDs, iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user}, resourceNodes)
}
