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

package u1_21_202110211130

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common"
)

// XRequest http request encapsulation
func XRequest(url, method string, header http.Header, body io.Reader) (string, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		reply := InternalError(common.BcsErrCommHttpNewRequest, common.BcsErrCommHttpNewRequestStr)
		return reply.Error(), fmt.Errorf("fail to new a http request. err:%s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	req.Close = true

	//header
	if header != nil {
		req.Header = header
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
	}

	rsp, err := client.Do(req)
	if err != nil {
		reply := InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr)
		return reply.Error(), fmt.Errorf("fail to do http request. err:%s", err.Error())
	}

	defer rsp.Body.Close()

	replyData, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		reply := InternalError(common.BcsErrCommHttpReadRsp, common.BcsErrCommHttpReadRspStr)
		return reply.Error(), fmt.Errorf("read response failed. err:%s", err.Error())
	}

	return string(replyData), nil
}

// InternalError internal error type exchange
func InternalError(code int, message string) error {

	_, err := createRespone(code, message, make(map[string]interface{}))

	return err
}

func createRespone(code int, message string, data interface{}) (string, error) {

	b, err := createResponeEx(code, message, data)

	return string(b), err
}

func createResponeEx(code int, message string, data interface{}) ([]byte, error) {
	bResult := false
	if 0 == code {
		bResult = true
	} else {
		appName := os.Args[0]
		szArr := strings.Split(appName, "/")
		if len(szArr) >= 2 {
			appName = szArr[1]
		}
		message = "(" + appName + "):" + message
	}

	resp := APIRespone{bResult, code, message, data}
	b, err := json.Marshal(resp)
	if err != nil {
		return []byte(""), err
	}

	return b, errors.New(string(b))
}

// APIResponse response for api request
type APIRespone struct {
	Result  bool        `json:"result"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
