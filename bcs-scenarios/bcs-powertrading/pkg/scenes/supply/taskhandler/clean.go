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

// Package taskhandler xxx
package taskhandler

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/panjf2000/ants/v2"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/bksops"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/job"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/storage"
)

// MachineCleanHandler machine clean handler
type MachineCleanHandler struct {
	Storage     storage.Storage
	BksopsCli   bksops.Client
	JobCli      job.Client
	conf        *CleanConf
	Interval    int
	Concurrency int
}

// CleanConf clean config struct
type CleanConf struct {
	templateId   string
	templateName string
}

const (
	cleanTemplateId   = "10180"
	cleanTemplateName = "TKExIEG混部集群第三方节点上架前清理"
)

// Init init handler
func (h *MachineCleanHandler) Init() {
	checkConf := &CleanConf{
		templateId:   cleanTemplateId,
		templateName: cleanTemplateName,
	}
	if os.Getenv("bcsCleanTemplateID") != "" {
		checkConf.templateId = os.Getenv("bcsCleanTemplateID")
	}
	if os.Getenv("bcsCleanTemplateName") != "" {
		checkConf.templateName = os.Getenv("bcsCleanTemplateName")
	}
	h.conf = checkConf
}

// Run run handler
func (h *MachineCleanHandler) Run(ctx context.Context) {
	tick := time.NewTicker(time.Second * time.Duration(h.Interval))
	for {
		select {
		case now := <-tick.C:
			// main loops
			blog.Infof("############## clean controller ticker: %s################", now.Format(time.RFC3339))
			h.controllerLoops(ctx)
		case <-ctx.Done():
			blog.Infof("clean Controller is asked to exit")
			return
		}
	}
}

func (h *MachineCleanHandler) controllerLoops(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			blog.Errorf("[machineTestHandler] panic in MachineTestHandler, info: %v, stack:%s", r,
				string(debug.Stack()))
		}
	}()
	tasks, err := h.Storage.ListTasks(ctx, storage.CleanTask, &storage.ListOptions{})
	if err != nil {
		blog.Errorf("list machineTestTask error:%s", err.Error())
		return
	}
	wg := sync.WaitGroup{}
	pool, err := ants.NewPool(h.Concurrency)
	if err != nil {
		blog.Errorf("[producer] init new pool err:%v", err)
		return
	}
	blog.Infof("[producer] concurrency:%d", h.Concurrency)
	defer pool.Release()
	for index := range tasks {
		task := tasks[index]
		if task.Status != storage.TaskRunning {
			continue
		}
		wg.Add(1)
		err := pool.Submit(func() {
			defer func() {
				if r := recover(); r != nil {
					blog.Errorf("[machineTestHandler] panic in handle one task, info: %v, stack:%s", r,
						string(debug.Stack()))
					wg.Done()
				}
			}()
			h.handleOneTask(ctx, task)
			wg.Done()
		})
		if err != nil {
			blog.Errorf("submit task to ch pool err:%v", err)
		}
	}
	wg.Wait()
}

func (h *MachineCleanHandler) handleOneTask(ctx context.Context, task *storage.MachineTask) {
	switch task.CurrentStep {
	case storage.BkOpsTaskCheck:
		BksopsCheck(ctx, h.BksopsCli, h.JobCli, h.Storage, task, &OpsConf{
			TemplateId:   templateId,
			TemplateName: templateName,
		})
		return
	case storage.BkOpsTaskClean:
		h.bksopsClean(ctx, task)
		return
	default:
		blog.Errorf("unsupported step %s", task.CurrentStep)
	}
}
func (h *MachineCleanHandler) bksopsClean(ctx context.Context, task *storage.MachineTask) {
	if task.Detail[storage.BkOpsTaskClean] == nil || task.Detail[storage.BkOpsTaskClean].BksOpsTaskID == "" {
		constants := make(map[string]string)
		ipStr := ""
		for key, ip := range task.IPList {
			ipStr += ip
			if key != len(task.IPList)-1 {
				ipStr += ","
			}
		}
		constants["${node_ip_list}"] = ipStr
		constants["${biz_cc_id}"] = task.BusinessID
		blog.Infof("req constant:%v", constants)
		err := createAndStartBksOps(ctx, h.BksopsCli, h.Storage, task, constants, &OpsConf{
			TemplateId:   h.conf.templateId,
			TemplateName: h.conf.templateName,
		})
		if err != nil {
			blog.Errorf("%s task check error:%s", task.TaskID, err.Error())
			task.RetryTimes++
			if task.RetryTimes == 3 {
				task.Message = fmt.Sprintf("retry task %s 3 times, finished this task, error:%s",
					task.TaskID, err.Error())
				blog.Errorf("retry task %s 3 times, finished this task", task.TaskID)
				task.Detail[storage.BkOpsTaskClean].Status = storage.TaskFailed
				task.Status = storage.TaskFailed
			}
			_, err = h.Storage.UpdateTask(ctx, task, &storage.UpdateOptions{})
			if err != nil {
				blog.Errorf("update task %s error:%s", task.TaskID, err.Error())
			}
		}
		return
	}
	finished, pass, checkErr := CheckJobStatus(ctx, h.BksopsCli, h.JobCli, h.Storage, task, false,
		"执行清理", false)
	if checkErr != nil {
		blog.Errorf("%s task check job status error:%s", task.TaskID, checkErr.Error())
		task.RetryTimes++
		if task.RetryTimes == 3 {
			task.Message = fmt.Sprintf("retry task %s 3 times, finished this task, error:%s",
				task.TaskID, checkErr.Error())
			blog.Errorf("retry task %s 3 times, finished this task", task.TaskID)
			task.Detail[storage.BkOpsTaskClean].Status = storage.TaskFailed
			task.Status = storage.TaskFailed
		}
		_, err := h.Storage.UpdateTask(ctx, task, &storage.UpdateOptions{})
		if err != nil {
			blog.Errorf("update task %s error:%s", task.TaskID, err.Error())
		}
		return
	}
	if !finished {
		return
	}
	if pass {
		task.Detail[storage.BkOpsTaskClean].Status = storage.TaskFinished
	} else {
		task.Detail[storage.BkOpsTaskClean].Status = storage.TaskFailed
	}
	if checkIfContinue(task.IPList, task.Summary[storage.MachineCheckFailure]) {
		task.CurrentStep = storage.BkOpsTaskCheck
		task.Detail[storage.BkOpsTaskCheck].Status = storage.TaskRunning
	} else {
		task.Status = storage.TaskFailed
	}
	_, err := h.Storage.UpdateTask(ctx, task, &storage.UpdateOptions{})
	if err != nil {
		blog.Errorf("update task %s error:%s", task.TaskID, err.Error())
	}
}
