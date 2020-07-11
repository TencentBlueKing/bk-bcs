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

package watch

import (
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

func TestGetWatchResource(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{})

	request := newReqWatch(req)
	defer request.exit()

	if _, err := request.get(); err != nil {
		t.Errorf("get() failed! err: %v", err)
	}
}

func TestPutWatchResource(t *testing.T) {
	bodyStr := "{\"data\":{\"a\":\"b\"}}"
	expect := map[string]interface{}{
		"a": "b",
	}
	r, _ := http.NewRequest("PUT", "/", ioutil.NopCloser(strings.NewReader(bodyStr)))
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{})

	request := newReqWatch(req)
	defer request.exit()

	if err := request.put(); err != nil {
		t.Errorf("put() failed! err: %v", err)
	}

	expect["updateTime"] = request.data.(map[string]interface{})["updateTime"]
	if !reflect.DeepEqual(request.data, expect) {
		t.Errorf("put() failed! \nput_data:\n%v:\nexpect:\n%v\n", request.data, expect)
	}
}

func TestDeleteWatchResource(t *testing.T) {
	r, _ := http.NewRequest("DELETE", "/", nil)
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{})

	request := newReqWatch(req)
	defer request.exit()

	if err := request.remove(); err != nil {
		t.Errorf("remove() failed! err: %v", err)
	}
}

func TestListWatchResource(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{})

	request := newReqWatch(req)
	defer request.exit()

	if _, err := request.list(); err != nil {
		t.Errorf("list() failed! err: %v", err)
	}
}
