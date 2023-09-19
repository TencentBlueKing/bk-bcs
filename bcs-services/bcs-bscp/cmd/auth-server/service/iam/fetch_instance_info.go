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

package iam

import (
	"bscp.io/cmd/auth-server/types"
	"bscp.io/pkg/iam/client"
	"bscp.io/pkg/kit"
	pbds "bscp.io/pkg/protocol/data-service"
)

// FetchInstanceInfo obtain resource instance details in batch.
func (i *IAM) FetchInstanceInfo(kt *kit.Kit, resType client.TypeID, ft *types.FetchInstanceInfoFilter) (
	[]*types.InstanceInfo, error) {

	// Note: f.Attrs need to deal with, if add attribute authentication.

	req := &pbds.FetchInstanceInfoReq{
		ResourceType: string(resType),
		Ids:          ft.IDs,
	}
	resp, err := i.ds.FetchInstanceInfo(kt.RpcCtx(), req)
	if err != nil {
		return nil, err
	}
	results := make([]*types.InstanceInfo, 0)
	for _, detail := range resp.Details {
		results = append(results, &types.InstanceInfo{
			ID:            detail.Id,
			DisplayName:   detail.DisplayName,
			BKIAMApprover: detail.Approver,
			BKIAMPath:     detail.Path,
		})
	}
	return results, nil
}
