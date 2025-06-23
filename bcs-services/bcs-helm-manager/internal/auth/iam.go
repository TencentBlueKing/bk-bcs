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

package auth

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/options"
)

var (
	// IAMClient iam client
	IAMClient func(tenantID string) iam.PermClient
	// ProjectIamClient project iam client
	ProjectIamClient func(tenantID string) *project.BCSProjectPerm
	// ClusterIamClient cluster iam client
	ClusterIamClient func(tenantID string) *cluster.BCSClusterPerm
	// NamespaceIamClient namespace iam client
	NamespaceIamClient func(tenantID string) *namespace.BCSNamespacePerm
)

// InitPermClient new a perm client
func InitPermClient() {
	ProjectIamClient = func(tenantID string) *project.BCSProjectPerm {
		return project.NewBCSProjectPermClient(IAMClient(tenantID))
	}
	ClusterIamClient = func(tenantID string) *cluster.BCSClusterPerm {
		return cluster.NewBCSClusterPermClient(IAMClient(tenantID))
	}
	NamespaceIamClient = func(tenantID string) *namespace.BCSNamespacePerm {
		return namespace.NewBCSNamespacePermClient(IAMClient(tenantID))
	}
}

// GetUserNamespacePermList get user namespace perm
func GetUserNamespacePermList(username, projectID, clusterID string, namespaces []string, tenantID string) (
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

	if !options.GlobalOptions.JWT.Enable {
		for _, v := range namespaces {
			nsID := utils.CalcIAMNsID(clusterID, v)
			permissions[nsID] = make(map[string]interface{})
			permissions[nsID][namespace.NameSpaceScopedCreate.String()] = true
			permissions[nsID][namespace.NameSpaceScopedView.String()] = true
			permissions[nsID][namespace.NameSpaceScopedUpdate.String()] = true
			permissions[nsID][namespace.NameSpaceScopedDelete.String()] = true
		}
		return permissions, nil
	}
	perms, err := IAMClient(tenantID).BatchResourceMultiActionsAllowed(actionIDs, iam.PermissionRequest{
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
func ReleaseResourcePermCheck(projectCode, clusterID string, namespaceCreated, clusterScope bool,
	namespaces []string) (bool, string, []utils.ResourceAction, error) {
	if namespaceCreated {
		return false, "", nil, fmt.Errorf("共享集群不支持通过 Helm 创建命名空间")
	}
	if clusterScope {
		return false, "", nil, fmt.Errorf("共享集群不支持通过 Helm 创建集群域资源")
	}
	// 检测命名空间是否属于该项目
	var client *kubernetes.Clientset
	var err error
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
		if ns.Annotations[options.GlobalOptions.SharedCluster.AnnotationKeyProjCode] != projectCode {
			return false, "", nil, fmt.Errorf("命名空间 %s 在该共享集群中不属于指定项目", v)
		}
	}
	return true, "", nil, nil
}

func getRelatedActionIDs(projectID, clusterID string, namespaceCreated, clusterScope bool, // nolint
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
