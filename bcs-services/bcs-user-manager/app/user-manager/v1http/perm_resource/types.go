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
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/iam"
)

const (
	// NoLimit no limit definition
	NoLimit = 999999999
	// MaxPageSize max limit of a page
	MaxPageSize = 1000
)

const (
	// SuccessCode success code
	SuccessCode    = 0
	// SuccessMessage success message
	SuccessMessage = "success"

	// IDField id field
	IDField   = "id"
	// NameField display_name
	NameField = "display_name"
)

// Response the response body
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	// maybe a [] or {}
	Data interface{} `json:"data"`
}

// Method PullResource type
type Method string

const (
	// ListAttrMethod list_ttr
	ListAttrMethod             Method = "list_attr"
	// ListAttrValueMethod list_attr_value
	ListAttrValueMethod        Method = "list_attr_value"
	// ListInstanceMethod list_instance
	ListInstanceMethod         Method = "list_instance"
	// FetchInstanceInfoMethod fetch_instance_info
	FetchInstanceInfoMethod    Method = "fetch_instance_info"
	// ListInstanceByPolicyMethod list_instance_by_policy
	ListInstanceByPolicyMethod Method = "list_instance_by_policy"
	// SearchInstanceMethod search_instance
	SearchInstanceMethod       Method = "search_instance"
)

// PullResourceReq request resource
type PullResourceReq struct {
	Type   iam.TypeID  `json:"type"`
	Method Method      `json:"method"`
	Filter interface{} `json:"filter,omitempty"`
	Page   Page        `json:"page,omitempty"`
}

// UnmarshalJSON trans interface{} to filter type
func (req *PullResourceReq) UnmarshalJSON(raw []byte) error {
	data := struct {
		Type   iam.TypeID      `json:"type"`
		Method Method          `json:"method"`
		Filter json.RawMessage `json:"filter,omitempty"`
		Page   Page            `json:"page,omitempty"`
	}{}
	err := json.Unmarshal(raw, &data)
	if err != nil {
		return err
	}
	req.Type = data.Type
	req.Method = data.Method
	req.Page = data.Page
	if data.Filter == nil || len(data.Filter) == 0 {
		return nil
	}

	switch data.Method {
	case ListAttrValueMethod:
		filter := ListAttrValueFilter{}
		err := json.Unmarshal(data.Filter, &filter)
		if err != nil {
			return err
		}
		req.Filter = filter
	case SearchInstanceMethod:
		filter := SearchInstanceFilter{}
		err := json.Unmarshal(data.Filter, &filter)
		if err != nil {
			return err
		}
		req.Filter = filter
	case ListInstanceMethod:
		filter := ListInstanceFilter{}
		err := json.Unmarshal(data.Filter, &filter)
		if err != nil {
			return err
		}
		req.Filter = filter
	case FetchInstanceInfoMethod:
		filter := FetchInstanceInfoFilter{}
		err := json.Unmarshal(data.Filter, &filter)
		if err != nil {
			return err
		}
		req.Filter = filter
	case ListInstanceByPolicyMethod:
		filter := ListInstanceByPolicyFilter{}
		err := json.Unmarshal(data.Filter, &filter)
		if err != nil {
			return err
		}
		req.Filter = filter
	default:
		return fmt.Errorf("method %s is not supported", data.Method)
	}
	return nil
}

// Page xxx
type Page struct {
	Offset int64 `json:"offset"`
	Limit  int64 `json:"limit"`
}

// IsIllegal check page valid
func (page *Page) IsIllegal() bool {
	if page.Offset < 0 {
		page.Offset = 0
	}

	if (page.Limit > MaxPageSize && page.Limit != NoLimit) || page.Limit <= 0{
		return true
	}
	return false
}

// ListAttrValueFilter filter by keyword or ids
type ListAttrValueFilter struct {
	Attr    string `json:"attr"`
	Keyword string `json:"keyword,omitempty"`
	// id type is string, int or bool
	IDs []interface{} `json:"ids,omitempty"`
}

// ListInstanceFilter filter by parent or keyword
type ListInstanceFilter struct {
	Parent *ParentFilter `json:"parent,omitempty"`
}

// ParentFilter parent level
type ParentFilter struct {
	Type iam.TypeID `json:"type"`
	ID   string     `json:"id"`
}

// SearchInstanceFilter filter by parent or keyword
type SearchInstanceFilter struct {
	Parent  *ParentFilter `json:"parent,omitempty"`
	Keyword string        `json:"keyword,omitempty"`
}

// ResourceTypeChainFilter resource chain
type ResourceTypeChainFilter struct {
	SystemID string     `json:"system_id"`
	ID       iam.TypeID `json:"id"`
}

// FetchInstanceInfoFilter find instance IDs attr
type FetchInstanceInfoFilter struct {
	IDs   []string `json:"ids"`
	Attrs []string `json:"attrs,omitempty"`
}

// ListInstanceByPolicyFilter filter bu policy
type ListInstanceByPolicyFilter struct {
	Expression interface{} `json:"expression"`
}

// AttrResource attr resource
type AttrResource struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// ListAttrValueResult get attr resource value
type ListAttrValueResult struct {
	Count   int64               `json:"count"`
	Results []AttrValueResource `json:"results"`
}

// AttrValueResource attr values
type AttrValueResource struct {
	// id type is string, int or bool
	ID          interface{} `json:"id"`
	DisplayName string      `json:"display_name"`
}

// ListInstanceResult list instance for resource
type ListInstanceResult struct {
	Count   int64              `json:"count"`
	Results []InstanceResource `json:"results"`
}

// InstanceResource instance resource
type InstanceResource struct {
	ID          string         `json:"id"`
	DisplayName string         `json:"display_name"`
	Path        []InstancePath `json:"path,omitempty"`
}

// InstancePath instance path
type InstancePath struct {
	Type        iam.TypeID `json:"type"`
	ID          string     `json:"id"`
	DisplayName string     `json:"display_name"`
}
