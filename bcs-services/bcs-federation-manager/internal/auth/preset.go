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
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
)

// NoAuthMethod method for no auth
var NoAuthMethod = []string{
	// federation topology
	"FederationManager.ListFederationClusterWithSubcluster",
	"FederationManager.ListFederationClusterWithNamespace",

	// tasks
	"FederationManager.ListTasks",
	"FederationManager.GetTask",
	"FederationManager.RetryTask",
	"FederationManager.GetTaskRecord",
}

// ActionPermissions action 对应权限中心的权限
var ActionPermissions = map[string]string{
	// federation cluster build
	"FederationManager.InstallFederation":   cluster.CanManageClusterOperation,
	"FederationManager.UnInstallFederation": cluster.CanManageClusterOperation,
	"FederationManager.RegisterSubcluster":  cluster.CanManageClusterOperation,
	"FederationManager.RemoveSubcluster":    cluster.CanManageClusterOperation,

	// federation cluster query
	"FederationManager.GetFederationCluster":            cluster.CanViewClusterOperation,
	"FederationManager.GetFederationByHostCluster":      cluster.CanViewClusterOperation,
	"FederationManager.ListProjectInstallingFederation": project.CanViewProjectOperation,
	"FederationManager.ListProjectFederation":           project.CanViewProjectOperation,

	// namespace manage federation cluster namespace should be able to manage federation cluster
	"FederationManager.CreateFederationClusterNamespace": cluster.CanManageClusterOperation,
	"FederationManager.UpdateFederationClusterNamespace": cluster.CanManageClusterOperation,
	"FederationManager.DeleteFederationClusterNamespace": cluster.CanManageClusterOperation,
	"FederationManager.GetFederationClusterNamespace":    namespace.CanViewNamespaceOperation,
	"FederationManager.ListFederationClusterNamespace":   namespace.CanListNamespaceOperation,

	// quota manage federation cluster namespace quota should be able to manage federation cluster
	"FederationManager.CreateFederationClusterNamespaceQuota": cluster.CanManageClusterOperation,
	"FederationManager.UpdateFederationClusterNamespaceQuota": cluster.CanManageClusterOperation,
	"FederationManager.DeleteFederationClusterNamespaceQuota": cluster.CanManageClusterOperation,
	"FederationManager.GetFederationClusterNamespaceQuota":    namespace.CanViewNamespaceOperation,
	"FederationManager.ListFederationClusterNamespaceQuota":   namespace.CanListNamespaceOperation,
}
