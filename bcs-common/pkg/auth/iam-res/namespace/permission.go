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

package namespace

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	blog "k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam-res/cluster"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam-res/project"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam-res/utils"
)

// BCSNamespacePerm namespace perm client
type BCSNamespacePerm struct {
	iamClient iam.PermClient
}

// NewBCSNamespacePermClient init namespace perm client
func NewBCSNamespacePermClient(cli iam.PermClient) *BCSNamespacePerm {
	return &BCSNamespacePerm{iamClient: cli}
}

// GenerateIAMApplicationURL build permission URL
func (bnp *BCSNamespacePerm) GenerateIAMApplicationURL(systemID string, applications []iam.ApplicationAction) (string,
	error) {
	url, err := bnp.iamClient.GetApplyURL(iam.ApplicationRequest{SystemID: systemID}, applications, iam.BkUser{
		BkUserName: iam.SystemUser,
	})
	if err != nil {
		return iam.IamAppURL, err
	}

	return url, nil
}

// CanCreateNamespace check user createNamespace perm
func (bnp *BCSNamespacePerm) CanCreateNamespace(user,
	projectID, clusterID string, isSharedCluster bool) (bool, string, []utils.ResourceAction, error) {
	// related actions
	resources := []utils.ResourceAction{
		{Type: string(project.SysProject), Resource: projectID, Action: project.ProjectView.String()},
		{Type: string(SysNamespace), Resource: clusterID, Action: NameSpaceCreate.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{project.ProjectView.String(), NameSpaceCreate.String()}
	projectNode := project.ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.
		BuildResourceNodes()
	clusterNode := cluster.ClusterResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID, ClusterID: clusterID}.
		BuildResourceNodes()
	namespaceNode := NamespaceResourceNode{IsClusterPerm: true, ProjectID: projectID, ClusterID: clusterID}.
		BuildResourceNodes()

	resourceNodes := [][]iam.ResourceNode{projectNode, namespaceNode}
	if !isSharedCluster {
		// ignore cluster_view permission for namepsace action in shared cluster
		resources = append(resources, utils.ResourceAction{
			Type: string(cluster.SysCluster), Resource: clusterID, Action: cluster.ClusterView.String()})
		relatedActionIDs = append(relatedActionIDs, cluster.ClusterView.String())
		resourceNodes = append(resourceNodes, clusterNode)
	}

	// get namespace permission by iam
	perms, err := bnp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, resourceNodes)
	if err != nil {
		return false, "", nil, err
	}
	blog.V(4).Infof("BCSNamespacePerm CanCreateNamespace user[%s] %+v", user, perms)

	// check namespace resource perms
	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    BCSNamespaceModule,
		Operation: CanCreateNamespaceOperation,
		User:      user,
	}, resources, perms)
	if err != nil {
		return false, "", nil, err
	}

	if allow {
		return allow, "", resources, nil
	}

	// generate apply url if namespace perm notAllow
	projectApp := project.BuildProjectApplicationInstance(project.ProjectApplicationAction{
		ActionID: project.ProjectView.String(),
		Data:     []string{projectID},
	})
	clusterApp := cluster.BuildClusterApplicationInstance(cluster.ClusterApplicationAction{
		ActionID: cluster.ClusterView.String(),
		Data: []cluster.ProjectClusterData{
			{
				Project: projectID,
				Cluster: clusterID,
			},
		},
	})
	nsApp := BuildNamespaceApplicationInstance(NamespaceApplicationAction{
		IsClusterPerm: true,
		ActionID:      NameSpaceCreate.String(),
	})
	apps := []iam.ApplicationAction{projectApp, nsApp}
	if !isSharedCluster {
		apps = append(apps, clusterApp)
	}

	url, err := bnp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, apps)
	if err != nil {
		return false, "", nil, err
	}
	return allow, url, resources, nil
}

