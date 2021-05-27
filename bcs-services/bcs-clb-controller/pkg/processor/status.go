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

package processor

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	clbingressType "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/clb/v1"
	cloudListenerType "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/network/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/serviceclient"

	"github.com/emicklei/go-restful"
)

// RemoteListenerStatus remote listener status
type RemoteListenerStatus struct {
	Name string
	cloudListenerType.CloudListenerSpec
	cloudListenerType.CloudListenerStatus
}

// CloudListenerStatus cloud listener status
type CloudListenerStatus struct {
	Name      string
	Namespace string
	cloudListenerType.CloudListenerSpec
}

// ClbIngressStatus clb ingress status
type ClbIngressStatus struct {
	Name      string
	Namespace string
	clbingressType.ClbIngressSpec
	clbingressType.ClbIngressStatus
	AppSvcs map[string]*serviceclient.AppService
}

// Status status
type Status struct {
	Ingresses       []*ClbIngressStatus     `json:"ingresses"`
	CloudListeners  []*CloudListenerStatus  `json:"cloudListeners"`
	RemoteListeners []*RemoteListenerStatus `json:"remoteListeners"`
}

const (
	// StatusCodeSuccess code for success status
	StatusCodeSuccess = 0
	// StatusCodeError code for error status
	StatusCodeError = 1
)

// StatusResponse status response
type StatusResponse struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    *Status `json:"data,omitempty"`
}

// GetStatusFunction get status function
func (p *Processor) GetStatusFunction() restful.RouteFunction {
	return p.status
}

func (p *Processor) status(req *restful.Request, resp *restful.Response) {

	response := &StatusResponse{}
	defer resp.WriteEntity(response)

	ingresses, err := p.ingressRegistry.ListIngresses()
	if err != nil {
		blog.Warnf("list ingresses err %s", err.Error())
		response.Code = StatusCodeError
		response.Message = "list ingresses failed"
		return
	}
	listeners, err := p.listenerClient.ListListeners()
	if err != nil {
		blog.Warnf("list listener failed, err %s", err.Error())
		response.Code = StatusCodeError
		response.Message = "list listener failed"
		return
	}
	remoteListeners, err := p.updater.ListRemoteListener()
	if err != nil {
		blog.Warnf("list remote listener failed, err %s", err.Error())
	}
	var retIngresses []*ClbIngressStatus
	for _, ingress := range ingresses {
		retAppServices := make(map[string]*serviceclient.AppService)
		if len(ingress.Spec.HTTP) != 0 {
			for _, http := range ingress.Spec.HTTP {
				appSvc, err := p.serviceClient.GetAppService(http.Namespace, http.ServiceName)
				if err != nil {
					continue
				}
				retAppServices[http.Namespace+"/"+http.ServiceName] = appSvc
			}
		}
		if len(ingress.Spec.HTTPS) != 0 {
			for _, https := range ingress.Spec.HTTPS {
				appSvc, err := p.serviceClient.GetAppService(https.Namespace, https.ServiceName)
				if err != nil {
					continue
				}
				retAppServices[https.Namespace+"/"+https.ServiceName] = appSvc
			}
		}
		if len(ingress.Spec.TCP) != 0 {
			for _, tcp := range ingress.Spec.TCP {
				appSvc, err := p.serviceClient.GetAppService(tcp.Namespace, tcp.ServiceName)
				if err != nil {
					continue
				}
				retAppServices[tcp.Namespace+"/"+tcp.ServiceName] = appSvc
			}
		}
		if len(ingress.Spec.UDP) != 0 {
			for _, udp := range ingress.Spec.UDP {
				appSvc, err := p.serviceClient.GetAppService(udp.Namespace, udp.ServiceName)
				if err != nil {
					continue
				}
				retAppServices[udp.Namespace+"/"+udp.ServiceName] = appSvc
			}
		}

		retIngresses = append(retIngresses, &ClbIngressStatus{
			Name:             ingress.GetName(),
			Namespace:        ingress.GetNamespace(),
			ClbIngressSpec:   ingress.Spec,
			ClbIngressStatus: ingress.Status,
			AppSvcs:          retAppServices,
		})
	}
	var retListeners []*CloudListenerStatus
	for _, listener := range listeners {
		retListeners = append(retListeners, &CloudListenerStatus{
			Name:              listener.GetName(),
			Namespace:         listener.GetNamespace(),
			CloudListenerSpec: listener.Spec,
		})
	}
	var retRemoteListeners []*RemoteListenerStatus
	for _, remoteListener := range remoteListeners {
		retRemoteListeners = append(retRemoteListeners, &RemoteListenerStatus{
			Name:                remoteListener.GetName(),
			CloudListenerSpec:   remoteListener.Spec,
			CloudListenerStatus: remoteListener.Status,
		})
	}

	response.Data = &Status{
		Ingresses:       retIngresses,
		CloudListeners:  retListeners,
		RemoteListeners: retRemoteListeners,
	}

}
