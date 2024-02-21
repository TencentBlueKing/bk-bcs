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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
)

// SkipAction task skip
type SkipAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.SkipTaskRequest
	resp  *cmproto.SkipTaskResponse

	task    *cmproto.Task
	cluster *cmproto.Cluster
}

// NewSkipAction create skip action
func NewSkipAction(model store.ClusterManagerModel) *SkipAction {
	return &SkipAction{
		model: model,
	}
}

func (sa *SkipAction) setResp(code uint32, msg string) {
	sa.resp.Code = code
	sa.resp.Message = msg
	sa.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (sa *SkipAction) getRelativeData() error {
	task, err := sa.model.GetTask(sa.ctx, sa.req.TaskID)
	if err != nil {
		blog.Errorf("RetryTaskAction Task %s failed when retry task, err %s", sa.req.TaskID, err.Error())
		return err
	}
	cluster, err := sa.model.GetCluster(sa.ctx, task.ClusterID)
	if err != nil {
		blog.Errorf("RetryTaskAction %s failed: %v", sa.req.TaskID, err)
		return err
	}

	sa.task = task
	sa.cluster = cluster

	return nil
}

func (sa *SkipAction) validate() error {
	if err := sa.req.Validate(); err != nil {
		return err
	}
	// check task status
	switch sa.task.Status {
	case cloudprovider.TaskStatusInit, cloudprovider.TaskStatusRunning,
		cloudprovider.TaskStatusSuccess, cloudprovider.TaskStatusPartFailure:
		errMsg := fmt.Errorf("task[%s] status[%s] doing or done when skip task", sa.task.TaskID, sa.task.Status)
		return errMsg
	case cloudprovider.TaskStatusFailure, cloudprovider.TaskStatusTimeout:
	}

	return sa.checkCurStepValidate()
}

func (sa *SkipAction) checkCurStepValidate() error {
	curStep := sa.task.GetCurrentStep()
	if curStep == "" {
		return fmt.Errorf("task[%s] current step empty", sa.task.TaskID)
	}
	steps := sa.task.GetSteps()
	_, ok := steps[curStep]
	if !ok {
		return fmt.Errorf("task[%s] curStep[%s] not exist steps", sa.task.TaskID, curStep)
	}

	return nil
}

func (sa *SkipAction) distributeTask() error {
	sa.task.Status = cloudprovider.TaskStatusRunning
	sa.task.Message = "task skiping"

	// update current step status SKIP
	sa.task.Steps[sa.task.GetCurrentStep()].Status = cloudprovider.TaskStatusSkip

	err := sa.model.UpdateTask(sa.ctx, sa.task)
	if err != nil {
		blog.Errorf("SkipTaskAction[%s] updateTask failed: %v", sa.cluster.ClusterID, err)
		return err
	}
	if err = taskserver.GetTaskServer().Dispatch(sa.task); err != nil {
		blog.Errorf("dispatch skip task[%s] for cluster %s failed, %s", sa.req.TaskID, sa.task.ClusterID, err.Error())
		return err
	}
	blog.Infof("skip cluster[%s] task[%s] type %s successfully", sa.task.ClusterID, sa.task.TaskID, sa.task.TaskType)

	utils.HandleTaskStepData(sa.ctx, sa.task)

	sa.resp.Data = sa.task
	return nil
}

// Handle handle skip task action
func (sa *SkipAction) Handle(
	ctx context.Context, req *cmproto.SkipTaskRequest, resp *cmproto.SkipTaskResponse) {

	if req == nil || resp == nil {
		blog.Errorf("skip task failed, req or resp is empty")
		return
	}
	sa.ctx = ctx
	sa.req = req
	sa.resp = resp

	if err := sa.getRelativeData(); err != nil {
		sa.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if err := sa.validate(); err != nil {
		sa.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := sa.distributeTask(); err != nil {
		sa.setResp(common.BcsErrClusterManagerTaskErr, err.Error())
		return
	}
	// handle cluster data status and not block task, finally task will update data status
	_ = updateTaskDataStatus(sa.model, sa.task)

	sa.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
