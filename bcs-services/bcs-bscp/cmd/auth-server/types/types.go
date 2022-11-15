/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package types NOTES
package types

import (
	"encoding/json"
	"strconv"
	"strings"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/client"
	"bscp.io/pkg/iam/sys"
	pbbase "bscp.io/pkg/protocol/core/base"
	"bscp.io/pkg/runtime/filter"

	pbstruct "github.com/golang/protobuf/ptypes/struct"
)

const (
	// SuccessCode blueking iam success resp code.
	SuccessCode = 0
	// UnauthorizedErrorCode iam token authorized failed error code.
	UnauthorizedErrorCode = 401

	// ListAttrMethod query the list of properties that a resource type can use to configure permissions.
	ListAttrMethod Method = "list_attr"
	// ListAttrValueMethod gets a list of values for an attribute of a resource type.
	ListAttrValueMethod Method = "list_attr_value"
	// ListInstanceMethod query instances based on filter criteria.
	ListInstanceMethod Method = "list_instance"
	// FetchInstanceInfoMethod obtain resource instance details in batch.
	FetchInstanceInfoMethod Method = "fetch_instance_info"
	// ListInstanceByPolicyMethod query resource instances based on policy expressions.
	ListInstanceByPolicyMethod Method = "list_instance_by_policy"
	// SearchInstanceMethod query instances based on filter criteria and search keywords.
	SearchInstanceMethod Method = "search_instance"

	// IDField instance id field name.
	IDField = "id"
	// NameField instance display name.
	NameField = "display_name"
	// ResTopology resource topology level. e.g: "/biz,1/set,1/module,1/"
	ResTopology = "_bk_iam_path_"
)

// Method pull resource method.
type Method string

// PullResourceReq blueking iam pull resource request.
type PullResourceReq struct {
	Type   client.TypeID `json:"type"`
	Method Method        `json:"method"`
	Filter interface{}   `json:"filter,omitempty"`
	Page   Page          `json:"page,omitempty"`
}

// Page blueking iam pull resource page.
type Page struct {
	Limit  uint `json:"limit"`
	Offset uint `json:"offset"`
}

// PbPage convert Page to pb page.
func (p *Page) PbPage() *pbbase.BasePage {
	page := &pbbase.BasePage{
		Start: uint32(p.Offset),
		Limit: uint32(p.Limit),
	}

	return page
}

// ListAttrValueFilter list attr value filter.
type ListAttrValueFilter struct {
	Attr    string `json:"attr"`
	Keyword string `json:"keyword,omitempty"`
	// id type is string, int or bool
	IDs []interface{} `json:"ids,omitempty"`
}

// ListInstanceFilter list instance filter.
type ListInstanceFilter struct {
	Parent  *ParentFilter `json:"parent,omitempty"`
	Keyword string        `json:"keyword,omitempty"`
}

// GetBizIDAndPbFilter get biz_id and pb filter from list instance filter.
func (f *ListInstanceFilter) GetBizIDAndPbFilter() (
	uint32, *pbstruct.Struct, error) {

	if f.Parent == nil {
		return 0, nil, errf.New(errf.InvalidParameter, "parent is required")
	}

	bizID := f.Parent.ID.BizID
	expr := &filter.Expression{
		Op:    filter.And,
		Rules: make([]filter.RuleFactory, 0),
	}

	if f.Parent.Type == sys.Business {
		// data service filter is not nil, need add is true filter.
		expr.Rules = append(expr.Rules, &filter.AtomRule{
			Field: "id",
			Op:    filter.GreaterThan.Factory(),
			Value: 0,
		})

	} else {
		field, err := getResourceNameField(f.Parent.Type)
		if err != nil {
			return 0, nil, err
		}

		expr.Rules = append(expr.Rules, &filter.AtomRule{
			Field: field,
			Op:    filter.Equal.Factory(),
			Value: f.Parent.ID.InstanceID,
		})
	}

	if len(f.Keyword) != 0 {
		expr.Rules = append(expr.Rules, &filter.AtomRule{
			Field: "name",
			Op:    filter.ContainsInsensitive.Factory(),
			Value: f.Keyword,
		})
	}

	pbFilter, err := expr.MarshalPB()
	if err != nil {
		return 0, nil, err
	}

	return bizID, pbFilter, nil
}

