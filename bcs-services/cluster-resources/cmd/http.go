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

// Package cmd http 接口实现
package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/contextx"
	httpUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/http"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/httpx"
)

// NewAPIRouter http handler
func NewAPIRouter(crs *clusterResourcesService) *mux.Router {
	r := mux.NewRouter()
	// add middleware
	r.Use(httpx.LoggingMiddleware)
	r.Use(httpx.AuthenticationMiddleware)
	r.Use(httpx.ParseProjectIDMiddleware)
	r.Use(httpx.AuthorizationMiddleware)

	// events 接口代理
	r.Methods("GET").Path("/clusterresources/api/v1/projects/{projectCode}/clusters/{clusterID}/events").
		Handler(httpx.ParseClusterIDMiddleware(http.HandlerFunc(StorageEvents(crs))))
	return r
}

// StorageEvents reverse proxy events
func StorageEvents(crs *clusterResourcesService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		targetURLPath := fmt.Sprintf("%s/bcsstorage/v1/events", config.G.Component.BCSStorageHost)

		targetURL, err := url.Parse(targetURLPath)
		if err != nil {
			httpx.ResponseSystemError(w, r, err)
			return
		}
		clusterID := contextx.GetClusterIDFromCtx(r.Context())
		query := r.URL.Query()
		query.Set("clusterId", clusterID)
		targetURL.RawQuery = query.Encode()

		proxy := httpUtil.NewHTTPReverseProxy(crs.clientTLSConfig, func(request *http.Request) {
			request.URL = targetURL
			request.Method = http.MethodGet
		})
		proxy.ServeHTTP(w, r)
	}
}
