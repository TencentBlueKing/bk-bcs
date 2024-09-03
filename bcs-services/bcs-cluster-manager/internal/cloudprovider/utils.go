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

package cloudprovider

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/modules"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/alarm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/alarm/bkmonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/alarm/tmp"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/nodeman"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	// BKSOPTask bk-sops common job
	BKSOPTask = "bksopsjob"
	// UnCordonNodesAction 节点可调度任务
	UnCordonNodesAction = "unCordonNodes"
	// CordonNodesAction 节点不可调度任务
	CordonNodesAction = "cordonNodes"
	// WatchTask watch component common job
	WatchTask = "watchjob"
	// InstallImagePullSecretAddonAction imagePull component common job
	InstallImagePullSecretAddonAction = "installImagePullSecret"
	// RemoveHostFromCmdbAction remove host action
	RemoveHostFromCmdbAction = "removeHostFromCmdb"
	// CheckNodeIpsInCmdbAction check node if in cmdb
	CheckNodeIpsInCmdbAction = "checkNodeIpsInCmdb"
	// InstallGseAgentAction install gseAgent action
	InstallGseAgentAction = "installGseAgent"
	// TransferHostModuleAction transfer module action
	TransferHostModuleAction = "transferHostModule"
	// EnsureAutoScalerAction install/update ca component
	EnsureAutoScalerAction = "ensureAutoScaler"
	// JobFastExecuteScriptAction execute script by job
	JobFastExecuteScriptAction = "jobFastExecuteScript"
	// InstallVclusterAction install vcluster
	InstallVclusterAction = "installVcluster"
	// DeleteVclusterAction uninstall vcluster
	DeleteVclusterAction = "deleteVcluster"
	// UpgradeVclusterAction upgrade vcluster
	UpgradeVclusterAction = "upgradeVcluster"
	// CreateNamespaceAction 创建命名空间任务
	CreateNamespaceAction = "createNamespace"
	// DeleteNamespaceAction 删除命名空间任务
	DeleteNamespaceAction = "deleteNamespace"
	// SetNodeLabelsAction 节点设置labels任务
	SetNodeLabelsAction = "nodeSetLabels"
	// SetNodeTaintsAction 节点设置labels任务
	SetNodeTaintsAction = "nodeSetTaints"
	// SetNodeAnnotationsAction 节点设置Annotations任务
	SetNodeAnnotationsAction = "nodeSetAnnotations"
	// CheckKubeAgentStatusAction 检测agent组件状态
	CheckKubeAgentStatusAction = "checkAgentStatus"
	// CreateResourceQuotaAction 创建资源配额任务
	CreateResourceQuotaAction = "createResourceQuota"
	// DeleteResourceQuotaAction 删除资源配额任务
	DeleteResourceQuotaAction = "deleteResourceQuota"
	// ResourcePoolLabelAction 设置资源池标签
	ResourcePoolLabelAction = "resourcePoolLabel"
	// CheckClusterCleanNodesAction 检测集群销毁节点状态
	CheckClusterCleanNodesAction = "checkClusterCleanNodes"
	// RemoveClusterNodesInnerTaintAction remove nodes inner taints
	RemoveClusterNodesInnerTaintAction = "removeClusterNodesInnerTaint"
	// LadderResourcePoolLabelAction 标签设置
	LadderResourcePoolLabelAction = "yunti-ResourcePoolLabelTask"
	// AddNodesShieldAlarmAction 屏蔽机器告警
	AddNodesShieldAlarmAction = "addNodesShieldAlarm"
)

var (
	defaultTaskID = "qwertyuiop123456"
	// TaskID inject taskID into ctx
	TaskID = "taskID"
	// StepNameKey inject stepName into ctx
	StepNameKey = "stepName"
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
	// NOCC:golint/type(设计如此)
	return context.WithValue(ctx, TaskID, taskID) // nolint
}

// GetTaskIDAndStepNameFromContext get taskID and stepName from context
func GetTaskIDAndStepNameFromContext(ctx context.Context) (taskID, stepName string) {
	if id, ok := ctx.Value(TaskID).(string); ok {
		taskID = id
	}

	if name, ok := ctx.Value(StepNameKey).(string); ok {
		stepName = name
	}

	return
}

// WithTaskIDAndStepNameForContext will return a new context wrapped taskID and stepName flag around the original ctx
func WithTaskIDAndStepNameForContext(ctx context.Context, taskID, stepName string) context.Context {
	// NOCC:golint/type(设计如此)
	ctx = context.WithValue(ctx, TaskID, taskID)         // nolint
	return context.WithValue(ctx, StepNameKey, stepName) // nolint
}

// CredentialData dependency data
type CredentialData struct {
	// Cloud cloud
	Cloud *proto.Cloud
	// Cluster cluster
	AccountID string
}

