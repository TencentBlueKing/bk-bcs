/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package u1_21_202110211130

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-upgrader/upgrader"
)

func getCCToken() (string, error) {

	data := map[string]string{
		"grant_type":  "client_credentials",
		"id_provider": "client",
	}
	dataByte, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	body := bytes.NewBuffer(dataByte)

	header := make(http.Header)
	header.Set("X-BK-APP-CODE", "bk_cmdb")
	header.Set("X-BK-APP-SECRET", BKAPPSECRET)
	header.Set("Content-Type", "application/json")

	replyData, err := bhttp.Request(GetCCTokenPath, http.MethodPost, header, body)
	if err != nil {
		return "", nil
	}

	resp := new(respGetCCToken)

	err = json.Unmarshal([]byte(replyData), resp)
	if err != nil {
		return "", err
	}
	if resp.Code != 0 {
		return "", nil
	}

	return resp.Data.AccessToken, nil
}

func getAllProject(helper upgrader.UpgradeHelper) ([]ccProject, error) {
	url := fmt.Sprintf(ALLPROJECTPATH, CCTOKEN)

	replyData, err := helper.RequestApiServer(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp := new(respAllProjectData)
	err = json.Unmarshal(replyData, resp)
	if err != nil {
		return nil, err
	}

	return resp.Results, nil
}

func allCluster(helper upgrader.UpgradeHelper) ([]allClusterData, error) {
	url := fmt.Sprintf(ALLCLUSTERPATH, CCTOKEN)

	replyData, err := helper.RequestApiServer(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp := make([]allClusterData, 0)
	err = json.Unmarshal(replyData, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func versionConfig(clusterID string, helper upgrader.UpgradeHelper) (*versionConfigData, error) {

	url := fmt.Sprintf(VERSIONCONFIGPATH, clusterID, CCTOKEN)

	replyData, err := helper.RequestApiServer(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp := new(versionConfigData)
	err = json.Unmarshal(replyData, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func clusterInfo(projectID string, clusterID string, helper upgrader.UpgradeHelper) (*clustersInfoData, error) {

	url := fmt.Sprintf(CLUSTERINFOPATH, projectID, clusterID, CCTOKEN)

	replyData, err := helper.RequestApiServer(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp := new(clustersInfoData)
	err = json.Unmarshal(replyData, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func allNodeList(helper upgrader.UpgradeHelper) ([]nodeListData, error) {

	url := fmt.Sprintf(AllNodeListPath, CCTOKEN)

	replyData, err := helper.RequestApiServer(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp := make([]nodeListData, 0)
	err = json.Unmarshal(replyData, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func allMasterList(helper upgrader.UpgradeHelper) ([]allMasterListData, error) {

	url := fmt.Sprintf(ALLMASTERLISTPATH, CCTOKEN)

	replyData, err := helper.RequestApiServer(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp := make([]allMasterListData, 0)
	err = json.Unmarshal(replyData, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
