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
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cloudaccount"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
)

// NoAuthMethod 不需要用户身份认证的方法
var NoAuthMethod = []string{
	// 集群相关
	"ClusterManager.ListCommonCluster",
	"ClusterManager.GetClustersMetaData",

	// 节点相关
	"ClusterManager.GetNodeInfo",
	"ClusterManager.GetNodeInfo",
	"ClusterManager.CheckNodeInCluster",
	"ClusterManager.CheckDrainNode",

	// cluster credential

	// nodeGroup resource
	"ClusterManager.GetExternalNodeScriptByGroupID",
	"ClusterManager.RecommendNodeGroupConf",

	// cloud template
	"ClusterManager.GetCloud",
	"ClusterManager.ListCloud",

	// vpc control
	"ClusterManager.ListCloudVPC",
	"ClusterManager.GetVPCCidr",

	// kubeConfig
	"ClusterManager.CheckCloudKubeConfig",
	"ClusterManager.CheckCloudKubeConfigConnect",

	// task resource
	"ClusterManager.RetryTask",
	"ClusterManager.GetTask",
	"ClusterManager.ListTask",
	"ClusterManager.SkipTask",

	// cluster autoscaling
	"ClusterManager.GetAutoScalingOption",
	"ClusterManager.ListAutoScalingOption",
	"ClusterManager.SyncAutoScalingOption",
	"ClusterManager.UpdateAsOptionDeviceProvider",

	// cloud account
	"ClusterManager.CreateCloudAccount",
	"ClusterManager.ListCloudAccount",
	"ClusterManager.ListCloudAccountToPerm",
	"ClusterManager.GetServiceRoles",

	// cloud module flag
	"ClusterManager.ListCloudModuleFlag",

	// common interface && support interface
	"ClusterManager.ListCloudRegions",
	"ClusterManager.GetCloudRegions",
	"ClusterManager.GetCloudRegionZones",
	"ClusterManager.ListCloudRegionCluster",
	"ClusterManager.ListCloudSubnets",
	"ClusterManager.ListCloudSecurityGroups",
	"ClusterManager.ListCloudInstanceTypes",
	"ClusterManager.ListCloudDiskTypes",
	"ClusterManager.ListCloudOsImage",
	"ClusterManager.ListCloudRuntimeInfo",
	"ClusterManager.ListCloudInstances",
	"ClusterManager.ListKeypairs",
	"ClusterManager.ListCloudProjects",
	"ClusterManager.ListCloudVpcs",
	"ClusterManager.ListCloudInstancesByPost",
	"ClusterManager.GetResourceGroups",

	"ClusterManager.ListOperationLogs",
	"ClusterManager.ListTaskStepLogs",
	"ClusterManager.ListTaskRecords",
	"ClusterManager.ListResourceSchema",
	"ClusterManager.GetResourceSchema",

	"ClusterManager.QueryPermByActionID",
	"ClusterManager.ListBKCloud",
	"ClusterManager.ListCCTopology",
	"ClusterManager.GetBkSopsTemplateList",
	"ClusterManager.GetBkSopsTemplateInfo",
	"ClusterManager.GetInnerTemplateValues",
	"ClusterManager.DebugBkSopsTask",
	"ClusterManager.Health",

	"ClusterManager.GetBatchCustomSetting",
	"ClusterManager.GetBizTopologyHost",
	"ClusterManager.GetTopologyNodes",
	"ClusterManager.GetScopeHostCheck",
	"ClusterManager.GetCloudAccountType",
	"ClusterManager.GetCloudBandwidthPackages",
	"ClusterManager.GetTopologyHostIdsNodes",
	"ClusterManager.GetHostsDetails",
	"ClusterManager.VerifyCloudAccount",
	"ClusterManager.CheckCidrConflictFromVpc",
	"ClusterManager.GetMasterSuggestedMachines",
	"ClusterManager.ListCloudNodePublicPrefix",
	"ClusterManager.GetClusterSharedProject",
}

