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

package cr

import (
	"fmt"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/requester"
)

func Test_GetPerfDetail(t *testing.T) {
	testCli := New(&apis.ClientOptions{}, requester.NewRequester())
	req := &GetPerfDetailReq{
		Dsl: &GetPerfDetailDsl{MatchExpr: []GetPerfDetailMatchExpr{{
			Key:      "IP",
			Values:   []string{"11.187.113.19"},
			Operator: "In",
		}, {
			Key:      "sync_date",
			Values:   []string{"2024-01-15"},
			Operator: "In",
		}}},
		Offset: 0,
		Limit:  1,
	}

	rsp, err := testCli.GetPerfDetail(req)
	fmt.Println(err)
	fmt.Println(rsp)

}
