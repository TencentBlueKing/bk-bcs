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
	"fmt"
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

	newAppID := ""
	{
		data, err := e2edata.CreateAppTestData(e2eTestBizID)
		assert.Nil(err)

		api := testHost(fmt.Sprintf(createAppAPIV2, e2eTestBizID))
		resp, err := httpRequest("POST", api, "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respBody(resp)
		assert.Nil(err)

		assert.EqualValues(true, gjson.Get(body, "result").Bool(), body)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "code").Int(), body)
		assert.Equal("OK", gjson.Get(body, "message").String(), body)

		assert.NotEmpty(gjson.Get(body, "data.app_id").String(), body)
		newAppID = gjson.Get(body, "data.app_id").String()
	}

	newCfgID := ""
	newCfgName := ""
	newCfgFpath := ""
	{
		data, err := e2edata.CreateConfigTestData(e2eTestBizID, newAppID, "/etc")
		assert.Nil(err)

		api := testHost(fmt.Sprintf(createConfigAPIV2, e2eTestBizID, newAppID))
		resp, err := httpRequest("POST", api, "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respBody(resp)
		assert.Nil(err)

		assert.EqualValues(true, gjson.Get(body, "result").Bool(), body)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "code").Int(), body)
		assert.Equal("OK", gjson.Get(body, "message").String(), body)
		assert.NotEmpty(gjson.Get(body, "data.cfg_id").String(), body)
		newCfgID = gjson.Get(body, "data.cfg_id").String()

		api = testHost(fmt.Sprintf(queryConfigAPIV2, e2eTestBizID, newAppID))
		queryResp, err := httpRequest("GET", api+"?"+"cfg_id="+newCfgID, "application/json", nil)
		assert.Nil(err)

		defer queryResp.Body.Close()
		assert.Equal(http.StatusOK, queryResp.StatusCode)

		queryBody, err := respBody(queryResp)
		assert.Nil(err)

		assert.EqualValues(true, gjson.Get(queryBody, "result").Bool(), queryBody)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(queryBody, "code").Int(), queryBody)
		assert.Equal("OK", gjson.Get(queryBody, "message").String(), queryBody)
		newCfgName = gjson.Get(queryBody, "data.name").String()
		newCfgFpath = gjson.Get(queryBody, "data.fpath").String()
	}

	newCommitID := ""
	{
		data, err := e2edata.CreateCommitTestData(e2eTestBizID, newAppID, newCfgID, pbcommon.CommitMode_CM_CONFIGS)
		assert.Nil(err)

		api := testHost(fmt.Sprintf(createCommitAPIV2, e2eTestBizID, newAppID))
		resp, err := httpRequest("POST", api, "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respBody(resp)
		assert.Nil(err)

		assert.EqualValues(true, gjson.Get(body, "result").Bool(), body)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "code").Int(), body)
		assert.Equal("OK", gjson.Get(body, "message").String(), body)
		assert.NotEmpty(gjson.Get(body, "data.commit_id").String(), body)
		newCommitID = gjson.Get(body, "data.commit_id").String()
	}

	{
		t.Logf("Case: create new release with unconfirmed commit")
		data, err := e2edata.CreateReleaseTestData(e2eTestBizID, newAppID, newCommitID, "")
		assert.Nil(err)

		api := testHost(fmt.Sprintf(createReleaseAPIV2, e2eTestBizID, newAppID))
		resp, err := httpRequest("POST", api, "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respBody(resp)
		assert.Nil(err)

		assert.EqualValues(false, gjson.Get(body, "result").Bool(), body)
		assert.EqualValues(pbcommon.ErrCode_E_CS_CREATE_RELEASE_WITH_UNCONFIRMED_COMMIT,
			gjson.Get(body, "code").Int(), body)
		assert.NotEqual("OK", gjson.Get(body, "message").String(), body)
	}

	{
		data, err := e2edata.ConfirmCommitTestData(e2eTestBizID, newAppID, newCommitID)
		assert.Nil(err)

		api := testHost(fmt.Sprintf(confirmCommitAPIV2, e2eTestBizID, newAppID, newCommitID))
		resp, err := httpRequest("POST", api, "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respBody(resp)
		assert.Nil(err)

		assert.EqualValues(true, gjson.Get(body, "result").Bool(), body)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "code").Int(), body)
		assert.Equal("OK", gjson.Get(body, "message").String(), body)
	}

	data, err := e2edata.CreateReleaseTestData(e2eTestBizID, newAppID, newCommitID, "")
	assert.Nil(err)

	newReleaseID := ""
	{
		t.Logf("Case: create new release with confirmed commit")
		api := testHost(fmt.Sprintf(createReleaseAPIV2, e2eTestBizID, newAppID))
		resp, err := httpRequest("POST", api, "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respBody(resp)
		assert.Nil(err)

		assert.EqualValues(true, gjson.Get(body, "result").Bool(), body)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "code").Int(), body)
		assert.Equal("OK", gjson.Get(body, "message").String(), body)
		assert.NotEmpty(gjson.Get(body, "data.release_id").String(), body)
		newReleaseID = gjson.Get(body, "data.release_id").String()
	}

	{
		t.Logf("Case: query release")
		api := testHost(fmt.Sprintf(queryReleaseAPIV2, e2eTestBizID, newAppID, newReleaseID))
		resp, err := httpRequest("GET", api+"?"+"biz_id="+e2eTestBizID+"&app_id="+newAppID+
			"&release_id="+newReleaseID, "application/json", nil)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respBody(resp)
		assert.Nil(err)

		assert.EqualValues(true, gjson.Get(body, "result").Bool(), body)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "code").Int(), body)
		assert.Equal("OK", gjson.Get(body, "message").String(), body)

		assert.Equal(e2eTestBizID, gjson.Get(body, "data.biz_id").String(), body)
		assert.Equal(newAppID, gjson.Get(body, "data.app_id").String(), body)
		assert.Equal(newCfgID, gjson.Get(body, "data.cfg_id").String(), body)
		assert.Equal(newCommitID, gjson.Get(body, "data.commit_id").String(), body)
		assert.Equal(newReleaseID, gjson.Get(body, "data.release_id").String(), body)
		assert.Empty(gjson.Get(body, "data.strategy_id").String(), body)
		assert.Equal(newCfgName, gjson.Get(body, "data.cfg_name").String(), body)
		assert.Equal(newCfgFpath, gjson.Get(body, "data.cfg_fpath").String(), body)
		assert.Equal(gjson.Get(data, "name").String(), gjson.Get(body, "data.name").String(), body)
		assert.Equal(e2eTestOperator, gjson.Get(body, "data.creator").String(), body)
		assert.Equal(e2eTestOperator, gjson.Get(body, "data.last_modify_by").String(), body)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(body, "data.memo").String(), body)
		assert.Equal("{}", gjson.Get(body, "data.strategies").String(), body)
		assert.EqualValues(pbcommon.ReleaseState_RS_INIT, gjson.Get(body, "data.state").Int(), body)
	}

	{
		t.Logf("Case: query history releases")
		data, err := e2edata.QueryHistoryReleasesTestData(e2eTestBizID, newAppID, newCfgID, true, 0, 100)
		assert.Nil(err)

		api := testHost(fmt.Sprintf(listReleaseAPIV2, e2eTestBizID, newAppID))
		resp, err := httpRequest("POST", api, "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respBody(resp)
		assert.Nil(err)

		assert.EqualValues(true, gjson.Get(body, "result").Bool(), body)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "code").Int(), body)
		assert.Equal("OK", gjson.Get(body, "message").String(), body)

		assert.NotZero(gjson.Get(body, "data.total_count").Int(), body)
		assert.NotZero(len(gjson.Get(body, "data.info").Array()), body)
	}

	{
		t.Logf("Case: update release")
		data, err := e2edata.UpdateReleaseTestData(e2eTestBizID, newAppID, newReleaseID)
		assert.Nil(err)

		api := testHost(fmt.Sprintf(updateReleaseAPIV2, e2eTestBizID, newAppID, newReleaseID))
		resp, err := httpRequest("PUT", api, "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respBody(resp)
		assert.Nil(err)

		assert.EqualValues(true, gjson.Get(body, "result").Bool(), body)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "code").Int(), body)
		assert.Equal("OK", gjson.Get(body, "message").String(), body)

		api = testHost(fmt.Sprintf(queryReleaseAPIV2, e2eTestBizID, newAppID, newReleaseID))
		queryResp, err := httpRequest("GET", api+"?"+"biz_id="+e2eTestBizID+"&app_id="+newAppID+
			"&release_id="+newReleaseID, "application/json", nil)
		assert.Nil(err)

		defer queryResp.Body.Close()
		assert.Equal(http.StatusOK, queryResp.StatusCode)

		queryBody, err := respBody(queryResp)
		assert.Nil(err)

		assert.EqualValues(true, gjson.Get(queryBody, "result").Bool(), queryBody)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(queryBody, "code").Int(), queryBody)
		assert.Equal("OK", gjson.Get(queryBody, "message").String(), queryBody)

		assert.Equal(newAppID, gjson.Get(queryBody, "data.app_id").String(), queryBody)
		assert.Equal(newCfgID, gjson.Get(queryBody, "data.cfg_id").String(), queryBody)
		assert.Equal(newCommitID, gjson.Get(queryBody, "data.commit_id").String(), queryBody)
		assert.Equal(newReleaseID, gjson.Get(queryBody, "data.release_id").String(), queryBody)
		assert.Equal(gjson.Get(data, "name").String(), gjson.Get(queryBody, "data.name").String(), queryBody)
		assert.Equal(e2eTestOperator, gjson.Get(queryBody, "data.last_modify_by").String(), queryBody)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(queryBody, "data.memo").String(), queryBody)
	}

	{
		t.Logf("Case: publish release")
		data, err := e2edata.PublishReleaseTestData(e2eTestBizID, newAppID, newReleaseID)
		assert.Nil(err)

		api := testHost(fmt.Sprintf(publishReleaseAPIV2, e2eTestBizID, newAppID, newReleaseID))
		resp, err := httpRequest("POST", api, "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respBody(resp)
		assert.Nil(err)

		assert.EqualValues(true, gjson.Get(body, "result").Bool(), body)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "code").Int(), body)
		assert.Equal("OK", gjson.Get(body, "message").String(), body)
	}

	{
		t.Logf("Case: update release after publish")
		data, err := e2edata.UpdateReleaseTestData(e2eTestBizID, newAppID, newReleaseID)
		assert.Nil(err)

		api := testHost(fmt.Sprintf(updateReleaseAPIV2, e2eTestBizID, newAppID, newReleaseID))
		resp, err := httpRequest("PUT", api, "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respBody(resp)
		assert.Nil(err)

		assert.EqualValues(false, gjson.Get(body, "result").Bool(), body)
		assert.EqualValues(pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, gjson.Get(body, "code").Int(), body)
		assert.NotEqual("OK", gjson.Get(body, "message").String(), body)
	}

	{
		t.Logf("Case: cancel release after publish")
		data, err := e2edata.CancelReleaseTestData(e2eTestBizID, newAppID, newReleaseID)
		assert.Nil(err)

		api := testHost(fmt.Sprintf(cancelReleaseAPIV2, e2eTestBizID, newAppID, newReleaseID))
		resp, err := httpRequest("POST", api, "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respBody(resp)
		assert.Nil(err)

		assert.EqualValues(false, gjson.Get(body, "result").Bool(), body)
		assert.EqualValues(pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, gjson.Get(body, "code").Int(), body)
		assert.NotEqual("OK", gjson.Get(body, "message").String(), body)
	}

	{
		api := testHost(fmt.Sprintf(deleteConfigAPIV2, e2eTestBizID, newAppID, newCfgID))
		resp, err := httpRequest("DELETE", api, "application/json", nil)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respBody(resp)
		assert.Nil(err)

		assert.EqualValues(true, gjson.Get(body, "result").Bool(), body)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "code").Int(), body)
		assert.Equal("OK", gjson.Get(body, "message").String(), body)
	}

	{
		api := testHost(fmt.Sprintf(deleteAppAPIV2, e2eTestBizID, newAppID))
		resp, err := httpRequest("DELETE", api, "application/json", nil)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respBody(resp)
		assert.Nil(err)

		assert.EqualValues(true, gjson.Get(body, "result").Bool(), body)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "code").Int(), body)
		assert.Equal("OK", gjson.Get(body, "message").String(), body)
	}
}
