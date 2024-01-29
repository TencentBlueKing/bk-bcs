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

// Package types NOTES
package types

import (
	"encoding/json"

	pbstruct "github.com/golang/protobuf/ptypes/struct"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/client"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/sys"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
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

// PullResourceResp blueking iam pull resource response.
type PullResourceResp struct {
	Code    int32       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ConvertToPb ...
func (r *PullResourceResp) ConvertToPb() (*pbstruct.Struct, error) {

	data := new(pbstruct.Struct)

	marshal, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	if err = data.UnmarshalJSON(marshal); err != nil {
		return nil, err
	}
	return data, nil
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

// getResourceNameField get the query instance field corresponding to the resource type.
// nolint: unused
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
	ID   string        `json:"id"`
}

// ResourceTypeChainFilter resource type chain filter.
type ResourceTypeChainFilter struct {
	SystemID string        `json:"system_id"`
	ID       client.TypeID `json:"id"`
}

// FetchInstanceInfoFilter fetch instance info filter.
type FetchInstanceInfoFilter struct {
	IDs   []string `json:"ids"`
	Attrs []string `json:"attrs,omitempty"`
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
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// InstanceInfo instance info.
type InstanceInfo struct {
	ID            string   `json:"id"`
	DisplayName   string   `json:"display_name"`
	BKIAMApprover []string `json:"_bk_iam_approver_"`
	BKIAMPath     []string `json:"_bk_iam_path_"`
}
