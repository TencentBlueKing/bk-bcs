/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package bscpclient

import (
	"net/http"

	"bscp.io/pkg/bscpclient/api"
	"bscp.io/pkg/criteria/constant"
)

// Client for call bscp api.
type Client struct {
	ApiClient *api.Client
}

// Config for client config
type Config struct {
	ApiHost string
}

// NewClient new a client
func NewClient(cfg Config) (*Client, error) {
	apiCli, err := api.NewApiClient(cfg.ApiHost, nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		ApiClient: apiCli,
	}, nil
}

// Header generate header for api request.
func Header(rid string) http.Header {
	header := http.Header{}
	header.Set(constant.UserKey, constant.BKUserForTestPrefix+"bscp-client")
	header.Set(constant.RidKey, rid)
	header.Set(constant.AppCodeKey, "bk-bscp-client")
	header.Add("Cookie", "bk_token="+constant.BKTokenForTest)

	return header
}
