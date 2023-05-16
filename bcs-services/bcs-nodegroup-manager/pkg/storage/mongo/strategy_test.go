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
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage/mongo/mocks"
)

func Test_CreateNodeGroupStrategy(t *testing.T) {
	strategy := &storage.NodeGroupMgrStrategy{
		Name:         "testStrategy1",
		Labels:       map[string]string{"test": "test"},
		ResourcePool: "resourcePool1",
		ReservedNodeGroup: &storage.GroupInfo{
			NodeGroupID: "111",
			ClusterID:   "111",
			Weight:      1,
		},
		ElasticNodeGroups: []*storage.GroupInfo{
			{
				NodeGroupID: "222",
				ClusterID:   "222",
				Weight:      1,
			},
			{
				NodeGroupID: "333",
				ClusterID:   "333",
				Weight:      2,
			},
		},
		Strategy: &storage.Strategy{
			Type:              "buffer",
			ScaleUpCoolDown:   0,
			ScaleUpDelay:      0,
			MinScaleUpSize:    0,
			ScaleDownDelay:    0,
			MaxIdleDelay:      0,
			ReservedTimeRange: "",
			Buffer: &storage.BufferStrategy{
				Low:  1,
				High: 2,
			},
		},
		Status: &storage.State{
			Status:      "normal",
			LastStatus:  "",
			Error:       "",
			Message:     "",
			CreatedTime: time.Now(),
			UpdatedTime: time.Now(),
		},
	}
	tests := []struct {
		name     string
		strategy *storage.NodeGroupMgrStrategy
		opt      *storage.CreateOptions
		wantErr  bool
		on       func(mockFields *MockFields)
	}{
		{
			name:     "normal",
			strategy: strategy,
			opt: &storage.CreateOptions{
				OverWriteIfExist: false,
			},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+strategyTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+strategyTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), mock.Anything).Return(drivers.ErrTableRecordNotFound)
				mockFields.table.On("Insert", context.Background(), mock.Anything).Return(1, nil)
			},
			wantErr: false,
		},
		{
			name:     "existErr",
			strategy: strategy,
			opt: &storage.CreateOptions{
				OverWriteIfExist: false,
			},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+strategyTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+strategyTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), mock.Anything).Return(nil)
			},
			wantErr: true,
		},
		{
			name:     "overwrite",
			strategy: strategy,
			opt: &storage.CreateOptions{
				OverWriteIfExist: true,
			},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+strategyTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+strategyTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "strategy_1").Return(true, nil)
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
			err := server.CreateNodeGroupStrategy(tt.strategy, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_GetNodeGroupStrategy(t *testing.T) {
	strategy := &storage.NodeGroupMgrStrategy{
		Name:         "testStrategy1",
		Labels:       map[string]string{"test": "test"},
		ResourcePool: "resourcePool1",
		ReservedNodeGroup: &storage.GroupInfo{
			NodeGroupID: "111",
			ClusterID:   "111",
			Weight:      1,
		},
		ElasticNodeGroups: []*storage.GroupInfo{
			{
				NodeGroupID: "222",
				ClusterID:   "222",
				Weight:      1,
			},
			{
				NodeGroupID: "333",
				ClusterID:   "333",
				Weight:      2,
			},
		},
		Strategy: &storage.Strategy{
			Type:              "buffer",
			ScaleUpCoolDown:   0,
			ScaleUpDelay:      0,
			MinScaleUpSize:    0,
			ScaleDownDelay:    0,
			MaxIdleDelay:      0,
			ReservedTimeRange: "",
			Buffer: &storage.BufferStrategy{
				Low:  1,
				High: 2,
			},
		},
		Status: &storage.State{
			Status:      "normal",
			LastStatus:  "",
			Error:       "",
			Message:     "",
			CreatedTime: time.Now(),
			UpdatedTime: time.Now(),
		},
	}
	tests := []struct {
		name         string
		strategyName string
		opt          *storage.GetOptions
		want         *storage.NodeGroupMgrStrategy
		wantErr      bool
		on           func(mockFields *MockFields)
	}{
		{
			name:         "normal",
			strategyName: "testStrategy1",
			opt: &storage.GetOptions{
				ErrIfNotExist:  false,
				GetSoftDeleted: false,
			},
			want:    strategy,
			wantErr: false,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+strategyTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+strategyTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroupMgrStrategy{}).
					Return(func(ctx context.Context, result interface{}) error {
						return reflectInterface(result, *strategy)
					})
			},
		},
		{
			name:         "notExist",
			strategyName: "testStrategy1",
			opt: &storage.GetOptions{
				ErrIfNotExist:  false,
				GetSoftDeleted: false,
			},
			wantErr: false,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+strategyTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+strategyTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), mock.Anything).Return(drivers.ErrTableRecordNotFound)
			},
		},
		{
			name:         "errNotExist",
			strategyName: "testStrategy1",
			opt: &storage.GetOptions{
				ErrIfNotExist:  true,
				GetSoftDeleted: false,
			},
			wantErr: true,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+strategyTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+strategyTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "strategy_1").Return(true, nil)
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
			ret, err := server.GetNodeGroupStrategy(tt.strategyName, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, ret)
		})
	}
}

