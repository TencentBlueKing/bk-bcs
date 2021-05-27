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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

// GetAction action for get cluster
type GetAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	req     *cmproto.GetClusterReq
	resp    *cmproto.GetClusterResp
	cluster *cmproto.Cluster
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
	ga.cluster = &cmproto.Cluster{
		ClusterID:           cluster.ClusterID,
		ClusterName:         cluster.ClusterName,
		FederationClusterID: cluster.FederationClusterID,
		Provider:            cluster.Provider,
		Region:              cluster.Region,
		VpcID:               cluster.VpcID,
		ProjectID:           cluster.ProjectID,
		BusinessID:          cluster.BusinessID,
		Environment:         cluster.Environment,
		EngineType:          cluster.EngineType,
		IsExclusive:         cluster.IsExclusive,
		ClusterType:         cluster.ClusterType,
		Labels:              cluster.Labels,
		Operators:           cluster.Operators,
		CreateTime:          cluster.CreateTime.String(),
		UpdateTime:          cluster.UpdateTime.String(),
	}
	return nil
}

func (ga *GetAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == types.BcsErrClusterManagerSuccess)
	ga.resp.Data = ga.cluster
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
		ga.setResp(types.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ga.getCluster(); err != nil {
		ga.setResp(types.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ga.setResp(types.BcsErrClusterManagerSuccess, types.BcsErrClusterManagerSuccessStr)

	return
}
