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
	"errors"
	"fmt"

	"bscp.io/cmd/auth-server/types"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbas "bscp.io/pkg/protocol/auth-server"
)

// PullResource callback function for iam to pull auth resource.
func (i *IAM) PullResource(ctx context.Context, req *pbas.PullResourceReq) (*pbas.PullResourceResp, error) {
	kt := kit.FromGrpcContext(ctx)
	resp := new(pbas.PullResourceResp)

	// if auth is disabled, returns error if iam calls pull resource callback function
	// if i.disableAuth {
	// 	err := errf.New(errf.Aborted, "authorize function is disabled, can not pull auth resource.")
	// 	logs.Errorf("authorize function is disabled, can not pull auth resource, rid: %s", kt.Rid)
	// 	return nil, err
	// }

	query, err := req.PullResourceReq()
	if err != nil {
		logs.Errorf("pb pull resource request convert failed, err: %v, rid: %s", err, kt.Rid)
		return resp, nil
	}

	// get response data for each iam query method, if callback method is not set, returns empty data
	switch query.Method {
	case types.ListInstanceMethod, types.SearchInstanceMethod:
		filter, ok := query.Filter.(types.ListInstanceFilter)
		if !ok {
			logs.Errorf("filter %v is not the right type for list_instance method, rid: %s", filter, kt.Rid)
			return nil, errors.New("filter type not right")
		}

		instance, err := i.ListInstances(kt, query.Type, &filter, query.Page)
		if err != nil {
			logs.Errorf("list instance failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if err = resp.SetData(instance); err != nil {
			logs.Errorf("set data failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

	case types.FetchInstanceInfoMethod:
		filter, ok := query.Filter.(types.FetchInstanceInfoFilter)
		if !ok {
			logs.Errorf("filter %v is not the right type for fetch_instance_info method, rid: %s", filter, kt.Rid)
			return nil, errors.New("filter type not right")
		}

		info, err := i.FetchInstanceInfo(kt, query.Type, &filter)
		if err != nil {
			logs.Errorf("fetch instance info failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if err = resp.SetData(info); err != nil {
			logs.Errorf("set data failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

	case types.ListAttrMethod:
		// attribute authentication is not needed for the time being,
		// so the interface does not need to be implemented
		logs.Errorf("pull resource method list_attr not support, rid: %s", kt.Rid)
		return resp, errors.New("list_attr not support")

	case types.ListAttrValueMethod:
		// attribute authentication is not needed for the time being,
		// so the interface does not need to be implemented
		logs.Errorf("pull resource method list_attr_value not support, rid: %s", kt.Rid)
		return resp, errors.New("list_attr_value not support")

	case types.ListInstanceByPolicyMethod:
		// sdk authentication is used, and there is no need to support this interface.
		logs.Errorf("pull resource method list_instance_by_policy not support, rid: %s", kt.Rid)
		return resp, errors.New("list_instance_by_policy not support")

	default:
		logs.Errorf("pull resource method %s not support, rid: %s", query.Method, kt.Rid)
		return resp, fmt.Errorf("%s not support", query.Method)
	}

	return resp, nil
}
