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

// TestRelease tests the release cases.
func TestRelease(t *testing.T) {
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

	newCfgsetName := ""
	newCfgsetid := ""
	{
		data, err := e2edata.CreateConfigSetTestData(newBid, newAppid)
		assert.Nil(err)

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
		newCfgsetName = gjson.Get(data, "name").String()
	}

	newCommitid := ""
	{
		data, err := e2edata.CreateCommitTestData(newBid, newAppid, newCfgsetid)
		assert.Nil(err)

		resp, err := http.Post(testhost(commitInterfaceV1), "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "commitid").String(), body)
		newCommitid = gjson.Get(body, "commitid").String()
	}

	{
		t.Logf("Case: create new release with unconfirmed commit")
		data, err := e2edata.CreateReleaseTestData(newBid, newCommitid, "")
		assert.Nil(err)

		resp, err := http.Post(testhost(releaseInterfaceV1), "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_BS_CREATE_RELEASE_WITH_UNCONFIRMED_COMMIT, gjson.Get(body, "errCode").Int(), body)
		assert.NotEqual("OK", gjson.Get(body, "errMsg").String(), body)
	}

	{
		data, err := e2edata.ConfirmCommitTestData(newBid, newCommitid)
		assert.Nil(err)

		req, err := http.NewRequest("PUT", testhost(commitConfirmInterfaceV1), strings.NewReader(data))
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

	newStrategies := ""
	newStrategyid := ""
	{
		data, err := e2edata.CreateStrategyTestData(newBid, newAppid)
		assert.Nil(err)

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
		newStrategies = gjson.Get(body, "strategy.content").String()
	}

	data, err := e2edata.CreateReleaseTestData(newBid, newCommitid, newStrategyid)
	assert.Nil(err)

	newReleaseid := ""
	{
		t.Logf("Case: create new release with confirmed commit")
		resp, err := http.Post(testhost(releaseInterfaceV1), "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "releaseid").String(), body)
		newReleaseid = gjson.Get(body, "releaseid").String()
	}

	{
		t.Logf("Case: query release")
		resp, err := http.Get(testhost(releaseInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&releaseid=" + newReleaseid)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)

		assert.Equal(newBid, gjson.Get(body, "release.bid").String(), body)
		assert.Equal(newAppid, gjson.Get(body, "release.appid").String(), body)
		assert.Equal(newCfgsetid, gjson.Get(body, "release.cfgsetid").String(), body)
		assert.Equal(newCommitid, gjson.Get(body, "release.commitid").String(), body)
		assert.Equal(newReleaseid, gjson.Get(body, "release.releaseid").String(), body)
		assert.Equal(newStrategyid, gjson.Get(body, "release.strategyid").String(), body)
		assert.Equal(newCfgsetName, gjson.Get(body, "release.cfgsetName").String(), body)
		assert.Equal(gjson.Get(data, "name").String(), gjson.Get(body, "release.name").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "release.creator").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "release.lastModifyBy").String(), body)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(body, "release.memo").String(), body)
		assert.Equal(newStrategies, gjson.Get(body, "release.strategies").String(), body)
		assert.EqualValues(pbcommon.ReleaseState_RS_INIT, gjson.Get(body, "release.state").Int(), body)
	}

	{
		t.Logf("Case: query history releases(all states)")
		operator := gjson.Get(data, "creator").String()
		resp, err := http.Get(testhost(releaseHistoryInterfaceV1) + "?" +
			"seq=1&bid=" + newBid + "&cfgsetid=" + newCfgsetid + "&queryType=0&operator=" + operator + "&index=0&limit=10")
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotZero(len(gjson.Get(body, "releases").Array()), body)
	}

	{
		t.Logf("Case: update release")
		data, err := e2edata.UpdateReleaseTestData(newBid, newReleaseid)
		assert.Nil(err)

		req, err := http.NewRequest("PUT", testhost(releaseInterfaceV1), strings.NewReader(data))
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

		r, err := http.Get(testhost(releaseInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&releaseid=" + newReleaseid)
		assert.Nil(err)

		defer r.Body.Close()
		assert.Equal(http.StatusOK, r.StatusCode)

		b, err := respbody(r)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(b, "errCode").Int(), b)
		assert.Equal("OK", gjson.Get(b, "errMsg").String(), b)
		assert.Equal(newAppid, gjson.Get(b, "release.appid").String(), b)
		assert.Equal(newCfgsetid, gjson.Get(b, "release.cfgsetid").String(), b)
		assert.Equal(newCommitid, gjson.Get(b, "release.commitid").String(), b)
		assert.Equal(newReleaseid, gjson.Get(b, "release.releaseid").String(), b)
		assert.Equal(gjson.Get(data, "name").String(), gjson.Get(b, "release.name").String(), b)
		assert.Equal(gjson.Get(data, "operator").String(), gjson.Get(b, "release.lastModifyBy").String(), b)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(b, "release.memo").String(), b)
	}

	{
		t.Logf("Case: publish release")
		data, err := e2edata.PublishReleaseTestData(newBid, newReleaseid)
		assert.Nil(err)

		resp, err := http.Post(testhost(releasePubInterfaceV1), "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	{
		t.Logf("Case: update release after publish")
		data, err := e2edata.UpdateReleaseTestData(newBid, newReleaseid)
		assert.Nil(err)

		req, err := http.NewRequest("PUT", testhost(releaseInterfaceV1), strings.NewReader(data))
		assert.Nil(err)

		req.Header.Add("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, gjson.Get(body, "errCode").Int(), body)
		assert.NotEqual("OK", gjson.Get(body, "errMsg").String(), body)
	}

	{
		t.Logf("Case: cancel release after publish")
		data, err := e2edata.CancelReleaseTestData(newBid, newReleaseid)
		assert.Nil(err)

		req, err := http.NewRequest("PUT", testhost(releaseCancelInterfaceV1), strings.NewReader(data))
		assert.Nil(err)

		req.Header.Add("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, gjson.Get(body, "errCode").Int(), body)
		assert.NotEqual("OK", gjson.Get(body, "errMsg").String(), body)
	}

	{
		data, err := e2edata.CreateReleaseTestData(newBid, newCommitid, newStrategyid)
		assert.Nil(err)

		resp, err := http.Post(testhost(releaseInterfaceV1), "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "releaseid").String(), body)
		newReleaseid = gjson.Get(body, "releaseid").String()
	}

	{
		t.Logf("Case: cancel release")
		data, err := e2edata.CancelReleaseTestData(newBid, newReleaseid)
		assert.Nil(err)

		req, err := http.NewRequest("PUT", testhost(releaseCancelInterfaceV1), strings.NewReader(data))
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
		t.Logf("Case: update release after cancel")
		data, err := e2edata.UpdateReleaseTestData(newBid, newReleaseid)
		assert.Nil(err)

		req, err := http.NewRequest("PUT", testhost(releaseInterfaceV1), strings.NewReader(data))
		assert.Nil(err)

		req.Header.Add("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, gjson.Get(body, "errCode").Int(), body)
		assert.NotEqual("OK", gjson.Get(body, "errMsg").String(), body)
	}
}
