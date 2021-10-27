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

package bcs_storage

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/klog"
	"net/http"
)

// Response struct store the bcs-storage's response message.
type Response struct {
	Result   bool            `json:"result"`
	Code     int             `json:"code"`           //operation code
	Message  string          `json:"message"`        //response message
	Data     json.RawMessage `json:"data,omitempty"` //response data
	Total    int32           `json:"total"`
	PageSize int32           `json:"pageSize"`
	Offset   int32           `json:"offset"`
}

type ResponseDataList []ResponseData

// ResponseData struct store the bcs-storage's resource message.
type ResponseData struct {
	Data         json.RawMessage `json:"data,omitempty"`
	UpdateTime   string          `json:"updateTime"`
	Id           string          `json:"_id"`
	ClusterId    string          `json:"clusterId""`
	Namespace    string          `json:"namespace"`
	ResourceName string          `json:"resourceName"`
	ResourceType string          `json:"resourceType"`
	CreateTime   string          `json:"createTime"`
}

// DoBcsStorageGetRequest function implements the bcs-storage request,
// which the inputs are fullPath, token(base64 needed), and contentType.
func DoBcsStorageGetRequest(fullPath string, tokenBase64 string, contentType string) (response *http.Response,
	err error) {
	if fullPath == "" {
		klog.Errorf("Http path is nil, please check again.\n")
		return nil, fmt.Errorf("Http path is nil, please check again.\n")
	}

	client := &http.Client{}
	request, err := http.NewRequest("GET", fullPath, nil)
	if err != nil {
		klog.Errorf("Get func NewRequest failed, %s\n", err)
		return nil, fmt.Errorf("Get func NewRequest failed, %s\n", err)
	}

	if tokenBase64 != "" {
		token, err := base64.StdEncoding.DecodeString(tokenBase64)
		if err != nil {
			return nil, err
		}
		var bearer = "Bearer " + string(token)
		request.Header.Add("Authorization", bearer)
	}

	request.Header.Set("Content-type", contentType)

	response, err = client.Do(request)
	if err != nil {
		klog.Errorf("Get func client.Do failed, %s\n", err)
		return nil, fmt.Errorf("Get func client.Do failed, %s\n", err)
	}
	return response, err
}

// DecodeResp function implements the convert of bcs-storage's http responds to ResponseData
func DecodeResp(response http.Response) ([]ResponseData, error) {

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		klog.Errorf("http storage get failed, code: %d, message: %s\n", response.StatusCode, response.Status)
		return nil, fmt.Errorf("remote err, code: %d, status: %s", response.StatusCode, response.Status)
	}
	rawData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		klog.Errorf("http storage get http status success, but read response body failed, %s\n", err)
		return nil, err
	}

	//format http response
	standardResponse := &Response{}
	if err := json.Unmarshal(rawData, standardResponse); err != nil {
		klog.Errorf("http storage decode GET %s http response failed, %s\n", "standarResponse", err)
		return nil, err
	}
	if standardResponse.Code != 0 {
		klog.Errorf("http storage GET failed, %s\n", standardResponse.Message)
		return nil, fmt.Errorf("remote err: %s", standardResponse.Message)
	}
	if len(standardResponse.Data) == 0 {
		klog.Errorln("http storage GET success, but got no data")
		return nil, fmt.Errorf("Previous data err.\n ")
	}

	var responseData []ResponseData
	if err := json.Unmarshal(standardResponse.Data, &responseData); err != nil {
		klog.Errorf("http storage decode data object %s failed, %s\n", "responsedata", err)
		return nil, fmt.Errorf("json decode: %s", err)
	}
	return responseData, err
}
