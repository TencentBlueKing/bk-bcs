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

package cluster

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	cmcommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
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
}

// NewUpdateAction create update action
func NewUpdateAction(model store.ClusterManagerModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

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

func (ua *UpdateAction) getCluster() error {
	cluster, err := ua.model.GetCluster(ua.ctx, ua.req.ClusterID)
	if err != nil {
		return err
	}
	ua.cluster = cluster
	return nil
}

func (ua *UpdateAction) updateCluster() error {
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
	if len(ua.req.EngineType) != 0 {
		ua.cluster.EngineType = ua.req.EngineType
	}
	if ua.req.IsExclusive != nil && ua.req.IsExclusive.GetValue() != ua.cluster.IsExclusive {
		ua.cluster.IsExclusive = ua.req.IsExclusive.GetValue()
	}
	if len(ua.req.ClusterType) != 0 {
		ua.cluster.ClusterType = ua.req.ClusterType
	}
	if len(ua.req.FederationClusterID) != 0 {
		ua.cluster.FederationClusterID = ua.req.FederationClusterID
	}
	if ua.req.Labels != nil {
		ua.cluster.Labels = ua.req.Labels
	}
	if len(ua.req.Status) != 0 {
		ua.cluster.Status = ua.req.Status
	}
	if ua.req.BcsAddons != nil {
		ua.cluster.BcsAddons = ua.req.BcsAddons
	}
	if ua.req.ExtraAddons != nil {
		ua.cluster.ExtraAddons = ua.req.ExtraAddons
	}
	if len(ua.req.ManageType) != 0 {
		ua.cluster.ManageType = ua.req.ManageType
	}
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
	if len(ua.req.SystemID) != 0 {
		ua.cluster.SystemID = ua.req.SystemID
	}

	if len(ua.req.NetworkType) != 0 {
		ua.cluster.NetworkType = ua.req.NetworkType
	}

	if len(ua.req.ModuleID) != 0 {
		ua.cluster.ModuleID = ua.req.ModuleID
	}
	if len(ua.req.Description) > 0 {
		ua.cluster.Description = ua.req.Description
	}
	if len(ua.req.ClusterCategory) > 0 {
		ua.cluster.ClusterCategory = ua.req.ClusterCategory
	}
	if len(ua.req.ExtraClusterID) > 0 {
		ua.cluster.ExtraClusterID = ua.req.ExtraClusterID
	}
	if ua.req.IsCommonCluster != nil {
		ua.cluster.IsCommonCluster = ua.req.IsCommonCluster.GetValue()
	}
	if ua.req.IsShared != nil {
		ua.cluster.IsShared = ua.req.IsShared.GetValue()
	}
	if len(ua.req.CreateTime) > 0 {
		ua.cluster.CreateTime = ua.req.CreateTime
	}
	if len(ua.req.Creator) > 0 {
		ua.cluster.Creator = ua.req.Creator
	}
	if len(ua.req.ImportCategory) > 0 {
		ua.cluster.ImportCategory = ua.req.ImportCategory
	}

	for _, ip := range ua.req.Master {
		if ua.cluster.Master == nil {
			ua.cluster.Master = make(map[string]*cmproto.Node)
		}
		// add more details for Master Node
		ua.cluster.Master[ip] = &cmproto.Node{
			InnerIP: ip,
			Status:  common.StatusRunning,
		}
	}
	ua.cluster.UpdateTime = time.Now().Format(time.RFC3339)

	// update DB clusterInfo & passcc cluster
	err := ua.model.UpdateCluster(ua.ctx, ua.cluster)
	if err != nil {
		return err
	}
	updatePassCCClusterInfo(ua.cluster)
	return nil
}

func (ua *UpdateAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
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
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
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

func (ua *UpdateNodeAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	if len(ua.req.ClusterID) == 0 && len(ua.req.NodeGroupID) == 0 && len(ua.req.Status) == 0 {
		return fmt.Errorf("UpdateNodeAction validate failed: body empty")
	}

	return nil
}

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

func (ua *UpdateNodeAction) updateNodes() error {
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
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	if ua.resp.Data == nil {
		ua.resp.Data = &cmproto.NodeStatus{}
	}
	ua.resp.Data.Success = ua.success
	ua.resp.Data.Failed = ua.failed
}

// Handle handles update nodes request
func (ua *UpdateNodeAction) Handle(ctx context.Context, req *cmproto.UpdateNodeRequest, resp *cmproto.UpdateNodeResponse) {
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

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
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
	project        *cmproto.Project
	option         *cloudprovider.CommonOption
	nodeGroup      *cmproto.NodeGroup
	currentNodeCnt uint64
}

// NewAddNodesAction create addNodes action
func NewAddNodesAction(model store.ClusterManagerModel) *AddNodesAction {
	return &AddNodesAction{
		model: model,
	}
}

func (ua *AddNodesAction) addNodesToCluster() error {
	// get cloudprovider cluster implementation
	clusterMgr, err := cloudprovider.GetClusterMgr(ua.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s ClusterManager for add nodes %v to Cluster %s failed, %s",
			ua.cloud.CloudProvider, ua.req.Nodes, ua.req.ClusterID, err.Error(),
		)
		return err
	}
	reinstall := len(ua.req.InitLoginPassword) != 0

	// check cloud CIDR
	available, err := clusterMgr.CheckClusterCidrAvailable(ua.cluster, &cloudprovider.CheckClusterCIDROption{
		CurrentNodeCnt:  ua.currentNodeCnt,
		IncomingNodeCnt: uint64(len(ua.nodes)),
	})
	if !available {
		blog.Infof("AddNodesAction addNodesToCluster failed: %v", err)
		return err
	}

	// default reinstall system when add node to cluster
	task, err := clusterMgr.AddNodesToCluster(ua.cluster, ua.nodes, &cloudprovider.AddNodesOption{
		CommonOption: *ua.option,
		// input passwd not empty
		Reinstall:    reinstall,
		InitPassword: ua.req.InitLoginPassword,
		Cloud:        ua.cloud,
		NodeGroupID:  ua.req.NodeGroupID,
		Operator:     ua.req.Operator,
	})
	if err != nil {
		blog.Errorf("cloudprovider %s addNodes %v to Cluster %s failed, %s",
			ua.cloud.CloudProvider, ua.req.Nodes, ua.req.ClusterID, err.Error(),
		)
		return err
	}

	// create task
	if err := ua.model.CreateTask(ua.ctx, task); err != nil {
		blog.Errorf("save addNodesToCluster cluster task for cluster %s failed, %s",
			ua.cluster.ClusterID, err.Error(),
		)
		return err
	}

	// dispatch task
	if err := taskserver.GetTaskServer().Dispatch(task); err != nil {
		blog.Errorf("dispatch addNodesToCLuster cluster task for cluster %s failed, %s",
			ua.cluster.ClusterName, err.Error(),
		)
		return err
	}

	ua.task = task
	blog.Infof("add nodes %v to cluster %s with cloudprovider %s processing, task info: %v",
		ua.req.Nodes, ua.req.ClusterID, ua.cloud.CloudProvider, task,
	)
	ua.resp.Data = task
	return ua.saveNodesToStorage(common.StatusInitialization)
}

func (ua *AddNodesAction) saveNodesToStorage(status string) error {
	for _, node := range ua.nodes {
		node.ClusterID = ua.req.ClusterID
		node.NodeGroupID = ua.req.NodeGroupID
		node.Status = status

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

		if err := ua.model.UpdateNode(ua.ctx, node); err != nil {
			blog.Errorf("save node %s under cluster %s to storage failed, %s",
				node.InnerIP, ua.req.ClusterID, err.Error(),
			)
			return err
		}
		blog.Infof("update node %s under cluster %s to storage successfully", node.InnerIP, ua.req.ClusterID)
	}
	return nil
}

func (ua *AddNodesAction) checkNodeInCluster() error {
	//check if nodes are already in cluster
	nodeStatus := []string{common.StatusRunning, common.StatusInitialization, common.StatusDeleting}
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

	cloud, project, err := actions.GetProjectAndCloud(ua.model, ua.cluster.ProjectID, ua.cluster.Provider)
	if err != nil {
		blog.Errorf("get Cluster %s provider %s and Project %s failed, %s",
			ua.cluster.ClusterID, ua.cluster.Provider, ua.cluster.ProjectID, err.Error(),
		)
		return err
	}
	ua.cloud = cloud
	ua.project = project

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
	cmOption, err := cloudprovider.GetCredential(ua.project, ua.cloud)
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s when add nodes %s to cluster %s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, ua.req.Nodes, ua.req.ClusterID, err.Error(),
		)
		return err
	}
	cmOption.Region = ua.cluster.Region

	ua.option = cmOption
	ua.option.Region = ua.cluster.Region

	// cluster check instance if exist, validate nodes existence
	nodeList, err := nodeMgr.ListNodesByIP(ua.req.Nodes, &cloudprovider.ListNodesOption{
		Common:       cmOption,
		ClusterVPCID: ua.cluster.VpcID,
	})
	if err != nil {
		blog.Errorf("validate nodes %s existence failed, %s", ua.req.Nodes, err.Error())
		return err
	}
	if len(nodeList) == 0 {
		blog.Errorf("add nodes %v to Cluster %s validate failed, all Nodes are not under control",
			ua.req.Nodes, ua.req.ClusterID,
		)
		return fmt.Errorf("all nodes don't controlled by cloudprovider %s", ua.cloud.CloudProvider)
	}
	ua.nodes = nodeList

	blog.Infof("add nodes %v to Cluster %s validate successfully", ua.req.Nodes, ua.req.ClusterID)
	return nil
}

