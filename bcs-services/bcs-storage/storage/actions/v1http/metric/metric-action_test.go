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

package metric

import (
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

func TestGetMetric(t *testing.T) {
	expect := operator.M{
		"clusterId": "",
		"namespace": "",
		"type":      "",
		"name":      "",
	}
	r, _ := http.NewRequest("GET", "/", nil)
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{})

	request := newReqMetric(req)
	defer request.exit()

	_, err := request.getMetric()
	if err != nil {
		t.Errorf("getMetric() failed! err: %v", err)
	}

	if condition := operator.MockCombineCondition(request.condition); !reflect.DeepEqual(condition, expect) {
		t.Errorf("getMetric() failed! \ncondition:\n%v:\nexpect:\n%v\n", condition, expect)
	}
}

func TestPutMetric(t *testing.T) {
	bodyStr := "{\"data\":{\"a\":\"b\"}}"
	expect := operator.M{
		"data":      map[string]interface{}{"a": "b"},
		"clusterId": "",
		"namespace": "",
		"type":      "",
		"name":      "",
	}
	r, _ := http.NewRequest("PUT", "/", ioutil.NopCloser(strings.NewReader(bodyStr)))
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{ChangeInfo: &operator.ChangeInfo{Matched: 1, Updated: 1}})

	request := newReqMetric(req)
	defer request.exit()

	if err := request.put(); err != nil {
		t.Errorf("put() failed! err: %v", err)
	}

	expect["updateTime"] = request.data["updateTime"]
	expect["createTime"] = request.data["createTime"]
	if !reflect.DeepEqual(request.data, expect) {
		t.Errorf("put() failed! \nput_data:\n%v:\nexpect:\n%v\n", request.data, expect)
	}
}

func TestDeleteMetric(t *testing.T) {
	r, _ := http.NewRequest("DELETE", "/", nil)
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{ChangeInfo: &operator.ChangeInfo{Matched: 1, Removed: 1}})

	request := newReqMetric(req)
	defer request.exit()

	err := request.remove()
	if err != nil {
		t.Errorf("remove() failed! err: %v", err)
	}
}

func TestQueryMetric(t *testing.T) {
	expect := operator.M{
		"and": []interface{}{
			operator.M{"clusterId": ""},
			operator.M{"in": operator.M{"namespace": []string{"ns"}}},
			operator.M{"in": operator.M{"type": []string{"ha", "ho"}}},
			operator.M{"in": operator.M{"name": []string{"n"}}},
		}}

	r, _ := http.NewRequest("GET", "/?namespace=ns&type=ha,ho&name=n", nil)
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{})

	request := newReqMetric(req)
	defer request.exit()

	_, err := request.queryMetric()
	if err != nil {
		t.Errorf("queryMetric() failed! err: %v", err)
	}
	if condition := operator.MockCombineCondition(request.condition); !reflect.DeepEqual(condition, expect) {
		t.Errorf("queryMetric() failed! \ncondition:\n%v:\nexpect:\n%v\n", condition, expect)
	}
}

func TestListMetricTables(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{})

	request := newReqMetric(req)
	defer request.exit()

	_, err := request.tables()
	if err != nil {
		t.Errorf("tables() failed! err: %v", err)
	}
}
