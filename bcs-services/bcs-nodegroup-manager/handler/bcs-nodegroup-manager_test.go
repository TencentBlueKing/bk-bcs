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

package handler

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Tencent/bk-bcs/bcs-common/common"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage/mocks"
	nodegroupmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/proto"
)

func Test_CreateNodePoolMgrStrategy(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		req     *nodegroupmanager.CreateNodePoolMgrStrategyReq
		rsp     *nodegroupmanager.CreateNodePoolMgrStrategyRsp
		want    *nodegroupmanager.CreateNodePoolMgrStrategyRsp
		wantErr bool
		on      func(storageCli *mocks.Storage)
	}{
		{
			name: "normal",
			ctx:  context.Background(),
			req: &nodegroupmanager.CreateNodePoolMgrStrategyReq{
				Option: &nodegroupmanager.CreateOptions{
					OverWriteIfExist: true,
					Operator:         "test",
				},
				Strategy: &nodegroupmanager.NodeGroupStrategy{
					Name: "normal",
				},
			},
			rsp: &nodegroupmanager.CreateNodePoolMgrStrategyRsp{},
			want: &nodegroupmanager.CreateNodePoolMgrStrategyRsp{
				Code:    0,
				Message: "success",
				Result:  true,
			},
			wantErr: false,
			on: func(storageCli *mocks.Storage) {
				storageCli.On("CreateNodeGroupStrategy", mock.Anything, &storage.CreateOptions{OverWriteIfExist: true}).Return(nil)
			},
		},
		{
			name: "existErr",
			ctx:  context.Background(),
			req: &nodegroupmanager.CreateNodePoolMgrStrategyReq{
				Option: &nodegroupmanager.CreateOptions{
					OverWriteIfExist: false,
					Operator:         "test",
				},
				Strategy: &nodegroupmanager.NodeGroupStrategy{
					Name: "testexist",
				},
			},
			rsp: &nodegroupmanager.CreateNodePoolMgrStrategyRsp{},
			want: &nodegroupmanager.CreateNodePoolMgrStrategyRsp{
				Code:    common.AdditionErrorCode + 500,
				Message: "",
				Result:  false,
			},
			wantErr: false,
			on: func(storageCli *mocks.Storage) {
				storageCli.On("CreateNodeGroupStrategy", mock.Anything, &storage.CreateOptions{OverWriteIfExist: false}).Return(fmt.Errorf("exist"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageCli := mocks.NewStorage(t)
			tt.on(storageCli)
			handler := New(storageCli)
			err := handler.CreateNodePoolMgrStrategy(tt.ctx, tt.req, tt.rsp)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want.Code, tt.rsp.Code)
			assert.Equal(t, tt.want.Result, tt.rsp.Result)
		})
	}
}

