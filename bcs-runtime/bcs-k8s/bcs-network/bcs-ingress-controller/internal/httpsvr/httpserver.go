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

// Package httpsvr 对外提供http接口（通过service域名访问）
package httpsvr

import (
	"github.com/emicklei/go-restful"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud/aws"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/generator"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/nodecache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/option"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/portbindingcontroller"
)

// HttpServerClient http server client
type HttpServerClient struct {
	Mgr manager.Manager

	NodeCache         *nodecache.NodeCache
	NodePortBindCache *portbindingcontroller.NodePortBindingCache

	IngressLiConverter *generator.IngressListenerConverter
	AgaSupporter       *aws.AgaSupporter

	Ops *option.ControllerOption
}

// InitRouters init router
func InitRouters(ws *restful.WebService, httpServerClient *HttpServerClient) {
	ws.Route(ws.GET("/api/v1/ingresss").To(httpServerClient.listIngress))
	ws.Route(ws.GET("/api/v1/portpools").To(httpServerClient.listPortPool))
	ws.Route(ws.GET("/api/v1/nodeportbindings").To(httpServerClient.getNodePortBindings))
	ws.Route(ws.GET("/api/v1/listeners/{condition}/{namespace}/{name}").To(httpServerClient.listListener))

	ws.Route(ws.GET("/api/v1/node").To(httpServerClient.listNode))
	ws.Route(ws.GET("/api/v1/aga_entrance").To(httpServerClient.getPodRelatedAgaEntrance))

	ws.Route(ws.GET("/api/v1/check_bind_status").To(httpServerClient.CheckBindStatus))

	ws.Route(ws.GET("/readiness_probe").To(httpServerClient.readinessProbe))

}
