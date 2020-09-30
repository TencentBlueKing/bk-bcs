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
 *
 */

package api

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	moc_bkdata "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/apigateway/bkdata/mock"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/api/proto/logmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/config"
	k8s "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/mock/manager"
)

// TestObtainDataID test obtain dataid method
func TestObtainDataID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCreator := moc_bkdata.NewMockClientCreatorInterface(ctrl)
	mockClient := moc_bkdata.NewMockClientInterface(ctrl)
	mockLogManager := k8s.NewMockLogManagerInterface(ctrl)
	errRet := fmt.Errorf("error ObtainDataID test")
	mockClient.EXPECT().ObtainDataID(gomock.Any()).Return(int64(-1), errRet)
	mockClient.EXPECT().ObtainDataID(gomock.Any()).Return(int64(21903), nil).Times(2)
	mockClient.EXPECT().SetCleanStrategy(gomock.Any()).Return(nil)
	mockClient.EXPECT().SetCleanStrategy(gomock.Any()).Return(errRet)
	mockCreator.EXPECT().NewClientFromConfig(gomock.Any()).Return(mockClient).Times(3)

	server := &LogManagerServerImpl{
		logManager:          mockLogManager,
		apiHost:             "http://127.0.0.1:8080",
		bkdataClientCreator: mockCreator,
	}
	resp := proto.ObtainDataidResp{}
	err := server.ObtainDataID(context.Background(), &proto.ObtainDataidReq{}, &resp)
	if err == nil {
		t.Errorf("LogManagerServerImpl.ObtainDataID returns dataid(%d), error(%+v), expect dataid(-1), error(%+v)", resp.DataID, err, errRet)
	}
	err = server.ObtainDataID(context.Background(), &proto.ObtainDataidReq{}, &resp)
	if err != nil {
		t.Errorf("LogManagerServerImpl.ObtainDataID returns dataid(%d), error(%+v), expect dataid(21903), error(%+v)", resp.DataID, err, nil)
	}
	err = server.ObtainDataID(context.Background(), &proto.ObtainDataidReq{}, &resp)
	if resp.Message == "" {
		t.Errorf("LogManagerServerImpl.ObtainDataID returns dataid(%d), error(%+v), expect dataid(21903), error(%+v)", resp.DataID, err, errRet)
	}
}

// TestSetCleanStrategy test create data clean stategy method
func TestSetCleanStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCreator := moc_bkdata.NewMockClientCreatorInterface(ctrl)
	mockClient := moc_bkdata.NewMockClientInterface(ctrl)
	mockLogManager := k8s.NewMockLogManagerInterface(ctrl)
	errRet := fmt.Errorf("error SetCleanStrategy test")
	mockClient.EXPECT().SetCleanStrategy(gomock.Any()).Return(errRet)
	mockClient.EXPECT().SetCleanStrategy(gomock.Any()).Return(nil)
	mockCreator.EXPECT().NewClientFromConfig(gomock.Any()).Return(mockClient).Times(2)

	server := &LogManagerServerImpl{
		logManager:          mockLogManager,
		apiHost:             "http://127.0.0.1:8080",
		bkdataClientCreator: mockCreator,
	}
	resp := proto.CommonResp{}
	err := server.CreateCleanStrategy(context.Background(), &proto.CreateCleanStrategyReq{}, &resp)
	if err == nil {
		t.Errorf("LogManagerServerImpl.CreateCleanStrategy returns error(%+v), expect error(%+v)", err, errRet)
	}
	err = server.CreateCleanStrategy(context.Background(), &proto.CreateCleanStrategyReq{}, &resp)
	if err != nil {
		t.Errorf("LogManagerServerImpl.CreateCleanStrategy returns error(%+v), expect error(%+v)", err, errRet)
	}
}

