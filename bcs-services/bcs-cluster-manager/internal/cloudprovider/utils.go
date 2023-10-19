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
 *
 */

package cloudprovider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/modules"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/nodeman"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	// BKSOPTask bk-sops common job
	BKSOPTask = "bksopsjob"
	// WatchTask watch component common job
	WatchTask = "watchjob"
	// EnsureAutoScalerAction install/update ca component
	EnsureAutoScalerAction = "ensureAutoScaler"
)

var (
	defaultTaskID = "qwertyuiop123456"
	// TaskID inject taskID into ctx
	TaskID = "taskID"
)

// GetTaskIDFromContext get taskID from context
func GetTaskIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(TaskID).(string); ok {
		return id
	}

	return defaultTaskID
}

// WithTaskIDForContext will return a new context wrapped taskID flag around the original ctx
func WithTaskIDForContext(ctx context.Context, taskID string) context.Context {
	return context.WithValue(ctx, TaskID, taskID)
}

// CredentialData dependency data
type CredentialData struct {
	// Cloud cloud
	Cloud *proto.Cloud
	// Cluster cluster
	AccountID string
}

// GetCredential get specified credential information according Cloud configuration, if Cloud conf is nil, try Cluster Account.
// @return CommonOption: option can be nil if no credential conf in cloud or cluster account or when cloudprovider don't support authentication
// GetCredential get cloud credential by cloud or cluster
func GetCredential(data *CredentialData) (*CommonOption, error) {
	if data.Cloud == nil && data.AccountID == "" {
		return nil, fmt.Errorf("lost cloud/account information")
	}

	option := &CommonOption{}
	// get credential from cloud
	if data.Cloud.CloudCredential != nil {
		option.Account = &proto.Account{
			SecretID:          data.Cloud.CloudCredential.Key,
			SecretKey:         data.Cloud.CloudCredential.Secret,
			SubscriptionID:    data.Cloud.CloudCredential.SubscriptionID,
			TenantID:          data.Cloud.CloudCredential.TenantID,
			ResourceGroupName: data.Cloud.CloudCredential.ResourceGroupName,
			ClientID:          data.Cloud.CloudCredential.ClientID,
			ClientSecret:      data.Cloud.CloudCredential.ClientSecret,
		}
	}

	// if credential not exist cloud, get from cluster account
	if option.Account == nil && data.AccountID != "" {
		// try to get credential in cluster
		account, err := GetStorageModel().GetCloudAccount(context.Background(), data.Cloud.CloudID, data.AccountID)
		if err != nil {
			return nil, fmt.Errorf("GetCloudAccount failed: %v", err)
		}
		option.Account = account.Account
	}

	// set cloud basic confInfo
	option.CommonConf = CloudConf{
		CloudInternalEnable: data.Cloud.ConfInfo.CloudInternalEnable,
		CloudDomain:         data.Cloud.ConfInfo.CloudDomain,
		MachineDomain:       data.Cloud.ConfInfo.MachineDomain,
	}

	// check cloud credential info
	err := checkCloudCredentialValidate(data.Cloud, option)
	if err != nil {
		return nil, fmt.Errorf("checkCloudCredentialValidate %s failed: %v", data.Cloud.CloudProvider, err)
	}

	return option, nil
}

func checkCloudCredentialValidate(cloud *proto.Cloud, option *CommonOption) error {
	validate, err := GetCloudValidateMgr(cloud.CloudProvider)
	if err != nil {
		return err
	}
	err = validate.ImportCloudAccountValidate(option.Account)
	if err != nil {
		return err
	}

	return nil
}

// TaskType taskType
type TaskType string

// String toString
func (tt TaskType) String() string {
	return string(tt)
}

