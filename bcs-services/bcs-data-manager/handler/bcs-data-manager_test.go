/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package handler

import (
	"context"
	bcsCommon "github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/mock"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetClusterInfo(t *testing.T) {
	storeServer := mock.NewMockStore()
	mockPmCli := mock.NewMockPmClient()
	resourceGetter := common.NewGetter(false, []string{}, "stag", mockPmCli)
	handler := NewBcsDataManager(storeServer, resourceGetter)
	ctx := context.Background()
	req := &bcsdatamanager.GetClusterInfoRequest{ClusterID: "testCluster"}
	rsp := &bcsdatamanager.GetClusterInfoResponse{}
	err := handler.GetClusterInfo(ctx, req, rsp)
	assert.Nil(t, err)
	assert.Equal(t, uint32(bcsCommon.BcsSuccess), rsp.GetCode())
	reqErr := &bcsdatamanager.GetClusterInfoRequest{ClusterID: "testErr"}
	rspErr := &bcsdatamanager.GetClusterInfoResponse{}
	err = handler.GetClusterInfo(ctx, reqErr, rspErr)
	assert.Nil(t, err)
	assert.Equal(t, uint32(bcsCommon.AdditionErrorCode+500), rspErr.GetCode())
}

func TestGetClusterInfoList(t *testing.T) {
	storeServer := mock.NewMockStore()
	mockPmCli := mock.NewMockPmClient()
	resourceGetter := common.NewGetter(false, []string{}, "stag", mockPmCli)
	handler := NewBcsDataManager(storeServer, resourceGetter)
	ctx := context.Background()
	req := &bcsdatamanager.GetClusterListRequest{Project: "testProject"}
	rsp := &bcsdatamanager.GetClusterListResponse{}
	err := handler.GetClusterListByProject(ctx, req, rsp)
	assert.Nil(t, err)
	assert.Equal(t, uint32(bcsCommon.BcsSuccess), rsp.GetCode())
	reqErr := &bcsdatamanager.GetClusterListRequest{Project: "testErr"}
	rspErr := &bcsdatamanager.GetClusterListResponse{}
	err = handler.GetClusterListByProject(ctx, reqErr, rspErr)
	assert.Nil(t, err)
	assert.Equal(t, uint32(bcsCommon.AdditionErrorCode+500), rspErr.GetCode())
}

func TestGetAllClusterList(t *testing.T) {
	storeServer := mock.NewMockStore()
	mockPmCli := mock.NewMockPmClient()
	resourceGetter := common.NewGetter(false, []string{}, "stag", mockPmCli)
	handler := NewBcsDataManager(storeServer, resourceGetter)
	ctx := context.Background()
	req := &bcsdatamanager.GetClusterListRequest{Project: "testProject"}
	rsp := &bcsdatamanager.GetClusterListResponse{}
	err := handler.GetAllClusterList(ctx, req, rsp)
	assert.Nil(t, err)
	assert.Equal(t, uint32(bcsCommon.BcsSuccess), rsp.GetCode())
	reqErr := &bcsdatamanager.GetClusterListRequest{Project: "testErr"}
	rspErr := &bcsdatamanager.GetClusterListResponse{}
	err = handler.GetAllClusterList(ctx, reqErr, rspErr)
	assert.Nil(t, err)
	assert.Equal(t, uint32(bcsCommon.AdditionErrorCode+500), rspErr.GetCode())
}

func TestGetNamespaceInfo(t *testing.T) {
	storeServer := mock.NewMockStore()
	mockPmCli := mock.NewMockPmClient()
	resourceGetter := common.NewGetter(false, []string{}, "stag", mockPmCli)
	handler := NewBcsDataManager(storeServer, resourceGetter)
	ctx := context.Background()
	req := &bcsdatamanager.GetNamespaceInfoRequest{ClusterID: "testCluster", Namespace: "testNs"}
	rsp := &bcsdatamanager.GetNamespaceInfoResponse{}
	err := handler.GetNamespaceInfo(ctx, req, rsp)
	assert.Nil(t, err)
	assert.Equal(t, uint32(bcsCommon.BcsSuccess), rsp.GetCode())
	reqErr := &bcsdatamanager.GetNamespaceInfoRequest{ClusterID: "testErr", Namespace: "testErr"}
	rspErr := &bcsdatamanager.GetNamespaceInfoResponse{}
	err = handler.GetNamespaceInfo(ctx, reqErr, rspErr)
	assert.Nil(t, err)
	assert.Equal(t, uint32(bcsCommon.AdditionErrorCode+500), rspErr.GetCode())
}

func TestGetNamespaceInfoList(t *testing.T) {
	storeServer := mock.NewMockStore()
	mockPmCli := mock.NewMockPmClient()
	resourceGetter := common.NewGetter(false, []string{}, "stag", mockPmCli)
	handler := NewBcsDataManager(storeServer, resourceGetter)
	ctx := context.Background()
	req := &bcsdatamanager.GetNamespaceInfoListRequest{ClusterID: "testCluster"}
	rsp := &bcsdatamanager.GetNamespaceInfoListResponse{}
	err := handler.GetNamespaceInfoList(ctx, req, rsp)
	assert.Nil(t, err)
	assert.Equal(t, uint32(bcsCommon.BcsSuccess), rsp.GetCode())
	reqErr := &bcsdatamanager.GetNamespaceInfoListRequest{ClusterID: "testErr"}
	rspErr := &bcsdatamanager.GetNamespaceInfoListResponse{}
	err = handler.GetNamespaceInfoList(ctx, reqErr, rspErr)
	assert.Nil(t, err)
	assert.Equal(t, uint32(bcsCommon.AdditionErrorCode+500), rspErr.GetCode())
}

func TestGetAllProjectList(t *testing.T) {
	storeServer := mock.NewMockStore()
	mockPmCli := mock.NewMockPmClient()
	resourceGetter := common.NewGetter(false, []string{}, "stag", mockPmCli)
	handler := NewBcsDataManager(storeServer, resourceGetter)
	ctx := context.Background()
	req := &bcsdatamanager.GetAllProjectListRequest{}
	rsp := &bcsdatamanager.GetAllProjectListResponse{}
	err := handler.GetAllProjectList(ctx, req, rsp)
	assert.Nil(t, err)
	assert.Equal(t, uint32(bcsCommon.BcsSuccess), rsp.GetCode())
}

func TestGetProjectInfo(t *testing.T) {
	storeServer := mock.NewMockStore()
	mockPmCli := mock.NewMockPmClient()
	resourceGetter := common.NewGetter(false, []string{}, "stag", mockPmCli)
	handler := NewBcsDataManager(storeServer, resourceGetter)
	ctx := context.Background()
	req := &bcsdatamanager.GetProjectInfoRequest{Project: "testProject"}
	rsp := &bcsdatamanager.GetProjectInfoResponse{}
	err := handler.GetProjectInfo(ctx, req, rsp)
	assert.Nil(t, err)
	assert.Equal(t, uint32(bcsCommon.BcsSuccess), rsp.GetCode())
	reqErr := &bcsdatamanager.GetProjectInfoRequest{Project: "testErr"}
	rspErr := &bcsdatamanager.GetProjectInfoResponse{}
	err = handler.GetProjectInfo(ctx, reqErr, rspErr)
	assert.Nil(t, err)
	assert.Equal(t, uint32(bcsCommon.AdditionErrorCode+500), rspErr.GetCode())
}

func TestGetWorkloadInfo(t *testing.T) {
	storeServer := mock.NewMockStore()
	mockPmCli := mock.NewMockPmClient()
	resourceGetter := common.NewGetter(false, []string{}, "stag", mockPmCli)
	handler := NewBcsDataManager(storeServer, resourceGetter)
	ctx := context.Background()
	req := &bcsdatamanager.GetWorkloadInfoRequest{ClusterID: "testCluster", Namespace: "testNs", WorkloadType: "testType", WorkloadName: "testName"}
	rsp := &bcsdatamanager.GetWorkloadInfoResponse{}
	err := handler.GetWorkloadInfo(ctx, req, rsp)
	assert.Nil(t, err)
	assert.Equal(t, uint32(bcsCommon.BcsSuccess), rsp.GetCode())
	reqErr := &bcsdatamanager.GetWorkloadInfoRequest{ClusterID: "testErr", Namespace: "testErr", WorkloadType: "testErr", WorkloadName: "testErr"}
	rspErr := &bcsdatamanager.GetWorkloadInfoResponse{}
	err = handler.GetWorkloadInfo(ctx, reqErr, rspErr)
	assert.Nil(t, err)
	assert.Equal(t, uint32(bcsCommon.AdditionErrorCode+500), rspErr.GetCode())
}

func TestGetWorkloadInfoList(t *testing.T) {
	storeServer := mock.NewMockStore()
	mockPmCli := mock.NewMockPmClient()
	resourceGetter := common.NewGetter(false, []string{}, "stag", mockPmCli)
	handler := NewBcsDataManager(storeServer, resourceGetter)
	ctx := context.Background()
	req := &bcsdatamanager.GetWorkloadInfoListRequest{ClusterID: "testCluster", Namespace: "testNs", WorkloadType: "testType"}
	rsp := &bcsdatamanager.GetWorkloadInfoListResponse{}
	err := handler.GetWorkloadInfoList(ctx, req, rsp)
	assert.Nil(t, err)
	assert.Equal(t, uint32(bcsCommon.BcsSuccess), rsp.GetCode())
	reqErr := &bcsdatamanager.GetWorkloadInfoListRequest{ClusterID: "testErr", Namespace: "testErr", WorkloadType: "testErr"}
	rspErr := &bcsdatamanager.GetWorkloadInfoListResponse{}
	err = handler.GetWorkloadInfoList(ctx, reqErr, rspErr)
	assert.Nil(t, err)
	assert.Equal(t, uint32(bcsCommon.AdditionErrorCode+500), rspErr.GetCode())
}
