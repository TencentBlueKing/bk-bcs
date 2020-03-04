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

package esb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"bk-bcs/bcs-common/common/blog"
)

const (
	EsbRequestPayloadAppcode   = "app_code"
	EsbRequestPayloadAppsecret = "app_secret"
	EsbRequestPayloadOperator  = "operator"
)

type APIResponse struct {
	Result  bool        `json:"result"`
	Code    string      `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

type EsbClient struct {
	//esb app code
	AppCode string
	//esb app secret
	AppSecret string
	//esb app operator
	AppOperator string
	//esb url
	EsbUrl string
	//request esb
}

func NewEsbClient(appCode, appSecret, appOperator, esbUrl string) (*EsbClient, error) {
	esb := &EsbClient{
		EsbUrl: esbUrl,
	}

	//Decrypt app parameters
	/*var err error
	esb.AppCode, err = encrypt.DesDecryptFromBase([]byte(appCode))
	if err != nil {
		return nil, fmt.Errorf("decrypt appCode %s failed: %s", appCode, err.Error())
	}
	esb.AppSecret, err = encrypt.DesDecryptFromBase([]byte(appSecret))
	if err != nil {
		return nil, fmt.Errorf("decrypt appSecret %s failed: %s", appSecret, err.Error())
	}
	esb.AppOperator, err = encrypt.DesDecryptFromBase([]byte(appOperator))
	if err != nil {
		return nil, fmt.Errorf("decrypt Operator %s fialed: %s", appOperator, err.Error())
	}*/
	esb.AppCode = appCode
	esb.AppSecret = appSecret
	esb.AppOperator = appOperator

	return esb, nil
}

//method=http.method: POST、GET、PUT、DELETE
//request url = esb.EsbUrl/url
//payload is request body
//if error!=nil, then request esb failed, error.Error() is failed message
//if error==nil, []byte is response body information
func (esb *EsbClient) RequestEsb(method, url string, payload map[string]interface{}) ([]byte, error) {
	if payload == nil {
		return nil, fmt.Errorf("payload can't be nil")
	}
	//set payload app parameter
	payload[EsbRequestPayloadAppcode] = esb.AppCode
	payload[EsbRequestPayloadAppsecret] = esb.AppSecret
	payload[EsbRequestPayloadOperator] = esb.AppOperator
	payloadBytes, _ := json.Marshal(payload)
	//new request body
	body := bytes.NewBuffer(payloadBytes)
	//request url
	url = fmt.Sprintf("%s%s", esb.EsbUrl, url)

	//new request object
	req, _ := http.NewRequest(method, url, body)
	//set header application/json
	req.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request esb %s failed: %s", url, err.Error())
	}
	defer resp.Body.Close()

	// Parse body as JSON
	var result APIResponse
	respBody, _ := ioutil.ReadAll(resp.Body)
	blog.V(3).Infof("request esb %s resp body(%s)", url, string(respBody))

	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, fmt.Errorf("non-Json body(%s) response: %s", string(respBody), err.Error())
	}

	//http response status code != 200
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("response code %d body %s", resp.StatusCode, respBody)
	}
	//esb response result failed
	if !result.Result {
		return nil, fmt.Errorf("request esb %s failed, code:%s message:%s", url, result.Code, result.Message)
	}
	//marshal result.data to []byte
	by, _ := json.Marshal(result.Data)
	return by, nil
}
