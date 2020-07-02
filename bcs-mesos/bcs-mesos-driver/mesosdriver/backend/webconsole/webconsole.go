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
	"net/url"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-driver/mesosdriver/config"

	"github.com/gorilla/websocket"
)

//WebconsoleProxy mesos web console proxy implementation for mesos-driver
type WebconsoleProxy struct {
	// Backend returns the backend URL which the proxy uses to reverse proxy
	Backend func(*http.Request) (*url.URL, error)
	//Certificatio configuration for backend console proxy
	CertConfig *config.CertConfig
}

//NewWebconsoleProxy create proxy instance for mesos-driver
func NewWebconsoleProxy(port uint, certConfig *config.CertConfig) *WebconsoleProxy {
	backend := func(req *http.Request) (*url.URL, error) {
		v := req.URL.Query()
		hostIp := v.Get("host_ip")
		if hostIp == "" {
			return nil, fmt.Errorf("param host_ip must not be empty")
		}
		host := net.JoinHostPort(hostIp, strconv.Itoa(int(port)))

		u := &url.URL{}
		u.Host = host
		u.Fragment = req.URL.Fragment
		//! this is hard code in WebconsoleProxy from 1.17.x, it's not elegent.
		//todo(DeveloperJim): discussion about mechanism that we can avoid hard code
		u.Path = strings.Replace(req.URL.Path, "/mesosdriver/v4/webconsole", "/bcsapi/v1/consoleproxy", 1)
		u.RawQuery = req.URL.RawQuery
		blog.Infof(u.String())
		return u, nil
	}

	return &WebconsoleProxy{
		Backend:    backend,
		CertConfig: certConfig,
	}
}

//ServeHTTP original http interface implementation
func (w *WebconsoleProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	backendURL, err := w.Backend(req)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	if websocket.IsWebSocketUpgrade(req) {
		websocketProxy := NewWebsocketProxy(w.CertConfig, backendURL)
		websocketProxy.ServeHTTP(rw, req)
		return
	}

	httpProxy, err := NewHttpReverseProxy(backendURL, w.CertConfig)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	httpProxy.ServeHTTP(rw, req)
	return
}
