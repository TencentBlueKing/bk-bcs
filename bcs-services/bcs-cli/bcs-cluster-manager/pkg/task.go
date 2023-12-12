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

package pkg

import (
	"fmt"

	clsapi "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

var (
	getTaskURL = "/bcsapi/v4/clustermanager/v1/task/%s"
)

// GetTask get task status by taskID
func (c *ClusterMgrClient) GetTask(req *clsapi.GetTaskRequest) (
	*clsapi.GetTaskResponse, error) {
	if len(req.TaskID) == 0 {
		return nil, fmt.Errorf("lost task ID")
	}
	totalURL := fmt.Sprintf(getTaskURL, req.TaskID)
	resp := &clsapi.GetTaskResponse{}
	if err := c.Get(totalURL, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
