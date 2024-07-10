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

// Package commonhandler xxx
package commonhandler

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/emicklei/go-restful"
	"gopkg.in/go-playground/validator.v9"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/operationlog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/lock"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// Validate local implementation
var Validate = validator.New()

// Handler handler for Common service
type Handler struct {
	model  store.ClusterManagerModel
	locker lock.DistributedLock
}

// NewCommonHandler create common handler
func NewCommonHandler(model store.ClusterManagerModel, locker lock.DistributedLock) *Handler {
	return &Handler{
		model:  model,
		locker: locker,
	}
}

// DownloadTaskRecords download task records
func (h *Handler) DownloadTaskRecords(request *restful.Request, response *restful.Response) {
	blog.V(3).Infof("xreq %s, host %s, url %s, src %s",
		utils.GetXRequestIDFromHTTPRequest(request.Request),
		request.Request.Host,
		request.Request.URL,
		request.Request.RemoteAddr)
	start := time.Now()
	code := 200

	resourceType := request.QueryParameter("resourceType")
	resourceID := request.QueryParameter("resourceID")
	startTimeStr := request.QueryParameter("startTime")
	startTime, err := strconv.ParseUint(startTimeStr, 10, 64)
	if err != nil {
		startTime = 0
	}
	endTimeStr := request.QueryParameter("endTime")
	endTime, err := strconv.ParseUint(endTimeStr, 10, 64)
	if err != nil {
		endTime = 0
	}
	limitStr := request.QueryParameter("limit")
	limit, err := strconv.ParseUint(limitStr, 10, 32)
	if err != nil {
		limit = 0
	}
	pageStr := request.QueryParameter("page")
	page, err := strconv.ParseUint(pageStr, 10, 32)
	if err != nil {
		page = 0
	}
	simpleStr := request.QueryParameter("simple")
	simple := false
	if simpleStr == "true" {
		simple = true
	}
	taskIDNullStr := request.QueryParameter("taskIDNull")
	taskIDNull := false
	if taskIDNullStr == "true" {
		taskIDNull = true
	}
	clusterID := request.QueryParameter("clusterID")
	projectID := request.QueryParameter("projectID")
	status := request.QueryParameter("status")
	taskType := request.QueryParameter("taskType")
	v2Str := request.QueryParameter("v2")
	v2 := false
	if v2Str == "true" {
		v2 = true
	}
	ipList := request.QueryParameter("ipList")
	taskID := request.QueryParameter("taskID")
	taskName := request.QueryParameter("taskName")
	resourceName := request.QueryParameter("resourceName")

	// 获取操作日志
	operReq := &cmproto.ListOperationLogsRequest{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		StartTime:    startTime,
		EndTime:      endTime,
		Limit:        uint32(limit),
		Page:         uint32(page),
		Simple:       simple,
		TaskIDNull:   taskIDNull,
		ClusterID:    clusterID,
		ProjectID:    projectID,
		Status:       status,
		TaskType:     taskType,
		V2:           v2,
		IpList:       ipList,
		TaskID:       taskID,
		TaskName:     taskName,
		ResourceName: resourceName,
	}
	operRsp := &cmproto.ListOperationLogsResponse{}
	operationlog.NewListOperationLogsAction(h.model).Handle(context.Background(), operReq, operRsp)
	if operRsp.Code != common.BcsErrClusterManagerSuccess {
		code = httpCodeClientError
		message := fmt.Sprintf("get operation log failed, err %s", operRsp.Message)
		blog.Warnf("get operation log failed, err %s", operRsp.Message)
		WriteClientError(response, common.BcsErrClusterManagerStoreOperationFailed, message)
		metrics.ReportAPIRequestMetric("DownloadTaskRecords", "http", strconv.Itoa(code), start)
		return
	}

	str := ""
	for _, task := range operRsp.Data.Results {
		// 获取任务步骤日志
		recordReq := &cmproto.TaskRecordsRequest{
			TaskID: task.TaskID,
		}
		recordRsp := &cmproto.TaskRecordsResponse{}
		operationlog.NewTaskRecordsAction(h.model).Handle(context.Background(), recordReq, recordRsp)
		if recordRsp.Code != common.BcsErrClusterManagerSuccess {
			code = httpCodeClientError
			message := fmt.Sprintf("get operation log failed, err %s", recordRsp.Message)
			blog.Warnf("get operation log failed, err %s", recordRsp.Message)
			WriteClientError(response, common.BcsErrClusterManagerStoreOperationFailed, message)
			metrics.ReportAPIRequestMetric("DownloadTaskRecords", "http", strconv.Itoa(code), start)
			return
		}

		str += fmt.Sprintf("%s taskID:%s %s user:%s status:%s\n",
			task.CreateTime, task.TaskID, task.Message, task.OpUser, task.Status)
		for _, step := range task.Task.StepSequence {
			stepObj := task.Task.Steps[step]
			str += fmt.Sprintf("%s %s status:%s\n", stepObj.Start, stepObj.TaskName, stepObj.Status)

			for _, record := range recordRsp.Data.Step {
				if record.Name == stepObj.TaskName {
					for _, data := range record.Data {
						str += fmt.Sprintf("%s %s %s\n",
							time.UnixMilli(data.Timestamp).Format(time.DateTime), data.Level, data.Log)
					}
				}
			}
		}
	}

	filename := fmt.Sprintf("bcs-cluster-manager-taskrecords-%s.log", time.Now().Format("20060102150405"))
	response.AddHeader("Content-Type", "application/octet-stream")
	response.AddHeader("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	response.Write([]byte(str))
}
