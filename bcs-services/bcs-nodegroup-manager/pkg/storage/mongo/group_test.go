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

func TestModelGroup_CreateNodeGroup(t *testing.T) {
	group := &storage.NodeGroup{
		NodeGroupID:  "nodegroup1",
		ClusterID:    "cluster1",
		MaxSize:      10,
		MinSize:      0,
		DesiredSize:  5,
		UpcomingSize: 2,
		NodeIPs:      []string{"127.0.0.1"},
		Status:       storage.ScaleUpState,
		LastStatus:   storage.ScaleUpState,
		Message:      "",
		UpdatedTime:  time.Now(),
		IsDeleted:    false,
	}
	tests := []struct {
		name    string
		group   *storage.NodeGroup
		opt     *storage.CreateOptions
		wantErr bool
		on      func(mockFields *MockFields)
	}{
		{
			name:  "normal",
			group: group,
			opt: &storage.CreateOptions{
				OverWriteIfExist: false,
			},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+nodeGroupTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+nodeGroupTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), mock.Anything).Return(drivers.ErrTableRecordNotFound)
				mockFields.table.On("Insert", context.Background(), mock.Anything).Return(1, nil)
			},
			wantErr: false,
		},
		{
			name:  "existErr",
			group: group,
			opt: &storage.CreateOptions{
				OverWriteIfExist: false,
			},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+nodeGroupTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+nodeGroupTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), mock.Anything).Return(nil)
			},
			wantErr: true,
		},
		{
			name:  "overwrite",
			group: group,
			opt: &storage.CreateOptions{
				OverWriteIfExist: true,
			},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+nodeGroupTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+nodeGroupTableName).Return(mockFields.table)
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
			err := server.CreateNodeGroup(tt.group, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestModelGroup_GetNodeGroup(t *testing.T) {
	group := &storage.NodeGroup{
		NodeGroupID:  "nodegroup1",
		ClusterID:    "cluster1",
		MaxSize:      10,
		MinSize:      0,
		DesiredSize:  5,
		UpcomingSize: 2,
		NodeIPs:      []string{"127.0.0.1"},
		Status:       storage.ScaleUpState,
		LastStatus:   storage.ScaleUpState,
		Message:      "",
		UpdatedTime:  time.Now(),
		IsDeleted:    false,
	}
	tests := []struct {
		name        string
		nodeGroupID string
		opt         *storage.GetOptions
		want        *storage.NodeGroup
		wantErr     bool
		on          func(mockFields *MockFields)
	}{
		{
			name:        "normal",
			nodeGroupID: "testStrategy1",
			opt: &storage.GetOptions{
				ErrIfNotExist:  false,
				GetSoftDeleted: false,
			},
			want:    group,
			wantErr: false,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+nodeGroupTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+nodeGroupTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroup{}).
					Return(func(ctx context.Context, result interface{}) error {
						return reflectInterface(result, *group)
					})
			},
		},
		{
			name:        "notExist",
			nodeGroupID: "testStrategy1",
			opt: &storage.GetOptions{
				ErrIfNotExist:  false,
				GetSoftDeleted: false,
			},
			wantErr: false,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+nodeGroupTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+nodeGroupTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), mock.Anything).Return(drivers.ErrTableRecordNotFound)
			},
		},
		{
			name:        "errNotExist",
			nodeGroupID: "testStrategy1",
			opt: &storage.GetOptions{
				ErrIfNotExist:  true,
				GetSoftDeleted: false,
			},
			wantErr: true,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+nodeGroupTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+nodeGroupTableName).Return(mockFields.table)
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
			ret, err := server.GetNodeGroup(tt.nodeGroupID, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, ret)
		})
	}
}

func TestModelGroup_ListNodeGroups(t *testing.T) {
	group := &storage.NodeGroup{
		NodeGroupID:  "nodegroup1",
		ClusterID:    "cluster1",
		MaxSize:      10,
		MinSize:      0,
		DesiredSize:  5,
		UpcomingSize: 2,
		NodeIPs:      []string{"127.0.0.1"},
		Status:       storage.ScaleUpState,
		LastStatus:   storage.ScaleUpState,
		Message:      "",
		UpdatedTime:  time.Now(),
		IsDeleted:    false,
	}
	tests := []struct {
		name    string
		opt     *storage.ListOptions
		wantErr bool
		want    []*storage.NodeGroup
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
			want:    []*storage.NodeGroup{group},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+nodeGroupTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+nodeGroupTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithSort", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithStart", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithLimit", mock.Anything).Return(mockFields.find)
				mockFields.find.On("All", context.Background(), mock.Anything).Return(func(ctx context.Context, result interface{}) error {
					return reflectInterface(result, []*storage.NodeGroup{group})
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
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+nodeGroupTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+nodeGroupTableName).Return(mockFields.table)
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
			rsp, err := server.ListNodeGroups(tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, rsp)
		})
	}
}

