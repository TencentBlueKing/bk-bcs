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

package qcloud

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"

	"github.com/google/go-querystring/query"
)

//Response common api response from qcloud
type Response struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	CodeDesc string `json:"codeDesc"`
}

//TaskResponse response task id with qcloud task
type TaskResponse struct {
	Response `json:",inline"`
	Data     TaskData `json:"data"`
}

//TaskData data holder for TaskResponse
type TaskData struct {
	Output interface{} `json:"output,omitempty"`
	Status int         `json:"status,omitempty"`
	TaskID int         `json:"taskId,omitempty"`
}

//APIMeta base data structure for qcloud API
//always use HmacSHA1 method creating signature
type APIMeta struct {
	Action    string `url:"Action"`
	Nonce     uint   `url:"Nonce"`
	Region    string `url:"Region"`
	SecretID  string `url:"SecretId"`
	Signature string `url:"Signature,omitempty"` //method is HmacSHA1
	Timestamp uint   `url:"Timestamp"`
}

//TaskRequest request for task status
type TaskRequest struct {
	APIMeta `url:",inline"`
	TaskID  int `url:"taskId"`
}

//GroupList id list
type GroupList []string

//EncodeValues interface for url encoding
func (l GroupList) EncodeValues(key string, urlv *url.Values) error {
	for i, v := range l {
		k := fmt.Sprintf("%s.%d", key, i)
		urlv.Set(k, v)
	}
	return nil
}

//Signature create signature of request data
//param method: http method, GET or POST
//param url: qcloud request url
//param obj: object to encode,
//example data before hamcSHA1 : "GETcvm.api.qcloud.com/v2/index.php?Action=DescribeInstances&InstanceIds.0=ins-09dx96dg&Nonce=11886&Region=ap-guangzhou&SecretId=xxxxxxxxxx&SignatureMethod=HmacSHA1&Timestamp=1465185768"
func Signature(key, method, url string, obj interface{}) (string, error) {
	if obj == nil {
		return "", fmt.Errorf("Can not signature nil object")
	}
	//1. encode url param
	value, err := query.Values(obj)
	if err != nil {
		return "", fmt.Errorf("construct query.Values from %v failed, err %s", obj, err)
	}
	v := BCSValues{
		val: value,
	}
	str := v.Encode()
	if len(str) == 0 {
		return "", fmt.Errorf("empty data from object")
	}
	//2. construct data
	data := method + url + "?" + str
	//3. hmac
	mac := hmac.New(sha1.New, []byte(key))
	if _, err := mac.Write([]byte(data)); err != nil {
		return "", err
	}
	//4. base64
	sigBytes := mac.Sum(nil)
	sigStr := base64.StdEncoding.EncodeToString(sigBytes)
	return sigStr, nil
}

//BCSValues bcs values encode without value encode
type BCSValues struct {
	val url.Values
}

//Encode encodes the values into "URL encoded" form
//("bar=baz&foo=quux") sorted by key.
func (v BCSValues) Encode() string {
	if v.val == nil {
		return ""
	}
	var buf bytes.Buffer
	keys := make([]string, 0, len(v.val))
	for k := range v.val {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v.val[k]
		prefix := url.QueryEscape(k) + "="
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(prefix)
			buf.WriteString(v)
		}
	}
	return buf.String()
}
