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

package taskhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/bksops"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/job"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/storage"
	powertrading "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/proto"
)

// OpsConf config
type OpsConf struct {
	TemplateId   string
	TemplateName string
}

func createAndStartBksOps(ctx context.Context, cli bksops.Client, storageCli storage.Storage, task *storage.MachineTask,
	constants map[string]string, conf *OpsConf) error {
	createRsp, err := cli.CreateTask(conf.TemplateId, task.BusinessID,
		conf.TemplateName, constants)
	if err != nil {
		return fmt.Errorf("create %s task error:%s", conf.TemplateId, err.Error())
	}
	blog.Infof("taskID: %d", createRsp.Data.TaskID)
	task.Detail[task.CurrentStep].BksOpsTaskID = strconv.Itoa(int(createRsp.Data.TaskID))
	blog.Infof("task taskid:%s", task.Detail[task.CurrentStep].BksOpsTaskID)
	_, err = storageCli.UpdateTask(ctx, task, &storage.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("update task %s error:%s", task.TaskID, err.Error())
	}
	_, startErr := cli.StartTask(task.Detail[task.CurrentStep].BksOpsTaskID, task.BusinessID)
	if startErr != nil {
		return fmt.Errorf("start %s task error:%s", conf.TemplateId, startErr.Error())
	}
	return nil
}

// CheckJobStatus check job status
func CheckJobStatus(ctx context.Context, opsCli bksops.Client, jobCli job.Client, storageCli storage.Storage,
	task *storage.MachineTask, stdOutPut bool, nodeName string, lastStep bool) (bool, bool, error) {
	blog.Infof("begin to check job status")
	jobId := task.Detail[task.CurrentStep].JobID
	if jobId == "" {
		id, err := getJobID(opsCli, task, nodeName)
		if err != nil {
			return false, false, err
		}
		task.Detail[task.CurrentStep].JobID = id
		_, err = storageCli.UpdateTask(ctx, task, &storage.UpdateOptions{})
		if err != nil {
			return false, false, fmt.Errorf("update task %s error:%s", task.TaskID, err.Error())
		}
	}
	task.Detail[task.CurrentStep].JobID = jobId
	blog.Infof("job id %s", jobId)
	jobStatus, err := jobCli.GetJobStatus("biz", jobId, task.BusinessID)
	if err != nil {
		return false, false, fmt.Errorf("%s GetJobStatus error:%s", jobId, err.Error())
	}
	if !jobStatus.Data.Finished {
		return false, false, nil
	}
	var pass bool
	if stdOutPut {
		pass, err = getJobResultWithMessage(jobCli, task, lastStep, jobId, jobStatus)
		if err != nil {
			blog.Errorf("getStandardJobResult failed:%s", err.Error())
			return true, pass, err
		}
	} else {
		pass = getJobResult(task, lastStep, jobStatus)
	}
	return true, pass, nil
}

