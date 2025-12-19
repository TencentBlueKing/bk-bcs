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

// Package operationlog xxxx
package operationlog

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	autils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/util"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
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

	if ua.req.GetProjectID() != "" && ua.req.GetClusterID() != "" {
		cluster, err := ua.model.GetCluster(ua.ctx, ua.req.GetClusterID())
		if err != nil {
			return err
		}

		if cluster.GetProjectID() != ua.req.GetProjectID() {
			return fmt.Errorf("project[%s] not match cluster[%s]", ua.req.GetProjectID(), ua.req.GetClusterID())
		}
	}

	return nil
}

func (ua *ListOperationLogsAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ua *ListOperationLogsAction) filterOperationLogs() ([]*cmproto.TaskOperationLog, int, error) {
	var (
		conds   = make([]bson.E, 0)
		condDst = make([]bson.E, 0)
	)
	if ua.req.ResourceType != "" {
		conds = append(conds, util.Condition(operator.Eq, "resourcetype", []string{ua.req.ResourceType}))
	}
	if ua.req.ResourceID != "" {
		conds = append(conds, util.Condition(operator.Eq, "resourceid", []string{ua.req.ResourceID}))
	}
	if ua.req.ClusterID != "" {
		conds = append(conds, util.Condition(operator.Eq, "clusterid", []string{ua.req.ClusterID}))
	}
	if ua.req.ProjectID != "" {
		conds = append(conds, util.Condition(operator.Eq, "projectid", []string{ua.req.ProjectID}))
	}
	if ua.req.TaskID != "" {
		conds = append(conds, util.Condition(operator.Eq, "taskid", []string{ua.req.TaskID}))
	}
	if ua.req.TaskName != "" {
		conds = append(conds, util.Condition(operator.Eq, "taskname", []string{ua.req.TaskName}))
	}
	if ua.req.ResourceName != "" {
		conds = append(conds, util.Condition(util.Regex, "resourcename", []string{ua.req.ResourceName}))
	}
	if ua.req.OpUser != "" {
		conds = append(conds, util.Condition(operator.Eq, "opuser", []string{ua.req.OpUser}))
	}

	if ua.req.StartTime > 0 && ua.req.EndTime > 0 {
		// time range condition
		start := time.Unix(int64(ua.req.StartTime), 0).Format(time.RFC3339)
		end := time.Unix(int64(ua.req.EndTime), 0).Format(time.RFC3339)
		conds = append(conds, util.Condition(util.Range, "createtime", []string{start, end}))
	}

	// default taskID empty filter
	if !ua.req.TaskIDNull {
		conds = append(conds, util.Condition(operator.Ne, "taskid", []string{""}))
	}

	if len(ua.req.IpList) > 0 {
		ipList := strings.Split(ua.req.IpList, ",")
		condDst = append(condDst, util.Condition(operator.In, "nodeiplist", ipList))
	}
	if ua.req.Status != "" {
		condDst = append(condDst, util.Condition(operator.Eq, "status", []string{ua.req.Status}))
	}
	if ua.req.TaskType != "" {
		condDst = append(condDst, util.Condition(util.Regex, "tasktype", []string{ua.req.TaskType}))
	}

	sumLogs, err := ua.model.ListAggreOperationLog(ua.ctx, conds, condDst, &options.ListOption{
		Count: true,
	})
	if err != nil {
		return nil, 0, err
	}

	count := len(sumLogs)
	offset := (ua.req.Page - 1) * ua.req.Limit
	sort := map[string]int{"createtime": -1}

	opLogs, err := ua.model.ListAggreOperationLog(ua.ctx, conds, condDst, &options.ListOption{
		Limit: int64(ua.req.Limit), Offset: int64(offset), Sort: sort})
	if err != nil {
		return nil, 0, err
	}

	return opLogs, count, nil
}

func (ua *ListOperationLogsAction) fetchV2OperationLogs() error { // nolint
	opLogs, count, err := ua.filterOperationLogs()
	if err != nil {
		return err
	}

	ua.resp.Data = &cmproto.ListOperationLogsResponseData{
		Count:   uint32(count),
		Results: []*cmproto.OperationLogDetail{},
	}
	var taskIDs []string
	for _, v := range opLogs {
		if len(v.TaskID) > 0 {
			taskIDs = append(taskIDs, v.TaskID)
		}

		createTime := utils.TransTimeFormat(v.CreateTime)
		ua.resp.Data.Results = append(ua.resp.Data.Results, &cmproto.OperationLogDetail{
			ResourceType: v.ResourceType,
			ResourceID:   v.ResourceID,
			TaskID:       v.TaskID,
			Message:      v.Message,
			OpUser:       v.OpUser,
			CreateTime:   createTime,
			TaskType:     v.TaskType,
			Status:       v.Status,
			ResourceName: v.GetResourceName(),
		})
	}

	// get task
	if len(taskIDs) == 0 || ua.req.Simple {
		return nil
	}
	return ua.appendTasks(taskIDs)
}

