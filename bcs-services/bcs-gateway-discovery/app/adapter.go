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

package app

import (
	"fmt"
	"strings"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-services/bcs-gateway-discovery/register"
)

var defaultModules = []string{
	types.BCS_MODULE_STORAGE,
	types.BCS_MODULE_MESOSDRIVER,
	types.BCS_MODULE_KUBERNETEDRIVER,
	types.BCS_MODULE_NETWORKDETECTION,
	types.BCS_MODULE_USERMANAGER,
	types.BCS_MODULE_KUBEAGENT,
}

var defaultDomain = "bkbcs.tencent.com"
var defaultClusterIDKey = "BCS-ClusterID"
var defaultMediaTypeKey = "Content-Type"
var defaultAcceptKey = "Accept"
var defaultMediaType = "application/json"

//Handler for module automatic reflection
type Handler func(module string, svcs []*types.ServerInfo) (*register.Service, error)

//NewAdapter create service data convertion
func NewAdapter(option *ServerOptions, standardModules []string) *Adapter {
	adp := &Adapter{
		admintoken: option.AuthToken,
		modules:    standardModules,
		handlers:   make(map[string]Handler),
	}
	//init all service data
	adp.initDefaultModules()
	adp.initAdditionalModules()
	return adp
}

//Adapter uses for converting bkbcs service discovery
// information to inner register data structures
type Adapter struct {
	admintoken string
	modules    []string
	handlers   map[string]Handler
}

//GetService interface for all service data convertion
func (adp *Adapter) GetService(module string, svcs []*types.ServerInfo) (*register.Service, error) {
	resources := strings.Split(module, "/")
	//module name
	handler := adp.handlers[resources[0]]
	if handler == nil {
		blog.Errorf("gateway-discovery didn't register %s handler", module)
		return nil, fmt.Errorf("handle for %s not registe", module)
	}
	return handler(module, svcs)
}

//initDefaultModules init original proxy rule, it's better compatible with originals
func (adp *Adapter) initDefaultModules() error {
	adp.handlers[types.BCS_MODULE_STORAGE] = adp.constructStorage
	blog.Infof("gateway-discovery init module %s proxy rules", types.BCS_MODULE_STORAGE)
	adp.handlers[types.BCS_MODULE_MESOSDRIVER] = adp.constructMesosDriver
	blog.Infof("gateway-discovery init module %s proxy rules", types.BCS_MODULE_MESOSDRIVER)
	adp.handlers[types.BCS_MODULE_KUBERNETEDRIVER] = adp.constructKubeDriver
	blog.Infof("gateway-discovery init module %s proxy rules", types.BCS_MODULE_KUBERNETEDRIVER)
	adp.handlers[types.BCS_MODULE_NETWORKDETECTION] = adp.constructNetworkDetection
	blog.Infof("gateway-discovery init module %s proxy rules", types.BCS_MODULE_NETWORKDETECTION)
	//kube-apiserver information get by userManager
	adp.handlers[types.BCS_MODULE_USERMANAGER] = adp.constructUserMgr
	blog.Infof("gateway-discovery init module %s proxy rules", types.BCS_MODULE_USERMANAGER)
	//kube-apiserver proxy rule is not compatible with bcs-api
	adp.handlers[types.BCS_MODULE_KUBEAGENT] = adp.constructKubeAPIServer
	blog.Infof("gateway-discovery init module %s proxy rules", types.BCS_MODULE_KUBEAGENT)
	return nil
}

func (adp *Adapter) initAdditionalModules() error {
	if len(adp.modules) == 0 {
		return nil
	}
	for _, name := range adp.modules {
		adp.handlers[name] = adp.constructStandardProxy
		blog.Infof("gateway-dicovery init standard proxy handler for module %s", name)
	}
	return nil
}

