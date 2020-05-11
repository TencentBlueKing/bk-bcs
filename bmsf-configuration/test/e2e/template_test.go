/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package e2e

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/internal/structs"
	e2edata "bk-bscp/test/e2e/testdata"
)

func randName(prefix string) string {
	return fmt.Sprintf("%s-%s-%d", prefix, time.Now().Format("2006-01-02-15:04:05"), time.Now().Nanosecond())
}

func httpFunc(functionName, url, data string) (string, error) {
	req, err := http.NewRequest(functionName, url, strings.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("err status code %d", resp.StatusCode)
	}

	body, err := respbody(resp)
	if err != nil {
		return "", err
	}
	return body, nil
}

func httpDelete(urlStr string, data map[string]string) (string, error) {
	params := url.Values{}
	for key, value := range data {
		params.Add(key, value)
	}
	realURL := urlStr + "?" + params.Encode()

	fmt.Println("delete ", realURL)

	req, err := http.NewRequest("DELETE", realURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("err status code %d", resp.StatusCode)
	}

	body, err := respbody(resp)
	if err != nil {
		return "", err
	}
	return body, nil
}

func httpGet(urlStr string, data map[string]string) (string, error) {

	params := url.Values{}
	for key, value := range data {
		params.Add(key, value)
	}
	realURL := urlStr + "?" + params.Encode()

	fmt.Println("get ", realURL)

	resp, err := http.Get(realURL)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("err status code %d", resp.StatusCode)
	}

	body, err := respbody(resp)
	if err != nil {
		return "", err
	}
	return body, nil
}

var (
	newBid             = ""
	newAppid           = ""
	newClusterid1      = ""
	cluster1Data       = ""
	newClusterid1Zone1 = ""
	cluster1Zone1Data  = ""
	newClusterid1Zone2 = ""
	cluster1Zone2Data  = ""
	newClusterid2      = ""
	cluster2Data       = ""
	newClusterid2Zone1 = ""
	cluster2Zone1Data  = ""
	newClusterid2Zone2 = ""
	cluster2Zone2Data  = ""
)

func prepareClusterZone(t *testing.T, assert *assert.Assertions) {
	// create business
	{
		data, err := e2edata.CreateBusinessTestData()
		assert.Nil(err)

		body, err := httpFunc("POST", testhost(businessInterfaceV1), data)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "bid").String(), body)
		newBid = gjson.Get(body, "bid").String()
	}

	//  create app
	{
		data, err := e2edata.CreateAppTestData(newBid)
		assert.Nil(err)

		body, err := httpFunc("POST", testhost(appInterfaceV1), data)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "appid").String(), body)
		newAppid = gjson.Get(body, "appid").String()
	}

	// create cluster1
	// create zone1 of cluster1
	// create zone2 of cluster1
	cluster1Data, _ = e2edata.CreateClusterTestDataWithLabel(newBid, newAppid)
	{
		t.Logf("Case: create new cluster 1")
		body, err := httpFunc("POST", testhost(clusterInterfaceV1), cluster1Data)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "clusterid").String(), body)
		newClusterid1 = gjson.Get(body, "clusterid").String()

	}
	cluster1Zone1Data, _ = e2edata.CreateZoneTestData(newBid, newAppid, newClusterid1)
	{
		t.Logf("Case: create new cluster 1 zone 1")
		body, err := httpFunc("POST", testhost(zoneInterfaceV1), cluster1Zone1Data)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "zoneid").String(), body)
		newClusterid1Zone1 = gjson.Get(body, "zoneid").String()
	}
	cluster1Zone2Data, _ = e2edata.CreateZoneTestData(newBid, newAppid, newClusterid1)
	{
		t.Logf("Case: create new cluster 1 zone 2")
		body, err := httpFunc("POST", testhost(zoneInterfaceV1), cluster1Zone2Data)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "zoneid").String(), body)
		newClusterid1Zone2 = gjson.Get(body, "zoneid").String()
	}

	// create cluster2
	// create zone1 of cluster2
	// create zone2 of cluster2
	cluster2Data, _ = e2edata.CreateClusterTestDataWithLabel(newBid, newAppid)
	{
		t.Logf("Case: create new cluster 2")
		body, err := httpFunc("POST", testhost(clusterInterfaceV1), cluster2Data)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "clusterid").String(), body)
		newClusterid2 = gjson.Get(body, "clusterid").String()

	}
	cluster2Zone1Data, _ = e2edata.CreateZoneTestData(newBid, newAppid, newClusterid1)
	{
		t.Logf("Case: create new cluster 2 zone 1")
		body, err := httpFunc("POST", testhost(zoneInterfaceV1), cluster2Zone1Data)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "zoneid").String(), body)
		newClusterid2Zone1 = gjson.Get(body, "zoneid").String()
	}
	cluster2Zone2Data, _ = e2edata.CreateZoneTestData(newBid, newAppid, newClusterid1)
	{
		t.Logf("Case: create new cluster 2 zone 2")
		body, err := httpFunc("POST", testhost(zoneInterfaceV1), cluster2Zone2Data)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "zoneid").String(), body)
		newClusterid2Zone2 = gjson.Get(body, "zoneid").String()
	}
}

