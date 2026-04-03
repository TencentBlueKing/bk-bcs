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

// Package operation operation operate
package operation

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/utils"

	clustermgr "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs/clustermanager"
)

// OperationAction operation action interface
type OperationAction interface { // nolint
	ListOperationLogs(ctx context.Context, req *types.ListOperationLogsReq) (*types.ListOperationLogsResp, error)
}

// Action action for operation
type Action struct{}

// NewOperationAction new operation action
func NewOperationAction() OperationAction {
	return &Action{}
}

// ListOperationLogs list operation logs
func (a *Action) ListOperationLogs(ctx context.Context, req *types.ListOperationLogsReq) (
	*types.ListOperationLogsResp, error) {
	logs, err := clustermgr.ListOperationLogs(ctx, &clustermanager.ListOperationLogsRequest{
		V2:           req.V2,
		TaskIDNull:   req.TaskIDNull,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		ResourceID:   req.ResourceID,
		ResourceName: req.ResourceName,
		Status:       req.Status,
		OpUser:       req.OpUser,
		Page:         req.Page,
		Limit:        req.Limit,
		ResourceType: req.ResourceType,
		Simple:       req.Simple,
		ClusterID:    req.ClusterID,
		ProjectID:    req.ProjectID,
		TaskType:     req.TaskType,
		IpList:       req.IpList,
		TaskID:       req.TaskID,
		TaskName:     req.TaskName,
	})
	if err != nil {
		return nil, utils.SystemError(err)
	}

	result := &types.ListOperationLogsResp{
		Count:   logs.Count,
		Results: make([]*types.OperationLogDetail, 0),
	}
	for _, log := range logs.Results {
		result.Results = append(result.Results, &types.OperationLogDetail{
			TaskID:       log.TaskID,
			Status:       log.Status,
			Message:      log.Message,
			ResourceID:   log.ResourceID,
			OpUser:       log.OpUser,
			CreateTime:   log.CreateTime,
			ResourceName: log.ResourceName,
			Task: func() *types.Task {
				if log.Task != nil {
					return &types.Task{
						TaskID:        log.Task.TaskID,
						TaskType:      log.Task.TaskType,
						Status:        log.Task.Status,
						Message:       log.Task.Message,
						Start:         log.Task.Start,
						End:           log.Task.End,
						ExecutionTime: log.Task.ExecutionTime,
						CurrentStep:   log.Task.CurrentStep,
						StepSequence:  log.Task.StepSequence,
						Steps: func() map[string]*types.Step {
							steps := make(map[string]*types.Step)
							for k, v := range log.Task.Steps {
								steps[k] = &types.Step{
									Name:          v.Name,
									System:        v.System,
									Link:          v.Link,
									Params:        v.Params,
									Retry:         v.Retry,
									Start:         v.Start,
									End:           v.End,
									ExecutionTime: v.ExecutionTime,
									Status:        v.Status,
									Message:       v.Message,
									LastUpdate:    v.LastUpdate,
									TaskMethod:    v.TaskMethod,
									TaskName:      v.TaskName,
									SkipOnFailed:  v.SkipOnFailed,
									Translate:     v.Translate,
									AllowSkip:     v.AllowSkip,
									MaxRetry:      v.MaxRetry,
								}
							}
							return steps
						}(),
						ClusterID:      log.Task.ClusterID,
						ProjectID:      log.Task.ProjectID,
						Creator:        log.Task.Creator,
						LastUpdate:     log.Task.LastUpdate,
						Updater:        log.Task.Updater,
						ForceTerminate: log.Task.ForceTerminate,
						TaskName:       log.Task.TaskName,
						NodeGroupID:    log.Task.NodeGroupID,
					}
				}
				return nil
			}(),
		})
	}

	return result, nil
}
