/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package task

import (
	"context"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
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
		for k, v := range tasks[i].CommonParams {
			if utils.StringInSlice(strings.ToLower(k), Passwd) || utils.StringContainInSlice(v, Passwd) {
				delete(tasks[i].CommonParams, k)
			}
		}

		if len(la.req.NodeIP) > 0 {
			exist := strContains(tasks[i].NodeIPList, la.req.NodeIP)
			if exist {
				la.TaskList = append(la.TaskList, &tasks[i])
			}
		} else {
			la.TaskList = append(la.TaskList, &tasks[i])
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

// Handle handle list cluster credential
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
	return
}
