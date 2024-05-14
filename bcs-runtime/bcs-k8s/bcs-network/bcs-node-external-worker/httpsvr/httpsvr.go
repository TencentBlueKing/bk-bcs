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

// Package httpsvr xxx
package httpsvr

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-node-external-worker/options"
)

// HttpServerClient http server client
type HttpServerClient struct {
	Ops options.Options

	UUID string
}

// Init xxx
func (server *HttpServerClient) Init() error {
	s := httpserver.NewHttpServer(server.Ops.ListenPort, "0.0.0.0", "")

	s.SetInsecureServer("0.0.0.0", server.Ops.ListenPort)
	ws := s.NewWebService("/node-external-worker", nil)
	server.initRouters(ws)

	router := s.GetRouter()
	webContainer := s.GetWebContainer()
	router.Handle("/node-external-worker/{sub_path:.*}", webContainer)
	if err := s.ListenAndServeMux(false); err != nil {
		return fmt.Errorf("http ListenAndServe error %s", err.Error())
	}
	return nil
}

// InitRouters init router
func (server *HttpServerClient) initRouters(ws *restful.WebService) {
	ws.Route(ws.GET("/api/v1/health_check").To(server.healthCheck))
}

func (server *HttpServerClient) healthCheck(request *restful.Request, response *restful.Response) {
	resp := CreateResponseData(nil, "success", server.UUID)
	_, _ = response.Write(resp)
}

// SetUUID set uuid
func (server *HttpServerClient) SetUUID(uUID string) {
	server.UUID = uUID
}
