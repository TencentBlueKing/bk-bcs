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

package discovery

import (
	"fmt"
	"path"
	"sync"

	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-mesos/pkg/client/informers"
	"github.com/Tencent/bk-bcs/bcs-mesos/pkg/client/internalclientset"
	bkbcsv2 "github.com/Tencent/bk-bcs/bcs-mesos/pkg/client/lister/bkbcs/v2"
	monitorv1 "github.com/Tencent/bk-bcs/bcs-mesos/pkg/client/lister/monitor/v1"
	apismonitorv1 "github.com/Tencent/bk-bcs/bcs-mesos/pkg/apis/monitor/v1"
	apisbkbcsv2 "github.com/Tencent/bk-bcs/bcs-mesos/pkg/apis/bkbcs/v2"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-service-prometheus/types"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/cache"
)

type serviceMonitor struct {
	sync.RWMutex
	kubeconfig     string
	sdFilePath     string
	module         string
	promFilePrefix string

	eventHandler   EventHandleFunc
	//endpoints
	endpointLister bkbcsv2.BcsEndpointLister
	endpointInformer cache.SharedIndexInformer
	//service monitor
	serviceMonitorLister monitorv1.ServiceMonitorLister
	serviceMonitorInformer cache.SharedIndexInformer
	initSuccess    bool

	svrMonitors map[string]*serviceEndpoint
}

type serviceEndpoint struct {
	serviceM *apismonitorv1.ServiceMonitor
	endpoints map[string]*apisbkbcsv2.BcsEndpoint
	cPorts map[string]apismonitorv1.Endpoint
}

func (s *serviceEndpoint) getPrometheusConfigs()[]*types.PrometheusSdConfig{
	promConfigs := make([]*types.PrometheusSdConfig,0)
	for _,bcsEndpoint :=range s.endpoints {
		conf := &types.PrometheusSdConfig{
			Targets: make([]string, 0),
			Labels: bcsEndpoint.Labels,
		}
		for _,endpoint :=range bcsEndpoint.Spec.Endpoints {
			for _,cPort := range endpoint.Ports {
				portInfo,ok := s.cPorts[cPort.Name]
				if !ok {
					blog.V(3).Infof("BcsEndpoint(%s) endpoint(%s) port(%s) don't matched, and continue",
						bcsEndpoint.GetUuid(), endpoint.ContainerIP, cPort.Name)
					continue
				}
				conf.Targets = append(conf.Targets, fmt.Sprintf("%s:%d%s", endpoint.ContainerIP, cPort.ContainerPort, portInfo.Path))
			}
		}

		promConfigs = append(promConfigs, conf)
	}

	return promConfigs
}

// new serviceMonitor for discovery node cadvisor targets
func NewServiceMonitor(kubeconfig string, promFilePrefix, module string) (Discovery, error) {
	disc := &serviceMonitor{
		kubeconfig:     kubeconfig,
		module:         module,
		promFilePrefix: promFilePrefix,
		svrMonitors: make(map[string]*serviceEndpoint),
	}

	return disc, nil
}

func (disc *serviceMonitor) Start() error {
	cfg, err := clientcmd.BuildConfigFromFlags("", disc.kubeconfig)
	if err != nil {
		blog.Errorf("build kubeconfig %s error %s", disc.kubeconfig, err.Error())
		return err
	}
	stopCh := make(chan struct{})
	//internal clientset for informer BcsLogConfig Crd
	internalClientset, err := internalclientset.NewForConfig(cfg)
	if err != nil {
		blog.Errorf("build internal clientset by kubeconfig %s error %s", disc.kubeconfig, err.Error())
		return err
	}
	internalFactory := informers.NewSharedInformerFactory(internalClientset, 0)
	disc.endpointLister = internalFactory.Bkbcs().V2().BcsEndpoints().Lister()
	disc.endpointInformer = internalFactory.Bkbcs().V2().BcsEndpoints().Informer()
	blog.Infof("build bkbcsClientset for config %s success", disc.kubeconfig)

	//init monitor clientset
	disc.serviceMonitorLister = internalFactory.Monitor().V1().ServiceMonitors().Lister()
	disc.serviceMonitorInformer = internalFactory.Monitor().V1().ServiceMonitors().Informer()
	internalFactory.Start(stopCh)
	// Wait for all caches to sync.
	internalFactory.WaitForCacheSync(stopCh)
	blog.Infof("build monitorClientset for config %s success", disc.kubeconfig)
	err = disc.initServiceMonitor()
	if err!=nil {
		return err
	}
	//trigger event handler
	disc.eventHandler(disc.module)
	//add k8s resources event handler functions
	disc.serviceMonitorInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    disc.OnServiceMonitorAdd,
			UpdateFunc: disc.OnServiceMonitorUpdate,
			DeleteFunc: disc.OnServiceMonitorDelete,
		},
	)
	disc.endpointInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    disc.OnEndpointsAdd,
			UpdateFunc: disc.OnEndpointsUpdate,
			DeleteFunc: disc.OnEndpointsDelete,
		},
	)
	disc.initSuccess = true
	return nil
}

