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

package pkg

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/viper"
)

// HttpClient http client
type HttpClient struct {
	*http.Client
}

// Get http client get function
func (c *HttpClient) Get(url string, result interface{}) (err error) {
	return c.do(http.MethodGet, url, nil, result)
}

// Post http client post function
func (c *HttpClient) Post(url string, data, result interface{}) (err error) {
	return c.do(http.MethodPost, url, data, result)
}

// Put http client put function
func (c *HttpClient) Put(url string, data, result interface{}) (err error) {
	return c.do(http.MethodPut, url, data, result)
}

// Delete http client delete function
func (c *HttpClient) Delete(url string, result interface{}) (err error) {
	return c.do(http.MethodDelete, url, nil, result)
}

func (c *HttpClient) do(method, url string, data, result interface{}) (err error) {
	var byt []byte
	if data != nil {
		byt, err = json.Marshal(data)
		if err != nil {
			return
		}
	}

	baseURL := viper.GetString("config.apiserver")
	req, err := http.NewRequest(method, "https://"+baseURL+url, bytes.NewBuffer(byt))
	if err != nil {
		return err
	}

	config := &Config{
		APIServer: baseURL,
		AuthToken: viper.GetString("config.bcs_token"),
		Operator:  viper.GetString("config.operator"),
	}

	req.Header.Set("Content-Type", "application/json")
	if len(config.AuthToken) != 0 {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.AuthToken))
	}

	resp, err := c.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error status code: %d", resp.StatusCode)
	}

	if result == nil {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = resp.Body.Close()
	if err != nil {
		return
	}

	return json.Unmarshal(body, result)
}

// NewHttpClientWithConfiguration new http client with config
func NewHttpClientWithConfiguration(ctx context.Context) *HttpClient {
	return &HttpClient{
		&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}
