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

package node

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// UnCordonNodeAction action for uncordon node
type UnCordonNodeAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.UnCordonNodeRequest
	resp  *cmproto.UnCordonNodeResponse
	k8sOp *clusterops.K8SOperator

	failed []string
}

// NewUnCordonNodeAction create update action
func NewUnCordonNodeAction(model store.ClusterManagerModel, k8sOp *clusterops.K8SOperator) *UnCordonNodeAction {
	return &UnCordonNodeAction{
		model:  model,
		k8sOp:  k8sOp,
		failed: make([]string, 0),
	}
}

func (ua *UnCordonNodeAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	return nil
}

func (ua *UnCordonNodeAction) unCordonClusterNodes() error {
	for _, ip := range ua.req.InnerIPs {
		err := ua.k8sOp.ClusterUpdateScheduleNode(ua.ctx, clusterops.NodeInfo{
			ClusterID: ua.req.ClusterID,
			NodeIP:    ip,
			Desired:   false,
		})
		if err != nil {
			blog.Errorf("unCordonClusterNodes[%s] failed: %+v", ip, err)
			ua.failed = append(ua.failed, ip)
		}
	}

	return nil
}

func (ua *UnCordonNodeAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	ua.resp.Fail = ua.failed
}

// Handle handles node uncordon
func (ua *UnCordonNodeAction) Handle(ctx context.Context, req *cmproto.UnCordonNodeRequest, resp *cmproto.UnCordonNodeResponse) {
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
	return
}

// CordonNodeAction action for cordon node
type CordonNodeAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.CordonNodeRequest
	resp  *cmproto.CordonNodeResponse
	k8sOp *clusterops.K8SOperator

	failed []string
}

// NewCordonNodeAction create update action
func NewCordonNodeAction(model store.ClusterManagerModel, k8sOp *clusterops.K8SOperator) *CordonNodeAction {
	return &CordonNodeAction{
		model:  model,
		k8sOp:  k8sOp,
		failed: make([]string, 0),
	}
}

func (ua *CordonNodeAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	return nil
}

func (ua *CordonNodeAction) cordonClusterNodes() error {
	for _, ip := range ua.req.InnerIPs {
		err := ua.k8sOp.ClusterUpdateScheduleNode(ua.ctx, clusterops.NodeInfo{
			ClusterID: ua.req.ClusterID,
			NodeIP:    ip,
			Desired:   true,
		})
		if err != nil {
			blog.Errorf("cordonClusterNodes[%s] failed: %+v", ip, err)
			ua.failed = append(ua.failed, ip)
		}
	}

	return nil
}

func (ua *CordonNodeAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	ua.resp.Fail = ua.failed
}

// Handle handles node cordon
func (ua *CordonNodeAction) Handle(ctx context.Context, req *cmproto.CordonNodeRequest, resp *cmproto.CordonNodeResponse) {
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
	return
}