// CanViewNamespace check user viewNamespace perm
func (bnp *BCSNamespacePerm) CanViewNamespace(user,
	projectID, clusterID, namespace string, isSharedCluster bool) (bool, string, []utils.ResourceAction, error) {
	namespaceID := utils.CalcIAMNsID(clusterID, namespace)
	// related actions
	resources := []utils.ResourceAction{
		{Type: string(project.SysProject), Resource: projectID, Action: project.ProjectView.String()},
		{Type: string(SysNamespace), Resource: namespaceID, Action: NameSpaceView.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{project.ProjectView.String(), NameSpaceView.String()}
	projectNode := project.ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.
		BuildResourceNodes()
	clusterNode := cluster.ClusterResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID, ClusterID: clusterID}.
		BuildResourceNodes()
	namespaceNode := NamespaceResourceNode{ProjectID: projectID, ClusterID: clusterID, Namespace: namespaceID}.
		BuildResourceNodes()
	resourceNodes := [][]iam.ResourceNode{projectNode, namespaceNode}
	if !isSharedCluster {
		// ignore cluster_view permission for namepsace action in shared cluster
		resources = append(resources, utils.ResourceAction{
			Type: string(cluster.SysCluster), Resource: clusterID, Action: cluster.ClusterView.String()})
		relatedActionIDs = append(relatedActionIDs, cluster.ClusterView.String())
		resourceNodes = append(resourceNodes, clusterNode)
	}

	// get namespace permission by iam
	perms, err := bnp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, resourceNodes)
	if err != nil {
		return false, "", nil, err
	}
	blog.V(4).Infof("BCSNamespacePerm CanViewNamespace user[%s] %+v", user, perms)

	// check namespace resource perms
	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    BCSNamespaceModule,
		Operation: CanViewNamespaceOperation,
		User:      user,
	}, resources, perms)
	if err != nil {
		return false, "", nil, err
	}

	if allow {
		return allow, "", nil, nil
	}

	// generate apply url if namespace perm notAllow
	projectApp := project.BuildProjectApplicationInstance(project.ProjectApplicationAction{
		ActionID: project.ProjectView.String(),
		Data:     []string{projectID},
	})
	clusterApp := cluster.BuildClusterApplicationInstance(cluster.ClusterApplicationAction{
		ActionID: cluster.ClusterView.String(),
		Data: []cluster.ProjectClusterData{
			{Project: projectID, Cluster: clusterID},
		},
	})
	nsApp := BuildNamespaceApplicationInstance(NamespaceApplicationAction{
		ActionID: NameSpaceView.String(),
		Data: []ProjectNamespaceData{
			{Project: projectID, Cluster: clusterID, Namespace: namespaceID},
		},
	})
	apps := []iam.ApplicationAction{projectApp, nsApp}
	if !isSharedCluster {
		apps = append(apps, clusterApp)
	}

	url, err := bnp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, apps)
	if err != nil {
		return false, "", nil, err
	}
	return allow, url, resources, nil
}

// CanListNamespace check user listNamespace perm
func (bnp *BCSNamespacePerm) CanListNamespace(user,
	projectID, clusterID string, isSharedCluster bool) (bool, string, []utils.ResourceAction, error) {
	// related actions
	resources := []utils.ResourceAction{
		{Type: string(project.SysProject), Resource: projectID, Action: project.ProjectView.String()},
		{Type: string(SysNamespace), Resource: clusterID, Action: NameSpaceList.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{project.ProjectView.String(), NameSpaceList.String()}
	projectNode := project.ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.
		BuildResourceNodes()
	clusterNode := cluster.ClusterResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID, ClusterID: clusterID}.
		BuildResourceNodes()
	namespaceNode := NamespaceResourceNode{IsClusterPerm: true, ProjectID: projectID, ClusterID: clusterID}.
		BuildResourceNodes()

	resourceNodes := [][]iam.ResourceNode{projectNode, namespaceNode}
	if !isSharedCluster {
		// ignore cluster_view permission for namepsace action in shared cluster
		resources = append(resources, utils.ResourceAction{
			Type: string(cluster.SysCluster), Resource: clusterID, Action: cluster.ClusterView.String()})
		relatedActionIDs = append(relatedActionIDs, cluster.ClusterView.String())
		resourceNodes = append(resourceNodes, clusterNode)
	}

	// get namespace permission by iam
	perms, err := bnp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, resourceNodes)
	if err != nil {
		return false, "", nil, err
	}
	blog.V(4).Infof("BCSNamespacePerm CanListNamespace user[%s] %+v", user, perms)

	// check namespace resource perms
	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    BCSNamespaceModule,
		Operation: CanListNamespaceOperation,
		User:      user,
	}, resources, perms)
	if err != nil {
		return false, "", nil, err
	}

	if allow {
		return allow, "", nil, nil
	}

	// generate apply url if namespace perm notAllow
	projectApp := project.BuildProjectApplicationInstance(project.ProjectApplicationAction{
		ActionID: project.ProjectView.String(),
		Data:     []string{projectID},
	})
	clusterApp := cluster.BuildClusterApplicationInstance(cluster.ClusterApplicationAction{
		ActionID: cluster.ClusterView.String(),
		Data: []cluster.ProjectClusterData{
			{Project: projectID, Cluster: clusterID},
		},
	})
	nsApp := BuildNamespaceApplicationInstance(NamespaceApplicationAction{
		ActionID:      NameSpaceList.String(),
		IsClusterPerm: true,
	})
	apps := []iam.ApplicationAction{projectApp, nsApp}
	if !isSharedCluster {
		apps = append(apps, clusterApp)
	}

	url, err := bnp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, apps)
	if err != nil {
		return false, "", nil, err
	}
	return allow, url, resources, nil
}

