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

package manager

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
)

type HttpClient struct {
	header  map[string]string
	httpCli *http.Client
}

func NewHttpClient() *HttpClient {
	return &HttpClient{
		httpCli: &http.Client{},
		header:  make(map[string]string),
	}
}

func (client *HttpClient) SetTimeOut(timeOut time.Duration) {
	client.httpCli.Timeout = timeOut
}

func (client *HttpClient) SetHeader(key, value string) {
	client.header[key] = value
}

func (client *HttpClient) GET(url string, header http.Header, data []byte) (int, []byte, error) {
	return client.Request(url, "GET", header, data)
}

func (client *HttpClient) POST(url string, header http.Header, data []byte) (int, []byte, error) {
	return client.Request(url, "POST", header, data)
}

func (client *HttpClient) DELETE(url string, header http.Header, data []byte) (int, []byte, error) {
	return client.Request(url, "DELETE", header, data)
}

func (client *HttpClient) PUT(url string, header http.Header, data []byte) (int, []byte, error) {
	return client.Request(url, "PUT", header, data)
}

func (client *HttpClient) Request(url, method string, header http.Header, data []byte) (int, []byte, error) {
	var req *http.Request
	var errReq error
	if data != nil {
		req, errReq = http.NewRequest(method, url, bytes.NewReader(data))
	} else {
		req, errReq = http.NewRequest(method, url, nil)
	}

	if errReq != nil {
		return 0, nil, errReq
	}

	req.Close = true

	if header != nil {
		req.Header = header
	}

	for key, value := range client.header {
		req.Header.Set(key, value)
	}

	rsp, err := client.httpCli.Do(req)
	if err != nil {
		return 0, nil, err
	}

	defer func() {
		err1 := rsp.Body.Close()
		if err1 != nil {
			blog.Errorf("close http response body error %s", err1.Error())
		}
	}()

	by, err := ioutil.ReadAll(rsp.Body)

	return rsp.StatusCode, by, err
}
