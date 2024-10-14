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

// Package operationlog xxxx
package operationlog

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	autils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// TaskRecordsAction action for list operation logs
type TaskRecordsAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.ListTaskRecordsRequest
	resp  *cmproto.ListTaskRecordsResponse
}

var (
	statusMap = map[string]string{
		cloudprovider.TaskStatusInit:        "LOADING",
		cloudprovider.TaskStatusRunning:     "LOADING",
		cloudprovider.TaskStatusSuccess:     "SUCCESS",
		cloudprovider.TaskStatusPartFailure: "HALFSUCCESS",
		cloudprovider.TaskStatusFailure:     "FAILED",
		cloudprovider.TaskStatusNotStarted:  "WAITING",
		cloudprovider.TaskStatusTimeout:     "TERMINATE",
	}
)

// NewTaskRecordsAction create action
func NewTaskRecordsAction(model store.ClusterManagerModel) *TaskRecordsAction {
	return &TaskRecordsAction{
		model: model,
	}
}

func (ua *TaskRecordsAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	return nil
}

func (ua *TaskRecordsAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ua *TaskRecordsAction) fetchTaskRecords() error {
	task, err := ua.model.GetTask(ua.ctx, ua.req.TaskID)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}

	ua.resp.Data = &cmproto.TaskRecordsResponseData{
		Status: task.Status,
		Step:   []*cmproto.TaskRecordStep{},
	}

	allowRetry := true
	// attention: 开启CA节点自动扩缩容不允许手动重试
	if utils.StringContainInSlice(task.TaskType, []string{cloudprovider.UpdateNodeGroupDesiredNode.String(),
		cloudprovider.CleanNodeGroupNodes.String()}) &&
		task.GetCommonParams()[cloudprovider.ManualKey.String()] == common.False {
		allowRetry = false
	}

	// 默认 所有的失败任务都允许重试, 自动化任务不允许重试
	// 仅仅只有设置失败跳过标志的步骤才允许失败跳过
	for _, step := range task.StepSequence {
		for k, v := range task.Steps {
			if step == k {
				if status, ok := statusMap[v.Status]; ok {
					v.Status = status
				} else {
					v.Status = ""
				}

				ua.resp.Data.Step = append(ua.resp.Data.Step, &cmproto.TaskRecordStep{
					Name:       v.Name,
					Status:     v.Status,
					StartTime:  utils.TransStrToTs(v.Start),
					EndTime:    utils.TransStrToTs(v.End),
					Data:       []*cmproto.TaskRecordStepData{},
					AllowSkip:  v.AllowSkip,
					AllowRetry: allowRetry,
				})
			}
		}
	}

	return ua.appendTaskRecords(task)
}

func (ua *TaskRecordsAction) appendTaskRecords(task *cmproto.Task) error {
	// resource condition
	cond := operator.M{"taskid": task.TaskID}
	resourceCond := operator.NewLeafCondition(operator.Eq, cond)
	conds := []*operator.Condition{resourceCond}
	logsCond := operator.NewBranchCondition(operator.And, conds...)

	// list operation logs
	sort := map[string]int{"createtime": 1}
	logs, err := ua.model.ListTaskStepLog(ua.ctx, logsCond, &options.ListOption{Sort: sort})
	if err != nil {
		return err
	}

	for k, v := range ua.resp.Data.Step {
		for _, y := range logs {
			if y.StepName == v.Name {
				createTime := utils.TransStrToTs(y.CreateTime)
				ua.resp.Data.Step[k].Data = append(ua.resp.Data.Step[k].Data, &cmproto.TaskRecordStepData{
					Log:       y.Message,
					Timestamp: createTime,
					Level:     y.Level,
				})
			}
		}
	}

	for k, v := range ua.resp.Data.Step {
		if step, ok := task.Steps[v.Name]; ok {
			ua.resp.Data.Step[k].Name = autils.Translate(ua.ctx,
				step.TaskMethod, step.TaskName, step.Translate)
		}
	}

	return nil
}

// Handle handles task records
func (ua *TaskRecordsAction) Handle(
	ctx context.Context, req *cmproto.ListTaskRecordsRequest, resp *cmproto.ListTaskRecordsResponse) {
	if req == nil || resp == nil {
		blog.Errorf("task records failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	err := ua.validate()
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	err = ua.fetchTaskRecords()
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
