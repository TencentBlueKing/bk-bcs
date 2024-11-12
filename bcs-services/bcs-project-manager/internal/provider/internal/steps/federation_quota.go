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

// Package steps xxx
package steps

import (
	"encoding/json"
	"fmt"

	common_task "github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

var (
	// OperationCreate "create"
	OperationCreate = "create"
	// OperationDelete "delete"
	OperationDelete = "delete"
	// OperationUpdate "update"
	OperationUpdate = "update"
	// ScaleUp "up"
	ScaleUp = "up"
	// ScaleDown "down"
	ScaleDown = "down"
)

const (
	federationQuotaStepName   = "创建联邦配额对象"
	federationQuotaStepMethod = "federation-quota"
)

// NewFederationQuotaStep federation quota step
func NewFederationQuotaStep() common_task.StepBuilder {
	return &federationQuotaStep{}
}

// federationQuotaStep itsm approve step
type federationQuotaStep struct{}

// GetName stepName
func (f federationQuotaStep) GetName() string {
	return federationQuotaStepMethod
}

// Alias method name
func (f federationQuotaStep) Alias() string {
	return federationQuotaStepMethod
}

func (f federationQuotaStep) getParams(task *types.Task) (*FederationQuotaStepParams, error) {
	step, ok := task.GetStep(f.Alias())
	if !ok {
		return nil, fmt.Errorf("task[%s] step[%s] not exist", task.GetTaskID(), f.GetName())
	}
	quotaParams, ok := step.GetParam(utils.FederationQuotaDataKey.String())
	if !ok {
		return nil, fmt.Errorf("task[%s] step[%s] user empty", task.GetTaskID(), f.GetName())
	}

	var federationQuotaParams FederationQuotaStepParams
	err := json.Unmarshal([]byte(quotaParams), &federationQuotaParams)
	if err != nil {
		return nil, fmt.Errorf("task[%s] step[%s] unmarshal params failed, %s", task.GetTaskID(), f.GetName(),
			err.Error())
	}

	return &federationQuotaParams, nil
}

// DoWork for worker exec task
func (f federationQuotaStep) DoWork(task *types.Task) error {
	// 属性一致的联邦quota 对象创建/更新/删除
	params, err := f.getParams(task)
	if err != nil {
		return err
	}

	// 调用联邦接口操作集群中的联邦quota对象
	switch params.Operation {
	case OperationCreate:
		return f.createFederationQuota(task, params)
	case OperationDelete:
		return f.deleteFederationQuota(task, params)
	case OperationUpdate:
		return f.updateFederationQuota(task, params)
	}

	return nil
}

func (f federationQuotaStep) createFederationQuota(task *types.Task, quota *FederationQuotaStepParams) error {
	// 调用联邦接口创建联邦quota对象
	logging.Info("createFederationQuota[%s] %s %s success", task.GetTaskID(), quota.NameSpace, quota.Name)

	return nil
}

func (f federationQuotaStep) updateFederationQuota(task *types.Task, quota *FederationQuotaStepParams) error {
	// 调用联邦接口更新联邦quota对象
	logging.Info("updateFederationQuota[%s] %s %s success", task.GetTaskID(), quota.NameSpace, quota.Name)

	// 更新联邦
	switch quota.Scale {
	case ScaleUp:
	case ScaleDown:
	}

	return nil
}

func (f federationQuotaStep) deleteFederationQuota(task *types.Task, quota *FederationQuotaStepParams) error {
	// 调用联邦接口删除联邦quota对象
	logging.Info("deleteFederationQuota[%s] %s %s success", task.GetTaskID(), quota.NameSpace, quota.Name)

	return nil
}

// BuildStep build step
func (f federationQuotaStep) BuildStep(kvs []common_task.KeyValue, opts ...types.StepOption) *types.Step {
	step := types.NewStep(f.GetName(), f.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}

// FederationQuotaStepParams xxx
type FederationQuotaStepParams struct {
	Operation string                 `json:"operation"`
	Scale     string                 `json:"scale"`
	QuotaId   string                 `json:"quota_id"`
	ClusterId string                 `json:"cluster_id"`
	NameSpace string                 `json:"namespace"`
	Name      string                 `json:"name"`
	Cpu       *bcsproject.DeviceInfo `json:"cpu"`
	Mem       *bcsproject.DeviceInfo `json:"mem"`
	Gpu       *bcsproject.DeviceInfo `json:"gpu"`
}

// validate validate xxx
func (fqs FederationQuotaStepParams) validate() error {
	if !stringx.StringInSlice(fqs.Operation, []string{OperationCreate, OperationUpdate, OperationDelete}) {
		return fmt.Errorf("operation %s is invalid", fqs.Operation)
	}

	if fqs.NameSpace == "" || fqs.Name == "" {
		return fmt.Errorf("namespace or name is empty")
	}

	switch fqs.Operation {
	case OperationCreate:
		if fqs.Cpu == nil && fqs.Mem == nil && fqs.Gpu == nil {
			return fmt.Errorf("cpu or mem or gpu is empty")
		}
		return nil
	case OperationDelete:
		if fqs.QuotaId == "" {
			return fmt.Errorf("namespace or name is empty")
		}
		return nil
	case OperationUpdate:
		if fqs.QuotaId == "" {
			return fmt.Errorf("namespace or name is empty")
		}
		if fqs.Cpu == nil && fqs.Mem == nil && fqs.Gpu == nil {
			return fmt.Errorf("cpu or mem or gpu is empty")
		}

		return nil
	}

	return nil
}

// BuildParams build kvs
func (fqs FederationQuotaStepParams) BuildParams() ([]common_task.KeyValue, error) {
	kvs := make([]common_task.KeyValue, 0)

	err := fqs.validate()
	if err != nil {
		return nil, err
	}

	quotaParams, err := json.Marshal(fqs)
	if err != nil {
		return nil, err
	}

	kvs = append(kvs, common_task.KeyValue{
		Key:   utils.FederationQuotaDataKey,
		Value: string(quotaParams),
	})

	return kvs, nil
}
