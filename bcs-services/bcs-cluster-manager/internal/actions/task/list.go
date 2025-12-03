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

package task

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	iutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ListAction action for list online cluster credential
type ListAction struct {
	ctx        context.Context
	model      store.ClusterManagerModel
	req        *cmproto.ListTaskRequest
	resp       *cmproto.ListTaskResponse
	TaskList   []*cmproto.Task
	LatestTask *cmproto.Task
}

// NewListAction create list action for cluster credential
func NewListAction(model store.ClusterManagerModel) *ListAction {
	return &ListAction{
		model: model,
	}
}

func (la *ListAction) listTask() error {
	condM := make(operator.M)

	if len(la.req.ClusterID) != 0 {
		condM["clusterid"] = la.req.ClusterID
	}
	if len(la.req.ProjectID) != 0 {
		condM["projectid"] = la.req.ProjectID
	}
	if len(la.req.Creator) != 0 {
		condM["creator"] = la.req.Creator
	}
	if len(la.req.Updater) != 0 {
		condM["updater"] = la.req.Updater
	}
	if len(la.req.TaskType) != 0 {
		condM["tasktype"] = la.req.TaskType
	}
	if len(la.req.Status) != 0 {
		condM["status"] = la.req.Status
	}
	if len(la.req.NodeGroupID) != 0 {
		condM["nodegroupid"] = la.req.NodeGroupID
	}

	cond := operator.NewLeafCondition(operator.Eq, condM)
	// default listTask descending sort by start
	tasks, err := la.model.ListTask(la.ctx, cond, &storeopt.ListOption{
		Sort: map[string]int{
			"start": -1,
		},
	})
	if err != nil {
		return err
	}
	for i := range tasks {
		utils.HandleTaskStepData(la.ctx, tasks[i])
		// actions.FormatTaskTime(&tasks[i])

		if len(la.req.NodeIP) > 0 {
			exist := iutils.StringContainInSlice(la.req.NodeIP, tasks[i].NodeIPList)
			if exist {
				la.TaskList = append(la.TaskList, tasks[i])
			}
		} else {
			la.TaskList = append(la.TaskList, tasks[i])
		}
	}

	if len(la.TaskList) > 0 {
		la.LatestTask = la.TaskList[0]
	}

	return nil
}

func (la *ListAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.TaskList
	la.resp.LatestTask = la.LatestTask
}

// Handle list cluster credential
func (la *ListAction) Handle(
	ctx context.Context, req *cmproto.ListTaskRequest, resp *cmproto.ListTaskResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list Task failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listTask(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// ListV2Action action for list online cluster credential
type ListV2Action struct {
	ctx      context.Context
	model    store.ClusterManagerModel
	req      *cmproto.ListTaskV2Request
	resp     *cmproto.ListTaskV2Response
	TaskList []*cmproto.Task
	count    int64
}

// NewListV2Action create list action for cluster credential
func NewListV2Action(model store.ClusterManagerModel) *ListV2Action {
	return &ListV2Action{
		model: model,
	}
}

func (la *ListV2Action) listV2Task() error {
	condM := make(operator.M)
	condArr := make([]*operator.Condition, 0)

	if len(la.req.ClusterID) != 0 {
		condM["clusterid"] = la.req.ClusterID
	}
	if len(la.req.ProjectID) != 0 {
		condM["projectid"] = la.req.ProjectID
	}
	if len(la.req.Creator) != 0 {
		condM["creator"] = la.req.Creator
	}
	if len(la.req.Updater) != 0 {
		condM["updater"] = la.req.Updater
	}
	if len(la.req.TaskType) != 0 {
		condM["tasktype"] = la.req.TaskType
	}
	if len(la.req.Status) != 0 {
		condM["status"] = la.req.Status
	}
	if len(la.req.NodeGroupID) != 0 {
		condM["nodegroupid"] = la.req.NodeGroupID
	}

	if len(condM) > 0 {
		condArr = append(condArr, operator.NewLeafCondition(operator.Eq, condM))
	}

	if la.req.StartTime > 0 && la.req.EndTime > 0 {
		condS := make(operator.M)
		condT := make(operator.M)
		condS["start"] = time.Unix(int64(la.req.StartTime), 0).Format(time.RFC3339)
		condT["start"] = time.Unix(int64(la.req.EndTime), 0).Format(time.RFC3339)
		condArr = append(condArr, operator.NewLeafCondition(operator.Gte, condS))
		condArr = append(condArr, operator.NewLeafCondition(operator.Lte, condT))
	}

	// default listTask descending sort by start
	tasks, count, err := la.model.ListTaskWithCount(la.ctx, operator.NewBranchCondition(operator.And, condArr...),
		&storeopt.ListOption{
			Offset: int64((la.req.Page - 1) * la.req.Limit),
			Limit:  int64(la.req.Limit),
			Sort: map[string]int{
				"start": -1,
			},
		})
	if err != nil {
		return err
	}
	for i := range tasks {
		utils.HandleTaskStepData(la.ctx, tasks[i])
		// actions.FormatTaskTime(&tasks[i])

		if len(la.req.NodeIP) > 0 {
			exist := iutils.StringContainInSlice(la.req.NodeIP, tasks[i].NodeIPList)
			if exist {
				la.TaskList = append(la.TaskList, tasks[i])
			} else {
				// 如果ip过滤不通过，总数需要减一
				count--
			}
		} else {
			la.TaskList = append(la.TaskList, tasks[i])
		}
	}
	la.count = count

	return nil
}

func (la *ListV2Action) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = &cmproto.ListTaskV2ResponseData{
		Count:   uint32(la.count),
		Results: la.TaskList,
	}
}

// Handle list cluster credential
func (la *ListV2Action) Handle(
	ctx context.Context, req *cmproto.ListTaskV2Request, resp *cmproto.ListTaskV2Response) {
	if req == nil || resp == nil {
		blog.Errorf("list Task failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listV2Task(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
