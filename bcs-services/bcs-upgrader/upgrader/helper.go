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
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-upgrader/app/options"
)

// UpgradeHelper is a helper for upgrade
type UpgradeHelper interface {
	// HelperName return the name of the helper
	HelperName() string
}

// Helper is an implementation for interface UpgradeHelper
type Helper struct {
	DB     drivers.DB
	config options.HttpCliConfig
	// http 调用bcs-saas cc模块提供的接口
	httpClient *httpclient.HttpClient
	// 调用bcs-cluster-manager提供的接口
	clusterManagerClient *httpclient.HttpClient
}

// HelperOpt is option for Helper
type HelperOpt struct {
	DB     drivers.DB
	config options.HttpCliConfig
}

// Name is the method of Helper to implement interface UpgradeHelper
func (h *Helper) HelperName() string {
	return "bcs-upgrade-helper"
}

// NewUpgradeHelper new a Helper instance
func NewUpgradeHelper(opt *HelperOpt) *Helper {

	// init clusterManager cli
	clusterManagerCli := httpclient.NewHttpClient()
	clusterManagerCli.SetHeader("Content-Type", "application/json")
	clusterManagerCli.SetHeader("Authorization", "Bearer "+opt.config.BcsApiGatewayToken)

	// if https
	if opt.config.ClusterManagerCertConfig != nil && opt.config.ClusterManagerCertConfig.IsSSL {
		clusterManagerCli.SetTlsVerity(opt.config.ClusterManagerCertConfig.CAFile,
			opt.config.ClusterManagerCertConfig.CertFile, opt.config.ClusterManagerCertConfig.KeyFile,
			opt.config.ClusterManagerCertConfig.CertPasswd)
	}

	// init bcs-saas cc cli
	httpCli := httpclient.NewHttpClient()
	httpCli.SetHeader("Content-Type", "application/json")
	// if https
	if opt.config.HttpCliCertConfig != nil && opt.config.HttpCliCertConfig.IsSSL {
		httpCli.SetTlsVerity(opt.config.HttpCliCertConfig.CAFile, opt.config.HttpCliCertConfig.CertFile,
			opt.config.HttpCliCertConfig.KeyFile, opt.config.HttpCliCertConfig.CertPasswd)
	}

	return &Helper{
		DB:                   opt.DB,
		httpClient:           httpCli,
		clusterManagerClient: clusterManagerCli,
		config:               opt.config,
	}
}

//HttpRequest : http cli request
func (h *Helper) HttpRequest(method, path string, payload []byte) ([]byte, error) {
	url := fmt.Sprintf(h.config.CcHOST+path, h.config.SsmAccessToken)
	return h.requestApiServer(h.httpClient, method, url, payload)
}

// ClusterManagerRequest : ClusterManager cli request
func (h *Helper) ClusterManagerRequest(method, path string, payload []byte) ([]byte, error) {
	url := h.config.ClusterManagerHost + path
	return h.requestApiServer(h.clusterManagerClient, method, url, payload)
}

//method=http.method: POST、GET、PUT、DELETE
//request url = address/url
//payload is request body
//if error!=nil, then request mesos failed, errom.Error() is failed message
//if error==nil, []byte is response body information
func (h *Helper) requestApiServer(cli *httpclient.HttpClient, method, url string, payload []byte) ([]byte, error) {

	var err error
	var by []byte

	switch method {
	case "GET":
		by, err = cli.GET(url, nil, payload)
	case "POST":
		by, err = cli.POST(url, nil, payload)
	case "DELETE":
		by, err = cli.DELETE(url, nil, payload)
	case "PUT":
		by, err = cli.PUT(url, nil, payload)
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

// SetSsmToken  SetSsmToken
func (h *Helper) SetSsmToken() error {

	data := map[string]string{
		"grant_type":  "client_credentials",
		"id_provider": "client",
	}
	dataByte, err := json.Marshal(data)
	if err != nil {
		return err
	}

	cli := httpclient.NewHttpClient()

	cli.SetHeader("Content-Type", "application/json")
	cli.SetHeader("X-BK-APP-CODE", "bk_cmdb")
	cli.SetHeader("X-BK-APP-SECRET", h.config.BkAppSecret)

	replyData, err := cli.POST(h.config.SsmHost, nil, dataByte)
	if err != nil {
		return err
	}

	type respGetCCToken struct {
		AccessToken string `json:"access_token"`
	}

	resp := new(respGetCCToken)
	err = json.Unmarshal(replyData, resp)
	if err != nil {
		return err
	}

	h.config.SsmAccessToken = resp.AccessToken

	return nil
}
