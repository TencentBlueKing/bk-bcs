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
 *
 */

package common

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/template"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"

	"github.com/avast/retry-go"
)

// RegisterCommonActions register common actions
func RegisterCommonActions() map[string]interface{} {
	return map[string]interface{}{
		cloudprovider.BKSOPTask:                  RunBKsopsJob,
		cloudprovider.UnCordonNodesAction:        UnCordonNodesTask,
		cloudprovider.CordonNodesAction:          CordonNodesTask,
		cloudprovider.WatchTask:                  EnsureWatchComponentTask,
		cloudprovider.InstallGseAgentAction:      InstallGSEAgentTask,
		cloudprovider.TransferHostModuleAction:   TransferHostModuleTask,
		cloudprovider.RemoveHostFromCmdbAction:   RemoveHostFromCMDBTask,
		cloudprovider.JobFastExecuteScriptAction: JobExecuteScriptTask,
		cloudprovider.EnsureAutoScalerAction:     EnsureAutoScalerTask,

		cloudprovider.InstallVclusterAction:      InstallVclusterTask,
		cloudprovider.DeleteVclusterAction:       UnInstallVclusterTask,
		cloudprovider.CreateNamespaceAction:      CreateNamespaceTask,
		cloudprovider.DeleteNamespaceAction:      DeleteNamespaceTask,
		cloudprovider.SetNodeLabelsAction:        SetNodeLabelsTask,
		cloudprovider.SetNodeAnnotationsAction:   SetNodeAnnotationsTask,
		cloudprovider.CheckKubeAgentStatusAction: CheckKubeAgentStatusTask,
		cloudprovider.CreateResourceQuotaAction:  CreateResourceQuotaTask,
		cloudprovider.DeleteResourceQuotaAction:  DeleteResourceQuotaTask,
		cloudprovider.ResourcePoolLabelAction:    SetResourcePoolDeviceLabels,
	}
}

const (
	// 默认bksops传入task_id参数
	injectTaskID = "${task_id}"
)

//* here are common tasks that for backgroup running
//* backgroup task running depends on machinery framework