var (
	// CreateCluster task
	CreateCluster TaskType = "CreateCluster"
	// ImportCluster task
	ImportCluster TaskType = "ImportCluster"
	// DeleteCluster task
	DeleteCluster TaskType = "DeleteCluster"
	// AddNodesToCluster task
	AddNodesToCluster TaskType = "AddNodesToCluster"
	// RemoveNodesFromCluster task
	RemoveNodesFromCluster TaskType = "RemoveNodesFromCluster"

	// CreateNodeGroup task
	CreateNodeGroup TaskType = "CreateNodeGroup"
	// UpdateNodeGroup task
	UpdateNodeGroup TaskType = "UpdateNodeGroup"
	// DeleteNodeGroup task
	DeleteNodeGroup TaskType = "DeleteNodeGroup"
	// MoveNodesToNodeGroup task
	MoveNodesToNodeGroup TaskType = "MoveNodesToNodeGroup"
	// SwitchNodeGroupAutoScaling task
	SwitchNodeGroupAutoScaling TaskType = "SwitchNodeGroupAutoScaling"
	// UpdateNodeGroupDesiredNode task
	UpdateNodeGroupDesiredNode TaskType = "UpdateNodeGroupDesiredNode"
	// CleanNodeGroupNodes task
	CleanNodeGroupNodes TaskType = "CleanNodeGroupNodes"
	// UpdateAutoScalingOption task
	UpdateAutoScalingOption TaskType = "UpdateAutoScalingOption"
	// SwitchAutoScalingOptionStatus task
	SwitchAutoScalingOptionStatus TaskType = "SwitchAutoScalingOptionStatus"

	// ApplyInstanceMachinesTask apply instance subTask
	ApplyInstanceMachinesTask TaskType = "ApplyInstanceMachinesTask"
)

// GetTaskType getTaskType by cloud
func GetTaskType(cloud string, taskName TaskType) string {
	return fmt.Sprintf("%s-%s", cloud, taskName.String())
}

// CloudDependBasicInfo cloud depend cluster info
type CloudDependBasicInfo struct {
	// Cluster info
	Cluster *proto.Cluster
	// Cloud info
	Cloud *proto.Cloud
	// NodeGroup info
	NodeGroup *proto.NodeGroup
	// CmOption option
	CmOption *CommonOption
}

// GetClusterDependBasicInfo get cluster/cloud/nodeGroup depend info, nodeGroup may be nil.
// only get metadata, try not to change it
func GetClusterDependBasicInfo(clusterID string, cloudID string, nodeGroupID string) (*CloudDependBasicInfo, error) {
	var (
		cluster   *proto.Cluster
		cloud     *proto.Cloud
		nodeGroup *proto.NodeGroup
		err       error
	)

	cloud, cluster, err = actions.GetCloudAndCluster(GetStorageModel(), cloudID, clusterID)
	if err != nil {
		return nil, err
	}

	// cloud credential info
	cmOption, err := GetCredential(&CredentialData{
		Cloud:     cloud,
		AccountID: cluster.CloudAccountID,
	})
	if err != nil {
		return nil, err
	}
	cmOption.Region = cluster.Region

	if len(nodeGroupID) > 0 {
		nodeGroup, err = actions.GetNodeGroupByGroupID(GetStorageModel(), nodeGroupID)
		if err != nil {
			return nil, err
		}
	}

	return &CloudDependBasicInfo{cluster, cloud, nodeGroup, cmOption}, nil
}

// ParseNodeIpOrIdFromCommonMap parse nodeIDs or nodeIPs by chart
func ParseNodeIpOrIdFromCommonMap(taskCommonMap map[string]string, key string, chart string) []string {
	val, ok := taskCommonMap[key]
	if !ok || val == "" {
		return nil
	}

	return strings.Split(val, chart)
}

// UpdateNodeStatus update node status; isInstanceIP true, instance is InstanceIP; isInstanceIP true,
// instance is InstanceID
func UpdateNodeStatus(isInstanceIP bool, instance, status string) error {
	var (
		node *proto.Node
		err  error
	)
	if isInstanceIP {
		node, err = GetStorageModel().GetNodeByIP(context.Background(), instance)
	} else {
		node, err = GetStorageModel().GetNode(context.Background(), instance)
	}
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		return err
	}

	if errors.Is(err, drivers.ErrTableRecordNotFound) {
		return nil
	}

	node.Status = status
	err = GetStorageModel().UpdateNode(context.Background(), node)
	if err != nil {
		return err
	}

	return nil
}

// UpdateClusterStatus set cluster status
func UpdateClusterStatus(clusterID string, status string) error {
	cluster, err := GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		return err
	}

	cluster.Status = status
	err = GetStorageModel().UpdateCluster(context.Background(), cluster)
	if err != nil {
		return err
	}

	return nil
}

// UpdateCluster update cluster
func UpdateCluster(cluster *proto.Cluster) error {
	err := GetStorageModel().UpdateCluster(context.Background(), cluster)
	if err != nil {
		return err
	}

	return nil
}

