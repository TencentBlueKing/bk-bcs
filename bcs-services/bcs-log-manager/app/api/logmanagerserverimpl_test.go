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

	moc_bkdata "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/apigateway/bkdata/mock"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/api/proto/logmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/config"
	"github.com/golang/mock/gomock"
)

type MockLogManager = k8s.LogManager

// TestObtainDataID test obtain dataid method
func TestObtainDataID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCreator := moc_bkdata.NewMockClientCreatorInterface(ctrl)
	mockClient := moc_bkdata.NewMockClientInterface(ctrl)
	errRet := fmt.Errorf("error ObtainDataID test")
	mockClient.EXPECT().ObtainDataID(gomock.Any()).Return(int64(-1), errRet)
	mockClient.EXPECT().ObtainDataID(gomock.Any()).Return(int64(21093), nil)
	mockCreator.EXPECT().NewClientFromConfig(gomock.Any()).Return(mockClient).Times(2)

	mockLogmanager := &k8s.LogManager{}
	server := &LogManagerServerImpl{
		logManager:          mockLogmanager,
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
}

// TestSetCleanStrategy test create data clean stategy method
func TestSetCleanStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCreator := moc_bkdata.NewMockClientCreatorInterface(ctrl)
	mockClient := moc_bkdata.NewMockClientInterface(ctrl)
	errRet := fmt.Errorf("error SetCleanStrategy test")
	mockClient.EXPECT().SetCleanStrategy(gomock.Any()).Return(errRet)
	mockClient.EXPECT().SetCleanStrategy(gomock.Any()).Return(nil)
	mockCreator.EXPECT().NewClientFromConfig(gomock.Any()).Return(mockClient).Times(2)

	mockLogmanager := &k8s.LogManager{}
	server := &LogManagerServerImpl{
		logManager:          mockLogmanager,
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
	logManager := &k8s.LogManager{
		AddLogCollectionTask: make(chan *k8s.RequestMessage),
	}
	errRet := fmt.Errorf("error CreateLogCollectionTask test")
	go mockLogManagerGoroutine(logManager, errRet)
	server := &LogManagerServerImpl{
		logManager: logManager,
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
		t.Errorf("LogManagerServerImpl.CreateLogCollectionTask returns error(%+v), expect error(%+v)", err, errRet)
	}

	err = server.CreateLogCollectionTask(context.Background(), normalReq, &resp)
	if err == nil {
		t.Errorf("LogManagerServerImpl.CreateLogCollectionTask returns error(%+v), expect error(%+v)", err, fmt.Errorf("error receiving response data from log manager"))
	}
}

// TestDeleteLogCollectionTask test delete log collection task method
func TestDeleteLogCollectionTask(t *testing.T) {
	logManager := &k8s.LogManager{
		DeleteLogCollectionTask: make(chan *k8s.RequestMessage),
	}
	errRet := fmt.Errorf("error DeleteLogCollectionTask test")
	go mockLogManagerGoroutine(logManager, errRet)
	server := &LogManagerServerImpl{
		logManager: logManager,
	}
	resp := proto.CollectionTaskCommonResp{}
	err := server.DeleteLogCollectionTask(context.Background(), &proto.DeleteLogCollectionTaskReq{}, &resp)
	if err != nil {
		t.Errorf("LogManagerServerImpl.DeleteLogCollectionTask returns error(%+v), expect error(%+v)", err, nil)
	}

	err = server.DeleteLogCollectionTask(context.Background(), &proto.DeleteLogCollectionTaskReq{}, &resp)
	if err == nil {
		t.Errorf("LogManagerServerImpl.DeleteLogCollectionTask returns error(%+v), expect error(%+v)", err, errRet)
	}

	err = server.DeleteLogCollectionTask(context.Background(), &proto.DeleteLogCollectionTaskReq{}, &resp)
	if err == nil {
		t.Errorf("LogManagerServerImpl.DeleteLogCollectionTask returns error(%+v), expect error(%+v)", err, fmt.Errorf("error receiving response data from log manager"))
	}
}

// TestListLogCollectionTask test list log collection task method
func TestListLogCollectionTask(t *testing.T) {
	logManager := &k8s.LogManager{
		GetLogCollectionTask: make(chan *k8s.RequestMessage),
	}
	errRet := fmt.Errorf("error ListLogCollectionTask test")
	go mockLogManagerGoroutine(logManager, errRet)
	server := &LogManagerServerImpl{
		logManager: logManager,
	}
	resp := proto.ListLogCollectionTaskResp{}
	err := server.ListLogCollectionTask(context.Background(), &proto.ListLogCollectionTaskReq{}, &resp)
	if err != nil {
		t.Errorf("LogManagerServerImpl.DeleteLogCollectionTask returns error(%+v), expect error(%+v)", err, nil)
	}

	err = server.ListLogCollectionTask(context.Background(), &proto.ListLogCollectionTaskReq{}, &resp)
	if err == nil {
		t.Errorf("LogManagerServerImpl.DeleteLogCollectionTask returns error(%+v), expect error(%+v)", err, errRet)
	}

	err = server.ListLogCollectionTask(context.Background(), &proto.ListLogCollectionTaskReq{}, &resp)
	if err == nil {
		t.Errorf("LogManagerServerImpl.DeleteLogCollectionTask returns error(%+v), expect error(%+v)", err, fmt.Errorf("error receiving response data from log manager"))
	}
}

func mockLogManagerGoroutine(logManager *k8s.LogManager, errRet error) {
	times := 0
	for times < 4 {
		switch times {
		case 0:
			select {
			case msg := <-logManager.AddLogCollectionTask:
				msg.RespCh <- "termination"
			case msg := <-logManager.DeleteLogCollectionTask:
				msg.RespCh <- "termination"
			case msg := <-logManager.GetLogCollectionTask:
				msg.RespCh <- config.CollectionConfig{}
				msg.RespCh <- "termination"
			}
		case 1:
			select {
			case msg := <-logManager.AddLogCollectionTask:
				msg.RespCh <- errRet
			case msg := <-logManager.DeleteLogCollectionTask:
				msg.RespCh <- errRet
			case msg := <-logManager.GetLogCollectionTask:
				msg.RespCh <- errRet
			}
		case 2:
			select {
			case msg := <-logManager.AddLogCollectionTask:
				close(msg.RespCh)
			case msg := <-logManager.DeleteLogCollectionTask:
				close(msg.RespCh)
			case msg := <-logManager.GetLogCollectionTask:
				close(msg.RespCh)
			}
		}
		times++
	}
}