// GetCredential get specified credential information according Cloud configuration,
// if Cloud conf is nil, try Cluster Account.
// @return CommonOption: option can be nil if no credential conf in cloud or cluster account or
// when cloudprovider don't support authentication
// GetCredential get cloud credential by cloud or cluster
func GetCredential(data *CredentialData) (*CommonOption, error) {
	if data.Cloud == nil && data.AccountID == "" {
		return nil, fmt.Errorf("lost cloud/account information")
	}

	option := &CommonOption{}

	// if credential not exist account, get from common cloud
	if data.AccountID != "" {
		// try to get credential in cluster
		account, err := GetStorageModel().GetCloudAccount(context.Background(),
			data.Cloud.CloudID, data.AccountID, false)
		if err != nil {
			return nil, fmt.Errorf("GetCloudAccount failed: %v", err)
		}
		option.Account = account.Account
	}

	// get credential from cloud
	if option.Account == nil && data.Cloud.CloudCredential != nil {
		option.Account = &proto.Account{
			SecretID:             data.Cloud.CloudCredential.Key,
			SecretKey:            data.Cloud.CloudCredential.Secret,
			SubscriptionID:       data.Cloud.CloudCredential.SubscriptionID,
			TenantID:             data.Cloud.CloudCredential.TenantID,
			ResourceGroupName:    data.Cloud.CloudCredential.ResourceGroupName,
			ClientID:             data.Cloud.CloudCredential.ClientID,
			ClientSecret:         data.Cloud.CloudCredential.ClientSecret,
			ServiceAccountSecret: data.Cloud.CloudCredential.ServiceAccountSecret,
			GkeProjectID:         data.Cloud.CloudCredential.GkeProjectID,
		}
	}

	// set cloud basic confInfo
	option.CommonConf = CloudConf{
		CloudInternalEnable: data.Cloud.ConfInfo.CloudInternalEnable,
		CloudDomain:         data.Cloud.ConfInfo.CloudDomain,
		MachineDomain:       data.Cloud.ConfInfo.MachineDomain,
		VpcDomain:           data.Cloud.ConfInfo.VpcDomain,
	}

	// check cloud credential info
	err := checkCloudCredentialValidate(data.Cloud, option)
	if err != nil {
		return nil, fmt.Errorf("checkCloudCredentialValidate %s failed: %v", data.Cloud.CloudProvider, err)
	}

	return option, nil
}

// GetCloudCmOptionByCluster get common option by cluster
func GetCloudCmOptionByCluster(cls proto.Cluster) (*CommonOption, error) {
	cloud, err := GetStorageModel().GetCloud(context.Background(), cls.GetProvider())
	if err != nil {
		blog.Errorf("GetCloudCmOptionByCluster[%s:%s] get cloud failed: %v",
			cls.GetClusterID(), cls.GetProvider(), err)
		return nil, err
	}
	cmOption, err := GetCredential(&CredentialData{
		Cloud:     cloud,
		AccountID: cls.GetCloudAccountID(),
	})
	if err != nil {
		blog.Errorf("getCredential for cloudprovider[%s] when GetCloudCmOptionByCluster[%s:%s] failed, %s",
			cloud.CloudID, cls.ClusterID, cls.Region, err.Error())
		return nil, err
	}
	cmOption.Region = cls.GetRegion()

	return cmOption, nil
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

// CloudDependBasicInfo cloud depend cluster info
type CloudDependBasicInfo struct {
	// Cluster info
	Cluster *proto.Cluster
	// Cloud info
	Cloud *proto.Cloud
	// NodeGroup info
	NodeGroup *proto.NodeGroup
	// NodeTemplate info
	NodeTemplate *proto.NodeTemplate
	// CmOption option
	CmOption *CommonOption
}

// GetBasicInfoReq getDependBasicInfo, clusterID and cloudID must be not empty
type GetBasicInfoReq struct {
	ClusterID      string
	CloudID        string
	NodeGroupID    string
	NodeTemplateID string
}

// GetClusterDependBasicInfo get cluster/cloud/nodeGroup depend info, nodeGroup may be nil.
// only get metadata, try not to change it
func GetClusterDependBasicInfo(request GetBasicInfoReq) (*CloudDependBasicInfo, error) {
	var (
		cluster      *proto.Cluster
		cloud        *proto.Cloud
		nodeGroup    *proto.NodeGroup
		nodeTemplate *proto.NodeTemplate
		err          error
	)

	cloud, cluster, err = actions.GetCloudAndCluster(GetStorageModel(), request.CloudID, request.ClusterID)
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

	if len(request.NodeGroupID) > 0 {
		nodeGroup, err = actions.GetNodeGroupByGroupID(GetStorageModel(), request.NodeGroupID)
		if err != nil {
			return nil, err
		}
	}
	if len(request.NodeTemplateID) > 0 {
		nodeTemplate, err = actions.GetNodeTemplateByTemplateID(GetStorageModel(), request.NodeTemplateID)
		if err != nil {
			return nil, err
		}
	}

	return &CloudDependBasicInfo{cluster, cloud, nodeGroup,
		nodeTemplate, cmOption}, nil
}

// UpdateClusterStatus set cluster status
func UpdateClusterStatus(clusterID string, status string) (*proto.Cluster, error) {
	cluster, err := GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		return nil, err
	}

	cluster.Status = status
	err = GetStorageModel().UpdateCluster(context.Background(), cluster)
	if err != nil {
		return nil, err
	}

	return cluster, nil
}

// UpdateNodeGroupStatus set nodegroup status
func UpdateNodeGroupStatus(nodeGroupID, status string) error {
	group, err := GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		return err
	}

	group.Status = status
	err = GetStorageModel().UpdateNodeGroup(context.Background(), group)
	if err != nil {
		return err
	}

	return nil
}

// GetClusterByID get cluster by clusterID
func GetClusterByID(clusterID string) (*proto.Cluster, error) {
	return GetStorageModel().GetCluster(context.Background(), clusterID)
}

