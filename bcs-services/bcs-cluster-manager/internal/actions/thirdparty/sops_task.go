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

package thirdparty

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/google/uuid"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// DebugBkSopsTaskAction action for debug sops task
type DebugBkSopsTaskAction struct {
	ctx context.Context

	model store.ClusterManagerModel
	req   *cmproto.DebugBkSopsTaskRequest
	resp  *cmproto.DebugBkSopsTaskResponse
	task  *cmproto.Task
}

// NewDebugBkSopsTaskAction create list action for business templateList
func NewDebugBkSopsTaskAction(model store.ClusterManagerModel) *DebugBkSopsTaskAction {
	return &DebugBkSopsTaskAction{
		model: model,
	}
}

func (da *DebugBkSopsTaskAction) setResp(code uint32, msg string) {
	da.resp.Code = code
	da.resp.Message = msg
	if da.resp.Data == nil {
		da.resp.Data = &cmproto.DebugBkSopsTaskInfo{}
	}
	da.resp.Data.Task = da.task
}

func (da *DebugBkSopsTaskAction) validate() error {
	err := da.req.Validate()
	if err != nil {
		return err
	}
	return nil
}

// Handle handle debug bkSops task
func (da *DebugBkSopsTaskAction) Handle(
	ctx context.Context, req *cmproto.DebugBkSopsTaskRequest, resp *cmproto.DebugBkSopsTaskResponse) {
	if req == nil || resp == nil {
		blog.Errorf("ListTemplateListAction failed, req or resp is empty")
		return
	}
	da.ctx = ctx
	da.req = req
	da.resp = resp

	if err := da.validate(); err != nil {
		da.setResp(icommon.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := da.createDispatchTask(); err != nil {
		da.setResp(icommon.BcsErrClusterManagerTaskErr, err.Error())
		return
	}
	da.setResp(icommon.BcsErrClusterManagerSuccess, icommon.BcsErrClusterManagerSuccessStr)
}

func (da *DebugBkSopsTaskAction) createDispatchTask() error {
	task, err := da.buildDebugSopsTask()
	if err != nil {
		blog.Errorf("CreateDispatchTask BuildDebugSopsTask failed: %v", err)
		return err
	}
	da.task = task

	err = CreateDispatchTask(da.ctx, da.model, task)
	if err != nil {
		blog.Errorf("CreateDispatchTask CreateDispatchTask failed: %v", err)
		return err
	}

	return nil
}

// CreateDispatchTask create and dispatch task
func CreateDispatchTask(ctx context.Context, model store.ClusterManagerModel, task *cmproto.Task) error {
	// create task and dispatch task
	if err := model.CreateTask(ctx, task); err != nil {
		return err
	}
	if err := taskserver.GetTaskServer().Dispatch(task); err != nil {
		return err
	}

	return nil
}

// BuildUpdateDesiredNodesTask build update desired nodes task
func (da *DebugBkSopsTaskAction) buildDebugSopsTask() (*cmproto.Task, error) {
	// generate main task
	nowStr := time.Now().Format(time.RFC3339)
	task := &cmproto.Task{
		TaskID:         uuid.New().String(),
		TaskType:       cloudprovider.DebugBkSopsTask.String(),
		TaskName:       cloudprovider.DebugBkSopsTaskName.String(),
		Status:         cloudprovider.TaskStatusInit,
		Message:        "task initializing",
		Start:          nowStr,
		Steps:          make(map[string]*cmproto.Step),
		StepSequence:   make([]string, 0),
		ClusterID:      "",
		ProjectID:      "",
		Creator:        da.req.Operator,
		Updater:        da.req.Operator,
		LastUpdate:     nowStr,
		CommonParams:   make(map[string]string),
		ForceTerminate: false,
	}
	// generate taskName
	taskName := fmt.Sprintf("调试业务[%s:%s]标准运维任务", da.req.BusinessID, da.req.TemplateID)
	task.CommonParams[cloudprovider.TaskNameKey.String()] = taskName

	err := da.generateBKopsStep(task)
	if err != nil {
		return nil, err
	}
	// set current step
	if len(task.StepSequence) == 0 {
		return nil, fmt.Errorf("BuildDispatchDebugSopsTask task StepSequence empty")
	}
	task.CurrentStep = task.StepSequence[0]

	return task, nil
}

func (da *DebugBkSopsTaskAction) generateBKopsStep(task *cmproto.Task) error {
	now := time.Now().Format(time.RFC3339)

	stepName := cloudprovider.BKSOPTask + "-" + utils.RandomString(8)
	step := &cmproto.Step{
		Name:   stepName,
		System: "bksops",
		Params: make(map[string]string),
		Retry:  0,
		Start:  now,
		Status: cloudprovider.TaskStatusNotStarted,
		// method name is registered name to taskServer
		TaskMethod: cloudprovider.BKSOPTask,
		TaskName:   "标准运维任务",
	}
	step.Params[cloudprovider.BkSopsURLKey.String()] = ""
	step.Params[cloudprovider.BkSopsBizIDKey.String()] = da.req.BusinessID
	step.Params[cloudprovider.BkSopsTemplateIDKey.String()] = da.req.TemplateID
	step.Params[cloudprovider.BkSopsTemplateUserKey.String()] = da.req.Operator
	step.Params[cloudprovider.BkSopsTemplateSourceKey.String()] = da.req.TemplateSource

	newConstants := make(map[string]string, 0)
	for k, v := range da.req.Constant {
		if v == "" {
			continue
		}
		newConstants[fmt.Sprintf("${%s}", k)] = v
	}
	constantsbyte, err := json.Marshal(&newConstants)
	if err != nil {
		blog.Errorf("generateBKopsStep failed: %v", err)
		return err
	}
	step.Params[cloudprovider.BkSopsConstantsKey.String()] = string(constantsbyte)

	task.Steps[stepName] = step
	task.StepSequence = append(task.StepSequence, stepName)

	return nil
}
