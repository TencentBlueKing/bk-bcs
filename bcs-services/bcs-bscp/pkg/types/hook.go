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

package types

import (
	"errors"

	"bscp.io/pkg/dal/table"
)

// ListHooksOption defines options to list group.
type ListHooksOption struct {
	BizID  uint32    `json:"biz_id"`
	Name   string    `json:"name"`
	Tag    string    `json:"tag"`
	All    bool      `json:"all"`
	NotTag bool      `json:"not_tag"`
	Page   *BasePage `json:"page"`
}

// Validate the list group options
func (opt *ListHooksOption) Validate(po *PageOption) error {
	if opt.BizID <= 0 {
		return errors.New("invalid biz id, should >= 1")
	}

	if opt.Page == nil {
		return errors.New("page is null")
	}

	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	return nil
}

// ListHookDetails defines the response details of requested ListHooksOption.
type ListHookDetails struct {
	Count   uint32        `json:"count"`
	Details []*table.Hook `json:"details"`
}

// HookTagCount defines the response details of requested CountHookTag.
type HookTagCount struct {
	Tag    string `db:"tag" json:"tag"`
	Counts uint32 `db:"counts" json:"counts"`
}