// BksopsCheck bksops check task
func BksopsCheck(ctx context.Context, opsCli bksops.Client, jobCli job.Client, storageCli storage.Storage,
	task *storage.MachineTask, conf *OpsConf) {
	remainIPs := getRemainIPs(task.IPList, task.Summary[storage.MachineCheckFailure])
	task.Detail[storage.BkOpsTaskCheck].IPList = remainIPs
	if task.Detail[storage.BkOpsTaskCheck] == nil || task.Detail[storage.BkOpsTaskCheck].BksOpsTaskID == "" {
		constants := make(map[string]string)
		ipStr := ""
		for key, ip := range remainIPs {
			ipStr += ip
			if key != len(remainIPs)-1 {
				ipStr += ","
			}
		}
		constants["${node_ip_list}"] = ipStr
		constants["${biz_cc_id}"] = task.BusinessID
		blog.Infof("req constant:%v", constants)
		err := createAndStartBksOps(ctx, opsCli, storageCli, task, constants, conf)
		if err != nil {
			blog.Errorf("%s task check error:%s", task.TaskID, err.Error())
			task.RetryTimes++
			if task.RetryTimes == 3 {
				task.Message = fmt.Sprintf("retry task %s 3 times, finished this task, error:%s",
					task.TaskID, err.Error())
				blog.Errorf("retry task %s 3 times, finished this task", task.TaskID)
				task.Detail[storage.BkOpsTaskCheck].Status = storage.TaskFailed
				task.Status = storage.TaskFailed
			}
			_, err = storageCli.UpdateTask(ctx, task, &storage.UpdateOptions{})
			if err != nil {
				blog.Errorf("update task %s error:%s", task.TaskID, err.Error())
			}
		}
		return
	}
	finished, _, checkErr := CheckJobStatus(ctx, opsCli, jobCli, storageCli, task, true, "执行检测", true)
	if checkErr != nil {
		blog.Errorf("%s task check job status error:%s", task.TaskID, checkErr.Error())
		task.RetryTimes++
		if task.RetryTimes == 4 {
			blog.Errorf("retry task %s 3 times, finished this task", task.TaskID)
			task.Detail[storage.BkOpsTaskCheck].Status = storage.TaskFailed
			task.Detail[storage.BkOpsTaskCheck].Message = fmt.Sprintf("retry task %s 3 times, finished this task, "+
				"error:%s", task.TaskID, checkErr.Error())
			task.Status = storage.TaskFailed
		}
		_, err := storageCli.UpdateTask(ctx, task, &storage.UpdateOptions{})
		if err != nil {
			blog.Errorf("update task %s error:%s", task.TaskID, err.Error())
		}
		return
	}
	if !finished {
		return
	}
	task.Detail[storage.BkOpsTaskCheck].Status = storage.TaskFinished
	task.Status = storage.TaskFinished
	_, err := storageCli.UpdateTask(ctx, task, &storage.UpdateOptions{})
	if err != nil {
		blog.Errorf("update task %s error:%s", task.TaskID, err.Error())
	}
}

func getDefaultDateRange() []string {
	dates := make([]string, 0)
	for i := 0; i < memoryCheckDays; i++ {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		dates = append(dates, date)
	}
	return dates
}

func getRemainIPs(originIPs []string, failureIPs map[string][]string) []string {
	originIPMap := make(map[string]bool)
	for _, originIP := range originIPs {
		originIPMap[originIP] = true
	}
	if failureIPs == nil {
		return originIPs
	}
	for reason := range failureIPs {
		for _, failureIP := range failureIPs[reason] {
			originIPMap[failureIP] = false
		}
	}
	remainIPs := make([]string, 0)
	for ip := range originIPMap {
		if originIPMap[ip] {
			remainIPs = append(remainIPs, ip)
		}
	}
	return remainIPs
}

func checkIfContinue(originIPs []string, failureIPs map[string][]string) bool {
	machineNum := len(originIPs)
	failureNum := 0
	for reason := range failureIPs {
		failureNum += len(failureIPs[reason])
	}
	return failureNum < machineNum
}

func getJobID(opsCli bksops.Client, task *storage.MachineTask, nodeName string) (string, error) {
	var jobId string
	taskStatus, err := opsCli.GetTaskStatus(task.Detail[task.CurrentStep].BksOpsTaskID, task.BusinessID)
	if err != nil {
		return jobId, fmt.Errorf("GetTaskStatus %s error:%s", task.Detail[task.CurrentStep].BksOpsTaskID, err.Error())
	}
	nodeId := ""
	for node := range taskStatus.Data.Children {
		if taskStatus.Data.Children[node].Name == nodeName {
			nodeId = node
			break
		}
	}
	nodeDetail, err := opsCli.GetTaskNodeDetail(task.Detail[task.CurrentStep].BksOpsTaskID,
		task.BusinessID, nodeId)
	if err != nil {
		return jobId, fmt.Errorf("GetTaskNodeDetail %s error:%s",
			task.Detail[task.CurrentStep].BksOpsTaskID, err.Error())
	}
	for _, output := range nodeDetail.Data.Outputs {
		if output.Key == "job_inst_id" {
			switch x := output.Value.(type) {
			case int:
				id, _ := output.Value.(int)
				jobId = strconv.Itoa(id)
			case string:
				id, _ := output.Value.(string)
				jobId = id
			case float64:
				id, _ := output.Value.(float64)
				jobId = strconv.FormatFloat(id, 'f', -1, 64)
			default:
				blog.Errorf("not supported type:%s", x)
			}
		}
	}
	if jobId == "" {
		return jobId, fmt.Errorf("get jobId failed")
	}
	return jobId, nil
}

