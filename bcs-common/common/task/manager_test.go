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

// Package task is a package for task management
package task

import (
	"testing"

	"github.com/stretchr/testify/assert"

	hellostep "github.com/Tencent/bk-bcs/bcs-common/common/task/steps/hello"
	istep "github.com/Tencent/bk-bcs/bcs-common/common/task/steps/iface"
)

func TestDoWork(t *testing.T) {
	mgr := NewTaskManager()
	cfg := &ManagerConfig{
		ModuleName: "test",
		StepWorkers: []istep.StepWorkerInterface{
			&hellostep.Hello{},
		},
	}
	err := mgr.Init(cfg)
	assert.NoError(t, err)

	err = mgr.doWork("1", "1")
	assert.NoError(t, err)
}
