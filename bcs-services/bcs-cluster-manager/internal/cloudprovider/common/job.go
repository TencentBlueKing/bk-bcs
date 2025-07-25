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

package common

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/template"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/job"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
)

var (
	jobExecuteScriptStep = cloudprovider.StepInfo{
		StepMethod: cloudprovider.JobFastExecuteScriptAction,
		StepName:   "用户初始化作业",
	}
)

const (
	// PreInitJob xxx
	PreInitJob = "preInitJob"
	// PostInitJob xxx
	PostInitJob = "postInitJob"
	// PreInitStepJob xxx
	PreInitStepJob = "前置初始化作业"
	// PostInitStepJob xxx
	PostInitStepJob = "后置初始化作业"
)

// JobExecParas parameters
type JobExecParas struct {
	ClusterID string
	// Content base64 encode script
	Content  string
	NodeIps  string
	Operator string
	// StepName 自定义step名称
	StepName string
	// AllowSkipJobTask 任务失败时候是否允许自动跳过
	AllowSkipJobTask bool
	Translate        string
}

// BuildJobExecuteScriptStep build job execute script step
func BuildJobExecuteScriptStep(task *proto.Task, paras JobExecParas) {
	if paras.StepName != "" {
		jobExecuteScriptStep.StepName = paras.StepName
	}
	jobScriptStep := cloudprovider.InitTaskStep(jobExecuteScriptStep,
		cloudprovider.WithStepSkipFailed(paras.AllowSkipJobTask),
		cloudprovider.WithStepTranslate(paras.Translate),
	)

	if len(paras.NodeIps) == 0 {
		paras.NodeIps = template.NodeIPList
	}

	jobScriptStep.Params[cloudprovider.ClusterIDKey.String()] = paras.ClusterID
	jobScriptStep.Params[cloudprovider.ScriptContentKey.String()] = paras.Content
	jobScriptStep.Params[cloudprovider.NodeIPsKey.String()] = paras.NodeIps
	jobScriptStep.Params[cloudprovider.OperatorKey.String()] = paras.Operator

	task.Steps[jobExecuteScriptStep.StepMethod] = jobScriptStep
	task.StepSequence = append(task.StepSequence, jobExecuteScriptStep.StepMethod)
}

// JobScriptPara job paras
type JobScriptPara struct {
	BizID   string
	Script  string
	NodeIPs []string
}

func renderScript(ctx context.Context, clusterID, content, nodeIPs, operator string) (*JobScriptPara, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	cls, err := cloudprovider.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}
	paras := &JobScriptPara{
		BizID: cls.GetBusinessID(),
	}

	blog.Infof("renderScript[%s] before: %s,%s", taskID, content, nodeIPs)
	script, err := template.GetNodeTemplateScript(template.RenderVars{
		Cluster:  cls,
		IPList:   nodeIPs,
		Operator: operator,
		Render:   true,
	}, content)
	if err != nil {
		blog.Errorf("renderScript[%s] failed: %v", taskID, err)
		return nil, err
	}

	blog.Infof("renderScript[%s] success: %s", taskID, script)
	paras.Script = script
	paras.NodeIPs = strings.Split(nodeIPs, ",")

	return paras, nil
}

// JobExecuteScriptTask execute job script
func JobExecuteScriptTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start execute job script")
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("JobExecuteScriptTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("JobExecuteScriptTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// clusterID / scriptContent(base64编码) / nodeIPs / operator
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	content := step.Params[cloudprovider.ScriptContentKey.String()]
	nodeIPs := step.Params[cloudprovider.NodeIPsKey.String()]
	operator := step.Params[cloudprovider.OperatorKey.String()]

	// nodeIPs
	if nodeIPs == template.NodeIPList {
		if value, ok := template.DynamicParameterInject[nodeIPs]; ok {
			nodeIPs = state.Task.CommonParams[value]
		}
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDAndStepNameForContext(context.Background(), taskID, stepName)

	// render script && base64 encode
	jobParas, err := renderScript(ctx, clusterID, content, nodeIPs, operator)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("render script failed [%s]", err))
		blog.Errorf("JobExecuteScriptTask[%s] renderScript failed: %v", taskID, err)
		if step.GetSkipOnFailed() {
			_ = state.SkipFailure(start, stepName, err)
			return nil
		}
		_ = state.UpdateStepFailure(start, stepName, err)
		return err
	}
	blog.Infof("JobExecuteScriptTask[%s] renderScript successful[%s:%v:%s]", taskID,
		jobParas.BizID, jobParas.NodeIPs, jobParas.Script)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"render script successful")

	url, err := ExecuteScriptByJob(ctx, stepName, jobParas.BizID, jobParas.Script, jobParas.NodeIPs)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("execute script failed [%s]", err))
		blog.Errorf("JobExecuteScriptTask[%s] ExecuteScriptByJob failed: %v", taskID, err)
		state.TaskURL = url
		if step.GetSkipOnFailed() {
			_ = state.SkipFailure(start, stepName, err)
			return nil
		}
		_ = state.UpdateStepFailure(start, stepName, err)
		return err
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"execute job script successful")

	// update step
	state.TaskURL = url
	_ = state.UpdateStepSucc(start, stepName)

	return nil
}