// NOCC:golint/funlen(设计如此)
// nolint
func getJobResultWithMessage(jobCli job.Client, task *storage.MachineTask, lastStep bool, jobId string,
	jobStatus *job.StatusResponse) (bool, error) {
	task.Summary[storage.MachineCheckFailure][task.CurrentStep] = make([]string, 0)
	task.Summary[storage.MachineCheckSuccess][task.CurrentStep] = make([]string, 0)
	task.Summary[storage.MachineNeedClean][task.CurrentStep] = make([]string, 0)
	stepInstanceId := jobStatus.Data.StepInstanceList[0].StepInstanceID
	blog.Infof("stepInstanceId %d", stepInstanceId)
	batch := len(jobStatus.Data.StepInstanceList[0].StepIPResultList) / 20
	for i := 0; i <= batch; i++ {
		ips := make([]job.BatchLogIPRequest, 0)
		infos := make([]job.StepIPResult, 0)
		if i != batch {
			infos = jobStatus.Data.StepInstanceList[0].StepIPResultList[i*20 : (i+1)*20]
		} else {
			infos = jobStatus.Data.StepInstanceList[0].StepIPResultList[i*20:]
		}
		for _, info := range infos {
			ips = append(ips, job.BatchLogIPRequest{
				BkCloudID: info.BkCloudID,
				IP:        info.IP,
			})
		}
		batchRsp, batchErr := jobCli.GetBatchJobLog("biz", jobId, task.BusinessID,
			strconv.Itoa(int(stepInstanceId)), ips)
		if batchErr != nil {
			return false, fmt.Errorf("GetBatchJobLog error: %s", batchErr.Error())
		}
		blog.Infof("requestid:%s, log length:%s", batchRsp.JobRequestID, len(batchRsp.Data.ScriptTaskLogs))
		returnIPs := make(map[string]bool)
		for _, log := range batchRsp.Data.ScriptTaskLogs {
			returnIPs[log.IP] = true
			if task.Detail[task.CurrentStep].DetailList[log.IP] == nil {
				task.Detail[task.CurrentStep].DetailList[log.IP] = &powertrading.MachineTestMessage{
					Ip:          log.IP,
					Pass:        "false",
					AbleToClean: "false",
					Message:     "",
				}
			}
			finalMessageTag := "[final_message]"
			startIndex := strings.Index(log.LogContent, finalMessageTag)

			if startIndex == -1 {
				task.Detail[task.CurrentStep].DetailList[log.IP].Message = fmt.Sprintf("final_message tag not found")
				task.Summary[storage.MachineCheckFailure][task.CurrentStep] =
					append(task.Summary[storage.MachineCheckFailure][task.CurrentStep], log.IP)
				continue
			}

			startIndex += len(finalMessageTag)
			finalMessageContent := strings.TrimSpace(log.LogContent[startIndex:])
			result := &powertrading.MachineTestMessage{}
			marshalErr := json.Unmarshal([]byte(finalMessageContent), result)
			if marshalErr != nil {
				task.Detail[task.CurrentStep].DetailList[log.IP].Message =
					fmt.Sprintf("marshal rsp log to MachineTestMessage error:%s", marshalErr.Error())
				continue
			}
			task.Detail[task.CurrentStep].DetailList[log.IP] = result
			if result.Pass == "false" {
				if result.AbleToClean == "false" {
					task.Summary[storage.MachineCheckFailure][task.CurrentStep] =
						append(task.Summary[storage.MachineCheckFailure][task.CurrentStep], log.IP)
				} else {
					task.Summary[storage.MachineNeedClean][task.CurrentStep] =
						append(task.Summary[storage.MachineNeedClean][task.CurrentStep], log.IP)
				}
			} else {
				if lastStep {
					task.Summary[storage.MachineCheckSuccess][task.CurrentStep] =
						append(task.Summary[storage.MachineCheckSuccess][task.CurrentStep], log.IP)
				}
			}
		}
		for _, info := range infos {
			if !returnIPs[info.IP] {
				if task.Detail[task.CurrentStep].DetailList[info.IP] == nil {
					task.Detail[task.CurrentStep].DetailList[info.IP] = &powertrading.MachineTestMessage{
						Ip:      info.IP,
						Pass:    "false",
						Message: "",
					}
				}
				tag := info.Tag
				if info.Status == 18 {
					tag = "agent未安装"
				}
				task.Detail[task.CurrentStep].DetailList[info.IP].Message = fmt.Sprintf("status:%d, tag:%s",
					info.Status, tag)
				if lastStep {
					task.Summary[storage.MachineCheckFailure][task.CurrentStep] =
						append(task.Summary[storage.MachineCheckFailure][task.CurrentStep], info.IP)
				}
			}
		}
	}
	return true, nil
}

