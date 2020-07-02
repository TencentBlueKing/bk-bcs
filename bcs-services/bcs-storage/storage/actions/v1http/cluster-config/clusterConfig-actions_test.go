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

package clusterConfig

import (
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	restful "github.com/emicklei/go-restful"
)

func TestGetConfig(t *testing.T) {
	r, _ := http.NewRequest("GET", "/clusters/BCS-TEST-10001", nil)
	req := restful.NewRequest(r)

	mockTank := &operator.MockTank{Value: []interface{}{map[string]interface{}{
		"data": map[string]interface{}{
			"common": map[string]interface{}{"foo": "bar"},
			"conf":   map[string]interface{}{"one": map[string]interface{}{"hello": "world"}}}}}}
	getNewTank = operator.GetMockTankNewFunc(mockTank)

	request := newReqConfig(req)
	defer request.exit()

	if _, err := request.getSvcSet(); err != nil {
		t.Errorf("getSvcSet() failed! err: %v", err)
	}
	if _, err := request.getCls(); err != nil {
		t.Errorf("getCls() failed! err: %v", err)
	}

	mockTank.Value = []interface{}{map[string]interface{}{"data": "18.01.02"}}
	if _, err := request.getStableVersion(); err != nil {
		t.Errorf("getStableVersion() failed! err: %v", err)
	}
}

func TestPutClusterConfig(t *testing.T) {
	data := "{" +
		"\"zkIp\":[\"0.0.0.0\",\"127.0.0.1\"]," +
		"\"masterIp\":[\"127.0.0.2\",\"127.0.0.3\",\"127.0.0.4\"]," +
		"\"dnsIp\":[\"127.0.0.5\",\"127.0.0.6\"]," +
		"\"city\":\"shenzhen\"," +
		"\"jfrogUrl\":\"jrog.com\"" +
		"}"

	temp := map[string]interface{}{
		"data": "{" +
			"\"city\": \"${city}\"," +
			"\"jfrogUrl\": \"${jfrogUrl}\"," +
			"\"mesosZkRaw\": \"${mesosZkRaw}\"," +
			"\"mesosZkHost\": \"${mesosZkHost}\"," +
			"\"mesosZkHostSpace\": \"${mesosZkHostSpace}\"," +
			"\"mesosZkHostSemicolon\": \"${mesosZkHostSemicolon}\"," +
			"\"dnsUpStream\": \"${dnsUpStream}\"," +
			"\"clusterIdNumber\": \"${clusterIdNumber}\"," +
			"\"mesosMaster\": \"${mesosMaster}\"," +
			"\"mesosQuorum\": \"${mesosQuorum}\"" +
			"}",
	}
	expect := operator.M{
		"clusterId": "",
		"data": map[string]interface{}{
			"city":                 "shenzhen",
			"jfrogUrl":             "jrog.com",
			"mesosZkRaw":           "0.0.0.0,127.0.0.1",
			"mesosZkHost":          "0.0.0.0:2181,127.0.0.1:2181",
			"mesosZkHostSpace":     "0.0.0.0:2181 127.0.0.1:2181",
			"mesosZkHostSemicolon": "0.0.0.0:2181;127.0.0.1:2181",
			"dnsUpStream":          "127.0.0.5:53 127.0.0.6:53",
			"clusterIdNumber":      "",
			"mesosMaster":          "127.0.0.2,127.0.0.3,127.0.0.4",
			"mesosQuorum":          "2",
		},
	}

	r, _ := http.NewRequest("GET", "/clusters/BCS-TEST-10001", ioutil.NopCloser(strings.NewReader(data)))
	req := restful.NewRequest(r)

	mockTank := &operator.MockTank{Value: []interface{}{temp}, ChangeInfo: &operator.ChangeInfo{Matched: 1}}
	getNewTank = operator.GetMockTankNewFunc(mockTank)

	request := newReqConfig(req)
	defer request.exit()

	if err := request.putClsConfig(); err != nil {
		t.Errorf("putClsConfig() failed! err: %v", err)
	}

	expect["updateTime"] = request.data["updateTime"]
	expect["createTime"] = request.data["createTime"]
	if !reflect.DeepEqual(request.data, expect) {
		t.Errorf("putClsConfig() failed! \nresult:\n%v\nexpect:\n%v\n", request.data, expect)
	}
}

func TestPutStableVersion(t *testing.T) {
	r, _ := http.NewRequest("PUT", "/clusters/BCS-TEST-10001", ioutil.NopCloser(strings.NewReader("{\"version\":\"18.01.02\"}")))
	req := restful.NewRequest(r)

	getNewTank = operator.GetMockTankNewFunc(&operator.MockTank{})

	request := newReqConfig(req)
	defer request.exit()

	if err := request.putStableVersion(); err != nil {
		t.Errorf("putStableVersion() failed! err: %v", err)
	}

	if request.stableVerData != "18.01.02" {
		t.Errorf("putStableVersion() failed! \nput_version:\n%s\nexpect:\n18.01.02\n", request.stableVerData)
	}
}
