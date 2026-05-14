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

// Package store xxx
package store

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	client "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc"

	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/account"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/cloud"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/cloudvpc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/clustercredential"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/etcd"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/machinery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/moduleflag"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/node"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/nodegroup"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/nodetemplate"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/notifytemplate"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/operationlog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/resourcequota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/scalingoption"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/task"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/templateconfig"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/tke"
	stypes "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/types"
	itypes "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

var storeClient ClusterManagerModel

// ClusterManagerModel database operation for
type ClusterManagerModel interface {
	// cluster information storage management
	CreateCluster(ctx context.Context, cluster *types.Cluster) error
	UpdateCluster(ctx context.Context, cluster *types.Cluster) error
	DeleteCluster(ctx context.Context, clusterID string) error
	GetCluster(ctx context.Context, clusterID string) (*types.Cluster, error)
	ListCluster(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]*types.Cluster, error)

	// node information storage management
	CreateNode(ctx context.Context, node *types.Node) error
	UpdateNode(ctx context.Context, node *types.Node) error
	UpdateClusterNodeByNodeID(ctx context.Context, node *types.Node) error
	DeleteNode(ctx context.Context, nodeID string) error
	DeleteClusterNode(ctx context.Context, clusterID, nodeID string) error
	DeleteClusterNodeByName(ctx context.Context, clusterID, nodeName string) error
	DeleteNodesByIPs(ctx context.Context, ips []string) error
	DeleteClusterNodesByIPs(ctx context.Context, clusterID string, ips []string) error
	DeleteNodesByNodeIDs(ctx context.Context, nodeIDs []string) error
	DeleteNodeByIP(ctx context.Context, ip string) error
	DeleteClusterNodeByIP(ctx context.Context, clusterID, ip string) error
	DeleteNodesByClusterID(ctx context.Context, clusterID string) error
	GetNode(ctx context.Context, nodeID string) (*types.Node, error)
	GetNodeByName(ctx context.Context, clusterID, name string) (*types.Node, error)
	GetNodeByIP(ctx context.Context, ip string) (*types.Node, error)
	GetClusterNode(ctx context.Context, clusterID, nodeID string) (*types.Node, error)
	GetClusterNodeByIP(ctx context.Context, clusterID, ip string) (*types.Node, error)
	ListNode(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]*types.Node, error)

	// namespace information storage management
	CreateNamespace(ctx context.Context, ns *types.Namespace) error
	UpdateNamespace(ctx context.Context, ns *types.Namespace) error
	DeleteNamespace(ctx context.Context, name, federationClusterID string) error
	GetNamespace(ctx context.Context, name, federationClusterID string) (*types.Namespace, error)
	ListNamespace(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]*types.Namespace, error)

	// quota information storage management
	CreateQuota(ctx context.Context, quota *types.ResourceQuota) error
	UpdateQuota(ctx context.Context, quota *types.ResourceQuota) error
	DeleteQuota(ctx context.Context, namespace, federationClusterID, clusterID string) error
	GetQuota(ctx context.Context, namespace, federationClusterID, clusterID string) (*types.ResourceQuota, error)
	ListQuota(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]*types.ResourceQuota, error)
	BatchDeleteQuotaByCluster(ctx context.Context, clusterID string) error

	// credential information storage management
	PutClusterCredential(ctx context.Context, clusterCredential *types.ClusterCredential) error
	GetClusterCredential(ctx context.Context, serverKey string) (*types.ClusterCredential, bool, error)
	DeleteClusterCredential(ctx context.Context, serverKey string) error
	ListClusterCredential(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
		[]*types.ClusterCredential, error)

	// TKE CIDR information storage management
	CreateTkeCidr(ctx context.Context, cidr *types.TkeCidr) error
	UpdateTkeCidr(ctx context.Context, cidr *types.TkeCidr) error
	DeleteTkeCidr(ctx context.Context, vpc string, cidr string) error
	GetTkeCidr(ctx context.Context, vpc string, cidr string) (*types.TkeCidr, error)
	ListTkeCidr(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]*types.TkeCidr, error)
	ListTkeCidrCount(ctx context.Context, opt *options.ListOption) ([]*types.TkeCidrCount, error)

	// project information storage management
	CreateProject(ctx context.Context, project *types.Project) error
	UpdateProject(ctx context.Context, project *types.Project) error
	DeleteProject(ctx context.Context, projectID string) error
	GetProject(ctx context.Context, projectID string) (*types.Project, error)
	ListProject(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]*types.Project, error)

	// cloud information storage management
	CreateCloud(ctx context.Context, cloud *types.Cloud) error
	UpdateCloud(ctx context.Context, cloud *types.Cloud) error
	DeleteCloud(ctx context.Context, cloudID string) error
	GetCloud(ctx context.Context, cloudID string) (*types.Cloud, error)
	GetCloudByProvider(ctx context.Context, provider string) (*types.Cloud, error)
	ListCloud(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]*types.Cloud, error)

	// cloud vpc information storage manager
	CreateCloudVPC(ctx context.Context, vpc *types.CloudVPC) error
	UpdateCloudVPC(ctx context.Context, vpc *types.CloudVPC) error
	DeleteCloudVPC(ctx context.Context, cloudID string, vpcID string) error
	ListCloudVPC(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]*types.CloudVPC, error)
	GetCloudVPC(ctx context.Context, cloudID, vpcID string) (*types.CloudVPC, error)

	// cloud account info storage manager
	CreateCloudAccount(ctx context.Context, account *types.CloudAccount) error
	UpdateCloudAccount(ctx context.Context, account *types.CloudAccount, skipEncrypt bool) error
	DeleteCloudAccount(ctx context.Context, cloudID string, accountID string) error
	ListCloudAccount(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]*types.CloudAccount, error)
	GetCloudAccount(ctx context.Context, cloudID, accountID string, skipDecrypt bool) (*types.CloudAccount, error)

	// cloud nodeTemplate info storage management
	CreateNodeTemplate(ctx context.Context, template *types.NodeTemplate) error
	UpdateNodeTemplate(ctx context.Context, template *types.NodeTemplate) error
	DeleteNodeTemplate(ctx context.Context, projectID string, templateID string) error
	ListNodeTemplate(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
		[]*types.NodeTemplate, error)
	GetNodeTemplate(ctx context.Context, projectID, templateID string) (*types.NodeTemplate, error)
	GetNodeTemplateByID(ctx context.Context, templateID string) (*types.NodeTemplate, error)

	// notifyTemplate info storage management
	CreateNotifyTemplate(ctx context.Context, template *types.NotifyTemplate) error
	UpdateNotifyTemplate(ctx context.Context, template *types.NotifyTemplate) error
	DeleteNotifyTemplate(ctx context.Context, projectID string, templateID string) error
	ListNotifyTemplate(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
		[]*types.NotifyTemplate, error)
	GetNotifyTemplate(ctx context.Context, projectID, templateID string) (*types.NotifyTemplate, error)
	GetNotifyTemplateByID(ctx context.Context, templateID string) (*types.NotifyTemplate, error)

	// nodegroup information storage management
	CreateNodeGroup(ctx context.Context, group *types.NodeGroup) error
	UpdateNodeGroup(ctx context.Context, group *types.NodeGroup) error
	DeleteNodeGroup(ctx context.Context, groupID string) error
	GetNodeGroup(ctx context.Context, groupID string) (*types.NodeGroup, error)
	CountNodeGroup(ctx context.Context, cond *operator.Condition) (int64, error)
	ListNodeGroup(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]*types.NodeGroup, error)
	DeleteNodeGroupByClusterID(ctx context.Context, clusterID string) error

	// task information storage management
	CreateTask(ctx context.Context, task *types.Task) error
	UpdateTask(ctx context.Context, task *types.Task) error
	PatchTask(ctx context.Context, taskID string, patchs map[string]interface{}) error
	DeleteTask(ctx context.Context, taskID string) error
	GetTask(ctx context.Context, taskID string) (*types.Task, error)
	ListTask(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]*types.Task, error)
	DeleteFinishedTaskByDate(ctx context.Context, startTime, endTime string) error
	ListMachineryTasks(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]stypes.Task, error)
	GetTasksFieldDistinct(ctx context.Context, fieldName string, filter interface{}) ([]string, error)

	// OperationLog
	CreateOperationLog(ctx context.Context, log *types.OperationLog) error
	DeleteOperationLogByResourceID(ctx context.Context, resourceIndex string) error
	DeleteOperationLogByResourceType(ctx context.Context, resType string) error
	ListOperationLog(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]*types.OperationLog, error)
	CountOperationLog(ctx context.Context, cond *operator.Condition) (int64, error)
	ListAggreOperationLog(ctx context.Context, condSrc, condDst []bson.E,
		opt *options.ListOption) ([]*types.TaskOperationLog, error)
	DeleteOperationLogByDate(ctx context.Context, startTime, endTime string) error

	// TaskStepLog
	CreateTaskStepLogInfo(ctx context.Context, taskID, stepName, message string)
	CreateTaskStepLogWarn(ctx context.Context, taskID, stepName, message string)
	CreateTaskStepLogError(ctx context.Context, taskID, stepName, message string)
	DeleteTaskStepLogByTaskID(ctx context.Context, taskID string) error
	CountTaskStepLog(ctx context.Context, cond *operator.Condition) (int64, error)
	ListTaskStepLog(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]*types.TaskStepLog, error)

	// project information storage management
	CreateAutoScalingOption(ctx context.Context, option *types.ClusterAutoScalingOption) error
	UpdateAutoScalingOption(ctx context.Context, option *types.ClusterAutoScalingOption) error
	DeleteAutoScalingOption(ctx context.Context, clusterID string) error
	GetAutoScalingOption(ctx context.Context, clusterID string) (*types.ClusterAutoScalingOption, error)
	ListAutoScalingOption(ctx context.Context, cond *operator.Condition,
		opt *options.ListOption) ([]*types.ClusterAutoScalingOption, error)

	// cloudModuleFlag storage management
	CreateCloudModuleFlag(ctx context.Context, flag *types.CloudModuleFlag) error
	UpdateCloudModuleFlag(ctx context.Context, flag *types.CloudModuleFlag) error
	DeleteCloudModuleFlag(ctx context.Context, cloudID, version, module, flag string) error
	GetCloudModuleFlag(ctx context.Context, cloudID, version, module, flag string) (*types.CloudModuleFlag, error)
	ListCloudModuleFlag(ctx context.Context, cond *operator.Condition,
		opt *options.ListOption) ([]*types.CloudModuleFlag, error)

	// TemplateConfig storage management
	CreateTemplateConfig(ctx context.Context, config *types.TemplateConfig) error
	UpdateTemplateConfig(ctx context.Context, config *types.TemplateConfig) error
	DeleteTemplateConfig(ctx context.Context, templateConfigID string) error
	GetTemplateConfig(ctx context.Context, businessID, provider, configType string) (*types.TemplateConfig, error)
	GetTemplateConfigByID(ctx context.Context, templateConfigID string) (*types.TemplateConfig, error)
	ListTemplateConfigs(
		ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]*types.TemplateConfig, error)
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
	*operationlog.ModelTaskStepLog
	*account.ModelCloudAccount
	*nodetemplate.ModelNodeTemplate
	*moduleflag.ModelCloudModuleFlag
	*machinery.ModelMachineryTask
	*notifytemplate.ModelNotifyTemplate
	*templateconfig.ModelTemplateConfig
}