// CanUpdateNamespace check user updateNamespace perm
func (bnp *BCSNamespacePerm) CanUpdateNamespace(user,
	projectID, clusterID, namespace string, isSharedCluster bool) (bool, string, []utils.ResourceAction, error) {
	namespaceID := utils.CalcIAMNsID(clusterID, namespace)
	// related actions
	resources := []utils.ResourceAction{
		{Type: string(project.SysProject), Resource: projectID, Action: project.ProjectView.String()},
		{Type: string(SysNamespace), Resource: namespaceID, Action: NameSpaceUpdate.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{project.ProjectView.String(), NameSpaceUpdate.String()}
	projectNode := project.ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.
		BuildResourceNodes()
	clusterNode := cluster.ClusterResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID, ClusterID: clusterID}.
		BuildResourceNodes()
	namespaceNode := NamespaceResourceNode{ProjectID: projectID, ClusterID: clusterID, Namespace: namespaceID}.
		BuildResourceNodes()
	resourceNodes := [][]iam.ResourceNode{projectNode, namespaceNode}
	if !isSharedCluster {
		// ignore cluster_view permission for namepsace action in shared cluster
		resources = append(resources, utils.ResourceAction{
			Type: string(cluster.SysCluster), Resource: clusterID, Action: cluster.ClusterView.String()})
		relatedActionIDs = append(relatedActionIDs, cluster.ClusterView.String())
		resourceNodes = append(resourceNodes, clusterNode)
	}

	// get namespace permission by iam
	perms, err := bnp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, resourceNodes)
	if err != nil {
		return false, "", nil, err
	}
	blog.V(4).Infof("BCSNamespacePerm CanUpdateNamespace user[%s] %+v", user, perms)

	// check namespace resource perms
	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    BCSNamespaceModule,
		Operation: CanUpdateNamespaceOperation,
		User:      user,
	}, resources, perms)
	if err != nil {
		return false, "", nil, err
	}

	if allow {
		return allow, "", nil, nil
	}

	// generate apply url if namespace perm notAllow
	projectApp := project.BuildProjectApplicationInstance(project.ProjectApplicationAction{
		ActionID: project.ProjectView.String(),
		Data:     []string{projectID},
	})
	clusterApp := cluster.BuildClusterApplicationInstance(cluster.ClusterApplicationAction{
		ActionID: cluster.ClusterView.String(),
		Data: []cluster.ProjectClusterData{
			{Project: projectID, Cluster: clusterID},
		},
	})
	nsApp := BuildNamespaceApplicationInstance(NamespaceApplicationAction{
		ActionID: NameSpaceUpdate.String(),
		Data: []ProjectNamespaceData{
			{Project: projectID, Cluster: clusterID, Namespace: namespaceID},
		},
	})
	apps := []iam.ApplicationAction{projectApp, nsApp}
	if !isSharedCluster {
		apps = append(apps, clusterApp)
	}

	url, err := bnp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, apps)
	if err != nil {
		return false, "", nil, err
	}
	return allow, url, resources, nil
}

