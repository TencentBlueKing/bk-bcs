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

package driver

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	protoExec "github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-executor/process-executor/protobuf/executor"

	"github.com/golang/protobuf/proto"
)

type HttpConnection struct {
	endpoint string //remote http endpoint info
	uri      string //remote http endpoint uri
	streamID string //http header Mesos-Stream-Id

	client *http.Client
}

func NewHttpConnection(endpoint, uri string) *HttpConnection {
	httpTransport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		ResponseHeaderTimeout: 5 * time.Second,
	}

	return &HttpConnection{
		endpoint: endpoint,
		uri:      uri,
		client: &http.Client{
			Transport: httpTransport,
		},
	}
}

func (conn *HttpConnection) Send(call *protoExec.Call, keepAlive bool) (*http.Response, error) {
	//create targetURL
	targetURL := fmt.Sprintf("%s%s", conn.endpoint, conn.uri)
	// proto serialization
	payLoad, err := proto.Marshal(call)
	if err != nil {
		blog.Errorf("proto.Marshal call error %s", err.Error())
		return nil, err
	}

	//create http request
	request, err := http.NewRequest("POST", targetURL, bytes.NewReader(payLoad))
	if err != nil {
		blog.Errorf("NewRequest url %s error %s", targetURL, err.Error())
		return nil, err
	}

	request.Header.Set("Content-Type", "application/x-protobuf")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "bcs-container-executor/1.0")
	request.Header.Set("Connection", "Keep-Alive")

	response, err := conn.client.Do(request)
	if err != nil {
		blog.Errorf("conn request uri %s error %s", targetURL, err.Error())
		return nil, err
	}

	if response.StatusCode != http.StatusAccepted && response.StatusCode != http.StatusOK {
		reply, _ := ioutil.ReadAll(response.Body)
		err = fmt.Errorf("request url %s response statuscode %d body %s", targetURL, response.StatusCode, string(reply))
		blog.Errorf(err.Error())
		return nil, err
	}

	if keepAlive {
		return response, nil
	}

	response.Body.Close()
	return nil, nil
}
