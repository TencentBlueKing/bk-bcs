/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package iam

import (
	"context"

	"bscp.io/cmd/auth-server/types"
	"bscp.io/pkg/iam/client"
	"bscp.io/pkg/kit"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbds "bscp.io/pkg/protocol/data-service"
)

// ListInstances query instances based on filter criteria.
func (i *IAM) ListInstances(kt *kit.Kit, resType client.TypeID, filter *types.ListInstanceFilter,
	page types.Page) (*types.ListInstanceResult, error) {

	bizID, pbFilter, err := filter.GetBizIDAndPbFilter()
	if err != nil {
		return nil, err
	}

	countReq := &pbds.ListInstancesReq{
		BizId:        bizID,
		ResourceType: string(resType),
		Filter:       pbFilter,
		Page:         &pbbase.BasePage{Count: true},
	}
	countResp, err := i.ds.ListInstances(kt.RpcCtx(), countReq)
	if err != nil {
		return nil, err
	}

	req := &pbds.ListInstancesReq{
		BizId:        bizID,
		ResourceType: string(resType),
		Filter:       pbFilter,
		Page:         page.PbPage(),
	}
	resp, err := i.ds.ListInstances(kt.RpcCtx(), req)
	if err != nil {
		return nil, err
	}

	instances := make([]types.InstanceResource, 0)
	for _, one := range resp.Details {
		instances = append(instances, types.InstanceResource{
			ID: types.InstanceID{
				BizID:      bizID,
				InstanceID: one.Id,
			},
			DisplayName: one.Name,
		})
	}

	result := &types.ListInstanceResult{
		Count:   countResp.Count,
		Results: instances,
	}
	return result, nil
}

// ListInstancesWithAttributes list resource instances that user is privileged to access by policy, returns id list.
func (i *IAM) ListInstancesWithAttributes(ctx context.Context, opts *client.ListWithAttributes) (idList []string,
	err error) {

	// Note implement this when attribute auth is enabled
	return make([]string, 0), nil
}
