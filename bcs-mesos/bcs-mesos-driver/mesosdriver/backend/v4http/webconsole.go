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

package v4http

import (
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-driver/mesosdriver/backend/webconsole"

	restful "github.com/emicklei/go-restful"
)

//initMesosWebconsole init webconsole proxy, including normal http proxy & websocket proxy
func (s *Scheduler) initMesosWebconsole() {
	s.consoleProxy = webconsole.NewWebconsoleProxy(s.config.MesosWebconsoleProxyPort, s.config.ClientCert)
}

//webconsoleForwarding forwarding webconsole information to specified host
func (s *Scheduler) webconsoleForwarding(req *restful.Request, resp *restful.Response) {
	// name := req.PathParameter("name")
	// kubeURL := defaultCRDURL
	// if len(name) != 0 {
	// 	kubeURL = filepath.Join(defaultCRDURL, name)
	// }
	// rawRequest := req.Request
	// original := rawRequest.URL.String()
	// tmpURL, _ := url.Parse(s.localProxy.config.Host)
	// rawRequest.URL.Scheme = tmpURL.Scheme
	// rawRequest.URL.Host = tmpURL.Host
	// rawRequest.URL.Path = kubeURL
	// blog.Infof("bcs-mesos-driver CustomResourceDefinition forwarding from %s to %s", original, rawRequest.URL.String())
	// s.localProxy.crdsProxy.ServeHTTP(resp, rawRequest)
	s.consoleProxy.ServeHTTP(resp.ResponseWriter, req.Request)
}