// GetNodeGroupByID get nodeGroup by groupID
func GetNodeGroupByID(nodeGroupId string) (*proto.NodeGroup, error) {
	return GetStorageModel().GetNodeGroup(context.Background(), nodeGroupId)
}

// UpdateCluster set cluster status
func UpdateCluster(cluster *proto.Cluster) error {
	err := GetStorageModel().UpdateCluster(context.Background(), cluster)
	if err != nil {
		return err
	}

	return nil
}

// GetClusterCredentialByClusterID get cluster credential what agent report
func GetClusterCredentialByClusterID(ctx context.Context, clusterID string) (bool, error) {
	_, exist, err := GetStorageModel().GetClusterCredential(ctx, clusterID)
	if err != nil {
		blog.Errorf("GetClusterCredentialByClusterID[%s] failed: %v", clusterID, err)
		return false, err
	}

	return exist, nil
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

	if server == "" || (token == "" && (clientCert == "" || clientKey == "")) {
		return fmt.Errorf("importClusterCredential parse kubeConfig "+
			"failed: %v", "[server|token｜clientCert] empty")
	}

	// need to handle crypt
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

func getNodeNormalStatus(extraStatus []string) []string {
	defaultStatus := []string{common.StatusRunning, common.StatusInitialization,
		common.StatusAddNodesFailed, common.StatusResourceApplyFailed,
		common.StatusDeleting}

	if len(extraStatus) > 0 {
		defaultStatus = append(defaultStatus, extraStatus...)
	}

	return defaultStatus
}

func getClusterOrPoolNodes(clusterId, nodePoolId string) ([]*proto.Node, error) {
	condM := make(operator.M)
	condM["clusterid"] = clusterId

	if len(nodePoolId) > 0 {
		condM["nodegroupid"] = nodePoolId
	}
	cond := operator.NewLeafCondition(operator.Eq, condM)
	return GetStorageModel().ListNode(context.Background(), cond, &storeopt.ListOption{})
}

// ListNodesInClusterNodePool list nodeGroup nodes
func ListNodesInClusterNodePool(clusterID, nodePoolID string) ([]*proto.Node, error) {
	nodes, err := getClusterOrPoolNodes(clusterID, nodePoolID)
	if err != nil {
		blog.Errorf("ListNodesInClusterNodePool NodeGroup %s all Nodes failed, %s", nodePoolID, err.Error())
		return nil, err
	}

	// sum running & creating nodes, these status are ready to serve workload
	var (
		goodNodes []*proto.Node
	)
	for _, node := range nodes {
		if utils.StringInSlice(node.Status, getNodeNormalStatus(nil)) {
			goodNodes = append(goodNodes, node)
		}
	}

	return goodNodes, nil
}

// ListClusterNodes list cluster nodes (运行中/上架中状态)
func ListClusterNodes(clusterID string) ([]*proto.Node, error) {
	nodes, err := getClusterOrPoolNodes(clusterID, "")
	if err != nil {
		blog.Errorf("ListClusterNodes %s all Nodes failed, %s", clusterID, err.Error())
		return nil, err
	}

	// sum running & creating nodes, these status are ready to serve workload
	var (
		goodNodes []*proto.Node
	)
	for _, node := range nodes {
		if utils.StringInSlice(node.Status, getNodeNormalStatus([]string{common.StatusResourceApplying})) {
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
			desiredNodes := taskList[i].CommonParams[ScalingNodesNumKey.String()]
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
			group.AutoScaling.DesiredSize -= uint32(nodeNum)
		} else {
			group.AutoScaling.DesiredSize = 0
			blog.Warnf("updateNodeGroupDesiredSize abnormal, desiredSize[%v] scaleNodesNum[%v]",
				group.AutoScaling.DesiredSize, nodeNum)
		}
	} else {
		group.AutoScaling.DesiredSize += uint32(nodeNum)
	}

	err = GetStorageModel().UpdateNodeGroup(context.Background(), group)
	if err != nil {
		blog.Errorf("updateNodeGroupDesiredSize failed when CA scale nodes: %v", err)
		return err
	}

	return nil
}

// SaveNodeInfoToDB save node to DB
func SaveNodeInfoToDB(ctx context.Context, node *proto.Node, isIP bool) error {
	var (
		oldNode *proto.Node
		err     error
	)
	taskID := GetTaskIDFromContext(ctx)

	if isIP {
		oldNode, err = GetStorageModel().GetNodeByIP(context.Background(), node.InnerIP)
	} else {
		oldNode, err = GetStorageModel().GetNode(context.Background(), node.NodeID)
	}
	blog.Infof("SaveNodeInfoToDB[%s] node[%s:%s] node[%+v] err: %v", taskID, node.InnerIP, node.NodeID, oldNode, err)

	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		return fmt.Errorf("saveNodeInfoToDB[%s] getNode[%s] failed: %v", taskID, node.NodeID, err)
	}

	if oldNode == nil {
		// check repeated cluster ips
		inDb, inCluster := checkRepeatedNodes(ctx, node)
		blog.Infof("SaveNodeInfoToDB[%s] cluster[%s] nodeGroup[%s] checkRepeatedNodes[%+v:%+v]",
			taskID, node.ClusterID, node.NodeGroupID, inDb, inCluster)

		if inDb && !inCluster {
			err = GetStorageModel().DeleteNodeByIP(context.Background(), node.InnerIP)
			if err != nil {
				return err
			}
		}

		err = GetStorageModel().CreateNode(context.Background(), node)
		if err != nil {
			return fmt.Errorf("saveNodeInfoToDB[%s] createNode[%s] failed: %v", taskID, node.InnerIP, err)
		}

		blog.Infof("saveNodeInfoToDB[%s] createNode[%s:%s] success", taskID, node.InnerIP, node.NodeID)

		return nil
	}

	blog.Infof("saveNodeInfoToDB[%s] exist node[%s:%s]", taskID, node.InnerIP, node.NodeID)
	err = GetStorageModel().UpdateNode(context.Background(), node)
	if err != nil {
		return fmt.Errorf("saveNodeInfoToDB updateNode[%s] failed: %v", node.InnerIP, err)
	}

	return nil
}

