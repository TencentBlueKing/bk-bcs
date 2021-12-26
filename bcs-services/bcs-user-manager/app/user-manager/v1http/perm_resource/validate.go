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

package perm_resource

import (
	"errors"
	"fmt"
)

var (
	// ErrPullResourceReq err pull resource
	ErrPullResourceReq = errors.New("pull resource request not init")
)

// ValidateListAttrValueRequest request list_attr_value
func ValidateListAttrValueRequest(req *PullResourceReq) (*ListAttrValueFilter, error) {
	if req == nil {
		return nil, ErrPullResourceReq
	}

	filter, ok := req.Filter.(ListAttrValueFilter)
	if !ok {
		errMsg := fmt.Errorf("request filter %s is not right type for list_attr_value method", filter)
		return nil, errMsg
	}

	if filter.Attr == "" {
		errMsg := fmt.Errorf("request filter %s attr is null for list_attr_value method", req.Filter)
		return nil, errMsg
	}

	if req.Page.IsIllegal() {
		errMsg := fmt.Errorf("request page limit %d exceeds max page size", req.Page.Limit)
		return nil, errMsg
	}

	return &filter, nil
}

// ValidateListInstanceRequest request list_instance
func ValidateListInstanceRequest(req *PullResourceReq) (*ListInstanceFilter, error) {
	if req == nil {
		return nil, ErrPullResourceReq
	}

	if req.Page.IsIllegal() {
		errMsg := fmt.Errorf("request page limit %d exceeds max page size", req.Page.Limit)
		return nil, errMsg
	}
	if req.Filter == nil {
		return nil, nil
	}
	filter, ok := req.Filter.(ListInstanceFilter)
	if !ok {
		errMsg := fmt.Errorf("request filter %v is not the right type for list_instance method", filter)
		return nil, errMsg
	}
	return &filter, nil
}

// ValidateSearchInstanceRequest request search_instance
func ValidateSearchInstanceRequest(req *PullResourceReq) (*SearchInstanceFilter, error) {
	if req == nil {
		return nil, ErrPullResourceReq
	}

	if req.Page.IsIllegal() {
		errMsg := fmt.Errorf("request page limit %d exceeds max page size", req.Page.Limit)
		return nil, errMsg
	}
	if req.Filter == nil {
		return nil, nil
	}

	filter, ok := req.Filter.(SearchInstanceFilter)
	if !ok {
		errMsg := fmt.Errorf("request filter %s is not the right type for search_instance method", filter)
		return nil, errMsg
	}
	return &filter, nil
}

// ValidateFetchInstanceRequest request fetch_instance
func ValidateFetchInstanceRequest(req *PullResourceReq) (*FetchInstanceInfoFilter, error) {
	if req == nil {
		return nil, ErrPullResourceReq
	}

	if req.Filter == nil {
		errMsg := fmt.Errorf("request FetchInstanceInfoFilter filter cann't be null")
		return nil, errMsg
	}

	filter, ok := req.Filter.(FetchInstanceInfoFilter)
	if !ok {
		errMsg := fmt.Errorf("request filter %s is not the right type for fetch_instance method", filter)
		return nil, errMsg
	}

	if len(filter.IDs) == 0 {
		errMsg := fmt.Errorf("FetchInstanceInfoFilter must be input instance ids")
		return nil, errMsg
	}

	return &filter, nil
}

// ValidateListInstanceByPolicyRequest request list_instance_by_policy
func ValidateListInstanceByPolicyRequest(req *PullResourceReq)  (*ListInstanceByPolicyFilter, error) {
	if req == nil {
		return nil, ErrPullResourceReq
	}

	filter, ok := req.Filter.(ListInstanceByPolicyFilter)
	if !ok {
		errMsg := fmt.Errorf("request filter %s is not the right type for list_instance_by_policy method", filter)
		return nil, errMsg
	}

	if req.Page.IsIllegal() {
		errMsg := fmt.Errorf("request page limit %d exceeds max page size", req.Page.Limit)
		return nil, errMsg
	}

	return &filter, nil
}

