/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mesoswebconsole

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestWebsocketProxyServeHTTP(t *testing.T) {
	req, err := http.NewRequest("GET", "/mesosdriver/v4/webconsole/{sub_path:.*}", bytes.NewReader(nil))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	u := "https://127.0.0.1:8087"

	backendURL, err := url.Parse(u)
	if err != nil {
		t.Fatal("error parse url")
	}
	wp := NewWebsocketProxy(nil, backendURL, nil)
	wp.ServeHTTP(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Error("should return 505 http code")
	}
}

func TestSetRequestHeader(t *testing.T) {
	req, err := http.NewRequest("GET", "/mesosdriver/v4/webconsole/{sub_path:.*}", bytes.NewReader(nil))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("BCS-ClusterID", "k8s-001")
	req.Header.Add("Sec-WebSocket-Protocol", "websocket")

	backendURL := url.URL{}
	wp := NewWebsocketProxy(nil, &backendURL, nil)
	requestHeader := wp.setRequestHeader(req)

	if requestHeader.Get("BCS-ClusterID") != "k8s-001" {
		t.Error("error set upgrade request header")
	}

	if requestHeader.Get("Sec-WebSocket-Protocol") != "websocket" {
		t.Error("error set upgrade request header")
	}
}
