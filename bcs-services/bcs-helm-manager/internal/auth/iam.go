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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
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
