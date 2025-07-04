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

package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	traceconst "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/constants"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/internal/logctx"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/options"
)

// HTTPRequest defines the http request
type HTTPRequest struct {
	Url         string
	Method      string
	QueryParams map[string]string
	Body        interface{}
	Header      map[string]string
}

// SendHTTPRequest the http request
func SendHTTPRequest(ctx context.Context, hr *HTTPRequest) ([]byte, error) {
	_, respBody, err := SendHTTPRequestReturnResponse(ctx, hr)
	return respBody, err
}

// SendHTTPRequestReturnResponse the http request return response
func SendHTTPRequestReturnResponse(ctx context.Context, hr *HTTPRequest) (*http.Response, []byte, error) {
	var req *http.Request
	var err error

	if !strings.Contains(hr.Url, "custom_api") {
		logctx.Infof(ctx, "do request '%s'", hr.Url)
	}
	if hr.Body != nil {
		var body []byte
		body, err = json.Marshal(hr.Body)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "marshal body failed")
		}
		// NOCC:Server Side Request Forgery(只是代码封装，所有 URL都是可信的)
		req, err = http.NewRequestWithContext(ctx, hr.Method, hr.Url, bytes.NewBuffer(body))
	} else {
		// NOCC:Server Side Request Forgery(只是代码封装，所有 URL都是可信的)
		req, err = http.NewRequestWithContext(ctx, hr.Method, hr.Url, nil)
	}
	if err != nil {
		return nil, nil, errors.Wrapf(err, "create http request failed")
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range hr.Header {
		req.Header.Set(k, v)
	}
	requestID := logctx.RequestID(ctx)
	if requestID != "" {
		req.Header.Set(traceconst.RequestIDHeaderKey, requestID)
	}

	if hr.QueryParams != nil {
		query := req.URL.Query()
		for k, v := range hr.QueryParams {
			query.Set(k, v)
		}
		req.URL.RawQuery = query.Encode()
	}

	var resp *http.Response
	httpClient := http.DefaultClient
	if !strings.Contains(hr.Url, "custom_api") {
		httpClient.Transport = options.GlobalOptions().HTTPProxyTransport()
	}
	for i := 0; i < 10; i++ {
		resp, err = httpClient.Do(req)
		if err == nil {
			break
		}
		logctx.Warnf(ctx, "do request '%s, %s' failed(retry=%d): %s", req.Method,
			req.URL.String(), i, err.Error())
		time.Sleep(time.Second)
	}
	if err != nil {
		return nil, nil, errors.Wrap(err, "http request failed")
	}
	defer resp.Body.Close()

	var respBody []byte
	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, errors.Wrap(err, "read response body failed")
	}
	if resp.StatusCode != http.StatusOK {
		return resp, nil, errors.Errorf("http response code not 200 but %d, resp: %s",
			resp.StatusCode, string(respBody))
	}
	return resp, respBody, nil
}
