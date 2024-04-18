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

// Package httputils xx
package httputils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/moul/http2curl"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/utils"
)

// HTTPRequest defines the http request
type HTTPRequest struct {
	Path        string
	Method      string
	QueryParams map[string][]string
	Body        interface{}
	Header      map[string]string
}

// DoRequest the http request
func DoRequest(ctx context.Context, hr *HTTPRequest) []byte {
	op := options.GlobalOption()
	var req *http.Request
	var err error
	urlStr := "https://" + path.Join(op.Server, op.ProxyPath, hr.Path)
	if hr.Body != nil {
		var body []byte
		body, err = json.Marshal(hr.Body)
		if err != nil {
			utils.ExitError(fmt.Sprintf("marshal request body failed: %s", err.Error()))
		}
		req, err = http.NewRequestWithContext(ctx, hr.Method, urlStr, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, hr.Method, urlStr, nil)
	}
	if err != nil {
		utils.ExitError(fmt.Sprintf("create http request failed: %s", err.Error()))
	}
	for k, v := range hr.Header {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+op.Token)
	if hr.QueryParams != nil {
		query := req.URL.Query()
		for k, v := range hr.QueryParams {
			for i := range v {
				query.Add(k, v[i])
			}
		}
		req.URL.RawQuery = query.Encode()
	}
	command, _ := http2curl.GetCurlCommand(req)
	blog.V(3).Infof(command.String())

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	cost := time.Since(start).Milliseconds()
	if err != nil {
		utils.ExitError(fmt.Sprintf("%s '%s' do request failed in %v: %s",
			strings.ToLower(req.Method), req.URL.String(), cost, err.Error()))
	} else {
		blog.V(3).Infof("%s %s %d in %v", strings.ToUpper(req.Method), req.URL.String(), resp.StatusCode, cost)
	}
	defer resp.Body.Close()
	blog.V(3).Infof("Response Headers:")
	for k, v := range resp.Header {
		blog.V(3).Infof("    %s: %s", k, strings.Join(v, ", "))
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.ExitError(fmt.Sprintf("read response body failed: %s", err.Error()))
	}
	blog.V(3).Infof("Response Body: %s", string(respBody))

	if resp.StatusCode != http.StatusOK {
		utils.ExitError(string(respBody))
	}
	return respBody
}