func Test_GetNodePoolMgrStrategy(t *testing.T) {
	tests := []struct {
		name    string
		req     *nodegroupmanager.GetNodePoolMgrStrategyReq
		rsp     *nodegroupmanager.GetNodePoolMgrStrategyRsp
		want    *nodegroupmanager.GetNodePoolMgrStrategyRsp
		wantErr bool
		on      func(storageCli *mocks.Storage)
	}{
		{
			name: "normal",
			req: &nodegroupmanager.GetNodePoolMgrStrategyReq{
				Name: "normal",
			},
			rsp: &nodegroupmanager.GetNodePoolMgrStrategyRsp{},
			want: &nodegroupmanager.GetNodePoolMgrStrategyRsp{
				Code:    0,
				Message: "",
				Data: &nodegroupmanager.NodeGroupStrategy{
					Name:              "normal",
					Kind:              "NodeGroupStrategy",
					ReservedNodeGroup: &nodegroupmanager.ReservedNodeGroup{},
					ElasticNodeGroups: []*nodegroupmanager.ElasticNodeGroup{},
					Strategy:          &nodegroupmanager.Strategy{},
				},
			},
			wantErr: false,
			on: func(storageCli *mocks.Storage) {
				storageCli.On("GetNodeGroupStrategy", "normal", &storage.GetOptions{}).Return(&storage.NodeGroupMgrStrategy{Name: "normal"}, nil)
			},
		},
		{
			name: "empty",
			req: &nodegroupmanager.GetNodePoolMgrStrategyReq{
				Name: "empty",
			},
			rsp: &nodegroupmanager.GetNodePoolMgrStrategyRsp{},
			want: &nodegroupmanager.GetNodePoolMgrStrategyRsp{
				Code:    0,
				Message: "",
				Data:    nil,
			},
			wantErr: false,
			on: func(storageCli *mocks.Storage) {
				storageCli.On("GetNodeGroupStrategy", "empty", &storage.GetOptions{}).Return(nil, nil)
			},
		},
		{
			name: "err",
			req: &nodegroupmanager.GetNodePoolMgrStrategyReq{
				Name: "err",
			},
			rsp: &nodegroupmanager.GetNodePoolMgrStrategyRsp{},
			want: &nodegroupmanager.GetNodePoolMgrStrategyRsp{
				Code:    common.AdditionErrorCode + 500,
				Message: "find error",
				Data:    nil,
			},
			wantErr: false,
			on: func(storageCli *mocks.Storage) {
				storageCli.On("GetNodeGroupStrategy", "err", &storage.GetOptions{}).Return(nil, fmt.Errorf("find error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageCli := mocks.NewStorage(t)
			tt.on(storageCli)
			handler := New(storageCli)
			err := handler.GetNodePoolMgrStrategy(context.Background(), tt.req, tt.rsp)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want.Data, tt.rsp.Data)
			assert.Equal(t, tt.want.Code, tt.rsp.Code)
			assert.Contains(t, tt.rsp.Message, tt.want.Message)
		})
	}
}

func Test_ListNodePoolMgrStrategies(t *testing.T) {
	tests := []struct {
		name    string
		req     *nodegroupmanager.ListNodePoolMgrStrategyReq
		rsp     *nodegroupmanager.ListNodePoolMgrStrategyRsp
		want    *nodegroupmanager.ListNodePoolMgrStrategyRsp
		wantErr bool
		on      func(storageCli *mocks.Storage)
	}{
		{
			name: "normal",
			req: &nodegroupmanager.ListNodePoolMgrStrategyReq{
				Limit: 1,
				Page:  0,
			},
			rsp: &nodegroupmanager.ListNodePoolMgrStrategyRsp{},
			want: &nodegroupmanager.ListNodePoolMgrStrategyRsp{
				Code:    0,
				Message: "success",
				Total:   0,
				Data:    []*nodegroupmanager.NodeGroupStrategy{{Name: "normal"}},
			},
			wantErr: false,
			on: func(storageCli *mocks.Storage) {
				storageCli.On("ListNodeGroupStrategies", &storage.ListOptions{
					Limit: 1,
					Page:  0,
				}).Return([]*storage.NodeGroupMgrStrategy{{
					Name: "normal",
				}}, nil)
			},
		},
		{
			name: "empty",
			req: &nodegroupmanager.ListNodePoolMgrStrategyReq{
				Limit: 1,
				Page:  0,
			},
			rsp: &nodegroupmanager.ListNodePoolMgrStrategyRsp{},
			want: &nodegroupmanager.ListNodePoolMgrStrategyRsp{
				Code:    0,
				Message: "success",
				Total:   0,
				Data:    []*nodegroupmanager.NodeGroupStrategy{},
			},
			wantErr: false,
			on: func(storageCli *mocks.Storage) {
				storageCli.On("ListNodeGroupStrategies", &storage.ListOptions{
					Limit: 1,
					Page:  0,
				}).Return([]*storage.NodeGroupMgrStrategy{}, nil)
			},
		},
		{
			name: "err",
			req: &nodegroupmanager.ListNodePoolMgrStrategyReq{
				Limit: 1,
				Page:  0,
			},
			rsp: &nodegroupmanager.ListNodePoolMgrStrategyRsp{},
			want: &nodegroupmanager.ListNodePoolMgrStrategyRsp{
				Code:    0,
				Message: "list error",
				Total:   0,
				Data:    nil,
			},
			wantErr: false,
			on: func(storageCli *mocks.Storage) {
				storageCli.On("ListNodeGroupStrategies", &storage.ListOptions{
					Limit: 1,
					Page:  0,
				}).Return(nil, fmt.Errorf("list error"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageCli := mocks.NewStorage(t)
			tt.on(storageCli)
			handler := New(storageCli)
			err := handler.ListNodePoolMgrStrategies(context.Background(), tt.req, tt.rsp)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, len(tt.want.Data), len(tt.rsp.Data))
		})
	}
}

func Test_UpdateNodePoolMgrStrategy(t *testing.T) {
	tests := []struct {
		name    string
		req     *nodegroupmanager.UpdateNodePoolMgrStrategyReq
		rsp     *nodegroupmanager.CreateNodePoolMgrStrategyRsp
		want    *nodegroupmanager.CreateNodePoolMgrStrategyRsp
		wantErr bool
		on      func(storageCli *mocks.Storage)
	}{
		{
			name: "normal",
			req: &nodegroupmanager.UpdateNodePoolMgrStrategyReq{
				Option: &nodegroupmanager.UpdateOptions{
					CreateIfNotExist:        false,
					OverwriteZeroOrEmptyStr: false,
					Operator:                "test",
				},
				Strategy: &nodegroupmanager.NodeGroupStrategy{Name: "normal"},
			},
			rsp: &nodegroupmanager.CreateNodePoolMgrStrategyRsp{},
			want: &nodegroupmanager.CreateNodePoolMgrStrategyRsp{
				Code:    0,
				Message: "success",
				Result:  true,
			},
			wantErr: false,
			on: func(storageCli *mocks.Storage) {
				storageCli.On("UpdateNodeGroupStrategy", mock.Anything, &storage.UpdateOptions{
					CreateIfNotExist:        false,
					OverwriteZeroOrEmptyStr: false,
				}).Return(nil, nil)
			},
		},
		{
			name: "notexist",
			req: &nodegroupmanager.UpdateNodePoolMgrStrategyReq{
				Option: &nodegroupmanager.UpdateOptions{
					CreateIfNotExist:        false,
					OverwriteZeroOrEmptyStr: false,
					Operator:                "test",
				},
				Strategy: &nodegroupmanager.NodeGroupStrategy{Name: "notexist"},
			},
			rsp: &nodegroupmanager.CreateNodePoolMgrStrategyRsp{},
			want: &nodegroupmanager.CreateNodePoolMgrStrategyRsp{
				Code:    common.AdditionErrorCode + 500,
				Message: "not found",
				Result:  false,
			},
			wantErr: false,
			on: func(storageCli *mocks.Storage) {
				storageCli.On("UpdateNodeGroupStrategy", mock.Anything, &storage.UpdateOptions{
					CreateIfNotExist:        false,
					OverwriteZeroOrEmptyStr: false,
				}).Return(nil, fmt.Errorf("not found"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageCli := mocks.NewStorage(t)
			tt.on(storageCli)
			handler := New(storageCli)
			err := handler.UpdateNodePoolMgrStrategy(context.Background(), tt.req, tt.rsp)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want.Code, tt.rsp.Code)
			assert.Equal(t, tt.want.Result, tt.rsp.Result)
			assert.Contains(t, tt.rsp.Message, tt.want.Message)
		})
	}
}

func Test_DeleteNodePoolMgrStrategy(t *testing.T) {
	tests := []struct {
		name    string
		req     *nodegroupmanager.DeleteNodePoolMgrStrategyReq
		rsp     *nodegroupmanager.DeleteNodePoolMgrStrategyRsp
		want    *nodegroupmanager.DeleteNodePoolMgrStrategyRsp
		wantErr bool
		on      func(storageCli *mocks.Storage)
	}{
		{
			name: "normal",
			req: &nodegroupmanager.DeleteNodePoolMgrStrategyReq{
				Name:     "normal",
				Operator: "test",
			},
			rsp: &nodegroupmanager.DeleteNodePoolMgrStrategyRsp{},
			want: &nodegroupmanager.DeleteNodePoolMgrStrategyRsp{
				Code:    0,
				Message: "success",
				Result:  true,
			},
			wantErr: false,
			on: func(storageCli *mocks.Storage) {
				storageCli.On("DeleteNodeGroupStrategy", "normal", &storage.DeleteOptions{}).Return(&storage.NodeGroupMgrStrategy{}, nil)
			},
		},
		{
			name: "empty",
			req: &nodegroupmanager.DeleteNodePoolMgrStrategyReq{
				Name:     "empty",
				Operator: "test",
			},
			rsp: &nodegroupmanager.DeleteNodePoolMgrStrategyRsp{},
			want: &nodegroupmanager.DeleteNodePoolMgrStrategyRsp{
				Code:    0,
				Message: "not exist",
				Result:  true,
			},
			wantErr: false,
			on: func(storageCli *mocks.Storage) {
				storageCli.On("DeleteNodeGroupStrategy", "empty", &storage.DeleteOptions{}).Return(nil, nil)
			},
		},
		{
			name: "err",
			req: &nodegroupmanager.DeleteNodePoolMgrStrategyReq{
				Name:     "err",
				Operator: "test",
			},
			rsp: &nodegroupmanager.DeleteNodePoolMgrStrategyRsp{},
			want: &nodegroupmanager.DeleteNodePoolMgrStrategyRsp{
				Code:    common.AdditionErrorCode + 500,
				Message: "delete error",
				Result:  false,
			},
			wantErr: false,
			on: func(storageCli *mocks.Storage) {
				storageCli.On("DeleteNodeGroupStrategy", "err", &storage.DeleteOptions{}).Return(nil, fmt.Errorf("delete error"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageCli := mocks.NewStorage(t)
			tt.on(storageCli)
			handler := New(storageCli)
			err := handler.DeleteNodePoolMgrStrategy(context.Background(), tt.req, tt.rsp)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want.Code, tt.rsp.Code)
			assert.Equal(t, tt.want.Result, tt.rsp.Result)
			assert.Contains(t, tt.rsp.Message, tt.want.Message)
		})
	}
}

func Test_GetClusterAutoscalerReview(t *testing.T) {
	tests := []struct {
		name    string
		req     *nodegroupmanager.ClusterAutoscalerReview
		rsp     *nodegroupmanager.ClusterAutoscalerReview
		want    *nodegroupmanager.ClusterAutoscalerReview
		wantErr bool
		on      func(storageCli *mocks.Storage)
	}{
		{
			name: "normalScaleUp",
			req: &nodegroupmanager.ClusterAutoscalerReview{Request: &nodegroupmanager.AutoscalerReviewRequest{
				Uid: "test-uid",
				NodeGroups: map[string]*nodegroupmanager.NodeGroup{"testNodegroup": {
					NodeGroupID:  "testNodegroup",
					MaxSize:      10,
					MinSize:      2,
					DesiredSize:  5,
					UpcomingSize: 1,
					NodeIPs:      []string{"1.1.1.1"},
				}},
			}},
			rsp: &nodegroupmanager.ClusterAutoscalerReview{},
			want: &nodegroupmanager.ClusterAutoscalerReview{Response: &nodegroupmanager.AutoscalerReviewResponse{
				ScaleUps: []*nodegroupmanager.NodeScaleUpPolicy{{
					NodeGroupID: "testNodegroup",
					DesiredSize: 8,
				}},
				ScaleDowns: []*nodegroupmanager.NodeScaleDownPolicy{},
				Uid:        "test-uid",
			}},
			wantErr: false,
			on: func(storageCli *mocks.Storage) {
				storageCli.On("GetNodeGroup", "testNodegroup", &storage.GetOptions{}).Return(&storage.NodeGroup{
					NodeGroupID:  "testNodegroup",
					ClusterID:    "",
					MaxSize:      0,
					MinSize:      0,
					DesiredSize:  8,
					UpcomingSize: 0,
					NodeIPs:      []string{},
					Status:       "",
				}, nil)
				storageCli.On("UpdateNodeGroup", mock.Anything, &storage.UpdateOptions{
					CreateIfNotExist:        true,
					OverwriteZeroOrEmptyStr: true,
				}).Return(&storage.NodeGroup{
					NodeGroupID:  "testNodegroup",
					ClusterID:    "",
					MaxSize:      10,
					MinSize:      2,
					DesiredSize:  8,
					UpcomingSize: 1,
					NodeIPs:      []string{"1.1.1.1"},
				}, nil)
				storageCli.On("GetNodeGroupAction", "testNodegroup", storage.ScaleUpState, &storage.GetOptions{}).Return(&storage.NodeGroupAction{
					NodeGroupID:        "testNodegroup",
					CreatedTime:        time.Time{},
					Event:              storage.ScaleUpState,
					DeltaNum:           3,
					NewDesiredNum:      5,
					OriginalDesiredNum: 2,
					OriginalNodeNum:    2,
					NodeIPs:            nil,
					Process:            0,
					Status:             "",
					UpdatedTime:        time.Now(),
					IsDeleted:          false,
				}, nil)
				// storageCli.On("GetNodeGroupAction", "testNodegroup", storage.ScaleDownState, &storage.GetOptions{}).Return(nil, nil)
				storageCli.On("UpdateNodeGroupAction", mock.Anything, mock.Anything).Return(nil, nil)
				storageCli.On("UpdateNodeGroup", mock.Anything, &storage.UpdateOptions{}).Return(nil, nil)
			},
		},
		{
			name: "normalScaleDown",
			req: &nodegroupmanager.ClusterAutoscalerReview{Request: &nodegroupmanager.AutoscalerReviewRequest{
				Uid: "test-uid",
				NodeGroups: map[string]*nodegroupmanager.NodeGroup{"testNodegroup": {
					NodeGroupID:  "testNodegroup",
					MaxSize:      10,
					MinSize:      2,
					DesiredSize:  5,
					UpcomingSize: 1,
					NodeIPs:      []string{"1.1.1.1"},
				}},
			}},
			rsp: &nodegroupmanager.ClusterAutoscalerReview{},
			want: &nodegroupmanager.ClusterAutoscalerReview{Response: &nodegroupmanager.AutoscalerReviewResponse{
				ScaleUps: []*nodegroupmanager.NodeScaleUpPolicy{},
				ScaleDowns: []*nodegroupmanager.NodeScaleDownPolicy{{
					NodeGroupID: "testNodegroup",
					Type:        "NodeNum",
					NodeNum:     1,
				}},
				Uid: "test-uid",
			}},
			wantErr: false,
			on: func(storageCli *mocks.Storage) {
				storageCli.On("GetNodeGroup", "testNodegroup", &storage.GetOptions{}).Return(&storage.NodeGroup{
					NodeGroupID:  "testNodegroup",
					ClusterID:    "",
					MaxSize:      0,
					MinSize:      0,
					DesiredSize:  1,
					UpcomingSize: 0,
					NodeIPs:      []string{},
					Status:       "",
				}, nil)
				storageCli.On("UpdateNodeGroup", mock.Anything, &storage.UpdateOptions{
					CreateIfNotExist:        true,
					OverwriteZeroOrEmptyStr: true,
				}).Return(&storage.NodeGroup{
					NodeGroupID:  "testNodegroup",
					ClusterID:    "",
					MaxSize:      10,
					MinSize:      2,
					DesiredSize:  1,
					UpcomingSize: 1,
					NodeIPs:      []string{"1.1.1.1"},
				}, nil)
				storageCli.On("GetNodeGroupAction", "testNodegroup", storage.ScaleUpState, &storage.GetOptions{}).Return(nil, nil)
				storageCli.On("GetNodeGroupAction", "testNodegroup", storage.ScaleDownState, &storage.GetOptions{}).Return(&storage.NodeGroupAction{
					NodeGroupID:        "testNodegroup",
					CreatedTime:        time.Time{},
					Event:              storage.ScaleUpState,
					DeltaNum:           3,
					NewDesiredNum:      1,
					OriginalDesiredNum: 2,
					OriginalNodeNum:    2,
					NodeIPs:            nil,
					Process:            0,
					Status:             "",
					UpdatedTime:        time.Now(),
					IsDeleted:          false,
				}, nil)
				storageCli.On("UpdateNodeGroupAction", mock.Anything, mock.Anything).Return(nil, nil)
				storageCli.On("UpdateNodeGroup", mock.Anything, &storage.UpdateOptions{}).Return(nil, nil)
			},
		},
		{
			name: "getNodegroupError",
			req: &nodegroupmanager.ClusterAutoscalerReview{Request: &nodegroupmanager.AutoscalerReviewRequest{
				Uid: "test-uid",
				NodeGroups: map[string]*nodegroupmanager.NodeGroup{"testNodegroup": {
					NodeGroupID:  "testNodegroup",
					MaxSize:      10,
					MinSize:      2,
					DesiredSize:  5,
					UpcomingSize: 1,
					NodeIPs:      []string{"1.1.1.1"},
				}},
			}},
			rsp:     &nodegroupmanager.ClusterAutoscalerReview{},
			wantErr: true,
			want: &nodegroupmanager.ClusterAutoscalerReview{Response: &nodegroupmanager.AutoscalerReviewResponse{
				Uid: "test-uid",
			}},
			on: func(storageCli *mocks.Storage) {
				storageCli.On("GetNodeGroup", "testNodegroup", &storage.GetOptions{}).Return(nil, fmt.Errorf("find error"))
			},
		},
		{
			name: "equal",
			req: &nodegroupmanager.ClusterAutoscalerReview{Request: &nodegroupmanager.AutoscalerReviewRequest{
				Uid: "test-uid",
				NodeGroups: map[string]*nodegroupmanager.NodeGroup{"testNodegroup": {
					NodeGroupID:  "testNodegroup",
					MaxSize:      10,
					MinSize:      2,
					DesiredSize:  5,
					UpcomingSize: 1,
					NodeIPs:      []string{"1.1.1.1"},
				}},
			}},
			rsp:     &nodegroupmanager.ClusterAutoscalerReview{},
			wantErr: false,
			want: &nodegroupmanager.ClusterAutoscalerReview{Response: &nodegroupmanager.AutoscalerReviewResponse{
				Uid: "test-uid",
				ScaleUps: []*nodegroupmanager.NodeScaleUpPolicy{{
					NodeGroupID: "testNodegroup",
					DesiredSize: 5,
				}},
				ScaleDowns: []*nodegroupmanager.NodeScaleDownPolicy{},
			}},
			on: func(storageCli *mocks.Storage) {
				storageCli.On("GetNodeGroup", "testNodegroup", &storage.GetOptions{}).Return(&storage.NodeGroup{
					NodeGroupID:  "testNodegroup",
					ClusterID:    "",
					MaxSize:      10,
					MinSize:      2,
					DesiredSize:  5,
					UpcomingSize: 1,
					NodeIPs:      []string{"1.1.1.1"},
					Status:       "",
				}, nil)
				storageCli.On("GetNodeGroupAction", "testNodegroup", storage.ScaleUpState, &storage.GetOptions{}).Return(&storage.NodeGroupAction{
					NodeGroupID:        "testNodegroup",
					CreatedTime:        time.Time{},
					Event:              storage.ScaleUpState,
					DeltaNum:           3,
					NewDesiredNum:      5,
					OriginalDesiredNum: 2,
					OriginalNodeNum:    2,
					NodeIPs:            nil,
					Process:            0,
					Status:             "",
					UpdatedTime:        time.Now(),
					IsDeleted:          false,
				}, nil)
				storageCli.On("UpdateNodeGroupAction", mock.Anything, mock.Anything).Return(nil, nil)
				storageCli.On("UpdateNodeGroup", mock.Anything, &storage.UpdateOptions{}).Return(nil, nil)
			},
		},
		{
			name: "emptyAction",
			req: &nodegroupmanager.ClusterAutoscalerReview{Request: &nodegroupmanager.AutoscalerReviewRequest{
				Uid: "test-uid",
				NodeGroups: map[string]*nodegroupmanager.NodeGroup{"testNodegroup": {
					NodeGroupID:  "testNodegroup",
					MaxSize:      10,
					MinSize:      2,
					DesiredSize:  5,
					UpcomingSize: 1,
					NodeIPs:      []string{"1.1.1.1"},
				}},
			}},
			rsp:     &nodegroupmanager.ClusterAutoscalerReview{},
			wantErr: false,
			want: &nodegroupmanager.ClusterAutoscalerReview{Response: &nodegroupmanager.AutoscalerReviewResponse{
				Uid:        "test-uid",
				ScaleUps:   []*nodegroupmanager.NodeScaleUpPolicy{},
				ScaleDowns: []*nodegroupmanager.NodeScaleDownPolicy{},
			}},
			on: func(storageCli *mocks.Storage) {
				storageCli.On("GetNodeGroup", "testNodegroup", &storage.GetOptions{}).Return(&storage.NodeGroup{
					NodeGroupID:  "testNodegroup",
					ClusterID:    "",
					MaxSize:      10,
					MinSize:      2,
					DesiredSize:  5,
					UpcomingSize: 1,
					NodeIPs:      []string{"1.1.1.1"},
					Status:       "",
				}, nil)
				storageCli.On("GetNodeGroupAction", "testNodegroup", storage.ScaleUpState, &storage.GetOptions{}).Return(nil, nil)
				storageCli.On("GetNodeGroupAction", "testNodegroup", storage.ScaleDownState, &storage.GetOptions{}).Return(nil, nil)
				// storageCli.On("UpdateNodeGroupAction", mock.Anything, mock.Anything).Return(nil, nil)
				// storageCli.On("UpdateNodeGroup", mock.Anything, &storage.UpdateOptions{}).Return(nil, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageCli := mocks.NewStorage(t)
			tt.on(storageCli)
			handler := New(storageCli)
			err := handler.GetClusterAutoscalerReview(context.Background(), tt.req, tt.rsp)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.req.Request, tt.rsp.Request)
			assert.Equal(t, tt.req.Request.Uid, tt.rsp.Response.Uid)
			assert.Equal(t, tt.want.Response.ScaleDowns, tt.rsp.Response.ScaleDowns)
			assert.Equal(t, tt.want.Response.ScaleUps, tt.rsp.Response.ScaleUps)
		})
	}
}

func Test_calculateProcess(t *testing.T) {
	tests := []struct {
		name    string
		desire  int
		current int
		want    int
	}{
		{
			name:    storage.ScaleUpState,
			desire:  100,
			current: 60,
			want:    60,
		},
		{
			name:    storage.ScaleDownState,
			desire:  100,
			current: 110,
			want:    90,
		},
		{
			name:    storage.ScaleDownState,
			desire:  1,
			current: 2,
			want:    50,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, calculateProcess(tt.current, tt.desire))
		})
	}
}

func Test_checkNodeGroupEqual(t *testing.T) {
	origin := &nodegroupmanager.NodeGroup{
		NodeGroupID:  "testNodeGroup",
		MaxSize:      10,
		MinSize:      2,
		DesiredSize:  8,
		UpcomingSize: 2,
		NodeTemplate: nil,
		NodeIPs:      []string{"1.1.1.1", "2.2.2.2"},
	}
	storageNodegroup := &storage.NodeGroup{
		NodeGroupID:  "testNodeGroup",
		MaxSize:      10,
		MinSize:      2,
		DesiredSize:  8,
		UpcomingSize: 2,
		NodeIPs:      []string{"1.1.1.1", "2.2.2.2"},
	}
	assert.Equal(t, true, checkNodeGroupEqual(origin, storageNodegroup))
	storageNodegroup.MaxSize = 8
	assert.Equal(t, false, checkNodeGroupEqual(origin, storageNodegroup))
	storageNodegroup.MaxSize = 10
	storageNodegroup.NodeIPs = []string{}
	assert.Equal(t, false, checkNodeGroupEqual(origin, storageNodegroup))
	storageNodegroup.NodeIPs = []string{"2.2.2.2", "3.3.3.3"}
	assert.Equal(t, false, checkNodeGroupEqual(origin, storageNodegroup))
}
