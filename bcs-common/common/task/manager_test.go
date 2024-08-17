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
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hellostep "github.com/Tencent/bk-bcs/bcs-common/common/task/steps/hello"
	istep "github.com/Tencent/bk-bcs/bcs-common/common/task/steps/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/stores/mem"
	mysqlstore "github.com/Tencent/bk-bcs/bcs-common/common/task/stores/mysql"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

func TestDoWork(t *testing.T) {
	// 使用结构体注册
	// istep.Register("hello", hellostep.NewHello())

	// 使用函数注册
	// istep.Register("sum", istep.StepWorkerFunc(hellostep.Sum))

	mgr := TaskManager{
		ctx:           context.Background(),
		store:         mem.New(),
		stepExecutors: istep.GetRegisters(),
	}
	mgr.initGlobalStorage()

	info := types.TaskInfo{
		TaskType: "example-test",
		TaskName: "example",
		Creator:  "bcs",
	}

	steps := []*types.Step{
		types.NewStep("hello", "test"),
		types.NewStep("sum", "test1").AddParam(hellostep.SumA.String(), "1").AddParam(hellostep.SumB.String(), "2"),
	}

	task := types.NewTask(info)
	task.Steps = steps

	require.NoError(t, GetGlobalStorage().CreateTask(context.Background(), task))

	for _, s := range steps {
		err := mgr.doWork(task.TaskID, s.Name)
		assert.ErrorIs(t, err, types.ErrNotImplemented)
	}
}

func TestDoWorkWithMySQL(t *testing.T) {
	if os.Getenv("MYSQL_DSN") == "" {
		t.Skip("skip test without mysql dsn")
	}

	// 使用结构体注册
	istep.Register("hello", hellostep.NewHello())

	// 使用函数注册
	istep.Register("sum", istep.StepExecutorFunc(hellostep.Sum))

	store, err := mysqlstore.New(os.Getenv("MYSQL_DSN"))
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, store.EnsureTable(ctx))

	mgr := TaskManager{
		ctx:           context.Background(),
		store:         store,
		stepExecutors: istep.GetRegisters(),
	}
	mgr.initGlobalStorage()

	info := types.TaskInfo{
		TaskType: "example-test",
		TaskName: "example",
		Creator:  "bcs",
	}

	steps := []*types.Step{
		types.NewStep("hello", "test"),
		types.NewStep("sum", "test1").AddParam(hellostep.SumA.String(), "1").AddParam(hellostep.SumB.String(), "2"),
	}

	task := types.NewTask(info)
	task.Steps = steps

	require.NoError(t, GetGlobalStorage().CreateTask(context.Background(), task))

	for _, s := range steps {
		err := mgr.doWork(task.TaskID, s.Name)
		assert.NoError(t, err)
	}
}
