/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package federated

import (
	"fmt"

	apiproxy "k8s.io/apimachinery/pkg/util/proxy"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/clientutil"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/proxy"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/rest"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/rest/apis"
)

type Handler struct {
	clusterId    string
	members      []string
	proxyHandler *apiproxy.UpgradeAwareHandler
	podHander    *apis.PodHandler
}

// NewHandler create handler
func NewHandler(clusterId string, members []string) (*Handler, error) {
	kubeConf, err := clientutil.GetKubeConfByClusterId(clusterId)
	if err != nil {
		return nil, fmt.Errorf("build proxy handler from config %s failed, err %s", kubeConf.String(), err.Error())
	}

	proxyHandler, err := proxy.NewProxyHandlerFromConfig(kubeConf)
	if err != nil {
		return nil, fmt.Errorf("build proxy handler from config %s failed, err %s", kubeConf.String(), err.Error())
	}

	stor, err := NewPodStor(members)
	if err != nil {
		return nil, err
	}

	podHander := apis.NewPodHandler(stor)

	return &Handler{
		clusterId:    clusterId,
		proxyHandler: proxyHandler,
		members:      members,
		podHander:    podHander,
	}, nil
}

// ServeHTTP serves http request
func (h *Handler) Serve(c *rest.RequestInfo) {
	err := rest.ErrInit

	switch c.Resource {
	case "pods":
		err = h.podHander.Serve(c)
	}

	// 未实现的功能, 使用代理请求
	if err == rest.ErrInit || err == rest.ErrNotImplemented {
		h.proxyHandler.ServeHTTP(c.Writer, c.Request)
		return
	}

	if err != nil {
		c.AbortWithError(err)
		return
	}
}
