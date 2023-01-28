/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mongo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage/mongo/mocks"
)

func TestModelTask_CreateTask(t *testing.T) {
	task := &storage.ScaleDownTask{
		TaskID:            "taskID1",
		TotalNum:          2,
		NodeGroupStrategy: "nodegroup1",
		ScaleDownGroups:   []*storage.ScaleDownDetail{},
		DrainDelayDays:    2,
		Deadline:          time.Now().Add(48 * time.Hour),
		CreatedTime:       time.Now(),
		UpdatedTime:       time.Now(),
		IsDelete:          false,
		IsExecuted:        false,
		Status:            "preparing",
	}
	tests := []struct {
		name    string
		task    *storage.ScaleDownTask
		opt     *storage.CreateOptions
		wantErr bool
		on      func(mockFields *MockFields)
	}{
		{
			name: "normal",
			task: task,
			opt: &storage.CreateOptions{
				OverWriteIfExist: false,
			},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+taskTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+taskTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "task_1").Return(true, nil)
				mockFields.table.On("HasIndex", context.Background(), "node_group_strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), mock.Anything).Return(drivers.ErrTableRecordNotFound)
				mockFields.table.On("Insert", context.Background(), mock.Anything).Return(1, nil)
			},
			wantErr: false,
		},
		{
			name: "existErr",
			task: task,
			opt: &storage.CreateOptions{
				OverWriteIfExist: false,
			},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+taskTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+taskTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "task_1").Return(true, nil)
				mockFields.table.On("HasIndex", context.Background(), "node_group_strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), mock.Anything).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "overwrite",
			task: task,
			opt: &storage.CreateOptions{
				OverWriteIfExist: true,
			},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+taskTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+taskTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "task_1").Return(true, nil)
				mockFields.table.On("HasIndex", context.Background(), "node_group_strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), mock.Anything).Return(nil)
				mockFields.table.On("Update", context.Background(), mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockFields{
				db:    mocks.NewDB(t),
				table: mocks.NewTable(t),
				find:  mocks.NewFind(t),
			}
			tt.on(mockDB)
			server := NewServer(mockDB.db)
			err := server.CreateTask(tt.task, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestModelTask_DeleteTask(t *testing.T) {
	task := &storage.ScaleDownTask{
		TaskID:            "taskID1",
		TotalNum:          2,
		NodeGroupStrategy: "nodegroup1",
		ScaleDownGroups:   []*storage.ScaleDownDetail{},
		DrainDelayDays:    2,
		Deadline:          time.Now().Add(48 * time.Hour),
		CreatedTime:       time.Now(),
		UpdatedTime:       time.Now(),
		IsDelete:          false,
		IsExecuted:        false,
		Status:            "preparing",
	}
	tests := []struct {
		name    string
		taskID  string
		opt     *storage.DeleteOptions
		wantErr bool
		want    *storage.ScaleDownTask
		on      func(mockFields *MockFields)
	}{
		{
			name:    "normal",
			taskID:  "taskID1",
			opt:     &storage.DeleteOptions{ErrIfNotExist: false},
			want:    task,
			wantErr: false,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+taskTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+taskTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "task_1").Return(true, nil)
				mockFields.table.On("HasIndex", context.Background(), "node_group_strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.ScaleDownTask{}).
					Return(func(ctx context.Context, result interface{}) error {
						return reflectInterface(result, *task)
					})
				mockFields.table.On("Update", context.Background(), mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "notExist",
			taskID:  "taskID1",
			opt:     &storage.DeleteOptions{ErrIfNotExist: false},
			wantErr: false,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+taskTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+taskTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "task_1").Return(true, nil)
				mockFields.table.On("HasIndex", context.Background(), "node_group_strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.ScaleDownTask{}).Return(drivers.ErrTableRecordNotFound)
			},
		},
		{
			name:    "err",
			taskID:  "taskID1",
			opt:     &storage.DeleteOptions{ErrIfNotExist: false},
			wantErr: true,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+taskTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+taskTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "task_1").Return(true, nil)
				mockFields.table.On("HasIndex", context.Background(), "node_group_strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.ScaleDownTask{}).Return(fmt.Errorf("db error"))
			},
		},
		{
			name:    "notExistErr",
			taskID:  "taskID1",
			opt:     &storage.DeleteOptions{ErrIfNotExist: true},
			wantErr: true,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+taskTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+taskTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "task_1").Return(true, nil)
				mockFields.table.On("HasIndex", context.Background(), "node_group_strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.ScaleDownTask{}).Return(drivers.ErrTableRecordNotFound)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockFields{
				db:    mocks.NewDB(t),
				table: mocks.NewTable(t),
				find:  mocks.NewFind(t),
			}
			tt.on(mockDB)
			server := NewServer(mockDB.db)
			rsp, err := server.DeleteTask(tt.taskID, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, rsp)
		})
	}
}

func TestModelTask_GetTask(t *testing.T) {
	task := &storage.ScaleDownTask{
		TaskID:            "taskID1",
		TotalNum:          2,
		NodeGroupStrategy: "nodegroup1",
		ScaleDownGroups:   []*storage.ScaleDownDetail{},
		DrainDelayDays:    2,
		Deadline:          time.Now().Add(48 * time.Hour),
		CreatedTime:       time.Now(),
		UpdatedTime:       time.Now(),
		IsDelete:          false,
		IsExecuted:        false,
		Status:            "preparing",
	}
	tests := []struct {
		name    string
		taskID  string
		opt     *storage.GetOptions
		want    *storage.ScaleDownTask
		wantErr bool
		on      func(mockFields *MockFields)
	}{
		{
			name:   "normal",
			taskID: "taskID1",
			opt: &storage.GetOptions{
				ErrIfNotExist:  false,
				GetSoftDeleted: false,
			},
			want:    task,
			wantErr: false,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+taskTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+taskTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "task_1").Return(true, nil)
				mockFields.table.On("HasIndex", context.Background(), "node_group_strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.ScaleDownTask{}).
					Return(func(ctx context.Context, result interface{}) error {
						return reflectInterface(result, *task)
					})
			},
		},
		{
			name:   "notExist",
			taskID: "taskID1",
			opt: &storage.GetOptions{
				ErrIfNotExist:  false,
				GetSoftDeleted: false,
			},
			wantErr: false,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+taskTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+taskTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "task_1").Return(true, nil)
				mockFields.table.On("HasIndex", context.Background(), "node_group_strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), mock.Anything).Return(drivers.ErrTableRecordNotFound)
			},
		},
		{
			name:   "errNotExist",
			taskID: "taskID1",
			opt: &storage.GetOptions{
				ErrIfNotExist:  true,
				GetSoftDeleted: false,
			},
			wantErr: true,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+taskTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+taskTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "task_1").Return(true, nil)
				mockFields.table.On("HasIndex", context.Background(), "node_group_strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), mock.Anything).Return(drivers.ErrTableRecordNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockFields{
				db:    mocks.NewDB(t),
				table: mocks.NewTable(t),
				find:  mocks.NewFind(t),
			}
			tt.on(mockDB)
			server := NewServer(mockDB.db)
			ret, err := server.GetTask(tt.taskID, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, ret)
		})
	}

}