//constructMesosDriver convert bcs-mesos-driver service information
// to custom service definition. this is compatible with original bcs-api proxy
func (adp *Adapter) constructMesosDriver(module string, svcs []*types.ServerInfo) (*register.Service, error) {
	if len(svcs) == 0 {
		//todo(DeveloperJim): if all service instances down, shall we update proxy rules?
		return nil, fmt.Errorf("ServerInfo lost")
	}
	resources := strings.Split(module, "/")
	if len(resources) != 2 {
		blog.Errorf("contruct MesosDriver server info for %s failed, module name is invalid", module)
		return nil, fmt.Errorf("module information is invalid")
	}
	bkbcsID := strings.Split(resources[1], "-")
	if len(bkbcsID) != 3 {
		blog.Errorf("contruct MesosDriver Server Info failed, ClusterID is invalid in Module name [%s]", module)
		return nil, fmt.Errorf("mesosdriver clusterID is invalid")
	}
	name := types.BCS_MODULE_MESOSDRIVER + "-" + bkbcsID[2]
	hostName := bkbcsID[2] + "." + types.BCS_MODULE_MESOSDRIVER + "." + defaultDomain
	labels := make(map[string]string)
	upcaseID := strings.ToUpper(resources[1])
	labels["module"] = types.BCS_MODULE_MESOSDRIVER
	labels["service"] = "bkbcs-cluster"
	labels["scheduler"] = "mesos"
	labels["cluster"] = upcaseID
	regSvc := &register.Service{
		Name:     name,
		Protocol: svcs[0].Scheme,
		Host:     hostName,
		Path:     "/mesosdriver/v4/",
		Retries:  1,
		Labels:   labels,
	}
	//setting route information
	rt := register.Route{
		Name:     name,
		Protocol: svcs[0].Scheme,
		//* contains mesosdriver & mesoswebconsole proxy rules
		//! path /bcsapi/v1 must maintain until bk-bcs-saas move to new api-gateway
		Paths:       []string{"/bcsapi/v4/scheduler/mesos/", "/bcsapi/v1/"},
		PathRewrite: true,
		Header: map[string]string{
			defaultClusterIDKey: upcaseID,
		},
		Service: name,
		Labels:  labels,
		Plugin: &register.Plugins{
			AuthOption: &register.BCSAuthOption{
				Name: "bkbcs-auth",
				//sending auth request to usermanager.bkbcs.tencent.com
				AuthEndpoints: fmt.Sprintf("https://%s.%s", types.BCS_MODULE_USERMANAGER, defaultDomain),
				AuthToken:     adp.admintoken,
				Module:        types.BCS_MODULE_MESOSDRIVER,
			},
		},
	}
	regSvc.Routes = append(regSvc.Routes, rt)
	//setting upstream backend information
	bcks := adp.constructBackends(svcs)
	regSvc.Backends = append(regSvc.Backends, bcks...)
	return regSvc, nil
}

//constructKubeDriver convert bcs-kube-driver service information
// to custom service definition. this is compatible with original bcs-api proxy
func (adp *Adapter) constructKubeDriver(module string, svcs []*types.ServerInfo) (*register.Service, error) {
	if len(svcs) == 0 {
		//todo(DeveloperJim): if all service instances down, shall we update proxy rules?
		return nil, fmt.Errorf("ServerInfo lost")
	}
	resources := strings.Split(module, "/")
	if len(resources) != 2 {
		blog.Errorf("contruct KubeDriver server info for %s failed, module name is invalid", module)
		return nil, fmt.Errorf("module information is invalid")
	}
	bkbcsID := strings.Split(resources[1], "-")
	if len(bkbcsID) != 3 {
		blog.Errorf("contruct KubeDriver Server Info failed, ClusterID is invalid in Module name [%s]", module)
		return nil, fmt.Errorf("kubedriver clusterID is invalid")
	}
	name := fmt.Sprintf("%s-%s", types.BCS_MODULE_KUBERNETEDRIVER, bkbcsID[2])
	hostName := fmt.Sprintf("%s.%s.%s", bkbcsID[2], types.BCS_MODULE_KUBERNETEDRIVER, defaultDomain)
	labels := make(map[string]string)
	upcaseID := strings.ToUpper(resources[1])
	labels["module"] = types.BCS_MODULE_KUBERNETEDRIVER
	labels["service"] = "bkbcs-cluster"
	labels["scheduler"] = "kubernetes"
	labels["cluster"] = upcaseID
	regSvc := &register.Service{
		Name:     name,
		Protocol: svcs[0].Scheme,
		Host:     hostName,
		Path:     "/k8sdriver/v4/",
		Retries:  1,
		Labels:   labels,
	}
	//setting route information
	rt := register.Route{
		Name:        name,
		Protocol:    svcs[0].Scheme,
		Paths:       []string{"/bcsapi/v4/scheduler/k8s/"},
		PathRewrite: true,
		Header: map[string]string{
			defaultClusterIDKey: upcaseID,
		},
		Service: name,
		Labels:  labels,
		Plugin: &register.Plugins{
			AuthOption: &register.BCSAuthOption{
				Name: "bkbcs-auth",
				//sending auth request to usermanager.bkbcs.tencent.com
				AuthEndpoints: fmt.Sprintf("https://%s.%s", types.BCS_MODULE_USERMANAGER, defaultDomain),
				AuthToken:     adp.admintoken,
				Module:        types.BCS_MODULE_KUBERNETEDRIVER,
			},
		},
	}
	regSvc.Routes = append(regSvc.Routes, rt)
	//setting upstream backend information
	bcks := adp.constructBackends(svcs)
	regSvc.Backends = append(regSvc.Backends, bcks...)
	return regSvc, nil
}