// ExecuteScriptByJob execute job script
func ExecuteScriptByJob(ctx context.Context, stepName, bizID, content string, ips []string) (string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	var (
		jobID uint64
		err   error
	)

	servers, err := GetIPCloudIDByNodeIPs(ctx, bizID, ips)
	if err != nil {
		blog.Errorf("task[%s] ExecuteScriptByJob failed: %v", taskID, err)
		return "", err
	}

	err = retry.Do(func() error {
		jobID, err = job.GetJobClient().ExecuteScript(ctx, job.ExecuteScriptParas{
			BizID:         bizID,
			ScriptContent: content,
			Servers:       servers,
		})
		if err != nil {
			return err
		}

		return nil
	}, retry.Attempts(3), retry.DelayType(retry.FixedDelay), retry.Delay(time.Millisecond*100))
	if err != nil {
		blog.Errorf("task[%s] ExecuteScriptByJob ExecuteScript failed: %v", taskID, err)
		return "", err
	}
	blog.Infof("task[%s] ExecuteScriptByJob[%v] ExecuteScript successful", taskID, jobID)

	// update job taskUrl to task
	_ = cloudprovider.SetTaskStepParas(taskID, stepName, cloudprovider.BkSopsTaskURLKey.String(),
		job.GetJobTaskLink(jobID))

	// check status
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Minute)
	defer cancel()

	err = loop.LoopDoFunc(ctx, func() error {
		status, errGet := job.GetJobClient().GetJobStatus(ctx, job.JobInfo{JobID: jobID, BizID: bizID})
		if errGet != nil {
			blog.Errorf("task[%s] ExecuteScriptByJob  GetJobStatus[%s:%v] failed: %v",
				taskID, bizID, jobID, errGet)
			return errGet
		}

		switch status {
		case job.Executing, job.UnExecuted:
			blog.Infof("task[%s] ExecuteScriptByJob  GetJobStatus[%s:%v] job status[%v]",
				taskID, bizID, jobID, status)

			cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
				fmt.Sprintf("job status [%v]", status))
			return nil
		case job.ExecuteSuccess:
			return loop.EndLoop
		case job.UnKnownStatus, job.ExecuteFailure:
			return fmt.Errorf("task[%s] ExecuteScriptByJob  GetJobStatus[%s:%v] failed: %v",
				taskID, bizID, jobID, status)
		}
		return nil
	}, loop.LoopInterval(5*time.Second))
	if err != nil {
		blog.Errorf("task[%s] ExecuteScriptByJob[%v] ExecuteScript failed: %v", taskID, jobID, err)
		return job.GetJobTaskLink(jobID), err
	}

	blog.Infof("task[%s] ExecuteScriptByJob[%v] ExecuteScript successful", taskID, jobID)
	return job.GetJobTaskLink(jobID), nil
}

func getBizHosts(bizID string) (map[int64]cmdb.HostData, error) {
	biz, err := strconv.Atoi(bizID)
	if err != nil {
		blog.Errorf("strconv BusinessID to int failed: %v", err)
		return nil, err
	}
	hosts, err := cmdb.GetCmdbClient().FetchAllHostsByBizID(biz, false)
	if err != nil {
		blog.Errorf("cmdb FetchAllHostsByBizID failed: %v", err)
		return nil, err
	}

	var (
		hostsMap = make(map[int64]cmdb.HostData, 0)
	)
	for i := range hosts {
		hostsMap[hosts[i].BKHostID] = hosts[i]
	}
	return hostsMap, nil
}

// GetIPCloudIDByNodeIPs get serverIP cloudInfo by cmdb
func GetIPCloudIDByNodeIPs(ctx context.Context, bizID string, ips []string) ([]job.ServerInfo, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// withoutBiz maybe get other biz host
	hostDetailData, err := cmdb.GetCmdbClient().QueryAllHostInfoWithoutBiz(ips)
	if err != nil {
		blog.Errorf("task[%s] GetIPCloudIDByNodeIPs failed: %v", taskID, err)
		return nil, err
	}

	bizHostsMap, err := getBizHosts(bizID)
	if err != nil {
		return nil, err
	}

	var (
		servers = make([]job.ServerInfo, 0)
	)
	for _, host := range hostDetailData {
		_, ok := bizHostsMap[host.BKHostID]
		if !ok {
			continue
		}

		servers = append(servers, job.ServerInfo{
			BkCloudID: uint64(host.BkCloudID),
			Ip:        host.BKHostInnerIP,
		})
	}

	blog.Infof("task[%s] GetIPCloudIDByNodeIPs[%v] successful[%v]", taskID, ips, servers)
	return servers, nil
}
