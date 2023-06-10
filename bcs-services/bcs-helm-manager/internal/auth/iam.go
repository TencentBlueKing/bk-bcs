/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package auth

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	// IAMClient iam client
	IAMClient iam.PermClient
	// ProjectIamClient project iam client
	ProjectIamClient *project.BCSProjectPerm
	// ClusterIamClient cluster iam client
	ClusterIamClient *cluster.BCSClusterPerm
	// NamespaceIamClient namespace iam client
	NamespaceIamClient *namespace.BCSNamespacePerm

	// ProjCodeAnnoKey 项目 Code 在命名空间 Annotations 中的 Key
	ProjCodeAnnoKey = "io.tencent.bcs.projectcode"
)

// InitPermClient new a perm client
func InitPermClient(iamClient iam.PermClient) {
	ProjectIamClient = project.NewBCSProjectPermClient(iamClient)
	ClusterIamClient = cluster.NewBCSClusterPermClient(iamClient)
	NamespaceIamClient = namespace.NewBCSNamespacePermClient(iamClient)
}

// GetUserNamespacePermList get user namespace perm
func GetUserNamespacePermList(username, projectID, clusterID string, namespaces []string) (
	map[string]map[string]interface{}, error) {
	permissions := make(map[string]map[string]interface{})

	actionIDs := []string{namespace.NameSpaceScopedCreate.String(), namespace.NameSpaceScopedView.String(),
		namespace.NameSpaceScopedUpdate.String(), namespace.NameSpaceScopedDelete.String()}
	resourceNodes := make([][]iam.ResourceNode, 0)
	for _, n := range namespaces {
		nsNode := namespace.NamespaceScopedResourceNode{
			SystemID:  iam.SystemIDBKBCS,
			ProjectID: projectID, ClusterID: clusterID, Namespace: utils.CalcIAMNsID(clusterID, n)}.
			BuildResourceNodes()
		resourceNodes = append(resourceNodes, nsNode)
	}

	perms, err := IAMClient.BatchResourceMultiActionsAllowed(actionIDs, iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: username}, resourceNodes)
	if err != nil {
		return nil, err
	}

	for nsID, perm := range perms {
		if permissions[nsID] == nil {
			permissions[nsID] = make(map[string]interface{})
		}
		for action, res := range perm {
			permissions[nsID][action] = res
		}
	}
	return permissions, nil
}

// ReleaseResourcePermCheck 检测用户是否有 release 中资源的创建、更新权限
func ReleaseResourcePermCheck(username, projectCode, projectID, clusterID string, namespaceCreated, clusterScope bool,
	namespaces []string, isShardCluster bool) (bool, string, []utils.ResourceAction, error) {
	// 如果是共享集群，且集群不属于该项目，说明是用户使用共享集群，需要单独鉴权
	cls, err := clustermanager.GetCluster(clusterID)
	if err != nil {
		return false, "", nil, err
	}
	if isShardCluster && cls.ProjectID != projectID {
		if namespaceCreated {
			return false, "", nil, fmt.Errorf("共享集群不支持通过 Helm 创建命名空间")
		}
		if clusterScope {
			return false, "", nil, fmt.Errorf("共享集群不支持通过 Helm 创建集群域资源")
		}
		// 检测命名空间是否属于该项目
		var client *kubernetes.Clientset
		client, err = component.GetK8SClientByClusterID(clusterID)
		if err != nil {
			return false, "", nil, err
		}
		for _, v := range namespaces {
			var ns *corev1.Namespace
			ns, err = client.CoreV1().Namespaces().Get(context.TODO(), v, v1.GetOptions{})
			if err != nil {
				return false, "", nil, err
			}
			if ns.Annotations[ProjCodeAnnoKey] != projectCode {
				return false, "", nil, fmt.Errorf("命名空间 %s 在该共享集群中不属于指定项目", v)
			}
		}
	}
	// related actions
	resources := getPermResources(projectID, clusterID, namespaceCreated, clusterScope, namespaces)

	// build request iam.request resourceNodes
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: username,
	}
	relatedActionIDs := getRelatedActionIDs(projectID, clusterID, namespaceCreated, clusterScope, namespaces)

	// get release permission by iam
	perms, err := IAMClient.BatchResourceMultiActionsAllowed(relatedActionIDs, req, getResourceNodes(
		projectID, clusterID, namespaceCreated, clusterScope, namespaces,
	))
	if err != nil {
		return false, "", nil, err
	}
	blog.V(4).Infof("ReleaseResourcePermCheck user[%s] %+v", username, perms)

	// check release resource perms
	allow, err := utils.CheckResourcePerms(utils.CheckResourceRequest{
		Module:    "BCSReleasePerm",
		Operation: "ReleaseResourcePermCheck",
		User:      username,
	}, resources, perms)
	if err != nil {
		return false, "", nil, err
	}

	if allow {
		return allow, "", nil, nil
	}

	// generate apply url if namespace perm notAllow
	applications := getApplications(projectID, clusterID, namespaceCreated, clusterScope, namespaces)
	url, _ := IAMClient.GetApplyURL(iam.ApplicationRequest{SystemID: iam.SystemIDBKBCS}, applications, iam.BkUser{
		BkUserName: iam.SystemUser,
	})
	return allow, url, resources, nil
}