func (ua *ListOperationLogsAction) fetchV1OperationLogs() error {
	// resource condition
	cond := operator.M{"resourcetype": ua.req.ResourceType}
	if len(ua.req.ResourceID) != 0 {
		cond["resourceid"] = ua.req.ResourceID
	}
	if len(ua.req.ProjectID) != 0 {
		cond["projectid"] = ua.req.ProjectID
	}
	if len(ua.req.ClusterID) != 0 {
		cond["clusterid"] = ua.req.ClusterID
	}

	resourceCond := operator.NewLeafCondition(operator.Eq, cond)
	// time range condition
	start := time.Unix(int64(ua.req.StartTime), 0).Format(time.RFC3339)
	end := time.Unix(int64(ua.req.EndTime), 0).Format(time.RFC3339)

	startTimeCond := operator.NewLeafCondition(operator.Gte, operator.M{"createtime": start})
	endTimeCond := operator.NewLeafCondition(operator.Lte, operator.M{"createtime": end})
	conds := []*operator.Condition{resourceCond, startTimeCond, endTimeCond}

	// default taskID empty filter
	if !ua.req.TaskIDNull {
		taskCond := operator.NewLeafCondition(operator.Ne, operator.M{"taskid": ""})
		conds = append(conds, taskCond)
	}
	logsCond := operator.NewBranchCondition(operator.And, conds...)

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
		if len(v.TaskID) == 0 {
			continue
		}
		taskIDs = append(taskIDs, v.TaskID)

		createTime := utils.TransTimeFormat(v.CreateTime)
		ua.resp.Data.Results = append(ua.resp.Data.Results, &cmproto.OperationLogDetail{
			ResourceType: v.ResourceType,
			ResourceID:   v.ResourceID,
			TaskID:       v.TaskID,
			Message:      v.Message,
			OpUser:       v.OpUser,
			CreateTime:   createTime,
			ResourceName: v.ResourceName,
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
		taskMap[tasks[i].TaskID] = tasks[i]
	}
	for i, v := range ua.resp.Data.Results {
		if t, ok := taskMap[v.TaskID]; ok {
			// remove sensitive info
			t.CommonParams = nil
			for i := range t.Steps {
				for k := range t.Steps[i].Params {
					if utils.StringInSlice(k,
						[]string{cloudprovider.BkSopsTaskURLKey.String(), cloudprovider.ShowSopsURLKey.String()}) {
						continue
					}
					delete(t.Steps[i].Params, k)
				}
				t.Steps[i].TaskName = autils.Translate(ua.ctx, t.Steps[i].TaskMethod,
					t.Steps[i].TaskName, t.Steps[i].Translate)
				if t.Steps[i].Start != "" {
					t.Steps[i].Start = utils.TransTimeFormat(t.Steps[i].Start)
				}
				if t.Steps[i].End != "" {
					t.Steps[i].End = utils.TransTimeFormat(t.Steps[i].End)
				}
			}
			startTime := utils.TransTimeFormat(t.Start)
			endTime := utils.TransTimeFormat(t.End)
			t.Start = startTime
			t.End = endTime

			t.TaskName = autils.Translate(ua.ctx, t.TaskType, t.TaskName, "")
			ua.resp.Data.Results[i].TaskType = t.TaskType

			allowRetry := true
			// attention: 开启CA节点自动扩缩容的任务不允许手动重试
			if utils.SliceContainInString([]string{cloudprovider.UpdateNodeGroupDesiredNode.String(),
				cloudprovider.CleanNodeGroupNodes.String()}, t.TaskType) &&
				t.GetCommonParams()[cloudprovider.ManualKey.String()] != common.True {
				allowRetry = false
			}

			if autils.CheckTaskStepPartFailureStatus(t) {
				ua.resp.Data.Results[i].Status = cloudprovider.TaskStatusPartFailure
			}

			ua.resp.Data.Results[i].AllowRetry = allowRetry
			ua.resp.Data.Results[i].Message = autils.TranslateMsg(ua.ctx, v.ResourceType, v.TaskType, v.Message, t)
			ua.resp.Data.Results[i].Task = t
		}
	}
	return nil
}

// Handle handles list operation logs
func (ua *ListOperationLogsAction) Handle(
	ctx context.Context, req *cmproto.ListOperationLogsRequest, resp *cmproto.ListOperationLogsResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list operation logs failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	err := ua.validate()
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if ua.req.V2 {
		err = ua.fetchV2OperationLogs()
	} else {
		err = ua.fetchV1OperationLogs()
	}
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
