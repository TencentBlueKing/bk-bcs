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

package cluster

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	cmcommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
)

// UpdateAction action for update cluster
type UpdateAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	req     *cmproto.UpdateClusterReq
	resp    *cmproto.UpdateClusterResp
	cluster *cmproto.Cluster
	cloud   *cmproto.Cloud
}

// NewUpdateAction create update action
func NewUpdateAction(model store.ClusterManagerModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

// validate check
func (ua *UpdateAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}
	if len(ua.req.EngineType) > 0 && !cmcommon.IsEngineTypeValid(ua.req.EngineType) {
		return fmt.Errorf("invalid engine type")
	}
	if len(ua.req.ClusterType) > 0 && !cmcommon.IsClusterTypeValid(ua.req.ClusterType) {
		return fmt.Errorf("invalid cluster type")
	}
	return nil
}

// getCluster cluster/cloud
func (ua *UpdateAction) getCluster() error {
	cluster, err := ua.model.GetCluster(ua.ctx, ua.req.ClusterID)
	if err != nil {
		return err
	}
	ua.cluster = cluster

	cloud, err := ua.model.GetCloud(ua.ctx, cluster.Provider)
	if err != nil {
		return err
	}
	ua.cloud = cloud

	return nil
}

// setBaseInfo 检查基本信息
func (ua *UpdateAction) setBaseInfo() {
	if len(ua.req.ClusterName) != 0 {
		ua.cluster.ClusterName = ua.req.ClusterName
	}
	if len(ua.req.Updater) != 0 {
		ua.cluster.Updater = ua.req.Updater
	}
	if len(ua.req.Provider) != 0 {
		ua.cluster.Provider = ua.req.Provider
	}
	if len(ua.req.Region) != 0 {
		ua.cluster.Region = ua.req.Region
	}
	if len(ua.req.VpcID) != 0 {
		ua.cluster.VpcID = ua.req.VpcID
	}
	if len(ua.req.ProjectID) != 0 {
		ua.cluster.ProjectID = ua.req.ProjectID
	}
	if len(ua.req.BusinessID) != 0 {
		ua.cluster.BusinessID = ua.req.BusinessID
	}
	if len(ua.req.Environment) != 0 {
		ua.cluster.Environment = ua.req.Environment
	}
	if len(ua.req.ClusterType) != 0 {
		ua.cluster.ClusterType = ua.req.ClusterType
	}
	if ua.req.Labels != nil {
		ua.cluster.Labels = ua.req.Labels
	}
	if len(ua.req.Status) != 0 {
		ua.cluster.Status = ua.req.Status
	}
	if len(ua.req.ManageType) != 0 {
		ua.cluster.ManageType = ua.req.ManageType
	}
	if len(ua.req.CreateTime) > 0 {
		ua.cluster.CreateTime = ua.req.CreateTime
	}
	if len(ua.req.Creator) > 0 {
		ua.cluster.Creator = ua.req.Creator
	}
	if len(ua.req.SystemID) != 0 {
		ua.cluster.SystemID = ua.req.SystemID
	}
	if len(ua.req.ModuleID) != 0 {
		ua.cluster.ModuleID = ua.req.ModuleID
	}
	if ua.req.Description != nil {
		ua.cluster.Description = ua.req.Description.GetValue()
	}
}

func (ua *UpdateAction) setSettingInfo() {
	if ua.req.NetworkSettings != nil {
		ua.cluster.NetworkSettings = ua.req.NetworkSettings
	}
	if ua.req.ClusterBasicSettings != nil {
		ua.cluster.ClusterBasicSettings = ua.req.ClusterBasicSettings
	}
	if ua.req.ClusterAdvanceSettings != nil {
		ua.cluster.ClusterAdvanceSettings = ua.req.ClusterAdvanceSettings
	}
	if ua.req.NodeSettings != nil {
		ua.cluster.NodeSettings = ua.req.NodeSettings
	}
	if ua.req.IsMixed != nil {
		ua.cluster.IsMixed = ua.req.IsMixed.GetValue()
	}
	if ua.req.SharedRanges != nil {
		ua.cluster.SharedRanges = ua.req.SharedRanges
	}
}

