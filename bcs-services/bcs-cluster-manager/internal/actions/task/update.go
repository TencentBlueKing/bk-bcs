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

package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
)

// UpdateAction update action for online cluster credential
type UpdateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.UpdateTaskRequest
	resp  *cmproto.UpdateTaskResponse
}

// NewUpdateAction create update action for online cluster credential
func NewUpdateAction(model store.ClusterManagerModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

func (ua *UpdateAction) updateTask(tsk *cmproto.Task) error {
	timeStr := time.Now().Format(time.RFC3339)
	// update field if required
	tsk.LastUpdate = timeStr
	tsk.Updater = ua.req.Updater
	if len(ua.req.Status) != 0 {
		tsk.Status = ua.req.Status
	}
	if len(ua.req.Message) != 0 {
		tsk.Message = ua.req.Message
	}
	if len(ua.req.CurrentStep) != 0 {
		tsk.CurrentStep = ua.req.CurrentStep
	}
	if ua.req.Steps != nil {
		tsk.Steps = ua.req.Steps
	}
	if len(ua.req.End) != 0 {
		tsk.End = ua.req.End
	}
	if ua.req.ExecutionTime > 0 {
		tsk.ExecutionTime = ua.req.ExecutionTime
	}
	if ua.req.ExecutionTime > 0 {
		tsk.ExecutionTime = ua.req.ExecutionTime
	}
	return ua.model.UpdateTask(ua.ctx, tsk)
}

func (ua *UpdateAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handle update cluster credential
func (ua *UpdateAction) Handle(
	ctx context.Context, req *cmproto.UpdateTaskRequest, resp *cmproto.UpdateTaskResponse) {

	if req == nil || resp == nil {
		blog.Errorf("update task failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := req.Validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// get old Task information, update fields if required
	destTsk, err := ua.model.GetTask(ua.ctx, req.TaskID)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("find Task %s failed when pre-update checking, err %s", req.TaskID, err.Error())
		return
	}
	if err := ua.updateTask(destTsk); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ua.resp.Data = destTsk
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// RetryAction update action for cluster task retry
type RetryAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.RetryTaskRequest
	resp  *cmproto.RetryTaskResponse

	task    *cmproto.Task
	cluster *cmproto.Cluster

	isPartFailure bool
}

// NewRetryAction create retry action for cluster retry
func NewRetryAction(model store.ClusterManagerModel) *RetryAction {
	return &RetryAction{
		model: model,
	}
}

func (ua *RetryAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ua *RetryAction) getRelativeData() error {
	task, err := ua.model.GetTask(ua.ctx, ua.req.TaskID)
	if err != nil {
		blog.Errorf("RetryTaskAction Task %s failed when retry task, err %s", ua.req.TaskID, err.Error())
		return err
	}
	cluster, err := ua.model.GetCluster(ua.ctx, task.ClusterID)
	if err != nil {
		blog.Errorf("RetryTaskAction %s failed: %v", ua.req.TaskID, err)
		return err
	}

	ua.task = task
	ua.cluster = cluster

	return nil
}

// ua.setResp(common.BcsErrClusterManagerTaskErr, errMsg)
func (ua *RetryAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	retry := allowTaskRetry(ua.task)
	if !retry {
		return fmt.Errorf("autoscaler task can't manual retry")
	}

	// check task status
	switch ua.task.Status {
	case cloudprovider.TaskStatusInit, cloudprovider.TaskStatusRunning:
		errMsg := fmt.Errorf("task[%s] status[%s] doing when retry task", ua.task.TaskID, ua.task.Status)
		return errMsg
	case cloudprovider.TaskStatusSuccess:
		if utils.CheckTaskStepPartFailureStatus(ua.task) &&
			strings.Contains(ua.task.TaskType, cloudprovider.AddNodesToCluster.String()) {
			ua.isPartFailure = true
			return nil
		}
		errMsg := fmt.Errorf("task[%s] status[%s] done when retry task", ua.task.TaskID, ua.task.Status)
		return errMsg
	case cloudprovider.TaskStatusFailure, cloudprovider.TaskStatusTimeout:
	}

	return nil
}

func (ua *RetryAction) distributeTask() error {
	// status is part failure && whole task is done will create new task
	if ua.isPartFailure {
		newTask, err := retryPartFailureTask(ua.model, ua.cluster, ua.task)
		if err != nil {
			return nil
		}
		ua.task = newTask
		// create task
		if err := ua.model.CreateTask(ua.ctx, ua.task); err != nil {
			blog.Errorf("save addNodesToCluster cluster task for cluster %s failed, %s",
				ua.cluster.ClusterID, err.Error(),
			)
			return err
		}
	} else {
		ua.task.Status = cloudprovider.TaskStatusRunning
		ua.task.Message = "task retrying"
		step, ok := ua.task.Steps[ua.task.CurrentStep]
		if ok && step.MaxRetry > 0 {
			step.Retry = 0
		}

		err := ua.model.UpdateTask(ua.ctx, ua.task)
		if err != nil {
			blog.Errorf("RetryTaskAction[%s] updateTask failed: %v", ua.cluster.ClusterID, err)
			return err
		}
	}

	if err := taskserver.GetTaskServer().Dispatch(ua.task); err != nil {
		blog.Errorf("dispatch retry task[%s] for cluster %s failed, %s", ua.req.TaskID, ua.task.ClusterID, err.Error())
		return err
	}
	blog.Infof("retry cluster[%s] task[%s] type %s successfully", ua.task.ClusterID, ua.task.TaskID, ua.task.TaskType)

	utils.HandleTaskStepData(ua.ctx, ua.task)

	ua.resp.Data = ua.task
	return nil
}

// Handle handle update cluster credential
func (ua *RetryAction) Handle(
	ctx context.Context, req *cmproto.RetryTaskRequest, resp *cmproto.RetryTaskResponse) {

	if req == nil || resp == nil {
		blog.Errorf("retry task failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.getRelativeData(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ua.distributeTask(); err != nil {
		ua.setResp(common.BcsErrClusterManagerTaskErr, err.Error())
		return
	}
	// handle cluster data status and not block task, finally task will update data status
	if !ua.isPartFailure {
		_ = updateTaskDataStatus(ua.model, ua.task)
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
