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
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
)

// DeleteAction action for delete cluster
type DeleteAction struct {
	ctx        context.Context
	req        *cmproto.DeleteClusterReq
	resp       *cmproto.DeleteClusterResp
	model      store.ClusterManagerModel
	cluster    *cmproto.Cluster
	nodes      []*cmproto.Node
	quotaList  []cmproto.ResourceQuota
	nodeGroups []cmproto.NodeGroup

	// cluster associate ca options
	scalingOption *cmproto.ClusterAutoScalingOption
	tasks         *cmproto.Task
	project       *cmproto.Project
	cloud         *cmproto.Cloud
	cmOptions     *cloudprovider.CommonOption
}

// NewDeleteAction delete cluster action
func NewDeleteAction(model store.ClusterManagerModel) *DeleteAction {
	return &DeleteAction{
		model: model,
	}
}

func (da *DeleteAction) queryQuotas() error {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"clusterid": da.req.ClusterID,
	})
	quotaList, err := da.model.ListQuota(da.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return err
	}
	da.quotaList = quotaList
	return nil
}

func (da *DeleteAction) queryNodeGroup() error {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"clusterid": da.req.ClusterID,
	})
	groupList, err := da.model.ListNodeGroup(da.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return err
	}
	da.nodeGroups = groupList
	return nil
}

func (da *DeleteAction) queryAutoScalingOption() error {
	option, err := da.model.GetAutoScalingOption(da.ctx, da.req.ClusterID)
	if err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			return nil
		}
		return err
	}
	da.scalingOption = option
	return nil
}

func (da *DeleteAction) getClusterAndNode() error {
	cluster, err := da.model.GetCluster(da.ctx, da.req.ClusterID)
	if err != nil {
		blog.Errorf("get Cluster %s failed, %s", da.req.ClusterID, err.Error())
		return err
	}
	da.cluster = cluster

	// get relative nodes by clusterID
	condM := make(operator.M)
	condM["clusterid"] = cluster.ClusterID
	cond := operator.NewLeafCondition(operator.Eq, condM)

	nodes, err := da.model.ListNode(da.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		blog.Errorf("get Cluster %s Nodes failed, %s", da.req.ClusterID, err.Error())
		return err
	}
	da.nodes = nodes

	return nil
}

func (da *DeleteAction) checkRelativeResource() error {
	if err := da.queryQuotas(); err != nil {
		return fmt.Errorf("query quotas in delete cluster failed, err %s", err.Error())
	}
	if err := da.queryNodeGroup(); err != nil {
		return fmt.Errorf("query NodeGroup in delete cluster failed, err %s", err.Error())
	}
	if err := da.queryAutoScalingOption(); err != nil {
		return fmt.Errorf("get AutoScalingOption in delete cluster failed, err %s", err.Error())
	}
	return nil
}

func (da *DeleteAction) canDelete() error {
	if len(da.nodeGroups) != 0 {
		return fmt.Errorf("cannot delete cluster, there is relative NodeGroup running")
	}
	if len(da.nodes) != 0 {
		return fmt.Errorf("cannot delete cluster, there are Nodes in cluster")
	}
	if len(da.quotaList) != 0 {
		return fmt.Errorf("cannot delete cluster, there is quots in cluster")
	}
	if da.scalingOption != nil {
		return fmt.Errorf("cannot delete cluster, there is relative AutoScalingOption")
	}
	return nil
}

