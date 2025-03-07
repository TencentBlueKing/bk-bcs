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

package mongo

import (
	"context"
	"testing"
	"time"

	bcsmongo "github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/stores/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

type TaskTestSuite struct {
	suite.Suite
	db        *bcsmongo.DB
	store     iface.Store
	tableName string
	ctx       context.Context
}

func (s *TaskTestSuite) SetupSuite() {
	s.tableName = "test"
	s.ctx = context.Background()

	// 创建 MongoDB 连接
	db, err := bcsmongo.NewDB(&bcsmongo.Options{
		Database: s.tableName,
		Hosts:    []string{"localhost:27017"},
		Username: "root",
		Password: "bcs",
	})
	if err != nil {
		s.T().Fatal(err)
	}
	s.db = db

	// 创建存储实例
	s.store = New(db, s.tableName)
}

func (s *TaskTestSuite) TearDownSuite() {
	s.db.Close()
}

func (s *TaskTestSuite) SetupTest() {
	// 每个测试前清理数据
	err := s.db.DropTable(s.ctx, s.tableName+"_task")
	if err != nil {
		s.T().Logf("drop table failed: %v", err)
	}
}

func (s *TaskTestSuite) createTestTask() *types.Task {
	now := time.Now()
	taskID := uuid.New().String()
	return &types.Task{
		TaskID:   taskID,
		TaskType: "test-type",
		TaskName: "test-name",
		Status:   "pending",
		Creator:  "test-user",
		Start:    now,
		End:      now,
		Steps: []*types.Step{
			{
				Name:   "step1",
				Status: "pending",
				Start:  now,
				End:    now,
			},
		},
	}
}

func (s *TaskTestSuite) TestCreateAndGetTask() {
	task := s.createTestTask()
	s.T().Logf("creating task: %+v", task)

	// 测试创建任务
	err := s.store.CreateTask(s.ctx, task)
	if err != nil {
		s.T().Logf("create task error: %v", err)
		s.FailNow("create task failed")
	}

	// 测试获取任务
	getTask, err := s.store.GetTask(s.ctx, task.TaskID)
	if err != nil {
		s.T().Logf("get task error: %v", err)
		s.FailNow("get task failed")
	}
	s.T().Logf("got task: %+v", getTask)
	s.Equal(task.TaskID, getTask.TaskID)
	s.Equal(task.TaskType, getTask.TaskType)
	s.Equal(task.TaskName, getTask.TaskName)
	s.Equal(task.Status, getTask.Status)
	s.Equal(task.Creator, getTask.Creator)
	s.Equal(len(task.Steps), len(getTask.Steps))
}

func (s *TaskTestSuite) TestCreateAndUpdateTask() {
	task := s.createTestTask()
	s.T().Logf("creating task: %+v", task)

	// 测试创建任务
	err := s.store.CreateTask(s.ctx, task)
	if err != nil {
		s.T().Logf("create task error: %v", err)
		s.FailNow("create task failed")
	}

	// 验证任务创建成功
	getTask, err := s.store.GetTask(s.ctx, task.TaskID)
	if err != nil {
		s.T().Logf("get task error: %v", err)
		s.FailNow("get task failed")
	}
	s.Equal(task.TaskID, getTask.TaskID)
	s.Equal("pending", getTask.Status)

	// 测试更新任务
	task.Status = "running"
	task.Message = "task is running"
	s.T().Logf("updating task: %+v", task)
	err = s.store.UpdateTask(s.ctx, task)
	if err != nil {
		s.T().Logf("update task error: %v", err)
		s.FailNow("update task failed")
	}

	// 验证更新结果
	updatedTask, err := s.store.GetTask(s.ctx, task.TaskID)
	if err != nil {
		s.T().Logf("get task error: %v", err)
		s.FailNow("get task failed")
	}
	s.T().Logf("updated task: %+v", updatedTask)
	s.Equal("running", updatedTask.Status)
	s.Equal("task is running", updatedTask.Message)
}

func (s *TaskTestSuite) TestListTask() {
	task := s.createTestTask()
	s.T().Logf("creating task: %+v", task)

	// 先创建任务
	err := s.store.CreateTask(s.ctx, task)
	if err != nil {
		s.T().Logf("create task error: %v", err)
		s.FailNow("create task failed")
	}

	// 测试空条件查询
	emptyOpt := &iface.ListOption{
		Offset: 0,
		Limit:  10,
	}
	s.T().Logf("listing tasks with empty option: %+v", emptyOpt)
	pagination, err := s.store.ListTask(s.ctx, emptyOpt)
	if err != nil {
		s.T().Logf("list task error: %v", err)
		s.FailNow("list task failed")
	}
	s.T().Logf("got tasks: %+v", pagination)
	s.Equal(int64(1), pagination.Count)
	s.Equal(1, len(pagination.Items))

	// 测试带条件查询
	listOpt := &iface.ListOption{
		TaskType: "test-type",
		Status:   "pending",
		Creator:  "test-user",
		Offset:   0,
		Limit:    10,
	}
	s.T().Logf("listing tasks with option: %+v", listOpt)
	pagination, err = s.store.ListTask(s.ctx, listOpt)
	if err != nil {
		s.T().Logf("list task error: %v", err)
		s.FailNow("list task failed")
	}
	s.T().Logf("got tasks: %+v", pagination)
	s.Equal(int64(1), pagination.Count)
	s.Equal(1, len(pagination.Items))

	// 测试不匹配条件
	notMatchOpt := &iface.ListOption{
		TaskType: "not-exist",
		Offset:   0,
		Limit:    10,
	}
	s.T().Logf("listing tasks with not match option: %+v", notMatchOpt)
	pagination, err = s.store.ListTask(s.ctx, notMatchOpt)
	if err != nil {
		s.T().Logf("list task error: %v", err)
		s.FailNow("list task failed")
	}
	s.T().Logf("got tasks: %+v", pagination)
	s.Equal(int64(0), pagination.Count)
	s.Equal(0, len(pagination.Items))
}

func (s *TaskTestSuite) TestCreateAndDeleteTask() {
	task := s.createTestTask()
	s.T().Logf("creating task: %+v", task)

	// 测试创建任务
	err := s.store.CreateTask(s.ctx, task)
	if err != nil {
		s.T().Logf("create task error: %v", err)
		s.FailNow("create task failed")
	}

	// 验证任务创建成功
	getTask, err := s.store.GetTask(s.ctx, task.TaskID)
	if err != nil {
		s.T().Logf("get task error: %v", err)
		s.FailNow("get task failed")
	}
	s.Equal(task.TaskID, getTask.TaskID)

	// 测试删除任务
	s.T().Logf("deleting task: %s", task.TaskID)
	err = s.store.DeleteTask(s.ctx, task.TaskID)
	if err != nil {
		s.T().Logf("delete task error: %v", err)
		s.FailNow("delete task failed")
	}

	// 验证任务已被删除
	_, err = s.store.GetTask(s.ctx, task.TaskID)
	s.Error(err, "task should be deleted")
}

func (s *TaskTestSuite) TestListTasks() {
	// 创建多个任务
	task1 := s.createTestTask()
	task1.TaskID = "test-task-1"
	task1.Status = "pending"

	task2 := s.createTestTask()
	task2.TaskID = "test-task-2"
	task2.Status = "running"
	task2.Creator = "another-user"

	task3 := s.createTestTask()
	task3.TaskID = "test-task-3"
	task3.TaskType = "another-type"
	task3.Status = "pending"

	// 创建任务
	for _, task := range []*types.Task{task1, task2, task3} {
		err := s.store.CreateTask(s.ctx, task)
		if err != nil {
			s.T().Logf("create task error: %v", err)
			s.FailNow("create task failed")
		}
	}

	// 测试空条件查询
	s.T().Log("testing empty condition list")
	emptyOpt := &iface.ListOption{
		Offset: 0,
		Limit:  10,
	}
	pagination, err := s.store.ListTask(s.ctx, emptyOpt)
	if err != nil {
		s.T().Logf("list task error: %v", err)
		s.FailNow("list task failed")
	}
	s.Equal(int64(3), pagination.Count)
	s.Equal(3, len(pagination.Items))

	// 测试按状态查询
	s.T().Log("testing status condition list")
	statusOpt := &iface.ListOption{
		Status: "pending",
		Offset: 0,
		Limit:  10,
	}
	pagination, err = s.store.ListTask(s.ctx, statusOpt)
	if err != nil {
		s.T().Logf("list task error: %v", err)
		s.FailNow("list task failed")
	}
	s.Equal(int64(2), pagination.Count)
	s.Equal(2, len(pagination.Items))

	// 测试按类型查询
	s.T().Log("testing type condition list")
	typeOpt := &iface.ListOption{
		TaskType: "another-type",
		Offset:   0,
		Limit:    10,
	}
	pagination, err = s.store.ListTask(s.ctx, typeOpt)
	if err != nil {
		s.T().Logf("list task error: %v", err)
		s.FailNow("list task failed")
	}
	s.Equal(int64(1), pagination.Count)
	s.Equal(1, len(pagination.Items))

	// 测试组合条件查询
	s.T().Log("testing combined condition list")
	combinedOpt := &iface.ListOption{
		TaskType: "test-type",
		Status:   "pending",
		Creator:  "test-user",
		Offset:   0,
		Limit:    10,
	}
	pagination, err = s.store.ListTask(s.ctx, combinedOpt)
	if err != nil {
		s.T().Logf("list task error: %v", err)
		s.FailNow("list task failed")
	}
	s.Equal(int64(1), pagination.Count)
	s.Equal(1, len(pagination.Items))
	s.Equal("test-task-1", pagination.Items[0].TaskID)

	// 测试分页
	s.T().Log("testing pagination")
	pageOpt := &iface.ListOption{
		Offset: 1,
		Limit:  1,
	}
	pagination, err = s.store.ListTask(s.ctx, pageOpt)
	if err != nil {
		s.T().Logf("list task error: %v", err)
		s.FailNow("list task failed")
	}
	s.Equal(int64(3), pagination.Count) // 总数应该是3
	s.Equal(1, len(pagination.Items))   // 但是只返回1条数据
}

func (s *TaskTestSuite) TestListTasksWithConditions() {
	// 创建多个不同条件的任务
	tasks := []*types.Task{
		{
			TaskID:   "task-1",
			TaskType: "type-1",
			TaskName: "name-1",
			Status:   "pending",
			Creator:  "user-1",
			Start:    time.Now().Add(-2 * time.Hour),
			End:      time.Now().Add(-1 * time.Hour),
			Steps: []*types.Step{{
				Name:   "step1",
				Status: "pending",
			}},
		},
		{
			TaskID:   "task-2",
			TaskType: "type-1",
			TaskName: "name-2",
			Status:   "running",
			Creator:  "user-2",
			Start:    time.Now().Add(-1 * time.Hour),
			End:      time.Now(),
			Steps: []*types.Step{{
				Name:   "step1",
				Status: "running",
			}},
		},
		{
			TaskID:   "task-3",
			TaskType: "type-2",
			TaskName: "name-3",
			Status:   "pending",
			Creator:  "user-1",
			Start:    time.Now(),
			End:      time.Now().Add(1 * time.Hour),
			Steps: []*types.Step{{
				Name:   "step1",
				Status: "pending",
			}},
		},
	}

	// 创建任务
	for _, task := range tasks {
		err := s.store.CreateTask(s.ctx, task)
		if err != nil {
			s.T().Logf("create task error: %v", err)
			s.FailNow("create task failed")
		}
	}

	// 测试用例
	testCases := []struct {
		name     string
		opt      *iface.ListOption
		expected struct {
			count    int64
			taskIDs  []string
			notFound []string
		}
	}{
		{
			name: "按任务类型查询",
			opt: &iface.ListOption{
				TaskType: "type-1",
			},
			expected: struct {
				count    int64
				taskIDs  []string
				notFound []string
			}{
				count:    2,
				taskIDs:  []string{"task-1", "task-2"},
				notFound: []string{"task-3"},
			},
		},
		{
			name: "按状态查询",
			opt: &iface.ListOption{
				Status: "pending",
			},
			expected: struct {
				count    int64
				taskIDs  []string
				notFound []string
			}{
				count:    2,
				taskIDs:  []string{"task-1", "task-3"},
				notFound: []string{"task-2"},
			},
		},
		{
			name: "按创建者查询",
			opt: &iface.ListOption{
				Creator: "user-1",
			},
			expected: struct {
				count    int64
				taskIDs  []string
				notFound []string
			}{
				count:    2,
				taskIDs:  []string{"task-1", "task-3"},
				notFound: []string{"task-2"},
			},
		},
		{
			name: "组合条件查询：类型和状态",
			opt: &iface.ListOption{
				TaskType: "type-1",
				Status:   "pending",
			},
			expected: struct {
				count    int64
				taskIDs  []string
				notFound []string
			}{
				count:    1,
				taskIDs:  []string{"task-1"},
				notFound: []string{"task-2", "task-3"},
			},
		},
		{
			name: "组合条件查询：创建者和状态",
			opt: &iface.ListOption{
				Creator: "user-1",
				Status:  "pending",
			},
			expected: struct {
				count    int64
				taskIDs  []string
				notFound []string
			}{
				count:    2,
				taskIDs:  []string{"task-1", "task-3"},
				notFound: []string{"task-2"},
			},
		},
		{
			name: "分页查询：第一页",
			opt: &iface.ListOption{
				Offset: 0,
				Limit:  2,
			},
			expected: struct {
				count    int64
				taskIDs  []string
				notFound []string
			}{
				count:   3,                            // 总数应该是3
				taskIDs: []string{"task-3", "task-2"}, // 按开始时间倒序
			},
		},
		{
			name: "分页查询：第二页",
			opt: &iface.ListOption{
				Offset: 2,
				Limit:  2,
			},
			expected: struct {
				count    int64
				taskIDs  []string
				notFound []string
			}{
				count:   3,                  // 总数应该是3
				taskIDs: []string{"task-1"}, // 最后一条数据
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.T().Logf("testing case: %s", tc.name)
			pagination, err := s.store.ListTask(s.ctx, tc.opt)
			if err != nil {
				s.T().Logf("list task error: %v", err)
				s.FailNow("list task failed")
			}

			// 验证总数
			s.Equal(tc.expected.count, pagination.Count, "total count should match")
			s.Equal(len(tc.expected.taskIDs), len(pagination.Items), "returned items count should match")

			// 验证返回的任务ID
			actualIDs := make([]string, 0, len(pagination.Items))
			for _, task := range pagination.Items {
				actualIDs = append(actualIDs, task.TaskID)
			}
			s.ElementsMatch(tc.expected.taskIDs, actualIDs, "returned task IDs should match")

			// 验证未找到的任务ID
			for _, notFoundID := range tc.expected.notFound {
				found := false
				for _, task := range pagination.Items {
					if task.TaskID == notFoundID {
						found = true
						break
					}
				}
				s.False(found, "task %s should not be found", notFoundID)
			}
		})
	}
}

func TestTaskSuite(t *testing.T) {
	suite.Run(t, new(TaskTestSuite))
}
