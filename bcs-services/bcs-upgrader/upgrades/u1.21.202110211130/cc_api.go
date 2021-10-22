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
	"errors"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
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
		return "", errors.New("")
	}

	return resp.Data.AccessToken, nil
}

func getAllProject() ([]ccProject, error) {
	url := fmt.Sprintf(ALLPROJECTPATH, CCTOKEN)

	replyData, err := bhttp.Request(url, http.MethodGet, nil, nil)
	if err != nil {
		blog.Errorf("err: %v", err)
		return nil, err
	}

	resp := new(respAllProject)
	err = json.Unmarshal([]byte(replyData), resp)
	if err != nil {
		blog.Errorf("http request failed, err: %v", err)
		return nil, err
	}
	if !resp.Result {
		blog.Errorf(" failed, err: %v", err)
		return nil, fmt.Errorf(resp.Message)
	}
	if resp.Code != 0 {
		return nil, errors.New(resp.Message)
	}

	return resp.Data.Results, nil
}

func allCluster() ([]allClusterData, error) {
	url := fmt.Sprintf(ALLCLUSTERPATH, CCTOKEN)

	replyData, err := bhttp.Request(url, http.MethodGet, nil, nil)
	if err != nil {
		return nil, err
	}
	resp := new(respAllCluster)
	err = json.Unmarshal([]byte(replyData), resp)
	if err != nil {
		return nil, err
	}
	if !resp.Result {
		return nil, errors.New(resp.Message)
	}
	if resp.Code != 0 {
		return nil, errors.New(resp.Message)
	}

	return resp.Data, nil
}

func versionConfig(clusterID string) (*versionConfigData, error) {

	url := fmt.Sprintf(VERSIONCONFIGPATH, clusterID, CCTOKEN)

	replyData, err := bhttp.Request(url, http.MethodGet, nil, nil)
	if err != nil {
		return nil, err
	}

	resp := new(respVersionConfig)
	err = json.Unmarshal([]byte(replyData), &resp)
	if err != nil {
		return nil, err
	}
	if !resp.Result {
		return nil, errors.New(resp.Message)
	}
	if resp.Code != 0 {
		return nil, errors.New(resp.Message)
	}

	return &resp.Data, nil
}

func clusterInfo(projectID string, clusterID string) (*clustersInfoData, error) {

	url := fmt.Sprintf(CLUSTERINFOPATH, projectID, clusterID, CCTOKEN)

	replyData, err := bhttp.Request(url, http.MethodGet, nil, nil)
	if err != nil {
		return nil, err
	}

	resp := new(respClustersInfo)
	err = json.Unmarshal([]byte(replyData), resp)
	if err != nil {
		return nil, err
	}
	if !resp.Result {
		return nil, errors.New(resp.Message)
	}
	if resp.Code != 0 {
		return nil, errors.New(resp.Message)
	}

	return &resp.Data, nil
}

func allNodeList() ([]nodeListData, error) {

	url := fmt.Sprintf(AllNodeListPath, CCTOKEN)

	replyData, err := bhttp.Request(url, http.MethodGet, nil, nil)
	if err != nil {
		return nil, err
	}

	resp := new(respNodeList)
	err = json.Unmarshal([]byte(replyData), resp)
	if err != nil {
		return nil, err
	}
	if !resp.Result {
		return nil, errors.New(resp.Message)
	}
	if resp.Code != 0 {
		return nil, errors.New(resp.Message)
	}
	return resp.Data, nil
}

func allMasterList() ([]allMasterListData, error) {

	url := fmt.Sprintf(ALLMASTERLISTPATH, CCTOKEN)

	replyData, err := bhttp.Request(url, http.MethodGet, nil, nil)
	if err != nil {
		return nil, err
	}

	resp := new(respAllMasterList)
	err = json.Unmarshal([]byte(replyData), resp)
	if err != nil {
		return nil, err
	}
	if !resp.Result {
		return nil, errors.New(resp.Message)
	}
	if resp.Code != 0 {
		return nil, errors.New(resp.Message)
	}
	return resp.Data, nil
}
