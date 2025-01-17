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

// Package http xxx
package http

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"

	"k8s.io/klog/v2"
)

// NewHTTPReverseProxy new http reverse proxy
func NewHTTPReverseProxy(clientTLSConfig *tls.Config, f func(request *http.Request)) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: f,
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
			klog.Errorf("new http reverse proxy request failed, err: %s", err.Error())
			rw.WriteHeader(http.StatusInternalServerError)
		},
		Transport: &http.Transport{
			TLSClientConfig: clientTLSConfig,
		},
	}
}
