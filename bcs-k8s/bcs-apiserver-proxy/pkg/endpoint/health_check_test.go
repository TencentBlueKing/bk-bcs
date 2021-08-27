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

package endpoint

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

var (
	addr        = "127.0.0.1"
	port uint32 = 8888
)

func TestHealthConfig_IsHTTPAPIHealth(t *testing.T) {
	server := fmt.Sprintf("%s:%d", addr, port)
	startHTTPServer(server)

	time.Sleep(2 * time.Second)

	health, err := NewHealthConfig("http", "/health")
	if err != nil {
		t.Fatalf("NewHealthConfig failed: %v", err)
		return
	}

	ok := health.IsHTTPAPIHealth(addr, port)
	if !ok {
		t.Fatalf("IsHTTPAPIHealth failed")
		return
	}

	t.Log("IsHTTPAPIHealth successful")
}

func startHTTPServer(addr string) {
	http.HandleFunc("/health", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("ok"))
	})
	go http.ListenAndServe(addr, nil)
}