// checkRepeatedNodes check ip repeated: ip in db, ip in cluster
func checkRepeatedNodes(ctx context.Context, n *proto.Node) (bool, bool) {
	var (
		ip = n.InnerIP
	)

	taskID := GetTaskIDFromContext(ctx)

	if ip != "" {
		existNode, err := GetStorageModel().GetNodeByIP(context.Background(), ip)
		if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Infof("checkRepeatedNodes[%s] GetNodeByIP[%s] failed: %v", taskID, ip, err)
			return false, false
		}
		// db not exist ip
		if existNode == nil {
			blog.Infof("checkRepeatedNodes[%s] IP[%s] not exist db", taskID, ip)
			return false, false
		}

		// check ip exist in cluster
		clusterID := existNode.ClusterID
		if clusterID == "" {
			blog.Infof("checkRepeatedNodes[%s] IP[%s] clusterID empty", taskID, ip)
			return true, false
		}
		found := checkNodeExistInCluster(clusterID, ip)
		if found {
			blog.Infof("checkRepeatedNodes[%s] IP[%s] exist in cluster[%s]", taskID, ip, clusterID)
			return true, true
		}
		blog.Infof("checkRepeatedNodes[%s] IP[%s] not exist in cluster[%s]", taskID, ip, clusterID)

		return true, false
	}

	return false, false
}

func checkNodeExistInCluster(clusterID, ip string) bool {
	if clusterID == "" || ip == "" {
		return false
	}

	found := false

	k8sClient := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), GetStorageModel())
	node, _ := k8sClient.GetClusterNode(context.Background(), clusterops.QueryNodeOption{
		ClusterID: clusterID,
		NodeIP:    ip,
	})
	if node != nil {
		found = true
		return found
	}

	return found
}

// GetInstanceIPsByID get InstanceIP by NodeID
func GetInstanceIPsByID(ctx context.Context, nodeIDs []string) []string {
	var (
		nodeIPs = make([]string, 0)
		taskID  = GetTaskIDFromContext(ctx)
	)

	for _, id := range nodeIDs {
		node, err := GetStorageModel().GetNode(context.Background(), id)
		if err != nil {
			blog.Errorf("GetInstanceIPsByID[%s] nodeID[%s] failed: %v", taskID, id, err)
			continue
		}

		nodeIPs = append(nodeIPs, node.InnerIP)
	}

	return nodeIPs
}

// GetInstanceIPsByName get InstanceIP by NodeName
func GetInstanceIPsByName(ctx context.Context, clusterID string, nodeNames []string) []string {
	var (
		taskID  = GetTaskIDFromContext(ctx)
		nodeIPs = make([]string, 0)
	)

	for _, name := range nodeNames {
		node, err := GetStorageModel().GetNodeByName(context.Background(), clusterID, name)
		if err != nil {
			blog.Errorf("GetInstanceIPsByName[%s] nodeName[%s] failed: %v", taskID, name, err)
			continue
		}

		nodeIPs = append(nodeIPs, node.InnerIP)
	}

	return nodeIPs
}

// GetNodesByInstanceIDs get nodes by instanceIDs
func GetNodesByInstanceIDs(instanceIDs []string) []*proto.Node {
	nodes := make([]*proto.Node, 0)
	for _, id := range instanceIDs {
		node, err := GetStorageModel().GetNode(context.Background(), id)
		if err != nil {
			continue
		}

		nodes = append(nodes, node)
	}

	return nodes
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

// UpdateNodeListStatus update nodeList status
func UpdateNodeListStatus(isInstanceIP bool, instances []string, status string) error {
	for i := range instances {
		err := UpdateNodeStatus(isInstanceIP, instances[i], status)
		if err != nil {
			// batch update if one failed need to handle, other than task failed
			continue
		}
	}

	return nil
}

// GetNodeByIpOrId get node info by ip or id
func GetNodeByIpOrId(isIp bool, instance string) (*proto.Node, error) {
	var (
		node *proto.Node
		err  error
	)
	if isIp {
		node, err = GetStorageModel().GetNodeByIP(context.Background(), instance)
	} else {
		node, err = GetStorageModel().GetNode(context.Background(), instance)
	}
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		return nil, err
	}

	if errors.Is(err, drivers.ErrTableRecordNotFound) {
		return nil, fmt.Errorf("instance[%s] not found", instance)
	}

	return node, nil
}