func getJobResult(task *storage.MachineTask, lastStep bool, jobStatus *job.StatusResponse) bool {
	task.Summary[storage.MachineCheckFailure][task.CurrentStep] = make([]string, 0)
	task.Summary[storage.MachineCheckSuccess][task.CurrentStep] = make([]string, 0)
	task.Summary[storage.MachineNeedClean][task.CurrentStep] = make([]string, 0)
	stepInstanceId := jobStatus.Data.StepInstanceList[0].StepInstanceID
	blog.Infof("stepInstanceId %d", stepInstanceId)
	pass := true
	for _, status := range jobStatus.Data.StepInstanceList {
		for _, stepIns := range status.StepIPResultList {
			if task.Detail[task.CurrentStep].DetailList[stepIns.IP] == nil {
				task.Detail[task.CurrentStep].DetailList[stepIns.IP] = &powertrading.MachineTestMessage{
					Ip:      stepIns.IP,
					Pass:    "false",
					Message: "",
				}
			}
			if stepIns.Status == 18 {
				pass = false
				task.Detail[task.CurrentStep].DetailList[stepIns.IP].Message = fmt.Sprintf("status:%d, agent未安装",
					stepIns.Status)
				task.Summary[storage.MachineCheckFailure][task.CurrentStep] =
					append(task.Summary[storage.MachineCheckFailure][task.CurrentStep], stepIns.IP)
			} else if stepIns.ExitCode == 0 {
				task.Detail[task.CurrentStep].DetailList[stepIns.IP] = &powertrading.MachineTestMessage{
					Ip:      stepIns.IP,
					Pass:    "true",
					Message: "",
				}
				if lastStep {
					task.Summary[storage.MachineCheckSuccess][task.CurrentStep] =
						append(task.Summary[storage.MachineCheckSuccess][task.CurrentStep], stepIns.IP)
				}
			} else {
				pass = false
				task.Detail[task.CurrentStep].DetailList[stepIns.IP] = &powertrading.MachineTestMessage{
					Ip:      stepIns.IP,
					Pass:    "false",
					Message: fmt.Sprintf("exit code: %d", stepIns.ExitCode),
				}
				task.Summary[storage.MachineCheckFailure][task.CurrentStep] =
					append(task.Summary[storage.MachineCheckFailure][task.CurrentStep], stepIns.IP)
			}
		}
	}
	return pass
}
