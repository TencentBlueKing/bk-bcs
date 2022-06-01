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

package nodegroup

import (
	"context"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// ListAction action for list online cluster credential
type ListAction struct {
	ctx       context.Context
	model     store.ClusterManagerModel
	req       *cmproto.ListNodeGroupRequest
	resp      *cmproto.ListNodeGroupResponse
	groupList []*cmproto.NodeGroup
}

// NewListAction create list action for cluster credential
func NewListAction(model store.ClusterManagerModel) *ListAction {
	return &ListAction{
		model: model,
	}
}

func (la *ListAction) listNodeGroup() error {
	condM := make(operator.M)

	if len(la.req.Name) != 0 {
		condM["name"] = la.req.Name
	}
	if len(la.req.ClusterID) != 0 {
		condM["clusterid"] = la.req.ClusterID
	}
	if len(la.req.ProjectID) != 0 {
		condM["projectid"] = la.req.ProjectID
	}
	if len(la.req.Region) != 0 {
		condM["region"] = la.req.Region
	}

	cond := operator.NewLeafCondition(operator.Eq, condM)
	condStatus := operator.NewLeafCondition(operator.Ne, operator.M{"status": common.StatusDeleted})
	branchCond := operator.NewBranchCondition(operator.And, cond, condStatus)
	groups, err := la.model.ListNodeGroup(la.ctx, branchCond, &storeopt.ListOption{})
	if err != nil {
		return err
	}
	for i := range groups {
		la.groupList = append(la.groupList, &groups[i])
	}
	return nil
}

func (la *ListAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.groupList
}

// Handle handle list cluster credential
func (la *ListAction) Handle(
	ctx context.Context, req *cmproto.ListNodeGroupRequest, resp *cmproto.ListNodeGroupResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list NodeGroup failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listNodeGroup(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

// ListNodesAction action for list online cluster credential
type ListNodesAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	group   *cmproto.NodeGroup
	cluster *cmproto.Cluster
	cloud   *cmproto.Cloud
	req     *cmproto.GetNodeGroupRequest
	resp    *cmproto.ListNodesInGroupResponse
}

// NewListNodesAction create list action for cluster credential
func NewListNodesAction(model store.ClusterManagerModel) *ListNodesAction {
	return &ListNodesAction{
		model: model,
	}
}

func (la *ListNodesAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (la *ListNodesAction) getRelativeResource() error {
	group, err := la.model.GetNodeGroup(la.ctx, la.req.NodeGroupID)
	if err != nil {
		blog.Errorf("get NodeGroup %s failed when list its all nodes, %s", la.req.NodeGroupID, err.Error())
		return err
	}
	la.group = group

	//get relative cluster for information injection
	cluster, err := la.model.GetCluster(la.ctx, la.group.ClusterID)
	if err != nil {
		blog.Errorf("can not get relative Cluster %s when list nodes in group", la.group.ClusterID)
		return fmt.Errorf("get relative cluster %s info err, %s", la.group.ClusterID, err.Error())
	}
	la.cluster = cluster

	cloud, err := actions.GetCloudByCloudID(la.model, la.group.Provider)
	if err != nil {
		blog.Errorf("can not get relative Cloud %s when list nodes in group for Cluster %s, %s",
			la.group.Provider, la.group.ClusterID, err.Error(),
		)
		return err
	}
	la.cloud = cloud

	return nil
}

func (la *ListNodesAction) listNodesInGroup() error {
	// create nodegroup with cloudprovider
	mgr, err := cloudprovider.GetNodeGroupMgr(la.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get NodeGroup Manager cloudprovider %s/%s for list nodes in group %s failed, %s",
			la.cloud.CloudID, la.cloud.CloudProvider, la.group.NodeGroupID, err.Error(),
		)
		return err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     la.cloud,
		AccountID: la.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get Credential for Cloud %s/%s when list nodes in group for cluster %s failed, %s",
			la.cloud.CloudID, la.cloud.CloudProvider, la.cluster.ClusterID, err.Error(),
		)
		return err
	}
	cmOption.Region = la.cluster.Region
	// cloud provider nodeGroup
	nodes, err := mgr.GetNodesInGroup(la.group, cmOption)
	if err != nil {
		blog.Errorf("list group nodes in cloudprovider %s/%s for Cluster %s failed, %s",
			la.cloud.CloudID, la.cloud.CloudProvider, la.cluster.ClusterID, err.Error(),
		)
		return err
	}
	la.resp.Data = nodes
	return nil
}

// Handle handle list cluster credential
func (la *ListNodesAction) Handle(
	ctx context.Context, req *cmproto.GetNodeGroupRequest, resp *cmproto.ListNodesInGroupResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list NodeGroup failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.getRelativeResource(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if err := la.listNodesInGroup(); err != nil {
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
