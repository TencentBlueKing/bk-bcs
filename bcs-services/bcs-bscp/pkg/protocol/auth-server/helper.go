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

package pbas

import (
	"encoding/json"
	"fmt"

	pbstruct "github.com/golang/protobuf/ptypes/struct"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/auth-server/types"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/client"
)

// Validate InitAuthCenterReq.
func (r *InitAuthCenterReq) Validate() error {
	if len(r.Host) == 0 {
		return errf.New(errf.InvalidParameter, "host is required")
	}

	return nil
}

// PullResourceReq convert pb PullResourceReq to types PullResourceReq.
func (r *PullResourceReq) PullResourceReq() (*types.PullResourceReq, error) {
	req := &types.PullResourceReq{
		Type:   client.TypeID(r.Type),
		Method: types.Method(r.Method),
	}

	if r.Page != nil {
		if err := r.Page.Validate(); err != nil {
			return nil, err
		}
		req.Page = types.Page{
			Offset: uint(r.Page.Offset),
			Limit:  uint(r.Page.Limit),
		}
	}

	if r.Filter == nil {
		return req, nil
	}

	jsonFilter, err := r.Filter.MarshalJSON()
	if err != nil {
		return nil, err
	}

	if len(jsonFilter) == 0 {
		return req, nil
	}

	switch req.Method {
	case types.ListAttrValueMethod:
		filter := types.ListAttrValueFilter{}
		err := json.Unmarshal(jsonFilter, &filter)
		if err != nil {
			return nil, err
		}
		req.Filter = filter

	case types.ListInstanceMethod, types.SearchInstanceMethod:
		filter := types.ListInstanceFilter{}
		err := json.Unmarshal(jsonFilter, &filter)
		if err != nil {
			return nil, err
		}
		req.Filter = filter

	case types.FetchInstanceInfoMethod:
		filter := types.FetchInstanceInfoFilter{}
		err := json.Unmarshal(jsonFilter, &filter)
		if err != nil {
			return nil, err
		}
		req.Filter = filter

	case types.ListInstanceByPolicyMethod:
		filter := types.ListInstanceByPolicyFilter{}
		err := json.Unmarshal(jsonFilter, &filter)
		if err != nil {
			return nil, err
		}
		req.Filter = filter

	default:
		return nil, fmt.Errorf("method %s is not supported", req.Method)
	}

	return req, nil
}

// Validate Page.
func (p *Page) Validate() error {
	if p.Limit == 0 {
		return errf.New(errf.InvalidParameter, "limit should > 0")
	}

	if p.Limit > client.BkIAMMaxPageSize {
		return errf.New(errf.InvalidParameter, fmt.Sprintf("limit should <= %d", client.BkIAMMaxPageSize))
	}

	return nil
}

// SetData set response result to resp.
func (p *PullResourceResp) SetData(data interface{}) error {
	if p == nil {
		return errf.New(errf.InvalidParameter, "pull resource resp is nil")
	}

	p.Data = new(pbstruct.Struct)

	marshal, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err = p.Data.UnmarshalJSON(marshal); err != nil {
		return err
	}

	return nil
}

// IsDataStruct view.DataStructInterface 实现
func (p *PullResourceResp) IsDataStruct() bool {
	return true
}
