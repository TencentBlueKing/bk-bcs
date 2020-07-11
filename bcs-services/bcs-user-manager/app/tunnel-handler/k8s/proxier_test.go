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

package k8s

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

type TestHandler struct {
	// ClusterVarName is the path parameter name of cluster_id
	ClusterVarName string
	// SubPathVarName is the path parameter name of sub-path needs to be forwarded
	SubPathVarName string
}

type BackendHandler struct {
}

func (b *BackendHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	rw.Write([]byte(path))
}

var DefaultTestHandler = NewTestHandler("cluster_id", "sub_path")

// NewTestHandler new a default TestHandler
func NewTestHandler(clusterVarName, subPathVarName string) *TestHandler {
	return &TestHandler{
		ClusterVarName: clusterVarName,
		SubPathVarName: subPathVarName,
	}
}

func (t *TestHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	backend := &BackendHandler{}
	subPath := mux.Vars(req)[t.SubPathVarName]
	fullPath := req.URL.Path

	handlerServer := stripLeaveSlash(fullPath[:len(fullPath)-len(subPath)], backend)
	handlerServer.ServeHTTP(rw, req)
}

func TestStripLeaveSlash(t *testing.T) {
	router := mux.NewRouter()
	router.Handle("/tunnels/clusters/{cluster_id}/{sub_path:.*}", DefaultTestHandler)

	urls := []string{"/tunnels/clusters/k8s-001/version", "/tunnels/clusters/k8s-001/apis/apps/v1/namespaces/default/deployments/test"}
	for _, url := range urls {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		stripPath := strings.TrimPrefix(url, "/tunnels/clusters/k8s-001")
		if rr.Body.String() != stripPath {
			t.Error("stripLeaveSlash got an unexpected path")
		}
	}
}

func TestTunnelServeHTTP(t *testing.T) {
	req, err := http.NewRequest("GET", "/tunnels/clusters/k8s-001/version", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	DefaultTunnelProxyDispatcher.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Error("tunnel api should provide admin token")
	}
}
