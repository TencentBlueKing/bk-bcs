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

// NoAuthMethod no auth method
var NoAuthMethod = []string{
	// 集群相关
	"ClusterManager.ListCommonCluster",
	"ClusterManager.GetClustersMetaData",

	// 节点相关
	"ClusterManager.CheckNodeInCluster",

	// cluster credential

	// nodeGroup resource

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

	// cluster autoscaling
	"ClusterManager.ListAutoScalingOption",
	"ClusterManager.SyncAutoScalingOption",

	// cloud account
	"ClusterManager.ListCloudAccountToPerm",

	// cloud module flag
	"ClusterManager.ListCloudModuleFlag",

	// common interface && support interface
	"ClusterManager.ListCloudRegions",
	"ClusterManager.ListCloudRuntimeInfo",

	"ClusterManager.ListOperationLogs",
	"ClusterManager.ListTaskRecords",
	"ClusterManager.ListResourceSchema",
	"ClusterManager.GetResourceSchema",

	"ClusterManager.QueryPermByActionID",
	"ClusterManager.ListBKCloud",
	"ClusterManager.ListCCTopology", // 集群id非必须
	"ClusterManager.GetInnerTemplateValues",
	"ClusterManager.DebugBkSopsTask",
	"ClusterManager.Health",

	"ClusterManager.VerifyCloudAccount",
}

