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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// FederateAction add single cluster into federation cluster
type FederateAction struct {
	ctx               context.Context
	model             store.ClusterManagerModel
	req               *cmproto.AddFederatedClusterReq
	resp              *cmproto.AddFederatedClusterResp
	federationCluster *cmproto.Cluster
	cluster           *cmproto.Cluster
}

// NewFederateAction create action for adding single cluster into federation cluster
func NewFederateAction(model store.ClusterManagerModel) *FederateAction {
	return &FederateAction{
		model: model,
	}
}

func (fa *FederateAction) validate() error {
	if err := fa.req.Validate(); err != nil {
		return err
	}
	return nil
}

func (fa *FederateAction) getFederationCluster() error {
	cluster, err := fa.model.GetCluster(fa.ctx, fa.req.FederationClusterID)
	if err != nil {
		return err
	}
	if cluster == nil {
		return fmt.Errorf("get federation cluster %s return empty", fa.req.FederationClusterID)
	}
	if cluster.ClusterType != common.ClusterTypeFederation {
		return fmt.Errorf("cluster %s is not federation cluster", fa.req.FederationClusterID)
	}
	fa.federationCluster = cluster
	return nil
}

func (fa *FederateAction) getSingleCluster() error {
	cluster, err := fa.model.GetCluster(fa.ctx, fa.req.ClusterID)
	if err != nil {
		return err
	}
	if cluster == nil {
		return fmt.Errorf("get cluster %s return empty", fa.req.ClusterID)
	}
	if cluster.ClusterType == common.ClusterTypeFederation {
		return fmt.Errorf("cluster %s is federation cluster, cannot join other federation cluster", fa.req.ClusterID)
	}
	if len(cluster.FederationClusterID) != 0 {
		return fmt.Errorf("cluster %s has already joined federation %s, cannot be federated into %s",
			fa.req.ClusterID, cluster.FederationClusterID, fa.req.FederationClusterID)
	}
	fa.cluster = cluster
	return nil
}

func (fa *FederateAction) updateSingleCluster() error {
	if fa.federationCluster.Environment != fa.cluster.Environment {
		return fmt.Errorf("cluster %s and federation %s not in the same environment",
			fa.cluster.ClusterID, fa.federationCluster.ClusterID)
	}
	if fa.federationCluster.ProjectID != fa.cluster.ProjectID ||
		fa.federationCluster.BusinessID != fa.cluster.BusinessID {
		return fmt.Errorf("cluster %s and federation %s has different projectID or businessID",
			fa.cluster.ClusterID, fa.federationCluster.ClusterID)
	}
	fa.cluster.FederationClusterID = fa.federationCluster.ClusterID
	fa.cluster.UpdateTime = time.Now().Format(time.RFC3339)
	return fa.model.UpdateCluster(fa.ctx, fa.cluster)
}

func (fa *FederateAction) setResp(code uint32, msg string) {
	fa.resp.Code = code
	fa.resp.Message = msg
	fa.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handles federate cluster request
func (fa *FederateAction) Handle(ctx context.Context,
	req *cmproto.AddFederatedClusterReq, resp *cmproto.AddFederatedClusterResp) {
	if req == nil || resp == nil {
		blog.Errorf("federate cluster failed, req or resp is empty")
		return
	}
	fa.ctx = ctx
	fa.req = req
	fa.resp = resp

	if err := fa.validate(); err != nil {
		fa.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := fa.getFederationCluster(); err != nil {
		fa.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	if err := fa.getSingleCluster(); err != nil {
		fa.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	if err := fa.updateSingleCluster(); err != nil {
		fa.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err := fa.model.CreateOperationLog(fa.ctx, &cmproto.OperationLog{
		ResourceType: common.Cluster.String(),
		ResourceID:   fa.cluster.ClusterID,
		TaskID:       "",
		Message:      fmt.Sprintf("添加集群%s为联邦集群%s", fa.req.ClusterID, fa.req.FederationClusterID),
		OpUser:       fa.cluster.Creator,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("AddFederatedCluster[%s] CreateOperationLog failed: %v", fa.req.ClusterID, err)
	}
	fa.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