// setAdditionalInfo 检查附加信息
func (ua *UpdateAction) setAdditionalInfo() {
	if len(ua.req.EngineType) != 0 {
		ua.cluster.EngineType = ua.req.EngineType
	}
	if ua.req.IsExclusive != nil && ua.req.IsExclusive.GetValue() != ua.cluster.IsExclusive {
		ua.cluster.IsExclusive = ua.req.IsExclusive.GetValue()
	}
	if len(ua.req.FederationClusterID) != 0 {
		ua.cluster.FederationClusterID = ua.req.FederationClusterID
	}
	if ua.req.BcsAddons != nil {
		ua.cluster.BcsAddons = ua.req.BcsAddons
	}
	if ua.req.ExtraAddons != nil {
		ua.cluster.ExtraAddons = ua.req.ExtraAddons
	}
	if len(ua.req.ImportCategory) > 0 {
		ua.cluster.ImportCategory = ua.req.ImportCategory
	}
	if len(ua.req.CloudAccountID) > 0 {
		ua.cluster.CloudAccountID = ua.req.CloudAccountID
	}
	if len(ua.req.ExtraInfo) > 0 {
		ua.cluster.ExtraInfo = ua.req.ExtraInfo
	}
	if ua.req.IsCommonCluster != nil {
		ua.cluster.IsCommonCluster = ua.req.IsCommonCluster.GetValue()
	}
	if ua.req.IsShared != nil {
		ua.cluster.IsShared = ua.req.IsShared.GetValue()
	}
	if len(ua.req.ClusterCategory) > 0 {
		ua.cluster.ClusterCategory = ua.req.ClusterCategory
	}
	if len(ua.req.NetworkType) != 0 {
		ua.cluster.NetworkType = ua.req.NetworkType
	}
	if len(ua.req.ExtraClusterID) > 0 {
		ua.cluster.ExtraClusterID = ua.req.ExtraClusterID
	}
}

// updateCluster update cluster info
func (ua *UpdateAction) updateCluster() error {
	// basic info
	ua.setBaseInfo()
	ua.setSettingInfo()
	// additional info
	ua.setAdditionalInfo()
	// update cluster cloud info
	err := utils.UpdateClusterCloudInfo(ua.cluster, ua.cloud)
	if err != nil {
		return err
	}

	// trans masterIPs
	if len(ua.req.Master) > 0 {
		ua.cluster.Master = make(map[string]*cmproto.Node)

		for _, ip := range ua.req.Master {
			node, err := ua.transNodeIPToCloudNode(ip) // nolint
			if err != nil {
				blog.Errorf("updateCluster transNodeIPToCloudNode failed: %v", err)
				ua.cluster.Master[ip] = &cmproto.Node{
					InnerIP: ip,
					Status:  common.StatusRunning,
				}
			} else {
				ua.cluster.Master[ip] = node
			}
		}
	}
	ua.cluster.UpdateTime = time.Now().Format(time.RFC3339)

	// update DB clusterInfo & passcc cluster
	err = ua.model.UpdateCluster(ua.ctx, ua.cluster)
	if err != nil {
		return err
	}

	// save data to passcc
	updatePassCCClusterInfo(ua.cluster)
	return nil
}

func (ua *UpdateAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = code == common.BcsErrClusterManagerSuccess
}

