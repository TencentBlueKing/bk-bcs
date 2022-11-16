/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/rest"
	"bscp.io/pkg/runtime/gwparser"
	"bscp.io/pkg/runtime/shutdown"
)

// setupFilters setups all api filters here. All request would cross here, and we filter request base on URL.
func (p *proxy) setupFilters(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		kt, err := gwparser.Parse(r.Context(), r.Header)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, errf.Error(err).Error())
			return
		}

		body, err := peekRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, errf.Error(errf.New(errf.Unknown, err.Error())).Error())
			logs.Errorf("peek request failed, err: %v, rid: %s", err, kt.Rid)
			return
		}
		// request and response details landing log for monitoring and troubleshooting problem.
		logs.Infof("uri: %s, method: %s, body: %s, appcode: %s, user: %s, remote addr: %s, "+
			"rid: %s", r.RequestURI, r.Method, body, kt.AppCode, kt.User, r.RemoteAddr, kt.Rid)

		// handler request.
		mux.ServeHTTP(w, r)
	})
}

// Healthz service health check.
func (p *proxy) Healthz(w http.ResponseWriter, r *http.Request) {
	if shutdown.IsShuttingDown() {
		logs.Errorf("service healthz check failed, current service is shutting down")
		w.WriteHeader(http.StatusServiceUnavailable)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealth, "current service is shutting down"))
		return
	}

	if err := p.state.Healthz(cc.ApiServer().Service.Etcd); err != nil {
		logs.Errorf("etcd healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealth, "etcd healthz error, "+err.Error()))
		return
	}

	rest.WriteResp(w, rest.NewBaseResp(errf.OK, "healthy"))
	return
}

func peekRequest(req *http.Request) (string, error) {
	// content upload body maybe too big, can not peek.
	if !strings.HasPrefix(req.URL.Path, "/api/v1/api/create/content/upload") && req.Body != nil {
		byt, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return "", err
		}

		req.Body = ioutil.NopCloser(bytes.NewBuffer(byt))

		reg := regexp.MustCompile("\\s+")
		str := reg.ReplaceAllString(string(byt), "")
		return str, nil
	}

	return "", nil
}
