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

package task

import (
	"context"
	"errors"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"

	"github.com/google/uuid"
)

// CreateAction action for create namespace
type CreateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.CreateTaskRequest
	resp  *cmproto.CreateTaskResponse
}

// NewCreateAction create namespace action
func NewCreateAction(model store.ClusterManagerModel) *CreateAction {
	return &CreateAction{
		model: model,
	}
}

func (ca *CreateAction) createTask() error {
	timeStr := time.Now().Format(time.RFC3339)
	t := &cmproto.Task{
		TaskID:       uuid.New().String(),
		TaskType:     ca.req.TaskType,
		Start:        timeStr,
		Status:       "INITIALIZING",
		StepSequence: ca.req.StepSequence,
		Steps:        ca.req.Steps,
		ClusterID:    ca.req.ClusterID,
		ProjectID:    ca.req.ProjectID,
		Creator:      ca.req.Creator,
		LastUpdate:   timeStr,
	}
	ca.resp.Data = t
	//write to database before push to worker
	if err := ca.model.CreateTask(ca.ctx, t); err != nil {
		blog.Errorf("write task to storage failed, %s", err.Error())
		return err
	}
	// push to task worker
	return nil
}

func (ca *CreateAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle create namespace request
func (ca *CreateAction) Handle(ctx context.Context,
	req *cmproto.CreateTaskRequest, resp *cmproto.CreateTaskResponse) {
	if req == nil || resp == nil {
		blog.Errorf("create Task failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := req.Validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ca.createTask(); err != nil {
		if errors.Is(err, drivers.ErrTableRecordDuplicateKey) {
			ca.setResp(common.BcsErrClusterManagerDatabaseRecordDuplicateKey, err.Error())
			return
		}
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
