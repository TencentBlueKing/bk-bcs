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

package upgrader

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
)

// UpgradeHelper is a helper for upgrade
type UpgradeHelper interface {
	// HelperName return the name of the helper
	HelperName() string
	RequestApiServer(method, url string, payload []byte) ([]byte, error)

	SetHeaderClusterManager() *Helper
	SetTlsConfClusterManager() *Helper
}

// Helper is an implementation for interface UpgradeHelper
type Helper struct {
	DB drivers.DB
	// TODO: 添加clusterManagerClient调用bcs-cluster-manager提供的接口，添加httpClient去调用bcs-saas cc模块提供的接口

	// http 调用bcs-saas cc模块提供的接口
	httpClient *httpclient.HttpClient

	//
	clusterManagerClient *httpclient.HttpClient
}

// HelperOpt is option for Helper
type HelperOpt struct {
	DB drivers.DB
}

// Name is the method of Helper to implement interface UpgradeHelper
func (h *Helper) HelperName() string {
	return "bcs-upgrade-helper"
}

// NewUpgradeHelper new a Helper instance
func NewUpgradeHelper(opt *HelperOpt) *Helper {

	return &Helper{
		DB:         opt.DB,
		httpClient: httpclient.NewHttpClient(),
	}
}

func (h *Helper) SetHeaderClusterManager() *Helper {
	h.httpClient.SetHeader("Content-Type", "application/json")
	h.httpClient.SetHeader("Authorization", "Bearer g8Y9wYrT97kERysMDjMy1Gvdq3nI6Tid")

	return h
}

func (h *Helper) SetTlsConfClusterManager() *Helper {

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
	}

	h.httpClient.SetTlsVerityConfig(tlsConf)

	return h
}

//RequestMesosApiServer : RequestMesosApiServer
//method=http.method: POST、GET、PUT、DELETE
//request url = address/url
//payload is request body
//if error!=nil, then request mesos failed, errom.Error() is failed message
//if error==nil, []byte is response body information
func (h *Helper) RequestApiServer(method, url string, payload []byte) ([]byte, error) {

	var err error
	var by []byte

	switch method {
	case "GET":
		by, err = h.httpClient.GET(url, nil, payload)
	case "POST":
		by, err = h.httpClient.POST(url, nil, payload)
	case "DELETE":
		by, err = h.httpClient.DELETE(url, nil, payload)
	case "PUT":
		by, err = h.httpClient.PUT(url, nil, payload)
	default:
		err = fmt.Errorf("uri %s method %s is invalid", url, method)
	}
	if err != nil {
		return nil, err
	}

	//unmarshal response.body
	var result *commtypes.APIResponse
	err = json.Unmarshal(by, &result)
	if err != nil {
		return nil, fmt.Errorf("unmarshal body(%s) failed: %s", string(by), err.Error())
	}
	//if result.Result==false, then request failed
	if !result.Result {
		return nil, fmt.Errorf("request %s failed: %s", url, result.Message)
	}
	by, _ = json.Marshal(result.Data)
	return by, nil
}
