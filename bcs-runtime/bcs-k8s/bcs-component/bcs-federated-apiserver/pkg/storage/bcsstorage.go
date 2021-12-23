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

package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"k8s.io/klog/v2"
)

// BcsStorage is the interface for storage
type BcsStorage struct {
	Address   string
	Token     string
	URLPrefix string
	//memberClusters string
}

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

// ResponseDataList is the response data list
type ResponseDataList []ResponseData

// ResponseData struct store the bcs-storage's resource message.
type ResponseData struct {
	Data         json.RawMessage `json:"data,omitempty"`
	UpdateTime   string          `json:"updateTime"`
	Id           string          `json:"_id"`
	ClusterId    string          `json:"clusterId"`
	Namespace    string          `json:"namespace"`
	ResourceName string          `json:"resourceName"`
	ResourceType string          `json:"resourceType"`
	CreateTime   string          `json:"createTime"`
}

// NewBcsStorage creates a new BcsStorage
func NewBcsStorage(address, token, urlPrefix string) *BcsStorage {
	return &BcsStorage{
		Address:   address,
		Token:     token,
		URLPrefix: urlPrefix,
	}
}

// ListResources lists resources
func (bcsStorage *BcsStorage) ListResources(memberClusters, namespace, name, resourceType string, limit, offset int64) ([]ResponseData, error) {
	if memberClusters == "" {
		return nil, fmt.Errorf("memberClusters is empty")
	}
	url := fmt.Sprintf("%s/%s/%s?clusterId=%s", strings.TrimSuffix(bcsStorage.Address, "/"), strings.TrimSuffix(bcsStorage.URLPrefix, "/"), resourceType, memberClusters)
	if namespace != "" {
		url += fmt.Sprintf("&namespace=%s", namespace)
	}
	if name != "" {
		url += fmt.Sprintf("&resourceName=%s", name)
	}
	if limit != 0 {
		url += fmt.Sprintf("&limit=%d", limit)
	}
	if offset != 0 {
		url += fmt.Sprintf("&offset=%d", offset)
	}
	klog.V(4).InfoS("list resource", "url", url)
	//TODO 分页以及labelSelector
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		klog.Errorf("create get request error: %v", err)
		return nil, fmt.Errorf("NewRequest error, %+v\n", err)
	}

	if bcsStorage.Token != "" {
		var bearer = "Bearer " + string(bcsStorage.Token)
		request.Header.Add("Authorization", bearer)
	}

	request.Header.Set("Content-type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		klog.Errorf("get request error: %v", err)
		return nil, err
	}
	//decode
	return bcsStorage.decodeResponse(response)
}

func (bcsStorage *BcsStorage) decodeResponse(response *http.Response) ([]ResponseData, error) {
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		klog.Errorf("http storage get failed, code: %d, message: %s\n", response.StatusCode, response.Status)
		return nil, fmt.Errorf("remote err, code: %d, status: %s", response.StatusCode, response.Status)
	}
	rawData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		klog.Errorf("http storage get http status success, but read response body failed, %s\n", err)
		return nil, err
	}
	defer response.Body.Close()

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
