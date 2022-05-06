/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package perm

import (
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
)

// IAMPermError IAM 鉴权失败抛出的 Error 类型
type IAMPermError struct {
	Code          int
	Msg           string
	Username      string
	ActionReqList []ActionResourcesRequest
}

// Error ...
func (e *IAMPermError) Error() string {
	return strconv.Itoa(e.Code) + ": " + e.Msg
}

// Perms ...
func (e *IAMPermError) Perms() (map[string]interface{}, error) {
	applyURL, err := NewApplyURLGenerator().Gen(e.Username, e.ActionReqList)
	if err != nil {
		return nil, err
	}
	actionList := []map[string]interface{}{}
	for _, actionReq := range e.ActionReqList {
		actionList = append(actionList, map[string]interface{}{
			"resource_type": actionReq.ResType, "action_id": actionReq.ActionID,
		})
	}
	return map[string]interface{}{
		"perms": map[string]interface{}{
			"apply_url":   applyURL,
			"action_list": actionList,
		},
	}, nil
}

// NewIAMPermErr ...
func NewIAMPermErr(username, msg string, actionReqList []ActionResourcesRequest) error {
	return &IAMPermError{Code: errcode.NoIAMPerm, Username: username, Msg: msg, ActionReqList: actionReqList}
}
