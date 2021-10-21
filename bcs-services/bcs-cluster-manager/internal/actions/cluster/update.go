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
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	cmcommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

// UpdateAction action for update cluster
type UpdateAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	req     *cmproto.UpdateClusterReq
	resp    *cmproto.UpdateClusterResp
	cluster *types.Cluster
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
	if !cmcommon.IsEngineTypeValid(ua.req.EngineType) {
		return fmt.Errorf("invalid engine type")
	}
	if !cmcommon.IsClusterTypeValid(ua.req.ClusterType) {
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
	newCluster := &types.Cluster{
		ClusterID:           ua.cluster.ClusterID,
		ClusterName:         ua.req.ClusterName,
		Provider:            ua.req.Provider,
		Region:              ua.req.Region,
		VpcID:               ua.req.VpcID,
		ProjectID:           ua.req.ProjectID,
		BusinessID:          ua.req.BusinessID,
		Environment:         ua.req.Environment,
		EngineType:          ua.req.EngineType,
		IsExclusive:         ua.req.IsExclusive,
		ClusterType:         ua.req.ClusterType,
		Labels:              ua.req.Labels,
		Operators:           ua.req.Operators,
		CreateTime:          ua.cluster.CreateTime,
		UpdateTime:          time.Now(),
		FederationClusterID: ua.cluster.FederationClusterID,
	}
	return ua.model.UpdateCluster(ua.ctx, newCluster)
}

func (ua *UpdateAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == types.BcsErrClusterManagerSuccess)
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
		ua.setResp(types.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ua.getCluster(); err != nil {
		ua.setResp(types.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if err := ua.updateCluster(); err != nil {
		ua.setResp(types.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ua.setResp(types.BcsErrClusterManagerSuccess, types.BcsErrClusterManagerSuccessStr)
	return
}
