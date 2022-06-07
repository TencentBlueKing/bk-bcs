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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/prom"
	"time"

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

// GetAllProjectList get project list
func (e *BcsDataManager) GetAllProjectList(ctx context.Context,
	req *bcsdatamanager.GetAllProjectListRequest, rsp *bcsdatamanager.GetAllProjectListResponse) error {
	blog.Infof("Received GetAllProjectList.Call request. Dimension:%s, page:%d, size:%d",
		req.GetDimension(), req.Page, req.Size)
	start := time.Now()
	result, total, err := e.model.GetProjectList(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get project list error: %v", err)
		rsp.Code = common.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetAllProjectList", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = common.BcsSuccessStr
	rsp.Code = common.BcsSuccess
	rsp.Total = uint32(total)
	prom.ReportAPIRequestMetric("GetAllProjectList", "grpc", prom.StatusOK, start)
	return nil
}

// GetProjectInfo get project info
func (e *BcsDataManager) GetProjectInfo(ctx context.Context,
	req *bcsdatamanager.GetProjectInfoRequest, rsp *bcsdatamanager.GetProjectInfoResponse) error {
	blog.Infof("Received GetProjectInfo.Call request. Project id: %s, dimension:%s",
		req.GetProject(), req.GetDimension())
	start := time.Now()
	if req.GetProject() == "" && req.GetBusiness() == "" {
		rsp.Message = fmt.Sprintf("get project info error, projectId or businessID is required")
		rsp.Code = common.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetProjectInfo", "grpc", prom.StatusErr, start)
		return nil
	}
	result, err := e.model.GetProjectInfo(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get project info error: %v", err)
		rsp.Code = common.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetProjectInfo", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = common.BcsSuccessStr
	rsp.Code = common.BcsSuccess
	prom.ReportAPIRequestMetric("GetProjectInfo", "grpc", prom.StatusOK, start)
	return nil
}

// GetAllClusterList get all cluster list
func (e *BcsDataManager) GetAllClusterList(ctx context.Context, req *bcsdatamanager.GetClusterListRequest,
	rsp *bcsdatamanager.GetClusterListResponse) error {
	blog.Infof("Received GetAllClusterList.Call request. Dimension:%s, page:%s, size:%s",
		req.GetDimension(), req.GetPage(), req.GetSize())
	start := time.Now()
	result, total, err := e.model.GetClusterInfoList(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get cluster list info error: %v", err)
		rsp.Code = common.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetAllClusterList", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = common.BcsSuccessStr
	rsp.Code = common.BcsSuccess
	rsp.Total = uint32(total)
	prom.ReportAPIRequestMetric("GetAllClusterList", "grpc", prom.StatusOK, start)
	return nil
}

// GetClusterListByProject get cluster list by project
func (e *BcsDataManager) GetClusterListByProject(ctx context.Context, req *bcsdatamanager.GetClusterListRequest,
	rsp *bcsdatamanager.GetClusterListResponse) error {
	blog.Infof("Received GetClusterListByProject.Call request. Project id: %s, dimension:%s, page:%s, size:%s",
		req.GetProject(), req.GetDimension(), req.GetPage(), req.GetSize())
	start := time.Now()
	if req.GetProject() == "" && req.GetBusiness() == "" {
		rsp.Message = fmt.Sprintf("get cluster list info error, projectId or businessID is required")
		rsp.Code = common.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetClusterListByProject", "grpc", prom.StatusErr, start)
		return nil
	}
	result, total, err := e.model.GetClusterInfoList(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get cluster list info error: %v", err)
		rsp.Code = common.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetClusterListByProject", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = common.BcsSuccessStr
	rsp.Code = common.BcsSuccess
	rsp.Total = uint32(total)
	prom.ReportAPIRequestMetric("GetClusterListByProject", "grpc", prom.StatusOK, start)
	return nil
}

// GetClusterInfo get cluster info
func (e *BcsDataManager) GetClusterInfo(ctx context.Context, req *bcsdatamanager.GetClusterInfoRequest,
	rsp *bcsdatamanager.GetClusterInfoResponse) error {
	blog.Infof("Received GetClusterInfo.Call request. cluster id:%s, dimension: %s",
		req.GetClusterID(), req.GetDimension())
	result, err := e.model.GetClusterInfo(ctx, req)
	start := time.Now()
	if err != nil {
		rsp.Message = fmt.Sprintf("get cluster info error: %v", err)
		rsp.Code = common.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetClusterInfo", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = common.BcsSuccessStr
	rsp.Code = common.BcsSuccess
	prom.ReportAPIRequestMetric("GetClusterInfo", "grpc", prom.StatusOK, start)
	return nil
}

// GetNamespaceInfoList get namespace info list
func (e *BcsDataManager) GetNamespaceInfoList(ctx context.Context, req *bcsdatamanager.GetNamespaceInfoListRequest,
	rsp *bcsdatamanager.GetNamespaceInfoListResponse) error {
	blog.Infof("Received GetNamespaceInfoList.Call request. cluster id:%s, dimension: %s, page:%s, size:%s",
		req.GetClusterID(), req.GetDimension(), req.GetPage(), req.GetSize())
	start := time.Now()
	result, total, err := e.model.GetNamespaceInfoList(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get namespace list info error: %v", err)
		rsp.Code = common.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetNamespaceInfoList", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = common.BcsSuccessStr
	rsp.Code = common.BcsSuccess
	rsp.Total = uint32(total)
	prom.ReportAPIRequestMetric("GetNamespaceInfoList", "grpc", prom.StatusOK, start)
	return nil
}

// GetNamespaceInfo get namespace info
func (e *BcsDataManager) GetNamespaceInfo(ctx context.Context, req *bcsdatamanager.GetNamespaceInfoRequest,
	rsp *bcsdatamanager.GetNamespaceInfoResponse) error {
	blog.Infof("Received GetNamespaceInfo.Call request. cluster id:%s, namespace:%s, dimension: %s",
		req.GetClusterID(), req.Namespace, req.Dimension)
	start := time.Now()
	result, err := e.model.GetNamespaceInfo(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get namespace info error: %v", err)
		rsp.Code = common.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetNamespaceInfo", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = common.BcsSuccessStr
	rsp.Code = common.BcsSuccess
	prom.ReportAPIRequestMetric("GetNamespaceInfo", "grpc", prom.StatusOK, start)
	return nil
}

// GetWorkloadInfoList get workload info list
func (e *BcsDataManager) GetWorkloadInfoList(ctx context.Context, req *bcsdatamanager.GetWorkloadInfoListRequest,
	rsp *bcsdatamanager.GetWorkloadInfoListResponse) error {
	blog.Infof("Received GetWorkloadInfoList.Call request, cluster id: %s, namespace: %s, dimension: %s, "+
		"type: %s, page: %s, size: %s",
		req.GetClusterID(), req.GetNamespace(), req.GetDimension(), req.GetWorkloadType(), req.GetPage(), req.GetSize())
	start := time.Now()
	result, total, err := e.model.GetWorkloadInfoList(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get workload list info error: %v", err)
		rsp.Code = common.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetWorkloadInfo", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = common.BcsSuccessStr
	rsp.Code = common.BcsSuccess
	rsp.Total = uint32(total)
	prom.ReportAPIRequestMetric("GetWorkloadInfo", "grpc", prom.StatusOK, start)
	return nil
}

// GetWorkloadInfo get workload info
func (e *BcsDataManager) GetWorkloadInfo(ctx context.Context, req *bcsdatamanager.GetWorkloadInfoRequest,
	rsp *bcsdatamanager.GetWorkloadInfoResponse) error {
	blog.Infof("Received GetWorkloadInfo.Call request. cluster id: %s, namespace: %s, dimension: %s, "+
		"type: %s, name: %s",
		req.GetClusterID(), req.GetNamespace(), req.Dimension, req.GetWorkloadType(), req.GetWorkloadName())
	start := time.Now()
	result, err := e.model.GetWorkloadInfo(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get workload info error: %v", err)
		rsp.Code = common.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetWorkloadInfo", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = common.BcsSuccessStr
	rsp.Code = common.BcsSuccess
	prom.ReportAPIRequestMetric("GetWorkloadInfo", "grpc", prom.StatusOK, start)
	return nil
}