// transCloudNodeToDNodes by req nodeIPs trans to cloud node
func (ua *UpdateAction) transNodeIPToCloudNode(ip string) (*cmproto.Node, error) {
	nodeMgr, err := cloudprovider.GetNodeMgr(ua.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s NodeManager Cluster %s failed, %s",
			ua.cloud.CloudProvider, ua.req.ClusterID, err.Error())
		return nil, err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ua.cloud,
		AccountID: "",
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s cluster %s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, ua.req.ClusterID, err.Error())
		return nil, err
	}
	cmOption.Region = ua.cluster.Region

	// cluster check instance if exist, validate nodes existence
	node, err := nodeMgr.GetNodeByIP(ip, &cloudprovider.GetNodeOption{
		Common:       cmOption,
		ClusterVPCID: ua.cluster.VpcID,
	})
	if err != nil {
		blog.Errorf("validate nodes %s existence failed, %s", ip, err.Error())
		return nil, err
	}
	blog.Infof("get cloud[%s] IP[%s] to Node successfully", ua.cloud.CloudProvider, ip)

	return node, nil
}

// Handle handles update cluster request
func (ua *UpdateAction) Handle(ctx context.Context, req *cmproto.UpdateClusterReq, resp *cmproto.UpdateClusterResp) {
	if req == nil || resp == nil {
		blog.Errorf("update cluster failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ua.getCluster(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if err := ua.updateCluster(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// create operationLog
	err := ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.Cluster.String(),
		ResourceID:   ua.cluster.ClusterID,
		TaskID:       "",
		Message:      fmt.Sprintf("集群%s修改基本信息", ua.cluster.ClusterID),
		OpUser:       auth.GetUserFromCtx(ua.ctx),
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    ua.cluster.ClusterID,
		ProjectID:    ua.cluster.ProjectID,
		ResourceName: ua.cluster.ClusterName,
	})
	if err != nil {
		blog.Errorf("UpdateCluster[%s] CreateOperationLog failed: %v", ua.cluster.ClusterID, err)
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// UpdateNodeAction action for update node
type UpdateNodeAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	req     *cmproto.UpdateNodeRequest
	resp    *cmproto.UpdateNodeResponse
	success []string
	failed  []string
}

// NewUpdateNodeAction create update action
func NewUpdateNodeAction(model store.ClusterManagerModel) *UpdateNodeAction {
	return &UpdateNodeAction{
		model: model,
	}
}

// validate check
func (ua *UpdateNodeAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	if len(ua.req.ClusterID) == 0 && len(ua.req.NodeGroupID) == 0 && len(ua.req.Status) == 0 {
		return fmt.Errorf("UpdateNodeAction validate failed: body empty")
	}

	return nil
}

// updateNodeInfo update node info
func (ua *UpdateNodeAction) updateNodeInfo(nodeIP string) error {
	node, err := ua.model.GetNodeByIP(ua.ctx, nodeIP)
	if err != nil {
		return err
	}
	if len(ua.req.ClusterID) > 0 {
		node.ClusterID = ua.req.ClusterID
	}
	if len(ua.req.NodeGroupID) > 0 {
		node.NodeGroupID = ua.req.NodeGroupID
	}
	if len(ua.req.Status) > 0 {
		node.Status = ua.req.Status
	}

	err = ua.model.UpdateNode(ua.ctx, node)
	if err != nil {
		return err
	}

	return nil
}

func (ua *UpdateNodeAction) updateNodes() error { // nolint
	for _, ip := range ua.req.InnerIPs {
		err := ua.updateNodeInfo(ip)
		if err != nil {
			ua.failed = append(ua.failed, ip)
			continue
		}

		ua.success = append(ua.success, ip)
	}

	return nil
}

func (ua *UpdateNodeAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = code == common.BcsErrClusterManagerSuccess
	if ua.resp.Data == nil {
		ua.resp.Data = &cmproto.NodeStatus{}
	}
	ua.resp.Data.Success = ua.success
	ua.resp.Data.Failed = ua.failed
}

// Handle handles update nodes request
func (ua *UpdateNodeAction) Handle(ctx context.Context, req *cmproto.UpdateNodeRequest,
	resp *cmproto.UpdateNodeResponse) {
	if req == nil || resp == nil {
		blog.Errorf("update cluster node failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ua.updateNodes(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err := ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.Cluster.String(),
		ResourceID:   "",
		TaskID:       "",
		Message:      "更新node信息",
		OpUser:       auth.GetUserFromCtx(ua.ctx),
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    ua.req.ClusterID,
	})
	if err != nil {
		blog.Errorf("UpdateNode CreateOperationLog failed: %v", err)
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// AddNodesAction action for add nodes to cluster
type AddNodesAction struct {
	ctx            context.Context
	model          store.ClusterManagerModel
	req            *cmproto.AddNodesRequest
	resp           *cmproto.AddNodesResponse
	cluster        *cmproto.Cluster
	nodes          []*cmproto.Node
	cloud          *cmproto.Cloud
	task           *cmproto.Task
	option         *cloudprovider.CommonOption
	nodeGroup      *cmproto.NodeGroup
	nodeTemplate   *cmproto.NodeTemplate
	currentNodeCnt uint64
	nodeScheduler  bool
}

// NewAddNodesAction create addNodes action
func NewAddNodesAction(model store.ClusterManagerModel) *AddNodesAction {
	return &AddNodesAction{
		model: model,
	}
}

// addExternalNodesToCluster handle external nodes
func (ua *AddNodesAction) addExternalNodesToCluster() error {
	groupCloud, err := actions.GetCloudByCloudID(ua.model, ua.nodeGroup.Provider)
	if err != nil {
		blog.Errorf("AddNodesAction addExternalNodesToCluster failed: %v", err)
		return err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     groupCloud,
		AccountID: ua.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s when add nodes %s to cluster %s failed, %s",
			groupCloud.CloudID, groupCloud.CloudProvider, ua.req.Nodes, ua.req.ClusterID, err.Error(),
		)
		return err
	}
	cmOption.Region = ua.nodeGroup.Region

	// get external node implement nodeMgr
	nodeGroupMgr, err := cloudprovider.GetNodeGroupMgr(groupCloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s NodeGroupManager for add nodes %v to Cluster %s failed, %s",
			ua.cloud.CloudProvider, ua.req.Nodes, ua.req.ClusterID, err.Error(),
		)
		return err
	}
	task, err := nodeGroupMgr.AddExternalNodeToCluster(ua.nodeGroup, ua.nodes, &cloudprovider.AddExternalNodesOption{
		CommonOption: *cmOption,
		Operator:     ua.req.Operator,
		Cloud:        groupCloud,
		Cluster:      ua.cluster,
	})
	if err != nil {
		blog.Errorf("cloudprovider %s addExternalNodes %v to Cluster %s failed, %s",
			groupCloud.CloudProvider, ua.req.Nodes, ua.req.ClusterID, err.Error(),
		)
		return err
	}

	// create task
	if err = ua.model.CreateTask(ua.ctx, task); err != nil {
		blog.Errorf("save addNodesToCluster cluster task for cluster %s failed, %s",
			ua.cluster.ClusterID, err.Error(),
		)
		return err
	}

	// dispatch task
	if err = taskserver.GetTaskServer().Dispatch(task); err != nil {
		blog.Errorf("dispatch addNodesToCLuster cluster task for cluster %s failed, %s",
			ua.cluster.ClusterName, err.Error(),
		)
		return err
	}

	ua.task = task
	blog.Infof("add nodes %v to cluster %s with cloudprovider %s processing, task info: %v",
		ua.req.Nodes, ua.req.ClusterID, groupCloud.CloudProvider, task,
	)
	ua.resp.Data = task

	return ua.saveNodesToStorage(common.StatusInitialization)
}

// addNodesToCluster handle normal nodes
func (ua *AddNodesAction) addNodesToCluster() error {
	// get cloudprovider cluster implementation
	clusterMgr, err := cloudprovider.GetClusterMgr(ua.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s ClusterManager for add nodes %v to Cluster %s failed, %s",
			ua.cloud.CloudProvider, ua.req.Nodes, ua.req.ClusterID, err.Error(),
		)
		return err
	}

	// check cloud CIDR && autoScale cluster cidr
	available, err := clusterMgr.CheckClusterCidrAvailable(ua.cluster, &cloudprovider.CheckClusterCIDROption{
		CommonOption:    *ua.option,
		IncomingNodeCnt: uint64(len(ua.nodes)),
		ExternalNode:    ua.req.IsExternalNode,
	})
	if !available {
		blog.Infof("AddNodesAction addNodesToCluster failed: %v", err)
		return err
	}

	// add externalNodes to cluster
	if ua.req.IsExternalNode {
		return ua.addExternalNodesToCluster()
	}

	// default reinstall system when add node to cluster
	task, err := clusterMgr.AddNodesToCluster(ua.cluster, ua.nodes, &cloudprovider.AddNodesOption{
		CommonOption: *ua.option,
		Login: func() *cmproto.NodeLoginInfo {
			loginInfo := &cmproto.NodeLoginInfo{
				InitLoginUsername: ua.req.Login.GetInitLoginUsername(),
				InitLoginPassword: "",
				KeyPair:           &cmproto.KeyInfo{},
			}
			if len(ua.req.Login.GetInitLoginPassword()) > 0 {
				loginInfo.InitLoginPassword, _ = encrypt.Encrypt(nil, ua.req.Login.GetInitLoginPassword())
			}
			if len(ua.req.Login.GetKeyPair().GetKeySecret()) > 0 {
				loginInfo.KeyPair.KeySecret, _ = encrypt.Encrypt(nil, ua.req.Login.GetKeyPair().GetKeySecret())
			}
			if len(ua.req.Login.GetKeyPair().GetKeyPublic()) > 0 {
				loginInfo.KeyPair.KeyPublic, _ = encrypt.Encrypt(nil, ua.req.Login.GetKeyPair().GetKeyPublic())
			}
			return loginInfo
		}(),
		Cloud:        ua.cloud,
		NodeTemplate: ua.nodeTemplate,
		NodeGroupID:  ua.req.NodeGroupID,
		Operator:     ua.req.Operator,
		NodeSchedule: ua.nodeScheduler,
	})
	if err != nil {
		blog.Errorf("cloudprovider %s addNodes %v to Cluster %s failed, %s",
			ua.cloud.CloudProvider, ua.req.Nodes, ua.req.ClusterID, err.Error(),
		)
		return err
	}

	// create task
	if err = ua.model.CreateTask(ua.ctx, task); err != nil {
		blog.Errorf("save addNodesToCluster cluster task for cluster %s failed, %s",
			ua.cluster.ClusterID, err.Error(),
		)
		return err
	}

	// dispatch task
	if err = taskserver.GetTaskServer().Dispatch(task); err != nil {
		blog.Errorf("dispatch addNodesToCLuster cluster task for cluster %s failed, %s",
			ua.cluster.ClusterName, err.Error(),
		)
		return err
	}

	utils.HandleTaskStepData(ua.ctx, task)
	blog.Infof("add nodes %v to cluster %s with cloudprovider %s processing, task info: %v",
		ua.req.Nodes, ua.req.ClusterID, ua.cloud.CloudProvider, task,
	)
	ua.task = task
	ua.resp.Data = task
	return ua.saveNodesToStorage(common.StatusInitialization)
}