func (da *DeleteAction) cleanLocalInformation() error {
	// importer cluster only delete cluster related data
	if da.isImporterCluster() {
		da.req.IsForced = true
	}
	if da.req.IsForced {
		// clean cluster autoscaling option
		if da.scalingOption != nil {
			if err := da.model.DeleteAutoScalingOption(da.ctx, da.scalingOption.ClusterID); err != nil {
				blog.Errorf("clean Cluster AutoScalingOption %s storage information failed, %s",
					da.req.ClusterID, err.Error())
				return err
			}
		}
		for _, group := range da.nodeGroups {
			if err := da.model.DeleteNodeGroup(da.ctx, group.NodeGroupID); err != nil {
				blog.Errorf("clean Cluster %s NodeGroup %s storage information failed, %s",
					da.req.ClusterID, group.NodeGroupID, err.Error())
				return err
			}
		}
		if len(da.nodes) > 0 {
			nodeIPs := make([]string, 0)
			for _, node := range da.nodes {
				nodeIPs = append(nodeIPs, node.InnerIP)
			}

			err := da.model.DeleteNodesByIPs(da.ctx, nodeIPs)
			if err != nil {
				blog.Errorf("clean Cluster %s node %v storage information failed, %s",
					da.req.ClusterID, nodeIPs, err.Error())
				return err
			}
		}
		// delete all namespace quotas related to certain cluster
		if len(da.quotaList) != 0 {
			if err := da.model.BatchDeleteQuotaByCluster(da.ctx, da.req.ClusterID); err != nil {
				blog.Errorf("clean Cluster %s Quota information failed, %s",
					da.req.ClusterID, err.Error())
				return err
			}
		}
	} else {
		// can not force delete, we need to check all relative resource
		// we don't delete cluster when relative resource are running
		if err := da.canDelete(); err != nil {
			return err
		}
	}
	// release cidr
	if err := da.releaseClusterCIDR(da.cluster); err != nil {
		return err
	}

	// finally clean cluster
	da.cluster.Status = common.StatusDeleted
	if err := da.model.UpdateCluster(da.ctx, da.cluster); err != nil {
		return err
	}
	return nil
}

func (da *DeleteAction) deleteRelativeResource() error {
	//clean cluster autoscaling option, No relative entity in cloud infrastructure
	if da.scalingOption != nil {
		if err := da.model.DeleteAutoScalingOption(da.ctx, da.scalingOption.ClusterID); err != nil {
			blog.Errorf("clean Cluster AutoScalingOption %s storage information failed, %s",
				da.req.ClusterID, err.Error())
			return err
		}
	}
	// delete all namespace quotas related to certain cluster
	// no relative entity in cloud infrastructure
	if len(da.quotaList) != 0 {
		if err := da.model.BatchDeleteQuotaByCluster(da.ctx, da.req.ClusterID); err != nil {
			blog.Errorf("clean Cluster %s Quota information failed, %s",
				da.req.ClusterID, err.Error())
			return err
		}
	}
	//! NodeGroup update status for other operation deny, it manage Nodes,
	//! we don't delete it here, we delete it after all Nodes are releasing
	for _, group := range da.nodeGroups {
		group.Status = common.StatusDeleting
		if err := da.model.UpdateNodeGroup(da.ctx, &group); err != nil {
			blog.Errorf("setting Cluster %s relative NodeGroup %s to status DELETEING failed, %s",
				da.req.ClusterID, group.NodeGroupID, err.Error())
			return err
		}
	}

	// delete all nodes related to certain cluster
	if len(da.nodes) > 0 {
		nodeIPs := make([]string, 0)
		for i := range da.nodes {
			nodeIPs = append(nodeIPs, da.nodes[i].InnerIP)
		}
		if err := da.model.DeleteNodesByIPs(da.ctx, nodeIPs); err != nil {
			blog.Errorf("delete Cluster %s relative Nodes %v failed, %s",
				da.req.ClusterID, nodeIPs, err.Error())
			return err
		}
	}

	return nil
}

func (da *DeleteAction) setResp(code uint32, msg string) {
	da.resp.Code = code
	da.resp.Message = msg
	da.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (da *DeleteAction) releaseClusterCIDR(cls *cmproto.Cluster) error {
	if len(cls.NetworkSettings.GetClusterIPv4CIDR()) > 0 {
		cidr, err := da.model.GetTkeCidr(da.ctx, cls.VpcID, cls.NetworkSettings.ClusterIPv4CIDR)
		if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Errorf("delete cluster release cidr[%s] failed: %v", cls.NetworkSettings.ClusterIPv4CIDR, err)
			return err
		}

		if cidr == nil {
			blog.Infof("delete cluster release cidr[%s] not found", cls.NetworkSettings.ClusterIPv4CIDR)
			return nil
		}

		if cidr.Cluster == cls.ClusterID && cidr.Status == common.TkeCidrStatusUsed {
			// update cidr and save to DB
			updateCidr := cidr
			updateCidr.Status = common.TkeCidrStatusAvailable
			updateCidr.Cluster = ""
			updateCidr.UpdateTime = time.Now().String()
			err = da.model.UpdateTkeCidr(da.ctx, updateCidr)
			if err != nil {
				blog.Errorf("delete cluster release cidr[%s] failed: %v", cls.NetworkSettings.ClusterIPv4CIDR, err)
				return err
			}
		}
	}

	return nil
}