func (ua *AddNodesAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	// check instance init login passwd
	// generate passwd if passwd empty
	pwlen := len(ua.req.InitLoginPassword)
	if pwlen != 0 && (pwlen < 8 || pwlen > 16) {
		return fmt.Errorf("when setting initLoginPassword, its length must be in [8, 16]")
	}

	// get cluster basic info(project/cluster/cloud)
	err := ua.getClusterBasicInfo()
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
	if len(ua.req.Nodes) > common.ClusterAddNodesLimit {
		errMsg := fmt.Errorf("add nodes failed: cluster[%s] add NodesLimit exceed %d",
			ua.cluster.ClusterID, common.ClusterAddNodesLimit)
		blog.Errorf(errMsg.Error())
		return errMsg
	}

	// addNodes exist in cluster
	if err = ua.checkNodeInCluster(); err != nil {
		return err
	}

	// check nodegroup information
	// if need to add node to nodegroup, must cluster provider equal nodeGroup provider
	// if len(nodeGroupID) == 0, add node to cluster but not belong nodeGroup
	if len(ua.req.NodeGroupID) != 0 {
		blog.Infof("add Nodes %v to NodeGroup %s when AddNodesToCluster %s",
			ua.req.Nodes, ua.req.NodeGroupID, ua.req.ClusterID,
		)
		//try to get nodegroup
		nodeGroup, err := ua.model.GetNodeGroup(ua.ctx, ua.req.NodeGroupID)
		if err != nil {
			blog.Errorf("get NodeGroup %s failed when AddNodesToCluster, %s", ua.req.NodeGroupID, err.Error())
			return err
		}
		ua.nodeGroup = nodeGroup

		// cloud provider should equal nodeGroup provider, provider is cloudID
		if ua.cluster.Provider != nodeGroup.Provider {
			blog.Errorf("add nodes %v to Cluster %s failed, Cluster and NodeGroup provider [%s/%s] must be same",
				ua.req.Nodes, ua.req.ClusterID, ua.cluster.Provider, nodeGroup.Provider,
			)
			return fmt.Errorf("nodegroup and cluseter cloudprovider must be same")
		}
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
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handles update cluster request
func (ua *AddNodesAction) Handle(ctx context.Context, req *cmproto.AddNodesRequest, resp *cmproto.AddNodesResponse) {
	if req == nil || resp == nil {
		blog.Errorf("update cluster failed, req or resp is empty")
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
		blog.Infof("only create nodes %v information to local storage under cluster %s successfully", req.Nodes, req.ClusterID)
		ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
		return
	}

	// check node if exist in cloud_provider
	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// generate async task to call cloud provider for add nodes
	// 1. task to add node in cluster 2. init node status initialization
	if err := ua.addNodesToCluster(); err != nil {
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	blog.Infof(
		"add nodes %v inforamtion to local storage under cluster %s successfully",
		req.Nodes, req.ClusterID,
	)

	err := ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.Cluster.String(),
		ResourceID:   ua.cluster.ClusterID,
		TaskID:       ua.task.TaskID,
		Message:      fmt.Sprintf("集群%s添加节点", ua.cluster.ClusterID),
		OpUser:       req.Operator,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("AddNodesToCluster[%s] CreateOperationLog failed: %v", ua.cluster.ClusterID, err)
	}
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
