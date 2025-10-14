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

	wholeTasks, err := la.model.ListClusterWholeTaskMetrics(la.ctx, la.req.Start, la.req.End)
	if err != nil {
		return err
	}
	subSuccessTasks, err := la.model.ListClusterSubSuccessTaskMetrics(la.ctx, la.req.Start, la.req.End)
	if err != nil {
		return err
	}
	subFailTasks, err := la.model.ListClusterSubFailTaskMetrics(la.ctx, la.req.Start, la.req.End)
	if err != nil {
		return err
	}
	metricsData := convertClusterMetricsData(wholeTasks, subSuccessTasks, subFailTasks)
	la.resp.Data = metricsData
	return nil
}

// ClusterSubTaskUniq sub task metrics unique
type ClusterSubTaskUniq map[string]ClusterSubTaskMetricsTopReason

// ClusterSubTaskMetricsTopReason sub task metrics top reason
type ClusterSubTaskMetricsTopReason struct {
	cmproto.ClusterSubTaskMetrics
	Reason      string
	MaxFailTask int
}

// convertClusterMetricsData convert metrics data to proto data
func convertClusterMetricsData(wholeTasks []*itypes.ClusterWholeTaskMetrics,
	subSuccessTasks []*itypes.ClusterSubSuccessTaskMetrics,
	subFailTasks []*itypes.ClusterSubFailTaskMetrics) []*cmproto.ClusterTaskMetrics {

	metricsData := []*cmproto.ClusterTaskMetrics{}
	metricsDataUnique := map[string]ClusterSubTaskUniq{}
	for i := range wholeTasks {
		metricsDataUnique[wholeTasks[i].ClusterId] = ClusterSubTaskUniq{}
	}

	for i := range subSuccessTasks {
		if _, ok := metricsDataUnique[subSuccessTasks[i].ClusterId]; ok {
			metricsDataUnique[subSuccessTasks[i].ClusterId][subSuccessTasks[i].TaskType] = ClusterSubTaskMetricsTopReason{
				ClusterSubTaskMetrics: cmproto.ClusterSubTaskMetrics{
					SubSuccessRate:       subSuccessTasks[i].SuccessRate,
					SubAvgExecutionTime:  subSuccessTasks[i].AvgExecutionTime,
					SubTaskFails:         int32(subSuccessTasks[i].FailTasks),
					SubTaskFailReason:    []string{},
					SubTaskFailTopReason: "",
					TaskType:             subSuccessTasks[i].TaskType,
				},
			}
		}
	}

	for i := range subFailTasks {
		if sub, ok := metricsDataUnique[subFailTasks[i].ClusterId]; ok {
			if stm, ok := sub[subFailTasks[i].TaskType]; ok {
				stm.SubTaskFailReason = append(stm.SubTaskFailReason, subFailTasks[i].Message)
				if stm.MaxFailTask < subFailTasks[i].FailTasks {
					stm.MaxFailTask = subFailTasks[i].FailTasks
					stm.SubTaskFailTopReason = subFailTasks[i].Message
				}
				sub[subFailTasks[i].TaskType] = stm
			}
		}
	}

	for i := range wholeTasks {
		if sub, ok := metricsDataUnique[wholeTasks[i].ClusterId]; ok {
			subTaskMetrics := []*cmproto.ClusterSubTaskMetrics{}
			for _, s := range sub {
				subTaskMetrics = append(subTaskMetrics, &cmproto.ClusterSubTaskMetrics{
					SubSuccessRate:       s.ClusterSubTaskMetrics.SubSuccessRate,
					SubAvgExecutionTime:  s.ClusterSubTaskMetrics.SubAvgExecutionTime,
					SubTaskFails:         s.ClusterSubTaskMetrics.SubTaskFails,
					SubTaskFailReason:    s.ClusterSubTaskMetrics.SubTaskFailReason,
					SubTaskFailTopReason: s.ClusterSubTaskMetrics.SubTaskFailTopReason,
					TaskType:             s.ClusterSubTaskMetrics.TaskType,
				})
			}
			metricsData = append(metricsData, &cmproto.ClusterTaskMetrics{
				ClusterId:        wholeTasks[i].ClusterId,
				SubTaskMetrics:   subTaskMetrics,
				SuccessRate:      wholeTasks[i].SuccessRate,
				AvgExecutionTime: wholeTasks[i].AvgExecutionTime,
			})
		}
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
		blog.Errorf("list cluster Task Metrics failed, req or resp is empty")
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

// ListBusinessTaskMetricsAction action for list online cluster credential
type ListBusinessTaskMetricsAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.ListBusinessTaskMetricsRequest
	resp  *cmproto.ListBusinessTaskMetricsResponse
}

// NewListBusinessTaskMetricsAction create list action for business task metrics
func NewListBusinessTaskMetricsAction(model store.ClusterManagerModel) *ListBusinessTaskMetricsAction {
	return &ListBusinessTaskMetricsAction{
		model: model,
	}
}

func (la *ListBusinessTaskMetricsAction) listTaskMetrics() error {

	wholeTasks, err := la.model.ListBusinessWholeTaskMetrics(la.ctx, la.req.Start, la.req.End)
	if err != nil {
		return err
	}
	subSuccessTasks, err := la.model.ListBusinessSubSuccessTaskMetrics(la.ctx, la.req.Start, la.req.End)
	if err != nil {
		return err
	}
	subFailTasks, err := la.model.ListBusinessSubFailTaskMetrics(la.ctx, la.req.Start, la.req.End)
	if err != nil {
		return err
	}
	metricsData := convertBusinessMetricsData(wholeTasks, subSuccessTasks, subFailTasks)
	la.resp.Data = metricsData
	return nil
}

// BusinessSubTaskUniq sub task metrics unique
type BusinessSubTaskUniq map[string]BusinessSubTaskMetricsTopReason

// BusinessSubTaskMetricsTopReason sub task metrics top reason
type BusinessSubTaskMetricsTopReason struct {
	cmproto.BussinessSubTaskMetrics
	Reason      string
	MaxFailTask int
}

// convertBusinessMetricsData convert metrics data to proto data
func convertBusinessMetricsData(wholeTasks []*itypes.BusinessWholeTaskMetrics,
	subSuccessTasks []*itypes.BusinessSubSuccessTaskMetrics,
	subFailTasks []*itypes.BusinessSubFailTaskMetrics) []*cmproto.BusinessTaskMetrics {

	metricsData := []*cmproto.BusinessTaskMetrics{}
	metricsDataUnique := map[string]BusinessSubTaskUniq{}
	for i := range wholeTasks {
		metricsDataUnique[wholeTasks[i].BusinessId] = BusinessSubTaskUniq{}
	}

	for i := range subSuccessTasks {
		if _, ok := metricsDataUnique[subSuccessTasks[i].BusinessId]; ok {
			metricsDataUnique[subSuccessTasks[i].BusinessId][subSuccessTasks[i].TaskType] = BusinessSubTaskMetricsTopReason{
				BussinessSubTaskMetrics: cmproto.BussinessSubTaskMetrics{
					SubSuccessRate:       subSuccessTasks[i].SuccessRate,
					SubAvgExecutionTime:  subSuccessTasks[i].AvgExecutionTime,
					SubTaskFails:         int32(subSuccessTasks[i].FailTasks),
					SubTaskFailReason:    []string{},
					SubTaskFailTopReason: "",
					TaskType:             subSuccessTasks[i].TaskType,
				},
			}
		}
	}

	for i := range subFailTasks {
		if sub, ok := metricsDataUnique[subFailTasks[i].BusinessId]; ok {
			if stm, ok := sub[subFailTasks[i].TaskType]; ok {
				stm.SubTaskFailReason = append(stm.SubTaskFailReason, subFailTasks[i].Message)
				if stm.MaxFailTask < subFailTasks[i].FailTasks {
					stm.MaxFailTask = subFailTasks[i].FailTasks
					stm.SubTaskFailTopReason = subFailTasks[i].Message
				}
				sub[subFailTasks[i].TaskType] = stm
			}
		}
	}

	for i := range wholeTasks {
		if sub, ok := metricsDataUnique[wholeTasks[i].BusinessId]; ok {
			subTaskMetrics := []*cmproto.BussinessSubTaskMetrics{}
			for _, s := range sub {
				subTaskMetrics = append(subTaskMetrics, &cmproto.BussinessSubTaskMetrics{
					SubSuccessRate:       s.BussinessSubTaskMetrics.SubSuccessRate,
					SubAvgExecutionTime:  s.BussinessSubTaskMetrics.SubAvgExecutionTime,
					SubTaskFails:         s.BussinessSubTaskMetrics.SubTaskFails,
					SubTaskFailReason:    s.BussinessSubTaskMetrics.SubTaskFailReason,
					SubTaskFailTopReason: s.BussinessSubTaskMetrics.SubTaskFailTopReason,
					TaskType:             s.BussinessSubTaskMetrics.TaskType,
				})
			}
			metricsData = append(metricsData, &cmproto.BusinessTaskMetrics{
				BusinessId:       wholeTasks[i].BusinessId,
				SubTaskMetrics:   subTaskMetrics,
				SuccessRate:      wholeTasks[i].SuccessRate,
				AvgExecutionTime: wholeTasks[i].AvgExecutionTime,
			})
		}
	}
	return metricsData
}

func (la *ListBusinessTaskMetricsAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle list cluster task metrics
func (la *ListBusinessTaskMetricsAction) Handle(
	ctx context.Context, req *cmproto.ListBusinessTaskMetricsRequest, resp *cmproto.ListBusinessTaskMetricsResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list business Task Metrics failed, req or resp is empty")
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
