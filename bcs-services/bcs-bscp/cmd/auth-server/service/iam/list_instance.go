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
	"context"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/auth-server/types"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/client"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbds "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// ListInstances query instances based on filter criteria.
func (i *IAM) ListInstances(kt *kit.Kit, resType client.TypeID, filter *types.ListInstanceFilter,
	page types.Page) (*types.ListInstanceResult, error) {

	req := &pbds.ListInstancesReq{
		ResourceType: string(resType),
		Page:         page.PbPage(),
	}
	if filter.Parent != nil {
		req.ParentType = string(filter.Parent.Type)
		req.ParentId = filter.Parent.ID
	}
	resp, err := i.ds.ListInstances(kt.RpcCtx(), req)
	if err != nil {
		return nil, err
	}

	instances := make([]types.InstanceResource, 0)
	for _, one := range resp.Details {
		instances = append(instances, types.InstanceResource{
			ID:          one.Id,
			DisplayName: one.Name,
		})
	}

	result := &types.ListInstanceResult{
		Count:   resp.Count,
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
