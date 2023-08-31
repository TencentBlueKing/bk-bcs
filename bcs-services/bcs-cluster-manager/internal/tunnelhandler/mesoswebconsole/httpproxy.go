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

package mesoswebconsole

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer"
)

// NewHTTPReverseProxy create a http reverse proxy
func NewHTTPReverseProxy(
	clientTLSConfig *tls.Config, target *url.URL, clusterDialer websocketDialer.Dialer) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
		req.URL.RawQuery = target.RawQuery
	}
	reverseProxy := &httputil.ReverseProxy{Director: director}
	// use the cluster tunnel dialer as transport Dial
	tp := &http.Transport{
		//Proxy: http.ProxyFromEnvironment,
		Dial: clusterDialer,
	}
	if target.Scheme == "https" {
		tp.TLSClientConfig = clientTLSConfig
	}
	reverseProxy.Transport = tp

	return reverseProxy
}
