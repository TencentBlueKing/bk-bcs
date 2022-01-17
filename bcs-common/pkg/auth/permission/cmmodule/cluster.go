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

package cmmodule

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/permission/utils"
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
		return false, "", ErrServerNotInited
	}

	// related actions
	resources := []ResourceAction{
		{Resource: projectID, Action: iam.ClusterCreate.String()},
		{Resource: projectID, Action: iam.ProjectView.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{iam.ClusterCreate.String(), iam.ProjectView.String()}
	clusterNode := ClusterResourceNode{
		IsCreateCluster: true, SystemID: iam.SystemIDBKBCS,
		ProjectID: projectID, ClusterID: ""}.BuildResourceNodes()
	projectNode := ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.BuildResourceNodes()

	perms, err := bcp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, [][]iam.ResourceNode{clusterNode, projectNode})
	if err != nil {
		return false, "", err
	}
	blog.V(4).Infof("BCSClusterPerm CanCreateCluster user[%s] %+v", user, perms)

	allow, err := CheckResourcePerms(CheckResourceRequest{
		Module:    utils.BCSClusterModule,
		Operation: utils.CanCreateClusterOperation,
		User:      user,
	}, resources, perms)
	if err != nil {
		return false, "", err
	}

	if allow {
		return allow, "", nil
	}

	// generate apply url
	clusterApp := BuildClusterApplicationInstance(ClusterApplicationAction{
		IsCreateCluster: true,
		ActionID:        iam.ClusterCreate.String(),
		Data: []ProjectClusterData{
			{
				Project: projectID,
				Cluster: "",
			},
		},
	})
	projectApp := BuildProjectApplicationInstance(ProjectApplicationAction{
		IsCreateProject: false,
		ActionID:        iam.ProjectView.String(),
		Data:            []string{projectID},
	})

	url, err := bcp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, []iam.ApplicationAction{clusterApp, projectApp})
	return allow, url, nil
}

// CanManageCluster check user manageCluster perm
func (bcp *BCSClusterPerm) CanManageCluster(user string, projectID string, clusterID string) (bool, string, error) {
	if bcp == nil {
		return false, "", ErrServerNotInited
	}

	// related actions
	resources := []ResourceAction{
		{Resource: clusterID, Action: iam.ClusterManage.String()},
		{Resource: projectID, Action: iam.ProjectView.String()},
		{Resource: clusterID, Action: iam.ClusterView.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{iam.ClusterManage.String(), iam.ProjectView.String(), iam.ClusterView.String()}
	clusterNode := ClusterResourceNode{
		IsCreateCluster: false, SystemID: iam.SystemIDBKBCS,
		ProjectID: projectID, ClusterID: clusterID}.BuildResourceNodes()
	projectNode := ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.BuildResourceNodes()

	perms, err := bcp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, [][]iam.ResourceNode{clusterNode, projectNode})
	if err != nil {
		return false, "", err
	}
	blog.V(4).Infof("BCSClusterPerm CanManageCluster user[%s] %+v", user, perms)

	allow, err := CheckResourcePerms(CheckResourceRequest{
		Module:    utils.BCSClusterModule,
		Operation: utils.CanManageClusterOperation,
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
		[]string{iam.ClusterManage.String(), iam.ClusterView.String()}, []ProjectClusterData{
			{
				Project: projectID,
				Cluster: clusterID,
			},
		})
	projectApp := BuildProjectApplicationInstance(ProjectApplicationAction{
		IsCreateProject: false,
		ActionID:        iam.ProjectView.String(),
		Data:            []string{projectID},
	})

	url, err := bcp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, append(clusterApps, projectApp))
	return allow, url, nil
}

// CanDeleteCluster check user deleteCluster perm
func (bcp *BCSClusterPerm) CanDeleteCluster(user string, projectID string, clusterID string) (bool, string, error) {
	if bcp == nil {
		return false, "", ErrServerNotInited
	}

	// related actions
	resources := []ResourceAction{
		{Resource: clusterID, Action: iam.ClusterDelete.String()},
		{Resource: projectID, Action: iam.ProjectView.String()},
		{Resource: clusterID, Action: iam.ClusterView.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{iam.ClusterDelete.String(), iam.ProjectView.String(), iam.ClusterView.String()}
	clusterNode := ClusterResourceNode{
		IsCreateCluster: false, SystemID: iam.SystemIDBKBCS,
		ProjectID: projectID, ClusterID: clusterID}.BuildResourceNodes()
	projectNode := ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.BuildResourceNodes()

	perms, err := bcp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, [][]iam.ResourceNode{clusterNode, projectNode})
	if err != nil {
		return false, "", err
	}
	blog.V(4).Infof("BCSClusterPerm CanDeleteCluster user[%s] %+v", user, perms)

	allow, err := CheckResourcePerms(CheckResourceRequest{
		Module:    utils.BCSClusterModule,
		Operation: utils.CanDeleteClusterOperation,
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
		[]string{iam.ClusterDelete.String(), iam.ClusterView.String()}, []ProjectClusterData{
			{
				Project: projectID,
				Cluster: clusterID,
			},
		})
	projectApp := BuildProjectApplicationInstance(ProjectApplicationAction{
		IsCreateProject: false,
		ActionID:        iam.ProjectView.String(),
		Data:            []string{projectID},
	})

	url, err := bcp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, append(clusterApps, projectApp))
	return allow, url, nil
}

// CanViewCluster check user viewCluster perm
func (bcp *BCSClusterPerm) CanViewCluster(user string, projectID string, clusterID string) (bool, string, error) {
	if bcp == nil {
		return false, "", ErrServerNotInited
	}

	// related actions
	resources := []ResourceAction{
		{Resource: clusterID, Action: iam.ClusterView.String()},
		{Resource: projectID, Action: iam.ProjectView.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{iam.ClusterView.String(), iam.ProjectView.String()}
	clusterNode := ClusterResourceNode{
		IsCreateCluster: false, SystemID: iam.SystemIDBKBCS,
		ProjectID: projectID, ClusterID: clusterID}.BuildResourceNodes()
	projectNode := ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.BuildResourceNodes()

	perms, err := bcp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, [][]iam.ResourceNode{clusterNode, projectNode})
	if err != nil {
		return false, "", err
	}
	blog.V(4).Infof("BCSClusterPerm CanViewCluster user[%s] %+v", user, perms)

	allow, err := CheckResourcePerms(CheckResourceRequest{
		Module:    utils.BCSClusterModule,
		Operation: utils.CanViewClusterOperation,
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
		[]string{iam.ClusterView.String()}, []ProjectClusterData{
			{
				Project: projectID,
				Cluster: clusterID,
			},
		})
	projectApp := BuildProjectApplicationInstance(ProjectApplicationAction{
		IsCreateProject: false,
		ActionID:        iam.ProjectView.String(),
		Data:            []string{projectID},
	})

	url, err := bcp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, append(clusterApps, projectApp))
	return allow, url, nil
}

// GenerateIAMApplicationURL build permission URL
func (bcp *BCSClusterPerm) GenerateIAMApplicationURL(systemID string, applications []iam.ApplicationAction) (string, error) {
	if bcp == nil {
		return iam.IamAppURL, ErrServerNotInited
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
		return nil, ErrServerNotInited
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
		return nil, ErrServerNotInited
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
