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
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/panjf2000/ants/v2"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/bkcc"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/bksops"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/clustermgr"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/cr"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/job"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/resourcemgr"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/storage"
	powertrading "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/proto"
)

// MachineTestHandler machine test handler
type MachineTestHandler struct {
	ClusterCli  clustermgr.Client
	ResourceCli resourcemgr.Client
	Storage     storage.Storage
	BksopsCli   bksops.Client
	JobCli      job.Client
	CrCli       cr.Client
	CcCli       bkcc.Client
	Interval    int
	Concurrency int
	CheckConf   *CheckConf
}

// CheckConf check config
type CheckConf struct {
	memPercent   float64
	memUsage     float64
	templateId   string
	templateName string
}

const (
	memoryCheckDays = 7
	memPercent      = 0.85
	memUsage        = 2
	templateId      = "10179"
	templateName    = "TKExIEG混部集群第三方节点上架前检测"
)

// Init init handler
func (h *MachineTestHandler) Init() {
	checkConf := &CheckConf{
		memPercent:   memPercent,
		memUsage:     memUsage,
		templateId:   templateId,
		templateName: templateName,
	}
	if os.Getenv("bcsTestTemplateID") != "" {
		checkConf.templateId = os.Getenv("bcsTestTemplateID")
	}
	if os.Getenv("bcsTestTemplateName") != "" {
		checkConf.templateName = os.Getenv("bcsTestTemplateName")
	}
	if os.Getenv("bcsTestMemPercent") != "" {
		checkConf.templateId = os.Getenv("bcsTestMemPercent")
	}
	if os.Getenv("bcsTestMemUsage") != "" {
		checkConf.templateId = os.Getenv("bcsTestMemUsage")
	}
	h.CheckConf = checkConf
}

// Run run handler
func (h *MachineTestHandler) Run(ctx context.Context) {
	tick := time.NewTicker(time.Second * time.Duration(h.Interval))
	for {
		select {
		case now := <-tick.C:
			// main loops
			blog.Infof("############## task controller ticker: %s################", now.Format(time.RFC3339))
			h.controllerLoops(ctx)
		case <-ctx.Done():
			blog.Infof("Task Controller is asked to exit")
			return
		}
	}
}

func (h *MachineTestHandler) controllerLoops(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			blog.Errorf("[machineTestHandler] panic in MachineTestHandler, info: %v, stack:%s", r,
				string(debug.Stack()))
		}
	}()
	tasks, err := h.Storage.ListTasks(ctx, storage.CheckTask, &storage.ListOptions{})
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

// handleOneTask handle one step in one time
func (h *MachineTestHandler) handleOneTask(ctx context.Context, task *storage.MachineTask) {
	switch task.CurrentStep {
	case storage.BusinessCheck:
		h.businessCheck(ctx, task)
		return
	case storage.ClusterCheck:
		h.clusterCheck(ctx, task)
		return
	case storage.ImportedCheck:
		h.importerCheck(ctx, task)
		return
	case storage.MemoryCheck:
		h.memoryCheck(ctx, task)
		return
	case storage.BkOpsTaskCheck:
		BksopsCheck(ctx, h.BksopsCli, h.JobCli, h.Storage, task, &OpsConf{
			TemplateId:   h.CheckConf.templateId,
			TemplateName: h.CheckConf.templateName,
		})
		return
	default:
		blog.Errorf("unsupported step %s", task.CurrentStep)
	}
}

// clusterCheck check if node has added to bcs cluster
func (h *MachineTestHandler) clusterCheck(ctx context.Context, task *storage.MachineTask) {
	blog.Infof("begin to cluster check task %s", task.TaskID)
	remainIPs := getRemainIPs(task.IPList, task.Summary[storage.MachineCheckFailure])
	task.Detail[storage.ClusterCheck].IPList = remainIPs
	task.Summary[storage.MachineCheckFailure][storage.ClusterCheck] = make([]string, 0)
	for _, ip := range remainIPs {
		node, err := h.ClusterCli.GetNodeDetail(ctx, ip)
		if err != nil {
			task.Detail[storage.ClusterCheck].DetailList[ip] = &powertrading.MachineTestMessage{
				Ip:      ip,
				Pass:    "false",
				Message: fmt.Sprintf("get node failed:%s", err.Error()),
			}
			task.Summary[storage.MachineCheckFailure][storage.ClusterCheck] =
				append(task.Summary[storage.MachineCheckFailure][storage.ClusterCheck], ip)
			continue
		}
		if node != nil {
			task.Detail[storage.ClusterCheck].DetailList[ip] = &powertrading.MachineTestMessage{
				Ip:      ip,
				Pass:    "false",
				Message: fmt.Sprintf("node has been added in cluster:%s", node.ClusterID),
			}
			task.Summary[storage.MachineCheckFailure][storage.ClusterCheck] =
				append(task.Summary[storage.MachineCheckFailure][storage.ClusterCheck], ip)
			continue
		}
	}
	task.Detail[storage.ClusterCheck].Status = storage.TaskFinished
	if !checkIfContinue(task.IPList, task.Summary[storage.MachineCheckFailure]) {
		task.Status = storage.TaskFailed
	} else {
		task.CurrentStep = storage.ImportedCheck
		task.Detail[storage.ImportedCheck].Status = storage.TaskRunning
	}
	_, err := h.Storage.UpdateTask(ctx, task, &storage.UpdateOptions{})
	if err != nil {
		blog.Errorf("update task %s error:%s", task.TaskID, err.Error())
	}
}

