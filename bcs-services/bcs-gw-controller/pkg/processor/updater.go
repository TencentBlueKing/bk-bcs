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
	"fmt"
	"strconv"
	"strings"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-services/bcs-clb-controller/pkg/common"
	svcclient "bk-bcs/bcs-services/bcs-clb-controller/pkg/serviceclient"
	"bk-bcs/bcs-services/bcs-gw-controller/pkg/gw"

	k8scache "k8s.io/client-go/tools/cache"
)

// Updater interface
type Updater interface {
	Update() error
}

// GWUpdater gw updater
type GWUpdater struct {
	cache    k8scache.Store
	gwClient gw.Interface
	opt      *Option
	// service client for service discovery
	serviceClient svcclient.Client
}

// GwServiceKeyFunc key function for cache
func GwServiceKeyFunc(obj interface{}) (string, error) {
	cacheService, err := parseService(obj)
	if err != nil {
		return "", err
	}
	return cacheService.Key(), nil
}

// parseService parse service from interface
func parseService(obj interface{}) (*gw.Service, error) {
	if obj == nil {
		return nil, fmt.Errorf("obj is nil")
	}
	cacheService, ok := obj.(*gw.Service)
	if !ok {
		return nil, fmt.Errorf("%v is not cache service", obj)
	}
	return cacheService, nil
}

// NewGWUpdater create gw updater
func NewGWUpdater() *GWUpdater {
	return &GWUpdater{
		cache: k8scache.NewStore(GwServiceKeyFunc),
	}
}

// SetServiceClient set service client for GWUpdater
func (updater *GWUpdater) SetServiceClient(svcClient svcclient.Client) {
	updater.serviceClient = svcClient
}

// SetGWClient set gw client for GWUpdater
func (updater *GWUpdater) SetGWClient(gwClient gw.Interface) {
	updater.gwClient = gwClient
}

// SetOption set command line option for GWUpdater
func (updater *GWUpdater) SetOption(opt *Option) {
	updater.opt = opt
}

// Update implements Updater interface
func (updater *GWUpdater) Update() error {
	appSvcs, err := updater.getAppService()
	if err != nil {
		blog.Errorf("get app service failed, err %s", err.Error())
		return fmt.Errorf("get app service failed, err %s", err.Error())
	}
	serviceMap := updater.convertAppServicesToGWServices(appSvcs)
	err = updater.ensureServices(serviceMap)
	if err != nil {
		blog.Errorf("ensure services failed, err %s", err.Error())
		return err
	}
	return nil
}

// ensureServices ensure services to gw server
func (updater *GWUpdater) ensureServices(serviceMap map[string]*gw.Service) error {
	// do delete services
	var deleteList []*gw.Service
	for _, key := range updater.cache.ListKeys() {
		_, ok := serviceMap[key]
		if !ok {
			cacheSvcObj, existed, err := updater.cache.GetByKey(key)
			if err != nil {
				blog.Warnf("get svc %s from cache failed, err %s", key, err.Error())
				continue
			}
			if existed {
				cacheSvc, err := parseService(cacheSvcObj)
				if err != nil {
					blog.Warnf("parse old svc obj failed, err %s", err.Error())
					continue
				}
				deleteList = append(deleteList, cacheSvc)
			}
		}
	}
	if len(deleteList) != 0 {
		err := updater.gwClient.Delete(deleteList)
		if err != nil {
			return fmt.Errorf("do delete failed, err %s", err.Error())
		}
		for _, svc := range deleteList {
			err = updater.cache.Delete(svc)
			if err != nil {
				blog.Errorf("delete svc %s from cache failed, err %s", svc.Key(), err.Error())
				continue
			}
		}
	}

	// do update services
	var updateList []*gw.Service
	for key, svc := range serviceMap {
		oldSvcObj, existed, err := updater.cache.Get(svc)
		if err != nil {
			blog.Warnf("get svc %s from cache failed, err %s", key, err.Error())
			continue
		}
		if existed {
			oldSvc, err := parseService(oldSvcObj)
			if err != nil {
				// when old svc is invalid, we should do update
				blog.Warnf("parse old svc obj failed, err %s", err.Error())
			} else {
				if !svc.Diff(oldSvc) {
					blog.Errorf("svc %s has no change, skip", key)
					continue
				}
			}
		}
		updateList = append(updateList, svc)
	}
	if len(updateList) != 0 {
		err := updater.gwClient.Update(updateList)
		if err != nil {
			return fmt.Errorf("do update failed, err %s", err.Error())
		}
		for _, svc := range updateList {
			err := updater.cache.Update(svc)
			if err != nil {
				blog.Errorf("update svc %s to cache failed, err %s", svc.Key(), err.Error())
				continue
			}
		}
	}
	return nil
}

// getAppService get AppService from service client
func (updater *GWUpdater) getAppService() ([]*svcclient.AppService, error) {
	appSvcs, err := updater.serviceClient.ListAppService(updater.opt.ServiceLabel)
	if err != nil {
		blog.Errorf("get AppServices by label %v failed, err %s", updater.opt.ServiceLabel, err.Error())
		return nil, fmt.Errorf("get AppServices by label %v failed, err %s", updater.opt.ServiceLabel, err.Error())
	}
	return appSvcs, nil
}

