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
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// UnCordonNodeAction action for uncordon node
type UnCordonNodeAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.UnCordonNodeRequest
	resp  *cmproto.UnCordonNodeResponse
	k8sOp *clusterops.K8SOperator

	cluster *cmproto.Cluster
}

// NewUnCordonNodeAction create update action
func NewUnCordonNodeAction(model store.ClusterManagerModel, k8sOp *clusterops.K8SOperator) *UnCordonNodeAction {
	return &UnCordonNodeAction{
		model: model,
		k8sOp: k8sOp,
	}
}

func (ua *UnCordonNodeAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	// get relative cluster for information injection
	cluster, err := ua.model.GetCluster(ua.ctx, ua.req.ClusterID)
	if err == nil {
		ua.cluster = cluster
	}

	return nil
}

func (ua *UnCordonNodeAction) unCordonClusterNodes() error {
	// get node names
	if len(ua.req.Nodes) == 0 && len(ua.req.InnerIPs) > 0 {
		option := clusterops.ListNodeOption{ClusterID: ua.req.ClusterID, NodeIPs: ua.req.InnerIPs}
		nodes, err := ua.k8sOp.ListClusterNodesByIPsOrNames(ua.ctx, option)
		if err != nil {
			blog.Errorf("get nodename by ips failed in cluster %s, err %s", ua.req.ClusterID, err.Error())
			return fmt.Errorf("get nodename by ips failed in cluster %s, err %s", ua.req.ClusterID, err.Error())
		}
		for _, v := range nodes {
			ua.req.Nodes = append(ua.req.Nodes, v.Name)
		}
	}

	successCh := make(chan *cmproto.NodeOperationStatusInfo, len(ua.req.Nodes))
	failCh := make(chan *cmproto.NodeOperationStatusInfo, len(ua.req.Nodes))

	barrier := utils.NewRoutinePool(50)
	defer barrier.Close()

	for i := range ua.req.Nodes {
		barrier.Add(1)
		go func(node string) {
			defer barrier.Done()
			ctx, cancel := context.WithTimeout(context.Background(), clusterops.DefaultTimeout)
			defer cancel()
			if err := ua.k8sOp.ClusterUpdateScheduleNode(ctx, clusterops.NodeInfo{
				ClusterID: ua.req.ClusterID,
				NodeName:  node,
				Desired:   false,
			}); err != nil {
				failCh <- &cmproto.NodeOperationStatusInfo{NodeName: node, Message: err.Error()}
				blog.Errorf("unCordonClusterNodes[%s] failed in cluster %s, err %s", node, ua.req.ClusterID, err.Error())
				return
			}
			successCh <- &cmproto.NodeOperationStatusInfo{NodeName: node}
		}(ua.req.Nodes[i])
	}
	barrier.Wait()
	close(successCh)
	close(failCh)

	ua.resp.Data = &cmproto.NodeOperationStatus{
		Success: make([]*cmproto.NodeOperationStatusInfo, 0),
		Fail:    make([]*cmproto.NodeOperationStatusInfo, 0),
	}
	for v := range successCh {
		ua.resp.Data.Success = append(ua.resp.Data.Success, v)
	}
	for v := range failCh {
		ua.resp.Data.Fail = append(ua.resp.Data.Fail, v)
	}

	// record operation log
	err := ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.Cluster.String(),
		ResourceID:   ua.req.GetClusterID(),
		TaskID:       "",
		Message:      fmt.Sprintf("集群%s设置节点可调度状态", ua.req.ClusterID),
		OpUser:       auth.GetUserFromCtx(ua.ctx),
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    ua.req.ClusterID,
		ProjectID:    ua.cluster.ProjectID,
		ResourceName: ua.cluster.ClusterName,
	})
	if err != nil {
		blog.Errorf("unCordonClusterNodes[%s] CreateOperationLog failed: %v", ua.cluster.ClusterID, err)
	}

	return nil
}

