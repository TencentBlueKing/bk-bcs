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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	itypes "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

// ListClusterTaskMetricsAction action for list online cluster credential
type ListClusterTaskMetricsAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.ListClusterTaskMetricsRequest
	resp  *cmproto.ListClusterTaskMetricsResponse
}

// NewListClusterTaskMetricsAction create list action for cluster task metrics
func NewListClusterTaskMetricsAction(model store.ClusterManagerModel) *ListClusterTaskMetricsAction {
	return &ListClusterTaskMetricsAction{
		model: model,
	}
}

func (la *ListClusterTaskMetricsAction) listTaskMetrics() error {

	tasks, err := la.model.ListTaskMetrics(la.ctx, la.req.ClusterID, la.req.Start, la.req.End, &storeopt.ListOption{})
	if err != nil {
		return err
	}
	metricsData := convertMetricsData(tasks)
	la.resp.Data = metricsData
	return nil
}

// convertMetricsData convert metrics data to proto data
func convertMetricsData(tasks []*itypes.ClusterTaskMetrics) []*cmproto.ClusterTask {
	metricsData := []*cmproto.ClusterTask{}
	metricsDataUnique := map[string][]*cmproto.ClusterTaskMetrics{}
	for i := range tasks {
		metricsDataUnique[tasks[i].ClusterId] = append(metricsDataUnique[tasks[i].ClusterId],
			&cmproto.ClusterTaskMetrics{
				SuccessRate:      tasks[i].SuccessRate,
				ExecutionTime:    tasks[i].AvgExecutionTime,
				FailureTopReason: tasks[i].TopFailReason,
				TaskType:         tasks[i].TaskType,
			})
	}

	for clusterID, metrics := range metricsDataUnique {
		metricsData = append(metricsData, &cmproto.ClusterTask{
			ClusterId:   clusterID,
			TaskMetrics: metrics,
		})
	}
	return metricsData
}

func (la *ListClusterTaskMetricsAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle list cluster task metrics
func (la *ListClusterTaskMetricsAction) Handle(
	ctx context.Context, req *cmproto.ListClusterTaskMetricsRequest, resp *cmproto.ListClusterTaskMetricsResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list Task Metrics failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listTaskMetrics(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