// UpdateNodeStatus update node status
func UpdateNodeStatus(isInstanceIP bool, instance, status string) error {
	node, err := GetNodeByIpOrId(isInstanceIP, instance)
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

// GetClusterMasterIPList get cluster masterIPs
func GetClusterMasterIPList(cluster *proto.Cluster) []string {
	masterIPs := make([]string, 0)
	for masterIP := range cluster.Master {
		masterIPs = append(masterIPs, masterIP)
	}

	return masterIPs
}

// StepOptions xxx
type StepOptions struct {
	Retry      uint32
	SkipFailed bool
	Translate  string
	AllowSkip  bool
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

// WithStepAllowSkip xxx
func WithStepAllowSkip(allow bool) StepOption {
	return func(opt *StepOptions) {
		opt.AllowSkip = allow
	}
}

// WithStepTranslate xxx
func WithStepTranslate(translate string) StepOption {
	return func(opt *StepOptions) {
		opt.Translate = translate
	}
}

// InitTaskStep init task step
func InitTaskStep(stepInfo StepInfo, opts ...StepOption) *proto.Step {
	defaultOptions := &StepOptions{
		Retry:      0,
		SkipFailed: false,
		Translate:  "",
		AllowSkip:  false,
	}
	for _, opt := range opts {
		opt(defaultOptions)
	}

	nowStr := time.Now().Format(time.RFC3339)
	return &proto.Step{
		Name:         stepInfo.StepMethod,
		System:       "api",
		Params:       make(map[string]string),
		Retry:        0,
		SkipOnFailed: defaultOptions.SkipFailed,
		Start:        nowStr,
		Status:       TaskStatusNotStarted,
		TaskMethod:   stepInfo.StepMethod,
		TaskName:     stepInfo.StepName,
		Translate:    defaultOptions.Translate,
		AllowSkip:    defaultOptions.AllowSkip,
	}
}

// GetIDToIPMap get instanceID to instanceIP map
func GetIDToIPMap(nodeIDs, nodeIPs []string) map[string]string {
	idToIPMap := make(map[string]string, 0)
	for i := range nodeIDs {
		if i < len(nodeIPs) {
			idToIPMap[nodeIDs[i]] = nodeIPs[i]
		}
	}

	return idToIPMap
}

// GetNodeIdToIpMapByNodeIds get nodeId mapTo nodeIp
func GetNodeIdToIpMapByNodeIds(nodeIds []string) map[string]string {
	nodes := GetNodesByInstanceIDs(nodeIds)

	idToIpmap := make(map[string]string, 0)
	for i := range nodes {
		idToIpmap[nodes[i].GetNodeID()] = nodes[i].GetInnerIP()
	}

	return idToIpmap
}

// IsExternalNodePool check group external nodePool
func IsExternalNodePool(group *proto.NodeGroup) bool {
	if group == nil {
		return false
	}
	switch group.GetNodeGroupType() {
	case common.External.String():
		return true
	case common.Normal.String(), "":
		return false
	}

	return false
}

// ParseNodeIpOrIdFromCommonMap parse nodeIDs or nodeIPs by chart
func ParseNodeIpOrIdFromCommonMap(taskCommonMap map[string]string, key string, chart string) []string {
	val, ok := taskCommonMap[key]
	if !ok || val == "" {
		return nil
	}

	return strings.Split(val, chart)
}

// ParseMapFromStepParas from step parse k1=v1;k2=v2; to map
func ParseMapFromStepParas(stepMap map[string]string, key string) map[string]string {
	val, ok := stepMap[key]
	if !ok || val == "" {
		return nil
	}

	return utils.StringsToMap(val)
}

// GetScaleOutModuleID get scaleOut module ID
func GetScaleOutModuleID(cls *proto.Cluster, asOption *proto.ClusterAutoScalingOption,
	template *proto.NodeTemplate, isGroup bool) string {
	if template != nil && template.Module != nil && template.Module.ScaleOutModuleID != "" {
		return template.Module.ScaleOutModuleID
	}
	if isGroup && len(cls.GetModuleID()) > 0 {
		return cls.GetModuleID()
	}
	if asOption != nil && asOption.Module != nil && asOption.Module.ScaleOutModuleID != "" {
		return asOption.Module.ScaleOutModuleID
	}
	if cls.GetClusterBasicSettings().GetModule().GetWorkerModuleID() != "" {
		return cls.GetClusterBasicSettings().GetModule().GetWorkerModuleID()
	}

	return ""
}

// GetScaleInModuleID get scaleIn module ID only from template
func GetScaleInModuleID(asOption *proto.ClusterAutoScalingOption, template *proto.NodeTemplate) string {
	if template != nil && template.Module != nil && template.Module.ScaleInModuleID != "" {
		return template.Module.ScaleInModuleID
	}
	if asOption != nil && asOption.Module != nil && asOption.Module.ScaleInModuleID != "" {
		return asOption.Module.ScaleInModuleID
	}

	return ""
}

// GetBusinessID get business id, default cluster business id
func GetBusinessID(cls *proto.Cluster, asOption *proto.ClusterAutoScalingOption,
	template *proto.NodeTemplate, scale bool) string {
	getBizID := func(scale bool, scaleOut, scaleIn string) string {
		switch scale {
		case true:
			return scaleOut
		case false:
			return scaleIn
		}
		return ""
	}

	if template != nil && template.Module != nil {
		return getBizID(scale, template.Module.ScaleOutBizID, template.Module.ScaleInBizID)
	}

	if asOption != nil && asOption.Module != nil {
		return getBizID(scale, asOption.Module.ScaleOutBizID, asOption.Module.ScaleInBizID)
	}

	return cls.GetBusinessID()
}

// GetBKCloudName get bk cloud name by id
func GetBKCloudName(bkCloudID int) string {
	cli := nodeman.GetNodeManClient()
	if cli == nil {
		return ""
	}
	list, err := cli.CloudList(context.Background())
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
	list, err := cli.ListTopology(context.Background(), bkBizID, false, false)
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

// GetBizMaintainers get biz maintainers
func GetBizMaintainers(bkBizID int) string {
	cli := cmdb.GetCmdbClient()
	if cli == nil {
		return ""
	}

	bizData, err := cli.GetBusinessMaintainer(bkBizID)
	if err != nil {
		blog.Errorf("GetBizMaintainers failed: %v", err.Error())
		return ""
	}

	return bizData.BKBizMaintainer
}

// IsMasterIp check ip if is cluster master
func IsMasterIp(ip string, cls *proto.Cluster) bool {
	if cls == nil || cls.Master == nil {
		return false
	}

	_, ok := cls.Master[ip]
	return ok
}

// UpdateNodeGroupCloudAndModuleInfo update cloudID && moduleInfo
func UpdateNodeGroupCloudAndModuleInfo(nodeGroupID string, cloudGroupID string,
	consumer bool, clusterBiz string) error {
	group, err := GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		return err
	}
	if consumer {
		group.ConsumerID = cloudGroupID
	} else {
		group.CloudNodeGroupID = cloudGroupID
	}

	// update group module info
	if group.NodeTemplate != nil && group.NodeTemplate.Module != nil {
		if group.NodeTemplate.Module.ScaleOutBizID == "" {
			group.NodeTemplate.Module.ScaleOutBizID = clusterBiz
		}
		if group.NodeTemplate.Module.ScaleInBizID == "" {
			group.NodeTemplate.Module.ScaleInBizID = clusterBiz
		}
		if group.NodeTemplate.Module.ScaleOutModuleID != "" {
			scaleOutBiz, _ := strconv.Atoi(group.NodeTemplate.Module.ScaleOutBizID)
			scaleOutModule, _ := strconv.Atoi(group.NodeTemplate.Module.ScaleOutModuleID)
			group.NodeTemplate.Module.ScaleOutModuleName = GetModuleName(scaleOutBiz, scaleOutModule)
		}
		if group.NodeTemplate.Module.ScaleInModuleID != "" {
			scaleInBiz, _ := strconv.Atoi(group.NodeTemplate.Module.ScaleInBizID)
			scaleInModule, _ := strconv.Atoi(group.NodeTemplate.Module.ScaleInModuleID)
			group.NodeTemplate.Module.ScaleInModuleName = GetModuleName(scaleInBiz, scaleInModule)
		}
	}
	err = GetStorageModel().UpdateNodeGroup(context.Background(), group)
	if err != nil {
		return err
	}

	return nil
}

// ShieldHostAlarm shield host alarm for user
func ShieldHostAlarm(ctx context.Context, bizID string, ips []string) error {
	taskID, stepName := GetTaskIDAndStepNameFromContext(ctx)
	if len(ips) == 0 {
		return fmt.Errorf("ShieldHostAlarm[%s] ips empty", taskID)
	}

	biz, _ := strconv.Atoi(bizID)
	bizData, err := cmdb.GetCmdbClient().GetBusinessMaintainer(biz)
	if err != nil {
		blog.Errorf("ShieldHostAlarm[%s] GetBusinessMaintainer[%s] failed: %v", taskID, bizID, err)
		return err
	}
	maintainers := strings.Split(bizData.BKBizMaintainer, ",")
	if len(maintainers) == 0 {
		return fmt.Errorf("ShieldHostAlarm[%s] BKBizMaintainer[%s] empty", taskID, bizID)
	}

	hostData, err := cmdb.GetCmdbClient().QueryAllHostInfoWithoutBiz(ips)
	if err != nil {
		blog.Errorf("ShieldHostAlarm[%s] QueryAllHostInfoWithoutBiz[%+v] failed: %v", taskID, ips, err)
		return err
	}

	hosts := make([]alarm.HostInfo, 0)
	for i := range hostData {
		hosts = append(hosts, alarm.HostInfo{
			IP:      hostData[i].BKHostInnerIP,
			CloudID: uint64(hostData[i].BkCloudID),
		})
	}
	blog.Infof("ShieldHostAlarm[%s] bizID[%s] hostInfo[%+v]", taskID, bizID, hosts)

	var alarms = []alarm.AlarmInterface{tmp.GetBKAlarmClient(), bkmonitor.GetBkMonitorClient()}

	for i := range alarms {
		err = alarms[i].ShieldHostAlarmConfig(maintainers[0], &alarm.ShieldHost{
			BizID:    bizID,
			HostList: hosts,
		})
		if err != nil {
			blog.Errorf("ShieldHostAlarm[%s][%s] ShieldHostAlarmConfig failed: %v", taskID, alarms[i].Name(), err)
			continue
		}

		GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
			fmt.Sprintf("[%s] successful", alarms[i].Name()))

		blog.Infof("ShieldHostAlarm[%s][%s] ShieldHostAlarmConfig success", taskID, alarms[i].Name())
	}

	return nil
}

// UpdateAutoScalingOptionModuleInfo update cluster ca moduleInfo
func UpdateAutoScalingOptionModuleInfo(clusterID string) error {
	cls, err := GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		return err
	}

	asOption, err := GetStorageModel().GetAutoScalingOption(context.Background(), clusterID)
	if err != nil {
		return err
	}
	// update asOption module info
	if asOption.Module != nil {
		if asOption.Module.ScaleOutBizID == "" {
			asOption.Module.ScaleOutBizID = cls.BusinessID
		}
		if asOption.Module.ScaleInBizID == "" {
			asOption.Module.ScaleInBizID = cls.BusinessID
		}
		if asOption.Module.ScaleOutModuleID != "" {
			scaleOutBiz, _ := strconv.Atoi(asOption.Module.ScaleOutBizID)
			scaleOutModule, _ := strconv.Atoi(asOption.Module.ScaleOutModuleID)
			asOption.Module.ScaleOutModuleName = GetModuleName(scaleOutBiz, scaleOutModule)
		}
		if asOption.Module.ScaleInModuleID != "" {
			scaleInBiz, _ := strconv.Atoi(asOption.Module.ScaleInBizID)
			scaleInModule, _ := strconv.Atoi(asOption.Module.ScaleInModuleID)
			asOption.Module.ScaleInModuleName = GetModuleName(scaleInBiz, scaleInModule)
		}
	}
	err = GetStorageModel().UpdateAutoScalingOption(context.Background(), asOption)
	if err != nil {
		return err
	}

	return nil
}

