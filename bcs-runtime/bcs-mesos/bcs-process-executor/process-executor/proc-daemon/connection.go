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

package proc_daemon

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"encoding/json"
	"github.com/Tencent/bk-bcs/bcs-common/common"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	"strings"
)

const (
	DefaultHttpDomain = "http://xxxxxxx"
)

type HttpConnection struct {
	endpoint string //remote http endpoint info
	client   *http.Client
}

func NewHttpConnection(endpoint string) *HttpConnection {
	httpTransport := &http.Transport{
		Dial: func(proto, addr string) (conn net.Conn, err error) {
			return net.Dial("unix", endpoint)
		},
		ResponseHeaderTimeout: 5 * time.Second,
	}

	return &HttpConnection{
		endpoint: endpoint,
		client: &http.Client{
			Transport: httpTransport,
		},
	}
}

func (cli *HttpConnection) requestProcessDaemon(method, uri string, data []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/bcsapi/v1/processdaemon%s", DefaultHttpDomain, uri)

	req, err := http.NewRequest(method, url, strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest method %s url %s error %s", method, url, err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := cli.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http client request method %s url %s error %s", method, url, err.Error())
	}
	defer resp.Body.Close()

	by, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("request url %s read Body error %s", url, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request url %s resp statusCode %d body %s", url, resp.StatusCode, string(by))
	}

	var api *bhttp.APIRespone
	err = json.Unmarshal(by, &api)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal data %s to bhttp.APIRespone error %s", string(by), err.Error())
	}

	if api.Code != common.BcsSuccess {
		return nil, fmt.Errorf("request url %s resp code %d message %s", url, api.Code, api.Message)
	}

	by, err = json.Marshal(api.Data)
	if err != nil {
		return nil, fmt.Errorf("request url %s marshal response data error %s", url, err.Error())
	}

	return by, nil
}