// ActionPermissions actions perms
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
	"ClusterManager.ListClusterV2":                project.CanViewProjectOperation,
	"ClusterManager.ListBusinessCluster":          "", // 内部接口，无须权限校验
	"ClusterManager.ListCommonCluster":            "",
	"ClusterManager.CreateVirtualCluster":         cluster.CanCreateClusterOperation,
	"ClusterManager.DeleteVirtualCluster":         cluster.CanDeleteClusterOperation,
	"ClusterManager.UpdateVirtualClusterQuota":    cluster.CanManageClusterOperation,
	"ClusterManager.AddSubnetToCluster":           cluster.CanManageClusterOperation,
	"ClusterManager.UpdateClusterModule":          cluster.CanManageClusterOperation,
	"ClusterManager.SwitchClusterUnderlayNetwork": cluster.CanManageClusterOperation,
	"ClusterManager.GetClusterSharedProject":      cluster.CanViewClusterOperation,
	"ClusterManager.GetClustersMetaData":          "",

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
	"ClusterManager.CheckDrainNode":              cluster.CanViewClusterOperation,
	"ClusterManager.ListCloudNodePublicPrefix":   cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.ListCloudNodes":              "",
	"ClusterManager.SyncClusterNodes":            cluster.CanManageClusterOperation,

	// federation cluster
	"ClusterManager.InitFederationCluster": "",
	"ClusterManager.AddFederatedCluster":   cluster.CanCreateClusterOperation,

	// cluster credential
	"ClusterManager.GetClusterCredential":    cluster.CanManageClusterOperation,
	"ClusterManager.UpdateClusterCredential": cluster.CanManageClusterOperation,
	"ClusterManager.UpdateClusterKubeConfig": cluster.CanManageClusterOperation,
	"ClusterManager.DeleteClusterCredential": cluster.CanManageClusterOperation,
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
	"ClusterManager.RecommendNodeGroupConf": cloudaccount.CanUseCloudAccountOperation,

	"ClusterManager.UpdateGroupDesiredNode":         cluster.CanManageClusterOperation,
	"ClusterManager.UpdateGroupDesiredSize":         cluster.CanManageClusterOperation,
	"ClusterManager.UpdateGroupMinMaxSize":          cluster.CanManageClusterOperation,
	"ClusterManager.EnableNodeGroupAutoScale":       cluster.CanManageClusterOperation,
	"ClusterManager.DisableNodeGroupAutoScale":      cluster.CanManageClusterOperation,
	"ClusterManager.GetExternalNodeScriptByGroupID": cluster.CanViewClusterOperation,
	"ClusterManager.TransNodeGroupToNodeTemplate":   cluster.CanManageClusterOperation,
	"ClusterManager.CleanDbHistoryData":             "", // health类，无须权限
	"ClusterManager.GetProjectResourceQuotaUsage":   project.CanViewProjectOperation,

	// cloud template
	"ClusterManager.CreateCloud": "",
	"ClusterManager.UpdateCloud": "",
	"ClusterManager.DeleteCloud": "",
	"ClusterManager.GetCloud":    "",
	"ClusterManager.ListCloud":   "",

	// vpc control
	"ClusterManager.CreateCloudVPC":             "",
	"ClusterManager.UpdateCloudVPC":             "",
	"ClusterManager.DeleteCloudVPC":             "",
	"ClusterManager.ListCloudVPC":               "",
	"ClusterManager.GetVPCCidr":                 "",
	"ClusterManager.CheckCidrConflictFromVpc":   cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.GetCloudBandwidthPackages":  cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.GetMasterSuggestedMachines": cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.ListCloudVpcs":              cloudaccount.CanUseCloudAccountOperation,

	// kubeconfig
	"ClusterManager.CheckCloudKubeConfig":        "",
	"ClusterManager.CheckCloudKubeConfigConnect": cluster.CanViewClusterOperation,

	// task
	"ClusterManager.CreateTask": cluster.CanManageClusterOperation,
	"ClusterManager.RetryTask":  cluster.CanManageClusterOperation,
	"ClusterManager.UpdateTask": cluster.CanManageClusterOperation,
	"ClusterManager.DeleteTask": cluster.CanManageClusterOperation,
	"ClusterManager.GetTask":    cluster.CanViewClusterOperation,
	"ClusterManager.ListTask":   project.CanViewProjectOperation,
	"ClusterManager.SkipTask":   cluster.CanViewClusterOperation,

	// cluster auto scaling option
	"ClusterManager.CreateAutoScalingOption":      cluster.CanManageClusterOperation,
	"ClusterManager.UpdateAutoScalingOption":      cluster.CanManageClusterOperation,
	"ClusterManager.DeleteAutoScalingOption":      cluster.CanManageClusterOperation,
	"ClusterManager.GetAutoScalingOption":         cluster.CanViewClusterOperation,
	"ClusterManager.ListAutoScalingOption":        cluster.CanViewClusterOperation,
	"ClusterManager.UpdateAutoScalingStatus":      cluster.CanManageClusterOperation,
	"ClusterManager.SyncAutoScalingOption":        "",
	"ClusterManager.UpdateAsOptionDeviceProvider": cluster.CanManageClusterOperation,

	// NodeTemplate
	"ClusterManager.CreateNodeTemplate": project.CanEditProjectOperation,
	"ClusterManager.UpdateNodeTemplate": project.CanEditProjectOperation,
	"ClusterManager.DeleteNodeTemplate": project.CanEditProjectOperation,
	"ClusterManager.ListNodeTemplate":   project.CanViewProjectOperation,
	"ClusterManager.GetNodeTemplate":    project.CanViewProjectOperation,

	// NotifyTemplate
	"ClusterManager.CreateNotifyTemplate": project.CanViewProjectOperation,
	"ClusterManager.DeleteNotifyTemplate": project.CanViewProjectOperation,
	"ClusterManager.ListNotifyTemplate":   project.CanViewProjectOperation,

	// cloud account
	"ClusterManager.CreateCloudAccount":     cloudaccount.CanCreateCloudAccountOperation,
	"ClusterManager.UpdateCloudAccount":     cloudaccount.CanManageCloudAccountOperation,
	"ClusterManager.DeleteCloudAccount":     cloudaccount.CanManageCloudAccountOperation,
	"ClusterManager.ListCloudAccount":       cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.MigrateCloudAccount":    "", // 内部接口
	"ClusterManager.VerifyCloudAccount":     "", // 校验账号，无须权限校验账号，无须权限
	"ClusterManager.GetServiceRoles":        cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.ListCloudAccountToPerm": "", // 内部接口，无须权限校验

	// cloud component paras
	"ClusterManager.CreateCloudModuleFlag": "",
	"ClusterManager.UpdateCloudModuleFlag": "",
	"ClusterManager.DeleteCloudModuleFlag": "",
	"ClusterManager.ListCloudModuleFlag":   "",

	// common interface
	"ClusterManager.ListCloudRegions":         "",
	"ClusterManager.GetCloudRegions":          cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.GetCloudRegionZones":      cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.ListCloudRegionCluster":   cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.ListCloudSubnets":         cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.ListCloudSecurityGroups":  cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.ListCloudInstanceTypes":   cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.ListCloudDiskTypes":       cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.ListCloudOsImage":         cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.ListCloudRuntimeInfo":     "",
	"ClusterManager.ListCloudInstances":       cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.ListOperationLogs":        "",
	"ClusterManager.ListProjectOperationLogs": project.CanViewProjectOperation,
	"ClusterManager.ListTaskStepLogs":         project.CanViewProjectOperation,
	"ClusterManager.ListResourceSchema":       "",
	"ClusterManager.GetResourceSchema":        "",
	"ClusterManager.QueryPermByActionID":      "",
	"ClusterManager.ListCloudInstancesByPost": cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.GetResourceGroups":        cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.GetCloudAccountType":      cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.ListCloudProjects":        cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.ListKeypairs":             cloudaccount.CanUseCloudAccountOperation,
	"ClusterManager.ListTaskRecords":          "",

	"ClusterManager.ListBKCloud":              "",
	"ClusterManager.ListCCTopology":           cluster.CanViewClusterOperation,
	"ClusterManager.GetBkSopsTemplateList":    CanOperatorBiz,
	"ClusterManager.GetBkSopsTemplateInfo":    CanOperatorBiz,
	"ClusterManager.GetInnerTemplateValues":   "",
	"ClusterManager.DebugBkSopsTask":          "",
	"ClusterManager.GetBatchCustomSetting":    CanOperatorBiz,
	"ClusterManager.GetBizTopologyHost":       CanOperatorBiz,
	"ClusterManager.GetHostsDetails":          CanOperatorBiz,
	"ClusterManager.GetProviderResourceUsage": "",
	"ClusterManager.GetScopeHostCheck":        CanOperatorBiz,
	"ClusterManager.GetTopologyHostIdsNodes":  CanOperatorBiz,
	"ClusterManager.GetTopologyNodes":         CanOperatorBiz,

	"ClusterManager.Health": "",

	// template config
	"ClusterManager.CreateTemplateConfig": "", // 内部接口，无须权限校验
	"ClusterManager.DeleteTemplateConfig": "", // 内部接口，无须权限校验
	"ClusterManager.ListTemplateConfig":   "", // 内部接口，无须权限校验
	"ClusterManager.UpdateTemplateConfig": "", // 内部接口，无须权限校验
}
