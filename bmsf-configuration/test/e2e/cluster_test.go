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

// TestCluster tests the cluster cases.
func TestCluster(t *testing.T) {
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

	data, err := e2edata.CreateClusterTestData(newBid, newAppid)
	assert.Nil(err)

	newClusterid := ""
	{
		t.Logf("Case: create new cluster")
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
	}

	{
		t.Logf("Case: create repeated cluster")
		resp, err := http.Post(testhost(clusterInterfaceV1), "application/json", strings.NewReader(data))
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_DM_ALREADY_EXISTS, gjson.Get(body, "errCode").Int(), body)
		assert.NotEqual("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(newClusterid, gjson.Get(body, "clusterid").String(), body)
	}

	name := gjson.Get(data, "name").String()
	{
		t.Logf("Case: query cluster by name")
		resp, err := http.Get(testhost(clusterInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&appid=" + newAppid + "&name=" + name)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(newBid, gjson.Get(body, "cluster.bid").String(), body)
		assert.Equal(newAppid, gjson.Get(body, "cluster.appid").String(), body)
		assert.Equal(newClusterid, gjson.Get(body, "cluster.clusterid").String(), body)
		assert.Equal(name, gjson.Get(body, "cluster.name").String(), body)
		assert.Equal(gjson.Get(data, "rClusterid").String(), gjson.Get(body, "cluster.rClusterid").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "cluster.creator").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "cluster.lastModifyBy").String(), body)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(body, "cluster.memo").String(), body)
	}

	{
		t.Logf("Case: query cluster by clusterid")
		resp, err := http.Get(testhost(clusterInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&clusterid=" + newClusterid)
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(newBid, gjson.Get(body, "cluster.bid").String(), body)
		assert.Equal(newAppid, gjson.Get(body, "cluster.appid").String(), body)
		assert.Equal(newClusterid, gjson.Get(body, "cluster.clusterid").String(), body)
		assert.Equal(name, gjson.Get(body, "cluster.name").String(), body)
		assert.Equal(gjson.Get(data, "rClusterid").String(), gjson.Get(body, "cluster.rClusterid").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "cluster.creator").String(), body)
		assert.Equal(gjson.Get(data, "creator").String(), gjson.Get(body, "cluster.lastModifyBy").String(), body)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(body, "cluster.memo").String(), body)
	}

	{
		t.Logf("Case: query cluster list")
		resp, err := http.Get(testhost(clusterListInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&appid=" + newAppid + "&index=0&limit=10")
		assert.Nil(err)

		defer resp.Body.Close()
		assert.Equal(http.StatusOK, resp.StatusCode)

		body, err := respbody(resp)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotZero(len(gjson.Get(body, "clusters").Array()), body)
	}

	{
		t.Logf("Case: update cluster")
		data, err := e2edata.UpdateClusterTestData(newBid, newClusterid)
		assert.Nil(err)

		req, err := http.NewRequest("PUT", testhost(clusterInterfaceV1), strings.NewReader(data))
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

		r, err := http.Get(testhost(clusterInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&clusterid=" + newClusterid)
		assert.Nil(err)

		defer r.Body.Close()
		assert.Equal(http.StatusOK, r.StatusCode)

		b, err := respbody(r)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(b, "errCode").Int(), b)
		assert.Equal("OK", gjson.Get(b, "errMsg").String(), b)
		assert.Equal(newBid, gjson.Get(b, "cluster.bid").String(), b)
		assert.Equal(newAppid, gjson.Get(b, "cluster.appid").String(), b)
		assert.Equal(newClusterid, gjson.Get(b, "cluster.clusterid").String(), b)
		assert.Equal(gjson.Get(data, "name").String(), gjson.Get(b, "cluster.name").String(), b)
		assert.Equal(gjson.Get(data, "operator").String(), gjson.Get(b, "cluster.lastModifyBy").String(), b)
		assert.Equal(gjson.Get(data, "memo").String(), gjson.Get(b, "cluster.memo").String(), b)
		assert.Equal(gjson.Get(data, "rClusterid").String(), gjson.Get(b, "cluster.rClusterid").String(), b)
		assert.Equal(gjson.Get(data, "state").Int(), gjson.Get(b, "cluster.state").Int(), b)
	}

	{
		t.Logf("Case: delete cluster")
		req, err := http.NewRequest("DELETE", testhost(clusterInterfaceV1)+"?"+"seq=1&bid="+newBid+"&clusterid="+newClusterid+"&operator=e2e", nil)
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

		r, err := http.Get(testhost(clusterInterfaceV1) + "?" + "seq=1&bid=" + newBid + "&clusterid=" + newClusterid)
		assert.Nil(err)

		defer r.Body.Close()
		assert.Equal(http.StatusOK, r.StatusCode)

		b, err := respbody(r)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_DM_NOT_FOUND, gjson.Get(b, "errCode").Int(), b)
		assert.NotEqual("OK", gjson.Get(b, "errMsg").String(), b)
	}
}
