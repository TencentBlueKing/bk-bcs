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

package app

import (
	"bytes"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestHttpConnection(t *testing.T) {
	httpcli := httpclient.NewHttpClient()
	httpcli.SetTlsNoVerity()
	httpcli.SetTimeOut(time.Duration(60) * time.Second)
	client := httpcli.GetClient()

	var req *http.Request
	var errReq error

	for {
		select {
		case <-time.After(30 * time.Second):
			fmt.Printf("exit test")
			return
		case <-time.After(10 * time.Second):

			req, errReq = http.NewRequest("POST", "http://127.0.0.1:9090/api/v1/do/data/1430/export", bytes.NewReader([]byte("hello world")))

			if errReq != nil {
				fmt.Printf("\nrequest error:%s\n", errReq.Error())
				return
			}
			rsp, err := client.Do(req)
			if err != nil {
				fmt.Printf("failed to do request")
				return
			}
			rpy, err := ioutil.ReadAll(rsp.Body)
			if err != nil {
				fmt.Printf("\nread data failed\n")
				return
			}

			fmt.Printf("\nrpy: %s\n", string(rpy))
		}
	}

	//httpcli.Post("http://127.0.0.1:9090/api/v1/do/data/1430/export", nil, []byte("hello world"))
	//select {}
}
