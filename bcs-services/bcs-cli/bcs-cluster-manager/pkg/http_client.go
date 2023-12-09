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
 */

package pkg

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ClusterMgrClient http client
type ClusterMgrClient struct {
	cli    *http.Client
	config *Config
}

// Get http client get function
func (c *ClusterMgrClient) Get(url string, result interface{}) error {
	return c.do(http.MethodGet, url, nil, result)
}

// Post http client post function
func (c *ClusterMgrClient) Post(url string, data, result interface{}) error {
	return c.do(http.MethodPost, url, data, result)
}

// Put http client put function
func (c *ClusterMgrClient) Put(url string, data, result interface{}) error {
	return c.do(http.MethodPut, url, data, result)
}

// Delete http client delete function
func (c *ClusterMgrClient) Delete(url string, result interface{}) error {
	return c.do(http.MethodDelete, url, nil, result)
}

func (c *ClusterMgrClient) do(method, url string, data, result interface{}) error {
	var byt []byte
	var err error
	if data != nil {
		byt, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}

	if result == nil {
		return fmt.Errorf("lost expected response")
	}

	totalURL := c.config.APIServer + url
	req, err := http.NewRequest(method, totalURL, bytes.NewBuffer(byt))
	if err != nil {
		return fmt.Errorf("create http request failed, %s", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	if len(c.config.AuthToken) != 0 {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.AuthToken))
	}

	resp, err := c.cli.Do(req)
	if err != nil {
		return fmt.Errorf("send http request failed, %s", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status code is not expected: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading http response body failed, %s", err.Error())
	}
	defer resp.Body.Close()

	return json.Unmarshal(body, result)
}

// NewClusterMgrClient new http client with config
func NewClusterMgrClient(config *Config) *ClusterMgrClient {
	return &ClusterMgrClient{
		cli: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint
			},
		},
		config: config,
	}
}