func deleteClusterZone(t *testing.T, assert *assert.Assertions) {
	// delete zone1 of cluster1
	{
		t.Logf("Case: delete clueter1 zone1")
		body, err := httpDelete(testhost(zoneInterfaceV1), map[string]string{
			"seq":       "0",
			"bid":       newBid,
			"clusterid": newClusterid1,
			"zoneid":    newClusterid1Zone1,
			"operator":  "e2e",
		})
		assert.Nil(err)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}
	// delete zone2 of cluster1
	{
		t.Logf("Case: delete clueter1 zone2")
		body, err := httpDelete(testhost(zoneInterfaceV1), map[string]string{
			"seq":       "0",
			"bid":       newBid,
			"clusterid": newClusterid1,
			"zoneid":    newClusterid1Zone2,
			"operator":  "e2e",
		})
		assert.Nil(err)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}
	// delete zone1 of cluster2
	{
		t.Logf("Case: delete clueter2 zone1")
		body, err := httpDelete(testhost(zoneInterfaceV1), map[string]string{
			"seq":       "0",
			"bid":       newBid,
			"clusterid": newClusterid2,
			"zoneid":    newClusterid2Zone1,
			"operator":  "e2e",
		})
		assert.Nil(err)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}
	// delete zone2 of cluster2
	{
		t.Logf("Case: delete clueter2 zone1")
		body, err := httpDelete(testhost(zoneInterfaceV1), map[string]string{
			"seq":       "0",
			"bid":       newBid,
			"clusterid": newClusterid2,
			"zoneid":    newClusterid2Zone2,
			"operator":  "e2e",
		})
		assert.Nil(err)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	// delete cluster1
	{
		t.Logf("Case: delete clueter1")
		body, err := httpDelete(testhost(clusterInterfaceV1), map[string]string{
			"seq":       "0",
			"bid":       newBid,
			"clusterid": newClusterid1,
			"operator":  "e2e",
		})
		assert.Nil(err)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	// delete cluster2
	{
		t.Logf("Case: delete clueter2")
		body, err := httpDelete(testhost(clusterInterfaceV1), map[string]string{
			"seq":       "0",
			"bid":       newBid,
			"clusterid": newClusterid2,
			"operator":  "e2e",
		})
		assert.Nil(err)
		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}
}

func TestVariable(t *testing.T) {

	assert := assert.New(t)

	prepareClusterZone(t, assert)

	/// =================== test global var =====================
	// create global vars
	newGlobalVarid := ""
	globalVarData, err := e2edata.CreateVarTestData(0, newBid, "", "", "")
	assert.Nil(err)
	{
		t.Logf("Case: create new global var")
		body, err := httpFunc("POST", testhost(variableInterfaceV1), globalVarData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "vid").String(), body)
		newGlobalVarid = gjson.Get(body, "vid").String()
	}

	// update global var
	updateGlobalVarData, err := e2edata.UpdateVarTestData(0, newBid, newGlobalVarid)
	{
		t.Logf("Case: update global var")
		body, err := httpFunc("PUT", testhost(variableInterfaceV1), updateGlobalVarData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	// query global var
	{
		t.Logf("Case: query global var")
		body, err := httpGet(testhost(variableInterfaceV1), map[string]string{
			"seq": "0",
			"bid": newBid,
			"vid": newGlobalVarid,
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(gjson.Get(updateGlobalVarData, "key").String(), gjson.Get(body, "var.key").String(), body)
		assert.Equal(gjson.Get(updateGlobalVarData, "value").String(), gjson.Get(body, "var.value").String(), body)
		assert.Equal(gjson.Get(updateGlobalVarData, "memo").String(), gjson.Get(body, "var.memo").String(), body)
		assert.Equal(gjson.Get(updateGlobalVarData, "vid").String(), gjson.Get(body, "var.vid").String(), body)
	}

	// list global var
	{
		t.Logf("Case: query global var list")
		body, err := httpGet(testhost(variableListInterfaceV1), map[string]string{
			"seq":   "0",
			"bid":   newBid,
			"limit": "10",
		})
		assert.Nil(err)

		arrs := gjson.Get(body, "vars").Array()
		assert.NotEqual(0, len(arrs), arrs)

		for _, e := range arrs {
			if e.Get("vid").String() == newGlobalVarid {
				assert.Equal(gjson.Get(updateGlobalVarData, "key").String(), e.Get("key").String(), e)
				assert.Equal(gjson.Get(updateGlobalVarData, "value").String(), e.Get("value").String(), e)
				assert.Equal(gjson.Get(updateGlobalVarData, "memo").String(), e.Get("memo").String(), e)
			}
		}
	}

	// delete global var
	{
		t.Logf("Case: delete global var")
		body, err := httpDelete(testhost(variableInterfaceV1), map[string]string{
			"seq":      "0",
			"bid":      newBid,
			"vid":      newGlobalVarid,
			"operator": "e2e",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	// =================== test cluster var =====================
	// create cluster vars
	newClusterVarid := ""
	clusterVarData, err := e2edata.CreateVarTestData(1, newBid, gjson.Get(cluster1Data, "name").String(), gjson.Get(cluster1Data, "labels").String(), "")
	assert.Nil(err)
	{
		t.Logf("Case: create new cluster var")
		body, err := httpFunc("POST", testhost(variableInterfaceV1), clusterVarData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "vid").String(), body)
		newClusterVarid = gjson.Get(body, "vid").String()
	}

	// update cluster var
	updateClusterVarData, err := e2edata.UpdateVarTestData(1, newBid, newClusterVarid)
	{
		t.Logf("Case: update cluster var")
		body, err := httpFunc("PUT", testhost(variableInterfaceV1), updateClusterVarData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	// query cluster var
	{
		t.Logf("Case: query cluster var")
		body, err := httpGet(testhost(variableInterfaceV1), map[string]string{
			"seq":  "0",
			"bid":  newBid,
			"vid":  newClusterVarid,
			"type": "1",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(gjson.Get(updateClusterVarData, "key").String(), gjson.Get(body, "var.key").String(), body)
		assert.Equal(gjson.Get(updateClusterVarData, "value").String(), gjson.Get(body, "var.value").String(), body)
		assert.Equal(gjson.Get(updateClusterVarData, "memo").String(), gjson.Get(body, "var.memo").String(), body)
		assert.Equal(gjson.Get(updateClusterVarData, "vid").String(), gjson.Get(body, "var.vid").String(), body)
	}

	// list cluster var
	{
		t.Logf("Case: query cluster var list")
		body, err := httpGet(testhost(variableListInterfaceV1), map[string]string{
			"seq":           "0",
			"bid":           newBid,
			"cluster":       gjson.Get(cluster1Data, "name").String(),
			"clusterLabels": gjson.Get(cluster1Data, "labels").String(),
			"type":          "1",
			"limit":         "10",
		})
		assert.Nil(err)

		arrs := gjson.Get(body, "vars").Array()
		assert.NotEqual(0, len(arrs), arrs)

		for _, e := range arrs {
			if e.Get("vid").String() == newGlobalVarid {
				assert.Equal(gjson.Get(updateClusterVarData, "key").String(), e.Get("key").String(), e)
				assert.Equal(gjson.Get(updateClusterVarData, "value").String(), e.Get("value").String(), e)
				assert.Equal(gjson.Get(updateClusterVarData, "memo").String(), e.Get("memo").String(), e)
			}
		}
	}

	// delete cluster var
	{
		t.Logf("Case: delete cluster var")
		body, err := httpDelete(testhost(variableInterfaceV1), map[string]string{
			"seq":      "0",
			"bid":      newBid,
			"vid":      newClusterVarid,
			"type":     "1",
			"operator": "e2e",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	// =================== test zone var =====================
	// create zone vars
	newZoneVarid := ""
	zoneVarData, err := e2edata.CreateVarTestData(1, newBid, gjson.Get(cluster1Data, "name").String(), gjson.Get(cluster1Data, "labels").String(), gjson.Get(cluster1Zone1Data, "name").String())
	assert.Nil(err)
	{
		t.Logf("Case: create new cluster var")
		body, err := httpFunc("POST", testhost(variableInterfaceV1), zoneVarData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "vid").String(), body)
		newZoneVarid = gjson.Get(body, "vid").String()
	}

	// update zone var
	updateZoneVarData, err := e2edata.UpdateVarTestData(1, newBid, newZoneVarid)
	{
		t.Logf("Case: update zone var")
		body, err := httpFunc("PUT", testhost(variableInterfaceV1), updateZoneVarData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	// query zone var
	{
		t.Logf("Case: query zone var")
		body, err := httpGet(testhost(variableInterfaceV1), map[string]string{
			"seq":           "0",
			"bid":           newBid,
			"cluster":       gjson.Get(cluster1Data, "name").String(),
			"clusterLabels": gjson.Get(cluster1Data, "labels").String(),
			"vid":           newZoneVarid,
			"type":          "1",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(gjson.Get(updateZoneVarData, "key").String(), gjson.Get(body, "var.key").String(), body)
		assert.Equal(gjson.Get(updateZoneVarData, "value").String(), gjson.Get(body, "var.value").String(), body)
		assert.Equal(gjson.Get(updateZoneVarData, "memo").String(), gjson.Get(body, "var.memo").String(), body)
		assert.Equal(gjson.Get(updateZoneVarData, "vid").String(), gjson.Get(body, "var.vid").String(), body)
	}

	// list zone var
	{
		t.Logf("Case: query zone var list")
		body, err := httpGet(testhost(variableListInterfaceV1), map[string]string{
			"seq":           "0",
			"bid":           newBid,
			"cluster":       gjson.Get(cluster1Data, "name").String(),
			"clusterLabels": gjson.Get(cluster1Data, "labels").String(),
			"zone":          gjson.Get(cluster1Zone1Data, "name").String(),
			"type":          "1",
			"limit":         "10",
		})
		assert.Nil(err)

		arrs := gjson.Get(body, "vars").Array()
		assert.NotEqual(0, len(arrs), arrs)

		for _, e := range arrs {
			if e.Get("vid").String() == newZoneVarid {
				assert.Equal(gjson.Get(updateZoneVarData, "key").String(), e.Get("key").String(), e)
				assert.Equal(gjson.Get(updateZoneVarData, "value").String(), e.Get("value").String(), e)
				assert.Equal(gjson.Get(updateZoneVarData, "memo").String(), e.Get("memo").String(), e)
			}
		}
	}

	// delete zone var
	{
		t.Logf("Case: delete zone var")
		body, err := httpDelete(testhost(variableInterfaceV1), map[string]string{
			"seq":      "0",
			"bid":      newBid,
			"vid":      newZoneVarid,
			"type":     "1",
			"operator": "e2e",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	deleteClusterZone(t, assert)

}

func TestConfigTemplateSet(t *testing.T) {

	assert := assert.New(t)

	prepareClusterZone(t, assert)

	// =================== test config template =====================
	// create config template set
	newFpath := randName("/e2e-fpath")
	newConfigTemplateSetID := ""
	newConfigTemplateSetData, _ := e2edata.CreateConfigTemplateSetTestData(newBid, newFpath)
	{
		t.Logf("Case: create config template set")
		body, err := httpFunc("POST", testhost(templatesetInterfaceV1), newConfigTemplateSetData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "setid").String(), body)
		newConfigTemplateSetID = gjson.Get(body, "setid").String()
	}

	// update config template set
	updateConfigTemplateSetData, _ := e2edata.UpdateConfigTemplateSetTestData(newBid, newConfigTemplateSetID)
	{
		t.Logf("Case: update config template set")
		body, err := httpFunc("PUT", testhost(templatesetInterfaceV1), updateConfigTemplateSetData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	// query config template set
	{
		t.Logf("Case: query config template set")
		body, err := httpGet(testhost(templatesetInterfaceV1), map[string]string{
			"seq":   "0",
			"bid":   newBid,
			"setid": newConfigTemplateSetID,
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(gjson.Get(updateConfigTemplateSetData, "name").String(), gjson.Get(body, "templateSet.name").String(), body)
		assert.Equal(gjson.Get(updateConfigTemplateSetData, "memo").String(), gjson.Get(body, "templateSet.memo").String(), body)
		assert.Equal(gjson.Get(newConfigTemplateSetData, "fpath").String(), gjson.Get(body, "templateSet.fpath").String(), body)
		assert.Equal(gjson.Get(newConfigTemplateSetData, "creator").String(), gjson.Get(body, "templateSet.creator").String(), body)
		assert.Equal(gjson.Get(updateConfigTemplateSetData, "operator").String(), gjson.Get(body, "templateSet.lastModifyBy").String(), body)
	}

	// list config template set
	{
		t.Logf("Case: list config template set")
		body, err := httpGet(testhost(templatesetListInterfaceV1), map[string]string{
			"seq":   "0",
			"bid":   newBid,
			"limit": "10",
		})
		assert.Nil(err)

		arrs := gjson.Get(body, "templateSets").Array()
		assert.NotEqual(0, len(arrs), arrs)

		for _, e := range arrs {
			if e.Get("setid").String() == newConfigTemplateSetID {
				assert.Equal(gjson.Get(updateConfigTemplateSetData, "name").String(), e.Get("name").String(), body)
				assert.Equal(gjson.Get(updateConfigTemplateSetData, "memo").String(), e.Get("memo").String(), body)
				assert.Equal(gjson.Get(newConfigTemplateSetData, "fpath").String(), e.Get("fpath").String(), body)
				assert.Equal(gjson.Get(newConfigTemplateSetData, "creator").String(), e.Get("creator").String(), body)
				assert.Equal(gjson.Get(updateConfigTemplateSetData, "operator").String(), e.Get("lastModifyBy").String(), body)
			}
		}
	}

	// delete config template set
	{
		t.Logf("Case: delete config template set")
		body, err := httpDelete(testhost(templatesetInterfaceV1), map[string]string{
			"seq":      "0",
			"bid":      newBid,
			"setid":    newConfigTemplateSetID,
			"operator": "e2e",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	deleteClusterZone(t, assert)

}

// TestConfigTemplate test config template
func TestConfigTemplate(t *testing.T) {

	assert := assert.New(t)

	prepareClusterZone(t, assert)

	// create config template set
	newFpath := randName("/e2e-fpath")
	newConfigTemplateSetID := ""
	newConfigTemplateSetData, _ := e2edata.CreateConfigTemplateSetTestData(newBid, newFpath)
	{
		t.Logf("Case: create config template set")
		body, err := httpFunc("POST", testhost(templatesetInterfaceV1), newConfigTemplateSetData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "setid").String(), body)
		newConfigTemplateSetID = gjson.Get(body, "setid").String()
	}

	// create config template
	newConfigTemplateData, _ := e2edata.CreateConfigTemplateTestData(newBid, newConfigTemplateSetID)
	newConfigTemplateID := ""
	{
		t.Logf("Case: create config template")
		body, err := httpFunc("POST", testhost(templateInterfaceV1), newConfigTemplateData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "templateid").String(), body)
		newConfigTemplateID = gjson.Get(body, "templateid").String()
	}

	// update config template
	updateConfigTemplateData, _ := e2edata.UpdateConfigTemplateTestData(newBid, newConfigTemplateID)
	{
		t.Logf("Case: update config template")
		body, err := httpFunc("PUT", testhost(templateInterfaceV1), updateConfigTemplateData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	// query config template
	{
		t.Logf("Case: query config template")
		body, err := httpGet(testhost(templateInterfaceV1), map[string]string{
			"seq":        "0",
			"bid":        newBid,
			"templateid": newConfigTemplateID,
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(gjson.Get(updateConfigTemplateData, "name").String(), gjson.Get(body, "configTemplate.name").String(), body)
		assert.Equal(gjson.Get(updateConfigTemplateData, "memo").String(), gjson.Get(body, "configTemplate.memo").String(), body)
		assert.Equal(gjson.Get(updateConfigTemplateData, "user").String(), gjson.Get(body, "configTemplate.user").String(), body)
		assert.Equal(gjson.Get(updateConfigTemplateData, "group").String(), gjson.Get(body, "configTemplate.group").String(), body)
		assert.Equal(gjson.Get(updateConfigTemplateData, "fileEncoding").String(), gjson.Get(body, "configTemplate.fileEncoding").String(), body)
		assert.Equal(gjson.Get(updateConfigTemplateData, "permission").Int(), gjson.Get(body, "configTemplate.permission").Int(), body)
		assert.Equal(newConfigTemplateSetID, gjson.Get(body, "configTemplate.setid").String(), body)
		assert.Equal(gjson.Get(newConfigTemplateSetData, "fpath").String(), gjson.Get(body, "configTemplate.fpath").String(), body)
		assert.Equal(gjson.Get(newConfigTemplateData, "creator").String(), gjson.Get(body, "configTemplate.creator").String(), body)
		assert.Equal(gjson.Get(updateConfigTemplateData, "operator").String(), gjson.Get(body, "configTemplate.lastModifyBy").String(), body)
	}

	// list config template
	{
		t.Logf("Case: list config template")
		body, err := httpGet(testhost(templateListInterfaceV1), map[string]string{
			"seq":   "0",
			"bid":   newBid,
			"setid": newConfigTemplateSetID,
			"limit": "10",
		})
		assert.Nil(err)

		arrs := gjson.Get(body, "configTemplates").Array()
		assert.NotEqual(0, len(arrs), arrs)

		for _, e := range arrs {
			if e.Get("templateid").String() == newConfigTemplateID {
				assert.Equal(gjson.Get(updateConfigTemplateData, "name").String(), e.Get("name").String(), body)
				assert.Equal(gjson.Get(updateConfigTemplateData, "memo").String(), e.Get("memo").String(), body)
				assert.Equal(gjson.Get(updateConfigTemplateData, "user").String(), e.Get("user").String(), body)
				assert.Equal(gjson.Get(updateConfigTemplateData, "group").String(), e.Get("group").String(), body)
				assert.Equal(gjson.Get(updateConfigTemplateData, "fileEncoding").String(), e.Get("fileEncoding").String(), body)
				assert.Equal(gjson.Get(updateConfigTemplateData, "permission").Int(), e.Get("permission").Int(), body)
				assert.Equal(newConfigTemplateSetID, e.Get("setid").String(), body)
				assert.Equal(gjson.Get(newConfigTemplateSetData, "fpath").String(), e.Get("fpath").String(), body)
				assert.Equal(gjson.Get(newConfigTemplateData, "creator").String(), e.Get("creator").String(), body)
				assert.Equal(gjson.Get(updateConfigTemplateData, "operator").String(), e.Get("lastModifyBy").String(), body)
			}
		}
	}

	// delete config template
	{
		t.Logf("Case: delete config template")
		body, err := httpDelete(testhost(templateInterfaceV1), map[string]string{
			"seq":        "0",
			"bid":        newBid,
			"templateid": newConfigTemplateID,
			"operator":   "e2e",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	// delete config template set
	{
		t.Logf("Case: delete config template set")
		body, err := httpDelete(testhost(templatesetInterfaceV1), map[string]string{
			"seq":      "0",
			"bid":      newBid,
			"setid":    newConfigTemplateSetID,
			"operator": "e2e",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	deleteClusterZone(t, assert)
}

// TestConfigTemplateVersion test config template version
func TestConfigTemplateVersion(t *testing.T) {

	assert := assert.New(t)

	prepareClusterZone(t, assert)

	// create config template set
	newFpath := randName("/e2e-fpath")
	newConfigTemplateSetID := ""
	newConfigTemplateSetData, _ := e2edata.CreateConfigTemplateSetTestData(newBid, newFpath)
	{
		t.Logf("Case: create config template set")
		body, err := httpFunc("POST", testhost(templatesetInterfaceV1), newConfigTemplateSetData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "setid").String(), body)
		newConfigTemplateSetID = gjson.Get(body, "setid").String()
	}

	// create config template
	newConfigTemplateData, _ := e2edata.CreateConfigTemplateTestData(newBid, newConfigTemplateSetID)
	newConfigTemplateID := ""
	{
		t.Logf("Case: create config template")
		body, err := httpFunc("POST", testhost(templateInterfaceV1), newConfigTemplateData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "templateid").String(), body)
		newConfigTemplateID = gjson.Get(body, "templateid").String()
	}

	// create config template version
	newVersionID := ""
	newTemplateVersionData, _ := e2edata.CreateTemplateVersionTestData(newBid, newConfigTemplateID)
	{
		t.Logf("Case: create config template version")
		body, err := httpFunc("POST", testhost(templateversionInterfaceV1), newTemplateVersionData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "versionid").String(), body)
		newVersionID = gjson.Get(body, "versionid").String()
	}

	// update config template version
	updateTemplateVersionData, _ := e2edata.UpdateTemplateVersionTestData(newBid, newVersionID)
	{
		t.Logf("Case: update config template version")
		body, err := httpFunc("PUT", testhost(templateversionInterfaceV1), updateTemplateVersionData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	// query config template version
	{
		t.Logf("Case: query config template version")
		body, err := httpGet(testhost(templateversionInterfaceV1), map[string]string{
			"bid":       newBid,
			"versionid": newVersionID,
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(gjson.Get(updateTemplateVersionData, "versionName").String(), gjson.Get(body, "templateVersion.versionName").String(), body)
		assert.Equal(gjson.Get(updateTemplateVersionData, "memo").String(), gjson.Get(body, "templateVersion.memo").String(), body)
		assert.Equal(gjson.Get(updateTemplateVersionData, "content").String(), gjson.Get(body, "templateVersion.content").String(), body)
		assert.Equal(gjson.Get(updateTemplateVersionData, "creator").String(), gjson.Get(body, "templateVersions.creator").String(), body)
	}

	// query config template version list
	{
		t.Logf("Case: query config template version list")
		body, err := httpGet(testhost(templateversionListInterfaceV1), map[string]string{
			"bid":        newBid,
			"templateid": newConfigTemplateID,
			"limit":      "10",
		})
		assert.Nil(err)

		arrs := gjson.Get(body, "templateVersions").Array()
		assert.NotEqual(0, len(arrs), arrs)

		for _, e := range arrs {
			if e.Get("versionid").String() == newVersionID {
				assert.Equal(gjson.Get(updateTemplateVersionData, "versionName").String(), e.Get("versionName").String(), body)
				assert.Equal(gjson.Get(updateTemplateVersionData, "memo").String(), e.Get("memo").String(), body)
				assert.Equal(gjson.Get(updateTemplateVersionData, "content").String(), e.Get("content").String(), body)
				assert.Equal(gjson.Get(newTemplateVersionData, "creator").String(), e.Get("creator").String(), body)
			}
		}
	}

	// delete config template version
	{
		t.Logf("Case: delete config template version")
		body, err := httpDelete(testhost(templateversionInterfaceV1), map[string]string{
			"bid":       newBid,
			"versionid": newVersionID,
			"operator":  "e2e",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	// delete config template
	{
		t.Logf("Case: delete config template")
		body, err := httpDelete(testhost(templateInterfaceV1), map[string]string{
			"seq":        "0",
			"bid":        newBid,
			"templateid": newConfigTemplateID,
			"operator":   "e2e",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	// delete config template set
	{
		t.Logf("Case: delete config template set")
		body, err := httpDelete(testhost(templatesetInterfaceV1), map[string]string{
			"seq":      "0",
			"bid":      newBid,
			"setid":    newConfigTemplateSetID,
			"operator": "e2e",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	deleteClusterZone(t, assert)

}

// TestTemplateBinding
func TestTemplateBinding(t *testing.T) {

	assert := assert.New(t)

	prepareClusterZone(t, assert)

	// create config template set
	newFpath := randName("/e2e-fpath")
	newConfigTemplateSetID := ""
	newConfigTemplateSetData, _ := e2edata.CreateConfigTemplateSetTestData(newBid, newFpath)
	{
		t.Logf("Case: create config template set")
		body, err := httpFunc("POST", testhost(templatesetInterfaceV1), newConfigTemplateSetData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "setid").String(), body)
		newConfigTemplateSetID = gjson.Get(body, "setid").String()
	}

	// create config template
	newConfigTemplateData, _ := e2edata.CreateConfigTemplateTestData(newBid, newConfigTemplateSetID)
	newConfigTemplateID := ""
	{
		t.Logf("Case: create config template")
		body, err := httpFunc("POST", testhost(templateInterfaceV1), newConfigTemplateData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "templateid").String(), body)
		newConfigTemplateID = gjson.Get(body, "templateid").String()
	}

	// create config template version
	newVersionID := ""
	newTemplateVersionData, _ := e2edata.CreateTemplateVersionTestData(newBid, newConfigTemplateID)
	{
		t.Logf("Case: create config template version")
		body, err := httpFunc("POST", testhost(templateversionInterfaceV1), newTemplateVersionData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "versionid").String(), body)
		newVersionID = gjson.Get(body, "versionid").String()
	}

	cluster1Labels := make(map[string]string)
	err := json.Unmarshal([]byte(gjson.Get(cluster1Data, "labels").String()), &cluster1Labels)
	assert.Nil(err)

	newRules := structs.RuleList{
		structs.Rule{
			Cluster:       gjson.Get(cluster1Data, "name").String(),
			ClusterLabels: cluster1Labels,
			Zones: []*structs.RuleZone{
				{
					Zone: gjson.Get(cluster1Zone1Data, "name").String(),
					Instances: []*structs.RuleInstance{
						{
							Index: "127.0.0.1",
							Variables: map[string]interface{}{
								"InnerIP":   "127.0.0.1",
								"Placement": "黄鹤楼",
							},
						},
						{
							Index: "127.0.0.2",
							Variables: map[string]interface{}{
								"InnerIP":   "127.0.0.2",
								"Placement": "鹦鹉洲",
							},
						},
					},
				},
				{
					Zone: gjson.Get(cluster1Zone2Data, "name").String(),
					Instances: []*structs.RuleInstance{
						{
							Index: "127.0.0.3",
							Variables: map[string]interface{}{
								"InnerIP":   "127.0.0.3",
								"Placement": "深圳湾",
							},
						},
						{
							Index: "127.0.0.4",
							Variables: map[string]interface{}{
								"InnerIP":   "127.0.0.4",
								"Placement": "火焰山",
							},
						},
					},
				},
			},
		},
	}
	newRulesStr, _ := json.Marshal(newRules)

	updateRules := structs.RuleList{
		structs.Rule{
			Cluster:       gjson.Get(cluster1Data, "name").String(),
			ClusterLabels: cluster1Labels,
			Zones: []*structs.RuleZone{
				{
					Zone: gjson.Get(cluster1Zone1Data, "name").String(),
					Instances: []*structs.RuleInstance{
						{
							Index: "127.0.0.1",
							Variables: map[string]interface{}{
								"InnerIP":   "127.0.0.1",
								"Placement": "黄鹤楼",
							},
						},
						{
							Index: "127.0.0.2",
							Variables: map[string]interface{}{
								"InnerIP":   "127.0.0.2",
								"Placement": "鹦鹉洲",
							},
						},
					},
				},
				{
					Zone: gjson.Get(cluster1Zone2Data, "name").String(),
					Instances: []*structs.RuleInstance{
						{
							Index: "127.0.0.7",
							Variables: map[string]interface{}{
								"InnerIP":   "127.0.0.7",
								"Placement": "深圳湾",
							},
						},
						{
							Index: "127.0.0.8",
							Variables: map[string]interface{}{
								"InnerIP":   "127.0.0.8",
								"Placement": "火焰山",
							},
						},
					},
				},
			},
		},
	}
	updateRulesStr, _ := json.Marshal(updateRules)

	// create template binding
	newTemplateBindingData, _ := e2edata.CreateTemplateBindingTestData(newBid, newConfigTemplateID, newAppid, newVersionID, string(newRulesStr))
	{
		t.Logf("Case: create template binding")
		body, err := httpFunc("POST", testhost(templatebindingInterfaceV1), newTemplateBindingData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	// update template binding
	updateTemplateBindingData, _ := e2edata.UpdateTemplateBindingTestData(newBid, newConfigTemplateID, newAppid, newVersionID, string(updateRulesStr))
	{
		t.Logf("Case: update template binding")
		body, err := httpFunc("PUT", testhost(templatebindingInterfaceV1), updateTemplateBindingData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	// query template binding
	innerCfgsetid := ""
	innerCommitid := ""
	{
		t.Logf("Case: query template binding")
		body, err := httpGet(testhost(templatebindingInterfaceV1), map[string]string{
			"seq":        "0",
			"bid":        newBid,
			"templateid": newConfigTemplateID,
			"appid":      newAppid,
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.Equal(newBid, gjson.Get(body, "configTemplateBinding.bid").String(), body)
		assert.Equal(newConfigTemplateID, gjson.Get(body, "configTemplateBinding.templateid").String(), body)
		assert.Equal(newAppid, gjson.Get(body, "configTemplateBinding.appid").String(), body)
		assert.Equal(string(updateRulesStr), gjson.Get(body, "configTemplateBinding.bindingParams").String(), body)
		innerCfgsetid = gjson.Get(body, "configTemplateBinding.cfgsetid").String()
		innerCommitid = gjson.Get(body, "configTemplateBinding.commitid").String()
	}

	// query template binding list
	{
		t.Logf("Case: query template binding list")
		body, err := httpGet(testhost(templatebindingListInterfaceV1), map[string]string{
			"seq":        "0",
			"bid":        newBid,
			"templateid": newConfigTemplateID,
			"limit":      "10",
		})
		assert.Nil(err)

		arrs := gjson.Get(body, "configTemplateBindings").Array()
		assert.NotEqual(0, len(arrs), arrs)

		for _, e := range arrs {
			if e.Get("templateid").String() == newConfigTemplateID && e.Get("appid").String() == newAppid {
				assert.Equal(newBid, e.Get("bid").String(), body)
				assert.Equal(newConfigTemplateID, e.Get("templateid").String(), body)
				assert.Equal(newAppid, e.Get("appid").String(), body)
				assert.Equal(gjson.Get(updateTemplateBindingData, "bindingParams").String(), e.Get("bindingParams").String(), body)
			}
		}
	}

	newInnerGlobalVarData, _ := e2edata.CreateCertainVarTestData(newBid, 0, "", "", "", "TITLE", "将进酒")
	newInnerGlobalVarID := ""
	{
		t.Logf("Case: create global var")
		body, err := httpFunc("POST", testhost(variableInterfaceV1), newInnerGlobalVarData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "vid").String(), body)
		newInnerGlobalVarID = gjson.Get(body, "vid").String()
	}

	newInnerClusterVarData, _ := e2edata.CreateCertainVarTestData(
		newBid, 1, gjson.Get(cluster1Data, "name").String(),
		gjson.Get(cluster1Data, "labels").String(), "", "AUTHOR", "李白")
	newInnerClusterID := ""
	{
		t.Logf("Case: create cluster var")
		body, err := httpFunc("POST", testhost(variableInterfaceV1), newInnerClusterVarData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "vid").String(), body)
		newInnerClusterID = gjson.Get(body, "vid").String()
	}

	newInnerZoneVar1Data, _ := e2edata.CreateCertainVarTestData(
		newBid, 2, gjson.Get(cluster1Data, "name").String(),
		gjson.Get(cluster1Data, "labels").String(), gjson.Get(cluster1Zone1Data, "name").String(), "DEST", "唐朝")
	newInnerZoneVar1ID := ""
	{
		t.Logf("Case: create zone var")
		body, err := httpFunc("POST", testhost(variableInterfaceV1), newInnerZoneVar1Data)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "vid").String(), body)
		newInnerZoneVar1ID = gjson.Get(body, "vid").String()
	}

	newInnerZoneVar2Data, _ := e2edata.CreateCertainVarTestData(
		newBid, 2, gjson.Get(cluster1Data, "name").String(),
		gjson.Get(cluster1Data, "labels").String(), gjson.Get(cluster1Zone2Data, "name").String(), "DEST", "宋朝")
	newInnerZoneVar2ID := ""
	{
		t.Logf("Case: create zone var")
		body, err := httpFunc("POST", testhost(variableInterfaceV1), newInnerZoneVar2Data)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
		assert.NotEmpty(gjson.Get(body, "vid").String(), body)
		newInnerZoneVar2ID = gjson.Get(body, "vid").String()
	}

	// confirm commit
	newConfirmCommitData, _ := e2edata.CreateConfirmCommitWithTemplateTestData(newBid, innerCommitid)
	{
		t.Logf("Case: confirm commit")
		body, err := httpFunc("PUT", testhost(commitConfirmInterfaceV1), newConfirmCommitData)
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	zoneVars := []string{
		"唐朝",
		"宋朝",
	}

	zoneids := []string{
		newClusterid1Zone1,
		newClusterid1Zone2,
	}

	{
		for _, c := range updateRules {
			vars := make(map[string]interface{})
			vars["TITLE"] = "将进酒"
			vars["AUTHOR"] = "李白"
			for zIndex, r := range c.Zones {
				vars["DEST"] = zoneVars[zIndex]
				for _, i := range r.Instances {
					t.Logf("Case: assert configs")
					body, err := httpGet(testhost(configsInterfaceV1), map[string]string{
						"seq":       "0",
						"bid":       newBid,
						"appid":     newAppid,
						"clusterid": newClusterid1,
						"zoneid":    zoneids[zIndex],
						"cfgsetid":  innerCfgsetid,
						"commitid":  innerCommitid,
						"index":     i.Index,
					})
					assert.Nil(err)
					assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
					assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)

					vars["InnerIP"] = i.Variables["InnerIP"]
					vars["Placement"] = i.Variables["Placement"]

					t, err := template.New("").Parse(gjson.Get(newTemplateVersionData, "content").String())
					assert.Nil(err)
					buffer := bytes.NewBuffer(nil)
					// rendering template.
					err = t.Execute(buffer, vars)
					assert.Nil(err)

					tmpBytes, err := base64.StdEncoding.DecodeString(gjson.Get(body, "configs.content").String())
					assert.Nil(err)

					assert.Equal(buffer.String(), string(tmpBytes), body)
				}
			}
		}
	}

	// delete var
	{
		t.Logf("Case: delete var")
		body, err := httpDelete(testhost(variableInterfaceV1), map[string]string{
			"seq":      "0",
			"bid":      newBid,
			"type":     "0",
			"vid":      newInnerGlobalVarID,
			"operator": "e2e",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)

		body, err = httpDelete(testhost(variableInterfaceV1), map[string]string{
			"seq":      "0",
			"bid":      newBid,
			"type":     "1",
			"vid":      newInnerClusterID,
			"operator": "e2e",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)

		body, err = httpDelete(testhost(variableInterfaceV1), map[string]string{
			"seq":      "0",
			"bid":      newBid,
			"type":     "2",
			"vid":      newInnerZoneVar1ID,
			"operator": "e2e",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)

		body, err = httpDelete(testhost(variableInterfaceV1), map[string]string{
			"seq":      "0",
			"bid":      newBid,
			"type":     "2",
			"vid":      newInnerZoneVar2ID,
			"operator": "e2e",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)

	}

	// delete template binding
	{
		t.Logf("Case: delete template binding")
		body, err := httpDelete(testhost(templatebindingInterfaceV1), map[string]string{
			"seq":        "0",
			"bid":        newBid,
			"templateid": newConfigTemplateID,
			"appid":      newAppid,
			"operator":   "e2e",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	// delete config template version
	{
		t.Logf("Case: delete config template version")
		body, err := httpDelete(testhost(templateversionInterfaceV1), map[string]string{
			"bid":       newBid,
			"versionid": newVersionID,
			"operator":  "e2e",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	// delete config template
	{
		t.Logf("Case: delete config template")
		body, err := httpDelete(testhost(templateInterfaceV1), map[string]string{
			"seq":        "0",
			"bid":        newBid,
			"templateid": newConfigTemplateID,
			"operator":   "e2e",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	// delete config template set
	{
		t.Logf("Case: delete config template set")
		body, err := httpDelete(testhost(templatesetInterfaceV1), map[string]string{
			"seq":      "0",
			"bid":      newBid,
			"setid":    newConfigTemplateSetID,
			"operator": "e2e",
		})
		assert.Nil(err)

		assert.EqualValues(pbcommon.ErrCode_E_OK, gjson.Get(body, "errCode").Int(), body)
		assert.Equal("OK", gjson.Get(body, "errMsg").String(), body)
	}

	deleteClusterZone(t, assert)
}