func Test_ListNodeGroupStrategies(t *testing.T) {
	strategy := &storage.NodeGroupMgrStrategy{
		Name:         "testStrategy1",
		Labels:       map[string]string{"test": "test"},
		ResourcePool: "resourcePool1",
		ReservedNodeGroup: &storage.GroupInfo{
			NodeGroupID: "111",
			ClusterID:   "111",
			Weight:      1,
		},
		ElasticNodeGroups: []*storage.GroupInfo{
			{
				NodeGroupID: "222",
				ClusterID:   "222",
				Weight:      1,
			},
			{
				NodeGroupID: "333",
				ClusterID:   "333",
				Weight:      2,
			},
		},
		Strategy: &storage.Strategy{
			Type:              "buffer",
			ScaleUpCoolDown:   0,
			ScaleUpDelay:      0,
			MinScaleUpSize:    0,
			ScaleDownDelay:    0,
			MaxIdleDelay:      0,
			ReservedTimeRange: "",
			Buffer: &storage.BufferStrategy{
				Low:  1,
				High: 2,
			},
		},
		Status: &storage.State{
			Status:      "normal",
			LastStatus:  "",
			Error:       "",
			Message:     "",
			CreatedTime: time.Now(),
			UpdatedTime: time.Now(),
		},
	}
	tests := []struct {
		name    string
		opt     *storage.ListOptions
		wantErr bool
		want    []*storage.NodeGroupMgrStrategy
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
			want:    []*storage.NodeGroupMgrStrategy{strategy},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+strategyTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+strategyTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithSort", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithStart", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithLimit", mock.Anything).Return(mockFields.find)
				mockFields.find.On("All", context.Background(), mock.Anything).Return(func(ctx context.Context, result interface{}) error {
					return reflectInterface(result, []*storage.NodeGroupMgrStrategy{strategy})
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
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+strategyTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+strategyTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "strategy_1").Return(true, nil)
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
			rsp, err := server.ListNodeGroupStrategies(tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, rsp)
		})
	}
}