func getPermResources(projectID, clusterID string, namespaceCreated, clusterScope bool,
	namespaces []string) []utils.ResourceAction {
	resources := []utils.ResourceAction{
		{Resource: projectID, Action: project.ProjectView.String()},
		{Resource: clusterID, Action: cluster.ClusterView.String()},
	}
	if clusterScope {
		resources = append(resources, utils.ResourceAction{Resource: clusterID,
			Action: cluster.ClusterScopedCreate.String()})
		resources = append(resources, utils.ResourceAction{Resource: clusterID,
			Action: cluster.ClusterScopedUpdate.String()})
	}
	if namespaceCreated {
		resources = append(resources, utils.ResourceAction{Resource: clusterID,
			Action: namespace.NameSpaceCreate.String()})
	}
	for _, v := range namespaces {
		namespaceID := utils.CalcIAMNsID(clusterID, v)
		resources = append(resources, utils.ResourceAction{Resource: namespaceID,
			Action: namespace.NameSpaceView.String()})
		resources = append(resources, utils.ResourceAction{Resource: namespaceID,
			Action: namespace.NameSpaceScopedCreate.String()})
		resources = append(resources, utils.ResourceAction{Resource: namespaceID,
			Action: namespace.NameSpaceScopedUpdate.String()})
	}
	return resources
}

func getRelatedActionIDs(projectID, clusterID string, namespaceCreated, clusterScope bool,
	namespaces []string) []string {
	relatedActionIDs := []string{project.ProjectView.String(), cluster.ClusterView.String()}
	if clusterScope {
		relatedActionIDs = append(relatedActionIDs, cluster.ClusterScopedCreate.String())
		relatedActionIDs = append(relatedActionIDs, cluster.ClusterScopedUpdate.String())
	}
	if namespaceCreated {
		relatedActionIDs = append(relatedActionIDs, namespace.NameSpaceCreate.String())
	}
	if len(namespaces) > 0 {
		relatedActionIDs = append(relatedActionIDs, namespace.NameSpaceView.String())
		relatedActionIDs = append(relatedActionIDs, namespace.NameSpaceScopedCreate.String())
		relatedActionIDs = append(relatedActionIDs, namespace.NameSpaceScopedUpdate.String())
	}
	return relatedActionIDs
}