// NewModelSet create model set
func NewModelSet(mongoOptions *mongo.Options) (ClusterManagerModel, error) {
	// init db
	db, err := mongo.NewDB(mongoOptions)
	if err != nil {
		blog.Errorf("init mongo db failed, err %s", err.Error())
		return nil, err
	}
	if err = db.Ping(); err != nil {
		blog.Errorf("ping mongo db failed, err %s", err.Error())
		return nil, err
	}
	blog.Infof("init mongo db successfully")

	mTaskDb, err := machinery.New(db, mongoOptions)
	if err != nil {
		return nil, err
	}

	storeClient = &ModelSet{ // nolint
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
		ModelTaskStepLog:       operationlog.NewTaskStepLog(db),
		ModelCloudAccount:      account.New(db),
		ModelNodeTemplate:      nodetemplate.New(db),
		ModelCloudModuleFlag:   moduleflag.New(db),
		ModelMachineryTask:     mTaskDb,
		ModelNotifyTemplate:    notifytemplate.New(db),
		ModelTemplateConfig:    templateconfig.New(db),
	}

	return storeClient, nil
}

// GetStoreModel get store client
func GetStoreModel() ClusterManagerModel {
	return storeClient
}

var etcdStoreClient EtcdStoreInterface

// GetEtcdModel get etcd client
func GetEtcdModel() EtcdStoreInterface {
	return etcdStoreClient
}

