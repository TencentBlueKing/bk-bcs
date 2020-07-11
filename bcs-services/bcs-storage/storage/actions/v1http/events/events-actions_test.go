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

package events

import (
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

func TestPutEvent(t *testing.T) {
	bodyStr := "{" +
		"\"id\":\"12345\"," +
		"\"env\":\"env\"," +
		"\"kind\":\"kind\"," +
		"\"level\":\"level\"," +
		"\"component\":\"component\"," +
		"\"type\":\"type\"," +
		"\"describe\":\"describe\"," +
		"\"clusterId\":\"BCS-TEST-10001\"," +
		"\"eventTime\":1516849200," +
		"\"extraInfo\":{\"namespace\":\"ns\",\"name\":\"n\",\"kind\":\"kind\"}," +
		"\"data\":{\"a\":\"b\"}" +
		"}"
	expect := operator.M{
		"id":        "12345",
		"env":       types.EventEnv("env"),
		"kind":      types.EventKind("kind"),
		"level":     types.EventLevel("level"),
		"component": types.EventComponent("component"),
		"type":      "type",
		"describe":  "describe",
		"clusterId": "BCS-TEST-10001",
		"eventTime": time.Unix(1516849200, 0),
		"extraInfo": types.EventExtraInfo{
			Namespace: "ns",
			Name:      "n",
			Kind:      types.ExtraKind("kind"),
		},
		"data": map[string]interface{}{
			"a": "b",
		},
	}
	r, _ := http.NewRequest("PUT", "/events", ioutil.NopCloser(strings.NewReader(bodyStr)))
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{})

	request := newReqEvent(req)
	defer request.exit()
	if err := request.insert(); err != nil || !reflect.DeepEqual(request.data, expect.Update("createTime", request.data["createTime"])) {
		t.Errorf("%v", request.data["extraInfo"])
		t.Errorf("insert() failed! \npost_data:\n%v\nexpect:\n%v\nerr:\n%v\n", request.data, expect, err)
	}
}

func TestListEvent(t *testing.T) {
	expect := operator.MockCombineCondition(
		operator.BaseCondition.AddOp(operator.In, "kind", []string{"kind1", "kind2"}).
			AddOp(operator.In, "clusterId", []string{"BCS-TEST-10001"}).
			AddOp(operator.In, "extraInfo.name", []string{"n1", "n2"}))
	r, _ := http.NewRequest("GET", "/events?clusterId=BCS-TEST-10001&kind=kind1,kind2&extraInfo.name=n1,n2&offset=12&length=20", nil)
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{})

	request := newReqEvent(req)
	defer request.exit()
	if _, _, err := request.listEvent(); err != nil {
		t.Errorf("listEvent() failed! err: %v", err)
	}
	if result := operator.MockCombineCondition(request.condition); !reflect.DeepEqual(result, expect) {
		t.Errorf("listEvent() failed! \nlist_condition:\n%v\nexpect:\n%v\n", result, expect)
	}
	if request.offset != 12 || request.limit != 20 {
		t.Errorf("listEvent() failed! \nexpect_offset=12 expect_limit=20\nresult_offset=%d result_limit=%d", request.offset, request.limit)
	}
}