func getResourceNodes(projectID, clusterID string, namespaceCreated, clusterScope bool,
	namespaces []string) [][]iam.ResourceNode {
	nodes := make([][]iam.ResourceNode, 0)
	nodes = append(nodes, project.ProjectResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID}.
		BuildResourceNodes())
	nodes = append(nodes, cluster.ClusterResourceNode{
		SystemID: iam.SystemIDBKBCS, ProjectID: projectID, ClusterID: clusterID}.
		BuildResourceNodes())
	if clusterScope {
		nodes = append(nodes, cluster.ClusterScopedResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID,
			ClusterID: clusterID}.
			BuildResourceNodes())
	}
	if namespaceCreated {
		nodes = append(nodes, namespace.NamespaceResourceNode{
			SystemID: iam.SystemIDBKBCS, IsClusterPerm: true, ProjectID: projectID, ClusterID: clusterID}.
			BuildResourceNodes())
	}
	for _, v := range namespaces {
		namespaceID := utils.CalcIAMNsID(clusterID, v)
		nodes = append(nodes, namespace.NamespaceResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID,
			ClusterID: clusterID, Namespace: namespaceID}.
			BuildResourceNodes())
		nodes = append(nodes, namespace.NamespaceScopedResourceNode{SystemID: iam.SystemIDBKBCS, ProjectID: projectID,
			ClusterID: clusterID, Namespace: namespaceID}.
			BuildResourceNodes())
	}
	return nodes
}

func getApplications(projectID, clusterID string, namespaceCreated, clusterScope bool,
	namespaces []string) []iam.ApplicationAction {
	apps := make([]iam.ApplicationAction, 0)
	apps = append(apps, project.BuildProjectApplicationInstance(project.ProjectApplicationAction{
		ActionID: project.ProjectView.String(),
		Data:     []string{projectID},
	}))
	apps = append(apps, cluster.BuildClusterApplicationInstance(cluster.ClusterApplicationAction{
		ActionID: cluster.ClusterView.String(),
		Data: []cluster.ProjectClusterData{
			{Project: projectID, Cluster: clusterID},
		},
	}))
	if clusterScope {
		apps = append(apps, cluster.BuildClusterScopedApplicationInstance(cluster.ClusterScopedApplicationAction{
			ActionID: cluster.ClusterScopedCreate.String(),
			Data: []cluster.ProjectClusterData{
				{Project: projectID, Cluster: clusterID},
			},
		}))
		apps = append(apps, cluster.BuildClusterScopedApplicationInstance(cluster.ClusterScopedApplicationAction{
			ActionID: cluster.ClusterScopedUpdate.String(),
			Data: []cluster.ProjectClusterData{
				{Project: projectID, Cluster: clusterID},
			},
		}))
	}
	if namespaceCreated {
		apps = append(apps, namespace.BuildNamespaceApplicationInstance(namespace.NamespaceApplicationAction{
			IsClusterPerm: true,
			ActionID:      namespace.NameSpaceCreate.String(),
			Data: []namespace.ProjectNamespaceData{
				{Project: projectID, Cluster: clusterID},
			},
		}))
	}
	for _, v := range namespaces {
		namespaceID := utils.CalcIAMNsID(clusterID, v)
		apps = append(apps, namespace.BuildNamespaceApplicationInstance(namespace.NamespaceApplicationAction{
			ActionID: namespace.NameSpaceView.String(),
			Data: []namespace.ProjectNamespaceData{
				{Project: projectID, Cluster: clusterID, Namespace: namespaceID},
			},
		}))
		apps = append(apps, namespace.BuildNamespaceScopedApplicationInstance(
			namespace.NamespaceScopedApplicationAction{
				ActionID: namespace.NameSpaceScopedCreate.String(),
				Data: []namespace.ProjectNamespaceData{
					{Project: projectID, Cluster: clusterID, Namespace: namespaceID},
				},
			}))
		apps = append(apps, namespace.BuildNamespaceScopedApplicationInstance(
			namespace.NamespaceScopedApplicationAction{
				ActionID: namespace.NameSpaceScopedUpdate.String(),
				Data: []namespace.ProjectNamespaceData{
					{Project: projectID, Cluster: clusterID, Namespace: namespaceID},
				},
			}))
	}
	return apps
}
