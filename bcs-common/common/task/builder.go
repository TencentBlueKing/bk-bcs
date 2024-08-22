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

package task

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

// NewByTaskBuilder init task from builder
func NewByTaskBuilder(builder types.TaskBuilder, opts ...types.TaskOption) (*types.Task, error) {
	// 声明step
	steps, err := builder.Steps()
	if err != nil {
		return nil, err
	}

	if len(steps) == 0 {
		return nil, fmt.Errorf("task steps empty")
	}

	task := types.NewTask(builder.TaskInfo(), opts...)
	task.Steps = steps
	task.CurrentStep = steps[0].GetName()

	// 自定义extraJson等
	newTask, err := builder.BuildTask(*task)
	if err != nil {
		return nil, err
	}

	return &newTask, nil
}