func TestModelGroup_UpdateNodeGroup(t *testing.T) {
	group := &storage.NodeGroup{
		NodeGroupID:  "nodegroup1",
		ClusterID:    "cluster1",
		MaxSize:      10,
		MinSize:      0,
		DesiredSize:  5,
		UpcomingSize: 2,
		NodeIPs:      []string{"127.0.0.1"},
		Status:       storage.ScaleUpState,
		LastStatus:   storage.ScaleUpState,
		Message:      "",
		UpdatedTime:  time.Now(),
		IsDeleted:    false,
	}
	tests := []struct {
		name    string
		group   *storage.NodeGroup
		opt     *storage.UpdateOptions
		wantErr bool
		want    *storage.NodeGroup
		on      func(mockFields *MockFields)
	}{
		{
			name:  "normal",
			group: group,
			opt: &storage.UpdateOptions{
				CreateIfNotExist:        true,
				OverwriteZeroOrEmptyStr: false,
			},
			wantErr: false,
			want:    group,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+nodeGroupTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+nodeGroupTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroup{}).
					Return(func(ctx context.Context, result interface{}) error {
						return reflectInterface(result, *group)
					})
				mockFields.table.On("Update", context.Background(), mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:  "create",
			group: group,
			opt: &storage.UpdateOptions{
				CreateIfNotExist:        true,
				OverwriteZeroOrEmptyStr: false,
			},
			wantErr: false,
			want:    group,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+nodeGroupTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+nodeGroupTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroup{}).Return(drivers.ErrTableRecordNotFound)
				mockFields.table.On("Insert", context.Background(), mock.Anything).Return(1, nil)
			},
		},
		{
			name:  "notExistErr",
			group: group,
			opt: &storage.UpdateOptions{
				CreateIfNotExist:        false,
				OverwriteZeroOrEmptyStr: false,
			},
			wantErr: true,
			want:    group,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+nodeGroupTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+nodeGroupTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroup{}).Return(drivers.ErrTableRecordNotFound)
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
			_, err := server.UpdateNodeGroup(tt.group, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestModelGroup_DeleteNodeGroup(t *testing.T) {
	group := &storage.NodeGroup{
		NodeGroupID:  "nodegroup1",
		ClusterID:    "cluster1",
		MaxSize:      10,
		MinSize:      0,
		DesiredSize:  5,
		UpcomingSize: 2,
		NodeIPs:      []string{"127.0.0.1"},
		Status:       storage.ScaleUpState,
		LastStatus:   storage.ScaleUpState,
		Message:      "",
		UpdatedTime:  time.Now(),
		IsDeleted:    false,
	}
	tests := []struct {
		name    string
		groupId string
		opt     *storage.DeleteOptions
		wantErr bool
		want    *storage.NodeGroup
		on      func(mockFields *MockFields)
	}{
		{
			name:    "normal",
			groupId: "nodegroup1",
			opt:     &storage.DeleteOptions{ErrIfNotExist: false},
			wantErr: false,
			want:    group,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+nodeGroupTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+nodeGroupTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroup{}).
					Return(func(ctx context.Context, result interface{}) error {
						return reflectInterface(result, *group)
					})
				mockFields.table.On("Update", context.Background(), mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "notExist",
			groupId: "nodegroup1",
			opt:     &storage.DeleteOptions{ErrIfNotExist: false},
			wantErr: false,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+nodeGroupTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+nodeGroupTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroup{}).Return(drivers.ErrTableRecordNotFound)
			},
		},
		{
			name:    "err",
			groupId: "nodegroup1",
			opt:     &storage.DeleteOptions{ErrIfNotExist: false},
			wantErr: true,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+nodeGroupTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+nodeGroupTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroup{}).Return(fmt.Errorf("db error"))
			},
		},
		{
			name:    "notExistErr",
			groupId: "nodegroup1",
			opt:     &storage.DeleteOptions{ErrIfNotExist: true},
			wantErr: true,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+nodeGroupTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+nodeGroupTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroup{}).Return(drivers.ErrTableRecordNotFound)
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
			rsp, err := server.DeleteNodeGroup(tt.groupId, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, rsp)
		})
	}
}