// UpdateClusterCredentialByConfig update clusterCredential by kubeConfig
func UpdateClusterCredentialByConfig(clusterID string, config *types.Config) error {
	// first import cluster need to auto generate clusterCredential info, subsequently kube-agent report to update
	// currently, bcs only support token auth, kubeConfigList length greater 0, get zeroth kubeConfig
	var (
		server     = ""
		caCertData = ""
		token      = ""
		clientCert = ""
		clientKey  = ""
	)
	if len(config.Clusters) > 0 {
		server = config.Clusters[0].Cluster.Server
		caCertData = string(config.Clusters[0].Cluster.CertificateAuthorityData)
	}
	if len(config.AuthInfos) > 0 {
		token = config.AuthInfos[0].AuthInfo.Token
		clientCert = string(config.AuthInfos[0].AuthInfo.ClientCertificateData)
		clientKey = string(config.AuthInfos[0].AuthInfo.ClientKeyData)
	}

	if server == "" || caCertData == "" || (token == "" && (clientCert == "" || clientKey == "")) {
		return fmt.Errorf("importClusterCredential parse kubeConfig failed: %v", "[server|caCertData|token] null")
	}

	now := time.Now().Format(time.RFC3339)
	err := GetStorageModel().PutClusterCredential(context.Background(), &proto.ClusterCredential{
		ServerKey:     clusterID,
		ClusterID:     clusterID,
		ClientModule:  modules.BCSModuleKubeagent,
		ServerAddress: server,
		CaCertData:    caCertData,
		UserToken:     token,
		ConnectMode:   modules.BCSConnectModeDirect,
		CreateTime:    now,
		UpdateTime:    now,
		ClientKey:     clientKey,
		ClientCert:    clientCert,
	})
	if err != nil {
		return err
	}

	return nil
}

// ListNodesInClusterNodePool list nodeGroup nodes
func ListNodesInClusterNodePool(clusterID, nodePoolID string) ([]*proto.Node, error) {
	condM := make(operator.M)
	condM["nodegroupid"] = nodePoolID
	condM["clusterid"] = clusterID
	cond := operator.NewLeafCondition(operator.Eq, condM)
	nodes, err := GetStorageModel().ListNode(context.Background(), cond, &storeopt.ListOption{})
	if err != nil {
		blog.Errorf("ListNodesInClusterNodePool NodeGroup %s all Nodes failed, %s", nodePoolID, err.Error())
		return nil, err
	}

	// sum running & creating nodes, these status are ready to serve workload
	var (
		goodNodes []*proto.Node
	)
	for _, node := range nodes {
		if node.Status == common.StatusRunning || node.Status == common.StatusInitialization {
			goodNodes = append(goodNodes, node)
		}
	}

	return goodNodes, nil
}

// GetNodesNumWhenApplyInstanceTask get nodeNum
func GetNodesNumWhenApplyInstanceTask(clusterID, nodeGroupID, taskType, status string, steps []string) (int, error) {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"clusterid":   clusterID,
		"tasktype":    taskType,
		"nodegroupid": nodeGroupID,
		"status":      status,
	})
	taskList, err := GetStorageModel().ListTask(context.Background(), cond, &storeopt.ListOption{})
	if err != nil {
		blog.Errorf("GetNodesNumWhenApplyInstanceTask failed: %v", err)
		return 0, err
	}

	currentScalingNodes := 0
	for i := range taskList {
		if utils.StringInSlice(taskList[i].CurrentStep, steps) {
			desiredNodes := taskList[i].CommonParams[ScalingKey.String()]
			nodeNum, err := strconv.Atoi(desiredNodes)
			if err != nil {
				blog.Errorf("GetNodesNumWhenApplyInstanceTask strconv desiredNodes failed: %v", err)
				continue
			}
			currentScalingNodes += nodeNum
		}
	}

	return currentScalingNodes, nil
}

// UpdateNodeGroupDesiredSize when scaleOutNodes failed
func UpdateNodeGroupDesiredSize(groupID string, nodeNum int, scaleOut bool) error {
	group, err := GetStorageModel().GetNodeGroup(context.Background(), groupID)
	if err != nil {
		blog.Errorf("updateNodeGroupDesiredSize failed when CA scale nodes: %v", err)
		return err
	}

	if scaleOut {
		if group.AutoScaling.DesiredSize >= uint32(nodeNum) {
			group.AutoScaling.DesiredSize = group.AutoScaling.DesiredSize - uint32(nodeNum)
		} else {
			group.AutoScaling.DesiredSize = 0
			blog.Warnf("updateNodeGroupDesiredSize abnormal, desiredSize[%v] scaleNodesNum[%v]",
				group.AutoScaling.DesiredSize, nodeNum)
		}
	} else {
		group.AutoScaling.DesiredSize = group.AutoScaling.DesiredSize + uint32(nodeNum)
	}

	err = GetStorageModel().UpdateNodeGroup(context.Background(), group)
	if err != nil {
		blog.Errorf("updateNodeGroupDesiredSize failed when CA scale nodes: %v", err)
		return err
	}

	return nil
}

