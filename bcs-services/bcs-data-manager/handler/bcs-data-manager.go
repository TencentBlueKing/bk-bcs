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

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

// BcsDataManager for handler
type BcsDataManager struct {
	model store.Server
}

// NewBcsDataManager create DataManager Handler
func NewBcsDataManager(model store.Server) *BcsDataManager {
	return &BcsDataManager{
		model: model,
	}
}

// GetProjectInfo get project info
func (e *BcsDataManager) GetProjectInfo(ctx context.Context, req *bcsdatamanager.GetProjectInfoRequest,
	rsp *bcsdatamanager.GetProjectInfoResponse) error {
	blog.Infof("Received GetProjectInfo.Call request. Project id: %s, dimension:%s",
		req.GetProjectID(), req.GetDimension())
	result, err := e.model.GetProjectInfo(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get project info error: %v", err)
		rsp.Code = common.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		return nil
	}
	rsp.Data = result
	rsp.Message = common.BcsSuccessStr
	rsp.Code = common.BcsSuccess
	return nil
}

// GetClusterInfoList get cluster info list
func (e *BcsDataManager) GetClusterInfoList(ctx context.Context, req *bcsdatamanager.GetClusterInfoListRequest,
	rsp *bcsdatamanager.GetClusterInfoListResponse) error {
	blog.Infof("Received GetClusterInfoList.Call request. Project id: %s, dimension:%s, page:%s, size:%s",
		req.GetProjectID(), req.GetDimension(), req.GetPage(), req.GetSize())
	result, total, err := e.model.GetClusterInfoList(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get cluster list info error: %v", err)
		rsp.Code = common.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		return nil
	}
	rsp.Data = result
	rsp.Message = common.BcsSuccessStr
	rsp.Code = common.BcsSuccess
	rsp.Total = uint32(total)
	return nil
}

// GetClusterInfo get cluster info
func (e *BcsDataManager) GetClusterInfo(ctx context.Context, req *bcsdatamanager.GetClusterInfoRequest,
	rsp *bcsdatamanager.GetClusterInfoResponse) error {
	blog.Infof("Received GetClusterInfo.Call request. cluster id:%s, dimension: %s",
		req.GetClusterID(), req.GetDimension())
	result, err := e.model.GetClusterInfo(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get cluster info error: %v", err)
		rsp.Code = common.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		return nil
	}
	rsp.Data = result
	rsp.Message = common.BcsSuccessStr
	rsp.Code = common.BcsSuccess
	return nil
}

// GetNamespaceInfoList get namespace info list
func (e *BcsDataManager) GetNamespaceInfoList(ctx context.Context, req *bcsdatamanager.GetNamespaceInfoListRequest,
	rsp *bcsdatamanager.GetNamespaceInfoListResponse) error {
	blog.Infof("Received GetNamespaceInfoList.Call request. cluster id:%s, dimension: %s, page:%s, size:%s",
		req.GetClusterID(), req.GetDimension(), req.GetPage(), req.GetSize())
	result, total, err := e.model.GetNamespaceInfoList(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get namespace list info error: %v", err)
		rsp.Code = common.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		return nil
	}
	rsp.Data = result
	rsp.Message = common.BcsSuccessStr
	rsp.Code = common.BcsSuccess
	rsp.Total = uint32(total)
	return nil
}

// GetNamespaceInfo get namespace info
func (e *BcsDataManager) GetNamespaceInfo(ctx context.Context, req *bcsdatamanager.GetNamespaceInfoRequest,
	rsp *bcsdatamanager.GetNamespaceInfoResponse) error {
	blog.Infof("Received GetNamespaceInfo.Call request. cluster id:%s, namespace:%s, dimension: %s",
		req.GetClusterID(), req.Namespace, req.Dimension)
	result, err := e.model.GetNamespaceInfo(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get namespace info error: %v", err)
		rsp.Code = common.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		return nil
	}
	rsp.Data = result
	rsp.Message = common.BcsSuccessStr
	rsp.Code = common.BcsSuccess
	return nil
}

// GetWorkloadInfoList get workload info list
func (e *BcsDataManager) GetWorkloadInfoList(ctx context.Context, req *bcsdatamanager.GetWorkloadInfoListRequest,
	rsp *bcsdatamanager.GetWorkloadInfoListResponse) error {
	blog.Infof("Received GetWorkloadInfoList.Call request, cluster id: %s, namespace: %s, dimension: %s, "+
		"type: %s, page: %s, size: %s",
		req.GetClusterID(), req.GetNamespace(), req.GetDimension(), req.GetWorkloadType(), req.GetPage(), req.GetSize())
	result, total, err := e.model.GetWorkloadInfoList(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get workload list info error: %v", err)
		rsp.Code = common.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		return nil
	}
	rsp.Data = result
	rsp.Message = common.BcsSuccessStr
	rsp.Code = common.BcsSuccess
	rsp.Total = uint32(total)
	return nil
}

// GetWorkloadInfo get workload info
func (e *BcsDataManager) GetWorkloadInfo(ctx context.Context, req *bcsdatamanager.GetWorkloadInfoRequest,
	rsp *bcsdatamanager.GetWorkloadInfoResponse) error {
	blog.Infof("Received GetWorkloadInfo.Call request. cluster id: %s, namespace: %s, dimension: %s, "+
		"type: %s, name: %s",
		req.GetClusterID(), req.GetNamespace(), req.Dimension, req.GetWorkloadType(), req.GetWorkloadName())
	result, err := e.model.GetWorkloadInfo(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get workload info error: %v", err)
		rsp.Code = common.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		return nil
	}
	rsp.Data = result
	rsp.Message = common.BcsSuccessStr
	rsp.Code = common.BcsSuccess
	return nil
}
