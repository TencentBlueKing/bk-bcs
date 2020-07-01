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

package dynamic

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

func TestGetResources(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{})

	request := newReqDynamic(req)
	defer request.exit()

	if _, err := request.nsGet(); err != nil {
		t.Errorf("nsGet() failed! err: %v", err)
	}

	if _, err := request.csList(); err != nil {
		t.Errorf("csList() failed! err: %v", err)
	}
}

func TestPutResources(t *testing.T) {
	csExpect := operator.M{
		"data": map[string]interface{}{
			"foo": "bar",
		},
		"resourceType": "",
		"clusterId":    "",
		"resourceName": "",
	}
	r, _ := http.NewRequest("PUT", "/", ioutil.NopCloser(strings.NewReader("{\"data\":{\"foo\":\"bar\"}}")))
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{ChangeInfo: &operator.ChangeInfo{Matched: 1, Updated: 1}})

	request := newReqDynamic(req)
	defer request.exit()

	err := request.csPut()
	if csExpect["updateTime"], csExpect["createTime"] = request.data["updateTime"], request.data["createTime"]; err != nil || !reflect.DeepEqual(request.data, csExpect) {
		t.Errorf("csPut() failed! \ndata:\n%v\nexpect:\n%v\nerr:\n%v\n", request.data, csExpect, err)
	}

	nsExpect := csExpect
	nsExpect["namespace"] = ""
	nsExpect["data"] = map[string]interface{}{"hello": "world"}
	r, _ = http.NewRequest("PUT", "/", ioutil.NopCloser(strings.NewReader("{\"data\":{\"hello\":\"world\"}}")))
	req = restful.NewRequest(r)
	request.req = req
	request.reset()
	err = request.nsPut()
	if nsExpect["updateTime"], nsExpect["createTime"] = request.data["updateTime"], request.data["createTime"]; err != nil || !reflect.DeepEqual(request.data, nsExpect) {
		t.Errorf("nsPut() failed! \ndata:\n%v\nexpect:\n%v\nerr:\n%v\n", request.data, nsExpect, err)
	}
}

func TestDeleteResources(t *testing.T) {
	r, _ := http.NewRequest("DELETE", "/", nil)
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{ChangeInfo: &operator.ChangeInfo{Matched: 1, Removed: 1}})

	request := newReqDynamic(req)
	defer request.exit()

	if err := request.nsRemove(); err != nil {
		t.Errorf("nsRemove() failed! err: %v", err)
	}
	if err := request.csRemove(); err != nil {
		t.Errorf("csRemove() failed! err: %v", err)
	}
}

func TestListResources(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{ChangeInfo: &operator.ChangeInfo{Matched: 1}})

	request := newReqDynamic(req)
	defer request.exit()

	if _, err := request.nsList(); err != nil {
		t.Errorf("nsList() failed! err: %v", err)
	}
	if _, err := request.csList(); err != nil {
		t.Errorf("csList() failed! err: %v", err)
	}
}

func TestDeleteBatchResource(t *testing.T) {
	expect := operator.M{
		"and": []interface{}{
			operator.M{
				"clusterId":    "",
				"resourceType": "",
			},
			operator.M{"and": []interface{}{
				operator.M{"gt": operator.M{"updateTime": time.Unix(1516849200, 0)}},
				operator.M{"lt": operator.M{"updateTime": time.Unix(1516849201, 0)}},
			}},
		},
	}
	r, _ := http.NewRequest("DELETE", "/", ioutil.NopCloser(strings.NewReader("{\"updateTimeBegin\":1516849200,\"updateTimeEnd\":1516849201}")))
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{ChangeInfo: &operator.ChangeInfo{Matched: 1, Removed: 1}})

	request := newReqDynamic(req)
	defer request.exit()

	if err := request.csBatchRemove(); err != nil {
		t.Errorf("csBatchRemove() failed! err: %v", err)
	}
	if condition := operator.MockCombineCondition(request.condition); !reflect.DeepEqual(condition, expect) {
		t.Errorf("csBatchRemove() failed! \ncondition:\n%v\nexpect:\n%v\n", condition, expect)
	}

	expect["and"].([]interface{})[0].(operator.M)["namespace"] = ""
	r, _ = http.NewRequest("DELETE", "/", ioutil.NopCloser(strings.NewReader("{\"updateTimeBegin\":1516849200,\"updateTimeEnd\":1516849201}")))
	req = restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{ChangeInfo: &operator.ChangeInfo{Matched: 1, Removed: 1}})

	request = newReqDynamic(req)
	defer request.exit()

	if err := request.nsBatchRemove(); err != nil {
		t.Errorf("nsBatchRemove() failed! err: %v", err)
	}
	if condition := operator.MockCombineCondition(request.condition); !reflect.DeepEqual(condition, expect) {
		t.Errorf("nsBatchRemove() failed! \ncondition:\n%v\nexpect:\n%v\n", condition, expect)
	}
}
