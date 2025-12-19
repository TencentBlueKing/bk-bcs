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

package node

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// ListClusterNodeAction action for list cluster node
type ListClusterNodeAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.ListClusterNodesRequest
	resp  *cmproto.ListClusterNodesResponse
}

// NewListClusterNodeAction list cluster node action
func NewListClusterNodeAction(model store.ClusterManagerModel) *ListClusterNodeAction {
	return &ListClusterNodeAction{
		model: model,
	}
}

func (l *ListClusterNodeAction) validate() error {
	if err := l.req.Validate(); err != nil {
		return err
	}

	return nil
}

func (l *ListClusterNodeAction) setResp(code uint32, msg string) {
	l.resp.Code = code
	l.resp.Message = msg
	l.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handles list nodes request
func (l *ListClusterNodeAction) Handle(ctx context.Context, req *cmproto.ListClusterNodesRequest,
	resp *cmproto.ListClusterNodesResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list cluster nodes failed, req or resp is empty")
		return
	}
	l.ctx = ctx
	l.req = req
	l.resp = resp

	if err := l.validate(); err != nil {
		l.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := l.listNodes(); err != nil {
		l.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	l.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// listNodes merge cluster and db nodes
func (l *ListClusterNodeAction) listNodes() error {
	condM := make(operator.M)
	if len(l.req.ClusterID) != 0 && l.req.ClusterID != "-" {
		condM["clusterid"] = l.req.ClusterID
	}
	if len(l.req.Region) != 0 {
		condM["region"] = l.req.Region
	}
	if len(l.req.VpcID) != 0 {
		condM["vpc"] = l.req.VpcID
	}
	if len(l.req.NodeGroupID) != 0 {
		condM["nodegroupid"] = l.req.NodeGroupID
	}
	if len(l.req.InstanceType) != 0 {
		condM["instancetype"] = l.req.InstanceType
	}
	if len(l.req.Status) != 0 {
		condM["status"] = l.req.Status
	}
	if len(l.req.InnerIP) != 0 {
		condM["innerip"] = l.req.InnerIP
	}
	cond := operator.NewLeafCondition(operator.Eq, condM)
	nodes, err := l.model.ListNode(l.ctx, cond, &options.ListOption{
		Offset: int64(l.req.Offset),
		Limit:  int64(l.req.Limit),
	})
	if err != nil {
		blog.Errorf("list nodes in cluster %s failed, %s", l.req.ClusterID, err.Error())
		return err
	}

	// remove passwd
	if !l.req.ShowPwd {
		removeNodeSensitiveInfo(nodes)
	}

	cmNodes := make([]*cmproto.ClusterNode, 0)
	for i := range nodes {
		cmNodes = append(cmNodes, transNodeToClusterNode(l.model, nodes[i]))
	}

	l.resp.Data = cmNodes

	return nil
}

func removeNodeSensitiveInfo(nodes []*cmproto.Node) {
	for i := range nodes {
		nodes[i].Passwd = ""
	}
}

func transNodeToClusterNode(model store.ClusterManagerModel, node *cmproto.Node) *cmproto.ClusterNode {
	var (
		nodeGroupName = ""
	)
	if node.NodeGroupID != "" {
		group, err := model.GetNodeGroup(context.Background(), node.NodeGroupID)
		if err != nil {
			blog.Warnf("transNodeToClusterNode GetNodeGroup[%s] failed: %v", node.NodeGroupID, err)
		} else {
			nodeGroupName = group.Name
		}
	}

	return &cmproto.ClusterNode{
		NodeID:        node.NodeID,
		InnerIP:       node.InnerIP,
		InstanceType:  node.InstanceType,
		CPU:           node.CPU,
		Mem:           node.Mem,
		GPU:           node.GPU,
		Status:        node.Status,
		ZoneID:        node.ZoneID,
		NodeGroupID:   node.NodeGroupID,
		ClusterID:     node.ClusterID,
		VPC:           node.VPC,
		Region:        node.Region,
		Passwd:        node.Passwd,
		Zone:          node.Zone,
		DeviceID:      node.DeviceID,
		NodeName:      node.NodeName,
		NodeGroupName: nodeGroupName,
		InnerIPv6:     node.InnerIPv6,
		TaskID:        node.TaskID,
		ZoneName:      node.ZoneName,
		FailedReason:  node.FailedReason,
	}
}
