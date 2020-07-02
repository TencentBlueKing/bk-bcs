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

package hostConfig

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

func TestGetHost(t *testing.T) {
	expect := operator.M{
		"ip": "",
	}
	r, _ := http.NewRequest("GET", "/", nil)
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{Value: []interface{}{map[string]interface{}{"updateTime": time.Unix(1516849200, 0)}}})

	request := newReqHost(req)
	defer request.exit()

	rs, err := request.getHost()
	if err != nil {
		t.Errorf("getHost() failed! err: %v", err)
		return
	}

	if condition := operator.MockCombineCondition(request.condition); !reflect.DeepEqual(condition, expect) {
		t.Errorf("getHost() failed! \ncondition:\n%v\nexpect:\n%v\n", condition, expect)
	}

	if ti := rs[0].(map[string]interface{})["updateTime"]; ti != "2018-01-25 11:00:00" {
		t.Errorf("getHost() failed! \nupdateTime:\n%v\nexpect:\n2018-01-25 11:00:00\n", ti)
	}
}

func TestPutHost(t *testing.T) {
	strBody := "{" +
		"\"clusterId\":\"BCS-TEST-10001\"," +
		"\"ip\":\"\"," +
		"\"data\":{\"a\":\"b\"}" +
		"}"
	expect := operator.M{
		"clusterId": "BCS-TEST-10001",
		"ip":        "",
		"data": map[string]interface{}{
			"a": "b",
		},
	}
	r, _ := http.NewRequest("PUT", "/", ioutil.NopCloser(strings.NewReader(strBody)))
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{ChangeInfo: &operator.ChangeInfo{Matched: 1, Updated: 1}})

	request := newReqHost(req)
	defer request.exit()

	err := request.putHost()
	if err != nil {
		t.Errorf("putHost() failed! err: %v", err)
		return
	}

	expect["updateTime"] = request.data["updateTime"]
	expect["createTime"] = request.data["createTime"]
	if !reflect.DeepEqual(request.data, expect) {
		t.Errorf("putHost() failed! \nput_data:\n%v\nexpect:\n%v\n", request.data, expect)
	}
}

func TestDeleteHost(t *testing.T) {
	r, _ := http.NewRequest("DELETE", "/", nil)
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{ChangeInfo: &operator.ChangeInfo{Matched: 1, Removed: 1}})
	request := newReqHost(req)
	defer request.exit()

	err := request.removeHost()
	if err != nil {
		t.Errorf("removeHost() failed! err: %v", err)
		return
	}
}

func TestListHost(t *testing.T) {
	expect := operator.M{
		"clusterId": "",
	}
	r, _ := http.NewRequest("GET", "/", nil)
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{})

	request := newReqHost(req)
	defer request.exit()

	_, err := request.queryHost()
	if err != nil {
		t.Errorf("queryHost() failed! err: %v", err)
		return
	}

	if condition := operator.MockCombineCondition(request.condition); !reflect.DeepEqual(condition, expect) {
		t.Errorf("queryHost() failed! \ncondition:\n%v\nexpect:\n%v\n", condition, expect)
	}
}

func TestPostClusterRelation(t *testing.T) {
	bodyStr := "{\"ips\":[\"127.0.0.1\",\"127.0.0.2\"]}"
	expect := operator.M{
		"in": operator.M{"ip": []string{"127.0.0.1", "127.0.0.2"}},
	}
	r, _ := http.NewRequest("POST", "/", ioutil.NopCloser(strings.NewReader(bodyStr)))
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{ChangeInfo: &operator.ChangeInfo{Matched: 0, Updated: 0}})

	request := newReqHost(req)
	defer request.exit()

	if err := request.doRelation(false); err != nil {
		t.Errorf("doRelation() POST failed! err: %v", err)
		return
	}

	if condition := operator.MockCombineCondition(request.condition); !reflect.DeepEqual(condition, expect) {
		t.Errorf("doRelation() POST failed! \ncondition:\n%v\nexpect:\n%v\n", condition, expect)
	}
}

func TestPutClusterRelation(t *testing.T) {
	bodyStr := "{\"ips\":[\"127.0.0.1\",\"127.0.0.2\"]}"
	expect := operator.M{
		"in": operator.M{"ip": []string{"127.0.0.1", "127.0.0.2"}},
	}
	r, _ := http.NewRequest("PUT", "/", ioutil.NopCloser(strings.NewReader(bodyStr)))
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{ChangeInfo: &operator.ChangeInfo{Matched: 1, Updated: 1}})

	request := newReqHost(req)
	defer request.exit()

	if err := request.doRelation(true); err != nil {
		t.Errorf("doRelation() PUT failed! err: %v", err)
		return
	}

	if condition := operator.MockCombineCondition(request.condition); !reflect.DeepEqual(condition, expect) {
		t.Errorf("doRelation() POST failed! \ncondition:\n%v\nexpect:\n%v\n", condition, expect)
	}
}