// CanDeleteNamespace check user deleteNamespace perm
func (bnp *BCSNamespacePerm) CanDeleteNamespace(user,
	projectID, clusterID, namespace string, isSharedCluster bool) (bool, string, []utils.ResourceAction, error) {
	namespaceID := utils.CalcIAMNsID(clusterID, namespace)
	// related actions
	resources := []utils.ResourceAction{
		{Type: string(project.SysProject), Resource: projectID, Action: project.ProjectView.String()},
		{Type: string(SysNamespace), Resource: namespaceID, Action: NameSpaceDelete.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{project.ProjectView.String(), NameSpaceDelete.String()}
	projectNode := project.ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.
		BuildResourceNodes()
	clusterNode := cluster.ClusterResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID, ClusterID: clusterID}.
		BuildResourceNodes()
	namespaceNode := NamespaceResourceNode{ProjectID: projectID, ClusterID: clusterID, Namespace: namespaceID}.
		BuildResourceNodes()
	resourceNodes := [][]iam.ResourceNode{projectNode, namespaceNode}
	if !isSharedCluster {
		// ignore cluster_view permission for namepsace action in shared cluster
		resources = append(resources, utils.ResourceAction{
			Type: string(cluster.SysCluster), Resource: clusterID, Action: cluster.ClusterView.String()})
		relatedActionIDs = append(relatedActionIDs, cluster.ClusterView.String())
		resourceNodes = append(resourceNodes, clusterNode)
	}

	// get namespace permission by iam
	perms, err := bnp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, resourceNodes)
	if err != nil {
		return false, "", nil, err
	}
	blog.V(4).Infof("BCSNamespacePerm CanDeleteNamespace user[%s] %+v", user, perms)

	// check namespace resource perms
	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    BCSNamespaceModule,
		Operation: CanDeleteNamespaceOperation,
		User:      user,
	}, resources, perms)
	if err != nil {
		return false, "", nil, err
	}

	if allow {
		return allow, "", nil, nil
	}

	// generate apply url if namespace perm notAllow
	projectApp := project.BuildProjectApplicationInstance(project.ProjectApplicationAction{
		ActionID: project.ProjectView.String(),
		Data:     []string{projectID},
	})
	clusterApp := cluster.BuildClusterApplicationInstance(cluster.ClusterApplicationAction{
		ActionID: cluster.ClusterView.String(),
		Data: []cluster.ProjectClusterData{
			{Project: projectID, Cluster: clusterID},
		},
	})
	nsApp := BuildNamespaceApplicationInstance(NamespaceApplicationAction{
		ActionID: NameSpaceDelete.String(),
		Data: []ProjectNamespaceData{
			{Project: projectID, Cluster: clusterID, Namespace: namespaceID},
		},
	})
	apps := []iam.ApplicationAction{projectApp, nsApp}
	if !isSharedCluster {
		apps = append(apps, clusterApp)
	}

	url, err := bnp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, apps)
	if err != nil {
		return false, "", nil, err
	}
	return allow, url, resources, nil
}

