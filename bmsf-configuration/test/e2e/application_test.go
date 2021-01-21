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

// TestApp tests the application cases.
func TestApp(t *testing.T) {
	assert := assert.New(t)

	data, err := e2edata.CreateAppTestData(e2eTestBizID)
	assert.Nil(err)

	newAppID := ""
	{
		t.Logf("Case: create new application")
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

	{
		t.Logf("Case: create repeated application")
		api := testHost(fmt.Sprintf(createAppAPIV2, e2eTestBizID))
		resp, err := httpRequest("POST", api, "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respBody(resp)
		assert.Nil(err)

		assert.EqualValues(false, gjson.Get(body, "result").Bool(), body)
		assert.EqualValues(pbcommon.ErrCode_E_CS_ALREADY_EXISTS, gjson.Get(body, "code").Int(), body)
		assert.NotEqual("OK", gjson.Get(body, "message").String(), body)

		assert.Equal(newAppID, gjson.Get(body, "data.app_id").String(), body)
	}

	name := gjson.Get(data, "name").String()
	{
		t.Logf("Case: query application by name")
		api := testHost(fmt.Sprintf(queryAppAPIV2, e2eTestBizID))
		resp, err := httpRequest("GET", api+"?"+"name="+name, "application/json", nil)
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
		assert.Equal(name, gjson.Get(body, "data.name").String(), body)
		assert.Equal(e2eTestOperator, gjson.Get(body, "data.creator").String(), body)
		assert.Equal(e2eTestOperator, gjson.Get(body, "data.last_modify_by").String(), body)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(body, "data.memo").String(), body)
		assert.Equal(gjson.Get(data, "deploy_type").Int(), gjson.Get(body, "data.deploy_type").Int(), body)
	}

	{
		t.Logf("Case: query application by app_id")
		api := testHost(fmt.Sprintf(queryAppAPIV2, e2eTestBizID))
		resp, err := httpRequest("GET", api+"?"+"app_id="+newAppID, "application/json", nil)
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
		assert.Equal(name, gjson.Get(body, "data.name").String(), body)
		assert.Equal(e2eTestOperator, gjson.Get(body, "data.creator").String(), body)
		assert.Equal(e2eTestOperator, gjson.Get(body, "data.last_modify_by").String(), body)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(body, "data.memo").String(), body)
		assert.Equal(gjson.Get(data, "deploy_type").Int(), gjson.Get(body, "data.deploy_type").Int(), body)
	}

	{
		t.Logf("Case: query application list")
		data, err := e2edata.QueryAppListTestData(e2eTestBizID, true, 0, 100)
		assert.Nil(err)

		api := testHost(fmt.Sprintf(listAppAPIV2, e2eTestBizID))
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
		t.Logf("Case: update application")
		data, err := e2edata.UpdateAppTestData(e2eTestBizID, newAppID)
		assert.Nil(err)

		api := testHost(fmt.Sprintf(updateAppAPIV2, e2eTestBizID, newAppID))
		resp, err := httpRequest("PUT", api, "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respBody(resp)
		assert.Nil(err)

		assert.EqualValues(true, gjson.Get(body, "result").Bool(), body)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "code").Int(), body)
		assert.Equal("OK", gjson.Get(body, "message").String(), body)

		api = testHost(fmt.Sprintf(queryAppAPIV2, e2eTestBizID))
		queryResp, err := httpRequest("GET", api+"?"+"app_id="+newAppID, "application/json", nil)
		assert.Nil(err)

		defer queryResp.Body.Close()
		assert.Equal(http.StatusOK, queryResp.StatusCode)

		queryBody, err := respBody(queryResp)
		assert.Nil(err)

		assert.EqualValues(true, gjson.Get(queryBody, "result").Bool(), queryBody)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(queryBody, "code").Int(), queryBody)
		assert.Equal("OK", gjson.Get(queryBody, "message").String(), queryBody)

		assert.Equal(e2eTestBizID, gjson.Get(queryBody, "data.biz_id").String(), queryBody)
		assert.Equal(newAppID, gjson.Get(queryBody, "data.app_id").String(), queryBody)
		assert.Equal(gjson.Get(data, "name").String(), gjson.Get(queryBody, "data.name").String(), queryBody)
		assert.Equal(gjson.Get(data, "deploy_type").Int(), gjson.Get(queryBody, "data.deploy_type").Int(), queryBody)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(queryBody, "data.memo").String(), queryBody)
		assert.Equal(gjson.Get(data, "state").Int(), gjson.Get(queryBody, "data.state").Int(), queryBody)
	}

	{
		t.Logf("Case: delete application")
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

		api = testHost(fmt.Sprintf(queryAppAPIV2, e2eTestBizID))
		queryResp, err := httpRequest("GET", api+"?"+"app_id="+newAppID, "application/json", nil)
		assert.Nil(err)

		defer queryResp.Body.Close()
		assert.Equal(http.StatusOK, queryResp.StatusCode)

		queryBody, err := respBody(queryResp)
		assert.Nil(err)

		assert.EqualValues(false, gjson.Get(queryBody, "result").Bool(), queryBody)
		assert.EqualValues(pbcommon.ErrCode_E_DM_NOT_FOUND, gjson.Get(queryBody, "code").Int(), queryBody)
		assert.NotEqual("OK", gjson.Get(queryBody, "message").String(), queryBody)
	}
}