// saveNodesToStorage save nodes to db
func (ua *AddNodesAction) saveNodesToStorage(status string) error {
	for _, node := range ua.nodes {
		node.ClusterID = ua.req.ClusterID
		node.NodeGroupID = ua.req.NodeGroupID
		node.Status = status
		node.NodeTemplateID = ua.req.NodeTemplateID

		oldNode, err := ua.model.GetNodeByIP(ua.ctx, node.InnerIP)
		if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
			return err
		}
		if oldNode == nil {
			err = ua.model.CreateNode(ua.ctx, node)
			if err != nil {
				blog.Errorf("save node %s under cluster %s to storage failed, %s",
					node.InnerIP, ua.req.ClusterID, err.Error(),
				)
				return err
			}
			blog.Infof("save node %s under cluster %s to storage successfully", node.InnerIP, ua.req.ClusterID)
			continue
		}

		if err = ua.model.UpdateNode(ua.ctx, node); err != nil {
			blog.Errorf("save node %s under cluster %s to storage failed, %s",
				node.InnerIP, ua.req.ClusterID, err.Error(),
			)
			return err
		}
		blog.Infof("update node %s under cluster %s to storage successfully", node.InnerIP, ua.req.ClusterID)
	}
	return nil
}

// checkManagedClusterNodeNum check managed cluster nodes num
func (ua *AddNodesAction) checkManagedClusterNodeNum() error {
	/*
		if ua.cluster.ManageType != common.ClusterManageTypeManaged {
			return nil
		}
	*/

	nodeStatus := []string{common.StatusRunning, common.StatusInitialization}
	nodes, err := GetClusterStatusNodes(ua.model, ua.cluster, nodeStatus)
	if err != nil {
		blog.Errorf("checkManagedClusterNodeNum[%s] GetClusterStatusNodes failed: %v", ua.cluster.ClusterID, err)
		return err
	}

	blog.Infof("checkManagedClusterNodeNum[%s] GetClusterStatusNodes[%v]", ua.cluster.ClusterID, len(nodes))
	if len(nodes) > 0 {
		return nil
	}
	ua.nodeScheduler = true

	return nil
}

