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

package v1

// UptimeCheckConfig 拨测配置
type UptimeCheckConfig struct {
	Enabled bool `json:"enabled"`
	// Name    string `json:"name,omitempty"`

	Protocol string `json:"protocol,omitempty"` // if not set, use listener protocol as default,
	// support HTTP(S)/ TCP/ UDP/ ICMP
	// Target   []string `json:"target"`
	Port int64 `json:"port,omitempty"` // if not set, use listeners port as default

	Nodes  []string `json:"nodes"`            // 发起拨测的节点名
	Groups []string `json:"groups,omitempty"` // 拨测任务归属的任务组

	Timeout int64 `json:"timeout,omitempty"` // 请求超时时间 单位ms, default 3000
	Period  int64 `json:"period,omitempty"`  // 访问周期， 单位s， default 60

	Response       string `json:"response,omitempty"`            // 期待响应信息
	ResponseFormat string `json:"response_format,omitempty"`     // 响应格式， 如raw|in, hex|in
	Request        string `json:"request,omitempty"`             // 请求内容
	RequestFormat  string `json:"request_format,omitempty"`      // 请求格式， 如raw, hex
	WaitResponse   bool   `json:"wait_empty_response,omitempty"` // 是否等待响应, 默认true

	// HTTP(S) Related properties
	Method       string     `json:"method,omitempty"` // GET POST PUT DELETE PATCH
	Authorize    *Authorize `json:"authorize,omitempty"`
	Body         *Body      `json:"body,omitempty"`
	QueryParams  []*Params  `json:"query_params,omitempty"`
	Headers      []*Params  `json:"headers,omitempty"`
	ResponseCode string     `json:"response_code,omitempty"`
	URLList      []string   `json:"url_list,omitempty"`
}

// Params http check params
type Params struct {
	IsEnabled bool   `json:"is_enabled"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Desc      string `json:"desc"`
}

// Authorize auth config
type Authorize struct {
	AuthType           string      `json:"auth_type,omitempty"` // basic_auth / bearer_token
	AuthConfig         *AuthConfig `json:"auth_config,omitempty"`
	InsecureSkipVerify bool        `json:"insecure_skip_verify,omitempty"`
}

// AuthConfig auth config
type AuthConfig struct {
	Token    string `json:"token,omitempty"`    // set when auth_type = 'bearer_token'
	UserName string `json:"username,omitempty"` // set when auth_type = 'basic_auth'
	PassWord string `json:"password,omitempty"` // set when auth_type = 'basic_auth'
}

// Body http body
type Body struct {
	DataType    string    `json:"data_type,omitempty"` // default / raw / form_data/ x_www_form_urlencoded
	Params      []*Params `json:"params,omitempty"`    // set when DataType in [form_data, x_www_form_urlencoded]
	Content     string    `json:"content,omitempty"`
	ContentType string    `json:"content_type,omitempty"`
}
