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

package store

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/account"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/cloud"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/cloudvpc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/clustercredential"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/node"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/nodegroup"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/operationlog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/resourcequota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/scalingoption"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/task"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/tke"
)

// ClusterManagerModel database operation for
type ClusterManagerModel interface {
	CreateCluster(ctx context.Context, cluster *types.Cluster) error
	UpdateCluster(ctx context.Context, cluster *types.Cluster) error
	DeleteCluster(ctx context.Context, clusterID string) error
	GetCluster(ctx context.Context, clusterID string) (*types.Cluster, error)
	ListCluster(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]types.Cluster, error)

	CreateNode(ctx context.Context, node *types.Node) error
	UpdateNode(ctx context.Context, node *types.Node) error
	DeleteNode(ctx context.Context, nodeID string) error
	DeleteNodesByIPs(ctx context.Context, ips []string) error
	DeleteNodesByNodeIDs(ctx context.Context, nodeIDs []string) error
	DeleteNodeByIP(ctx context.Context, ip string) error
	DeleteNodesByClusterID(ctx context.Context, clusterID string) error
	GetNode(ctx context.Context, nodeID string) (*types.Node, error)
	GetNodeByIP(ctx context.Context, ip string) (*types.Node, error)
	ListNode(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]*types.Node, error)

	CreateNamespace(ctx context.Context, ns *types.Namespace) error
	UpdateNamespace(ctx context.Context, ns *types.Namespace) error
	DeleteNamespace(ctx context.Context, name, federationClusterID string) error
	GetNamespace(ctx context.Context, name, federationClusterID string) (*types.Namespace, error)
	ListNamespace(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]types.Namespace, error)

	CreateQuota(ctx context.Context, quota *types.ResourceQuota) error
	UpdateQuota(ctx context.Context, quota *types.ResourceQuota) error
	DeleteQuota(ctx context.Context, namespace, federationClusterID, clusterID string) error
	GetQuota(ctx context.Context, namespace, federationClusterID, clusterID string) (*types.ResourceQuota, error)
	ListQuota(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]types.ResourceQuota, error)
	BatchDeleteQuotaByCluster(ctx context.Context, clusterID string) error

	PutClusterCredential(ctx context.Context, clusterCredential *types.ClusterCredential) error
	GetClusterCredential(ctx context.Context, serverKey string) (*types.ClusterCredential, bool, error)
	DeleteClusterCredential(ctx context.Context, serverKey string) error
	ListClusterCredential(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
		[]types.ClusterCredential, error)

	//TKE CIDR information storage management
	CreateTkeCidr(ctx context.Context, cidr *types.TkeCidr) error
	UpdateTkeCidr(ctx context.Context, cidr *types.TkeCidr) error
	DeleteTkeCidr(ctx context.Context, vpc string, cidr string) error
	GetTkeCidr(ctx context.Context, vpc string, cidr string) (*types.TkeCidr, error)
	ListTkeCidr(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]types.TkeCidr, error)
	ListTkeCidrCount(ctx context.Context, opt *options.ListOption) ([]types.TkeCidrCount, error)

	//project information storage management
	CreateProject(ctx context.Context, project *types.Project) error
	UpdateProject(ctx context.Context, project *types.Project) error
	DeleteProject(ctx context.Context, projectID string) error
	GetProject(ctx context.Context, projectID string) (*types.Project, error)
	ListProject(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]types.Project, error)

	//cloud information storage management
	CreateCloud(ctx context.Context, cloud *types.Cloud) error
	UpdateCloud(ctx context.Context, cloud *types.Cloud) error
	DeleteCloud(ctx context.Context, cloudID string) error
	GetCloud(ctx context.Context, cloudID string) (*types.Cloud, error)
	ListCloud(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]types.Cloud, error)

	//cloud vpc information storage manager
	CreateCloudVPC(ctx context.Context, vpc *types.CloudVPC) error
	UpdateCloudVPC(ctx context.Context, vpc *types.CloudVPC) error
	DeleteCloudVPC(ctx context.Context, cloudID string, vpcID string) error
	ListCloudVPC(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]types.CloudVPC, error)
	GetCloudVPC(ctx context.Context, cloudID, vpcID string) (*types.CloudVPC, error)

	//cloud account info storage manager
	CreateCloudAccount(ctx context.Context, account *types.CloudAccount) error
	UpdateCloudAccount(ctx context.Context, account *types.CloudAccount) error
	DeleteCloudAccount(ctx context.Context, cloudID string, accountID string) error
	ListCloudAccount(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]types.CloudAccount, error)
	GetCloudAccount(ctx context.Context, cloudID, accountID string) (*types.CloudAccount, error)

	//nodegroup information storage management
	CreateNodeGroup(ctx context.Context, group *types.NodeGroup) error
	UpdateNodeGroup(ctx context.Context, group *types.NodeGroup) error
	DeleteNodeGroup(ctx context.Context, groupID string) error
	GetNodeGroup(ctx context.Context, groupID string) (*types.NodeGroup, error)
	ListNodeGroup(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]types.NodeGroup, error)
	DeleteNodeGroupByClusterID(ctx context.Context, clusterID string) error

	//task information storage management
	CreateTask(ctx context.Context, task *types.Task) error
	UpdateTask(ctx context.Context, task *types.Task) error
	PatchTask(ctx context.Context, taskID string, patchs map[string]interface{}) error
	DeleteTask(ctx context.Context, taskID string) error
	GetTask(ctx context.Context, taskID string) (*types.Task, error)
	ListTask(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]types.Task, error)

	// OperationLog
	CreateOperationLog(ctx context.Context, log *types.OperationLog) error
	DeleteOperationLogByResourceID(ctx context.Context, resourceIndex string) error
	DeleteOperationLogByResourceType(ctx context.Context, resType string) error
	ListOperationLog(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]types.OperationLog, error)

	//project information storage management
	CreateAutoScalingOption(ctx context.Context, option *types.ClusterAutoScalingOption) error
	UpdateAutoScalingOption(ctx context.Context, option *types.ClusterAutoScalingOption) error
	DeleteAutoScalingOption(ctx context.Context, clusterID string) error
	GetAutoScalingOption(ctx context.Context, clusterID string) (*types.ClusterAutoScalingOption, error)
	ListAutoScalingOption(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]types.ClusterAutoScalingOption, error)
}