func TestModelTask_ListTasks(t *testing.T) {
	task := &storage.ScaleDownTask{
		TaskID:            "taskID1",
		TotalNum:          2,
		NodeGroupStrategy: "nodegroup1",
		ScaleDownGroups:   []*storage.ScaleDownDetail{},
		DrainDelayDays:    2,
		Deadline:          time.Now().Add(48 * time.Hour),
		CreatedTime:       time.Now(),
		UpdatedTime:       time.Now(),
		IsDelete:          false,
		IsExecuted:        false,
		Status:            "preparing",
	}
	tests := []struct {
		name    string
		opt     *storage.ListOptions
		wantErr bool
		want    []*storage.ScaleDownTask
		on      func(mockFields *MockFields)
	}{
		{
			name: "normal",
			opt: &storage.ListOptions{
				Limit:                  1,
				Page:                   0,
				ReturnSoftDeletedItems: false,
			},
			wantErr: false,
			want:    []*storage.ScaleDownTask{task},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+taskTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+taskTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "task_1").Return(true, nil)
				mockFields.table.On("HasIndex", context.Background(), "node_group_strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithSort", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithStart", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithLimit", mock.Anything).Return(mockFields.find)
				mockFields.find.On("All", context.Background(), mock.Anything).Return(func(ctx context.Context, result interface{}) error {
					return reflectInterface(result, []*storage.ScaleDownTask{task})
				})
			},
		},
		{
			name: "err",
			opt: &storage.ListOptions{
				Limit:                  1,
				Page:                   0,
				ReturnSoftDeletedItems: false,
			},
			wantErr: true,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+taskTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+taskTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "task_1").Return(true, nil)
				mockFields.table.On("HasIndex", context.Background(), "node_group_strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithSort", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithStart", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithLimit", mock.Anything).Return(mockFields.find)
				mockFields.find.On("All", context.Background(), mock.Anything).Return(fmt.Errorf("db err"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockFields{
				db:    mocks.NewDB(t),
				table: mocks.NewTable(t),
				find:  mocks.NewFind(t),
			}
			tt.on(mockDB)
			server := NewServer(mockDB.db)
			rsp, err := server.ListTasks(tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, rsp)
		})
	}
}