// IsInDependentCluster check independent cluster
func IsInDependentCluster(cluster *proto.Cluster) bool {
	return cluster.ManageType == common.ClusterManageTypeIndependent
}

// IsManagedCluster check managed cluster
func IsManagedCluster(cluster *proto.Cluster) bool {
	return cluster.ManageType == common.ClusterManageTypeManaged
}

// GetCRDByKubeConfig get crd by kubeConfig
func GetCRDByKubeConfig(kubeConfig string) (*v1.CustomResourceDefinitionList, error) {
	_, err := types.GetKubeConfigFromYAMLBody(false, types.YamlInput{
		FileName:    "",
		YamlContent: kubeConfig,
	})

	if err != nil {
		return nil, fmt.Errorf("checkKubeConfig get kubeConfig from YAML body failed: %v", err)
	}

	// 解析 kubeConfig 字符串
	cfg, err := clientcmd.NewClientConfigFromBytes([]byte(kubeConfig))
	if err != nil {
		return nil, err
	}

	// 获取 Kubernetes 配置
	config, err := cfg.ClientConfig()
	if err != nil {
		return nil, err
	}

	// 使用 Kubernetes 配置创建一个 Kubernetes 客户端
	cli, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// 获取 CRD
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	return cli.ApiextensionsV1().CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
}