// ModelSet a set of client
type ModelSet struct {
	*cluster.ModelCluster
	*node.ModelNode
	*clustercredential.ModelClusterCredential
	*namespace.ModelNamespace
	*resourcequota.ModelResourceQuota
	*tke.ModelTkeCidr
	*cloud.ModelCloud
	*project.ModelProject
	*nodegroup.ModelNodeGroup
	*task.ModelTask
	*scalingoption.ModelAutoScalingOption
	*cloudvpc.ModelCloudVPC
	*operationlog.ModelOperationLog
	*account.ModelCloudAccount
}

// NewModelSet create model set
func NewModelSet(db drivers.DB) ClusterManagerModel {
	return &ModelSet{
		ModelCluster:           cluster.New(db),
		ModelNode:              node.New(db),
		ModelClusterCredential: clustercredential.New(db),
		ModelNamespace:         namespace.New(db),
		ModelResourceQuota:     resourcequota.New(db),
		ModelTkeCidr:           tke.New(db),
		ModelCloud:             cloud.New(db),
		ModelProject:           project.New(db),
		ModelNodeGroup:         nodegroup.New(db),
		ModelTask:              task.New(db),
		ModelAutoScalingOption: scalingoption.New(db),
		ModelCloudVPC:          cloudvpc.New(db),
		ModelOperationLog:      operationlog.New(db),
		ModelCloudAccount:      account.New(db),
	}
}
