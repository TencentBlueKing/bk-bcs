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
	v1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/proxy"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/rest"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/rest/apis"
)

// Handler federated cluster hander
type Handler struct {
	clusterId          string
	members            []string
	proxyHandler       *proxy.ProxyHandler
	podHander          *apis.PodHandler
	deploymentHandler  *apis.DeploymentHandler
	statefulSetHandler *apis.StatefulSetHandler
	serviceHandler     *apis.ServiceHandler
	configmapHandler   *apis.ConfigMapHandler
	secretHandler      *apis.SecretHandler
	eventHandler       *apis.EventHandler
	clusterHandler     *apis.ClusterHandler
}

// NewHandler create federated cluster handler
func NewHandler(clusterId string, members []string) (*Handler, error) {
	proxyHandler, err := proxy.NewProxyHandler(clusterId)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		clusterId:    clusterId,
		proxyHandler: proxyHandler,
		members:      members,
	}

	if err := h.Register(clusterId, members); err != nil {
		return nil, err
	}
	return h, nil
}

// Register 注册对应的资源实现
func (h *Handler) Register(clusterId string, members []string) error {
	stor, err := NewPodStor(members)
	if err != nil {
		return err
	}
	h.podHander = apis.NewPodHandler(stor)

	deployStor, err := NewDeploymentStor(clusterId, members)
	if err != nil {
		return err
	}
	h.deploymentHandler = apis.NewDeploymentHandler(deployStor)

	statefulsetStor, err := NewStatefulSetStor(clusterId, members)
	if err != nil {
		return err
	}
	h.statefulSetHandler = apis.NewStatefulSetHandler(statefulsetStor)

	serviceStor, err := NewServiceStor(clusterId, members)
	if err != nil {
		return err
	}
	h.serviceHandler = apis.NewServiceHandler(serviceStor)

	configMapStor, err := NewConfigMapStor(clusterId, members)
	if err != nil {
		return err
	}
	h.configmapHandler = apis.NewConfigMapHandler(configMapStor)

	secretStor, err := NewSecretStor(clusterId, members)
	if err != nil {
		return err
	}
	h.secretHandler = apis.NewSecretHandler(secretStor)

	eventStor, err := NewEventStor(clusterId, members)
	if err != nil {
		return err
	}
	h.eventHandler = apis.NewEventHandler(eventStor)

	clusterStor, err := NewClusterStor(clusterId, members)
	if err != nil {
		return err
	}
	h.clusterHandler = apis.NewClusterHandler(clusterStor)

	return nil
}

// ServeHTTP serves http request
func (h *Handler) Serve(c *rest.RequestContext) {
	err := rest.ErrInit

	switch c.Resource {
	case string(v1.ResourcePods):
		err = h.podHander.Serve(c)
	case "deployments":
		err = h.deploymentHandler.Serve(c)
	case "statefulsets":
		err = h.statefulSetHandler.Serve(c)
	case "services":
		err = h.serviceHandler.Serve(c)
	case "configmaps":
		err = h.configmapHandler.Serve(c)
	case "secrets":
		err = h.secretHandler.Serve(c)
	case "events":
		err = h.eventHandler.Serve(c)
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
