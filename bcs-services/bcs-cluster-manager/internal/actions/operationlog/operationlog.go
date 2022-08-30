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

// Package operationlog xxx
package operationlog

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// ListOperationLogsAction action for list operation logs
type ListOperationLogsAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.ListOperationLogsRequest
	resp  *cmproto.ListOperationLogsResponse
}

// NewListOperationLogsAction create action
func NewListOperationLogsAction(model store.ClusterManagerModel) *ListOperationLogsAction {
	return &ListOperationLogsAction{
		model: model,
	}
}

func (ua *ListOperationLogsAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	return nil
}

func (ua *ListOperationLogsAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ua *ListOperationLogsAction) fetchOperationLogs() error {
	// resource condition
	cond := operator.M{"resourcetype": ua.req.ResourceType}
	if len(ua.req.ResourceID) != 0 {
		cond["resourceid"] = ua.req.ResourceID
	}
	resourceCond := operator.NewLeafCondition(operator.Eq, cond)
	// time range condition
	start := time.Unix(int64(ua.req.StartTime), 0).Format(time.RFC3339)
	end := time.Unix(int64(ua.req.EndTime), 0).Format(time.RFC3339)
	startTimeCond := operator.NewLeafCondition(operator.Gte, operator.M{"createtime": start})
	endTimeCond := operator.NewLeafCondition(operator.Lte, operator.M{"createtime": end})
	logsCond := operator.NewBranchCondition(operator.And, resourceCond, startTimeCond, endTimeCond)

	// list operation logs
	count, err := ua.model.CountOperationLog(ua.ctx, logsCond)
	if err != nil {
		return err
	}
	offset := (ua.req.Page - 1) * ua.req.Limit
	sort := map[string]int{"createtime": -1}
	opLogs, err := ua.model.ListOperationLog(ua.ctx, logsCond, &options.ListOption{
		Limit: int64(ua.req.Limit), Offset: int64(offset), Sort: sort})
	if err != nil {
		return err
	}
	ua.resp.Data = &cmproto.ListOperationLogsResponseData{
		Count:   uint32(count),
		Results: []*cmproto.OperationLogDetail{},
	}
	var taskIDs []string
	for _, v := range opLogs {
		if len(v.TaskID) != 0 {
			taskIDs = append(taskIDs, v.TaskID)
		}
		createTime, err := common.Format3399ToLocalTime(v.CreateTime)
		if err != nil {
			blog.Warnf("parse time failed, err: %s", err.Error())
		}
		ua.resp.Data.Results = append(ua.resp.Data.Results, &cmproto.OperationLogDetail{
			ResourceType: v.ResourceType,
			ResourceID:   v.ResourceID,
			TaskID:       v.TaskID,
			Message:      v.Message,
			OpUser:       v.OpUser,
			CreateTime:   createTime,
		})
	}

	// get task
	if len(taskIDs) == 0 || ua.req.Simple {
		return nil
	}
	return ua.appendTasks(taskIDs)
}

func (ua *ListOperationLogsAction) appendTasks(taskIDs []string) error {
	taskCond := operator.NewLeafCondition(operator.In, operator.M{"taskid": taskIDs})
	tasks, err := ua.model.ListTask(ua.ctx, taskCond, &options.ListOption{All: true})
	if err != nil {
		return err
	}
	taskMap := make(map[string]*cmproto.Task, 0)
	for i := range tasks {
		taskMap[tasks[i].TaskID] = &tasks[i]
	}
	for i, v := range ua.resp.Data.Results {
		if len(v.TaskID) == 0 {
			continue
		}
		if t, ok := taskMap[v.TaskID]; ok {
			// remove sensitive info
			t.CommonParams = nil
			for i := range t.Steps {
				t.Steps[i].Params = nil
			}
			startTime, err := common.Format3399ToLocalTime(t.Start)
			if err != nil {
				blog.Warnf("parse time failed, err: %s", err.Error())
			}
			endTime, err := common.Format3399ToLocalTime(t.End)
			if err != nil {
				blog.Warnf("parse time failed, err: %s", err.Error())
			}
			t.Start = startTime
			t.End = endTime
			ua.resp.Data.Results[i].Task = t
		}
	}
	return nil
}

// Handle handles list operation logs
func (ua *ListOperationLogsAction) Handle(ctx context.Context, req *cmproto.ListOperationLogsRequest,
	resp *cmproto.ListOperationLogsResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list operation logs failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ua.fetchOperationLogs(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