// UpdateVirtualNodeStatus update virtual nodes status
func UpdateVirtualNodeStatus(clusterId, nodeGroupId, taskID string) error {
	if clusterId == "" || nodeGroupId == "" || taskID == "" {
		blog.Infof("UpdateVirtualNodeStatus[%s] validate data", taskID)
		return nil
	}

	condM := make(operator.M)
	condM["nodegroupid"] = nodeGroupId
	condM["clusterid"] = clusterId
	condM["taskid"] = taskID
	cond := operator.NewLeafCondition(operator.Eq, condM)

	nodes, err := GetStorageModel().ListNode(context.Background(), cond, &storeopt.ListOption{})
	if err != nil {
		blog.Errorf("UpdateVirtualNodeStatus[%s] NodeGroup %s all Nodes failed, %s",
			taskID, nodeGroupId, err.Error())
		return err
	}

	blog.Infof("UpdateVirtualNodeStatus[%s] ListNodes[%+v] success", taskID, nodes)
	for i := range nodes {
		blog.Infof("UpdateVirtualNodeStatus[%s] node status", nodes[i].NodeID)
		nodes[i].Status = common.StatusResourceApplyFailed
		err = GetStorageModel().UpdateNode(context.Background(), nodes[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteVirtualNodes delete virtual nodes
func DeleteVirtualNodes(clusterId, nodeGroupId, taskID string) error {
	if clusterId == "" || nodeGroupId == "" || taskID == "" {
		blog.Infof("DeleteVirtualNodes[%s] validate data", taskID)
		return nil
	}

	condM := make(operator.M)
	condM["nodegroupid"] = nodeGroupId
	condM["clusterid"] = clusterId
	condM["taskid"] = taskID
	cond := operator.NewLeafCondition(operator.Eq, condM)

	nodes, err := GetStorageModel().ListNode(context.Background(), cond, &storeopt.ListOption{})
	if err != nil {
		blog.Errorf("ListNodesInClusterNodePool[%s] NodeGroup %s all Nodes failed, %s",
			taskID, nodeGroupId, err.Error())
		return err
	}

	blog.Infof("DeleteVirtualNodes[%s] ListNodes[%+v] success", taskID, nodes)
	for i := range nodes {
		blog.Infof("DeleteVirtualNodes[%s] node[%s] status", taskID, nodes[i].NodeID)

		if !strings.HasPrefix(nodes[i].GetNodeID(), "bcs") {
			continue
		}
		err = GetStorageModel().DeleteNode(context.Background(), nodes[i].GetNodeID())
		if err != nil {
			return err
		}
	}

	return nil
}

// GetAnnotationsByNg get annotations by nodeGroup
func GetAnnotationsByNg(group *proto.NodeGroup) map[string]string {
	if group == nil || group.NodeTemplate == nil || len(group.NodeTemplate.Annotations) == 0 {
		return nil
	}

	return group.GetNodeTemplate().GetAnnotations()
}

// GetLabelsByNg get labels by nodeGroup
func GetLabelsByNg(group *proto.NodeGroup) map[string]string {
	if group == nil || group.NodeTemplate == nil || len(group.NodeTemplate.Labels) == 0 {
		return nil
	}

	return group.GetNodeTemplate().GetLabels()
}

// GetTaintsByNg get taints by nodeGroup
func GetTaintsByNg(group *proto.NodeGroup) []*proto.Taint {
	if group == nil || group.NodeTemplate == nil || len(group.NodeTemplate.Taints) == 0 {
		return nil
	}

	return group.GetNodeTemplate().GetTaints()
}

// GetTransModuleInfo get trans moduleID
func GetTransModuleInfo(cls *proto.Cluster, asOption *proto.ClusterAutoScalingOption, group *proto.NodeGroup) string {
	if group != nil && group.NodeTemplate != nil && group.NodeTemplate.Module != nil &&
		len(group.NodeTemplate.Module.ScaleOutModuleID) != 0 {
		return group.NodeTemplate.Module.ScaleOutModuleID
	}

	if asOption != nil && asOption.GetModule() != nil && asOption.GetModule().GetScaleInModuleID() != "" {
		return asOption.GetModule().GetScaleOutModuleID()
	}

	return cls.GetClusterBasicSettings().GetModule().GetWorkerModuleID()
}

// UpdateClusterErrMessage update cluster failed message
func UpdateClusterErrMessage(clusterId string, message string) error {
	cluster, errLocal := GetClusterByID(clusterId)
	if errLocal == nil {
		// record cluster connect failed reason
		cluster.Message = message
		_ = UpdateCluster(cluster)

		return nil
	}

	return errLocal
}

// UpdateNodeGroupCloudNodeGroupID set nodegroup cloudNodeGroupID
func UpdateNodeGroupCloudNodeGroupID(nodeGroupID string, newGroup *proto.NodeGroup) error {
	group, err := GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		return err
	}

	group.CloudNodeGroupID = newGroup.CloudNodeGroupID
	if group.AutoScaling != nil && group.AutoScaling.VpcID == "" {
		group.AutoScaling.VpcID = newGroup.AutoScaling.VpcID
	}
	if group.LaunchTemplate != nil {
		group.LaunchTemplate.InstanceChargeType = newGroup.LaunchTemplate.InstanceChargeType
	}
	err = GetStorageModel().UpdateNodeGroup(context.Background(), group)
	if err != nil {
		return err
	}

	return nil
}

// GetClusterResourceGroup cluster resource group
func GetClusterResourceGroup(cls *proto.Cluster) string {
	if cls.GetExtraInfo() == nil {
		return ""
	}

	rg, ok := cls.GetExtraInfo()[common.ClusterResourceGroup]
	if ok {
		return rg
	}

	return ""
}

// GetNodeResourceGroup other resource group
func GetNodeResourceGroup(cls *proto.Cluster) string {
	if cls.GetExtraInfo() == nil {
		return ""
	}

	rg, ok := cls.GetExtraInfo()[common.NodeResourceGroup]
	if ok {
		return rg
	}

	return ""
}

// GetNetworkResourceGroup other resource group
func GetNetworkResourceGroup(cls *proto.Cluster) string {
	if cls.GetExtraInfo() == nil {
		return ""
	}

	rg, ok := cls.GetExtraInfo()[common.NetworkResourceGroup]
	if ok {
		return rg
	}

	return ""
}

// ListProjectNotifyTemplates list project notify templates
func ListProjectNotifyTemplates(projectId string) ([]proto.NotifyTemplate, error) {
	condM := make(operator.M)
	condM["projectid"] = projectId

	cond := operator.NewLeafCondition(operator.Eq, condM)
	templates, err := GetStorageModel().ListNotifyTemplate(context.Background(), cond, &storeopt.ListOption{})
	if err != nil {
		return nil, err
	}

	return templates, nil
}

// GetOverlayCidrBlocks get overlayIps from vpc
func GetOverlayCidrBlocks(cloudId, vpcId string) ([]*net.IPNet, error) {
	vpc, err := GetStorageModel().GetCloudVPC(context.Background(), cloudId, vpcId)
	if err != nil {
		return nil, err
	}

	cidrs := make([]string, 0)
	for i := range vpc.GetOverlay().GetCidrs() {
		if vpc.GetOverlay().GetCidrs()[i].GetBlock() {
			continue
		}
		cidrs = append(cidrs, vpc.GetOverlay().GetCidrs()[i].GetCidr())
	}

	var blocks []*net.IPNet
	for _, v := range cidrs {
		_, ipnet, _ := net.ParseCIDR(v)
		blocks = append(blocks, ipnet)
	}
	return blocks, nil
}

// GetCloudByProvider get cloud by provider
func GetCloudByProvider(provider string) (*proto.Cloud, error) {
	return GetStorageModel().GetCloudByProvider(context.Background(), provider)
}