// checkNodeInCluster check node id in cluster
func (ua *AddNodesAction) checkNodeInCluster() error {
	// get all masterIPs
	masterIPs := GetAllMasterIPs(ua.model)

	// check if nodes are already in cluster
	nodeStatus := []string{common.StatusRunning, common.StatusInitialization,
		common.StatusDeleting, common.StatusAddNodesFailed}
	clusterCond := operator.NewLeafCondition(operator.Eq, operator.M{"clusterid": ua.cluster.ClusterID})
	statusCond := operator.NewLeafCondition(operator.In, operator.M{"status": nodeStatus})
	cond := operator.NewBranchCondition(operator.And, clusterCond, statusCond)

	nodes, err := ua.model.ListNode(ua.ctx, cond, &options.ListOption{})
	if err != nil {
		blog.Errorf("get Cluster %s all Nodes failed when AddNodesToCluster, %s", ua.req.ClusterID, err.Error())
		return err
	}
	newNodeIP := make(map[string]string)
	for _, ip := range ua.req.Nodes {
		if cls, ok := masterIPs[ip]; ok {
			blog.Errorf("add nodes %v to Cluster %s failed, Node %s is duplicated",
				ua.req.Nodes, ua.req.ClusterID, ip)
			return fmt.Errorf("node %s is already in Cluster[%s]", ip, cls.ClusterID)
		}
		newNodeIP[ip] = ip
	}
	for _, node := range nodes {
		if _, ok := newNodeIP[node.InnerIP]; ok {
			blog.Errorf("add nodes %v to Cluster %s failed, Node %s is duplicated",
				ua.req.Nodes, ua.req.ClusterID, node.InnerIP,
			)
			return fmt.Errorf("node %s is already in Cluster", node.InnerIP)
		}
	}

	ua.currentNodeCnt = uint64(len(nodes))
	blog.Infof("cluster[%s] AddNodesAction currentNodeCnt %v", ua.cluster.ClusterID, ua.currentNodeCnt)
	return nil
}