func (disc *serviceMonitor) GetPrometheusSdConfig(module string) ([]*types.PrometheusSdConfig, error) {
	promConfigs := make([]*types.PrometheusSdConfig, 0)
	disc.Lock()
	for _,svrMonitor := range disc.svrMonitors {
		promConfigs = append(promConfigs, svrMonitor.getPrometheusConfigs()...)
	}
	disc.Unlock()

	return promConfigs, nil
}

func (disc *serviceMonitor) GetPromSdConfigFile(module string) string {
	return path.Join(disc.promFilePrefix, fmt.Sprintf("%s%s", module, DiscoveryFileName))
}

func (disc *serviceMonitor) RegisterEventFunc(handleFunc EventHandleFunc) {
	disc.eventHandler = handleFunc
}

func (disc *serviceMonitor) OnServiceMonitorAdd(obj interface{}) {
	if !disc.initSuccess {
		return
	}

	serviceM, ok := obj.(*apismonitorv1.ServiceMonitor)
	if !ok {
		blog.Errorf("cannot convert to *apismonitorv1.ServiceMonitor: %v", obj)
		return
	}
	o := &serviceEndpoint{
		serviceM: serviceM,
		endpoints: make(map[string]*apisbkbcsv2.BcsEndpoint),
	}
	rms := labels.NewSelector()
	for _,o :=range serviceM.GetSelector() {
		rms.Add(o)
	}
	endpoints,err := disc.endpointLister.BcsEndpoints(serviceM.Namespace).List(rms)
	if err!=nil {
		blog.Errorf("ServiceMonitor(%s) get Endpoints failed: %s", serviceM.GetUuid(), err.Error())
		return
	}
	for _,v :=range endpoints {
		o.endpoints[v.GetUuid()] = v
		blog.Infof("ServiceMonitor(%s) add selected BcsEndpoint(%s) success", serviceM.GetUuid(), v.GetUuid())
	}
	disc.Lock()
	disc.svrMonitors[serviceM.GetUuid()] = o
	disc.Unlock()
	blog.Infof("handle Add event ServiceMonitor(%s) success", serviceM.GetUuid())
	disc.eventHandler(disc.module)
}

// if on update event, then don't need to update sd config
func (disc *serviceMonitor) OnServiceMonitorUpdate(old, cur interface{}) {
	if !disc.initSuccess {
		return
	}

	serviceM, ok := cur.(*apismonitorv1.ServiceMonitor)
	if !ok {
		blog.Errorf("cannot convert to *apismonitorv1.ServiceMonitor: %v", cur)
		return
	}
	o := &serviceEndpoint{
		serviceM: serviceM,
		endpoints: make(map[string]*apisbkbcsv2.BcsEndpoint),
	}
	rms := labels.NewSelector()
	for _,o :=range serviceM.GetSelector() {
		rms.Add(o)
	}
	endpoints,err := disc.endpointLister.BcsEndpoints(serviceM.Namespace).List(rms)
	if err!=nil {
		blog.Errorf("ServiceMonitor(%s) get Endpoints failed: %s", serviceM.GetUuid(), err.Error())
		return
	}
	for _,v :=range endpoints {
		o.endpoints[v.GetUuid()] = v
		blog.Infof("ServiceMonitor(%s) add selected BcsEndpoint(%s) success", serviceM.GetUuid(), v.GetUuid())
	}
	disc.Lock()
	disc.svrMonitors[serviceM.GetUuid()] = o
	disc.Unlock()
	blog.Infof("handle Update event ServiceMonitor(%s) success", serviceM.GetUuid())
	disc.eventHandler(disc.module)
}

func (disc *serviceMonitor) OnServiceMonitorDelete(obj interface{}) {
	if !disc.initSuccess {
		return
	}
	serviceM, ok := obj.(*apismonitorv1.ServiceMonitor)
	if !ok {
		blog.Errorf("cannot convert to *apismonitorv1.ServiceMonitor: %v", obj)
		return
	}
	disc.Lock()
	delete(disc.svrMonitors, serviceM.GetUuid())
	disc.Unlock()
	blog.Infof("handle Delete event ServiceMonitor(%s) success", serviceM.GetUuid())
	// call event handler
	disc.eventHandler(disc.module)
}

func (disc *serviceMonitor) OnEndpointsAdd(obj interface{}) {
	if !disc.initSuccess {
		return
	}

	endpoint, ok := obj.(*apisbkbcsv2.BcsEndpoint)
	if !ok {
		blog.Errorf("cannot convert to *apisbkbcsv2.BcsEndpoint: %v", obj)
		return
	}
	matched := false
	for _,sm :=range disc.svrMonitors {
		serviceM := sm.serviceM
		if !serviceM.Match(endpoint.Labels) {
			blog.V(3).Infof("ServiceMonitor(%s) don't match BcsEndpoint(%s), and continue", serviceM.GetUuid(), endpoint.GetUuid())
			continue
		}
		matched = true
		disc.Lock()
		sm.endpoints[endpoint.GetUuid()] = endpoint
		disc.Unlock()
		blog.Infof("ServiceMonitor(%s) add selected BcsEndpoint(%s) success", serviceM.GetUuid(), endpoint.GetUuid())
	}

	if matched {
		// call event handler
		disc.eventHandler(disc.module)
	}
}

