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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage/mongo/mocks"
)

func TestModelEvent_CreateNodeGroupEvent(t *testing.T) {
	event := &storage.NodeGroupEvent{
		NodeGroupID: "nodegroup1",
		ClusterID:   "cluster1",
		EventTime:   time.Now(),
		Event:       storage.ScaleUpState,
		MaxNum:      10,
		MinNum:      0,
		DesiredNum:  5,
		Reason:      "test",
		Message:     "test",
		IsDeleted:   false,
	}

	tests := []struct {
		name    string
		event   *storage.NodeGroupEvent
		opt     *storage.CreateOptions
		wantErr bool
		on      func(mockFields *MockFields)
	}{
		{
			name:  "normal",
			event: event,
			opt: &storage.CreateOptions{
				OverWriteIfExist: false,
			},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+eventTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+eventTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Insert", context.Background(), mock.Anything).Return(1, nil)
			},
			wantErr: false,
		},
		{
			name:  "insertErr",
			event: event,
			opt: &storage.CreateOptions{
				OverWriteIfExist: true,
			},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+eventTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+eventTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Insert", context.Background(), mock.Anything).Return(0, fmt.Errorf("db error"))
			},
			wantErr: true,
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
			err := server.CreateNodeGroupEvent(tt.event, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestModelEvent_ListNodeGroupEvent(t *testing.T) {
	event := &storage.NodeGroupEvent{
		NodeGroupID: "nodegroup1",
		ClusterID:   "cluster1",
		EventTime:   time.Now(),
		Event:       storage.ScaleUpState,
		MaxNum:      10,
		MinNum:      0,
		DesiredNum:  5,
		Reason:      "test",
		Message:     "test",
		IsDeleted:   false,
	}
	tests := []struct {
		name        string
		nodegroupId string
		opt         *storage.ListOptions
		wantErr     bool
		want        []*storage.NodeGroupEvent
		on          func(mockFields *MockFields)
	}{
		{
			name: "normal",
			opt: &storage.ListOptions{
				Limit:                  1,
				Page:                   0,
				ReturnSoftDeletedItems: false,
			},
			nodegroupId: "testNodeGroup",
			wantErr:     false,
			want:        []*storage.NodeGroupEvent{event},
			on: func(mockFields *MockFields) {
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+eventTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+eventTableName).Return(mockFields.table)
				mockFields.table.On("HasIndex", context.Background(), mock.Anything).Return(true, nil)
				mockFields.table.On("Find", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithSort", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithStart", mock.Anything).Return(mockFields.find)
				mockFields.find.On("WithLimit", mock.Anything).Return(mockFields.find)
				mockFields.find.On("All", context.Background(), mock.Anything).Return(func(ctx context.Context, result interface{}) error {
					return reflectInterface(result, []*storage.NodeGroupEvent{event})
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
				mockFields.db.On("HasTable", context.Background(), tableNamePrefix+eventTableName).Return(true, nil)
				mockFields.db.On("Table", tableNamePrefix+eventTableName).Return(mockFields.table)
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
			rsp, err := server.ListNodeGroupEvent(tt.nodegroupId, tt.opt)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, rsp)
		})
	}
}
