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

package alarms

import (
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

func TestPostAlarm(t *testing.T) {
	bodyStr := "{" +
		"\"clusterId\":\"BCS-TEST-10001\"," +
		"\"namespace\":\"ns\"," +
		"\"message\":\"ms\"," +
		"\"source\":\"src\"," +
		"\"module\":\"md\"," +
		"\"type\":\"tp\"," +
		"\"receivedTime\":1516849200" +
		"}"
	expect := operator.M{
		"clusterId":    "BCS-TEST-10001",
		"namespace":    "ns",
		"message":      "ms",
		"source":       "src",
		"module":       "md",
		"type":         "tp",
		"receivedTime": time.Unix(1516849200, 0),
		"data":         nil,
	}
	r, _ := http.NewRequest("POST", "/alarms", ioutil.NopCloser(strings.NewReader(bodyStr)))
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{})

	request := newReqAlarm(req)
	defer request.exit()
	if err := request.insert(); err != nil || !reflect.DeepEqual(request.data, expect.Update("createTime", request.data["createTime"])) {
		t.Errorf("insert() failed! \npost_data:\n%v\nexpect:\n%v\nerr:\n%v\n", request.data, expect, err)
	}
}

func TestListAlarm(t *testing.T) {
	expect := operator.MockCombineCondition(
		operator.BaseCondition.AddOp(operator.In, "clusterId", []string{"BCS-TEST-10001"}).And(
			operator.BaseCondition.AddOp(operator.In, "namespace", []string{"ns1", "ns2"})).And(
			operator.BaseCondition.AddOp(operator.Con, "type", "tp1").Or(
				operator.BaseCondition.AddOp(operator.Con, "type", "tp2"))))
	r, _ := http.NewRequest("GET", "/alarms?clusterId=BCS-TEST-10001&namespace=ns1,ns2&type=tp1,tp2&offset=12&length=20", nil)
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{})

	request := newReqAlarm(req)
	defer request.exit()
	if _, _, err := request.listAlarm(); err != nil {
		t.Errorf("listAlarm() failed! err: %v", err)
	}
	if result := operator.MockCombineCondition(request.condition); !reflect.DeepEqual(result, expect) {
		t.Errorf("listAlarm() failed! \nlist_condition:\n%v\nexpect:\n%v\n", result, expect)
	}
	if request.offset != 12 || request.limit != 20 {
		t.Errorf("listAlarm() failed! \nexpect_offset=12 expect_limit=20\nresult_offset=%d result_limit=%d", request.offset, request.limit)
	}
}
