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

// RemoveNodes 指定IP地址从节点组中移除,但是该节点仍然在集群中
func (c *NodeGroupMgr) RemoveNodes(req types.RemoveNodesFromGroupReq) (types.RemoveNodesFromGroupResp, error) {
	var (
		resp types.RemoveNodesFromGroupResp
		err  error
	)

	servResp, err := c.client.RemoveNodesFromGroup(c.ctx, &clustermanager.RemoveNodesFromGroupRequest{
		ClusterID:   req.ClusterID,
		Nodes:       req.Nodes,
		NodeGroupID: req.NodeGroupID,
	})

	if err != nil {
		return resp, err
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	resp.TaskID = servResp.GetData().GetTaskID()

	return resp, nil
}