// importerCheck check if node has imported to bcs device pool
func (h *MachineTestHandler) importerCheck(ctx context.Context, task *storage.MachineTask) {
	blog.Infof("begin to import check task %s", task.TaskID)
	remainIPs := getRemainIPs(task.IPList, task.Summary[storage.MachineCheckFailure])
	task.Detail[storage.ImportedCheck].IPList = remainIPs
	hosts, err := h.CcCli.ListHostByCC(ctx, remainIPs, task.BusinessID)
	if err != nil {
		blog.Errorf("list host by cc failed:%s", err.Error())
		task.Detail[storage.ImportedCheck].Status = storage.TaskFailed
		task.Detail[storage.ImportedCheck].Message = fmt.Sprintf("list host by cc failed:%s", err.Error())
		task.Status = storage.TaskFailed
		_, updateErr := h.Storage.UpdateTask(ctx, task, &storage.UpdateOptions{})
		if updateErr != nil {
			blog.Errorf("update task %s error:%s", task.TaskID, updateErr.Error())
		}
		return
	}
	assetList := make([]string, 0)
	for _, host := range hosts {
		assetList = append(assetList, host.BKAssetID)
	}
	devices, err := h.ResourceCli.ListDeviceByAssetIds(ctx, int64(len(assetList)), assetList)
	if err != nil {
		blog.Errorf("list host by assetList from devicePool failed:%s", err.Error())
		task.Detail[storage.ImportedCheck].Status = storage.TaskFailed
		task.Detail[storage.ImportedCheck].Message =
			fmt.Sprintf("list host by assetList from devicePool failed:%s", err.Error())
		task.Status = storage.TaskFailed
		_, updateErr := h.Storage.UpdateTask(ctx, task, &storage.UpdateOptions{})
		if updateErr != nil {
			blog.Errorf("update task %s error:%s", task.TaskID, updateErr.Error())
		}
		return
	}
	task.Summary[storage.MachineCheckFailure][storage.ImportedCheck] = make([]string, 0)
	for _, device := range devices {
		task.Detail[storage.ImportedCheck].DetailList[*device.Info.InnerIP] = &powertrading.MachineTestMessage{
			Ip:      *device.Info.InnerIP,
			Pass:    "false",
			Message: fmt.Sprintf("%s has imported to device pool", *device.Info.InnerIP),
		}
		task.Summary[storage.MachineCheckFailure][storage.ImportedCheck] =
			append(task.Summary[storage.MachineCheckFailure][storage.ImportedCheck], *device.Info.InnerIP)
	}
	task.Detail[storage.ImportedCheck].Status = storage.TaskFinished
	if !checkIfContinue(task.IPList, task.Summary[storage.MachineCheckFailure]) {
		task.Status = storage.TaskFailed
	} else {
		task.CurrentStep = storage.MemoryCheck
		task.Detail[storage.MemoryCheck].Status = storage.TaskRunning
	}
	_, updateErr := h.Storage.UpdateTask(ctx, task, &storage.UpdateOptions{})
	if updateErr != nil {
		blog.Errorf("update task %s error:%s", task.TaskID, updateErr.Error())
	}
}

