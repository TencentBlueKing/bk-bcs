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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage/mongo/mocks"
)

func Test_CreateNodeGroupAction(t *testing.T) {
	action := &storage.NodeGroupAction{
		NodeGroupID: "nodegroup1",
		ClusterID:   "cluster1",
		CreatedTime: time.Now(),
		Event:       storage.ScaleUpState,
		DeltaNum:    2,
		NodeIPs:     []string{"127.0.0.1"},
		Process:     0,
		Status:      "test",
		UpdatedTime: time.Now(),
		IsDeleted:   false,
	}
	tests := []struct {
		name    string
		action  *storage.NodeGroupAction
		opt     *storage.CreateOptions
		wantErr bool
		on      func(mockFields *MockFields)
	}{
		{
			name:   "normal",
			action: action,
			opt: &storage.CreateOptions{
				OverWriteIfExist: false,
			},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+actionTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+actionTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), mock.Anything).Return(drivers.ErrTableRecordNotFound)
				mockFields.table.On("Insert", context.Background(), mock.Anything).Return(1, nil)
			},
			wantErr: false,
		},
		{
			name:   "existErr",
			action: action,
			opt: &storage.CreateOptions{
				OverWriteIfExist: false,
			},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+actionTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+actionTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), mock.Anything).Return(nil)
			},
			wantErr: true,
		},
		{
			name:   "overwrite",
			action: action,
			opt: &storage.CreateOptions{
				OverWriteIfExist: true,
			},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+actionTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+actionTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
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
			err := server.CreateNodeGroupAction(tt.action, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestModelAction_GetNodeGroupActions(t *testing.T) {
	action := &storage.NodeGroupAction{
		NodeGroupID: "nodegroup1",
		ClusterID:   "cluster1",
		CreatedTime: time.Now(),
		Event:       storage.ScaleUpState,
		DeltaNum:    2,
		NodeIPs:     []string{"127.0.0.1"},
		Process:     0,
		Status:      "test",
		UpdatedTime: time.Now(),
		IsDeleted:   false,
	}
	tests := []struct {
		name        string
		nodeGroupID string
		event       string
		opt         *storage.GetOptions
		want        *storage.NodeGroupAction
		wantErr     bool
		on          func(mockFields *MockFields)
	}{
		{
			name:        "normal",
			nodeGroupID: "testStrategy1",
			event:       storage.ScaleUpState,
			opt: &storage.GetOptions{
				ErrIfNotExist:  false,
				GetSoftDeleted: false,
			},
			want:    action,
			wantErr: false,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+actionTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+actionTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroupAction{}).
					Return(func(ctx context.Context, result interface{}) error {
						return reflectInterface(result, *action)
					})
			},
		},
		{
			name:        "notExist",
			nodeGroupID: "testStrategy1",
			event:       storage.ScaleUpState,
			opt: &storage.GetOptions{
				ErrIfNotExist:  false,
				GetSoftDeleted: false,
			},
			wantErr: false,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+actionTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+actionTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), mock.Anything).Return(drivers.ErrTableRecordNotFound)
			},
		},
		{
			name:        "errNotExist",
			nodeGroupID: "testStrategy1",
			event:       storage.ScaleUpState,
			opt: &storage.GetOptions{
				ErrIfNotExist:  true,
				GetSoftDeleted: false,
			},
			wantErr: true,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+actionTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+actionTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
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
			ret, err := server.GetNodeGroupAction(tt.nodeGroupID, tt.event, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, ret)
		})
	}
}

func TestModelAction_ListNodeGroupAction(t *testing.T) {
	action := &storage.NodeGroupAction{
		NodeGroupID: "nodegroup1",
		ClusterID:   "cluster1",
		CreatedTime: time.Now(),
		Event:       storage.ScaleUpState,
		DeltaNum:    2,
		NodeIPs:     []string{"127.0.0.1"},
		Process:     0,
		Status:      "test",
		UpdatedTime: time.Now(),
		IsDeleted:   false,
	}
	tests := []struct {
		name        string
		nodegroupId string
		opt         *storage.ListOptions
		wantErr     bool
		want        []*storage.NodeGroupAction
		on          func(mockFields *MockFields)
	}{
		{
			name: "normal",
			opt: &storage.ListOptions{
				Limit:                  1,
				Page:                   0,
				ReturnSoftDeletedItems: false,
			},
			nodegroupId: "testStrategy",
			wantErr:     false,
			want:        []*storage.NodeGroupAction{action},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+actionTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+actionTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithSort", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithStart", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithLimit", mock.Anything).Return(mockFields.find)
				mockFields.find.On("All", context.Background(), mock.Anything).Return(func(ctx context.Context, result interface{}) error {
					return reflectInterface(result, []*storage.NodeGroupAction{action})
				})
			},
		},
		{
			name:        "err",
			nodegroupId: "testStrategy",
			opt: &storage.ListOptions{
				Limit:                  1,
				Page:                   0,
				ReturnSoftDeletedItems: false,
			},
			wantErr: true,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+actionTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+actionTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
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
			rsp, err := server.ListNodeGroupAction(tt.nodegroupId, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, rsp)
		})
	}
}

