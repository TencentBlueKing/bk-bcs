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
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cloudaccount"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
)

// ActionPermissions action 对应权限中心的权限
var ActionPermissions = map[string]string{
	// cluster
	"ClusterManager.CreateCluster":          cluster.CanCreateClusterOperation,
	"ClusterManager.RetryCreateClusterTask": cluster.CanCreateClusterOperation,
	"ClusterManager.ImportCluster":          cluster.CanCreateClusterOperation,
	"ClusterManager.UpdateCluster":          cluster.CanManageClusterOperation,
	"ClusterManager.AddNodesToCluster":      cluster.CanManageClusterOperation,
	"ClusterManager.DeleteNodesFromCluster": cluster.CanManageClusterOperation,
	"ClusterManager.ListNodesInCluster":     cluster.CanViewClusterOperation,
	"ClusterManager.ListMastersInCluster":   cluster.CanViewClusterOperation,
	"ClusterManager.DeleteCluster":          cluster.CanDeleteClusterOperation,
	"ClusterManager.GetCluster":             cluster.CanViewClusterOperation,
	"ClusterManager.ListCluster":            project.CanViewProjectOperation,
	"ClusterManager.ListProjectCluster":     project.CanViewProjectOperation,

	// node
	"ClusterManager.GetNode":          cluster.CanViewClusterOperation,
	"ClusterManager.UpdateNode":       cluster.CanManageClusterOperation,
	"ClusterManager.CordonNode":       cluster.CanManageClusterOperation,
	"ClusterManager.UnCordonNode":     cluster.CanManageClusterOperation,
	"ClusterManager.DrainNode":        cluster.CanManageClusterOperation,
	"ClusterManager.UpdateNodeLabels": cluster.CanManageClusterOperation,
	"ClusterManager.UpdateNodeTaints": cluster.CanManageClusterOperation,
	// cluster credential
	"ClusterManager.GetClusterCredential":    cluster.CanManageClusterOperation,
	"ClusterManager.UpdateClusterCredential": cluster.CanManageClusterOperation,
	"ClusterManager.DeleteClusterCredential": cluster.CanManageClusterOperation,
	"ClusterManager.ListClusterCredential":   cluster.CanManageClusterOperation,
	// nodeGroup
	"ClusterManager.CreateNodeGroup":           cluster.CanManageClusterOperation,
	"ClusterManager.UpdateNodeGroup":           cluster.CanManageClusterOperation,
	"ClusterManager.DeleteNodeGroup":           cluster.CanManageClusterOperation,
	"ClusterManager.GetNodeGroup":              cluster.CanViewClusterOperation,
	"ClusterManager.ListNodeGroup":             cluster.CanViewClusterOperation,
	"ClusterManager.MoveNodesToGroup":          cluster.CanManageClusterOperation,
	"ClusterManager.RemoveNodesFromGroup":      cluster.CanManageClusterOperation,
	"ClusterManager.CleanNodesInGroup":         cluster.CanManageClusterOperation,
	"ClusterManager.CleanNodesInGroupV2":       cluster.CanManageClusterOperation,
	"ClusterManager.ListNodesInGroup":          cluster.CanViewClusterOperation,
	"ClusterManager.UpdateGroupDesiredNode":    cluster.CanManageClusterOperation,
	"ClusterManager.UpdateGroupDesiredSize":    cluster.CanManageClusterOperation,
	"ClusterManager.UpdateGroupMinMaxSize":     cluster.CanManageClusterOperation,
	"ClusterManager.EnableNodeGroupAutoScale":  cluster.CanManageClusterOperation,
	"ClusterManager.DisableNodeGroupAutoScale": cluster.CanManageClusterOperation,
	// task
	"ClusterManager.CreateTask": cluster.CanManageClusterOperation,
	"ClusterManager.RetryTask":  cluster.CanManageClusterOperation,
	"ClusterManager.UpdateTask": cluster.CanManageClusterOperation,
	"ClusterManager.DeleteTask": cluster.CanManageClusterOperation,
	"ClusterManager.GetTask":    cluster.CanViewClusterOperation,
	"ClusterManager.ListTask":   cluster.CanViewClusterOperation,
	// cluster auto scaling option
	"ClusterManager.CreateAutoScalingOption": cluster.CanManageClusterOperation,
	"ClusterManager.UpdateAutoScalingOption": cluster.CanManageClusterOperation,
	"ClusterManager.DeleteAutoScalingOption": cluster.CanManageClusterOperation,
	"ClusterManager.GetAutoScalingOption":    cluster.CanViewClusterOperation,
	"ClusterManager.ListAutoScalingOption":   cluster.CanViewClusterOperation,
	"ClusterManager.UpdateAutoScalingStatus": cluster.CanManageClusterOperation,
	// NodeTemplate
	"ClusterManager.CreateNodeTemplate": project.CanEditProjectOperation,
	"ClusterManager.UpdateNodeTemplate": project.CanEditProjectOperation,
	"ClusterManager.DeleteNodeTemplate": project.CanEditProjectOperation,
	"ClusterManager.ListNodeTemplate":   project.CanViewProjectOperation,
	// cloud account
	"ClusterManager.CreateCloudAccount": cloudaccount.CanManageCloudAccountOperation,
	"ClusterManager.UpdateCloudAccount": cloudaccount.CanManageCloudAccountOperation,
	"ClusterManager.DeleteCloudAccount": cloudaccount.CanManageCloudAccountOperation,
	"ClusterManager.ListCloudAccount":   cloudaccount.CanUseCloudAccountOperation,
	// vpc
	"ClusterManager.ListCloudVPCV2": "",
}