// CanCreateNamespaceScopedResource check user createNamespaceScopedResource perm
func (bnp *BCSNamespacePerm) CanCreateNamespaceScopedResource(user, projectID, clusterID, namespace string) (bool,
	string, []utils.ResourceAction, error) {
	namespaceID := utils.CalcIAMNsID(clusterID, namespace)
	// related actions
	resources := []utils.ResourceAction{
		{Resource: projectID, Action: project.ProjectView.String()},
		{Resource: clusterID, Action: cluster.ClusterView.String()},
		{Resource: namespaceID, Action: NameSpaceView.String()},
		{Resource: namespaceID, Action: NameSpaceScopedCreate.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{project.ProjectView.String(), cluster.ClusterView.String(), NameSpaceView.String(),
		NameSpaceScopedCreate.String(),
	}
	projectNode := project.ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.
		BuildResourceNodes()
	clusterNode := cluster.ClusterResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID, ClusterID: clusterID}.
		BuildResourceNodes()
	namespaceNode := NamespaceResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID, ClusterID: clusterID,
		Namespace: namespaceID}.
		BuildResourceNodes()
	namespaceScopedNode := NamespaceScopedResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID,
		ClusterID: clusterID, Namespace: namespaceID}.
		BuildResourceNodes()

	// get namespace permission by iam
	perms, err := bnp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, [][]iam.ResourceNode{
		projectNode, clusterNode, namespaceNode, namespaceScopedNode})
	if err != nil {
		return false, "", nil, err
	}
	blog.V(4).Infof("BCSNamespacePerm CanCreateNamespaceScopedResource user[%s] %+v", user, perms)

	// check namespace resource perms
	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    BCSNamespaceModule,
		Operation: CanCreateNamespaceScopedResourceOperation,
		User:      user,
	}, resources, perms)
	if err != nil {
		return false, "", nil, err
	}

	if allow {
		return allow, "", nil, nil
	}

	// generate apply url if namespace perm notAllow
	projectApp := project.BuildProjectApplicationInstance(project.ProjectApplicationAction{
		ActionID: project.ProjectView.String(),
		Data:     []string{projectID},
	})
	clusterApp := cluster.BuildClusterApplicationInstance(cluster.ClusterApplicationAction{
		ActionID: cluster.ClusterView.String(),
		Data:     []cluster.ProjectClusterData{{Project: projectID, Cluster: clusterID}},
	})
	nsApp := BuildNamespaceApplicationInstance(NamespaceApplicationAction{
		ActionID: NameSpaceView.String(),
		Data:     []ProjectNamespaceData{{Project: projectID, Cluster: clusterID, Namespace: namespaceID}},
	})
	nssApp := BuildNSScopedAppInstance(NamespaceScopedApplicationAction{
		ActionID: NameSpaceScopedCreate.String(),
		Data:     []ProjectNamespaceData{{Project: projectID, Cluster: clusterID, Namespace: namespaceID}},
	})

	url, _ := bnp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, []iam.ApplicationAction{
		clusterApp, projectApp, nsApp, nssApp,
	})
	return allow, url, resources, nil
}

// CanViewNamespaceScopedResource check user viewNamespaceScopedResource perm
func (bnp *BCSNamespacePerm) CanViewNamespaceScopedResource(user, projectID, clusterID, namespace string) (bool,
	string, []utils.ResourceAction, error) {
	namespaceID := utils.CalcIAMNsID(clusterID, namespace)
	// related actions
	resources := []utils.ResourceAction{
		{Resource: projectID, Action: project.ProjectView.String()},
		{Resource: clusterID, Action: cluster.ClusterView.String()},
		{Resource: namespaceID, Action: NameSpaceView.String()},
		{Resource: namespaceID, Action: NameSpaceScopedView.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{project.ProjectView.String(), cluster.ClusterView.String(), NameSpaceView.String(),
		NameSpaceScopedView.String(),
	}
	projectNode := project.ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.
		BuildResourceNodes()
	clusterNode := cluster.ClusterResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID, ClusterID: clusterID}.
		BuildResourceNodes()
	namespaceNode := NamespaceResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID, ClusterID: clusterID,
		Namespace: namespaceID}.
		BuildResourceNodes()
	namespaceScopedNode := NamespaceScopedResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID,
		ClusterID: clusterID, Namespace: namespaceID}.
		BuildResourceNodes()

	// get namespace permission by iam
	perms, err := bnp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, [][]iam.ResourceNode{
		projectNode, clusterNode, namespaceNode, namespaceScopedNode})
	if err != nil {
		return false, "", nil, err
	}
	blog.V(4).Infof("BCSNamespacePerm CanViewNamespaceScopedResource user[%s] %+v", user, perms)

	// check namespace resource perms
	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    BCSNamespaceModule,
		Operation: CanViewNamespaceScopedResourceOperation,
		User:      user,
	}, resources, perms)
	if err != nil {
		return false, "", nil, err
	}

	if allow {
		return allow, "", nil, nil
	}

	// generate apply url if namespace perm notAllow
	projectApp := project.BuildProjectApplicationInstance(project.ProjectApplicationAction{
		ActionID: project.ProjectView.String(),
		Data:     []string{projectID},
	})
	clusterApp := cluster.BuildClusterApplicationInstance(cluster.ClusterApplicationAction{
		ActionID: cluster.ClusterView.String(),
		Data: []cluster.ProjectClusterData{
			{Project: projectID, Cluster: clusterID},
		},
	})
	nsApp := BuildNamespaceApplicationInstance(NamespaceApplicationAction{
		ActionID: NameSpaceView.String(),
		Data:     []ProjectNamespaceData{{Project: projectID, Cluster: clusterID, Namespace: namespaceID}},
	})
	nssApp := BuildNSScopedAppInstance(NamespaceScopedApplicationAction{
		ActionID: NameSpaceScopedView.String(),
		Data:     []ProjectNamespaceData{{Project: projectID, Cluster: clusterID, Namespace: namespaceID}},
	})

	url, _ := bnp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, []iam.ApplicationAction{
		clusterApp, projectApp, nsApp, nssApp,
	})
	return allow, url, resources, nil
}

