/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package component

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/parnurzeal/gorequest"
	"github.com/stretchr/testify/assert"
)

// ref: https://github.com/parnurzeal/gorequest/blob/develop/gorequest_test.go
func TestRequest(t *testing.T) {
	// 预制值
	case1_empty := "/"
	case2_set_header := "/set_header"
	retData := "hello world"

	// 设置一个 http service
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check method is GET before going to check other features
		if r.Method != "GET" {
			t.Errorf("Expected method %q; got %q", "GET", r.Method)
		}
		if r.Header == nil {
			t.Error("Expected non-nil request Header")
		}
		w.Write([]byte(retData))
		switch r.URL.Path {
		default:
			t.Errorf("No testing for this case yet : %q", r.URL.Path)
		case case1_empty:
			t.Logf("case %v ", case1_empty)
		case case2_set_header:
			t.Logf("case %v ", case2_set_header)
			if r.Header.Get("API-Key") != "fookey" {
				t.Errorf("Expected 'API-Key' == %q; got %q", "fookey", r.Header.Get("API-Key"))
			}
		}
	}))

	defer ts.Close()

	// 发起请求
	req := gorequest.SuperAgent{
		Url:    ts.URL,
		Method: "GET",
	}
	timeout := 10
	body, err := Request(req, timeout, "", map[string]string{})
	assert.Nil(t, err)
	assert.Equal(t, body, retData)

	Request(req, timeout, "", map[string]string{"API-Key": "fookey"})
}
