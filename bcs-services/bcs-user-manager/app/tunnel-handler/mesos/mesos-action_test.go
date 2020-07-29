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

package mesos

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	"github.com/emicklei/go-restful"
)

func TestRequest2mesosapi(t *testing.T) {
	req, err := http.NewRequest("GET", "/mesosdriver/v4/{sub_path:.*}", bytes.NewReader(nil))
	if err != nil {
		t.Fatal(err)
	}
	request := restful.NewRequest(req)
	rr := httptest.NewRecorder()
	response := restful.NewResponse(rr)
	handlerGetActions(request, response)
	body, readErr := ioutil.ReadAll(rr.Body)
	if readErr != nil {
		t.Fatal("error read response body")
	}
	respData := bhttp.APIRespone{}
	err = json.Unmarshal(body, &respData)
	if err != nil {
		t.Fatal("error when unmarshall response body")
	}
	if respData.Code != common.BcsErrCommHttpParametersFailed {
		t.Error("header BCS-ClusterID empty should return a parameters failed code")
	}
}