// if on update event, then don't need to update sd config
func (disc *serviceMonitor) OnEndpointsUpdate(old, cur interface{}) {
	if !disc.initSuccess {
		return
	}
	oldEndpoint, ok := old.(*apisbkbcsv2.BcsEndpoint)
	if !ok {
		blog.Errorf("cannot convert to *apisbkbcsv2.BcsEndpoint: %v", old)
		return
	}
	curEndpoint, ok := cur.(*apisbkbcsv2.BcsEndpoint)
	if !ok {
		blog.Errorf("cannot convert to *apisbkbcsv2.BcsEndpoint: %v", cur)
		return
	}
	changed := checkEndpointsChanged(oldEndpoint.Spec.BcsEndpoint, curEndpoint.Spec.BcsEndpoint)
	if !changed {
		blog.Infof("OnEndpointsUpdate BcsEndpoint(%s) don't change", oldEndpoint.GetUuid())
		return
	}

	matched := false
	for _,sm :=range disc.svrMonitors {
		serviceM := sm.serviceM
		if !serviceM.Match(curEndpoint.Labels) {
			blog.V(3).Infof("ServiceMonitor(%s) don't match BcsEndpoint(%s), and continue", serviceM.GetUuid(), curEndpoint.GetUuid())
			continue
		}
		matched = true
		disc.Lock()
		sm.endpoints[curEndpoint.GetUuid()] = curEndpoint
		disc.Unlock()
		blog.Infof("ServiceMonitor(%s) update selected BcsEndpoint(%s) success", serviceM.GetUuid(), curEndpoint.GetUuid())
	}

	if matched {
		// call event handler
		disc.eventHandler(disc.module)
	}
}

func checkEndpointsChanged(old, cur commtypes.BcsEndpoint)bool{
	if len(old.Endpoints)!=len(cur.Endpoints) {
		return true
	}
	endpoints := make(map[string]bool)
	for _,in := range old.Endpoints {
		endpoints[in.ContainerIP] = false
	}
	for _,in := range cur.Endpoints {
		endpoints[in.ContainerIP] = true
	}
	for k,v :=range endpoints {
		if !v {
			blog.Infof("BcsEndpoint(%s) ContainerIP(%s) changed", k)
			return true
		}
	}

	return false
}

func (disc *serviceMonitor) OnEndpointsDelete(obj interface{}) {
	if !disc.initSuccess {
		return
	}
	endpoint, ok := obj.(*apisbkbcsv2.BcsEndpoint)
	if !ok {
		blog.Errorf("cannot convert to *apisbkbcsv2.BcsEndpoint: %v", obj)
		return
	}
	matched := false
	for _,sm :=range disc.svrMonitors {
		serviceM := sm.serviceM
		if !serviceM.Match(endpoint.Labels) {
			blog.V(3).Infof("ServiceMonitor(%s) don't match BcsEndpoint(%s), and continue", serviceM.GetUuid(), endpoint.GetUuid())
			continue
		}
		matched = true
		disc.Lock()
		delete(sm.endpoints, endpoint.GetUuid())
		disc.Unlock()
		blog.Infof("ServiceMonitor(%s) delete selected BcsEndpoint(%s) success", serviceM.GetUuid(), endpoint.GetUuid())
	}

	if matched {
		// call event handler
		disc.eventHandler(disc.module)
	}
}

func (disc *serviceMonitor) initServiceMonitor()error{
	svrs,err := disc.serviceMonitorLister.ServiceMonitors("").List(labels.Everything())
	if err!=nil {
		blog.Errorf("List ServiceMonitors failed: %s", err.Error())
		return err
	}

	for _,svr :=range svrs {
		blog.Infof("init ServiceMonitor(%s) starting", svr.GetUuid())
		o := &serviceEndpoint{
			serviceM: svr,
			endpoints: make(map[string]*apisbkbcsv2.BcsEndpoint),
			cPorts: make(map[string]apismonitorv1.Endpoint),
		}
		for _,endpoint :=range svr.Spec.Endpoints {
			o.cPorts[endpoint.Port] = endpoint
		}
		rms := labels.NewSelector()
		for _,o :=range svr.GetSelector() {
			rms.Add(o)
		}
		endpoints,err := disc.endpointLister.BcsEndpoints(svr.Namespace).List(rms)
		if err!=nil {
			blog.Errorf("get Endpoints failed: %s", err.Error())
			continue
		}
		for _,v :=range endpoints {
			o.endpoints[v.GetUuid()] = v
			blog.Infof("ServiceMonitor(%s) add selected BcsEndpoint(%s) success", svr.GetUuid(), v.GetUuid())
		}
		disc.svrMonitors[svr.GetUuid()] = o
	}
	blog.Infof("Init ServiceMonitor done")
	return nil
}