// TestCreateLogCollectionTask test create log collection task method
func TestCreateLogCollectionTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogManager := k8s.NewMockLogManagerInterface(ctrl)
	mockLogManager.EXPECT().HandleAddLogCollectionTask(gomock.Any(), gomock.Any()).Return(&proto.CollectionTaskCommonResp{})
	mockLogManager.EXPECT().HandleAddLogCollectionTask(gomock.Any(), gomock.Any()).Return(nil)
	mockLogManager.EXPECT().HandleAddLogCollectionTask(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, conf *config.CollectionConfig) *proto.CollectionTaskCommonResp {
		time.Sleep(time.Second * 12)
		return nil
	})
	server := &LogManagerServerImpl{
		logManager: mockLogManager,
	}
	resp := proto.CollectionTaskCommonResp{}
	emptyReq := &proto.CreateLogCollectionTaskReq{}
	normalReq := &proto.CreateLogCollectionTaskReq{
		Config: &proto.LogCollectionTaskConfig{
			Config: &proto.LogCollectionTaskConfigSpec{},
		},
	}
	err := server.CreateLogCollectionTask(context.Background(), normalReq, &resp)
	if err != nil {
		t.Errorf("LogManagerServerImpl.CreateLogCollectionTask returns error(%+v), expect error(%+v)", err, nil)
	}

	err = server.CreateLogCollectionTask(context.Background(), emptyReq, &resp)
	if err == nil {
		t.Errorf("LogManagerServerImpl.CreateLogCollectionTask returns error(%+v), expect error(%+v)", err, fmt.Errorf("Error in CreateLogCollectionTask: no LogCollectionConfig specified"))
	}

	err = server.CreateLogCollectionTask(context.Background(), normalReq, &resp)
	if err == nil {
		t.Errorf("LogManagerServerImpl.CreateLogCollectionTask returns error(%+v), expect error(%+v)", err, fmt.Errorf("Log Manager internal error"))
	}

	err = server.CreateLogCollectionTask(context.Background(), normalReq, &resp)
	if err == nil {
		t.Errorf("LogManagerServerImpl.CreateLogCollectionTask returns error(%+v), expect error(%+v)", err, fmt.Errorf("LogManagerServerImpl DeleteLogCollectionTask timeout"))
	}
}

// TestDeleteLogCollectionTask test delete log collection task method
func TestDeleteLogCollectionTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogManager := k8s.NewMockLogManagerInterface(ctrl)
	mockLogManager.EXPECT().HandleDeleteLogCollectionTask(gomock.Any(), gomock.Any()).Return(&proto.CollectionTaskCommonResp{})
	mockLogManager.EXPECT().HandleDeleteLogCollectionTask(gomock.Any(), gomock.Any()).Return(nil)
	mockLogManager.EXPECT().HandleDeleteLogCollectionTask(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, filter *config.CollectionFilterConfig) *proto.CollectionTaskCommonResp {
		time.Sleep(time.Second * 12)
		return nil
	})
	server := &LogManagerServerImpl{
		logManager: mockLogManager,
	}
	resp := proto.CollectionTaskCommonResp{}
	err := server.DeleteLogCollectionTask(context.Background(), &proto.DeleteLogCollectionTaskReq{}, &resp)
	if err != nil {
		t.Errorf("LogManagerServerImpl.DeleteLogCollectionTask returns error(%+v), expect error(%+v)", err, nil)
	}

	err = server.DeleteLogCollectionTask(context.Background(), &proto.DeleteLogCollectionTaskReq{}, &resp)
	if err == nil {
		t.Errorf("LogManagerServerImpl.DeleteLogCollectionTask returns error(%+v), expect error(%+v)", err, fmt.Errorf("Log Manager internal error"))
	}

	err = server.DeleteLogCollectionTask(context.Background(), &proto.DeleteLogCollectionTaskReq{}, &resp)
	if err == nil {
		t.Errorf("LogManagerServerImpl.DeleteLogCollectionTask returns error(%+v), expect error(%+v)", err, fmt.Errorf("LogManagerServerImpl DeleteLogCollectionTask timeout"))
	}
}

// TestListLogCollectionTask test list log collection task method
func TestListLogCollectionTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogManager := k8s.NewMockLogManagerInterface(ctrl)
	mockLogManager.EXPECT().HandleListLogCollectionTask(gomock.Any(), gomock.Any()).Return(make(map[string][]config.CollectionConfig))
	mockLogManager.EXPECT().HandleListLogCollectionTask(gomock.Any(), gomock.Any()).Return(nil)
	mockLogManager.EXPECT().HandleListLogCollectionTask(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, filter *config.CollectionFilterConfig) map[string][]config.CollectionConfig {
		time.Sleep(time.Second * 12)
		return nil
	})
	server := &LogManagerServerImpl{
		logManager: mockLogManager,
	}
	resp := proto.ListLogCollectionTaskResp{}
	err := server.ListLogCollectionTask(context.Background(), &proto.ListLogCollectionTaskReq{}, &resp)
	if err != nil {
		t.Errorf("LogManagerServerImpl.DeleteLogCollectionTask returns error(%+v), expect error(%+v)", err, nil)
	}

	err = server.ListLogCollectionTask(context.Background(), &proto.ListLogCollectionTaskReq{}, &resp)
	if err == nil {
		t.Errorf("LogManagerServerImpl.DeleteLogCollectionTask returns error(%+v), expect error(%+v)", err, fmt.Errorf("Log Manager internal error"))
	}

	err = server.ListLogCollectionTask(context.Background(), &proto.ListLogCollectionTaskReq{}, &resp)
	if err == nil {
		t.Errorf("LogManagerServerImpl.DeleteLogCollectionTask returns error(%+v), expect error(%+v)", err, fmt.Errorf("LogManagerServerImpl DeleteLogCollectionTask timeout"))
	}
}
