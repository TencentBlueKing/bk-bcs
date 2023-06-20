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

package task

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// Create 创建任务
func (c *TaskMgr) Create(req types.CreateTaskReq) (types.CreateTaskResp, error) {
	var (
		resp types.CreateTaskResp
		err  error
	)

	steps := make(map[string]*clustermanager.Step)
	for k, v := range req.Steps {
		steps[k] = &clustermanager.Step{
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
		}
	}

	servResp, err := c.client.CreateTask(c.ctx, &clustermanager.CreateTaskRequest{
		TaskType:     req.TaskType,
		Status:       req.Status,
		StepSequence: req.StepSequence,
		Steps:        steps,
		ClusterID:    req.ClusterID,
		ProjectID:    req.ProjectID,
		Creator:      "bcs",
	})
	if err != nil {
		return resp, err
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	resp.TaskID = servResp.Data.TaskID

	return resp, nil
}
