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

package cluster

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// ListBusinessClusterAction list action for business clusters
type ListBusinessClusterAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	iam   iam.PermClient

	req         *cmproto.ListBusinessClusterReq
	resp        *cmproto.ListBusinessClusterResp
	clusterList []*cmproto.BusinessCluster
}

// NewListBusinessClusterAction create list action for business cluster
func NewListBusinessClusterAction(model store.ClusterManagerModel, iam iam.PermClient) *ListBusinessClusterAction {
	return &ListBusinessClusterAction{
		model: model,
		iam:   iam,
	}
}

func (la *ListBusinessClusterAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.clusterList
}

// Handle list business cluster request
func (la *ListBusinessClusterAction) Handle(ctx context.Context,
	req *cmproto.ListBusinessClusterReq, resp *cmproto.ListBusinessClusterResp) {
	if req == nil || resp == nil {
		blog.Errorf("list business cluster failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listBusinessCluster(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

func (la *ListBusinessClusterAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}

	// check operator host permission
	canUse := CheckUserHasPerm(la.req.BusinessID, la.req.Operator)
	if !canUse {
		errMsg := fmt.Errorf("list business cluster failed: user[%s] no perm in bizID[%s]",
			la.req.Operator, la.req.BusinessID)
		blog.Errorf(errMsg.Error())
		return errMsg
	}

	return nil
}

// listBusinessCluster get business clusters
func (la *ListBusinessClusterAction) listBusinessCluster() error {
	condM := make(operator.M)

	if len(la.req.BusinessID) != 0 {
		condM["businessid"] = la.req.BusinessID
	}

	condCluster := operator.NewLeafCondition(operator.Eq, condM)
	condStatus := operator.NewLeafCondition(operator.Ne, operator.M{"status": common.StatusDeleted})

	branchCond := operator.NewBranchCondition(operator.And, condCluster, condStatus)
	clusterList, err := la.model.ListCluster(la.ctx, branchCond, &storeopt.ListOption{})
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		blog.Errorf("ListBusinessClusterAction ListCluster failed: %v", err)
		return err
	}

	// cluster sort
	var (
		otherCluster   = make([]*cmproto.BusinessCluster, 0)
		runningCluster = make([]*cmproto.BusinessCluster, 0)
	)
	for i := range clusterList {
		if clusterList[i].IsShared {
			clusterList[i].IsShared = false
		}

		if clusterList[i].Status == common.StatusRunning {
			runningCluster = append(runningCluster, clusterToBusinessCluster(clusterList[i]))
		} else {
			otherCluster = append(otherCluster, clusterToBusinessCluster(clusterList[i]))
		}
	}

	la.clusterList = append(la.clusterList, otherCluster...)
	la.clusterList = append(la.clusterList, runningCluster...)

	return nil
}

func clusterToBusinessCluster(cluster *cmproto.Cluster) *cmproto.BusinessCluster {
	return &cmproto.BusinessCluster{
		ClusterID:       cluster.ClusterID,
		ClusterName:     cluster.ClusterName,
		Provider:        cluster.Provider,
		Region:          cluster.Region,
		VpcID:           cluster.VpcID,
		ProjectID:       cluster.ProjectID,
		BusinessID:      cluster.BusinessID,
		Environment:     cluster.Environment,
		EngineType:      cluster.EngineType,
		ClusterType:     cluster.ClusterType,
		Labels:          cluster.Labels,
		Creator:         cluster.Creator,
		CreateTime:      cluster.CreateTime,
		UpdateTime:      cluster.UpdateTime,
		SystemID:        cluster.SystemID,
		ManageType:      cluster.ManageType,
		Status:          cluster.Status,
		Updater:         cluster.Updater,
		NetworkType:     cluster.NetworkType,
		ModuleID:        cluster.ModuleID,
		IsCommonCluster: cluster.IsCommonCluster,
		Description:     cluster.Description,
		ClusterCategory: cluster.ClusterCategory,
		IsShared:        cluster.IsShared,
	}
}
