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

package cloudvpc

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// Delete 删除云私有网络
func (c *CloudVPCMgr) Delete(req types.DeleteCloudVPCReq) error {
	resp, err := c.client.DeleteCloudVPC(c.ctx, &clustermanager.DeleteCloudVPCRequest{
		CloudID: req.CloudID,
		VpcID:   req.VPCID,
	})
	if err != nil {
		return err
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}
