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
	"fmt"

	"bscp.io/cmd/auth-server/types"
	"bscp.io/pkg/iam/client"
	"bscp.io/pkg/kit"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/runtime/filter"
)

// FetchInstanceInfo obtain resource instance details in batch.
func (i *IAM) FetchInstanceInfo(kt *kit.Kit, resType client.TypeID, ft *types.FetchInstanceInfoFilter) (
	[]map[string]interface{}, error) {

	// Note: f.Attrs need to deal with, if add attribute authentication.

	groups := make(map[uint32][]uint32, 0)
	for _, id := range ft.IDs {
		groups[id.BizID] = append(groups[id.BizID], id.InstanceID)
	}

	results := make([]map[string]interface{}, 0)
	for bizID, ids := range groups {
		expr := &filter.Expression{
			Op:    filter.And,
			Rules: make([]filter.RuleFactory, 0),
		}
		expr.Rules = append(expr.Rules, &filter.AtomRule{
			Field: "id",
			Op:    filter.In.Factory(),
			Value: ids,
		})
		pbFt, err := expr.MarshalPB()
		if err != nil {
			return nil, err
		}

		req := &pbds.ListInstancesReq{
			BizId:        bizID,
			ResourceType: string(resType),
			Filter:       pbFt,
			Page:         &pbbase.BasePage{Count: false},
		}
		resp, err := i.ds.ListInstances(kt.RpcCtx(), req)
		if err != nil {
			return nil, err
		}

		for _, one := range resp.Details {
			result := make(map[string]interface{}, 0)
			result[types.IDField] = types.InstanceID{
				BizID:      bizID,
				InstanceID: one.Id,
			}
			result[types.NameField] = one.Name
			result[types.ResTopology] = fmt.Sprintf("/biz,%d/", bizID)
			results = append(results, result)
		}
	}

	return results, nil
}
