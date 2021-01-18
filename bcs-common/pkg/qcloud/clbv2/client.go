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
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/google/go-querystring/query"
)

//Client basic signature https client
type Client struct {
	URL       string //https url info
	SecretKey string //secret key for api signature
}

//NewClient constructor function
func NewClient(url, secretKey string) *Client {
	return &Client{
		URL:       url,
		SecretKey: secretKey,
	}
}

//GetRequest create get request to qcloud api service
//return:
//	[]byte: http response body
//	error: err info if do
func (q *Client) GetRequest(obj interface{}) ([]byte, error) {
	//signature with all detail request data
	signature, err := Signature(q.SecretKey, "GET", strings.Replace(q.URL, "https://", "", -1), obj)
	if err != nil {
		blog.Errorf("Qcloud signature failed, %s", err)
		return nil, err
	}
	//setting signature for request
	v := reflect.ValueOf(obj).Elem()
	v.FieldByName("Signature").SetString(signature)
	//desc.Signature = signature
	value, _ := query.Values(obj)
	request := value.Encode()
	link := fmt.Sprintf("%s?%s", q.URL, request)
	blog.Infof("GET details %s", link)
	response, err := http.Get(link)
	if err != nil {
		blog.Errorf("GET request failed, %s", err)
		return nil, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		blog.Errorf("Reading response body failed, %s", err)
		return nil, err
	}
	blog.Infof("GET request response: %s", string(body))
	return body, nil
}

//PostRequest create get request to vpc service
//return:
//	[]byte: http response body
//	error: err info if do
func (q *Client) PostRequest(obj interface{}) ([]byte, error) {
	//signature with all detail request data
	signature, err := Signature(q.SecretKey, "POST", strings.Replace(q.URL, "https://", "", -1), obj)
	if err != nil {
		blog.Errorf("Qcloud signature failed, %s", err)
		return nil, err
	}
	//setting signature for request
	v := reflect.ValueOf(obj).Elem()
	v.FieldByName("Signature").SetString(signature)
	//desc.Signature = signature
	value, _ := query.Values(obj)
	request := value.Encode()
	//blog.Infof("Post detail: %s", request)
	response, err := http.Post(q.URL, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(request)))
	if err != nil {
		blog.Errorf("Post request failed, %s", err)
		return nil, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		blog.Errorf("Reading response body failed, %s", err)
		return nil, err
	}
	blog.Infof("Post request response: %s", string(body))
	return body, nil
}
