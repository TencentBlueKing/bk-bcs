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

// TestBusiness tests the business cases.
func TestBusiness(t *testing.T) {
	assert := assert.New(t)

	data, err := e2edata.CreateBusinessTestData()
	assert.Nil(err)

	newBid := ""
	{
		t.Logf("Case: create new business")
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

	{
		t.Logf("Case: create repeated business")
		resp, err := http.Post(testhost(businessInterfaceV1), "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_DM_ALREADY_EXISTS, gjson.Get(body, "errCode").Int(), body)
		assert.NotEqual("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(newBid, gjson.Get(body, "bid").String(), body)
	}

	name := gjson.Get(data, "name").String()
	{
		t.Logf("Case: query business by name")
		resp, err := http.Get(testhost(businessInterfaceV1) + "?" + "seq=1&name=" + name)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(newBid, gjson.Get(body, "business.bid").String(), body)
		assert.Equal(name, gjson.Get(body, "business.name").String(), body)
		assert.Equal(gjson.Get(data, "depid").String(), gjson.Get(body, "business.depid").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "business.creator").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "business.lastModifyBy").String(), body)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(body, "business.memo").String(), body)
	}

	{
		t.Logf("Case: query business by bid")
		resp, err := http.Get(testhost(businessInterfaceV1) + "?" + "seq=1&bid=" + newBid)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(newBid, gjson.Get(body, "business.bid").String(), body)
		assert.Equal(name, gjson.Get(body, "business.name").String(), body)
		assert.Equal(gjson.Get(data, "depid").String(), gjson.Get(body, "business.depid").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "business.creator").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "business.lastModifyBy").String(), body)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(body, "business.memo").String(), body)
	}

	{
		t.Logf("Case: query business list")
		resp, err := http.Get(testhost(businessListInterfaceV1) + "?" + "seq=1&index=0&limit=10")
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotZero(len(gjson.Get(body, "businesses").Array()), body)
	}

	{
		t.Logf("Case: update business")
		data, err := e2edata.UpdateBusinessTestData(newBid)
		assert.Nil(err)

		req, err := http.NewRequest("PUT", testhost(businessInterfaceV1), strings.NewReader(data))
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

		r, err := http.Get(testhost(businessInterfaceV1) + "?" + "seq=1&bid=" + newBid)
		assert.Nil(err)

		defer r.Body.Close()
		assert.Equal(http.StatusOK, r.StatusCode)

		b, err := respbody(r)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(b, "errCode").Int(), b)
		assert.Equal("OK", gjson.Get(b, "errMsg").String(), b)
		assert.Equal(newBid, gjson.Get(b, "business.bid").String(), b)
		assert.Equal(gjson.Get(data, "name").String(), gjson.Get(b, "business.name").String(), b)
		assert.Equal(gjson.Get(data, "depid").String(), gjson.Get(b, "business.depid").String(), b)
		assert.Equal(gjson.Get(data, "operator").String(), gjson.Get(b, "business.lastModifyBy").String(), b)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(b, "business.memo").String(), b)
		assert.Equal(gjson.Get(data, "state").Int(), gjson.Get(b, "business.state").Int(), b)
	}
}
