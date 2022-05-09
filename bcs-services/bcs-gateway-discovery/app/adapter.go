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
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/modules"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/register"

	"github.com/micro/go-micro/v2/registry"
)

var metricModeule = "metric"

var notStandardRouteModules = map[string]string{
	modules.BCSModuleStorage:          "storage",
	modules.BCSModuleMesosdriver:      "mesosdriver",
	modules.BCSModuleNetworkdetection: "networkdetection",
	modules.BCSModuleKubeagent:        "kubeagent",
	modules.BCSModuleUserManager:      "usermanager",
	metricModeule:                     metricModeule,
}

var defaultModules = []string{}

var defaultGrpcModules = map[string]string{
	"logmanager":  "LogManager",
	"meshmanager": "MeshManager",
}

var defaultHTTPModules = map[string]string{
	"logmanager":  "LogManager",
	"meshmanager": "MeshManager",
}

var defaultDomain = ".bkbcs.tencent.com"
var defaultClusterIDKey = "BCS-ClusterID"
var defaultMediaTypeKey = "Content-Type"
var defaultAcceptKey = "Accept"
var defaultMediaType = "application/json"
var defaultClusterName = "bkbcs-cluster"
var defaultServiceTag = "bkbcs-service"
var defaultPluginName = "bkbcs-auth"

//Handler for module automatic reflection
//@param: module, bkbcs module name, like usermanager, logmanager
//@param: svcs, bkbcs service instance definition
type Handler func(module string, svcs []*types.ServerInfo) (*register.Service, error)

//MicroHandler compatible for old module that registe in micro registry
type MicroHandler func(module string, svc *registry.Service) (*register.Service, error)

//NewAdapter create service data conversion
func NewAdapter(option *ServerOptions) *Adapter {
	adp := &Adapter{
		admintoken:    option.AuthToken,
		handlers:      make(map[string]Handler),
		microHandlers: make(map[string]MicroHandler),
	}
	//init all service data
	adp.initDefaultModules()
	adp.initCompatibleMicroModules()
	return adp
}

//Adapter uses for converting bkbcs service discovery
// information to inner register data structures
type Adapter struct {
	admintoken    string
	handlers      map[string]Handler
	microHandlers map[string]MicroHandler
}

//GetService interface for all service data conversion
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

//GetGrpcService interface for go-micro grpc module data conversion
//@param: module, all kind module name, such as logmanager, usermanager
//@param: svc, go-micro service definition, came form etcd registry
func (adp *Adapter) GetGrpcService(module string, svc *registry.Service) (*register.Service, error) {
	//get grpc Service Interface name
	interfaceName, ok := defaultGrpcModules[module]
	if !ok {
		return nil, fmt.Errorf("module %s do not registe", module)
	}
	// actual registered name & grpc proxy path
	actualName := fmt.Sprintf("%s-grpc", module)
	requestPath := fmt.Sprintf("/%s.%s/", module, interfaceName)
	hostName := fmt.Sprintf("%s%s", actualName, defaultDomain)
	labels := make(map[string]string)
	labels["module"] = module
	regSvc := &register.Service{
		Name:     actualName,
		Protocol: "grpcs",
		Host:     hostName,
		Retries:  1,
		Labels:   labels,
	}
	//setting route information
	rt := register.Route{
		Name:     actualName,
		Protocol: "grpcs",
		//grpc path proxy rule likes /logmanager.LogManager/
		Paths:       []string{requestPath},
		PathRewrite: false,
		Plugin: &register.Plugins{
			AuthOption: &register.BCSAuthOption{
				Name: defaultPluginName,
				//sending auth request to usermanager.bkbcs.tencent.com
				AuthEndpoints: fmt.Sprintf("https://%s%s", modules.BCSModuleUserManager, defaultDomain),
				AuthToken:     adp.admintoken,
				Module:        module,
			},
		},
		Service: actualName,
		Labels:  labels,
	}
	regSvc.Routes = append(regSvc.Routes, rt)
	//setting upstream backend information
	bcks := adp.constructUpstreamTarget(svc.Nodes)
	regSvc.Backends = append(regSvc.Backends, bcks...)
	return regSvc, nil
}