//constructStorage convert bcs-storage service information
// to custom service definition. this is compatible with original bcs-api proxy
func (adp *Adapter) constructStorage(module string, svcs []*types.ServerInfo) (*register.Service, error) {
	if len(svcs) == 0 {
		//todo(DeveloperJim): if all service instances down, shall we update proxy rules?
		return nil, fmt.Errorf("ServerInfo lost")
	}
	hostName := fmt.Sprintf("%s.%s", types.BCS_MODULE_STORAGE, defaultDomain)
	labels := make(map[string]string)
	labels["module"] = types.BCS_MODULE_STORAGE
	labels["service"] = "bkbcs-service"
	regSvc := &register.Service{
		Name:     module,
		Protocol: svcs[0].Scheme,
		Host:     hostName,
		Path:     "/bcsstorage/v1/",
		Retries:  1,
		Labels:   labels,
	}
	//setting route information
	rt := register.Route{
		Name:        module,
		Protocol:    svcs[0].Scheme,
		Paths:       []string{"/bcsapi/v4/storage/"},
		PathRewrite: true,
		Service:     module,
		Labels:      labels,
		Plugin: &register.Plugins{
			AuthOption: &register.BCSAuthOption{
				Name: "bkbcs-auth",
				//sending auth request to usermanager.bkbcs.tencent.com
				AuthEndpoints: fmt.Sprintf("https://%s.%s", types.BCS_MODULE_USERMANAGER, defaultDomain),
				AuthToken:     adp.admintoken,
				Module:        types.BCS_MODULE_STORAGE,
			},
		},
	}
	regSvc.Routes = append(regSvc.Routes, rt)
	//setting upstream backend information
	bcks := adp.constructBackends(svcs)
	regSvc.Backends = append(regSvc.Backends, bcks...)
	return regSvc, nil
}

