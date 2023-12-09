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
 */

// Package pkg xxx
package pkg

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
)

var apiGatewayPrefix = "/bcsapi/v4"

// Config describe the options Client need
type Config struct {
	// APIServer for bcs-api-gateway address
	APIServer string
	// AuthToken for bcs permission token
	AuthToken string
	// Operator for the bk-repo operations
	Operator string
}

// ProjectManagerClient 项目服务客户端
type ProjectManagerClient struct {
	cfg   *Config
	ctx   context.Context
	debug bool
}

// NewClientWithConfiguration new client with config
func NewClientWithConfiguration(ctx context.Context) *ProjectManagerClient {
	return &ProjectManagerClient{
		ctx: ctx,
		cfg: &Config{
			APIServer: viper.GetString("apiserver"),
			AuthToken: viper.GetString("authtoken"),
			Operator:  viper.GetString("operator"),
		},
		debug: viper.GetBool("debug"),
	}
}

// NewClientWithConfiguration new client with config
func NewClient(ctx context.Context, config *Config) *ProjectManagerClient {
	return &ProjectManagerClient{
		ctx:   ctx,
		cfg:   config,
		debug: viper.GetBool("debug"),
	}
}

func (p *ProjectManagerClient) do(urls string, httpType string, query url.Values, body interface{}) ([]byte, error) {
	urls = p.cfg.APIServer + apiGatewayPrefix + urls
	var req *http.Request
	var err error
	_, err = url.Parse(p.cfg.APIServer)
	if err != nil {
		return nil, err
	}
	var requestParams []byte
	if body != nil {
		requestParams, err = json.Marshal(body)
		if err != nil {
			return nil, errors.Wrapf(err, "marshal body failed")
		}
		req, err = http.NewRequestWithContext(p.ctx, httpType, urls, bytes.NewReader(requestParams))
	} else {
		req, err = http.NewRequestWithContext(p.ctx, httpType, urls, nil)
	}
	if query != nil {
		req.URL.RawQuery = query.Encode()
	}
	// 添加鉴权
	if len(p.cfg.AuthToken) != 0 {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", p.cfg.AuthToken))
	}
	if err != nil {
		return nil, errors.Wrapf(err, "create request failed")
	}
	// 打印请求参数
	if body != nil {
		p.glogBody("Request Body", requestParams)
	}

	// 打印请求前
	p.debugRequest(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "http do request failed")
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read response body failed")
	}

	// 打印请求后
	p.debugResponse(resp)
	p.glogBody("Response Body", respBody)

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrapf(errors.Errorf(string(respBody)), "http response status not 200 but %d",
			resp.StatusCode)
	}
	return respBody, nil
}

func (p *ProjectManagerClient) glogBody(prefix string, body []byte) {
	if p.debug {
		if bytes.IndexFunc(body, func(r rune) bool {
			return r < 0x0a
		}) != -1 {
			klog.Infof("%s:\n%s", prefix, truncateBody(hex.Dump(body)))
		} else {
			klog.Infof("%s: %s", prefix, truncateBody(string(body)))
		}
	}
}

func (p *ProjectManagerClient) debugRequest(req *http.Request) {
	if p.debug {
		// 把链接转成curl
		klog.Infof("%s", toCurl(req))
		// 打印请求方法和地址
		klog.Infof("%s %s", req.Method, req.URL.String())
	}
}

func (p *ProjectManagerClient) debugResponse(resp *http.Response) {
	if p.debug {
		klog.Info("Response Headers:")
		for key, values := range resp.Header {
			for _, value := range values {
				klog.Infof("    %s: %s", key, value)
			}
		}
	}
}

func truncateBody(body string) string {
	max := 0
	switch {
	case klog.V(0).Enabled():
		return body
	case klog.V(0).Enabled(): // nolint
		max = 10240
	}

	if len(body) <= max {
		return body
	}

	return body[:max] + fmt.Sprintf(" [truncated %d chars]", len(body)-max)
}

var knownAuthTypes = map[string]bool{
	"bearer":    true,
	"basic":     true,
	"negotiate": true,
}

func toCurl(req *http.Request) string {
	headers := ""
	for key, values := range req.Header {
		for _, value := range values {
			value = maskValue(key, value)
			headers += fmt.Sprintf(` -H %q`, fmt.Sprintf("%s: %s", key, value))
		}
	}

	return fmt.Sprintf("curl -v -X%s %s '%s'", req.Method, headers, req.URL.String())
}

// maskValue masks credential content from authorization headers
// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Authorization
func maskValue(key string, value string) string {
	if !strings.EqualFold(key, "Authorization") {
		return value
	}
	if len(value) == 0 {
		return ""
	}
	var authType string
	if i := strings.Index(value, " "); i > 0 {
		authType = value[0:i]
	} else {
		authType = value
	}
	if !knownAuthTypes[strings.ToLower(authType)] {
		return "<masked>"
	}
	if len(value) > len(authType)+1 {
		value = authType + " <masked>"
	} else {
		value = authType
	}
	return value
}
