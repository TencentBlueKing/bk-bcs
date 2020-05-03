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
	"encoding/base64"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	pbcommon "bk-bscp/internal/protocol/common"
	e2edata "bk-bscp/test/e2e/testdata"
)

// TestCommit tests the commit cases.
func TestCommit(t *testing.T) {
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
	}

	newClusterid := ""
	newClusterName := ""
	{
		data, err := e2edata.CreateClusterTestData(newBid, newAppid)
		assert.Nil(err)

		resp, err := http.Post(testhost(clusterInterfaceV1), "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "clusterid").String(), body)
		newClusterid = gjson.Get(body, "clusterid").String()
		newClusterName = gjson.Get(data, "name").String()
	}

	newZoneid := ""
	newZoneName := ""
	{
		data, err := e2edata.CreateZoneTestData(newBid, newAppid, newClusterid)
		assert.Nil(err)

		resp, err := http.Post(testhost(zoneInterfaceV1), "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "zoneid").String(), body)
		newZoneid = gjson.Get(body, "zoneid").String()
		newZoneName = gjson.Get(data, "name").String()
	}

	data, err := e2edata.CreateCommitTestData(newBid, newAppid, newCfgsetid)
	assert.Nil(err)

	newCommitid := ""
	{
		t.Logf("Case: create new commit without template")
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
		t.Logf("Case: query commit")
		resp, err := http.Get(testhost(commitInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&commitid=" + newCommitid)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(newBid, gjson.Get(body, "commit.bid").String(), body)
		assert.Equal(newAppid, gjson.Get(body, "commit.appid").String(), body)
		assert.Equal(newCfgsetid, gjson.Get(body, "commit.cfgsetid").String(), body)
		assert.Equal(newCommitid, gjson.Get(body, "commit.commitid").String(), body)
		assert.Equal(gjson.Get(data, "operator").String(), gjson.Get(body, "commit.operator").String(), body)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(body, "commit.memo").String(), body)
		assert.Equal(gjson.Get(data, "templateid").String(), gjson.Get(body, "commit.templateid").String(), body)
		assert.Equal(gjson.Get(data, "template").String(), gjson.Get(body, "commit.template").String(), body)
		assert.Equal(gjson.Get(data, "templateRule").String(), gjson.Get(body, "commit.templateRule").String(), body)
		assert.Empty(gjson.Get(body, "commit.releaseid").String(), body)
		assert.EqualValues(pbcommon.CommitState_CS_INIT, gjson.Get(body, "commit.state").Int(), body)
	}

	{
		t.Logf("Case: query history commits(all states)")
		operator := gjson.Get(data, "operator").String()
		resp, err := http.Get(testhost(commitHistoryInterfaceV1) + "?" +
			"seq=1&bid=" + newBid + "&appid=" + newAppid + "&cfgsetid=" + newCfgsetid + "&queryType=0&operator=" + operator + "&index=0&limit=10")
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotZero(len(gjson.Get(body, "commits").Array()), body)
	}

	{
		t.Logf("Case: update commit")
		data, err := e2edata.UpdateCommitTestData(newBid, newCommitid)
		assert.Nil(err)

		req, err := http.NewRequest("PUT", testhost(commitInterfaceV1), strings.NewReader(data))
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

		r, err := http.Get(testhost(commitInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&commitid=" + newCommitid)
		assert.Nil(err)

		defer r.Body.Close()
		assert.Equal(http.StatusOK, r.StatusCode)

		b, err := respbody(r)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(b, "errCode").Int(), b)
		assert.Equal("OK", gjson.Get(b, "errMsg").String(), b)
		assert.Equal(newBid, gjson.Get(b, "commit.bid").String(), b)
		assert.Equal(newAppid, gjson.Get(b, "commit.appid").String(), b)
		assert.Equal(newCfgsetid, gjson.Get(b, "commit.cfgsetid").String(), b)
		assert.Equal(newCommitid, gjson.Get(b, "commit.commitid").String(), b)
		assert.Equal(gjson.Get(data, "operator").String(), gjson.Get(b, "commit.operator").String(), b)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(b, "commit.memo").String(), b)
		assert.Equal(gjson.Get(data, "changes").String(), gjson.Get(b, "commit.changes").String(), b)
		assert.Equal(gjson.Get(data, "configs").String(), gjson.Get(b, "commit.configs").String(), b)
	}

	{
		t.Logf("Case: confirm commit without template")
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

		r, err := http.Get(testhost(commitInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&commitid=" + newCommitid)
		assert.Nil(err)

		defer r.Body.Close()
		assert.Equal(http.StatusOK, r.StatusCode)

		b, err := respbody(r)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(b, "errCode").Int(), b)
		assert.Equal("OK", gjson.Get(b, "errMsg").String(), b)
		assert.EqualValues(pbcommon.CommitState_CS_CONFIRMED, gjson.Get(b, "commit.state").Int(), b)
		configsBase64 := gjson.Get(b, "commit.configs").String()

		{
			t.Logf("Case: configs content without template")
			r, err := http.Get(testhost(configsInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&appid=" +
				newAppid + "&cfgsetid=" + newCfgsetid + "&commitid=" + newCommitid)
			assert.Nil(err)

			defer r.Body.Close()
			assert.Equal(http.StatusOK, r.StatusCode)

			b, err := respbody(r)
			assert.Nil(err)

			assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(b, "errCode").Int(), b)
			assert.Equal("OK", gjson.Get(b, "errMsg").String(), b)
			assert.Equal(newBid, gjson.Get(b, "configs.bid").String(), b)
			assert.Equal(newAppid, gjson.Get(b, "configs.appid").String(), b)
			assert.Equal(newCfgsetid, gjson.Get(b, "configs.cfgsetid").String(), b)
			assert.Equal(newCommitid, gjson.Get(b, "configs.commitid").String(), b)
			assert.Equal(configsBase64, gjson.Get(b, "configs.content").String(), b)
		}
	}

	{
		t.Logf("Case: cancel commit after confirm")
		data, err := e2edata.CancelCommitTestData(newBid, newCommitid)
		assert.Nil(err)

		req, err := http.NewRequest("PUT", testhost(commitCancelInterfaceV1), strings.NewReader(data))
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

	data, err = e2edata.CreateCommitWithTplTestData(newBid, newAppid, newCfgsetid, newClusterName, newZoneName)
	assert.Nil(err)

	{
		t.Logf("Case: create new commit with template")
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
		t.Logf("Case: confirm commit with template")
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

		r, err := http.Get(testhost(commitInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&commitid=" + newCommitid)
		assert.Nil(err)

		defer r.Body.Close()
		assert.Equal(http.StatusOK, r.StatusCode)

		b, err := respbody(r)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(b, "errCode").Int(), b)
		assert.Equal("OK", gjson.Get(b, "errMsg").String(), b)
		assert.EqualValues(pbcommon.CommitState_CS_CONFIRMED, gjson.Get(b, "commit.state").Int(), b)

		{
			t.Logf("Case: template rendering(cluster)")
			r, err := http.Get(testhost(configsInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&appid=" +
				newAppid + "&cfgsetid=" + newCfgsetid + "&commitid=" + newCommitid + "&clusterid=" + newClusterid)
			assert.Nil(err)

			defer r.Body.Close()
			assert.Equal(http.StatusOK, r.StatusCode)

			b, err := respbody(r)
			assert.Nil(err)

			assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(b, "errCode").Int(), b)
			assert.Equal("OK", gjson.Get(b, "errMsg").String(), b)
			assert.Equal(newBid, gjson.Get(b, "configs.bid").String(), b)
			assert.Equal(newAppid, gjson.Get(b, "configs.appid").String(), b)
			assert.Equal(newCfgsetid, gjson.Get(b, "configs.cfgsetid").String(), b)
			assert.Equal(newCommitid, gjson.Get(b, "configs.commitid").String(), b)
			assert.Equal(newClusterid, gjson.Get(b, "configs.clusterid").String(), b)

			content, err := base64.StdEncoding.DecodeString(gjson.Get(b, "configs.content").String())
			assert.Nil(err)
			assert.EqualValues("v1", content, b)
		}
		{
			t.Logf("Case: template rendering(zone)")
			r, err := http.Get(testhost(configsInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&appid=" +
				newAppid + "&cfgsetid=" + newCfgsetid + "&commitid=" + newCommitid + "&clusterid=" + newClusterid + "&zoneid=" + newZoneid)
			assert.Nil(err)

			defer r.Body.Close()
			assert.Equal(http.StatusOK, r.StatusCode)

			b, err := respbody(r)
			assert.Nil(err)

			assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(b, "errCode").Int(), b)
			assert.Equal("OK", gjson.Get(b, "errMsg").String(), b)
			assert.Equal(newBid, gjson.Get(b, "configs.bid").String(), b)
			assert.Equal(newAppid, gjson.Get(b, "configs.appid").String(), b)
			assert.Equal(newCfgsetid, gjson.Get(b, "configs.cfgsetid").String(), b)
			assert.Equal(newCommitid, gjson.Get(b, "configs.commitid").String(), b)
			assert.Equal(newClusterid, gjson.Get(b, "configs.clusterid").String(), b)
			assert.Equal(newZoneid, gjson.Get(b, "configs.zoneid").String(), b)

			content, err := base64.StdEncoding.DecodeString(gjson.Get(b, "configs.content").String())
			assert.Nil(err)
			assert.EqualValues("v2", content, b)
		}
		{
			t.Logf("Case: query configs list after template rendering")
			r, err := http.Get(testhost(configsListInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&appid=" +
				newAppid + "&cfgsetid=" + newCfgsetid + "&commitid=" + newCommitid + "&index=0&limit=10")
			assert.Nil(err)

			defer r.Body.Close()
			assert.Equal(http.StatusOK, r.StatusCode)

			b, err := respbody(r)
			assert.Nil(err)

			assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(b, "errCode").Int(), b)
			assert.Equal("OK", gjson.Get(b, "errMsg").String(), b)
			assert.NotZerof(len(gjson.Get(b, "cfgslist").Array()), b)
		}
	}

	data, err = e2edata.CreateCommitTestData(newBid, newAppid, newCfgsetid)
	assert.Nil(err)

	{
		t.Logf("Case: create new commit without template")
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
		t.Logf("Case: cancel commit after create")
		data, err := e2edata.CancelCommitTestData(newBid, newCommitid)
		assert.Nil(err)

		req, err := http.NewRequest("PUT", testhost(commitCancelInterfaceV1), strings.NewReader(data))
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
		t.Logf("Case: update commit after cancel")
		data, err := e2edata.UpdateCommitTestData(newBid, newCommitid)
		assert.Nil(err)

		req, err := http.NewRequest("PUT", testhost(commitInterfaceV1), strings.NewReader(data))
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