//constructKubeAPIServer convert kube-apiserver service information
// to custom service definition. this is `not` compatible with original bcs-api proxy
//! @param svc instance plays a trick, it's field HostName holding token from kubeagent
func (adp *Adapter) constructKubeAPIServer(module string, svcs []*types.ServerInfo) (*register.Service, error) {
	if len(svcs) == 0 {
		//todo(DeveloperJim): if all service instances down, shall we update proxy rules?
		return nil, fmt.Errorf("ServerInfo lost")
	}
	resources := strings.Split(module, "/")
	if len(resources) != 2 {
		blog.Errorf("contruct Kube-apiserver info for %s failed, module name is invalid", module)
		return nil, fmt.Errorf("module information is invalid")
	}
	bkbcsID := strings.Split(resources[1], "-")
	if len(bkbcsID) != 3 {
		blog.Errorf("contruct Kube-apiServer Info failed, ClusterID is invalid in Module name [%s]", module)
		return nil, fmt.Errorf("kubeagent clusterID is invalid")
	}
	name := types.BCS_MODULE_KUBEAGENT + "-" + bkbcsID[2]
	hostName := bkbcsID[2] + "." + types.BCS_MODULE_KUBEAGENT + "." + defaultDomain
	labels := make(map[string]string)
	upcaseID := strings.ToUpper(resources[1])
	labels["module"] = types.BCS_MODULE_KUBEAGENT
	labels["service"] = "bkbcs-cluster"
	labels["scheduler"] = "kubernetes"
	labels["cluster"] = upcaseID
	//create service & setting header plugin for kube-apiserver
	regSvc := &register.Service{
		Name:     name,
		Protocol: svcs[0].Scheme,
		Host:     hostName,
		Path:     "/",
		Retries:  1,
		Plugin: &register.Plugins{
			HeadOption: &register.HeaderOption{
				Clean: []string{"Authorization"},
				Add: map[string]string{
					"Authorization": fmt.Sprintf("Bearer %s", svcs[0].HostName),
				},
			},
		},
		Labels: labels,
	}

	//setting route information
	rt := register.Route{
		Name:        name,
		Protocol:    svcs[0].Scheme,
		Paths:       []string{fmt.Sprintf("/tunnels/clusters/%s/", upcaseID)},
		PathRewrite: true,
		Service:     name,
		Plugin: &register.Plugins{
			AuthOption: &register.BCSAuthOption{
				Name: "bkbcs-auth",
				//sending auth request to usermanager.bkbcs.tencent.com
				AuthEndpoints: fmt.Sprintf("https://%s.%s", types.BCS_MODULE_USERMANAGER, defaultDomain),
				AuthToken:     adp.admintoken,
				Module:        types.BCS_MODULE_KUBEAGENT,
			},
		},
		Labels: labels,
	}
	regSvc.Routes = append(regSvc.Routes, rt)
	//setting upstream backend information
	bcks := adp.constructBackends(svcs)
	regSvc.Backends = append(regSvc.Backends, bcks...)
	return regSvc, nil
}

//constructClusterMgr convert bcs-cluster-manager service information
// to custom service definition. this is compatible with original bcs-api proxy.
// and further more, api-gateway defines new standard proxy rule for it
func (adp *Adapter) constructUserMgr(module string, svcs []*types.ServerInfo) (*register.Service, error) {
	if len(svcs) == 0 {
		//todo(DeveloperJim): if all service instances down, shall we update proxy rules?
		return nil, fmt.Errorf("ServerInfo lost")
	}
	hostName := fmt.Sprintf("%s.%s", types.BCS_MODULE_USERMANAGER, defaultDomain)
	labels := make(map[string]string)
	labels["module"] = types.BCS_MODULE_USERMANAGER
	labels["service"] = "bkbcs-service"
	regSvc := &register.Service{
		Name:     module,
		Protocol: svcs[0].Scheme,
		Host:     hostName,
		Path:     fmt.Sprintf("/%s/", types.BCS_MODULE_USERMANAGER),
		Retries:  1,
		Labels:   labels,
	}
	//setting route information
	rt := register.Route{
		Name:        module,
		Protocol:    svcs[0].Scheme,
		Paths:       []string{fmt.Sprintf("/bcsapi/v4/%s/", types.BCS_MODULE_USERMANAGER)},
		PathRewrite: true,
		Service:     module,
		Labels:      labels,
	}
	regSvc.Routes = append(regSvc.Routes, rt)
	//setting upstream backend information
	bcks := adp.constructBackends(svcs)
	regSvc.Backends = append(regSvc.Backends, bcks...)
	return regSvc, nil
}

