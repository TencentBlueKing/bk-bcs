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

// Package component xxx
package component

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	goReq "github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
)

// CommonResp blueking common response
type CommonResp struct {
	Code    int    `json:"code"`
	Result  bool   `json:"result"`
	Message string `json:"message"`
}

// Request request third api client
func Request(req goReq.SuperAgent, timeout int, proxy string, headers map[string]string) (string, error) {
	client := goReq.New().Timeout(time.Duration(timeout) * time.Second)
	// request by method
	index := 0
	for k := range req.QueryData {
		if index == 0 {
			req.Url = fmt.Sprintf("%s?%s=%s", req.Url, k, req.QueryData.Get(k))
		} else {
			req.Url = fmt.Sprintf("%s&%s=%s", req.Url, k, req.QueryData.Get(k))
		}
		index++
	}
	client = client.CustomMethod(req.Method, req.Url)
	// set proxy
	if proxy != "" {
		client = client.Proxy(proxy)
	}
	// set headers
	for key, val := range headers {
		client = client.Set(key, val)
	}
	for key := range req.Header {
		client = client.Set(key, req.Header.Get(key))
	}
	// request data

	curlCmd := fmt.Sprintf("curl -X %s '%s' ", req.Method, req.Url)

	for key := range client.Header {
		curlCmd += fmt.Sprintf(" -H %q", fmt.Sprintf("%s: %s", key, client.Header.Get(key)))
	}

	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		logging.Error("Failed to encode request data to JSON: %s", err)
		return "", err
	}
	curlCmd += fmt.Sprintf("-d '%s'", string(dataBytes))
	fmt.Println(curlCmd)

	client = client.Send(req.Data)
	client = client.SetDebug(req.Debug)
	_, body, errs := client.End()

	if len(errs) > 0 {
		logging.Error(
			"request api error, url: %s, method: %s, params: %s, data: %s, err: %v",
			req.Url, req.Method, req.QueryData, req.Data, errs,
		)
		return "", errors.New(stringx.Errs2String(errs))
	}
	return body, nil
}

var (
	auditClient *audit.Client
	auditOnce   sync.Once
)

// GetAuditClient 获取审计客户端
func GetAuditClient() *audit.Client {
	if auditClient == nil {
		auditOnce.Do(func() {
			auditClient =
				audit.NewClient(config.GlobalConf.BcsGateway.Host, config.GlobalConf.BcsGateway.Token, nil)
		})
	}
	return auditClient
}

// GetAuthHeader 获取蓝鲸网关通用认证头
func GetAuthHeader() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
		"X-Bkapi-Authorization": fmt.Sprintf(`{"bk_app_code": "%s", "bk_app_secret": "%s", "bk_username": "%s"}`,
			config.GlobalConf.App.Code, config.GlobalConf.App.Secret, config.GlobalConf.App.BkUsername),
	}
}