func (ua *UnCordonNodeAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handles node uncordon
func (ua *UnCordonNodeAction) Handle(
	ctx context.Context, req *cmproto.UnCordonNodeRequest, resp *cmproto.UnCordonNodeResponse) {
	if req == nil || resp == nil {
		blog.Errorf("uncordon cluster node failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ua.unCordonClusterNodes(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// CordonNodeAction action for cordon node
type CordonNodeAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	req     *cmproto.CordonNodeRequest
	resp    *cmproto.CordonNodeResponse
	k8sOp   *clusterops.K8SOperator
	cluster *cmproto.Cluster
}

// NewCordonNodeAction create update action
func NewCordonNodeAction(model store.ClusterManagerModel, k8sOp *clusterops.K8SOperator) *CordonNodeAction {
	return &CordonNodeAction{
		model: model,
		k8sOp: k8sOp,
	}
}

func (ua *CordonNodeAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	// get relative cluster for information injection
	cluster, err := ua.model.GetCluster(ua.ctx, ua.req.ClusterID)
	if err == nil {
		ua.cluster = cluster
	}

	return nil
}

func (ua *CordonNodeAction) cordonClusterNodes() error {
	// get node names
	if len(ua.req.Nodes) == 0 && len(ua.req.InnerIPs) > 0 {
		option := clusterops.ListNodeOption{ClusterID: ua.req.ClusterID, NodeIPs: ua.req.InnerIPs}
		nodes, err := ua.k8sOp.ListClusterNodesByIPsOrNames(ua.ctx, option)
		if err != nil {
			blog.Errorf("get nodename by ips failed in cluster %s, err %s", ua.req.ClusterID, err.Error())
			return fmt.Errorf("get nodename by ips failed in cluster %s, err %s", ua.req.ClusterID, err.Error())
		}
		for _, v := range nodes {
			ua.req.Nodes = append(ua.req.Nodes, v.Name)
		}
	}

	successCh := make(chan *cmproto.NodeOperationStatusInfo, len(ua.req.Nodes))
	failCh := make(chan *cmproto.NodeOperationStatusInfo, len(ua.req.Nodes))

	barrier := utils.NewRoutinePool(50)
	defer barrier.Close()

	for i := range ua.req.Nodes {
		barrier.Add(1)
		go func(node string) {
			defer barrier.Done()
			ctx, cancel := context.WithTimeout(context.Background(), clusterops.DefaultTimeout)
			defer cancel()
			if err := ua.k8sOp.ClusterUpdateScheduleNode(ctx, clusterops.NodeInfo{
				ClusterID: ua.req.ClusterID,
				NodeName:  node,
				Desired:   true,
			}); err != nil {
				failCh <- &cmproto.NodeOperationStatusInfo{NodeName: node, Message: err.Error()}
				blog.Errorf("cordonClusterNodes[%s] failed in cluster %s, err %s", node, ua.req.ClusterID, err.Error())
				return
			}
			successCh <- &cmproto.NodeOperationStatusInfo{NodeName: node}
		}(ua.req.Nodes[i])
	}
	barrier.Wait()
	close(successCh)
	close(failCh)

	ua.resp.Data = &cmproto.NodeOperationStatus{
		Success: make([]*cmproto.NodeOperationStatusInfo, 0),
		Fail:    make([]*cmproto.NodeOperationStatusInfo, 0),
	}
	for v := range successCh {
		ua.resp.Data.Success = append(ua.resp.Data.Success, v)
	}
	for v := range failCh {
		ua.resp.Data.Fail = append(ua.resp.Data.Fail, v)
	}

	// record operation log
	err := ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.Cluster.String(),
		ResourceID:   ua.req.GetClusterID(),
		TaskID:       "",
		Message:      fmt.Sprintf("集群%s设置节点不可调度状态", ua.req.ClusterID),
		OpUser:       auth.GetUserFromCtx(ua.ctx),
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    ua.req.ClusterID,
		ProjectID:    ua.cluster.ProjectID,
		ResourceName: ua.cluster.ClusterName,
	})
	if err != nil {
		blog.Errorf("unCordonClusterNodes[%s] CreateOperationLog failed: %v", ua.cluster.ClusterID, err)
	}

	return nil
}

func (ua *CordonNodeAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handles node cordon
func (ua *CordonNodeAction) Handle(
	ctx context.Context, req *cmproto.CordonNodeRequest, resp *cmproto.CordonNodeResponse) {
	if req == nil || resp == nil {
		blog.Errorf("cordon cluster node failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ua.cordonClusterNodes(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