// RunBKsopsJob running bksops job and wait for results
func RunBKsopsJob(taskID string, stepName string) error {
	// step1: get BKops url and para by taskID
	// step2: create bkops task
	// step3: start task & query status

	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("RunBKsopsJob[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("RunBKsopsJob[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// get bksops common parameter
	url := step.Params[cloudprovider.BkSopsUrlKey.String()]
	bizID := step.Params[cloudprovider.BkSopsBizIDKey.String()]
	templateID := step.Params[cloudprovider.BkSopsTemplateIDKey.String()]
	operator := step.Params[cloudprovider.BkSopsTemplateUserKey.String()]
	templateSource := step.Params[cloudprovider.BkSopsTemplateSourceKey.String()]
	constants := step.Params[cloudprovider.BkSopsConstantsKey.String()]
	taskName := state.Task.CommonParams[cloudprovider.TaskNameKey.String()]

	if bizID == "" || operator == "" || templateID == "" || taskName == "" || constants == "" {
		errMsg := fmt.Sprintf("RunBKsopsJob[%s] validateParameter task[%s] step[%s] failed", taskID, taskID, stepName)
		blog.Errorf(errMsg)
		retErr := fmt.Errorf("RunBKsopsJob err, %s", errMsg)
		if step.GetSkipOnFailed() {
			_ = state.SkipFailure(start, stepName, err)
			return nil
		}
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// render constants dynamic value parameter
	consMap, err := renderDynamicParaToConstants(state.Task, constants)
	if err != nil {
		errMsg := fmt.Sprintf("RunBKsopsJob[%s] unmarshal constants failed[%v]", taskID, err)
		blog.Errorf(errMsg)
		retErr := fmt.Errorf("RunBKsopsJob err, %s", errMsg)
		if step.GetSkipOnFailed() {
			_ = state.SkipFailure(start, stepName, err)
			return nil
		}
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	timeOutCtx, cancel := context.WithTimeout(ctx, time.Minute*120)
	defer cancel()

	taskUrl, err := execBkSopsTask(timeOutCtx, createBkSopsTaskParas{
		url:            url,
		bizID:          bizID,
		templateID:     templateID,
		operator:       operator,
		templateSource: templateSource,
		taskName:       taskName,
		constants:      consMap,
		stepName:       stepName,
	})
	if err != nil {
		state.TaskUrl = taskUrl
		if step.GetSkipOnFailed() {
			_ = state.SkipFailure(start, stepName, err)
			return nil
		}
		_ = state.UpdateStepFailure(start, stepName, err)
		return err
	}

	state.TaskUrl = taskUrl
	_ = state.UpdateStepSucc(start, stepName)
	return nil
}

// renderDynamicParaToConstants extract constants parameter & inject dynamic value
func renderDynamicParaToConstants(task *cmproto.Task, constants string) (map[string]string, error) {
	consMap := map[string]string{}
	err := json.Unmarshal([]byte(constants), &consMap)
	if err != nil {
		return nil, err
	}

	// inject dynamic parameter
	for ck, cv := range consMap {
		if value, ok := template.DynamicParameterInject[cv]; ok {
			consMap[ck] = task.CommonParams[value]
		}
	}

	// default bksops task set taskID para
	consMap[injectTaskID] = task.TaskID

	return consMap, nil
}

func execBkSopsTask(ctx context.Context, paras createBkSopsTaskParas) (string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	taskResp, err := createBkSopsTask(ctx, paras)
	if err != nil {
		blog.Errorf("execBkSopsTask[%s] createBkSopsTask failed: %v", taskID, err)
		return "", err
	}
	blog.Infof("execBkSopsTask[%s] createBkSopsTask successful: taskID[%v]", taskID, taskResp.TaskID)

	// update bksops taskUrl to task
	cloudprovider.SetTaskStepParas(taskID, paras.stepName, cloudprovider.BkSopsTaskUrlKey.String(), taskResp.TaskURL)

	startTaskReq := startBkSopsTaskParas{
		bizID:    paras.bizID,
		taskID:   taskResp.TaskID,
		operator: paras.operator,
	}
	err = startBkSopsTask(ctx, startTaskReq)
	if err != nil {
		blog.Errorf("execBkSopsTask[%s] startBkSopsTask failed: %v", taskID, err)
		return taskResp.TaskURL, err
	}
	blog.Infof("execBkSopsTask[%s] startBkSopsTask successful", taskID)

	getTaskStatusReq := &TaskPathParas{
		BkBizID:  paras.bizID,
		TaskID:   fmt.Sprintf("%d", taskResp.TaskID),
		Operator: paras.operator,
	}

	err = loop.LoopDoFunc(ctx, func() error {
		data, errGet := BKOpsClient.GetTaskStatus("", getTaskStatusReq, &StartTaskRequest{})
		if errGet != nil {
			blog.Errorf("RunBKsopsJob[%s] execBkSopsTask GetTaskStatus failed: %v", taskID, errGet)
			return nil
		}

		blog.Infof("RunBKsopsJob[%s] execBkSopsTask GetTaskStatus[%s] status %s",
			taskID, getTaskStatusReq.TaskID, data.Data.State)
		if data.Data.State == FINISHED.String() {
			return loop.EndLoop
		}

		if data.Data.State == FAILED.String() || data.Data.State == REVOKED.String() ||
			data.Data.State == SUSPENDED.String() {
			blog.Errorf("RunBKsopsJob[%s] execBkSopsTask GetTaskStatus[%s] failed: status[%s]",
				taskID, getTaskStatusReq.TaskID, data.Data.State)
			retErr := fmt.Errorf("execBkSopsTask GetTaskStatus %s %s err: %v, url: %s",
				getTaskStatusReq.TaskID, data.Data.State, err, taskResp.TaskURL)
			return retErr
		}

		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("RunBKsopsJob[%s] execBkSopsTask failed: %v", taskID, err)
		return taskResp.TaskURL, err
	}

	blog.Infof("RunBKsopsJob[%s] execBkSopsTask successful", taskID)
	return taskResp.TaskURL, nil
}

type createBkSopsTaskParas struct {
	url        string
	bizID      string
	templateID string
	operator   string

	templateSource string
	taskName       string
	constants      map[string]string

	stepName string
}

type startBkSopsTaskParas struct {
	bizID    string
	taskID   int
	operator string
}

// createBkSopsTask 从模板创建标准运维任务
func createBkSopsTask(ctx context.Context, paras createBkSopsTaskParas) (*ResData, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// create task
	pathParas := &CreateTaskPathParas{
		BkBizID:    paras.bizID,
		TemplateID: paras.templateID,
		Operator:   paras.operator,
	}
	createTaskReq := &CreateTaskRequest{
		TemplateSource: paras.templateSource,
		TaskName:       paras.taskName,
		Constants:      paras.constants,
	}

	var (
		resp *CreateTaskResponse
		err  error
	)

	err = retry.Do(func() error {
		resp, err = BKOpsClient.CreateBkOpsTask(paras.url, pathParas, createTaskReq)
		if err != nil {
			return err
		}
		return nil
	}, retry.Attempts(3))
	if err != nil {
		blog.Errorf("RunBKsopsJob[%s] createBkSopsTask failed: %v", taskID, err)
		return nil, err
	}
	blog.Infof("RunBKsopsJob[%s] createBkSopsTask successful[%d]", taskID, resp.Data.TaskID)

	return resp.Data, nil
}

// startBkSopsTask 启动标准运维任务
func startBkSopsTask(ctx context.Context, paras startBkSopsTaskParas) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// start task
	startTaskReq := &TaskPathParas{
		BkBizID:  paras.bizID,
		TaskID:   fmt.Sprintf("%d", paras.taskID),
		Operator: paras.operator,
	}

	var (
		err error
	)
	err = retry.Do(func() error {
		_, errStart := BKOpsClient.StartBkOpsTask("", startTaskReq, &StartTaskRequest{})
		if errStart != nil {
			return errStart
		}
		return nil
	}, retry.Attempts(3))
	if err != nil {
		blog.Errorf("RunBKsopsJob[%s] startBkSopsTask failed: %v", taskID, err)
		return err
	}

	blog.Infof("RunBKsopsJob[%s] startBkSopsTask successful[%d]", taskID, startTaskReq.TaskID)

	return nil
}
