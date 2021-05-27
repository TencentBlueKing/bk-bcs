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
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	cmcommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

// CreateAction action for create cluster
type CreateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.CreateClusterReq
	resp  *cmproto.CreateClusterResp
}

// NewCreateAction create cluster action
func NewCreateAction(model store.ClusterManagerModel) *CreateAction {
	return &CreateAction{
		model: model,
	}
}

func (ca *CreateAction) validate() error {
	if err := ca.req.Validate(); err != nil {
		return err
	}
	if !cmcommon.IsEngineTypeValid(ca.req.EngineType) {
		return fmt.Errorf("invalid engine type")
	}
	if !cmcommon.IsClusterTypeValid(ca.req.ClusterType) {
		return fmt.Errorf("invalid cluster type")
	}
	return nil
}

func (ca *CreateAction) createCluster() error {
	createTime := time.Now()
	newCluster := &types.Cluster{
		ClusterID:   ca.req.ClusterID,
		ClusterName: ca.req.ClusterName,
		Provider:    ca.req.Provider,
		Region:      ca.req.Region,
		VpcID:       ca.req.VpcID,
		ProjectID:   ca.req.ProjectID,
		BusinessID:  ca.req.BusinessID,
		Environment: ca.req.Environment,
		EngineType:  ca.req.EngineType,
		IsExclusive: ca.req.IsExclusive,
		ClusterType: ca.req.ClusterType,
		Labels:      ca.req.Labels,
		Operators:   ca.req.Operators,
		CreateTime:  createTime,
		UpdateTime:  createTime,
	}
	return ca.model.CreateCluster(ca.ctx, newCluster)
}

func (ca *CreateAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == types.BcsErrClusterManagerSuccess)
}

// Handle create cluster request
func (ca *CreateAction) Handle(ctx context.Context, req *cmproto.CreateClusterReq, resp *cmproto.CreateClusterResp) {
	if req == nil || resp == nil {
		blog.Errorf("create cluster failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := ca.validate(); err != nil {
		ca.setResp(types.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ca.createCluster(); err != nil {
		if errors.Is(err, drivers.ErrTableRecordDuplicateKey) {
			ca.setResp(types.BcsErrClusterManagerDatabaseRecordDuplicateKey, err.Error())
			return
		}
		ca.setResp(types.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ca.setResp(types.BcsErrClusterManagerSuccess, types.BcsErrClusterManagerSuccessStr)
	return
}