// getCloudProjectInfo get cluster/cloud/project info
func (ua *AddNodesAction) getClusterBasicInfo() error {
	cluster, err := ua.model.GetCluster(ua.ctx, ua.req.ClusterID)
	if err != nil {
		blog.Errorf("get Cluster %s failed when AddNodesToCluster, %s", ua.req.ClusterID, err.Error())
		return err
	}
	ua.cluster = cluster

	cloud, err := actions.GetCloudByCloudID(ua.model, ua.cluster.Provider)
	if err != nil {
		blog.Errorf("get Cluster %s provider %s and Project %s failed, %s",
			ua.cluster.ClusterID, ua.cluster.Provider, ua.cluster.ProjectID, err.Error(),
		)
		return err
	}
	ua.cloud = cloud

	if len(ua.req.NodeTemplateID) > 0 {
		template, errGet := actions.GetNodeTemplateByTemplateID(ua.model, ua.req.NodeTemplateID)
		if errGet != nil {
			blog.Errorf("get Cluster %s getNodeTemplateByTemplateID %s failed, %s",
				ua.cluster.ClusterID, ua.req.NodeTemplateID, errGet.Error(),
			)
			return errGet
		}
		ua.nodeTemplate = template
	}
	if len(ua.req.NodeGroupID) > 0 {
		group, errGet := actions.GetNodeGroupByGroupID(ua.model, ua.req.NodeGroupID)
		if errGet != nil {
			blog.Errorf("get Cluster %s GetNodeGroupByGroupID %s failed, %s",
				ua.cluster.ClusterID, ua.req.NodeGroupID, errGet.Error(),
			)
			return errGet
		}
		ua.nodeGroup = group
	}

	return nil
}