// getResourceNameField get the query instance field corresponding to the resource type.
func getResourceNameField(resType client.TypeID) (string, error) {
	switch resType {
	case sys.Application:
		return "app_id", nil

	default:
		return "", errf.New(errf.InvalidParameter, "resource type not support")
	}
}

// ParentFilter parent filter.
type ParentFilter struct {
	Type client.TypeID `json:"type"`
	ID   InstanceID    `json:"id"`
}

// ResourceTypeChainFilter resource type chain filter.
type ResourceTypeChainFilter struct {
	SystemID string        `json:"system_id"`
	ID       client.TypeID `json:"id"`
}

// FetchInstanceInfoFilter fetch instance info filter.
type FetchInstanceInfoFilter struct {
	IDs   []InstanceID `json:"ids"`
	Attrs []string     `json:"attrs,omitempty"`
}

// ListInstanceByPolicyFilter list instance by policy filter.
type ListInstanceByPolicyFilter struct {
	// Expression *operator.Policy `json:"expression"`
}

// AttrResource attr resource.
type AttrResource struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// ListAttrValueResult list attr value result.
type ListAttrValueResult struct {
	Count   int64               `json:"count"`
	Results []AttrValueResource `json:"results"`
}

// AttrValueResource attr value resource.
type AttrValueResource struct {
	// id type is string, int or bool
	ID          interface{} `json:"id"`
	DisplayName string      `json:"display_name"`
}

// ListInstanceResult list instance result.
type ListInstanceResult struct {
	Count   uint32             `json:"count"`
	Results []InstanceResource `json:"results"`
}

// InstanceResource instance resource.
type InstanceResource struct {
	ID          InstanceID `json:"id"`
	DisplayName string     `json:"display_name"`
}

// InstanceID is iam resource id, like '{biz_id}' or '{biz_id}-{instance_id}'.
type InstanceID struct {
	// BizID must not is 0.
	BizID uint32 `json:"biz_id"`
	// InstanceID may is 0, if the auth resource type is a business.
	InstanceID uint32 `json:"instance_id"`
}

// bizIDAssembleSymbol used assemble biz_id and resource id's symbol. list instance return id need.
const bizIDAssembleSymbol = "-"

// UnmarshalJSON unmarshal a json raw to this instance.
func (i *InstanceID) UnmarshalJSON(raw []byte) error {
	id := strings.Trim(strings.TrimSpace(string(raw)), `\"`)
	elements := strings.Split(id, bizIDAssembleSymbol)

	if len(elements) > 2 || len(elements) == 0 {
		return errf.New(errf.InvalidParameter, "instance id not right format, should be "+
			"'{biz_id}' or '{biz_id}-{instance_id}'")
	}

	if len(elements) == 1 {
		bizID, err := strconv.ParseUint(elements[0], 10, 64)
		if err != nil {
			return err
		}

		if bizID == 0 {
			return errf.New(errf.InvalidParameter, "biz_id should > 0")
		}
		i.BizID = uint32(bizID)

		return nil
	}

	bizID, err := strconv.ParseUint(elements[0], 10, 64)
	if err != nil {
		return err
	}

	if bizID == 0 {
		return errf.New(errf.InvalidParameter, "biz_id should > 0")
	}
	i.BizID = uint32(bizID)

	instID, err := strconv.ParseUint(elements[1], 10, 64)
	if err != nil {
		return err
	}

	if instID == 0 {
		return errf.New(errf.InvalidParameter, "instance id should > 0")
	}
	i.InstanceID = uint32(instID)

	return nil
}

// MarshalJSON marshal instance id to json string.
func (i InstanceID) MarshalJSON() ([]byte, error) {
	if i.BizID == 0 && i.InstanceID == 0 {
		return nil, errf.New(errf.InvalidParameter, "bizID or instance id is required")
	}

	if i.BizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "bizID is required")
	}

	var id string
	if i.InstanceID == 0 {
		id = strconv.FormatUint(uint64(i.BizID), 10)
	} else {
		id = strconv.FormatUint(uint64(i.BizID), 10) + bizIDAssembleSymbol + strconv.FormatUint(uint64(i.InstanceID),
			10)
	}

	return json.Marshal(id)
}
