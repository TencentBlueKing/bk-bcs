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

// TestStrategy tests the cluster cases.
func TestStrategy(t *testing.T) {
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

	data, err := e2edata.CreateStrategyTestData(newBid, newAppid)
	assert.Nil(err)

	newStrategyid := ""
	{
		t.Logf("Case: create new strategy")
		resp, err := http.Post(testhost(strategyInterfaceV1), "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "strategyid").String(), body)
		newStrategyid = gjson.Get(body, "strategyid").String()
	}

	{
		t.Logf("Case: create repeated strategy")
		resp, err := http.Post(testhost(strategyInterfaceV1), "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_BS_ALREADY_EXISTS, gjson.Get(body, "errCode").Int(), body)
		assert.NotEqual("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(newStrategyid, gjson.Get(body, "strategyid").String(), body)
	}

	name := gjson.Get(data, "name").String()
	{
		t.Logf("Case: query strategy by name")
		resp, err := http.Get(testhost(strategyInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&appid=" + newAppid + "&name=" + name)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(newAppid, gjson.Get(body, "strategy.appid").String(), body)
		assert.Equal(newStrategyid, gjson.Get(body, "strategy.strategyid").String(), body)
		assert.Equal(name, gjson.Get(body, "strategy.name").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "strategy.creator").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "strategy.lastModifyBy").String(), body)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(body, "strategy.memo").String(), body)
	}

	{
		t.Logf("Case: query strategy by strategyid")
		resp, err := http.Get(testhost(strategyInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&strategyid=" + newStrategyid)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(newAppid, gjson.Get(body, "strategy.appid").String(), body)
		assert.Equal(newStrategyid, gjson.Get(body, "strategy.strategyid").String(), body)
		assert.Equal(name, gjson.Get(body, "strategy.name").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "strategy.creator").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "strategy.lastModifyBy").String(), body)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(body, "strategy.memo").String(), body)
	}

	{
		t.Logf("Case: query strategy list")
		resp, err := http.Get(testhost(strategyListInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&appid=" + newAppid + "&index=0&limit=10")
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotZero(len(gjson.Get(body, "strategies").Array()), body)
	}

	{
		t.Logf("Case: delete strategy")
		req, err := http.NewRequest("DELETE", testhost(strategyInterfaceV1)+"?"+"seq=1&bid="+newBid+"&strategyid="+newStrategyid+"&operator=e2e", nil)
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

		r, err := http.Get(testhost(strategyInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&strategyid=" + newStrategyid)
		assert.Nil(err)

		defer r.Body.Close()
		assert.Equal(http.StatusOK, r.StatusCode)

		b, err := respbody(r)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_DM_NOT_FOUND, gjson.Get(b, "errCode").Int(), b)
		assert.NotEqual("OK", gjson.Get(b, "errMsg").String(), b)
	}
}