// CanUpdateNamespaceScopedResource check user updateNamespaceScopedResource perm
func (bnp *BCSNamespacePerm) CanUpdateNamespaceScopedResource(user, projectID, clusterID, namespace string) (bool,
	string, []utils.ResourceAction, error) {
	namespaceID := utils.CalcIAMNsID(clusterID, namespace)
	// related actions
	resources := []utils.ResourceAction{
		{Resource: projectID, Action: project.ProjectView.String()},
		{Resource: clusterID, Action: cluster.ClusterView.String()},
		{Resource: namespaceID, Action: NameSpaceView.String()},
		{Resource: namespaceID, Action: NameSpaceScopedUpdate.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{project.ProjectView.String(), cluster.ClusterView.String(), NameSpaceView.String(),
		NameSpaceScopedUpdate.String(),
	}
	projectNode := project.ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.
		BuildResourceNodes()
	clusterNode := cluster.ClusterResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID, ClusterID: clusterID}.
		BuildResourceNodes()
	namespaceNode := NamespaceResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID, ClusterID: clusterID,
		Namespace: namespaceID}.
		BuildResourceNodes()
	namespaceScopedNode := NamespaceScopedResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID,
		ClusterID: clusterID, Namespace: namespaceID}.
		BuildResourceNodes()

	// get namespace permission by iam
	perms, err := bnp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, [][]iam.ResourceNode{
		projectNode, clusterNode, namespaceNode, namespaceScopedNode})
	if err != nil {
		return false, "", nil, err
	}
	blog.V(4).Infof("BCSNamespacePerm CanUpdateNamespaceScopedResource user[%s] %+v", user, perms)

	// check namespace resource perms
	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    BCSNamespaceModule,
		Operation: CanUpdateNamespaceScopedResourceOperation,
		User:      user,
	}, resources, perms)
	if err != nil {
		return false, "", nil, err
	}

	if allow {
		return allow, "", nil, nil
	}

	// generate apply url if namespace perm notAllow
	projectApp := project.BuildProjectApplicationInstance(project.ProjectApplicationAction{
		ActionID: project.ProjectView.String(),
		Data:     []string{projectID},
	})
	clusterApp := cluster.BuildClusterApplicationInstance(cluster.ClusterApplicationAction{
		ActionID: cluster.ClusterView.String(),
		Data: []cluster.ProjectClusterData{
			{Project: projectID, Cluster: clusterID},
		},
	})
	nsApp := BuildNamespaceApplicationInstance(NamespaceApplicationAction{
		ActionID: NameSpaceView.String(),
		Data:     []ProjectNamespaceData{{Project: projectID, Cluster: clusterID, Namespace: namespaceID}},
	})
	nssApp := BuildNSScopedAppInstance(NamespaceScopedApplicationAction{
		ActionID: NameSpaceScopedUpdate.String(),
		Data:     []ProjectNamespaceData{{Project: projectID, Cluster: clusterID, Namespace: namespaceID}},
	})

	url, _ := bnp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, []iam.ApplicationAction{
		clusterApp, projectApp, nsApp, nssApp,
	})
	return allow, url, resources, nil
}

