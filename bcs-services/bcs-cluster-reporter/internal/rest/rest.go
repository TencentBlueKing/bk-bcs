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

// Package rest xxx
package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// HTTPClient  xxx
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Request xxx
type Request struct {
	client HTTPClient
	url    *url.URL
	verb   string
	body   io.Reader
}

// NewRequest  xxx
func NewRequest(client HTTPClient, verb string, url *url.URL, data io.Reader) *Request {
	if client == nil {
		client = http.DefaultClient
	}
	return &Request{
		client: client,
		verb:   verb,
		url:    url,
		body:   data,
	}
}

// Do xxx
func (r *Request) Do() ([]byte, error) {
	client := r.client

	req, err := http.NewRequest(r.verb, r.url.String(), r.body)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var body []byte
	if resp.Body != nil {
		data, err := io.ReadAll(resp.Body)
		body = data
		if err != nil {
			return nil, err
		}
	}

	if strings.Contains(resp.Status, "200") && strings.Contains(resp.Status, "201") {
		return body, fmt.Errorf("get %s status from %s", resp.Status, r.url)
	}

	return body, nil
}

// DoWithMarshal xxx
func (r *Request) DoWithMarshal(result interface{}) error {
	data, err := r.Do()
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, result)
	if err != nil {
		return err
	}

	return nil
}

// BaseResponse xxx
type BaseResponse struct {
	Msg    string      `json:"message"`
	Result bool        `json:"result"`
	Code   interface{} `json:"code"`
	Data   interface{} `json:"data"`
}