// EtcdStoreInterface for etcd data
type EtcdStoreInterface interface {
	Create(ctx context.Context, key string, obj interface{}) error
	Delete(ctx context.Context, key string) error
	Get(ctx context.Context, key string, objPtr interface{}) error
	List(ctx context.Context, key string, listObj interface{}) error
}

// NewModelEtcd create etcd store
func NewModelEtcd(opts ...itypes.Option) (EtcdStoreInterface, error) {
	var etcdOptions itypes.Options
	for _, o := range opts {
		o(&etcdOptions)
	}

	var endpoints []string
	for _, addr := range etcdOptions.Endpoints {
		if len(addr) > 0 {
			endpoints = append(endpoints, addr)
		}
	}

	// set etcd client config
	var conf client.Config
	if etcdOptions.TLSConfig != nil {
		conf = client.Config{
			Endpoints: endpoints,
			TLS:       etcdOptions.TLSConfig,
		}
	} else {
		conf = client.Config{
			Endpoints: endpoints,
		}
	}
	conf.DialOptions = []grpc.DialOption{grpc.WithBlock()}
	conf.DialTimeout = 10 * time.Second
	etcdClient, err := client.New(conf)
	if err != nil {
		return nil, err
	}

	etcdStoreClient = etcd.NewEtcdStore(etcdOptions.Prefix, etcdClient)
	return etcdStoreClient, nil
}