// ActionPermissions action 对应权限中心的权限
var ActionPermissions = map[string]string{
	// cluster
	"ClusterManager.CreateCluster":                cluster.CanCreateClusterOperation,
	"ClusterManager.RetryCreateClusterTask":       cluster.CanCreateClusterOperation,
	"ClusterManager.ImportCluster":                cluster.CanCreateClusterOperation,
	"ClusterManager.UpdateCluster":                cluster.CanManageClusterOperation,
	"ClusterManager.DeleteCluster":                cluster.CanDeleteClusterOperation,
	"ClusterManager.GetCluster":                   cluster.CanViewClusterOperation,
	"ClusterManager.ListProjectCluster":           project.CanViewProjectOperation,
	"ClusterManager.ListCluster":                  project.CanViewProjectOperation,
	"ClusterManager.ListCommonCluster":            "",
	"ClusterManager.CreateVirtualCluster":         cluster.CanCreateClusterOperation,
	"ClusterManager.DeleteVirtualCluster":         cluster.CanDeleteClusterOperation,
	"ClusterManager.UpdateVirtualClusterQuota":    cluster.CanManageClusterOperation,
	"ClusterManager.AddSubnetToCluster":           cluster.CanManageClusterOperation,
	"ClusterManager.UpdateClusterModule":          cluster.CanManageClusterOperation,
	"ClusterManager.SwitchClusterUnderlayNetwork": cluster.CanManageClusterOperation,

	// node
	"ClusterManager.AddNodesToCluster":           cluster.CanManageClusterOperation,
	"ClusterManager.AddNodesToClusterV2":         cluster.CanManageClusterOperation,
	"ClusterManager.DeleteNodesFromCluster":      cluster.CanManageClusterOperation,
	"ClusterManager.BatchDeleteNodesFromCluster": cluster.CanManageClusterOperation,
	"ClusterManager.ListNodesInCluster":          cluster.CanViewClusterOperation,
	"ClusterManager.ListMastersInCluster":        cluster.CanViewClusterOperation,
	"ClusterManager.GetNode":                     cluster.CanViewClusterOperation,
	"ClusterManager.GetNodeInfo":                 cluster.CanViewClusterOperation,
	"ClusterManager.RecordNodeInfo":              "",
	"ClusterManager.CheckNodeInCluster":          "",
	"ClusterManager.UpdateNode":                  cluster.CanManageClusterOperation,
	"ClusterManager.CordonNode":                  cluster.CanManageClusterOperation,
	"ClusterManager.UnCordonNode":                cluster.CanManageClusterOperation,
	"ClusterManager.DrainNode":                   cluster.CanManageClusterOperation,
	"ClusterManager.UpdateNodeLabels":            cluster.CanManageClusterOperation,
	"ClusterManager.UpdateNodeTaints":            cluster.CanManageClusterOperation,
	"ClusterManager.UpdateNodeAnnotations":       cluster.CanManageClusterOperation,

	// federation cluster
	"ClusterManager.InitFederationCluster": "",
	"ClusterManager.AddFederatedCluster":   "",

	// cluster credential
	"ClusterManager.GetClusterCredential":    cluster.CanManageClusterOperation,
	"ClusterManager.UpdateClusterCredential": cluster.CanManageClusterOperation,
	"ClusterManager.UpdateClusterKubeConfig": cluster.CanManageClusterOperation,
	"ClusterManager.DeleteClusterCredential": "",
	"ClusterManager.ListClusterCredential":   "",

	// nodeGroup
	"ClusterManager.CreateNodeGroup":        cluster.CanManageClusterOperation,
	"ClusterManager.UpdateNodeGroup":        cluster.CanManageClusterOperation,
	"ClusterManager.DeleteNodeGroup":        cluster.CanManageClusterOperation,
	"ClusterManager.GetNodeGroup":           cluster.CanViewClusterOperation,
	"ClusterManager.ListNodeGroup":          cluster.CanViewClusterOperation,
	"ClusterManager.ListClusterNodeGroup":   cluster.CanViewClusterOperation,
	"ClusterManager.MoveNodesToGroup":       cluster.CanManageClusterOperation,
	"ClusterManager.RemoveNodesFromGroup":   cluster.CanManageClusterOperation,
	"ClusterManager.CleanNodesInGroup":      cluster.CanManageClusterOperation,
	"ClusterManager.CleanNodesInGroupV2":    cluster.CanManageClusterOperation,
	"ClusterManager.ListNodesInGroup":       cluster.CanViewClusterOperation,
	"ClusterManager.ListNodesInGroupV2":     cluster.CanViewClusterOperation,
	"ClusterManager.UpdateGroupAsTimeRange": cluster.CanManageClusterOperation,
	"ClusterManager.RecommendNodeGroupConf": "",

	"ClusterManager.UpdateGroupDesiredNode":         cluster.CanManageClusterOperation,
	"ClusterManager.UpdateGroupDesiredSize":         cluster.CanManageClusterOperation,
	"ClusterManager.UpdateGroupMinMaxSize":          cluster.CanManageClusterOperation,
	"ClusterManager.EnableNodeGroupAutoScale":       cluster.CanManageClusterOperation,
	"ClusterManager.DisableNodeGroupAutoScale":      cluster.CanManageClusterOperation,
	"ClusterManager.GetExternalNodeScriptByGroupID": cluster.CanManageClusterOperation,
	"ClusterManager.TransNodeGroupToNodeTemplate":   cluster.CanManageClusterOperation,
	"ClusterManager.CleanDbHistoryData":             "",
	"ClusterManager.GetProjectResourceQuotaUsage":   project.CanViewProjectOperation,

	// cloud template
	"ClusterManager.CreateCloud": "",
	"ClusterManager.UpdateCloud": "",
	"ClusterManager.DeleteCloud": "",
	"ClusterManager.GetCloud":    "",
	"ClusterManager.ListCloud":   "",

	// vpc control
	"ClusterManager.CreateCloudVPC": "",
	"ClusterManager.UpdateCloudVPC": "",
	"ClusterManager.DeleteCloudVPC": "",
	"ClusterManager.ListCloudVPC":   "",
	"ClusterManager.GetVPCCidr":     "",

	// kubeconfig
	"ClusterManager.CheckCloudKubeConfig":        "",
	"ClusterManager.CheckCloudKubeConfigConnect": "",

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
	"ClusterManager.SyncAutoScalingOption":   cluster.CanManageClusterOperation,
	// NodeTemplate
	"ClusterManager.CreateNodeTemplate": project.CanEditProjectOperation,
	"ClusterManager.UpdateNodeTemplate": project.CanEditProjectOperation,
	"ClusterManager.DeleteNodeTemplate": project.CanEditProjectOperation,
	"ClusterManager.ListNodeTemplate":   project.CanViewProjectOperation,
	"ClusterManager.GetNodeTemplate":    project.CanViewProjectOperation,

	// cloud account
	"ClusterManager.CreateCloudAccount":  cloudaccount.CanCreateCloudAccountOperation,
	"ClusterManager.UpdateCloudAccount":  cloudaccount.CanManageCloudAccountOperation,
	"ClusterManager.DeleteCloudAccount":  cloudaccount.CanManageCloudAccountOperation,
	"ClusterManager.ListCloudAccount":    cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.MigrateCloudAccount": "",
	"ClusterManager.VerifyCloudAccount":  "",
	"ClusterManager.GetServiceRoles":     "",

	// cloud component paras
	"ClusterManager.CreateCloudModuleFlag": "",
	"ClusterManager.UpdateCloudModuleFlag": "",
	"ClusterManager.DeleteCloudModuleFlag": "",
	"ClusterManager.ListCloudModuleFlag":   "",

	// common interface
	"ClusterManager.ListCloudRegions":         "",
	"ClusterManager.GetCloudRegions":          "",
	"ClusterManager.GetCloudRegionZones":      "",
	"ClusterManager.ListCloudRegionCluster":   "",
	"ClusterManager.ListCloudSubnets":         "",
	"ClusterManager.ListCloudSecurityGroups":  "",
	"ClusterManager.ListCloudInstanceTypes":   "",
	"ClusterManager.ListCloudDiskTypes":       "",
	"ClusterManager.ListCloudOsImage":         "",
	"ClusterManager.ListCloudRuntimeInfo":     "",
	"ClusterManager.ListCloudInstances":       "",
	"ClusterManager.ListOperationLogs":        "",
	"ClusterManager.ListTaskStepLogs":         "",
	"ClusterManager.TaskRecords":              "",
	"ClusterManager.ListResourceSchema":       "",
	"ClusterManager.GetResourceSchema":        "",
	"ClusterManager.QueryPermByActionID":      "",
	"ClusterManager.ListCloudInstancesByPost": "",

	"ClusterManager.ListBKCloud":            "",
	"ClusterManager.ListCCTopology":         "",
	"ClusterManager.GetBkSopsTemplateList":  "",
	"ClusterManager.GetBkSopsTemplateInfo":  "",
	"ClusterManager.GetInnerTemplateValues": "",
	"ClusterManager.DebugBkSopsTask":        "",

	"ClusterManager.Health": "",
}