func TestModelTask_ListTasksByStrategy(t *testing.T) {
	task := &storage.ScaleDownTask{
		TaskID:            "taskID1",
		TotalNum:          2,
		NodeGroupStrategy: "nodegroup1",
		ScaleDownGroups:   []*storage.ScaleDownDetail{},
		DrainDelayDays:    2,
		Deadline:          time.Now().Add(48 * time.Hour),
		CreatedTime:       time.Now(),
		UpdatedTime:       time.Now(),
		IsDelete:          false,
		IsExecuted:        false,
		Status:            "preparing",
	}
	tests := []struct {
		name         string
		opt          *storage.ListOptions
		strategyName string
		wantErr      bool
		want         []*storage.ScaleDownTask
		on           func(mockFields *MockFields)
	}{
		{
			name: "normal",
			opt: &storage.ListOptions{
				Limit:                  1,
				Page:                   0,
				ReturnSoftDeletedItems: false,
			},
			strategyName: "nodegroup1",
			wantErr:      false,
			want:         []*storage.ScaleDownTask{task},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+taskTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+taskTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "task_1").Return(true, nil)
				mockFields.table.On("HasIndex", context.Background(), "node_group_strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithSort", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithStart", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithLimit", mock.Anything).Return(mockFields.find)
				mockFields.find.On("All", context.Background(), mock.Anything).Return(func(ctx context.Context, result interface{}) error {
					return reflectInterface(result, []*storage.ScaleDownTask{task})
				})
			},
		},
		{
			name: "err",
			opt: &storage.ListOptions{
				Limit:                  1,
				Page:                   0,
				ReturnSoftDeletedItems: false,
			},
			strategyName: "nodegroup1",
			wantErr:      true,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+taskTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+taskTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "task_1").Return(true, nil)
				mockFields.table.On("HasIndex", context.Background(), "node_group_strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithSort", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithStart", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithLimit", mock.Anything).Return(mockFields.find)
				mockFields.find.On("All", context.Background(), mock.Anything).Return(fmt.Errorf("db err"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockFields{
				db:    mocks.NewDB(t),
				table: mocks.NewTable(t),
				find:  mocks.NewFind(t),
			}
			tt.on(mockDB)
			server := NewServer(mockDB.db)
			rsp, err := server.ListTasksByStrategy(tt.strategyName, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, rsp)
		})
	}
}

func TestModelTask_UpdateTask(t *testing.T) {
	task := &storage.ScaleDownTask{
		TaskID:            "taskID1",
		TotalNum:          2,
		NodeGroupStrategy: "nodegroup1",
		ScaleDownGroups:   []*storage.ScaleDownDetail{},
		DrainDelayDays:    2,
		Deadline:          time.Now().Add(48 * time.Hour),
		CreatedTime:       time.Now(),
		UpdatedTime:       time.Now(),
		IsDelete:          false,
		IsExecuted:        false,
		Status:            "preparing",
	}
	tests := []struct {
		name    string
		task    *storage.ScaleDownTask
		opt     *storage.UpdateOptions
		wantErr bool
		want    *storage.ScaleDownTask
		on      func(mockFields *MockFields)
	}{
		{
			name: "normal",
			task: task,
			opt: &storage.UpdateOptions{
				CreateIfNotExist:        true,
				OverwriteZeroOrEmptyStr: false,
			},
			wantErr: false,
			want:    task,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+taskTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+taskTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "task_1").Return(true, nil)
				mockFields.table.On("HasIndex", context.Background(), "node_group_strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.ScaleDownTask{}).
					Return(func(ctx context.Context, result interface{}) error {
						return reflectInterface(result, *task)
					})
				mockFields.table.On("Update", context.Background(), mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "create",
			task: task,
			opt: &storage.UpdateOptions{
				CreateIfNotExist:        true,
				OverwriteZeroOrEmptyStr: false,
			},
			wantErr: false,
			want:    task,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+taskTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+taskTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "task_1").Return(true, nil)
				mockFields.table.On("HasIndex", context.Background(), "node_group_strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.ScaleDownTask{}).Return(drivers.ErrTableRecordNotFound)
				mockFields.table.On("Insert", context.Background(), mock.Anything).Return(1, nil)
			},
		},
		{
			name: "notExistErr",
			task: task,
			opt: &storage.UpdateOptions{
				CreateIfNotExist:        false,
				OverwriteZeroOrEmptyStr: false,
			},
			wantErr: true,
			want:    task,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+taskTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+taskTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "task_1").Return(true, nil)
				mockFields.table.On("HasIndex", context.Background(), "node_group_strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.ScaleDownTask{}).Return(drivers.ErrTableRecordNotFound)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockFields{
				db:    mocks.NewDB(t),
				table: mocks.NewTable(t),
				find:  mocks.NewFind(t),
			}
			tt.on(mockDB)
			server := NewServer(mockDB.db)
			_, err := server.UpdateTask(tt.task, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
