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

package cloudaccount

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// ListToPerm 查询云凭证列表, 主要用于权限资源查询
func (c *CloudAccountMgr) ListToPerm(req types.ListCloudAccountToPermReq) (types.ListCloudAccountToPermResp, error) {
	var (
		resp types.ListCloudAccountToPermResp
		err  error
	)

	servResp, err := c.client.ListCloudAccountToPerm(c.ctx, &clustermanager.ListCloudAccountPermRequest{})
	if err != nil {
		return resp, err
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	resp.Data = make([]*types.CloudAccount, 0)

	for _, v := range servResp.Data {
		resp.Data = append(resp.Data, &types.CloudAccount{
			CloudID:     v.CloudID,
			ProjectID:   v.ProjectID,
			AccountID:   v.AccountID,
			AccountName: v.AccountName,
			Account: types.Account{
				SecretID:          v.Account.SecretID,
				SecretKey:         v.Account.SecretKey,
				SubscriptionID:    v.Account.SubscriptionID,
				TenantID:          v.Account.TenantID,
				ResourceGroupName: v.Account.ResourceGroupName,
				ClientID:          v.Account.ClientID,
				ClientSecret:      v.Account.ClientSecret,
			},
			Enable:    v.Enable,
			Creator:   v.Creator,
			CreatTime: v.CreatTime,
		})
	}

	return resp, nil
}