func Test_UpdateNodeGroupAction(t *testing.T) {
	action := &storage.NodeGroupAction{
		NodeGroupID: "nodegroup1",
		ClusterID:   "cluster1",
		CreatedTime: time.Now(),
		Event:       storage.ScaleUpState,
		DeltaNum:    2,
		NodeIPs:     []string{"127.0.0.1"},
		Process:     0,
		Status:      "test",
		UpdatedTime: time.Now(),
		IsDeleted:   false,
	}
	tests := []struct {
		name    string
		action  *storage.NodeGroupAction
		opt     *storage.UpdateOptions
		wantErr bool
		want    *storage.NodeGroupAction
		on      func(mockFields *MockFields)
	}{
		{
			name:   "normal",
			action: action,
			opt: &storage.UpdateOptions{
				CreateIfNotExist:        true,
				OverwriteZeroOrEmptyStr: false,
			},
			wantErr: false,
			want:    action,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+actionTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+actionTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroupAction{}).
					Return(func(ctx context.Context, result interface{}) error {
						return reflectInterface(result, *action)
					})
				mockFields.table.On("Update", context.Background(), mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:   "create",
			action: action,
			opt: &storage.UpdateOptions{
				CreateIfNotExist:        true,
				OverwriteZeroOrEmptyStr: false,
			},
			wantErr: false,
			want:    action,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+actionTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+actionTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroupAction{}).Return(drivers.ErrTableRecordNotFound)
				mockFields.table.On("Insert", context.Background(), mock.Anything).Return(1, nil)
			},
		},
		{
			name:   "notExistErr",
			action: action,
			opt: &storage.UpdateOptions{
				CreateIfNotExist:        false,
				OverwriteZeroOrEmptyStr: false,
			},
			wantErr: true,
			want:    action,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+actionTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+actionTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroupAction{}).Return(drivers.ErrTableRecordNotFound)
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
			_, err := server.UpdateNodeGroupAction(tt.action, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_DeleteNodeGroupAction(t *testing.T) {
	action := &storage.NodeGroupAction{
		NodeGroupID: "nodegroup1",
		ClusterID:   "cluster1",
		CreatedTime: time.Now(),
		Event:       storage.ScaleUpState,
		DeltaNum:    2,
		NodeIPs:     []string{"127.0.0.1"},
		Process:     0,
		Status:      "test",
		UpdatedTime: time.Now(),
		IsDeleted:   false,
	}
	tests := []struct {
		name    string
		action  *storage.NodeGroupAction
		opt     *storage.DeleteOptions
		wantErr bool
		want    *storage.NodeGroupAction
		on      func(mockFields *MockFields)
	}{
		{
			name:    "normal",
			action:  action,
			opt:     &storage.DeleteOptions{ErrIfNotExist: false},
			wantErr: false,
			want:    action,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+actionTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+actionTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroupAction{}).
					Return(func(ctx context.Context, result interface{}) error {
						return reflectInterface(result, *action)
					})
				mockFields.table.On("Update", context.Background(), mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "notExist",
			action:  action,
			opt:     &storage.DeleteOptions{ErrIfNotExist: false},
			wantErr: false,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+actionTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+actionTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroupAction{}).Return(drivers.ErrTableRecordNotFound)
			},
		},
		{
			name:    "err",
			action:  action,
			opt:     &storage.DeleteOptions{ErrIfNotExist: false},
			wantErr: true,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+actionTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+actionTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroupAction{}).Return(fmt.Errorf("db error"))
			},
		},
		{
			name:    "notExistErr",
			action:  action,
			opt:     &storage.DeleteOptions{ErrIfNotExist: true},
			wantErr: true,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+actionTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+actionTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroupAction{}).Return(drivers.ErrTableRecordNotFound)
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
			rsp, err := server.DeleteNodeGroupAction(tt.action, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, rsp)
		})
	}
}
