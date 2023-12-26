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

package dao

import (
	"fmt"
	"testing"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
)

func TestParseChangedSpecFields(t *testing.T) {
	pre := &table.App{
		ID:    1,
		BizID: 2,
		Spec: &table.AppSpec{
			Name:       "api",
			ConfigType: table.File,
			Memo:       "this is a memo",
		},
		Revision: nil,
	}

	cur := &table.App{
		ID:    1,
		BizID: 2,
		Spec: &table.AppSpec{
			Name:       "server",
			ConfigType: "",
			Memo:       "this is a changed memo!",
		},
		Revision: nil,
	}

	changed, err := parseChangedSpecFields(pre, cur)
	if err != nil {
		t.Errorf("test parse changed spec fields failed, err: %v", err)
		return
	}

	nameV, exist := changed["name"]
	if !exist {
		t.Error("test parse changed spec fields failed, name should be changed.")
		return
	}

	if fmt.Sprintf("%v", nameV) != "server" {
		t.Error("test parse changed spec fields failed, name's value is mistake.")
		return
	}

	memoV, exist := changed["memo"]
	if !exist {
		t.Error("test parse changed spec fields failed, memo should be changed.")
		return
	}

	if fmt.Sprintf("%v", memoV) != "this is a changed memo!" {
		t.Error("test parse changed spec fields failed, memo's value is mistake.")
		return
	}

}