// transCloudNodeToDNodes by req nodeIPs trans to cloud node
func (ua *AddNodesAction) transCloudNodeToDNodes() error {
	nodeMgr, err := cloudprovider.GetNodeMgr(ua.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s NodeManager for add nodes %v to Cluster %s failed, %s",
			ua.cloud.CloudProvider, ua.req.Nodes, ua.req.ClusterID, err.Error(),
		)
		return err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ua.cloud,
		AccountID: ua.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s when add nodes %s to cluster %s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, ua.req.Nodes, ua.req.ClusterID, err.Error(),
		)
		return err
	}
	cmOption.Region = ua.cluster.Region

	ua.option = cmOption
	ua.option.Region = ua.cluster.Region

	nodeList, err := ua.getClusterNodesByIPs(nodeMgr, cmOption)
	if err != nil {
		blog.Errorf("AddNodesAction transCloudNodeToDNodes getClusterNodesByIPs failed: %v", err)
		return err
	}
	ua.nodes = nodeList

	blog.Infof("add nodes %v to Cluster %s validate successfully", ua.req.Nodes, ua.req.ClusterID)
	return nil
}

// getClusterNodesByIPs cluster nodes by ips
func (ua *AddNodesAction) getClusterNodesByIPs(nodeMgr cloudprovider.NodeManager,
	cmOption *cloudprovider.CommonOption) ([]*cmproto.Node, error) {
	var (
		nodeList []*cmproto.Node
		err      error
	)

	// cluster check instance if exist, validate nodes existence
	if ua.req.IsExternalNode {
		// get cloud support external nodes
		nodeList, err = nodeMgr.ListExternalNodesByIP(ua.req.Nodes, &cloudprovider.ListNodesOption{
			Common: cmOption,
		})
	} else {
		// get cloud support CVM nodes
		nodeList, err = nodeMgr.ListNodesByIP(ua.req.Nodes, &cloudprovider.ListNodesOption{
			Common: cmOption,
		})
	}

	if err != nil {
		blog.Errorf("validate nodes %s existence failed, %s", ua.req.Nodes, err.Error())
		return nil, err
	}
	if len(nodeList) == 0 {
		blog.Errorf("add nodes %v to Cluster %s validate failed, all Nodes are not under control",
			ua.req.Nodes, ua.req.ClusterID,
		)
		return nil, fmt.Errorf("all nodes don't controlled by cloudprovider %s", ua.cloud.CloudProvider)
	}

	return nodeList, nil
}

// cloudCheckValidate cloud check
func (ua *AddNodesAction) cloudCheckValidate() error {
	validate, err := cloudprovider.GetCloudValidateMgr(ua.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("AddNodesAction cloudCheckValidate failed: %v", err)
		return err
	}
	err = validate.AddNodesToClusterValidate(ua.req, nil)
	if err != nil {
		blog.Errorf("AddNodesAction cloudCheckValidate failed: %v", err)
		return err
	}

	return nil
}

