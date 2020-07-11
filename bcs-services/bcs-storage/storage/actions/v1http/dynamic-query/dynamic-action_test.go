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

package dynamicquery

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

type mockFilter struct {
	BaseInField string `json:"baseInField" filter:"base.in.field"`
	IntField    string `json:"intField" filter:"int.field,int"`
	Int64Field  string `json:"int64Field" filter:"int64.field,int64"`
	TimeLField  string `json:"timeLField" filter:"time.field,timeL"`
	TimeRField  string `json:"timeRField" filter:"time.field,timeR"`
	BoolField   string `json:"boolField" filter:"bool.field,bool"`
}

func (m mockFilter) getCondition() *operator.Condition {
	return qGenerate(m, timestampsLayout)
}

func TestDoQuery(t *testing.T) {
	expect :=
		operator.M{"and": []interface{}{
			operator.M{"and": []interface{}{
				operator.M{"and": []interface{}{
					operator.M{"and": []interface{}{
						operator.M{"and": []interface{}{
							operator.M{"in": operator.M{"base.in.field": []string{"a", "b", "c"}}},
							operator.M{"int.field": 1},
						}},
						operator.M{"int64.field": int64(1234567890987654321)},
					}},
					operator.M{"gt": operator.M{"time.field": int64(1516849200)}},
				}},
				operator.M{"lt": operator.M{"time.field": int64(1516849201)}},
			}},
			operator.M{"bool.field": false},
		}}
	r, _ := http.NewRequest("GET", "/?baseInField=a,b,c&intField=1&int64Field=1234567890987654321&timeLField=1516849200&timeRField=1516849201&boolField=false", nil)
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{})

	request := newReqDynamic(req, &mockFilter{}, "hi")
	defer request.exit()

	if _, err := request.queryDynamic(); err != nil {
		t.Errorf("queryDynamic() failed! err: %v", err)
	}

	condition := operator.MockCombineCondition(request.condition)

	if !reflect.DeepEqual(condition, expect) {
		t.Errorf("queryDynamic() failed! \nquery_condition:\n%v\nexpect:\n%v\n", condition, expect)
	}
}