func (da *DeleteAction) validate(req *cmproto.DeleteClusterReq) error {
	if err := req.Validate(); err != nil {
		return err
	}

	if len(req.Operator) == 0 {
		return fmt.Errorf("operator empty when delete cluster")
	}

	if len(req.InstanceDeleteMode) == 0 {
		req.InstanceDeleteMode = cloudprovider.Retain.String()
	}

	if req.InstanceDeleteMode != cloudprovider.Terminate.String() && req.InstanceDeleteMode != cloudprovider.Retain.String() {
		return fmt.Errorf("deleteInstanceMode is terminate or retain when delete cluster")
	}

	return nil
}

func (da *DeleteAction) getCloudAndProjectInfo(ctx context.Context) error {
	cloud, err := da.model.GetCloud(ctx, da.cluster.Provider)
	if err != nil {
		blog.Errorf("get Cluster %s provider %s information failed, %s", da.cluster.ClusterID, da.cluster.Provider, err.Error())
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	da.cloud = cloud

	pro, err := da.model.GetProject(ctx, da.cluster.ProjectID)
	if err != nil {
		blog.Errorf("get Cluster %s Project %s information failed, %s", da.cluster.ClusterID, da.cluster.ProjectID, err.Error())
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	da.project = pro

	coption, err := cloudprovider.GetCredential(da.project, da.cloud)
	if err != nil {
		blog.Errorf("get Credential failed when delete Cluster %s, %s. Project %s, Cloud %s",
			da.cluster.ClusterID, da.project.ProjectID, da.cloud.CloudID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return err
	}
	coption.Region = da.cluster.Region
	da.cmOptions = coption

	return nil
}

// Handle delete cluster request
func (da *DeleteAction) Handle(ctx context.Context, req *cmproto.DeleteClusterReq, resp *cmproto.DeleteClusterResp) {
	if req == nil || resp == nil {
		blog.Errorf("delete cluster failed, req or resp is empty")
		return
	}
	da.ctx = ctx
	da.req = req
	da.resp = resp

	// delete parameter validate check
	if err := da.validate(req); err != nil {
		da.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// admin operation delete cluster database record when set DeleteClusterRecord true
	if da.req.DeleteClusterRecord {
		err := da.model.DeleteCluster(da.ctx, da.req.ClusterID)
		if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Errorf("DeleteClusterRecord %s err: %s", req.ClusterID, err.Error())
			da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
			return
		}
		da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
		return
	}

	// get cluster and nodes info
	if err := da.getClusterAndNode(); err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			da.setResp(common.BcsErrClusterManagerDatabaseRecordNotFound, err.Error())
			return
		}
	}

	// cluster is deleting && IsForced == false, return
	if da.cluster.Status == common.StatusDeleting || da.cluster.Status == common.StatusDeleted {
		blog.Warnf("Cluster %s is under %s and is not force deleting, simply return", req.ClusterID, da.cluster.Status)
		da.setResp(common.BcsErrClusterManagerTaskErr, "cluster is under deleting/deleted")
		// retrieve specified task then return to user
		return
	}

	// get cluster relative resource (quota / nodeGroup / autoScalingOptions) for checking
	if err := da.checkRelativeResource(); err != nil {
		blog.Errorf("check Cluster %s relative resource err: %s", req.ClusterID, err.Error())
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// version 1 only delete cluster info, manual delete cluster by cloud provider
	//     OnlyDeleteInfo = true && IsForced = true (delete relative resource and delete cluster)
	//     and IsForced = false (check resource, can't delete cluster if resource do not nil).
	// if delete importer cluster need to delete cluster extra data, thus set IsForced = true
	if req.OnlyDeleteInfo || da.isImporterCluster() {
		//clean all relative resource then delete cluster finally
		if err := da.cleanLocalInformation(); err != nil {
			blog.Errorf("only delete Cluster %s local information err, %s", req.ClusterID, err.Error())
			da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
			return
		}

		blog.Infof("only Delete Cluster %s local information successfully", req.ClusterID)
		da.resp.Data = da.cluster
		da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
		return
	}
	blog.Infof("try to delete Cluster %s entity in cloud and local information", req.ClusterID)

	// step1: call cloud provider interface to delete underlay cluster
	// step2: clean local resource information, update cluster status
	err := da.getCloudAndProjectInfo(ctx)
	if err != nil {
		return
	}

	// delete cluster relative resource directly when IsForced is true
	if req.IsForced {
		// delete relative cloud infrastructure entity & local information
		if err = da.deleteRelativeResource(); err != nil {
			blog.Infof("force delete Cluster %s relative resource err, %s", req.ClusterID, err.Error())
			da.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
			return
		}
		blog.Infof("force delete Cluster %s relative resource successfully", req.ClusterID)
	} else {
		// cluster is still keep other resource, users need to clean them separately
		if err = da.canDelete(); err != nil {
			da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
			return
		}
	}
	blog.Infof("all Cluster %s relative resource are handling successfully, try to delete Cluster", req.ClusterID)

	// create delete cluster task
	err = da.createDeleteClusterTask(req)
	if err != nil {
		return
	}
	blog.Infof("delete cluster[%s] task cloud[%s] provider[%s] successfully",
		da.cluster.ClusterName, da.cloud.CloudID, da.cloud.CloudProvider)

	// build operation log
	err = da.model.CreateOperationLog(da.ctx, &cmproto.OperationLog{
		ResourceType: common.Cluster.String(),
		ResourceID:   da.cluster.ClusterID,
		TaskID:       da.tasks.TaskID,
		Message:      fmt.Sprintf("删除%s集群%s", da.cluster.Provider, da.cluster.ClusterID),
		OpUser:       da.req.Operator,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("delete cluster[%s] CreateOperationLog failed: %v", da.cluster.ClusterID, err)
	}

	da.resp.Data = da.cluster
	da.resp.Task = da.tasks
	da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

func (da *DeleteAction) isImporterCluster() bool {
	return da.cluster.ClusterCategory == Importer
}

func (da *DeleteAction) createDeleteClusterTask(req *cmproto.DeleteClusterReq) error {
	clsMgr, err := cloudprovider.GetClusterMgr(da.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get Cluster %s real CloudProvider %s manager failed, %s", req.ClusterID,
			da.cloud.CloudProvider, err.Error())
		da.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return err
	}

	// update cluster status: deleting
	da.cluster.Status = common.StatusDeleting
	if err = da.model.UpdateCluster(da.ctx, da.cluster); err != nil {
		blog.Errorf("update Cluster %s to status DELETING failed, %s", req.ClusterID, err.Error())
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}

	// call cloud provider api to delete cluster by async tasks
	task, err := clsMgr.DeleteCluster(da.cluster, &cloudprovider.DeleteClusterOption{
		CommonOption: *da.cmOptions,
		IsForce:      req.IsForced,
		DeleteMode:   cloudprovider.DeleteMode(req.InstanceDeleteMode),
		Operator:     req.Operator,
		Cloud:        da.cloud,
		Cluster:      da.cluster,
	})
	if err != nil {
		blog.Errorf("delete Cluster %s by cloudprovider %s failed, %s",
			da.cluster.ClusterID, da.cloud.CloudID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerTaskErr, err.Error())
		return err
	}

	// create task and dispatch task
	if err := da.model.CreateTask(da.ctx, task); err != nil {
		blog.Errorf("save delete cluster task for cluster %s failed, %s",
			da.cluster.ClusterName, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	if err := taskserver.GetTaskServer().Dispatch(task); err != nil {
		blog.Errorf("dispatch delete cluster task for cluster %s failed, %s",
			da.cluster.ClusterName, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerTaskErr, err.Error())
		return err
	}

	da.tasks = task
	return nil
}

// DeleteNodesAction action for delete nodes from cluster
type DeleteNodesAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.DeleteNodesRequest
	resp  *cmproto.DeleteNodesResponse

	nodeIPList []string
	cluster    *cmproto.Cluster
	nodes      []*cmproto.Node
	task       *cmproto.Task
	project    *cmproto.Project
	cloud      *cmproto.Cloud
}

// NewDeleteNodesAction delete Nodes action
func NewDeleteNodesAction(model store.ClusterManagerModel) *DeleteNodesAction {
	return &DeleteNodesAction{
		model: model,
	}
}

func (da *DeleteNodesAction) setResp(code uint32, msg string) {
	da.resp.Code = code
	da.resp.Message = msg
	da.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (da *DeleteNodesAction) getNodesByClusterAndIPs() ([]string, error) {
	err := da.getClusterBasicInfo()
	if err != nil {
		return nil, err
	}

	// get relative nodes by clusterID
	clusterCond := operator.NewLeafCondition(operator.Eq, operator.M{"clusterid": da.cluster.ClusterID})
	nodeCond := operator.NewLeafCondition(operator.In, operator.M{"innerip": da.nodeIPList})
	cond := operator.NewBranchCondition(operator.And, clusterCond, nodeCond)

	nodes, err := da.model.ListNode(da.ctx, cond, &storeopt.ListOption{})
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		blog.Errorf("get Cluster %s Nodes failed, %s", da.req.ClusterID, err.Error())
		return nil, err
	}

	// check node if deleting
	for i := range nodes {
		if nodes[i].Status == common.StatusDeleting || nodes[i].Status == common.StatusInitialization {
			return nil, fmt.Errorf("DeleteNodesAction node[%s] status is %s", nodes[i].InnerIP, nodes[i].Status)
		}
	}

	// filter nodeGroup nodes
	filterNodes := make(map[string]string, 0)
	for i := range nodes {
		if nodes[i].NodeGroupID != "" {
			filterNodes[nodes[i].InnerIP] = ""
		}
	}

	deleteNodeIPs := make([]string, 0)
	for _, ip := range da.nodeIPList {
		if _, ok := filterNodes[ip]; !ok {
			deleteNodeIPs = append(deleteNodeIPs, ip)
		}
	}
	err = da.transCloudNodeToDNodes(deleteNodeIPs)
	if err != nil {
		return nil, err
	}

	return deleteNodeIPs, nil
}

// transCloudNodeToDNodes by req nodeIPs trans to cloud node
func (da *DeleteNodesAction) transCloudNodeToDNodes(ips []string) error {
	nodeMgr, err := cloudprovider.GetNodeMgr(da.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s NodeManager for add nodes %v to Cluster %s failed, %s",
			da.cloud.CloudProvider, da.req.Nodes, da.req.ClusterID, err.Error(),
		)
		return err
	}
	cmOption, err := cloudprovider.GetCredential(da.project, da.cloud)
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s when add nodes %s to cluster %s failed, %s",
			da.cloud.CloudID, da.cloud.CloudProvider, da.req.Nodes, da.req.ClusterID, err.Error(),
		)
		return err
	}
	cmOption.Region = da.cluster.Region

	// cluster check instance if exist, validate nodes existence
	nodeList, err := nodeMgr.ListNodesByIP(ips, &cloudprovider.ListNodesOption{
		Common:       cmOption,
		ClusterVPCID: da.cluster.VpcID,
	})
	if err != nil {
		blog.Errorf("validate nodes %s existence failed, %s", da.req.Nodes, err.Error())
		return err
	}
	if len(nodeList) == 0 {
		blog.Errorf("add nodes %v to Cluster %s validate failed, all Nodes are not under control",
			da.req.Nodes, da.req.ClusterID,
		)
		return fmt.Errorf("all nodes don't controlled by cloudprovider %s", da.cloud.CloudProvider)
	}
	da.nodes = nodeList

	blog.Infof("add nodes %v to Cluster %s validate successfully", da.req.Nodes, da.req.ClusterID)
	return nil
}

func (da *DeleteNodesAction) validate() error {
	err := da.req.Validate()
	if err != nil {
		return err
	}

	if da.req.DeleteMode == "" {
		da.req.DeleteMode = cloudprovider.Retain.String()
	} else {
		switch da.req.DeleteMode {
		case cloudprovider.Terminate.String(), cloudprovider.Retain.String():
		default:
			return fmt.Errorf("DeleteNodesAction DeleteMode musr be terminate or retain")
		}
	}

	da.nodeIPList = strings.Split(da.req.Nodes, ",")
	if len(da.nodeIPList) == 0 {
		return fmt.Errorf("DeleteNodesAction parameter lost nodeIPs")
	}

	if da.req.Operator == "" {
		return fmt.Errorf("DeleteNodesAction parameter lost operator")
	}

	return nil
}

func (da *DeleteNodesAction) getClusterBasicInfo() error {
	cluster, err := da.model.GetCluster(da.ctx, da.req.ClusterID)
	if err != nil {
		blog.Errorf("get Cluster %s failed, %s", da.req.ClusterID, err.Error())
		return err
	}
	da.cluster = cluster

	cloud, project, err := actions.GetProjectAndCloud(da.model, da.cluster.ProjectID, da.cluster.Provider)
	if err != nil {
		blog.Errorf("get Cluster %s provider %s and Project %s failed, %s",
			da.cluster.ClusterID, da.cluster.Provider, da.cluster.ProjectID, err.Error(),
		)
		return err
	}
	da.cloud = cloud
	da.project = project

	return nil
}

func (da *DeleteNodesAction) checkClusterNodeInfoDeletion() ([]string, error) {
	err := da.getClusterBasicInfo()
	if err != nil {
		return nil, err
	}

	// get relative nodes by clusterID
	clusterCond := operator.NewLeafCondition(operator.Eq, operator.M{"clusterid": da.cluster.ClusterID})
	nodeCond := operator.NewLeafCondition(operator.In, operator.M{"innerip": da.nodeIPList})
	cond := operator.NewBranchCondition(operator.And, clusterCond, nodeCond)

	nodes, err := da.model.ListNode(da.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		blog.Errorf("get Cluster %s Nodes failed, %s", da.req.ClusterID, err.Error())
		return nil, err
	}

	nodeInnerIPs := make([]string, 0)
	// filter nodeGroup nodes
	for i := range nodes {
		if nodes[i].NodeGroupID == "" {
			nodeInnerIPs = append(nodeInnerIPs, nodes[i].InnerIP)
		}
	}

	return nodeInnerIPs, nil
}

// Handle delete cluster nodes request
func (da *DeleteNodesAction) Handle(ctx context.Context, req *cmproto.DeleteNodesRequest, resp *cmproto.DeleteNodesResponse) {
	if req == nil || resp == nil {
		blog.Errorf("delete cluster failed, req or resp is empty")
		return
	}
	da.ctx = ctx
	da.req = req
	da.resp = resp

	// check request parameter validate
	if err := da.validate(); err != nil {
		da.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// only delete nodeInfo
	if da.req.OnlyDeleteInfo {
		nodeInnerIPs, err := da.checkClusterNodeInfoDeletion()
		if err != nil || len(nodeInnerIPs) == 0 {
			da.setResp(common.BcsErrClusterManagerDataEmptyErr, fmt.Sprintf("DeleteNodesAction checkClusterNodeInfoDeletion failed: %s",
				"nodeInnerIPs empty"))
			return
		}

		err = da.model.DeleteNodesByIPs(da.ctx, nodeInnerIPs)
		if err != nil {
			da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
			return
		}
		da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
		return
	}

	// step0: get validate node IP Info
	// step1: update node status deleting
	// step2: generate task to async delete node and finally delete db nodes
	// version 1 only to delete node info in db, need to manual delete cluster node by cloud provider
	nodeInnerIPs, err := da.getNodesByClusterAndIPs()
	if err != nil {
		blog.Errorf("DeleteNodesAction getNodesByClusterAndIPs failed: %v", err)
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err = da.updateNodesStatus(common.StatusDeleting, nodeInnerIPs)
	if err != nil {
		blog.Errorf("DeleteNodesAction updateNodesStatus failed: %v", err)
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// build delete nodes task
	err = da.deleteNodesFromClusterTask()
	if err != nil {
		blog.Errorf("delete nodes from cluster %s by cloudprovider %s failed, %s",
			da.cluster.ClusterID, da.cloud.CloudID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerTaskErr, err.Error())
		return
	}

	err = da.model.CreateOperationLog(da.ctx, &cmproto.OperationLog{
		ResourceType: common.Cluster.String(),
		ResourceID:   da.cluster.ClusterID,
		TaskID:       da.task.TaskID,
		Message:      fmt.Sprintf("集群%s下架节点", da.cluster.ClusterID),
		OpUser:       da.req.Operator,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("deleteNodes from cluster[%s] CreateOperationLog failed: %v", da.cluster.ClusterID, err)
	}

	da.resp.Data = da.task
	da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

func (da *DeleteNodesAction) deleteNodesFromClusterTask() error {
	// get cloudprovider cluster implementation
	clusterMgr, err := cloudprovider.GetClusterMgr(da.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s ClusterManager for delete nodes %v from Cluster %s failed, %s",
			da.cloud.CloudProvider, da.req.Nodes, da.req.ClusterID, err.Error(),
		)
		return err
	}

	//get credential for cloudprovider operation
	cmOption, err := cloudprovider.GetCredential(da.project, da.cloud)
	if err != nil {
		blog.Errorf("get Credential for Cloud %s/%s when delete Nodes %s in Cluster %s failed, %s",
			da.cloud.CloudID, da.cloud.CloudProvider, da.nodes, da.cluster.ClusterID, err.Error(),
		)
		return err
	}
	cmOption.Region = da.cluster.Region

	// default cluster delete nodes mode retain
	task, err := clusterMgr.DeleteNodesFromCluster(da.cluster, da.nodes, &cloudprovider.DeleteNodesOption{
		CommonOption: *cmOption,
		Operator:     da.req.Operator,
		DeleteMode:   da.req.DeleteMode,
		Cloud:        da.cloud,
	})
	if err != nil {
		blog.Errorf("cloudprovider %s deleteNodes %v from Cluster %s failed, %s",
			da.cloud.CloudProvider, da.req.Nodes, da.req.ClusterID, err.Error(),
		)
		return err
	}

	// create task
	if err := da.model.CreateTask(da.ctx, task); err != nil {
		blog.Errorf("save deleteNodes task for cluster %s failed, %s",
			da.cluster.ClusterID, err.Error(),
		)
		return err
	}

	// dispatch task
	if err := taskserver.GetTaskServer().Dispatch(task); err != nil {
		blog.Errorf("dispatch deleteNodesFromCluster task for cluster %s failed, %s",
			da.cluster.ClusterName, err.Error(),
		)
		return err
	}

	da.task = task
	blog.Infof("delete nodes %v from cluster %s with cloudprovider %s processing, task info: %v",
		da.req.Nodes, da.req.ClusterID, da.cloud.CloudProvider, task)

	return nil
}

func (da *DeleteNodesAction) updateNodesStatus(status string, ips []string) error {
	for i := range ips {
		node, err := da.model.GetNodeByIP(da.ctx, ips[i])
		if err != nil {
			if errors.Is(err, drivers.ErrTableRecordNotFound) {
				blog.Infof("DeleteNodesAction GetNodeByIP[%s] not exist", ips[i])
				continue
			}
			errMsg := fmt.Sprintf("DeleteNodesAction GetNodeByIP[%s] failed: %v", ips[i], err)
			blog.Errorf(errMsg)
			return errors.New(errMsg)
		}
		node.Status = status
		err = da.model.UpdateNode(da.ctx, node)
		if err != nil {
			errMsg := fmt.Sprintf("DeleteNodesAction UpdateNode[%s] failed: %v", ips[i], err)
			blog.Errorf(errMsg)
			return errors.New(errMsg)
		}
	}

	return nil
}