// validate check
func (ua *AddNodesAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	// get cluster basic info(project/cluster/cloud)
	err := ua.getClusterBasicInfo()
	if err != nil {
		return err
	}

	// cloud validate
	err = ua.cloudCheckValidate()
	if err != nil {
		return err
	}

	// check operator host permission
	canUse := CheckUseNodesPermForUser(ua.cluster.BusinessID, ua.req.Operator, ua.req.Nodes)
	if !canUse {
		errMsg := fmt.Errorf("add nodes failed: user[%s] no perm to use nodes[%v] in bizID[%s]",
			ua.req.Operator, ua.req.Nodes, ua.cluster.BusinessID)
		blog.Errorf(errMsg.Error())
		return errMsg
	}

	// cluster add nodes limit at a time
	limit := ua.cloud.ConfInfo.MaxWorkerNodeNum
	if limit == 0 {
		limit = common.ClusterAddNodesLimit
	}

	if len(ua.req.Nodes) > int(limit) {
		errMsg := fmt.Errorf("add nodes failed: cluster[%s] add NodesLimit exceed %d",
			ua.cluster.ClusterID, common.ClusterAddNodesLimit)
		blog.Errorf(errMsg.Error())
		return errMsg
	}

	// check managed_type nodes
	if err = ua.checkManagedClusterNodeNum(); err != nil {
		return err
	}

	// addNodes exist in cluster
	if err = ua.checkNodeInCluster(); err != nil {
		return err
	}

	// request nodes trans to cloudNode
	err = ua.transCloudNodeToDNodes()
	if err != nil {
		return err
	}

	return nil
}

func (ua *AddNodesAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = code == common.BcsErrClusterManagerSuccess
}

// Handle handles update cluster request
func (ua *AddNodesAction) Handle(ctx context.Context, req *cmproto.AddNodesRequest, resp *cmproto.AddNodesResponse) {
	if req == nil || resp == nil {
		blog.Errorf("add cluster nodes failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	// check request body validate
	if err := ua.req.Validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// only save nodes data to db
	if req.OnlyCreateInfo {
		err := ua.getClusterBasicInfo()
		if err != nil {
			ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
			return
		}

		err = ua.transCloudNodeToDNodes()
		if err != nil {
			errMsg := fmt.Sprintf("createCluster[%s] transCloudNodeToDNodes failed: %v", ua.req.ClusterID, err)
			ua.setResp(common.BcsErrClusterManagerDBOperation, errMsg)
			return
		}

		// only create nodes information to local storage
		blog.Infof("only create nodes %v information to local storage under cluster %s", req.Nodes, req.ClusterID)
		if err := ua.saveNodesToStorage(common.StatusRunning); err != nil {
			ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
			return
		}
		blog.Infof("only create nodes %v information to local storage under cluster %s successfully",
			req.Nodes, req.ClusterID)
		ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
		return
	}

	// check node if exist in cloud_provider
	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// generate async task to call cloud provider for add nodes
	// 1. check cluster cidr and auto scale cidr
	// 2. add external nodes by nodeGroupMgr
	// 3. task to add node in cluster
	// 4. init node status initialization
	if err := ua.addNodesToCluster(); err != nil {
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	blog.Infof(
		"add nodes %v information to local storage under cluster %s successfully",
		req.Nodes, req.ClusterID,
	)

	err := ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.Cluster.String(),
		ResourceID:   ua.cluster.ClusterID,
		TaskID:       ua.task.TaskID,
		Message:      fmt.Sprintf("集群%s添加节点", ua.cluster.ClusterID),
		OpUser:       req.Operator,
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    ua.cluster.ClusterID,
		ProjectID:    ua.cluster.ProjectID,
		ResourceName: ua.cluster.ClusterName,
	})
	if err != nil {
		blog.Errorf("AddNodesToCluster[%s] CreateOperationLog failed: %v", ua.cluster.ClusterID, err)
	}
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
