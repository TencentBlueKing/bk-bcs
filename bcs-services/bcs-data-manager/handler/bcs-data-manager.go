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
 */

// Package handler xxx
package handler

import (
	"context"
	"fmt"
	"time"

	bcsCommon "github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/prom"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

// BcsDataManager for handler
type BcsDataManager struct {
	mongoModel     store.Server
	tspiderModel   store.Server
	resourceGetter common.GetterInterface
}

// NewBcsDataManager create DataManager Handler
func NewBcsDataManager(mongoModel store.Server, tspiderModel store.Server, resourceGetter common.GetterInterface) *BcsDataManager {
	return &BcsDataManager{
		mongoModel:     mongoModel,
		tspiderModel:   tspiderModel,
		resourceGetter: resourceGetter,
	}
}

// GetAllProjectList get project list
// dimension: project only support day
// page: default 0
// size: default 10
// startTime: timestamp
// endTime: timestamp
func (e *BcsDataManager) GetAllProjectList(ctx context.Context,
	req *bcsdatamanager.GetAllProjectListRequest, rsp *bcsdatamanager.GetAllProjectListResponse) error {
	blog.Infof("Received GetAllProjectList.Call request. Dimension:%s, page:%d, size:%d, startTime=%s, endTime=%s",
		req.GetDimension(), req.Page, req.Size, time.Unix(req.GetStartTime(), 0),
		time.Unix(req.GetEndTime(), 0))
	start := time.Now()
	result, total, err := e.mongoModel.GetProjectList(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get project list error: %v", err)
		rsp.Code = bcsCommon.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetAllProjectList", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = bcsCommon.BcsSuccessStr
	rsp.Code = bcsCommon.BcsSuccess
	rsp.Total = uint32(total)
	prom.ReportAPIRequestMetric("GetAllProjectList", "grpc", prom.StatusOK, start)
	return nil
}

// GetProjectInfo get single project info
// projectID/projectCode/businessID is necessary
func (e *BcsDataManager) GetProjectInfo(ctx context.Context,
	req *bcsdatamanager.GetProjectInfoRequest, rsp *bcsdatamanager.GetProjectInfoResponse) error {
	blog.Infof("Received GetProjectInfo.Call request. Project id: %s, dimension:%s, startTime=%s, endTime=%s",
		req.GetProject(), req.GetDimension(), time.Unix(req.GetStartTime(), 0),
		time.Unix(req.GetEndTime(), 0))
	start := time.Now()
	if req.GetProject() == "" && req.GetProjectCode() == "" && req.GetBusiness() == "" {
		rsp.Message = "get project info error, projectId, businessID or projectCode is required"
		rsp.Code = bcsCommon.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetProjectInfo", "grpc", prom.StatusErr, start)
		return nil
	}
	if req.GetProject() == "" && req.GetBusiness() == "" && req.GetProjectCode() != "" {
		project, err := e.resourceGetter.GetProjectInfo(ctx, "", req.GetProjectCode(), nil)
		if err != nil {
			rsp.Message = fmt.Sprintf("get project info err:%v", err)
			rsp.Code = bcsCommon.AdditionErrorCode + 500
			blog.Errorf(rsp.Message)
			prom.ReportAPIRequestMetric("GetProjectInfo", "grpc", prom.StatusErr, start)
			return nil
		}
		if project == nil {
			rsp.Message = fmt.Sprintf("cannot get project info by project code:%s", req.GetProjectCode())
			rsp.Code = bcsCommon.AdditionErrorCode + 500
			blog.Errorf(rsp.Message)
			prom.ReportAPIRequestMetric("GetProjectInfo", "grpc", prom.StatusErr, start)
			return nil
		}
		req.Project = project.ProjectID
	}
	result, err := e.mongoModel.GetProjectInfo(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get project info error: %v", err)
		rsp.Code = bcsCommon.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetProjectInfo", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = bcsCommon.BcsSuccessStr
	rsp.Code = bcsCommon.BcsSuccess
	prom.ReportAPIRequestMetric("GetProjectInfo", "grpc", prom.StatusOK, start)
	return nil
}

// GetAllClusterList get all cluster list
// dimension: minute/hour/day
// page: default 0
// size: default 10
// startTime: timestamp
// endTime: timestamp
func (e *BcsDataManager) GetAllClusterList(ctx context.Context, req *bcsdatamanager.GetClusterListRequest,
	rsp *bcsdatamanager.GetClusterListResponse) error {
	blog.Infof("Received GetAllClusterList.Call request. Dimension:%s, page:%s, size:%s, startTime=%s, endTime=%s",
		req.GetDimension(), req.GetPage(), req.GetSize(), time.Unix(req.GetStartTime(), 0),
		time.Unix(req.GetEndTime(), 0))
	start := time.Now()
	result, total, err := e.mongoModel.GetClusterInfoList(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get cluster list info error: %v", err)
		rsp.Code = bcsCommon.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetAllClusterList", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = bcsCommon.BcsSuccessStr
	rsp.Code = bcsCommon.BcsSuccess
	rsp.Total = uint32(total)
	prom.ReportAPIRequestMetric("GetAllClusterList", "grpc", prom.StatusOK, start)
	return nil
}

// GetClusterListByProject get cluster list by project
// dimension: minute/hour/day
// page: default 0
// size: default 10
// startTime: timestamp
// endTime: timestamp
func (e *BcsDataManager) GetClusterListByProject(ctx context.Context, req *bcsdatamanager.GetClusterListRequest,
	rsp *bcsdatamanager.GetClusterListResponse) error {
	blog.Infof("Received GetClusterListByProject.Call request. Project id: %s, dimension:%s, page:%s, size:%s, "+
		"startTime=%s, endTime=%s", // nolint
		req.GetProject(), req.GetDimension(), req.GetPage(), req.GetSize(), time.Unix(req.GetStartTime(), 0),
		time.Unix(req.GetEndTime(), 0))
	start := time.Now()
	if req.GetProject() == "" && req.GetBusiness() == "" && req.GetProjectCode() == "" {
		rsp.Message = "get cluster list info error, projectId, projectCode or businessID is required"
		rsp.Code = bcsCommon.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetClusterListByProject", "grpc", prom.StatusErr, start)
		return nil
	}
	if req.GetProject() == "" && req.GetBusiness() == "" && req.GetProjectCode() != "" {
		project, err := e.resourceGetter.GetProjectInfo(ctx, "", req.GetProjectCode(), nil)
		if err != nil {
			rsp.Message = fmt.Sprintf("get project info err:%v", err)
			rsp.Code = bcsCommon.AdditionErrorCode + 500
			blog.Errorf(rsp.Message)
			prom.ReportAPIRequestMetric("GetClusterListByProject", "grpc", prom.StatusErr, start)
			return nil
		}
		if project == nil {
			rsp.Message = fmt.Sprintf("cannot get project info by project code:%s", req.GetProjectCode())
			rsp.Code = bcsCommon.AdditionErrorCode + 500
			blog.Errorf(rsp.Message)
			prom.ReportAPIRequestMetric("GetProjectInfo", "grpc", prom.StatusErr, start)
			return nil
		}
		req.Project = project.ProjectID
	}
	result, total, err := e.mongoModel.GetClusterInfoList(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get cluster list info error: %v", err)
		rsp.Code = bcsCommon.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetClusterListByProject", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = bcsCommon.BcsSuccessStr
	rsp.Code = bcsCommon.BcsSuccess
	rsp.Total = uint32(total)
	prom.ReportAPIRequestMetric("GetClusterListByProject", "grpc", prom.StatusOK, start)
	return nil
}

// GetClusterInfo get single cluster info
// dimension: minute/hour/day
// page: default 0
// size: default 10
// startTime: timestamp
// endTime: timestamp
// clusterID is necessary
func (e *BcsDataManager) GetClusterInfo(ctx context.Context, req *bcsdatamanager.GetClusterInfoRequest,
	rsp *bcsdatamanager.GetClusterInfoResponse) error {
	blog.Infof("Received GetClusterInfo.Call request. cluster id:%s, dimension: %s, startTime=%s, endTime=%s",
		req.GetClusterID(), req.GetDimension(), time.Unix(req.GetStartTime(), 0),
		time.Unix(req.GetEndTime(), 0))
	result, err := e.mongoModel.GetClusterInfo(ctx, req)
	start := time.Now()
	if err != nil {
		rsp.Message = fmt.Sprintf("get cluster info error: %v", err)
		rsp.Code = bcsCommon.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetClusterInfo", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = bcsCommon.BcsSuccessStr
	rsp.Code = bcsCommon.BcsSuccess
	prom.ReportAPIRequestMetric("GetClusterInfo", "grpc", prom.StatusOK, start)
	return nil
}

// GetNamespaceInfoList get namespace info list
// dimension: minute/hour/day
// page: default 0
// size: default 10
// startTime: timestamp
// endTime: timestamp
// clusterID is necessary
func (e *BcsDataManager) GetNamespaceInfoList(ctx context.Context, req *bcsdatamanager.GetNamespaceInfoListRequest,
	rsp *bcsdatamanager.GetNamespaceInfoListResponse) error {
	blog.Infof("Received GetNamespaceInfoList.Call request. cluster id:%s, dimension: %s, page:%s, size:%s, "+
		"startTime=%s, endTime=%s",
		req.GetClusterID(), req.GetDimension(), req.GetPage(), req.GetSize(), time.Unix(req.GetStartTime(), 0),
		time.Unix(req.GetEndTime(), 0))
	start := time.Now()
	result, total, err := e.mongoModel.GetNamespaceInfoList(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get namespace list info error: %v", err)
		rsp.Code = bcsCommon.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetNamespaceInfoList", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = bcsCommon.BcsSuccessStr
	rsp.Code = bcsCommon.BcsSuccess
	rsp.Total = uint32(total)
	prom.ReportAPIRequestMetric("GetNamespaceInfoList", "grpc", prom.StatusOK, start)
	return nil
}

// GetNamespaceInfo get single namespace info
// dimension: minute/hour/day
// page: default 0
// size: default 10
// startTime: timestamp
// endTime: timestamp
// clusterID and namespace is necessary
func (e *BcsDataManager) GetNamespaceInfo(ctx context.Context, req *bcsdatamanager.GetNamespaceInfoRequest,
	rsp *bcsdatamanager.GetNamespaceInfoResponse) error {
	blog.Infof("Received GetNamespaceInfo.Call request. cluster id:%s, namespace:%s, dimension: %s, "+
		"startTime=%s, endTime=%s",
		req.GetClusterID(), req.Namespace, req.Dimension, time.Unix(req.GetStartTime(), 0),
		time.Unix(req.GetEndTime(), 0))
	start := time.Now()
	result, err := e.mongoModel.GetNamespaceInfo(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get namespace info error: %v", err)
		rsp.Code = bcsCommon.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetNamespaceInfo", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = bcsCommon.BcsSuccessStr
	rsp.Code = bcsCommon.BcsSuccess
	prom.ReportAPIRequestMetric("GetNamespaceInfo", "grpc", prom.StatusOK, start)
	return nil
}

// GetWorkloadInfoList get workload info list
// dimension: minute/hour/day
// page: default 0
// size: default 10
// startTime: timestamp
// endTime: timestamp
// clusterID is necessary
func (e *BcsDataManager) GetWorkloadInfoList(ctx context.Context, req *bcsdatamanager.GetWorkloadInfoListRequest,
	rsp *bcsdatamanager.GetWorkloadInfoListResponse) error {
	blog.Infof("Received GetWorkloadInfoList.Call request, cluster id: %s, namespace: %s, dimension: %s, "+
		"type: %s, page: %s, size: %s, startTime=%s, endTime=%s",
		req.GetClusterID(), req.GetNamespace(), req.GetDimension(), req.GetWorkloadType(), req.GetPage(), req.GetSize(),
		time.Unix(req.GetStartTime(), 0), time.Unix(req.GetEndTime(), 0))
	start := time.Now()
	result, total, err := e.mongoModel.GetWorkloadInfoList(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get workload list info error: %v", err)
		rsp.Code = bcsCommon.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetWorkloadInfoList", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = bcsCommon.BcsSuccessStr
	rsp.Code = bcsCommon.BcsSuccess
	rsp.Total = uint32(total)
	prom.ReportAPIRequestMetric("GetWorkloadInfoList", "grpc", prom.StatusOK, start)
	return nil
}

// GetWorkloadInfo get single workload info
// dimension: minute/hour/day
// page: default 0
// size: default 10
// startTime: timestamp
// endTime: timestamp
// clusterID, namespace, workloadType and workloadName is necessary
func (e *BcsDataManager) GetWorkloadInfo(ctx context.Context, req *bcsdatamanager.GetWorkloadInfoRequest,
	rsp *bcsdatamanager.GetWorkloadInfoResponse) error {
	blog.Infof("Received GetWorkloadInfo.Call request. cluster id: %s, namespace: %s, dimension: %s, "+
		"type: %s, name: %s, startTime=%s, endTime=%s",
		req.GetClusterID(), req.GetNamespace(), req.Dimension, req.GetWorkloadType(), req.GetWorkloadName(),
		time.Unix(req.GetStartTime(), 0), time.Unix(req.GetEndTime(), 0))
	start := time.Now()
	result, err := e.mongoModel.GetWorkloadInfo(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get workload info error: %v", err)
		rsp.Code = bcsCommon.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetWorkloadInfo", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = bcsCommon.BcsSuccessStr
	rsp.Code = bcsCommon.BcsSuccess
	prom.ReportAPIRequestMetric("GetWorkloadInfo", "grpc", prom.StatusOK, start)
	return nil
}

// GetPodAutoscalerList get pod autoscaler list
// dimension: minute/hour/day
// page: default 0
// size: default 10
// startTime: timestamp
// endTime: timestamp
func (e *BcsDataManager) GetPodAutoscalerList(ctx context.Context, req *bcsdatamanager.GetPodAutoscalerListRequest,
	rsp *bcsdatamanager.GetPodAutoscalerListResponse) error {
	blog.Infof("Received GetPodAutoscalerList.Call request. cluster id: %s, namespace: %s, dimension: %s, "+
		"workloadType: %s, workloadName: %s, podAutoscalerType:%s, page:%d, size:%d, startTime=%s, endTime=%s",
		req.GetClusterID(), req.GetNamespace(), req.Dimension, req.GetWorkloadType(), req.GetWorkloadName(),
		req.GetPodAutoscalerType(), req.GetPage(), req.GetSize(), time.Unix(req.GetStartTime(), 0),
		time.Unix(req.GetEndTime(), 0))
	start := time.Now()
	result, total, err := e.mongoModel.GetPodAutoscalerList(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get workload info error: %v", err)
		rsp.Code = bcsCommon.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetPodAutoscalerList", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = bcsCommon.BcsSuccessStr
	rsp.Code = bcsCommon.BcsSuccess
	rsp.Total = uint32(total)
	prom.ReportAPIRequestMetric("GetPodAutoscalerList", "grpc", prom.StatusOK, start)
	return nil
}

// GetPodAutoscaler get pod autoscaler
// dimension: minute/hour/day, default minute
// page: default 0
// size: default 10
// startTime: timestamp
// endTime: timestamp
// clusterID, namespace, podAutoscalerType and podAutoscalerName is necessary
func (e *BcsDataManager) GetPodAutoscaler(ctx context.Context, req *bcsdatamanager.GetPodAutoscalerRequest,
	rsp *bcsdatamanager.GetPodAutoscalerResponse) error {
	blog.Infof("Received GetPodAutoscaler.Call request. cluster id: %s, namespace: %s, dimension: %s, "+
		"type: %s, name: %s, startTime=%s, endTime=%s",
		req.GetClusterID(), req.GetNamespace(), req.Dimension, req.GetPodAutoscalerType(), req.GetPodAutoscalerName(),
		time.Unix(req.GetStartTime(), 0), time.Unix(req.GetEndTime(), 0))
	start := time.Now()
	result, err := e.mongoModel.GetPodAutoscalerInfo(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get podAutoscaler info error: %v", err)
		rsp.Code = bcsCommon.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetPodAutoscaler", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = bcsCommon.BcsSuccessStr
	rsp.Code = bcsCommon.BcsSuccess
	prom.ReportAPIRequestMetric("GetPodAutoscaler", "grpc", prom.StatusOK, start)
	return nil
}

// GetPowerTrading get operations data of powertrading
// table: table name
// startTime: timestamp
// endTime: timestamp
// params: others params
func (e *BcsDataManager) GetPowerTrading(ctx context.Context, req *bcsdatamanager.GetPowerTradingDataRequest,
	rsp *bcsdatamanager.GetPowerTradingDataResponse) error {
	blog.Infof("Received GetPowerTrading.Call request. table name: %s, startTime: %s, endTime: %s, params: %+v",
		req.GetTable(), req.GetStartTime(), req.GetEndTime(), req.GetParams())

	start := time.Now()

	var storeDB store.Server
	if req.GetPreferStorage() == store.TspiderStore {
		// tspider store
		storeDB = e.tspiderModel
	} else {
		// default store
		storeDB = e.mongoModel
	}

	result, total, err := storeDB.GetPowerTradingInfo(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get powerTrading info error: %v", err)
		rsp.Code = bcsCommon.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetPowerTrading", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = bcsCommon.BcsSuccessStr
	rsp.Code = bcsCommon.BcsSuccess
	rsp.Total = uint32(total)
	rsp.Page = req.GetPage()
	rsp.Size = req.GetSize()
	prom.ReportAPIRequestMetric("GetPowerTrading", "grpc", prom.StatusOK, start)
	return nil
}

// GetCloudNativeWorkloadList get cloud native workloads list
// currentPage: min 1
// pageSize: max 10000
func (e *BcsDataManager) GetCloudNativeWorkloadList(ctx context.Context,
	req *bcsdatamanager.GetCloudNativeWorkloadListRequest, rsp *bcsdatamanager.GetCloudNativeWorkloadListResponse) error {
	blog.Infof("Received GetCloudNativeWorkload.Call request. PageSize: %d, CurrentPage: %d", req.GetPageSize(), req.GetCurrentPage())

	start := time.Now()
	result, err := e.tspiderModel.GetCloudNativeWorkloadList(ctx, req)
	if err != nil {
		rsp.Msg = fmt.Sprintf("Get cloud native workloads error, err: %s", err.Error())
		rsp.Code = 500
		blog.Errorf(rsp.Msg)
		prom.ReportAPIRequestMetric("GetCloudNativeWorkloadList", "grpc", prom.StatusErr, start)
		return nil
	}

	rsp.Code = bcsCommon.BcsSuccess
	rsp.Msg = bcsCommon.BcsSuccessStr
	rsp.Data = result

	//appid和data和platform在函数内部填充
	rsp.Data.Status = rsp.Code
	rsp.Data.Message = rsp.Msg
	rsp.Data.PageSize = req.GetPageSize()
	rsp.Data.CurrentPage = req.GetCurrentPage()
	rsp.Data.Timestamp = time.Now().Format("2006-01-02 15:04:05")

	return nil
}

// GetUserOperationDataList get bcs user operation data list
// type: project/cluster/operation
// startTime: unix timestamp
// endTime: unix timestamp
// page: min 1
// size: min 1
// projectCode is necessary
func (e *BcsDataManager) GetUserOperationDataList(ctx context.Context,
	req *bcsdatamanager.GetUserOperationDataListRequest, rsp *bcsdatamanager.GetUserOperationDataListResponse) error {

	start := time.Now()
	result, total, err := e.tspiderModel.GetUserOperationDataList(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get bcs user opration data error: %v", err)
		rsp.Code = bcsCommon.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetUserOperationDataList", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = bcsCommon.BcsSuccessStr
	rsp.Code = bcsCommon.BcsSuccess
	rsp.Total = uint32(total)

	prom.ReportAPIRequestMetric("GetUserOperationDataList", "grpc", prom.StatusOK, start)
	return nil
}

// GetWorkloadRequestResult get workload request result
func (e *BcsDataManager) GetWorkloadRequestResult(ctx context.Context, req *bcsdatamanager.GetWorkloadRequestRecommendResultReq,
	rsp *bcsdatamanager.GetWorkloadRequestRecommendResultRsp) error {
	blog.Infof("Received GetWorkloadRequestResult.Call request. cluster id: %s, namespace: %s, workloadType:%s"+
		"workloadName:%s",
		req.GetClusterID(), req.GetNamespace(), req.GetWorkloadType(), req.GetWorkloadName())
	start := time.Now()
	result, err := e.mongoModel.GetLatestWorkloadRequest(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("get workloadRequestResult info error: %v", err)
		rsp.Code = bcsCommon.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetWorkloadRequestResult", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result.GetData()
	rsp.Message = bcsCommon.BcsSuccessStr
	rsp.Code = bcsCommon.BcsSuccess
	prom.ReportAPIRequestMetric("GetWorkloadRequestResult", "grpc", prom.StatusOK, start)
	return nil
}

// GetWorkloadOriginRequestResult get workload origin request result
func (e *BcsDataManager) GetWorkloadOriginRequestResult(ctx context.Context, req *bcsdatamanager.GetWorkloadOriginRequestResultReq,
	rsp *bcsdatamanager.GetWorkloadOriginRequestResultRsp) error {
	blog.Infof("Received GetWorkloadOriginRequestResult.Call request. projectID:%s, cluster id: %s, "+
		"namespace: %s, workloadType:%s, workloadName:%s",
		req.ProjectID, req.GetClusterID(), req.GetNamespace(), req.GetWorkloadType(), req.GetWorkloadName())
	start := time.Now()
	result, err := e.mongoModel.ListWorkloadOriginRequest(ctx, req)
	if err != nil {
		rsp.Message = fmt.Sprintf("GetWorkloadOriginRequestResult info error: %v", err)
		rsp.Code = bcsCommon.AdditionErrorCode + 500
		blog.Errorf(rsp.Message)
		prom.ReportAPIRequestMetric("GetWorkloadOriginRequestResult", "grpc", prom.StatusErr, start)
		return nil
	}
	rsp.Data = result
	rsp.Message = bcsCommon.BcsSuccessStr
	rsp.Code = bcsCommon.BcsSuccess
	prom.ReportAPIRequestMetric("GetWorkloadOriginRequestResult", "grpc", prom.StatusOK, start)
	return nil
}
