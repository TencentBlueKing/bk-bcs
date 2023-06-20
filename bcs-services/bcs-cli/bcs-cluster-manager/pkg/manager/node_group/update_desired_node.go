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

package nodegroup

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// UpdateDesiredNode 更新节点池中DesiredNode信息
func (c *NodeGroupMgr) UpdateDesiredNode(req types.UpdateGroupDesiredNodeReq) (types.UpdateGroupDesiredNodeResp, error) {
	var (
		resp types.UpdateGroupDesiredNodeResp
		err  error
	)

	servResp, err := c.client.UpdateGroupDesiredNode(c.ctx, &clustermanager.UpdateGroupDesiredNodeRequest{
		NodeGroupID: req.NodeGroupID,
		DesiredNode: req.DesiredNode,
		Operator:    "bcs",
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
