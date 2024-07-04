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
	"io"
	"net/http"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/internal/logctx"
)

// HTTPRequest defines the http request
type HTTPRequest struct {
	Url         string
	Method      string
	QueryParams map[string]string
	Body        interface{}
	Header      map[string]string
}

// Send the http request
func Send(ctx context.Context, hr *HTTPRequest) ([]byte, error) {
	var req *http.Request
	var err error

	if hr.Body != nil {
		var body []byte
		body, err = json.Marshal(hr.Body)
		if err != nil {
			return nil, errors.Wrapf(err, "marshal body failed")
		}
		req, err = http.NewRequestWithContext(ctx, hr.Method, hr.Url, bytes.NewBuffer(body))
		blog.Infof("Body: %s", string(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, hr.Method, hr.Url, nil)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "create http request failed")
	}
	for k, v := range hr.Header {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/json")

	if hr.QueryParams != nil {
		query := req.URL.Query()
		for k, v := range hr.QueryParams {
			query.Set(k, v)
		}
		req.URL.RawQuery = query.Encode()
	}

	logctx.Infof(ctx, "http request: [%s] %s", req.Method, req.URL.String())
	var respErr error
	var respBody []byte
	defer func() {
		if respErr != nil {
			logctx.Errorf(ctx, "http request failed: %s", respErr.Error())
		} else {
			tmp := make(map[string]interface{})
			if err = json.Unmarshal(respBody, &tmp); err != nil {
				logctx.Infof(ctx, "http response: %s", strings.ReplaceAll(string(respBody), "\n", ""))
			} else {
				bs, _ := json.Marshal(tmp) // nolint
				logctx.Infof(ctx, "http response: %s", strings.ReplaceAll(string(bs), "\n", ""))
			}
		}
	}()

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		respErr = errors.Wrap(err, "http request failed when proxy send")
		return nil, respErr
	}
	defer resp.Body.Close()
	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		respErr = errors.Wrap(err, "read response body failed when proxy send")
		return nil, respErr
	}
	if resp.StatusCode != http.StatusOK {
		respErr = errors.Errorf("http response code not 200 but %d, resp: %s",
			resp.StatusCode, string(respBody))
		return nil, respErr
	}
	return respBody, respErr
}