// convertAppServicesToGWServices convert AppService list to GWService map
func (updater *GWUpdater) convertAppServicesToGWServices(appSvcs []*svcclient.AppService) map[string]*gw.Service {
	serviceMap := make(map[string]*gw.Service)
	for _, appSvc := range appSvcs {
		newService, err := updater.convertToGWService(appSvc)
		if err != nil {
			blog.Warnf("convert appSvc %s to gw service failed, err %s", appSvc.GetName()+"/"+appSvc.GetNamespace(), err.Error())
			continue
		}
		if _, ok := serviceMap[newService.Key()]; ok {
			existedService, _ := serviceMap[newService.Key()]
			existedService.LocationList = append(existedService.LocationList, newService.LocationList...)
		} else {
			serviceMap[newService.Key()] = newService
		}
	}
	return serviceMap
}

// convertToGWService convert AppService to GWService
func (updater *GWUpdater) convertToGWService(appSvc *svcclient.AppService) (*gw.Service, error) {
	domain, ok := appSvc.Labels[updater.opt.DomainLabelKey]
	if !ok {
		return nil, fmt.Errorf("label %s cannot be empty", updater.opt.DomainLabelKey)
	}
	proxyPort, ok := appSvc.Labels[updater.opt.ProxyPortLabelKey]
	if !ok {
		return nil, fmt.Errorf("label %s cannot be empty", updater.opt.ProxyPortLabelKey)
	}
	proxyPortInt, err := strconv.Atoi(proxyPort)
	if err != nil {
		return nil, fmt.Errorf("parse proxy port %s failed, err %s", proxyPort, err.Error())
	}
	// port is used to specify a service port
	// maybe there are multiple service ports for a service
	port, ok := appSvc.Labels[updater.opt.PortLabelKey]
	if !ok {
		return nil, fmt.Errorf("label %s cannot be empty", updater.opt.PortLabelKey)
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("parse port %s failed, err %s", port, err.Error())
	}
	// when path label value is empty, path will be /
	// when path label value is aaa.bbb, path will be /aaa/bbb
	path, ok := appSvc.Labels[updater.opt.PathLabelKey]
	if !ok {
		path = ""
	}
	path = "/" + strings.Replace(path, ".", "/", -1)
	gwSvc := &gw.Service{
		BizID:                 updater.opt.GwBizID,
		Domain:                domain,
		VPort:                 proxyPortInt,
		Type:                  gw.ProtocolHTTPS,
		SSLEnable:             true,
		SSLVerifyClientEnable: false,
	}
	gwLocation := &gw.Location{
		URL:    path,
		RSList: updater.getRealServerList(appSvc, portInt),
	}
	gwSvc.LocationList = append(gwSvc.LocationList, gwLocation)
	return gwSvc, nil
}

func (updater *GWUpdater) getRealServerList(appSvc *svcclient.AppService, port int) []*gw.RealServer {
	// find port according to port and clb rule
	var ruledSvcPort svcclient.ServicePort
	foundPort := false
	for _, svcPort := range appSvc.ServicePorts {
		if svcPort.ServicePort == port {
			ruledSvcPort = svcPort
			foundPort = true
		}
	}
	if !foundPort {
		blog.Warnf("find no port %d of AppService %s/%s", port, appSvc.GetNamespace(), appSvc.GetName())
		return nil
	}
	if len(appSvc.Nodes) == 0 {
		blog.Warnf("port %d of AppService %s/%s has no pods", port, appSvc.GetNamespace(), appSvc.GetName())
		return nil
	}
	var rsList []*gw.RealServer
	rsMap := make(map[string]*gw.RealServer)
	for _, node := range appSvc.Nodes {
		for _, port := range node.Ports {
			// svc port and node port may be associated by name port or port number
			if port.NodePort == ruledSvcPort.TargetPort || port.Name == ruledSvcPort.Name {
				var newRealServer *gw.RealServer
				// for overlay ip, we use service NodeIP and service NodePort as backend ip and port
				if updater.opt.BackendIPType == common.BackendIPTypeOverlay {
					newRealServer = &gw.RealServer{
						IP:     node.ProxyIP,
						Port:   ruledSvcPort.ProxyPort,
						Weight: 100,
					}
					// for underlay ip
					// use pod ip and port directly
				} else {
					newRealServer = &gw.RealServer{
						IP:     node.NodeIP,
						Port:   port.NodePort,
						Weight: 100,
					}
					// support pod with mesos bridge network
					if port.ProxyPort > 0 {
						newRealServer.IP = node.ProxyIP
						newRealServer.Port = port.ProxyPort
					}
				}
				// to filter same ip and port, cloud lb cannot bind the same backend twice
				if _, ok := rsMap[newRealServer.IP+strconv.Itoa(newRealServer.Port)]; ok {
					continue
				}
				rsList = append(rsList, newRealServer)
				rsMap[newRealServer.IP+strconv.Itoa(newRealServer.Port)] = newRealServer
				break
			}
		}
	}
	return rsList
}
