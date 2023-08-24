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

package session

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
)

// SecretSession defines the instance that to proxy to secret server
type SecretSession struct {
	op *proxy.SecretOption
}

// NewSecretSession create the session of secret
func NewSecretSession(op *proxy.SecretOption) *SecretSession {
	return &SecretSession{
		op: op,
	}
}

// ServeHTTP http.Handler implementation
func (s *SecretSession) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// backend real path with encoded format
	realPath := strings.TrimPrefix(req.URL.RequestURI(), common.GitOpsProxyURL)
	// !force https link
	fullPath := fmt.Sprintf("http://%s:%s%s", s.op.Address, s.op.Port, realPath)
	newURL, err := url.Parse(fullPath)
	if err != nil {
		err = errors.Errorf("secret session build new fullpath '%s' failed: %s", fullPath, err.Error())
		rw.WriteHeader(http.StatusInternalServerError)
		blog.Errorf(err.Error())
		_, _ = rw.Write([]byte(err.Error())) // nolint
		return
	}
	reverseProxy := httputil.ReverseProxy{
		Director: func(request *http.Request) {
			request.URL = newURL
		},
		ErrorHandler: func(res http.ResponseWriter, request *http.Request, e error) {
			blog.Errorf("secret session proxy '%s' with header '%s' failure: %s",
				fullPath, request.Header, e.Error())
			res.WriteHeader(http.StatusInternalServerError)
			_, _ = res.Write([]byte("secret session proxy failed")) // nolint
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint
		},
		ModifyResponse: func(r *http.Response) error {
			return nil
		},
	}
	blog.Infof("secret session serve: %s/%s", req.Method, fullPath)
	reverseProxy.ServeHTTP(rw, req)
}
