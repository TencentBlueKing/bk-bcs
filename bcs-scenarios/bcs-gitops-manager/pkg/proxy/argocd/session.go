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

package argocd

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
)

// Session purpose: simple revese proxy for argocd according kubernetes service.
// gitops proxy implements http.Handler interface.
type Session struct {
	option *proxy.GitOpsOptions
}

// ServeHTTP http.Handler implementation
func (s *Session) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// backend real path with encoded format
	realPath := strings.TrimPrefix(req.URL.RequestURI(), common.GitOpsProxyURL)
	// !force https link
	fullPath := fmt.Sprintf("https://%s%s", s.option.Service, realPath)
	newURL, err := url.Parse(fullPath)
	if err != nil {
		blog.Errorf("GitOps session build new fullpath %s failed, %s", fullPath, err.Error())
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("URL conversion failure in manager")) // nolint
		return
	}
	reverseProxy := httputil.ReverseProxy{
		Director: func(request *http.Request) {
			request.URL = newURL
			// setting login session token for pass through, for http 1.x
			token := s.option.Storage.GetToken(request.Context())
			request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			// for http 2
			request.Header.Set("Token", token)
		},
		ErrorHandler: func(res http.ResponseWriter, request *http.Request, e error) {
			blog.Errorf("GitOps proxy %s failure, %s. header: %+v", fullPath, e.Error(), request.Header)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("GitOps Proxy session failure")) // nolint
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint
		},
		ModifyResponse: func(r *http.Response) error {
			blog.Infof("GitOps proxy %s response header details: %+v, status %s, code: %d",
				fullPath, r.Header, r.Status, r.StatusCode)
			return nil
		},
	}
	// all ready to serve
	blog.Infof("GitOps serve %s %s", req.Method, fullPath)
	reverseProxy.ServeHTTP(rw, req)
}