//constructClusterMgr convert bcs-cluster-manager service information
// to custom service definition. this is compatible with original bcs-api proxy.
// and further more, api-gateway defines new standard proxy rule for it
func (adp *Adapter) constructNetworkDetection(module string, svcs []*types.ServerInfo) (*register.Service, error) {
	if len(svcs) == 0 {
		//todo(DeveloperJim): if all service instances down, shall we update proxy rules?
		return nil, fmt.Errorf("ServerInfo lost")
	}
	hostName := fmt.Sprintf("%s.%s", types.BCS_MODULE_NETWORKDETECTION, defaultDomain)
	labels := make(map[string]string)
	labels["module"] = types.BCS_MODULE_NETWORKDETECTION
	labels["service"] = "bkbcs-service"
	regSvc := &register.Service{
		Name:     module,
		Protocol: svcs[0].Scheme,
		Host:     hostName,
		Path:     "/detection/v4/",
		Retries:  1,
		Labels:   labels,
	}
	//setting route information
	rt := register.Route{
		Name:        module,
		Protocol:    svcs[0].Scheme,
		Paths:       []string{"/bcsapi/v4/detection/"},
		PathRewrite: true,
		Plugin: &register.Plugins{
			AuthOption: &register.BCSAuthOption{
				Name: "bkbcs-auth",
				//sending auth request to usermanager.bkbcs.tencent.com
				AuthEndpoints: fmt.Sprintf("https://%s.%s", types.BCS_MODULE_USERMANAGER, defaultDomain),
				AuthToken:     adp.admintoken,
				Module:        types.BCS_MODULE_NETWORKDETECTION,
			},
		},
		Service: module,
		Labels:  labels,
	}
	regSvc.Routes = append(regSvc.Routes, rt)
	//setting upstream backend information
	bcks := adp.constructBackends(svcs)
	regSvc.Backends = append(regSvc.Backends, bcks...)
	return regSvc, nil
}

//constructStandardProxy convert standard service information
// to custom service definition. standard proxy rule is below:
// /bcsapi/v4/{module}/ ==> /{module}/
func (adp *Adapter) constructStandardProxy(module string, svcs []*types.ServerInfo) (*register.Service, error) {
	if len(svcs) == 0 {
		//todo(DeveloperJim): if all service instances down, shall we update proxy rules?
		return nil, fmt.Errorf("ServerInfo lost")
	}
	hostName := fmt.Sprintf("%s.%s", module, defaultDomain)
	labels := make(map[string]string)
	labels["module"] = module
	labels["service"] = "bkbcs-service"
	regSvc := &register.Service{
		Name:     module,
		Protocol: svcs[0].Scheme,
		Host:     hostName,
		Path:     fmt.Sprintf("/%s/v4/", module),
		Retries:  1,
		Labels:   labels,
	}
	//setting route information
	apipath := fmt.Sprintf("/bcsapi/v4/%s/", module)
	rt := register.Route{
		Name:        module,
		Protocol:    svcs[0].Scheme,
		Paths:       []string{apipath},
		PathRewrite: true,
		Service:     module,
		Labels:      labels,
		Plugin: &register.Plugins{
			AuthOption: &register.BCSAuthOption{
				Name: "bkbcs-auth",
				//sending auth request to usermanager.bkbcs.tencent.com
				AuthEndpoints: fmt.Sprintf("https://%s.%s", types.BCS_MODULE_USERMANAGER, defaultDomain),
				AuthToken:     adp.admintoken,
				Module:        module,
			},
		},
	}
	regSvc.Routes = append(regSvc.Routes, rt)
	//setting upstream backend information
	bcks := adp.constructBackends(svcs)
	regSvc.Backends = append(regSvc.Backends, bcks...)
	return regSvc, nil
}

func (adp *Adapter) constructBackends(svcs []*types.ServerInfo) []register.Backend {
	var backends []register.Backend
	for _, svc := range svcs {
		var target string
		//todo(DeveloperJim): support ipv6 feature
		if svc.ExternalIp != "" && svc.ExternalPort != 0 {
			target = fmt.Sprintf("%s:%d", svc.ExternalIp, svc.ExternalPort)
		} else {
			//inner ipaddress
			target = fmt.Sprintf("%s:%d", svc.IP, svc.Port)
		}
		back := register.Backend{
			Target: target,
			Weight: 100,
		}
		backends = append(backends, back)
	}
	return backends
}
