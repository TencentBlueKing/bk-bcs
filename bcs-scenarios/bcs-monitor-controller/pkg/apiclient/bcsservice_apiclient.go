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

package apiclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

const (
	APIPathListCluster = "clustermanager/v1/cluster"
)

// ServiceBaseResp base response from service
type ServiceBaseResp struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Result  bool   `json:"result,omitempty"`
}

// ClusterInfo  集群信息
type ClusterInfo struct {
	ClusterID   string `json:"clusterID,omitempty"`
	ClusterName string `json:"clusterName,omitempty"`
	Region      string `json:"region,omitempty"`
	ProjectID   string `json:"projectID,omitempty"`
	BusinessID  string `json:"businessID,omitempty"`
	Environment string `json:"environment,omitempty"`
	EngineType  string `json:"engineType,omitempty"`
	Creator     string `json:"creator,omitempty"`
}

// ListClusterResp resp of list cluster
type ListClusterResp struct {
	ServiceBaseResp
	Data []ClusterInfo `json:"data"`
}

// BcsServiceApiClient api client to call bcs service
type BcsServiceApiClient struct {
	Token     string
	APIDomain string
}

// ListCluster get cluster info from clusterManager
func (c *BcsServiceApiClient) ListCluster() {
	headers := http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", c.Token)},
	}
	params := url.Values{
		"kind": {"k8s"},
	}
	res, err := c.doRequest(fmt.Sprintf("%s/%s", c.APIDomain, APIPathListCluster), params, headers)
	if err != nil {
		return
	}

	var listClusterResp ListClusterResp
	if inErr := json.Unmarshal(res, &listClusterResp); inErr != nil {
		blog.Errorf("unmarshal resp from clusterManager failed. raw result: %s, err: %s", string(res), inErr.Error())
		return
	}
	log.Printf("%d", len(listClusterResp.Data))

	log.Printf(string(res))
}

func (c *BcsServiceApiClient) doRequest(urlStr string, params url.Values,
	headers http.Header) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()
	req.Header = headers

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		blog.Errorf("read bcs service resp failed, err: %s", err.Error())
		return nil, err
	}
	return respBody, err
}
