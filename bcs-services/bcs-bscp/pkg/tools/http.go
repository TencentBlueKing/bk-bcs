/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"bscp.io/pkg/criteria/constant"
	"k8s.io/klog/v2"
)

var (
	maskKeys = map[string]struct{}{
		"bk_app_secret": {},
		"bk_token":      {},
		"Authorization": {},
	}
)

// RequestIDValue 获取 RequestId 值
func RequestIDValue(req *http.Request) string {
	for _, k := range constant.RidKeys {
		v := req.Header.Get(k)
		if v != "" {
			return v
		}
	}
	return ""
}

// curlLogTransport print curl log transport
type curlLogTransport struct {
	Transport http.RoundTripper
}

// NewCurlLogTransport make a new curl log transport, default transport can be nil
func NewCurlLogTransport(transport http.RoundTripper) http.RoundTripper {
	return &curlLogTransport{Transport: transport}
}

// reqToCurl curl 格式的请求日志
func reqToCurl(r *http.Request) string {
	headers := ""
	for key, values := range r.Header {
		for _, value := range values {
			if _, ok := maskKeys[key]; ok {
				value = "<masked>"
			}
			headers += fmt.Sprintf(" -H %q", fmt.Sprintf("%s: %s", key, value))
		}
	}

	// 过滤掉敏感信息
	rawURL := *r.URL
	queryValue := rawURL.Query()
	for key := range queryValue {
		if _, ok := maskKeys[key]; ok {
			queryValue.Set(key, "<masked>")
		}
	}
	rawURL.RawQuery = queryValue.Encode()

	reqMsg := fmt.Sprintf("curl -X %s '%s'%s", r.Method, rawURL.String(), headers)
	if r.Body != nil {
		switch body := r.Body.(type) {
		case io.Reader:
			reqMsg += " -d (io.Reader)"
		default:
			prtBodyBytes, err := json.Marshal(body)
			if err != nil {
				reqMsg += fmt.Sprintf(" -d %q (MarshalErr %s)", body, err)
			} else {
				reqMsg += fmt.Sprintf(" -d '%s'", prtBodyBytes)
			}
		}
	}
	if r.Form.Encode() != "" {
		encodeStr := r.Form.Encode()
		reqMsg += fmt.Sprintf(" -d %q", encodeStr)
		rawStr, _ := url.QueryUnescape(encodeStr)
		reqMsg += fmt.Sprintf(" -raw `%s`", rawStr)
	}

	return reqMsg
}

// RoundTrip Transport
func (t *curlLogTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	klog.Infof("[%s] REQ: %s", RequestIDValue(req), reqToCurl(req))
	resp, err := t.transport(req).RoundTrip(req)
	return resp, err
}

func (t *curlLogTransport) transport(req *http.Request) http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}
