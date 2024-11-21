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
	"context"
	"fmt"
	"strconv"
	"time"

	commontask "github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/utils"
)

const (
	waitStepName   = "等待执行"
	waitStepMethod = "loop-wait"
)

// WaitType xxx
type WaitType string

// String xxx
func (wy WaitType) String() string {
	return string(wy)
}

var (
	// WaitTime 当前时间等待多久后执行,单位时间戳
	WaitTime WaitType = "wait_time"
	// WaitTimePoint 在某个时间点执行, 单位时间戳
	WaitTimePoint WaitType = "wait_time_point"
)

// NewWaitStep quota manager itsm approve step
func NewWaitStep() commontask.StepBuilder {
	return &waitStep{}
}

// waitStep wait step
type waitStep struct{}

// GetName stepName
func (s waitStep) GetName() string {
	return waitStepMethod
}

// Alias method name
func (s waitStep) Alias() string {
	return waitStepMethod
}

func (s waitStep) getStepParams(task *types.Task) (*WaitStepParams, error) {
	step, ok := task.GetStep(s.Alias())
	if !ok {
		return nil, fmt.Errorf("task[%s] step[%s] not exist", task.GetTaskID(), s.GetName())
	}
	waitType, ok := step.GetParam(utils.WaitTypeKey.String())
	if !ok {
		return nil, fmt.Errorf("task[%s] step[%s] user empty", task.GetTaskID(), s.GetName())
	}
	waitTime, ok := step.GetParam(utils.WaitTimeKey.String())
	if !ok {
		return nil, fmt.Errorf("task[%s] step[%s] project empty", task.GetTaskID(), s.GetName())
	}

	stepParams := &WaitStepParams{
		WaitType: WaitType(waitType),
		WaitTime: waitTime,
	}

	endTimeUnix, ok := step.GetParam(utils.EndWaitTimeKey.String())
	if !ok {
		stepParams.EndTimeUnix = ""
	} else {
		stepParams.EndTimeUnix = endTimeUnix
	}

	return stepParams, nil
}

// DoWork for worker exec task
func (s waitStep) DoWork(task *types.Task) error {
	waitParams, err := s.getStepParams(task)
	if err != nil {
		return err
	}

	var (
		endUnix  int64
		timeUnix int64
	)

	if waitParams.EndTimeUnix == "" {
		timeUnix, err = strconv.ParseInt(waitParams.WaitTime, 10, 64)
		if err != nil {
			return err
		}
		switch waitParams.WaitType {
		case WaitTime:
			endUnix = time.Now().Unix() + timeUnix
		case WaitTimePoint:
			endUnix = timeUnix
		default:
			return fmt.Errorf("waitStep[%s] not supported waitType %s", task.GetTaskID(), waitParams.WaitType.String())
		}

		// 写入并更新step信息
	} else {
		endUnix, err = strconv.ParseInt(waitParams.EndTimeUnix, 10, 64)
		if err != nil {
			return err
		}
	}

	// 任务等待超时时间
	err = commontask.LoopDoFunc(context.Background(), func() error {

		logging.Info("waitStep[%s] current unix %v:%v", task.GetTaskID(), time.Now().Unix(), endUnix)

		if time.Now().Unix() > endUnix {
			return commontask.ErrEndLoop
		}
		return nil
	}, commontask.LoopInterval(20*time.Second))

	if err != nil {
		logging.Error("waitStep[%s] loop failed, err: %s",
			task.GetTaskID(), err.Error())
		return err
	}

	logging.Info("waitStep[%s] loop success", task.GetTaskID())

	return nil
}

// BuildStep build step
func (s waitStep) BuildStep(kvs []commontask.KeyValue, opts ...types.StepOption) *types.Step {
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}

// WaitStepParams xxx
type WaitStepParams struct {
	WaitType    WaitType
	WaitTime    string
	EndTimeUnix string
}

// TransWaitStepParamsToKeyValue 转换wait参数为key value
func TransWaitStepParamsToKeyValue(params WaitStepParams) []commontask.KeyValue {
	kvs := make([]commontask.KeyValue, 0)

	kvs = append(kvs, commontask.KeyValue{
		Key:   utils.WaitTypeKey,
		Value: params.WaitType.String(),
	})
	kvs = append(kvs, commontask.KeyValue{
		Key:   utils.WaitTimeKey,
		Value: params.WaitTime,
	})

	return kvs
}
