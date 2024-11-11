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

	"github.com/pkg/errors"
)

// HTTPRequest defines the http request
type HTTPRequest struct {
	Address     string
	Port        string
	Path        string
	Method      string
	QueryParams map[string]string
	Body        interface{}
	Header      map[string]string
}

// Send the http request
func Send(ctx context.Context, hr *HTTPRequest) ([]byte, error) {
	var req *http.Request
	var err error

	urlStr := fmt.Sprintf("http://%s:%s%s", hr.Address, hr.Port, hr.Path) // nolint
	if hr.Body != nil {
		var body []byte
		body, err = json.Marshal(hr.Body)
		if err != nil {
			return nil, errors.Wrapf(err, "marshal body failed")
		}
		// NOCC:Server Side Request Forgery(只是代码封装，所有 URL都是可信的)
		req, err = http.NewRequestWithContext(ctx, hr.Method, urlStr, bytes.NewBuffer(body))
	} else {
		// NOCC:Server Side Request Forgery(只是代码封装，所有 URL都是可信的)
		req, err = http.NewRequestWithContext(ctx, hr.Method, urlStr, nil)
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

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "http request failed when proxy send")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response body failed when proxy send")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("http response code not 200 but %d, resp: %s",
			resp.StatusCode, string(respBody))
	}
	return respBody, err
}
