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
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/template"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/nodeman"
)

// RegisterCommonActions register common actions
func RegisterCommonActions() map[string]interface{} {
	return map[string]interface{}{
		cloudprovider.BKSOPTask:              RunBKsopsJob,
		cloudprovider.WatchTask:              EnsureWatchComponentTask,
		cloudprovider.EnsureAutoScalerAction: EnsureAutoScalerTask,
	}
}

// * here are common tasks that for backgroup running
// * backgroup task running depends on machinery framework

// RunBKsopsJob running bksops job and wait for results
func RunBKsopsJob(taskID string, stepName string) error {
	// step1: get BKops url and para by taskID
	// step2: create bkops task
	// step3: start task & query status

	start := time.Now()

	// get task form database
	task, err := cloudprovider.GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("RunBKsopsJob[%s] task %s get detail task information from storage failed: %s, task retry",
			taskID, taskID, err.Error())
		return err
	}

	// task state check
	state := &cloudprovider.TaskState{
		Task:      task,
		JobResult: cloudprovider.NewJobSyncResult(task),
	}
	// check task already terminated
	if state.IsTerminated() {
		blog.Errorf("RunBKsopsJob[%s] task %s is terminated, step %s skip", taskID, taskID, stepName)
		return fmt.Errorf("task %s terminated", taskID)
	}
	// workflow switch current step to stepName when previous task exec successful
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("RunBKsopsJob[%s] task %s not turn ro run step %s, err %s", taskID, taskID, stepName, err.Error())
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("RunBKsopsJob[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}

	blog.Infof("RunBKsopsJob[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// get common parameter
	url := step.Params["url"]
	bizID := step.Params["template_biz_id"]
	templateID := step.Params["template_id"]
	operator := step.Params["template_user"]
	constants := step.Params["constants"]

	taskName := task.CommonParams["taskName"]

	if url == "" || bizID == "" || operator == "" || templateID == "" || taskName == "" || constants == "" {
		errMsg := fmt.Sprintf("RunBKsopsJob[%s] validateParameter task[%s] step[%s] failed", taskID, taskID, stepName)
		blog.Errorf(errMsg)
		retErr := fmt.Errorf("RunBKsopsJob err, %s", errMsg)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// extract constants parameter & inject dynamic value
	consMap := map[string]string{}
	err = json.Unmarshal([]byte(constants), &consMap)
	if err != nil {
		errMsg := fmt.Sprintf("RunBKsopsJob[%s] unmarshal constants failed[%v]", taskID, err)
		blog.Errorf(errMsg)

		retErr := fmt.Errorf("RunBKsopsJob err, %s", errMsg)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject dynamic parameter
	for ck, cv := range consMap {
		if value, ok := template.DynamicParameterInject[cv]; ok {
			consMap[ck] = task.CommonParams[value]
		}
	}

	// create task
	pathParas := &CreateTaskPathParas{
		BkBizID:    bizID,
		TemplateID: templateID,
		Operator:   operator,
	}
	createTaskReq := &CreateTaskRequest{
		TaskName:  taskName,
		Constants: consMap,
	}
	taskRes, err := BKOpsClient.CreateBkOpsTask(url, pathParas, createTaskReq)
	if err != nil {
		blog.Errorf("RunBKsopsJob[%s] CreateBkOpsTask task[%s] step[%s] failed; %v",
			taskID, task.TaskName, stepName, err)
		retErr := fmt.Errorf("CreateBkOpsTask err: %v", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("RunBKsopsJob[%s] createBkSopsTask successful: taskID[%v]", taskID, taskRes.Data.TaskID)

	// update bksops taskUrl to task
	newTask, err := cloudprovider.SetTaskStepParas(taskID, stepName, cloudprovider.BkSopsTaskUrlKey.String(), taskRes.Data.TaskURL)
	if err == nil {
		state.Task = newTask
	}

	// start task
	startTaskReq := &TaskPathParas{
		BkBizID:  bizID,
		TaskID:   fmt.Sprintf("%d", taskRes.Data.TaskID),
		Operator: operator,
	}
	_, err = BKOpsClient.StartBkOpsTask("", startTaskReq, &StartTaskRequest{})
	if err != nil {
		blog.Errorf("RunBKsopsJob[%s] StartBkOpsTask task[%s] step[%s] failed; %v", taskID, taskID, stepName, err)
		retErr := fmt.Errorf("StartBkOpsTask err: %s, url: %s", err.Error(), taskRes.Data.TaskURL)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
	defer cancel()

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			blog.Errorf("RunBKsopsJob[%s] GetTaskStatus task[%s] step[%s] failed; %v", taskID, taskID, stepName, ctx.Err())
			retErr := fmt.Errorf("GetTaskStatus %s %s err: %s, url: %s", startTaskReq.TaskID, "timeOut", ctx.Err(),
				taskRes.Data.TaskURL)
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return nil
		case <-ticker.C:
		}

		data, err := BKOpsClient.GetTaskStatus("", startTaskReq, &StartTaskRequest{})
		if err != nil {
			blog.Errorf("RunBKsopsJob[%s] GetTaskStatus failed: %v", taskID, err)
			continue
		}

		blog.Infof("RunBKsopsJob[%s] GetTaskStatus %s status %s", taskID, startTaskReq.TaskID, data.Data.State)
		if data.Data.State == FINISHED.String() {
			// update step
			_ = state.UpdateStepSucc(start, stepName)
			break
		}
		if data.Data.State == FAILED.String() || data.Data.State == REVOKED.String() ||
			data.Data.State == SUSPENDED.String() {
			blog.Errorf("RunBKsopsJob[%s] GetTaskStatus task[%s] step[%s] failed: %v", taskID, taskID, stepName, err)
			retErr := fmt.Errorf("GetTaskStatus %s %s err: %v, url: %s", startTaskReq.TaskID, data.Data.State, err,
				taskRes.Data.TaskURL)
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
	}

	return nil
}

// InstallGSEAgentTask install gse agent task
func InstallGSEAgentTask(taskID string, stepName string) error {
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// get bkBizID
	bkBizIDString := step.Params[cloudprovider.BKBizIDKey.String()]
	// get bkCloudID
	bkCloudIDstring := step.Params[cloudprovider.BKCloudIDKey.String()]
	// get nodeIPs
	nodeIPs := state.Task.CommonParams[cloudprovider.NodeIPsKey.String()]
	// get password
	passwd := step.Params[cloudprovider.PasswordKey.String()]
	// get user
	user := step.Params[cloudprovider.UsernameKey.String()]
	if len(user) == 0 {
		user = nodeman.RootAccount
	}

	if len(nodeIPs) == 0 {
		blog.Infof("InstallGSEAgentTask %s skip, cause of empty node", taskID)
		_ = state.SkipFailure(start, stepName, fmt.Errorf("empty node ip"))
		return nil
	}

	bkCloudID, err := strconv.Atoi(bkCloudIDstring)
	if err != nil {
		blog.Errorf("InstallGSEAgentTask %s failed, invalid bkCloudID, err %s", taskID, err.Error())
		_ = state.SkipFailure(start, stepName, fmt.Errorf("invalid bkCloudID, err %s", err.Error()))
		return nil
	}
	bkBizID, err := strconv.Atoi(bkBizIDString)
	if err != nil {
		blog.Errorf("InstallGSEAgentTask %s failed, invalid bkBizID, err %s", taskID, err.Error())
		_ = state.SkipFailure(start, stepName, fmt.Errorf("invalid bkBizID, err %s", err.Error()))
		return nil
	}

	nodeManClient := nodeman.GetNodeManClient()
	if nodeManClient == nil {
		blog.Errorf("nodeman client is not init")
		_ = state.SkipFailure(start, stepName, fmt.Errorf("nodeman client is not init"))
		return nil
	}

	// get apID from cloud list
	clouds, err := nodeManClient.CloudList()
	if err != nil {
		blog.Errorf("InstallGSEAgentTask %s get cloud list error, %s", taskID, err.Error())
		_ = state.SkipFailure(start, stepName, fmt.Errorf("get cloud list error, %s", err.Error()))
		return nil
	}
	apID := getAPID(bkCloudID, clouds)

	// install gse agent
	hosts := make([]nodeman.JobInstallHost, 0)
	ips := strings.Split(nodeIPs, ",")
	for _, v := range ips {
		hosts = append(hosts, nodeman.JobInstallHost{
			BKCloudID: bkCloudID,
			APID:      apID,
			BKBizID:   bkBizID,
			OSType:    nodeman.LinuxOSType,
			InnerIP:   v,
			LoginIP:   v,
			Account:   user,
			Port:      nodeman.DefaultPort,
			AuthType:  nodeman.PasswordAuthType,
			Password:  passwd,
		})
	}
	job, err := nodeManClient.JobInstall(nodeman.InstallAgentJob, hosts)
	if err != nil {
		blog.Errorf("InstallGSEAgentTask %s install gse agent job error, %s", taskID, err.Error())
		_ = state.SkipFailure(start, stepName, fmt.Errorf("install gse agent job error, %s", err.Error()))
		return nil
	}
	blog.Infof("InstallGSEAgentTask %s install gse agent job(%d) url %s", taskID, job.JobID, job.JobURL)

	// check status
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Minute)
	defer cancel()
	err = cloudprovider.LoopDoFunc(ctx, func() error {
		var err error
		detail, err := nodeManClient.JobDetails(job.JobID)
		if err != nil {
			blog.Errorf("InstallGSEAgentTask %s failed, get job detail err %s", taskID, err.Error())
			return err
		}
		switch detail.Status {
		case nodeman.JobRunning:
			blog.Infof("InstallGSEAgentTask %s checking job status, waiting", taskID)
			return nil
		case nodeman.JobSuccess:
			return cloudprovider.EndLoop
		case nodeman.JobFailed, nodeman.JobPartFailed:
			return fmt.Errorf("GSE Agent 安装失败，详情查看: %s", job.JobURL)
		}
		return nil
	}, cloudprovider.LoopInterval(5*time.Second))
	if err != nil {
		blog.Errorf("InstallGSEAgentTask %s check gse agent install job status failed: %v", taskID, err)
		_ = state.SkipFailure(start, stepName, fmt.Errorf("check gse agent install job status err: %s", err.Error()))
		return nil
	}

	// update step
	_ = state.UpdateStepSucc(start, stepName)

	return nil
}

func getAPID(bkCloudID int, clouds []nodeman.CloudListData) int {
	apID := nodeman.DefaultAPID
	for _, v := range clouds {
		if v.BKCloudID == 0 {
			continue
		}
		if v.BKCloudID == bkCloudID {
			apID = v.APID
			break
		}
	}
	return apID
}

// TransferHostModule transfer host module task
func TransferHostModule(taskID string, stepName string) error {
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// get bkBizID
	bkBizIDString := step.Params[cloudprovider.BKBizIDKey.String()]
	// get nodeIPs
	nodeIPs := state.Task.CommonParams[cloudprovider.NodeIPsKey.String()]
	// get moduleID
	moduleIDString := step.Params[cloudprovider.BKModuleIDKey.String()]

	if len(nodeIPs) == 0 {
		blog.Warnf("TransferHostModule %s skip, cause of empty node", taskID)
		_ = state.SkipFailure(start, stepName, fmt.Errorf("empty node ip"))
		return nil
	}

	bkBizID, err := strconv.Atoi(bkBizIDString)
	if err != nil {
		blog.Errorf("TransferHostModule %s failed, invalid bkBizID, err %s", taskID, err.Error())
		_ = state.SkipFailure(start, stepName, fmt.Errorf("invalid bkBizID, err %s", err.Error()))
		return nil
	}
	moduleID, err := strconv.Atoi(moduleIDString)
	if err != nil {
		blog.Errorf("TransferHostModule %s failed, invalid moduleID, err %s", taskID, err.Error())
		_ = state.SkipFailure(start, stepName, fmt.Errorf("invalid moduleID, err %s", err.Error()))
		return nil
	}

	nodeManClient := nodeman.GetNodeManClient()
	if nodeManClient == nil {
		blog.Errorf("TransferHostModule %s failed, nodeman client is not init", taskID)
		_ = state.SkipFailure(start, stepName, fmt.Errorf("nodeman client is not init"))
		return nil
	}
	cmdbClient := cmdb.GetCmdbClient()
	if cmdbClient == nil {
		blog.Errorf("TransferHostModule %s failed, cmdb client is not init", taskID)
		_ = state.SkipFailure(start, stepName, fmt.Errorf("cmdb client is not init"))
		return nil
	}

	// get host id from host list
	ips := strings.Split(nodeIPs, ",")
	var hostIDs []int
	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Minute)
	defer cancel()
	err = cloudprovider.LoopDoFunc(ctx, func() error {
		var err error
		hostIDs, err = nodeManClient.GetHostIDByIPs(bkBizID, ips)
		if err != nil {
			blog.Errorf("TransferHostModule %s failed, list nodeman hosts err %s", err.Error())
			return err
		}
		if len(hostIDs) == len(ips) {
			return cloudprovider.EndLoop
		}
		blog.Infof("TransferHostModule %s can't get all host id, waiting", taskID)
		return nil
	}, cloudprovider.LoopInterval(3*time.Second))
	if err != nil {
		blog.Errorf("TransferHostModule %s get host id failed: %v", taskID, err)
		_ = state.SkipFailure(start, stepName, fmt.Errorf("get host id err %s", err.Error()))
		return nil
	}

	if err := cmdbClient.TransferHostModule(bkBizID, hostIDs, []int{moduleID}, false); err != nil {
		blog.Errorf("TransferHostModule %s failed, bkBizID %d, hosts %v, err %s",
			taskID, bkBizID, hostIDs, err.Error())
		_ = state.SkipFailure(start, stepName,
			fmt.Errorf("TransferHostModule failed, bkBizID %d, hosts %v, err %s", bkBizID, hostIDs, err.Error()))
		return nil
	}

	blog.Infof("TransferHostModule %s sucessful", taskID)

	// update step
	_ = state.UpdateStepSucc(start, stepName)

	return nil
}

// RemoveHostFromCMDBTask remove host from cmdb task
func RemoveHostFromCMDBTask(taskID string, stepName string) error {
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}
	nodeManClient := nodeman.GetNodeManClient()
	if nodeManClient == nil {
		blog.Errorf("RemoveHostFromCMDBTask %s failed, nodeman client is not init", taskID)
		_ = state.SkipFailure(start, stepName, fmt.Errorf("nodeman client is not init"))
		return nil
	}
	cmdbClient := cmdb.GetCmdbClient()
	if cmdbClient == nil {
		blog.Errorf("RemoveHostFromCMDBTask %s failed, cmdb client is not init", taskID)
		_ = state.SkipFailure(start, stepName, fmt.Errorf("cmdb client is not init"))
		return nil
	}

	// get bkBizID
	bkBizIDString := step.Params[cloudprovider.BKBizIDKey.String()]
	// get nodeIPs
	nodeIPs := state.Task.CommonParams[cloudprovider.NodeIPsKey.String()]

	if len(nodeIPs) == 0 {
		blog.Infof("RemoveHostFromCMDBTask %s skip, cause of empty node", taskID)
		_ = state.SkipFailure(start, stepName, fmt.Errorf("empty node ip"))
		return nil
	}
	bkBizID, err := strconv.Atoi(bkBizIDString)
	if err != nil {
		blog.Errorf("RemoveHostFromCMDBTask %s failed, invalid bkBizID, err %s", taskID, err.Error())
		_ = state.SkipFailure(start, stepName, fmt.Errorf("invalid bkBizID, err %s", err.Error()))
		return nil
	}

	// get host id from host list
	ips := strings.Split(nodeIPs, ",")
	hostIDs, err := nodeManClient.GetHostIDByIPs(bkBizID, ips)
	if err != nil {
		blog.Errorf("RemoveHostFromCMDBTask %s failed, list nodeman hosts err %s", err.Error())
		_ = state.SkipFailure(start, stepName, fmt.Errorf("list nodeman hosts err %s", err.Error()))
		return nil
	}

	if len(hostIDs) == 0 {
		blog.Warnf("RemoveHostFromCMDBTask %s skip, cause of empty host", taskID)
		_ = state.SkipFailure(start, stepName, fmt.Errorf("empty host"))
		return nil
	}

	if err := cmdbClient.TransferHostToIdleModule(bkBizID, hostIDs); err != nil {
		blog.Errorf("RemoveHostFromCMDBTask %s TransferHostToIdleModule failed, bkBizID %d, hosts %v, err %s",
			taskID, bkBizID, hostIDs, err.Error())
		_ = state.SkipFailure(start, stepName,
			fmt.Errorf("TransferHostToIdleModule failed, bkBizID %d, hosts %v, err %s", bkBizID, hostIDs, err.Error()))
		return nil
	}

	if err := cmdbClient.TransferHostToResourceModule(bkBizID, hostIDs); err != nil {
		blog.Errorf("RemoveHostFromCMDBTask %s TransferHostToResourceModule failed, bkBizID %d, hosts %v, err %s",
			taskID, bkBizID, hostIDs, err.Error())
		_ = state.SkipFailure(start, stepName,
			fmt.Errorf("TransferHostToResourceModule failed, bkBizID %d, hosts %v, err %s",
				bkBizID, hostIDs, err.Error()))
		return nil
	}

	if err := cmdbClient.DeleteHost(hostIDs); err != nil {
		blog.Errorf("RemoveHostFromCMDBTask %s DeleteHost %v failed, %s", taskID, hostIDs, err.Error())
		_ = state.SkipFailure(start, stepName, fmt.Errorf("DeleteHost %v failed, %s", hostIDs, err.Error()))
		return nil
	}
	blog.Infof("RemoveHostFromCMDBTask %s sucessful", taskID)

	// update step
	_ = state.UpdateStepSucc(start, stepName)
	return nil
}