// CanDeleteNamespaceScopedResource check user deleteNamespaceScopedResource perm
func (bnp *BCSNamespacePerm) CanDeleteNamespaceScopedResource(user, projectID, clusterID, namespace string) (bool,
	string, []utils.ResourceAction, error) {
	namespaceID := utils.CalcIAMNsID(clusterID, namespace)
	// related actions
	resources := []utils.ResourceAction{
		{Resource: projectID, Action: project.ProjectView.String()},
		{Resource: clusterID, Action: cluster.ClusterView.String()},
		{Resource: namespaceID, Action: NameSpaceView.String()},
		{Resource: namespaceID, Action: NameSpaceScopedDelete.String()},
	}

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	relatedActionIDs := []string{project.ProjectView.String(), cluster.ClusterView.String(), NameSpaceView.String(),
		NameSpaceScopedDelete.String(),
	}
	projectNode := project.ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.
		BuildResourceNodes()
	clusterNode := cluster.ClusterResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID, ClusterID: clusterID}.
		BuildResourceNodes()
	namespaceNode := NamespaceResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID, ClusterID: clusterID,
		Namespace: namespaceID}.
		BuildResourceNodes()
	namespaceScopedNode := NamespaceScopedResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID,
		ClusterID: clusterID, Namespace: namespaceID}.
		BuildResourceNodes()

	// get namespace permission by iam
	perms, err := bnp.iamClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, [][]iam.ResourceNode{
		projectNode, clusterNode, namespaceNode, namespaceScopedNode})
	if err != nil {
		return false, "", nil, err
	}
	blog.V(4).Infof("BCSNamespacePerm CanDeleteNamespaceScopedResource user[%s] %+v", user, perms)

	// check namespace resource perms
	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    BCSNamespaceModule,
		Operation: CanDeleteNamespaceScopedResourceOperation,
		User:      user,
	}, resources, perms)
	if err != nil {
		return false, "", nil, err
	}

	if allow {
		return allow, "", nil, nil
	}

	// generate apply url if namespace perm notAllow
	projectApp := project.BuildProjectApplicationInstance(project.ProjectApplicationAction{
		ActionID: project.ProjectView.String(),
		Data:     []string{projectID},
	})
	clusterApp := cluster.BuildClusterApplicationInstance(cluster.ClusterApplicationAction{
		ActionID: cluster.ClusterView.String(),
		Data:     []cluster.ProjectClusterData{{Project: projectID, Cluster: clusterID}},
	})
	nsApp := BuildNamespaceApplicationInstance(NamespaceApplicationAction{
		ActionID: NameSpaceView.String(),
		Data:     []ProjectNamespaceData{{Project: projectID, Cluster: clusterID, Namespace: namespaceID}},
	})
	nssApp := BuildNSScopedAppInstance(NamespaceScopedApplicationAction{
		ActionID: NameSpaceScopedDelete.String(),
		Data:     []ProjectNamespaceData{{Project: projectID, Cluster: clusterID, Namespace: namespaceID}},
	})

	url, _ := bnp.GenerateIAMApplicationURL(iam.SystemIDBKBCS, []iam.ApplicationAction{
		clusterApp, projectApp, nsApp, nssApp,
	})
	return allow, url, resources, nil
}

// GetMultiNamespaceMultiActionPerm only support same instanceSelection
func (bnp *BCSNamespacePerm) GetMultiNamespaceMultiActionPerm(user string, namespaces []ProjectNamespaceData,
	actionIDs []string) (map[string]map[string]bool, error) {
	if bnp == nil {
		return nil, utils.ErrServerNotInited
	}

	resourceNodes := make([][]iam.ResourceNode, 0)
	for i := range namespaces {
		namespaceID := utils.CalcIAMNsID(namespaces[i].Cluster, namespaces[i].Namespace)
		namespaceNode := NamespaceResourceNode{
			// IsClusterPerm: true,
			SystemID:  iam.SystemIDBKBCS,
			ProjectID: namespaces[i].Project,
			ClusterID: namespaces[i].Cluster,
			Namespace: namespaceID,
		}.BuildResourceNodes()
		resourceNodes = append(resourceNodes, namespaceNode)
	}

	return bnp.iamClient.BatchResourceMultiActionsAllowed(actionIDs, iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user}, resourceNodes)
}
