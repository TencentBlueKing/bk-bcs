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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-upgrader/upgrader"
)

func getAllProject(helper upgrader.UpgradeHelper) ([]ccProject, error) {

	replyData, err := helper.HttpRequest(http.MethodGet, ALLPROJECTPATH, nil)
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

	replyData, err := helper.HttpRequest(http.MethodGet, ALLCLUSTERPATH, nil)
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

	url := fmt.Sprintf(VERSIONCONFIGPATH, clusterID, "%s")

	replyData, err := helper.HttpRequest(http.MethodGet, url, nil)
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

func clusterInfo(projectID, clusterID string, helper upgrader.UpgradeHelper) (*clustersInfoData, error) {

	url := fmt.Sprintf(CLUSTERINFOPATH, projectID, clusterID, "%s")

	replyData, err := helper.HttpRequest(http.MethodGet, url, nil)
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

	replyData, err := helper.HttpRequest(http.MethodGet, AllNodeListPath, nil)
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

	replyData, err := helper.HttpRequest(http.MethodGet, ALLMASTERLISTPATH, nil)
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
