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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

// ListAction list action for cluster
type ListAction struct {
	ctx         context.Context
	model       store.ClusterManagerModel
	req         *cmproto.ListClusterReq
	resp        *cmproto.ListClusterResp
	clusterList []*cmproto.Cluster
}

// NewListAction create list action for cluster
func NewListAction(model store.ClusterManagerModel) *ListAction {
	return &ListAction{
		model: model,
	}
}

func (la *ListAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}
	return nil
}

func (la *ListAction) listCluster() error {
	condM := make(operator.M)
	if len(la.req.ClusterName) != 0 {
		condM["clusterName"] = la.req.ClusterName
	}
	if len(la.req.Provider) != 0 {
		condM["provider"] = la.req.Provider
	}
	if len(la.req.Region) != 0 {
		condM["region"] = la.req.Region
	}
	if len(la.req.VpcID) != 0 {
		condM["vpcID"] = la.req.VpcID
	}
	if len(la.req.ProjectID) != 0 {
		condM["projectID"] = la.req.ProjectID
	}
	if len(la.req.BusinessID) != 0 {
		condM["businessID"] = la.req.BusinessID
	}
	if len(la.req.Environment) != 0 {
		condM["environment"] = la.req.Environment
	}
	if len(la.req.EngineType) != 0 {
		condM["engineType"] = la.req.EngineType
	}
	cond := operator.NewLeafCondition(operator.Eq, condM)
	clusterList, err := la.model.ListCluster(la.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return err
	}
	for _, cluster := range clusterList {
		la.clusterList = append(la.clusterList, &cmproto.Cluster{
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
		})
	}

	return nil
}

func (la *ListAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == types.BcsErrClusterManagerSuccess)
	la.resp.Data = la.clusterList
}

// Handle handle list cluster request
func (la *ListAction) Handle(ctx context.Context, req *cmproto.ListClusterReq, resp *cmproto.ListClusterResp) {
	if req == nil || resp == nil {
		blog.Errorf("list cluster failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(types.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listCluster(); err != nil {
		la.setResp(types.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(types.BcsErrClusterManagerSuccess, types.BcsErrClusterManagerSuccessStr)
	return
}