// SaveNodeInfoToDB save node to DB
func SaveNodeInfoToDB(node *proto.Node) error {
	instanceID := node.NodeID

	oldNode, err := GetStorageModel().GetNode(context.Background(), instanceID)
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		return fmt.Errorf("saveNodeInfoToDB getNode[%s] failed: %v", node.NodeID, err)
	}

	if oldNode == nil {
		err = GetStorageModel().CreateNode(context.Background(), node)
		if err != nil {
			return fmt.Errorf("saveNodeInfoToDB createNode[%s] failed: %v", node.NodeID, err)
		}

		return nil
	}

	err = GetStorageModel().UpdateNode(context.Background(), node)
	if err != nil {
		return fmt.Errorf("saveNodeInfoToDB updateNode[%s] failed: %v", node.NodeID, err)
	}

	return nil
}

// UpdateNodeStatusByInstanceID update node status
func UpdateNodeStatusByInstanceID(instanceID, status string) error {
	node, err := GetStorageModel().GetNode(context.Background(), instanceID)
	if err != nil {
		return err
	}

	node.Status = status

	err = GetStorageModel().UpdateNode(context.Background(), node)
	if err != nil {
		return err
	}

	return nil
}

// UpdateClusterSystemID set cluster systemID
func UpdateClusterSystemID(clusterID string, systemID string) error {
	cluster, err := GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		return err
	}

	cluster.SystemID = systemID
	err = GetStorageModel().UpdateCluster(context.Background(), cluster)
	if err != nil {
		return err
	}

	return nil
}

// GetBKCloudName get bk cloud name by id
func GetBKCloudName(bkCloudID int) string {
	cli := nodeman.GetNodeManClient()
	if cli == nil {
		return ""
	}
	list, err := cli.CloudList()
	if err != nil {
		blog.Errorf("get cloud list failed, err %s", err.Error())
		return ""
	}
	for _, v := range list {
		if v.BKCloudID == bkCloudID {
			return v.BKCloudName
		}
	}
	return ""
}

// GetModuleName get module name
func GetModuleName(bkBizID, bkModuleID int) string {
	cli := cmdb.GetCmdbClient()
	if cli == nil {
		return ""
	}
	list, err := cli.ListTopology(bkBizID)
	if err != nil {
		blog.Errorf("list topology failed, err %s", err.Error())
		return ""
	}
	if list == nil {
		return ""
	}
	name := ""
	for _, v := range list.Child {
		name = list.BKInstName + " / " + v.BKInstName
		for _, c := range v.Child {
			if c.BKInstID == bkModuleID {
				name += " / " + c.BKInstName
				return name
			}
		}
	}
	return name
}

// StepOptions xxx
type StepOptions struct {
	Retry      uint32
	SkipFailed bool
	TaskName   string
}

// StepOption xxx
type StepOption func(opt *StepOptions)

// WithStepRetry xxx
func WithStepRetry(retry uint32) StepOption {
	return func(opt *StepOptions) {
		opt.Retry = retry
	}
}

// WithStepSkipFailed xxx
func WithStepSkipFailed(skip bool) StepOption {
	return func(opt *StepOptions) {
		opt.SkipFailed = skip
	}
}

// WithStepTaskName xxx
func WithStepTaskName(taskName string) StepOption {
	return func(opt *StepOptions) {
		opt.TaskName = taskName
	}
}

// InitTaskStep init task step
func InitTaskStep(stepInfo StepInfo, opts ...StepOption) *proto.Step {
	defaultOptions := &StepOptions{Retry: 0}
	for _, opt := range opts {
		opt(defaultOptions)
	}
	if defaultOptions.TaskName != "" {
		stepInfo.StepName = defaultOptions.TaskName
	}

	nowStr := time.Now().Format(time.RFC3339)
	return &proto.Step{
		Name:         stepInfo.StepMethod,
		System:       "api",
		Params:       make(map[string]string),
		Retry:        0,
		SkipOnFailed: false,
		Start:        nowStr,
		Status:       TaskStatusNotStarted,
		TaskMethod:   stepInfo.StepMethod,
		TaskName:     stepInfo.StepName,
	}
}
