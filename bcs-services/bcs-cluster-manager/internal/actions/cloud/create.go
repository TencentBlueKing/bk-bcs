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

package cloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// CreateAction action for create namespace
type CreateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.CreateCloudRequest
	resp  *cmproto.CreateCloudResponse
}

// NewCreateAction create namespace action
func NewCreateAction(model store.ClusterManagerModel) *CreateAction {
	return &CreateAction{
		model: model,
	}
}

func (ca *CreateAction) createCloud() error {
	timeStr := time.Now().Format(time.RFC3339)
	cloud := &cmproto.Cloud{
		CloudID:             ca.req.CloudID,
		Name:                ca.req.Name,
		Editable:            ca.req.Editable,
		OpsPlugins:          ca.req.OpsPlugins,
		ExtraPlugins:        ca.req.ExtraPlugins,
		CloudCredential:     ca.req.CloudCredential,
		OsManagement:        ca.req.OsManagement,
		ClusterManagement:   ca.req.ClusterManagement,
		NodeGroupManagement: ca.req.NodeGroupManagement,
		CloudProvider:       ca.req.CloudProvider,
		Config:              ca.req.Config,
		Description:         ca.req.Description,
		Creator:             ca.req.Creator,
		CreatTime:           timeStr,
		UpdateTime:          timeStr,
		EngineType:          ca.req.EngineType,
		Enable:              ca.req.Enable,
	}
	return ca.model.CreateCloud(ca.ctx, cloud)
}

func (ca *CreateAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ca *CreateAction) validate() error {
	if err := ca.req.Validate(); err != nil {
		return err
	}

	if !utils.StringInSlice(ca.req.Enable, cloudEnable) {
		return fmt.Errorf("cloud enable parameter invalid, must be true or false")
	}

	return nil
}

// Handle create namespace request
func (ca *CreateAction) Handle(ctx context.Context,
	req *cmproto.CreateCloudRequest, resp *cmproto.CreateCloudResponse) {
	if req == nil || resp == nil {
		blog.Errorf("create cloud failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := ca.validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ca.createCloud(); err != nil {
		if errors.Is(err, drivers.ErrTableRecordDuplicateKey) {
			ca.setResp(common.BcsErrClusterManagerDatabaseRecordDuplicateKey, err.Error())
			return
		}
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err := ca.model.CreateOperationLog(ca.ctx, &cmproto.OperationLog{
		ResourceType: common.Cloud.String(),
		ResourceID:   req.CloudID,
		TaskID:       "",
		Message:      fmt.Sprintf("创建云[%s]模板", req.CloudID),
		OpUser:       req.Creator,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("CreateCloud[%s] CreateOperationLog failed: %v", req.CloudID, err)
	}

	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
