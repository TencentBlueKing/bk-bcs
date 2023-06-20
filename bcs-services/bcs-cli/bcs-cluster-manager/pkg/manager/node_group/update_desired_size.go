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

// UpdateDesiredSize 更新节点池期望的节点数,扩容前保证数据一致性
func (c *NodeGroupMgr) UpdateDesiredSize(req types.UpdateGroupDesiredSizeReq) error {
	resp, err := c.client.UpdateGroupDesiredSize(c.ctx, &clustermanager.UpdateGroupDesiredSizeRequest{
		NodeGroupID: req.NodeGroupID,
		DesiredSize: req.DesiredSize,
		Operator:    "bcs",
	})

	if err != nil {
		return err
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}
