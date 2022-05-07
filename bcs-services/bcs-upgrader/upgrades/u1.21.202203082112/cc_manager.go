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

package u1x21x202203082112

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
)

// CcManager 调用pass-cc接口
type CcManager interface {
	setToken(bkAppSecret, ssmHost string) error
	getAllProject() ([]ccProject, error)
	getAllCluster() ([]ccGetAllClusterData, error)
	versionConfig(clusterID string) (*ccVersionConfigData, error)
	clusterInfo(projectID, clusterID string) (*ccGetClustersInfoData, error)
	getAllNode() ([]ccGetAllNode, error)
	getAllMaster() ([]ccGetAllMaster, error)
	requestApiServer(method, url string) ([]byte, error)
}

type ccManager struct {
	httpCli *httpclient.HttpClient
	ccHost  string
	Token   string
}

func NewCcManager(ccHost string) CcManager {
	return &ccManager{
		httpCli: httpclient.NewHttpClient(),
		ccHost:  ccHost,
	}
}

func (c *ccManager) setToken(bkAppSecret, ssmHost string) error {
	data := map[string]string{
		"grant_type":  "client_credentials",
		"id_provider": "client",
	}
	dataByte, err := json.Marshal(data)
	if err != nil {
		return err
	}

	cli := httpclient.NewHttpClient()
	cli.SetHeader("Content-Type", "application/json")
	cli.SetHeader("X-BK-APP-CODE", "bk_cmdb")
	cli.SetHeader("X-BK-APP-SECRET", bkAppSecret)

	replyData, err := cli.Request(ssmHost, http.MethodPost, nil, dataByte)
	if err != nil {
		return err
	}
	type respGetCCToken struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}

	resp := new(respGetCCToken)
	err = json.Unmarshal(replyData, resp)
	if err != nil {
		return err
	}
	c.Token = resp.Data.AccessToken

	return nil
}

func (c *ccManager) getAllProject() ([]ccProject, error) {

	url := fmt.Sprintf(ccAllProjectPath, c.Token)
	respData, err := c.requestApiServer(http.MethodGet, url)
	if err != nil {
		return nil, err
	}

	data := new(ccGetAllProject)
	_ = json.Unmarshal(respData, data)

	return data.Results, nil
}

func (c *ccManager) getAllCluster() ([]ccGetAllClusterData, error) {

	url := fmt.Sprintf(ccAllClusterPath, c.Token)
	respData, err := c.requestApiServer(http.MethodGet, url)
	if err != nil {
		return nil, err
	}

	var data []ccGetAllClusterData

	_ = json.Unmarshal(respData, &data)

	return data, nil
}

func (c *ccManager) versionConfig(clusterID string) (*ccVersionConfigData, error) {

	url := fmt.Sprintf(ccVersionConfigPath, clusterID, c.Token)
	replyData, err := c.requestApiServer(http.MethodGet, url)
	if err != nil {
		return nil, err
	}

	resp := new(ccVersionConfigData)
	_ = json.Unmarshal(replyData, resp)

	//val, ok := replyData.(ccVersionConfigData)
	//if !ok {
	//	return nil, fmt.Errorf("")
	//}

	return resp, nil
}

func (c *ccManager) clusterInfo(projectID, clusterID string) (*ccGetClustersInfoData, error) {

	url := fmt.Sprintf(ccClusterInfoPath, projectID, clusterID, c.Token)
	respData, err := c.requestApiServer(http.MethodGet, url)
	if err != nil {
		return nil, err
	}

	resp := new(ccGetClustersInfoData)
	_ = json.Unmarshal(respData, resp)

	return resp, nil
}

func (c *ccManager) getAllNode() ([]ccGetAllNode, error) {

	url := fmt.Sprintf(ccAllNodeListPath, c.Token)
	respData, err := c.requestApiServer(http.MethodGet, url)
	if err != nil {
		return nil, err
	}

	var resp []ccGetAllNode

	_ = json.Unmarshal(respData, &resp)

	return resp, nil
}

func (c *ccManager) getAllMaster() ([]ccGetAllMaster, error) {

	url := fmt.Sprintf(ccAllMasterListPath, c.Token)
	respData, err := c.requestApiServer(http.MethodGet, url)
	if err != nil {
		return nil, err
	}

	var resp []ccGetAllMaster
	_ = json.Unmarshal(respData, &resp)

	return resp, nil
}

func (c *ccManager) requestApiServer(method, url string) ([]byte, error) {
	url = c.ccHost + url
	replyData, err := c.httpCli.Request(url, method, nil, nil)
	if err != nil {
		return nil, err
	}

	reply := new(ccResp)
	err = json.Unmarshal(replyData, reply)
	if err != nil {
		return nil, fmt.Errorf("unmarshal body(%s) failed: %s", string(replyData), err.Error())
	}

	if !reply.Result {
		return nil, fmt.Errorf("request %s failed: %s", url, reply.Message)
	}

	resp, _ := json.Marshal(reply.Data)

	return resp, nil

}
