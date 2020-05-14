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

// TestConfigSet tests the config set cases.
func TestConfigSet(t *testing.T) {
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

	newAppid := ""
	{
		data, err := e2edata.CreateAppTestData(newBid)
		assert.Nil(err)

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

	data, err := e2edata.CreateConfigSetTestData(newBid, newAppid)
	assert.Nil(err)

	newCfgsetid := ""
	{
		t.Logf("Case: create new config set")
		resp, err := http.Post(testhost(configsetInterfaceV1), "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "cfgsetid").String(), body)
		newCfgsetid = gjson.Get(body, "cfgsetid").String()
	}

	{
		t.Logf("Case: create repeated config set")
		resp, err := http.Post(testhost(configsetInterfaceV1), "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_DM_ALREADY_EXISTS, gjson.Get(body, "errCode").Int(), body)
		assert.NotEqual("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(newCfgsetid, gjson.Get(body, "cfgsetid").String(), body)
	}

	name := gjson.Get(data, "name").String()
	{
		t.Logf("Case: query config set by name")
		resp, err := http.Get(testhost(configsetInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&appid=" + newAppid + "&name=" + name)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(newBid, gjson.Get(body, "configSet.bid").String(), body)
		assert.Equal(newAppid, gjson.Get(body, "configSet.appid").String(), body)
		assert.Equal(newCfgsetid, gjson.Get(body, "configSet.cfgsetid").String(), body)
		assert.Equal(name, gjson.Get(body, "configSet.name").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "configSet.creator").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "configSet.lastModifyBy").String(), body)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(body, "configSet.memo").String(), body)
	}

	{
		t.Logf("Case: query config set by cfgsetid")
		resp, err := http.Get(testhost(configsetInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&cfgsetid=" + newCfgsetid)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(newBid, gjson.Get(body, "configSet.bid").String(), body)
		assert.Equal(newAppid, gjson.Get(body, "configSet.appid").String(), body)
		assert.Equal(newCfgsetid, gjson.Get(body, "configSet.cfgsetid").String(), body)
		assert.Equal(name, gjson.Get(body, "configSet.name").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "configSet.creator").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "configSet.lastModifyBy").String(), body)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(body, "configSet.memo").String(), body)
	}

	{
		t.Logf("Case: query config set list")
		resp, err := http.Get(testhost(configsetListInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&appid=" + newAppid + "&index=0&limit=10")
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotZero(len(gjson.Get(body, "configSets").Array()), body)
	}

	{
		t.Logf("Case: update config set")
		data, err := e2edata.UpdateConfigSetTestData(newBid, newCfgsetid)
		assert.Nil(err)

		req, err := http.NewRequest("PUT", testhost(configsetInterfaceV1), strings.NewReader(data))
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

		r, err := http.Get(testhost(configsetInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&cfgsetid=" + newCfgsetid)
		assert.Nil(err)

		defer r.Body.Close()
		assert.Equal(http.StatusOK, r.StatusCode)

		b, err := respbody(r)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(b, "errCode").Int(), b)
		assert.Equal("OK", gjson.Get(b, "errMsg").String(), b)

		assert.Equal(newBid, gjson.Get(b, "configSet.bid").String(), b)
		assert.Equal(newAppid, gjson.Get(b, "configSet.appid").String(), b)
		assert.Equal(newCfgsetid, gjson.Get(b, "configSet.cfgsetid").String(), b)
		assert.Equal(gjson.Get(data, "name").String(), gjson.Get(b, "configSet.name").String(), b)
		assert.Equal(gjson.Get(data, "operator").String(), gjson.Get(b, "configSet.lastModifyBy").String(), b)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(b, "configSet.memo").String(), b)
		assert.Equal(gjson.Get(data, "state").Int(), gjson.Get(b, "configSet.state").Int(), b)
	}

	{
		t.Logf("Case: delete config set")
		req, err := http.NewRequest("DELETE", testhost(configsetInterfaceV1)+"?"+"seq=1&bid="+newBid+"&cfgsetid="+newCfgsetid+"&operator=e2e", nil)
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

		r, err := http.Get(testhost(configsetInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&cfgsetid=" + newCfgsetid)
		assert.Nil(err)

		defer r.Body.Close()
		assert.Equal(http.StatusOK, r.StatusCode)

		b, err := respbody(r)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_DM_NOT_FOUND, gjson.Get(b, "errCode").Int(), b)
		assert.NotEqual("OK", gjson.Get(b, "errMsg").String(), b)
	}

	alreadyLocker := ""
	{
		t.Logf("Case: lock config set")
		data, err := e2edata.LockConfigSetTestData(newBid, newCfgsetid)
		assert.Nil(err)

		resp, err := http.Post(testhost(configsetLockInterfaceV1), "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		alreadyLocker = gjson.Get(data, "operator").String()
	}

	{
		t.Logf("Case: try lock config set by other operator")
		data, err := e2edata.LockConfigSetTestData(newBid, newCfgsetid)
		assert.Nil(err)

		resp, err := http.Post(testhost(configsetLockInterfaceV1), "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_DM_CFGSET_LOCK_FAILED, gjson.Get(body, "errCode").Int(), body)
		assert.NotEqual("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(alreadyLocker, gjson.Get(body, "locker").String(), body)
		assert.NotEmpty(gjson.Get(body, "lockTime").String(), body)
	}

	{
		t.Logf("Case: unlock other config set")
		data, err := e2edata.UnlockConfigSetTestData(newBid, newCfgsetid, "")
		assert.Nil(err)

		req, err := http.NewRequest("PUT", testhost(configsetLockInterfaceV1), strings.NewReader(data))
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
	}

	{
		t.Logf("Case: unlock own config set")
		data, err := e2edata.UnlockConfigSetTestData(newBid, newCfgsetid, alreadyLocker)
		assert.Nil(err)

		req, err := http.NewRequest("PUT", testhost(configsetLockInterfaceV1), strings.NewReader(data))
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
	}
}
