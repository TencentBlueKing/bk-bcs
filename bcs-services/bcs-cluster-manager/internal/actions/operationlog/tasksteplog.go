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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ListTaskStepLogsAction action for list task step logs
type ListTaskStepLogsAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.ListTaskStepLogsRequest
	resp  *cmproto.ListTaskStepLogsResponse
}

// NewListTaskStepLogsAction create action
func NewListTaskStepLogsAction(model store.ClusterManagerModel) *ListTaskStepLogsAction {
	return &ListTaskStepLogsAction{
		model: model,
	}
}

func (ua *ListTaskStepLogsAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	// default page and limit
	if ua.req.GetPage() <= 0 {
		ua.req.Page = 0
	}
	if ua.req.GetLimit() <= 0 {
		ua.req.Limit = 10
	}
	return nil
}

func (ua *ListTaskStepLogsAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ua *ListTaskStepLogsAction) fetchTaskStepLogs() error {
	// resource condition
	cond := operator.M{"taskid": ua.req.TaskID}
	if len(ua.req.StepName) != 0 {
		cond["stepname"] = ua.req.StepName
	}
	logsCond := operator.NewLeafCondition(operator.Eq, cond)

	// list operation logs
	count, err := ua.model.CountTaskStepLog(ua.ctx, logsCond)
	if err != nil {
		return err
	}
	offset := (ua.req.Page - 1) * ua.req.Limit
	sort := map[string]int{"createtime": 1}
	opLogs, err := ua.model.ListTaskStepLog(ua.ctx, logsCond, &options.ListOption{
		Limit: int64(ua.req.Limit), Offset: int64(offset), Sort: sort})
	if err != nil {
		return err
	}
	ua.resp.Data = &cmproto.ListTaskStepLogsResponseData{
		Count:   uint32(count),
		Results: []*cmproto.TaskStepLogDetail{},
	}

	for _, v := range opLogs {
		if len(v.TaskID) == 0 {
			continue
		}

		createTime := utils.TransTimeFormat(v.CreateTime)
		ua.resp.Data.Results = append(ua.resp.Data.Results, &cmproto.TaskStepLogDetail{
			TaskID:     v.TaskID,
			StepName:   v.StepName,
			Level:      v.Level,
			Message:    v.Message,
			CreateTime: createTime,
		})
	}

	return nil
}

// Handle handles list task step logs
func (ua *ListTaskStepLogsAction) Handle(
	ctx context.Context, req *cmproto.ListTaskStepLogsRequest, resp *cmproto.ListTaskStepLogsResponse) {
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

	err = ua.fetchTaskStepLogs()
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