// memoryCheck check if the memory usage satisfies condition
// memory usage in past week should satisfy the following condition:
// 1. maxPercent <=85%
// 2. totalMemory - maxUsage >= 2
func (h *MachineTestHandler) memoryCheck(ctx context.Context, task *storage.MachineTask) {
	blog.Infof("begin to handle task %s", task.TaskID)
	remainIPs := getRemainIPs(task.IPList, task.Summary[storage.MachineCheckFailure])
	task.Detail[storage.MemoryCheck].IPList = remainIPs
	dates := getDefaultDateRange()
	memoryMax := make(map[string]*storage.MachineSpecification)
	for index := range dates {
		req := &cr.GetPerfDetailReq{
			Dsl: &cr.GetPerfDetailDsl{MatchExpr: []cr.GetPerfDetailMatchExpr{{
				Key:      "IP",
				Values:   remainIPs,
				Operator: "In",
			}, {
				Key:      "sync_date",
				Values:   []string{dates[index]},
				Operator: "In",
			}}},
			Offset: 0,
			Limit:  len(remainIPs),
		}
		rsp, err := h.CrCli.GetPerfDetail(req)
		if err != nil {
			blog.Errorf("GetPerfDetail error:%s, req:%v", err.Error(), req)
			continue
		}
		for _, data := range rsp.Data.Items {
			if memoryMax[data.IP] == nil {
				memoryMax[data.IP] = &storage.MachineSpecification{
					TotalMem:   data.MemTotal,
					MemPercent: data.Mem4Linux / data.MemTotal,
					TotalCPU:   data.MaxCpuCoreAmount,
					CpuPercent: data.CpuPercent,
				}
			} else {
				if memoryMax[data.IP].MemPercent < (data.Mem4Linux / data.MemTotal) {
					memoryMax[data.IP].MemPercent = data.Mem4Linux / data.MemTotal
				}
				if memoryMax[data.IP].CpuPercent < data.CpuPercent {
					memoryMax[data.IP].CpuPercent = data.CpuPercent
				}
			}
		}
	}
	task.Summary[storage.MachineCheckFailure][storage.MemoryCheck] = make([]string, 0)
	for _, ip := range remainIPs {
		machineInfo := memoryMax[ip]
		if machineInfo == nil {
			continue
		}
		if (h.CheckConf.memPercent != 0 && (machineInfo.MemPercent >= h.CheckConf.memPercent)) ||
			(h.CheckConf.memUsage != 0 && (machineInfo.TotalMem*(1-machineInfo.MemPercent) <= h.CheckConf.memUsage)) {
			task.Detail[storage.MemoryCheck].DetailList[ip] = &powertrading.MachineTestMessage{
				Ip:          ip,
				Pass:        "false",
				AbleToClean: "false",
				Message: fmt.Sprintf("total %f, max percent %f, max mem usage %f, max usage in recent week "+
					"higher than 85 percent or remain less than 2", machineInfo.TotalMem, machineInfo.MemPercent,
					machineInfo.TotalMem*machineInfo.MemPercent),
			}
			task.Summary[storage.MachineCheckFailure][storage.MemoryCheck] =
				append(task.Summary[storage.MachineCheckFailure][storage.MemoryCheck], ip)
		}
	}
	task.Detail[storage.MemoryCheck].Status = storage.TaskFinished
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

// businessCheck check if machine belongs to specific business
func (h *MachineTestHandler) businessCheck(ctx context.Context, task *storage.MachineTask) {
	hosts, err := h.CcCli.ListHostByCC(ctx, task.IPList, task.BusinessID)
	if err != nil {
		task.Status = storage.TaskFailed
		task.Detail[storage.BusinessCheck].Status = storage.TaskFailed
		task.Detail[storage.BusinessCheck].Message = fmt.Sprintf("request bkcc failed:%s", err.Error())
		task.Summary[storage.MachineCheckFailure][storage.BusinessCheck] = task.IPList
		_, updateErr := h.Storage.UpdateTask(ctx, task, &storage.UpdateOptions{})
		if updateErr != nil {
			blog.Errorf("update task %s error:%s", task.TaskID, err.Error())
		}
		return
	}
	hostMap := make(map[string]*bkcc.CCHostInfo)
	for i := range hosts {
		hostMap[hosts[i].BKHostInnerIP] = &hosts[i]
	}
	notExist := make([]string, 0)
	for i := range task.IPList {
		if _, ok := hostMap[task.IPList[i]]; !ok {
			notExist = append(notExist, task.IPList[i])
		}
	}
	if len(notExist) != 0 {
		for _, notExistIP := range notExist {
			task.Detail[storage.BusinessCheck].DetailList[notExistIP] = &powertrading.MachineTestMessage{
				Ip:      notExistIP,
				Pass:    "false",
				Message: fmt.Sprintf("%s not exist in business %s", notExistIP, task.BusinessID),
			}
		}
		task.Summary[storage.MachineCheckFailure][storage.BusinessCheck] = notExist
	}
	if len(notExist) == len(task.IPList) {
		task.Detail[storage.BusinessCheck].Status = storage.TaskFinished
		task.Status = storage.TaskFailed
	} else {
		task.Detail[storage.BusinessCheck].Status = storage.TaskFinished
		task.CurrentStep = storage.ClusterCheck
	}
	task.Detail[storage.BusinessCheck].IPList = task.IPList
	_, err = h.Storage.UpdateTask(ctx, task, &storage.UpdateOptions{})
	if err != nil {
		blog.Errorf("update task %s error:%s", task.TaskID, err.Error())
	}
}
