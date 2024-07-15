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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/TencentBlueKing/iam-go-sdk/resource"
	"github.com/emicklei/go-restful"
)

const (
	// Code 字段信息 https://bk.tencent.com/docs/document/7.0/236/39677

	// NoAuthCode 无权限
	NoAuthCode = 401
	// NotFoundCode 未找到资源或查询资源的方式不存在
	NotFoundCode = 404
	// SystemErrCode 系统错误
	SystemErrCode = 500
	// RefuseRequestCode 请求被拒绝
	RefuseRequestCode = 422
	// InvalidKeyWord 关键字不合法
	InvalidKeyWord = 406
	// RatelimitCode 限流
	RatelimitCode = 429
)

var dispatcher = resource.NewDispatcher()

// ResourceDispatch dispatches the request to the corresponding resource provider
func ResourceDispatch(request *restful.Request, response *restful.Response) {
	handler := resource.NewDispatchHandler(dispatcher)
	handler(response.ResponseWriter, request.Request)
}

// ListResult is the result of list
type ListResult struct {
	Count   int           `json:"count"`
	Results []interface{} `json:"results"`
}

// Instance is the instance
type Instance struct {
	ID            string   `json:"id"`
	DisplayName   string   `json:"display_name"`
	BKIAMApprover []string `json:"_bk_iam_approver_,omitempty"`
}

// Filter is the iam provider request filter
type Filter struct {
	Attr      string         `json:"attr"`
	Keyword   string         `json:"keyword"`
	IDs       []string       `json:"ids"`
	Parent    ResourceParent `json:"parent"`
	Ancestors []Ancestor     `json:"ancestors"`
}

// ResourceParent is the iam provider request resource parent
type ResourceParent struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Ancestor is the iam provider request ancestor
type Ancestor struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func convertFilter(data map[string]interface{}) Filter {
	// Convert the map to a JSON string
	jsonData, err := json.Marshal(data)
	if err != nil {
		return Filter{}
	}

	var filter Filter
	// Unmarshal the JSON string into the struct
	err = json.Unmarshal(jsonData, &filter)
	if err != nil {
		return filter
	}
	return filter
}

func combineNameID(name, id string) string {
	return fmt.Sprintf("%s(%s)", name, id)
}

// SplitString 分割字符串, 允许半角逗号、分号及空格
func SplitString(str string) []string {
	str = strings.TrimSpace(str)
	str = strings.ReplaceAll(str, ";", ",")
	str = strings.ReplaceAll(str, " ", ",")
	if str == "" {
		return []string{}
	}
	return strings.Split(str, ",")
}