//GetHTTPService interface for go-micro http module data conversion
// only support standard new module api registration
func (adp *Adapter) GetHTTPService(module string, svc *registry.Service) (*register.Service, error) {
	found := false
	for info := range defaultHTTPModules {
		if module == info {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("go-micro http module %s do not registe", module)
	}
	if _, ok := notStandardRouteModules[module]; ok {
		return adp.microNotStandarModule(module, svc)
	}
	return adp.microStandarModule(module, svc)
}

// microStandarModule
func (adp *Adapter) microStandarModule(module string, svc *registry.Service) (*register.Service, error) {
	// actual registered name & grpc proxy path
	actualName := fmt.Sprintf("%s-http", module)
	//route path
	requestPath := fmt.Sprintf("/%s/", module)
	gatewayPath := fmt.Sprintf("/bcsapi/v4/%s/", module)
	hostName := fmt.Sprintf("%s%s", actualName, defaultDomain)
	labels := make(map[string]string)
	labels["module"] = module
	regSvc := &register.Service{
		Name:     actualName,
		Protocol: "https",
		Host:     hostName,
		Path:     requestPath,
		Retries:  1,
		Labels:   labels,
	}
	//setting route information
	rt := register.Route{
		Name:        actualName,
		Protocol:    "https",
		Paths:       []string{gatewayPath},
		PathRewrite: true,
		Plugin: &register.Plugins{
			AuthOption: &register.BCSAuthOption{
				Name: defaultPluginName,
				//sending auth request to usermanager.bkbcs.tencent.com
				AuthEndpoints: fmt.Sprintf("https://%s%s", modules.BCSModuleUserManager, defaultDomain),
				AuthToken:     adp.admintoken,
				Module:        module,
			},
		},
		Service: actualName,
		Labels:  labels,
	}
	regSvc.Routes = append(regSvc.Routes, rt)
	//setting upstream backend information
	var httpNodes []*registry.Node
	for _, node := range svc.Nodes {
		hostport := strings.Split(node.Address, ":")
		if len(hostport) != 2 {
			blog.Errorf("standard http module %s address formation error, mis-match with ip:port(%s), ID: %s", actualName, node.Address, node.Id)
			return nil, fmt.Errorf("node ip:port formation error")
		}
		grpcport, err := strconv.Atoi(hostport[1])
		if err != nil {
			blog.Errorf("http module %s node %s port is not integer. original %s", actualName, node.Id, node.Address)
			return nil, fmt.Errorf("node port is not integer")
		}
		//go-micro http port definition
		httpport := grpcport - 1
		newNode := &registry.Node{
			Id:       node.Id,
			Address:  fmt.Sprintf("%s:%d", hostport[0], httpport),
			Metadata: node.Metadata,
		}
		httpNodes = append(httpNodes, newNode)
	}
	bcks := adp.constructUpstreamTarget(httpNodes)
	regSvc.Backends = append(regSvc.Backends, bcks...)
	return regSvc, nil
}

// microNotStandarModule these modules are not standard grpc server, here is for compatible purpose
//@param: bkbcs module information
func (adp *Adapter) microNotStandarModule(module string, svc *registry.Service) (*register.Service, error) {
	handler := adp.microHandlers[module]
	if handler == nil {
		blog.Errorf("gateway-discovery didn't register %s micro handler", module)
		return nil, fmt.Errorf("handle for %s not registe", module)
	}
	return handler(module, svc)
}

//initDefaultModules init original proxy rule, it's better compatible with originals
func (adp *Adapter) initDefaultModules() error {
	adp.handlers[modules.BCSModuleStorage] = adp.constructStorage
	blog.Infof("gateway-discovery init module %s proxy rules", modules.BCSModuleStorage)
	adp.handlers[modules.BCSModuleMesosdriver] = adp.constructMesosDriver
	blog.Infof("gateway-discovery init module %s proxy rules", modules.BCSModuleMesosdriver)
	//kube-apiserver information get by userManager
	adp.handlers[modules.BCSModuleUserManager] = adp.constructUserMgr
	blog.Infof("gateway-discovery init module %s proxy rules", modules.BCSModuleUserManager)
	//kube-apiserver proxy rule is not compatible with bcs-api
	adp.handlers[modules.BCSModuleKubeagent] = adp.constructKubeAPIServer
	blog.Infof("gateway-discovery init module %s proxy rules", modules.BCSModuleKubeagent)
	return nil
}

//initDefaultModules init original proxy rule, it's better compatible with originals
func (adp *Adapter) initCompatibleMicroModules() error {
	adp.microHandlers[modules.BCSModuleStorage] = adp.microStorage
	blog.Infof("gateway-discovery init compatible module %s proxy rules", modules.BCSModuleStorage)
	adp.microHandlers[modules.BCSModuleMesosdriver] = adp.microMesosDriver
	blog.Infof("gateway-discovery init compatible module %s proxy rules", modules.BCSModuleMesosdriver)
	//kube-apiserver information get by userManager
	adp.microHandlers[modules.BCSModuleUserManager] = adp.microUserMgr
	blog.Infof("gateway-discovery init compatible module %s proxy rules", modules.BCSModuleUserManager)
	//metricservice information
	adp.microHandlers["metric"] = adp.microMetricService
	return nil
}

//constructMesosDriver convert bcs-mesos-driver service information
// to custom service definition. this is compatible with original bcs-api proxy
func (adp *Adapter) constructMesosDriver(module string, svcs []*types.ServerInfo) (*register.Service, error) {
	if len(svcs) == 0 {
		//if all service instances down, just keep it what it used to be
		// and wait until remote storage node re-registe
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
	name := modules.BCSModuleMesosdriver + "-" + bkbcsID[2]
	hostName := bkbcsID[2] + "." + modules.BCSModuleMesosdriver + defaultDomain
	labels := make(map[string]string)
	upcaseID := strings.ToUpper(resources[1])
	labels["module"] = modules.BCSModuleMesosdriver
	labels["service"] = defaultClusterName
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
				Name: defaultPluginName,
				//sending auth request to usermanager.bkbcs.tencent.com
				AuthEndpoints: fmt.Sprintf("https://%s%s", modules.BCSModuleUserManager, defaultDomain),
				AuthToken:     adp.admintoken,
				Module:        modules.BCSModuleMesosdriver,
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
		//if all service instances down, just keep it what it used to be
		// and wait until remote storage node re-registe
		return nil, fmt.Errorf("ServerInfo lost")
	}
	hostName := fmt.Sprintf("%s%s", modules.BCSModuleStorage, defaultDomain)
	labels := make(map[string]string)
	labels["module"] = modules.BCSModuleStorage
	labels["service"] = defaultServiceTag
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
				Name: defaultPluginName,
				//sending auth request to usermanager.bkbcs.tencent.com
				AuthEndpoints: fmt.Sprintf("https://%s%s", modules.BCSModuleUserManager, defaultDomain),
				AuthToken:     adp.admintoken,
				Module:        modules.BCSModuleStorage,
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
		//if all service instances down, just keep it what it used to be
		// and wait until remote storage node re-registe
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
	name := modules.BCSModuleKubeagent + "-" + bkbcsID[2]
	hostName := bkbcsID[2] + "." + modules.BCSModuleKubeagent + defaultDomain
	labels := make(map[string]string)
	upcaseID := strings.ToUpper(resources[1])
	labels["module"] = modules.BCSModuleKubeagent
	labels["service"] = defaultClusterName
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
		Paths:       []string{fmt.Sprintf("/clusters/%s/", upcaseID)},
		PathRewrite: true,
		Service:     name,
		Plugin: &register.Plugins{
			AuthOption: &register.BCSAuthOption{
				Name: defaultPluginName,
				//sending auth request to usermanager.bkbcs.tencent.com
				AuthEndpoints: fmt.Sprintf("https://%s%s", modules.BCSModuleUserManager, defaultDomain),
				AuthToken:     adp.admintoken,
				Module:        modules.BCSModuleKubeagent,
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
		//if all service instances down, just keep it what it used to be
		// and wait until remote storage node re-registe
		return nil, fmt.Errorf("ServerInfo lost")
	}
	hostName := fmt.Sprintf("%s%s", modules.BCSModuleUserManager, defaultDomain)
	labels := make(map[string]string)
	labels["module"] = modules.BCSModuleUserManager
	labels["service"] = defaultServiceTag
	regSvc := &register.Service{
		Name:     module,
		Protocol: svcs[0].Scheme,
		Host:     hostName,
		Path:     fmt.Sprintf("/%s/", modules.BCSModuleUserManager),
		Retries:  1,
		Labels:   labels,
	}
	//setting route information
	rt := register.Route{
		Name:        module,
		Protocol:    svcs[0].Scheme,
		Paths:       []string{fmt.Sprintf("/bcsapi/v4/%s/", modules.BCSModuleUserManager)},
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
		//if all service instances down, just keep it what it used to be
		// and wait until remote storage node re-registe
		return nil, fmt.Errorf("ServerInfo lost")
	}
	hostName := fmt.Sprintf("%s%s", modules.BCSModuleNetworkdetection, defaultDomain)
	labels := make(map[string]string)
	labels["module"] = modules.BCSModuleNetworkdetection
	labels["service"] = defaultServiceTag
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
				Name: defaultPluginName,
				//sending auth request to usermanager.bkbcs.tencent.com
				AuthEndpoints: fmt.Sprintf("https://%s%s", modules.BCSModuleUserManager, defaultDomain),
				AuthToken:     adp.admintoken,
				Module:        modules.BCSModuleNetworkdetection,
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

func (adp *Adapter) constructBackends(svcs []*types.ServerInfo) []register.Backend {
	var backends []register.Backend
	for _, svc := range svcs {
		var target string
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

func (adp *Adapter) constructUpstreamTarget(nodes []*registry.Node) []register.Backend {
	var backends []register.Backend
	for _, node := range nodes {
		back := register.Backend{
			Target: node.Address,
			Weight: 100,
		}
		backends = append(backends, back)
	}
	return backends
}

//********************************************
// micro registry implementation
//********************************************

//microMesosDriver convert bcs-mesos-driver service information
// to custom service definition. this is compatible with original bcs-api proxy
//@param: module, bkbcs module name without clusterID, like meshmanager, storage, mesosdriver
//@param: svc, micro registry information
func (adp *Adapter) microMesosDriver(module string, svc *registry.Service) (*register.Service, error) {
	//route path
	items := strings.Split(svc.Name, ".")
	name := module + "-" + items[0]
	hostName := svc.Name
	labels := make(map[string]string)
	labels["module"] = module
	labels["service"] = defaultClusterName
	labels["scheduler"] = "mesos"
	upcaseID := fmt.Sprintf("BCS-MESOS-%s", items[0])
	labels["cluster"] = upcaseID
	regSvc := &register.Service{
		Name:     name,
		Protocol: "https",
		Host:     hostName,
		Path:     "/mesosdriver/v4/",
		Retries:  1,
		Labels:   labels,
	}
	//setting route information
	rt := register.Route{
		Name:        name,
		Protocol:    "http",
		Paths:       []string{"/bcsapi/v4/scheduler/mesos/", "/bcsapi/v1/"},
		PathRewrite: true,
		Header: map[string]string{
			defaultClusterIDKey: upcaseID,
		},
		Plugin: &register.Plugins{
			AuthOption: &register.BCSAuthOption{
				Name: defaultPluginName,
				//sending auth request to usermanager.bkbcs.tencent.com
				AuthEndpoints: fmt.Sprintf("https://%s%s", modules.BCSModuleUserManager, defaultDomain),
				AuthToken:     adp.admintoken,
				Module:        module,
			},
		},
		Service: name,
		Labels:  labels,
	}
	regSvc.Routes = append(regSvc.Routes, rt)
	//setting upstream backend information
	bcks := adp.constructUpstreamTarget(svc.Nodes)
	regSvc.Backends = append(regSvc.Backends, bcks...)
	return regSvc, nil
}

//microStorage convert bcs-storage service information
// to custom service definition. this is compatible with original bcs-api proxy
func (adp *Adapter) microStorage(module string, svc *registry.Service) (*register.Service, error) {
	labels := make(map[string]string)
	labels["module"] = modules.BCSModuleStorage
	labels["service"] = defaultServiceTag
	regSvc := &register.Service{
		Name:     module,
		Protocol: "https",
		Host:     svc.Name,
		Path:     "/bcsstorage/v1/",
		Retries:  1,
		Labels:   labels,
	}
	//setting route information
	rt := register.Route{
		Name:        module,
		Protocol:    "http",
		Paths:       []string{"/bcsapi/v4/storage/"},
		PathRewrite: true,
		Service:     module,
		Labels:      labels,
		Plugin: &register.Plugins{
			AuthOption: &register.BCSAuthOption{
				Name: defaultPluginName,
				//sending auth request to usermanager.bkbcs.tencent.com
				AuthEndpoints: fmt.Sprintf("https://%s%s", modules.BCSModuleUserManager, defaultDomain),
				AuthToken:     adp.admintoken,
				Module:        modules.BCSModuleStorage,
			},
		},
	}
	regSvc.Routes = append(regSvc.Routes, rt)
	//setting upstream backend information
	bcks := adp.constructUpstreamTarget(svc.Nodes)
	regSvc.Backends = append(regSvc.Backends, bcks...)
	return regSvc, nil
}

//microUserMgr convert bcs-cluster-manager service information
// to custom service definition. this is compatible with original bcs-api proxy.
// and further more, api-gateway defines new standard proxy rule for it
func (adp *Adapter) microUserMgr(module string, svc *registry.Service) (*register.Service, error) {
	labels := make(map[string]string)
	labels["module"] = modules.BCSModuleUserManager
	labels["service"] = defaultServiceTag
	regSvc := &register.Service{
		Name:     module,
		Protocol: "https",
		Host:     svc.Name,
		Path:     fmt.Sprintf("/%s/", modules.BCSModuleUserManager),
		Retries:  1,
		Labels:   labels,
	}
	//setting route information
	rt := register.Route{
		Name:        module,
		Protocol:    "http",
		Paths:       []string{fmt.Sprintf("/bcsapi/v4/%s/", modules.BCSModuleUserManager)},
		PathRewrite: true,
		Service:     module,
		Labels:      labels,
	}
	regSvc.Routes = append(regSvc.Routes, rt)
	//setting upstream backend information
	bcks := adp.constructUpstreamTarget(svc.Nodes)
	regSvc.Backends = append(regSvc.Backends, bcks...)
	return regSvc, nil
}

//microMetricService convert bcs-cluster-manager service information
// to custom service definition. this is compatible with original bcs-api proxy.
// and further more, api-gateway defines new standard proxy rule for it
func (adp *Adapter) microMetricService(module string, svc *registry.Service) (*register.Service, error) {
	labels := make(map[string]string)
	labels["module"] = metricModeule
	labels["service"] = defaultServiceTag
	regSvc := &register.Service{
		Name:     module,
		Protocol: "https",
		Host:     svc.Name,
		Path:     fmt.Sprintf("/%s/", metricModeule),
		Retries:  1,
		Labels:   labels,
	}
	//setting route information
	rt := register.Route{
		Name:        module,
		Protocol:    "http",
		Paths:       []string{fmt.Sprintf("/bcsapi/v4/%s/", metricModeule)},
		PathRewrite: true,
		Service:     module,
		Labels:      labels,
	}
	regSvc.Routes = append(regSvc.Routes, rt)
	//setting upstream backend information
	bcks := adp.constructUpstreamTarget(svc.Nodes)
	regSvc.Backends = append(regSvc.Backends, bcks...)
	return regSvc, nil
}

//microNetworkDetection convert bcs-network-detection service information
// to custom service definition. this is compatible with original bcs-api proxy.
// and further more, api-gateway defines new standard proxy rule for it
func (adp *Adapter) microNetworkDetection(module string, svc *registry.Service) (*register.Service, error) {
	labels := make(map[string]string)
	labels["module"] = modules.BCSModuleNetworkdetection
	labels["service"] = defaultServiceTag
	regSvc := &register.Service{
		Name:     module,
		Protocol: "https",
		Host:     svc.Name,
		Path:     "/detection/v4/",
		Retries:  1,
		Labels:   labels,
	}
	//setting route information
	rt := register.Route{
		Name:        module,
		Protocol:    "http",
		Paths:       []string{"/bcsapi/v4/detection/"},
		PathRewrite: true,
		Plugin: &register.Plugins{
			AuthOption: &register.BCSAuthOption{
				Name: defaultPluginName,
				//sending auth request to usermanager.bkbcs.tencent.com
				AuthEndpoints: fmt.Sprintf("https://%s%s", modules.BCSModuleUserManager, defaultDomain),
				AuthToken:     adp.admintoken,
				Module:        modules.BCSModuleNetworkdetection,
			},
		},
		Service: module,
		Labels:  labels,
	}
	regSvc.Routes = append(regSvc.Routes, rt)
	//setting upstream backend information
	bcks := adp.constructUpstreamTarget(svc.Nodes)
	regSvc.Backends = append(regSvc.Backends, bcks...)
	return regSvc, nil
}
