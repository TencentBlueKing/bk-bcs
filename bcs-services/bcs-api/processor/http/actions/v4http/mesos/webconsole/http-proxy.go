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

package webconsole

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/config"
)

func NewHttpReverseProxy(target *url.URL, certConfig *config.CertConfig) (*httputil.ReverseProxy, error) {
	if certConfig.IsSSL {
		target.Scheme = "https"
	}

	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
		req.URL.RawQuery = target.RawQuery
	}
	reverseProxy := &httputil.ReverseProxy{Director: director}
	if certConfig.IsSSL {
		cliTls, err := ssl.ClientTslConfVerity(certConfig.CAFile, certConfig.CertFile, certConfig.KeyFile, certConfig.CertPasswd)
		if err != nil {
			blog.Errorf("set client tls config error %s", err.Error())
			return nil, fmt.Errorf("set client tls config error %s", err.Error())
		}
		reverseProxy.Transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 30 * time.Second,
			TLSClientConfig:     cliTls,
		}
	}
	return reverseProxy, nil
}
