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

// Package callback xxx
package callback

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/quota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/entity"
	uquota "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/quota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

const (
	// QuotaCallBackName quota
	QuotaCallBackName = "quota"
)

// NewQuotaCallback 创建资源管理回调实现
func NewQuotaCallback() task.CallbackInterface {
	return &quotaCallBack{}
}

type quotaCallBack struct{}

// GetName 回调方法名称
func (q *quotaCallBack) GetName() string {
	return QuotaCallBackName
}

// Callback 回调方法,根据任务成功状态更新实体对象状态
func (q *quotaCallBack) Callback(isSuccess bool, task *types.Task) {
	if q == nil {
		return
	}

	taskType, ok := task.GetCommonParams(utils.TaskType.String())
	if !ok {
		logging.Error("task[%s] get taskType failed", task.GetTaskID())
		return
	}
	quotaId, ok := task.GetCommonParams(utils.QuotaIdKey.String())
	if !ok {
		logging.Error("task[%s] get quotaId failed", task.GetTaskID())
		return
	}

	logging.Info("task[%s] takType[%s] execute %+v", task.GetTaskID(), taskType, isSuccess)

	switch taskType {
	case utils.CreateProjectQuota.GetJobType():
		result := buildQuotaJobResult(quota.Running.String(), quota.CreateFailure.String())
		result.setQuotaId(quotaId)
		err := result.createProjectQuota(isSuccess)
		if err != nil {
			logging.Error("task[%s] create project quota failed, err: %s", task.GetTaskID(), err.Error())
		}
	case utils.DeleteProjectQuota.GetJobType():
		result := buildQuotaJobResult(quota.Deleted.String(), quota.DeleteFailure.String())
		result.setQuotaId(quotaId)
		err := result.deleteProjectQuota(isSuccess)
		if err != nil {
			logging.Error("task[%s] delete project quota failed, err: %s", task.GetTaskID(), err.Error())
		}
	case utils.ScaleUpProjectQuota.GetJobType(), utils.ScaleDownProjectQuota.GetJobType():
		quotaResource, exist := task.GetCommonParams(utils.QuotaResource.String())
		if !exist {
			logging.Error("task[%s] get quota resource failed", task.GetTaskID())
			return
		}
		result := buildQuotaJobResult("", "")
		result.setQuotaId(quotaId)
		result.setTaskId(task.GetTaskID())
		result.setQuotaResource(quotaResource)

		// 调增 or 调减 quota
		scaleUp := false
		if taskType == utils.ScaleUpProjectQuota.GetJobType() {
			scaleUp = true
		}

		err := result.scaleUpOrDownProjectQuota(isSuccess, scaleUp)
		if err != nil {
			logging.Error("task[%s] scaleUp/scaleDown project quota failed, err: %s", task.GetTaskID(), err.Error())
		}
	default:
	}

}

func buildQuotaJobResult(success, failure string) *syncQuotaJobResult {
	return &syncQuotaJobResult{
		success: success,
		failure: failure,
	}
}

type syncQuotaJobResult struct {
	success       string
	failure       string
	quotaId       string
	quotaResource string
	taskId        string
}

func (job *syncQuotaJobResult) setSuccessStatus(success string) { // nolint
	job.success = success
}

func (job *syncQuotaJobResult) setFailureStatus(failure string) { // nolint
	job.failure = failure
}

func (job *syncQuotaJobResult) setQuotaId(quotaId string) {
	job.quotaId = quotaId
}

func (job *syncQuotaJobResult) setQuotaResource(resource string) {
	job.quotaResource = resource
}

func (job *syncQuotaJobResult) setTaskId(taskId string) {
	job.taskId = taskId
}

func (job *syncQuotaJobResult) getStatus(isSuccess bool) string {
	if isSuccess {
		return job.success
	}

	return job.failure
}

func (job *syncQuotaJobResult) createProjectQuota(isSuccess bool) error {
	updateField := entity.M{
		quota.FieldKeyQuotaId:    job.quotaId,
		quota.FieldKeyUpdateTime: time.Now().Format(time.RFC3339),
		quota.FieldKeyStatus:     job.getStatus(isSuccess),
	}

	return store.GetModel().UpdateProjectQuotaByField(context.Background(), updateField)
}

func (job *syncQuotaJobResult) deleteProjectQuota(isSuccess bool) error {
	updateField := entity.M{
		quota.FieldKeyQuotaId:    job.quotaId,
		quota.FieldKeyUpdateTime: time.Now().Format(time.RFC3339),
		quota.FieldKeyStatus:     job.getStatus(isSuccess),
	}

	return store.GetModel().UpdateProjectQuotaByField(context.Background(), updateField)
}

func (job *syncQuotaJobResult) scaleUpOrDownProjectQuota(isSuccess, scaleUp bool) error {
	originQuota, err := store.GetModel().GetProjectQuotaById(context.Background(), job.quotaId)
	if err != nil {
		return err
	}

	resource := &bcsproject.QuotaResource{}
	err = json.Unmarshal([]byte(job.quotaResource), resource)
	if err != nil {
		return err
	}

	logging.Info("scaleUpOrDownProjectQuota[%s] success: %v, scaleUp: %v, originQuota: %+v, quotaResource: %+v",
		job.taskId, isSuccess, scaleUp, originQuota, resource)

	// 任务执行失败则 不进行额度增减
	if !isSuccess {
		return nil
	}

	// 根据类型 进行对应额度增减
	switch originQuota.QuotaType {
	case quota.Host:
		if scaleUp {
			originQuota.Quota.HostResources.QuotaNum += resource.ZoneResources.GetQuotaNum()
		} else {
			if resource.ZoneResources.GetQuotaNum() >= originQuota.Quota.HostResources.QuotaNum {
				originQuota.Quota.HostResources.QuotaNum = 0
			} else {
				originQuota.Quota.HostResources.QuotaNum -= resource.ZoneResources.GetQuotaNum()
			}
		}
	case quota.Shared, quota.Federation:
		// default scale down
		addOrSub := false
		if scaleUp {
			addOrSub = true
		}

		if resource.GetCpu() != nil && len(resource.GetCpu().GetDeviceQuota()) > 0 {
			cpu, errLocal := uquota.ResourceCpuCompute(addOrSub, originQuota.Quota.Cpu.DeviceQuota,
				resource.GetCpu().GetDeviceQuota())
			if errLocal != nil {
				return errLocal
			}

			originQuota.Quota.Cpu.DeviceQuota = cpu
		}
		if resource.GetMem() != nil && len(resource.GetMem().GetDeviceQuota()) > 0 {
			mem, errLocal := uquota.ResourceMemoryCompute(addOrSub, originQuota.Quota.Mem.DeviceQuota,
				resource.GetMem().GetDeviceQuota())
			if errLocal != nil {
				return errLocal
			}
			originQuota.Quota.Mem.DeviceQuota = mem
		}

		if resource.GetGpu() != nil && len(resource.GetGpu().GetDeviceQuota()) > 0 {
			mem, errLocal := uquota.ResourceCpuCompute(addOrSub, originQuota.Quota.Gpu.DeviceQuota,
				resource.GetGpu().GetDeviceQuota())
			if errLocal != nil {
				return errLocal
			}
			originQuota.Quota.Gpu.DeviceQuota = mem
		}
	}

	return store.GetModel().UpdateProjectQuota(context.Background(), originQuota)
}
