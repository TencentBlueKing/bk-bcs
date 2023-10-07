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

// Package requester xxx
package requester

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/go-resty/resty/v2"
)

// Requester for http request
type Requester interface {
	DoGetRequest(url string, header map[string]string) ([]byte, error)
	DoPostRequest(url string, header map[string]string, data []byte) ([]byte, error)
	DoPutRequest(url string, header map[string]string, data []byte) ([]byte, error)
	DoPatchRequest(url string, header map[string]string, data []byte) ([]byte, error)
	DoDeleteRequest(url string, header map[string]string) ([]byte, error)
}

type requester struct {
	httpCli *resty.Client
}

// NewRequester return interface Requester
func NewRequester() Requester {
	return &requester{
		httpCli: resty.New(),
	}
}

// DoGetRequest do get request
func (r *requester) DoGetRequest(url string, header map[string]string) ([]byte, error) {
	rsp, err := r.httpCli.R().SetHeaders(header).Get(url)
	if err != nil {
		blog.Errorf("do get request error, url: %s, error:%v", url, err)
		return nil, fmt.Errorf("do get request error, url: %s, error:%v", url, err)
	}
	return rsp.Body(), nil
}

// DoPostRequest do post request
func (r *requester) DoPostRequest(url string, header map[string]string, data []byte) ([]byte, error) {
	rsp, err := r.httpCli.R().SetHeaders(header).SetBody(data).Post(url)
	if err != nil {
		blog.Errorf("do post request error, url: %s, error:%v", url, err)
		return nil, fmt.Errorf("do post request error, url: %s, error:%v", url, err)
	}
	return rsp.Body(), nil
}

// DoPutRequest do put request
func (r *requester) DoPutRequest(url string, header map[string]string, data []byte) ([]byte, error) {
	rsp, err := r.httpCli.R().SetHeaders(header).SetBody(data).Put(url)
	if err != nil {
		blog.Errorf("do put request error, url: %s, error:%v", url, err)
		return nil, fmt.Errorf("do put request error, url: %s, error:%v", url, err)
	}
	return rsp.Body(), nil
}

// DoPatchRequest do patch request
func (r *requester) DoPatchRequest(url string, header map[string]string, data []byte) ([]byte, error) {
	rsp, err := r.httpCli.R().SetHeaders(header).SetBody(data).Patch(url)
	if err != nil {
		blog.Errorf("do patch request error, url: %s, error:%v", url, err)
		return nil, fmt.Errorf("do patch request error, url: %s, error:%v", url, err)
	}
	return rsp.Body(), nil
}

// DoDeleteRequest do delete request
func (r *requester) DoDeleteRequest(url string, header map[string]string) ([]byte, error) {
	rsp, err := r.httpCli.R().SetHeaders(header).Delete(url)
	if err != nil {
		blog.Errorf("do delete request error, url: %s, error:%v", url, err)
		return nil, fmt.Errorf("do delete request error, url: %s, error:%v", url, err)
	}
	return rsp.Body(), nil
}
