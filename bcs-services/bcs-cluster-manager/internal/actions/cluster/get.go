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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// GetAction action for get cluster
type GetAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	req     *cmproto.GetClusterReq
	resp    *cmproto.GetClusterResp
	cluster *cmproto.Cluster
	cloud   *cmproto.Cloud
}

// NewGetAction create get action
func NewGetAction(model store.ClusterManagerModel) *GetAction {
	return &GetAction{
		model: model,
	}
}

func (ga *GetAction) validate() error {
	return ga.req.Validate()
}

func (ga *GetAction) getCluster() error {
	cluster, err := ga.model.GetCluster(ga.ctx, ga.req.ClusterID)
	if err != nil {
		return err
	}
	ga.cluster = cluster
	return nil
}

func (ga *GetAction) getCloud() error {
	cloud, err := ga.model.GetCloud(ga.ctx, ga.cluster.Provider)
	if err != nil {
		return err
	}
	ga.cloud = cloud
	return nil
}

func (ga *GetAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	ga.resp.Data = ga.cluster
	if ga.resp.Extra == nil {
		ga.resp.Extra = &cmproto.ExtraClusterInfo{}
	}
	ga.resp.Extra.ProviderType = ga.cloud.GetEngineType()
}

// Handle get cluster request
func (ga *GetAction) Handle(ctx context.Context, req *cmproto.GetClusterReq, resp *cmproto.GetClusterResp) {
	if req == nil || resp == nil {
		blog.Errorf("get cluster failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	if err := ga.validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ga.getCluster(); err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	if err := ga.getCloud(); err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)

	return
}

// GetNodeAction action for get cluster
type GetNodeAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	resp  *cmproto.GetNodeResponse
}

// NewGetNodeAction create get action
func NewGetNodeAction(model store.ClusterManagerModel) *GetNodeAction {
	return &GetNodeAction{
		model: model,
	}
}

func (ga *GetNodeAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle get node request, attention innerIP same in different cluster
func (ga *GetNodeAction) Handle(ctx context.Context, req *cmproto.GetNodeRequest, resp *cmproto.GetNodeResponse) {
	if req == nil || resp == nil {
		blog.Errorf("get node failed, req or resp is empty")
		return
	}
	ga.resp = resp

	if err := req.Validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	node, err := ga.model.GetNodeByIP(ctx, req.InnerIP)
	if err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	if !req.ShowPwd {
		node.Passwd = ""
	}
	resp.Data = append(resp.Data, node)
	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)

	return
}

// CheckNodeAction action for check node in cluster
type CheckNodeAction struct {
	ctx        context.Context
	model      store.ClusterManagerModel
	req        *cmproto.CheckNodesRequest
	resp       *cmproto.CheckNodesResponse
	nodeResult map[string]*cmproto.NodeResult
}

// NewCheckNodeAction create checkNode action
func NewCheckNodeAction(model store.ClusterManagerModel) *CheckNodeAction {
	return &CheckNodeAction{
		model:      model,
		nodeResult: make(map[string]*cmproto.NodeResult),
	}
}

func (ca *CheckNodeAction) validate() error {
	if err := ca.req.Validate(); err != nil {
		return err
	}

	return nil
}

func (ca *CheckNodeAction) checkNodesInCluster() error {
	if ca.nodeResult == nil {
		ca.nodeResult = make(map[string]*cmproto.NodeResult)
	}

	for i := range ca.req.InnerIPs {
		nodeResult, err := ca.getNodeResultByNodeIP(ca.req.InnerIPs[i])
		if err != nil {
			blog.Errorf("CheckNodeAction getNodeResultByNodeIP failed: %v", err)
			continue
		}

		ca.nodeResult[ca.req.InnerIPs[i]] = nodeResult
	}

	return nil
}

func (ca *CheckNodeAction) getNodeResultByNodeIP(nodeIP string) (*cmproto.NodeResult, error) {
	nodeResult := &cmproto.NodeResult{
		IsExist:     false,
		ClusterID:   "",
		ClusterName: "",
	}

	node, err := ca.model.GetNodeByIP(ca.ctx, nodeIP)
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		return nodeResult, err
	}

	if errors.Is(err, drivers.ErrTableRecordNotFound) {
		return nodeResult, nil
	}

	nodeResult.IsExist = true
	nodeResult.ClusterID = node.ClusterID
	if len(node.ClusterID) != 0 {
		cluster, err := ca.model.GetCluster(ca.ctx, node.ClusterID)
		if err == nil {
			nodeResult.ClusterName = cluster.GetClusterName()
		}
	}

	return nodeResult, nil
}

func (ca *CheckNodeAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	if ca.resp.Data == nil {
		ca.resp.Data = make(map[string]*cmproto.NodeResult)
	}
	ca.resp.Data = ca.nodeResult
}

// Handle handles check nodes in cluster request
func (ca *CheckNodeAction) Handle(ctx context.Context, req *cmproto.CheckNodesRequest, resp *cmproto.CheckNodesResponse) {
	if req == nil || resp == nil {
		blog.Errorf("check cluster node failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := ca.validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ca.checkNodesInCluster(); err != nil {
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
