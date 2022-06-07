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
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
)

// CmManager 调用cm接口
type CmManager interface {
	createProject(project cmCreateProject) error
	findProject(projectID string) (*cmGetProject, error)
	updateProject(project cmUpdateProject) error
	createClusters(clusters cmCreateCluster) error
	findCluster(clustersID string) (*bcsRespFindCluster, error)
	updateCluster(data bcsReqUpdateCluster) error
	createNode(data cmCreateNode) error
	deleteNode(data reqDeleteNode) error
	findClusterNode(clustersID string) ([]bcsNodeListData, error)
	requestApiServer(method, url string, payload []byte) ([]byte, error)
}

type cmManager struct {
	httpCli      *httpclient.HttpClient
	cmHost       string
	gatewayToken string
}

func NewCmManager(cmHost, gatewayToken string) CmManager {

	httpCli := httpclient.NewHttpClient()

	httpCli.SetHeader("Content-Type", "application/json")
	httpCli.SetHeader("Authorization", "Bearer "+gatewayToken)

	httpCli.SetTlsNoVerity()

	return &cmManager{
		httpCli: httpCli,
		cmHost:  cmHost,
	}
}

func (c *cmManager) createProject(project cmCreateProject) error {
	projectByte, err := json.Marshal(project)
	if err != nil {
		return err
	}
	_, err = c.requestApiServer(http.MethodPost, cmCreateProjectPath, projectByte)

	return err
}

func (c *cmManager) findProject(projectID string) (*cmGetProject, error) {
	url := fmt.Sprintf(cmProjectPath, projectID)
	replyData, err := c.requestApiServer(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp := new(cmGetProject)
	err = json.Unmarshal(replyData, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *cmManager) updateProject(project cmUpdateProject) error {
	projectJson, err := json.Marshal(project)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(cmProjectPath, project.ProjectID)
	_, err = c.requestApiServer(http.MethodPut, url, projectJson)
	return err
}

func (c *cmManager) createClusters(clusters cmCreateCluster) error {
	clustersJson, err := json.Marshal(clusters)
	if err != nil {
		return err
	}
	_, err = c.requestApiServer(http.MethodPost, cmCreateClusterPath, clustersJson)
	return err
}

func (c *cmManager) findCluster(clustersID string) (*bcsRespFindCluster, error) {

	url := fmt.Sprintf(cmClusterHost, clustersID)
	replyData, err := c.requestApiServer(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp := new(bcsRespFindCluster)
	err = json.Unmarshal(replyData, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *cmManager) updateCluster(data bcsReqUpdateCluster) error {
	url := fmt.Sprintf(cmClusterHost, data.ClusterID)

	dataJson, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = c.requestApiServer(http.MethodPut, url, dataJson)
	return err
}

func (c *cmManager) createNode(data cmCreateNode) error {
	nodeJson, err := json.Marshal(data)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(cmNodeHost, data.ClusterID)

	_, err = c.requestApiServer(http.MethodPost, url, nodeJson)
	return err
}

func (c *cmManager) deleteNode(data reqDeleteNode) error {
	dataJson, err := json.Marshal(data)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(cmNodeHost, data.ClusterID)

	_, err = c.requestApiServer(http.MethodDelete, url, dataJson)
	return err
}

func (c *cmManager) findClusterNode(clustersID string) ([]bcsNodeListData, error) {

	url := fmt.Sprintf(cmNodeHost, clustersID)

	replyData, err := c.requestApiServer(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp := make([]bcsNodeListData, 0)
	err = json.Unmarshal(replyData, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

//method=http.method: POST、GET、PUT、DELETE
//request url = address/url
//payload is request body
//if error!=nil, then request mesos failed, errom.Error() is failed message
//if error==nil, []byte is response body information
func (c *cmManager) requestApiServer(method, url string, payload []byte) ([]byte, error) {

	var err error
	var by []byte
	url = c.cmHost + url

	switch method {
	case "GET":
		by, err = c.httpCli.GET(url, nil, payload)
	case "POST":
		by, err = c.httpCli.POST(url, nil, payload)
	case "DELETE":
		by, err = c.httpCli.DELETE(url, nil, payload)
	case "PUT":
		by, err = c.httpCli.PUT(url, nil, payload)
	default:
		err = fmt.Errorf("uri %s method %s is invalid", url, method)
	}
	if err != nil {
		return nil, err
	}

	//unmarshal response.body
	var result *commtypes.APIResponse
	err = json.Unmarshal(by, &result)
	if err != nil {
		return nil, fmt.Errorf("unmarshal body(%s) failed: %s", string(by), err.Error())
	}
	//if result.Result==false, then request failed
	if !result.Result {
		return nil, fmt.Errorf("request %s failed: %s", url, result.Message)
	}
	by, _ = json.Marshal(result.Data)
	return by, nil
}
