/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	pbcommon "bk-bscp/internal/protocol/common"
	e2edata "bk-bscp/test/e2e/testdata"
)

// TestApp tests the application cases.
func TestApp(t *testing.T) {
	assert := assert.New(t)

	newBid := ""
	{
		data, err := e2edata.CreateBusinessTestData()
		assert.Nil(err)

		resp, err := http.Post(testhost(businessInterfaceV1), "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "bid").String(), body)
		newBid = gjson.Get(body, "bid").String()
	}

	data, err := e2edata.CreateAppTestData(newBid)
	assert.Nil(err)

	newAppid := ""
	{
		t.Logf("Case: create new application")
		resp, err := http.Post(testhost(appInterfaceV1), "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "appid").String(), body)
		newAppid = gjson.Get(body, "appid").String()
	}

	{
		t.Logf("Case: create repeated application")
		resp, err := http.Post(testhost(appInterfaceV1), "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_DM_ALREADY_EXISTS, gjson.Get(body, "errCode").Int(), body)
		assert.NotEqual("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(newAppid, gjson.Get(body, "appid").String(), body)
	}

	name := gjson.Get(data, "name").String()
	{
		t.Logf("Case: query application by name")
		resp, err := http.Get(testhost(appInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&name=" + name)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(newBid, gjson.Get(body, "app.bid").String(), body)
		assert.Equal(newAppid, gjson.Get(body, "app.appid").String(), body)
		assert.Equal(name, gjson.Get(body, "app.name").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "app.creator").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "app.lastModifyBy").String(), body)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(body, "app.memo").String(), body)
		assert.Equal(gjson.Get(data, "deployType").Int(), gjson.Get(body, "app.deployType").Int(), body)
	}

	{
		t.Logf("Case: query application by appid")
		resp, err := http.Get(testhost(appInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&appid=" + newAppid)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(newBid, gjson.Get(body, "app.bid").String(), body)
		assert.Equal(newAppid, gjson.Get(body, "app.appid").String(), body)
		assert.Equal(name, gjson.Get(body, "app.name").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "app.creator").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "app.lastModifyBy").String(), body)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(body, "app.memo").String(), body)
		assert.Equal(gjson.Get(data, "deployType").Int(), gjson.Get(body, "app.deployType").Int(), body)
	}

	{
		t.Logf("Case: query application list")
		resp, err := http.Get(testhost(appListInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&index=0&limit=10")
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotZero(len(gjson.Get(body, "apps").Array()), body)
	}

	{
		t.Logf("Case: update application")
		data, err := e2edata.UpdateAppTestData(newBid, newAppid)
		assert.Nil(err)

		req, err := http.NewRequest("PUT", testhost(appInterfaceV1), strings.NewReader(data))
		assert.Nil(err)

		req.Header.Add("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)

		r, err := http.Get(testhost(appInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&appid=" + newAppid)
		assert.Nil(err)

		defer r.Body.Close()
		assert.Equal(http.StatusOK, r.StatusCode)

		b, err := respbody(r)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(b, "errCode").Int(), b)
		assert.Equal("OK", gjson.Get(b, "errMsg").String(), b)
		assert.Equal(newBid, gjson.Get(b, "app.bid").String(), b)
		assert.Equal(newAppid, gjson.Get(b, "app.appid").String(), b)
		assert.Equal(gjson.Get(data, "name").String(), gjson.Get(b, "app.name").String(), b)
		assert.Equal(gjson.Get(data, "deployType").Int(), gjson.Get(b, "app.deployType").Int(), b)
		assert.Equal(gjson.Get(data, "operator").String(), gjson.Get(b, "app.lastModifyBy").String(), b)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(b, "app.memo").String(), b)
		assert.Equal(gjson.Get(data, "state").Int(), gjson.Get(b, "app.state").Int(), b)
	}

	{
		t.Logf("Case: delete application")
		req, err := http.NewRequest("DELETE", testhost(appInterfaceV1)+"?"+"seq=1&bid="+newBid+"&appid="+newAppid+"&operator=e2e", nil)
		assert.Nil(err)

		req.Header.Add("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)

		r, err := http.Get(testhost(appInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&appid=" + newAppid)
		assert.Nil(err)

		defer r.Body.Close()
		assert.Equal(http.StatusOK, r.StatusCode)

		b, err := respbody(r)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_DM_NOT_FOUND, gjson.Get(b, "errCode").Int(), b)
		assert.NotEqual("OK", gjson.Get(b, "errMsg").String(), b)
	}
}