func Test_UpdateNodeGroupStrategy(t *testing.T) {
	strategy := &storage.NodeGroupMgrStrategy{
		Name:         "testStrategy1",
		Labels:       map[string]string{"test": "test"},
		ResourcePool: "resourcePool1",
		ReservedNodeGroup: &storage.GroupInfo{
			NodeGroupID: "111",
			ClusterID:   "111",
			Weight:      1,
		},
		ElasticNodeGroups: []*storage.GroupInfo{
			{
				NodeGroupID: "222",
				ClusterID:   "222",
				Weight:      1,
			},
			{
				NodeGroupID: "333",
				ClusterID:   "333",
				Weight:      2,
			},
		},
		Strategy: &storage.Strategy{
			Type:              "buffer",
			ScaleUpCoolDown:   0,
			ScaleUpDelay:      0,
			MinScaleUpSize:    0,
			ScaleDownDelay:    0,
			MaxIdleDelay:      0,
			ReservedTimeRange: "",
			Buffer: &storage.BufferStrategy{
				Low:  1,
				High: 2,
			},
		},
		Status: &storage.State{
			Status:      "normal",
			LastStatus:  "",
			Error:       "",
			Message:     "",
			CreatedTime: time.Now(),
			UpdatedTime: time.Now(),
		},
	}
	tests := []struct {
		name     string
		strategy *storage.NodeGroupMgrStrategy
		opt      *storage.UpdateOptions
		wantErr  bool
		want     *storage.NodeGroupMgrStrategy
		on       func(mockFields *MockFields)
	}{
		{
			name:     "normal",
			strategy: strategy,
			opt: &storage.UpdateOptions{
				CreateIfNotExist:        true,
				OverwriteZeroOrEmptyStr: false,
			},
			wantErr: false,
			want:    strategy,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+strategyTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+strategyTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroupMgrStrategy{}).
					Return(func(ctx context.Context, result interface{}) error {
						return reflectInterface(result, *strategy)
					})
				mockFields.table.On("Update", context.Background(), mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:     "create",
			strategy: strategy,
			opt: &storage.UpdateOptions{
				CreateIfNotExist:        true,
				OverwriteZeroOrEmptyStr: false,
			},
			wantErr: false,
			want:    strategy,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+strategyTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+strategyTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroupMgrStrategy{}).Return(drivers.ErrTableRecordNotFound)
				mockFields.table.On("Insert", context.Background(), mock.Anything).Return(1, nil)
			},
		},
		{
			name:     "notExistErr",
			strategy: strategy,
			opt: &storage.UpdateOptions{
				CreateIfNotExist:        false,
				OverwriteZeroOrEmptyStr: false,
			},
			wantErr: true,
			want:    strategy,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+strategyTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+strategyTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroupMgrStrategy{}).Return(drivers.ErrTableRecordNotFound)
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
			_, err := server.UpdateNodeGroupStrategy(tt.strategy, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_DeleteNodeGroupStrategy(t *testing.T) {
	strategy := &storage.NodeGroupMgrStrategy{
		Name:         "testStrategy1",
		Labels:       map[string]string{"test": "test"},
		ResourcePool: "resourcePool1",
		ReservedNodeGroup: &storage.GroupInfo{
			NodeGroupID: "111",
			ClusterID:   "111",
			Weight:      1,
		},
		ElasticNodeGroups: []*storage.GroupInfo{
			{
				NodeGroupID: "222",
				ClusterID:   "222",
				Weight:      1,
			},
			{
				NodeGroupID: "333",
				ClusterID:   "333",
				Weight:      2,
			},
		},
		Strategy: &storage.Strategy{
			Type:              "buffer",
			ScaleUpCoolDown:   0,
			ScaleUpDelay:      0,
			MinScaleUpSize:    0,
			ScaleDownDelay:    0,
			MaxIdleDelay:      0,
			ReservedTimeRange: "",
			Buffer: &storage.BufferStrategy{
				Low:  1,
				High: 2,
			},
		},
		Status: &storage.State{
			Status:      "normal",
			LastStatus:  "",
			Error:       "",
			Message:     "",
			CreatedTime: time.Now(),
			UpdatedTime: time.Now(),
		},
	}
	tests := []struct {
		name         string
		strategyName string
		opt          *storage.DeleteOptions
		wantErr      bool
		want         *storage.NodeGroupMgrStrategy
		on           func(mockFields *MockFields)
	}{
		{
			name:         "normal",
			strategyName: "testStrategy1",
			opt:          &storage.DeleteOptions{ErrIfNotExist: false},
			wantErr:      false,
			want:         strategy,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+strategyTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+strategyTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroupMgrStrategy{}).
					Return(func(ctx context.Context, result interface{}) error {
						return reflectInterface(result, *strategy)
					})
				mockFields.table.On("Update", context.Background(), mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:         "notExist",
			strategyName: "testStrategy1",
			opt:          &storage.DeleteOptions{ErrIfNotExist: false},
			wantErr:      false,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+strategyTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+strategyTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroupMgrStrategy{}).Return(drivers.ErrTableRecordNotFound)
			},
		},
		{
			name:         "err",
			strategyName: "testStrategy1",
			opt:          &storage.DeleteOptions{ErrIfNotExist: false},
			wantErr:      true,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+strategyTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+strategyTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroupMgrStrategy{}).Return(fmt.Errorf("db error"))
			},
		},
		{
			name:         "notExistErr",
			strategyName: "testStrategy1",
			opt:          &storage.DeleteOptions{ErrIfNotExist: true},
			wantErr:      true,
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+strategyTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+strategyTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), "strategy_1").Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("One", context.Background(), &storage.NodeGroupMgrStrategy{}).Return(drivers.ErrTableRecordNotFound)
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
			rsp, err := server.DeleteNodeGroupStrategy(tt.strategyName, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, rsp)
		})
	}
}

type MockFields struct {
	db    *mocks.DB
	table *mocks.Table
	find  *mocks.Find
}

func reflectInterface(result interface{}, setValue interface{}) error {
	rval := reflect.ValueOf(result)
	switch rval.Kind() {
	case reflect.Ptr:
		if rval.IsNil() {
			return errors.New("cannot Decode to nil value")
		}
		rval = rval.Elem()
	case reflect.Map:
		if rval.IsNil() {
			return errors.New("cannot Decode to nil value")
		}
	default:
		return fmt.Errorf("argument to Decode must be a pointer or a map, but got %v", rval)
	}
	rval.Set(reflect.ValueOf(setValue))
	return nil
}
