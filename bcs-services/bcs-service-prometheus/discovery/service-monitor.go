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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-mesos/pkg/client/informers"
	"github.com/Tencent/bk-bcs/bcs-mesos/pkg/client/internalclientset"
	bkbcsv2 "github.com/Tencent/bk-bcs/bcs-mesos/pkg/client/lister/bkbcs/v2"
	monitorInformers "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/informers/externalversions"
	monitorClientset "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/clientset/versioned"
	monitorv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/listers/monitor/v1"
	apismonitorv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/monitor/v1"
	apisbkbcsv2 "github.com/Tencent/bk-bcs/bcs-mesos/pkg/apis/bkbcs/v2"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-service-prometheus/types"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/cache"
)

type serviceMonitor struct {
	kubeconfig     string
	sdFilePath     string
	cadvisorPort   string
	nodeExportPort int
	module         string

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
	endpoint map[string]*apisbkbcsv2.BcsEndpoint
}

// new serviceMonitor for discovery node cadvisor targets
func NewserviceMonitor(kubeconfig string, promFilePrefix, module string, cadvisorPort, nodeExportPort int) (Discovery, error) {
	disc := &serviceMonitor{
		kubeconfig:     kubeconfig,
		module:         module,
		svrMonitors: make(map[string]*serviceEndpoint),
	}
	switch module {
	case CadvisorModule:
		if cadvisorPort <= 0 {
			return nil, fmt.Errorf("cadvisorPort can't be zero")
		}
	case NodeexportModule:
		if nodeExportPort <= 0 {
			return nil, fmt.Errorf("nodeExportPort can't be zero")
		}
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
	internalFactory.Start(stopCh)
	// Wait for all caches to sync.
	internalFactory.WaitForCacheSync(stopCh)
	blog.Infof("build bkbcsClientset for config %s success", disc.kubeconfig)

	//init monitor clientset
	monitorClient,err := monitorClientset.NewForConfig(cfg)
	monitorFactory := monitorInformers.NewSharedInformerFactory(monitorClient, 0)
	disc.serviceMonitorLister = monitorFactory.Monitor().V1().ServiceMonitors().Lister()
	disc.serviceMonitorInformer = monitorFactory.Monitor().V1().ServiceMonitors().Informer()
	monitorFactory.Start(stopCh)
	monitorFactory.WaitForCacheSync(stopCh)
	blog.Infof("build monitorClientset for config %s success", disc.kubeconfig)

	disc.initSuccess = true
	disc.eventHandler(disc.module)
	return nil
}

func (disc *serviceMonitor) GetPrometheusSdConfig(module string) ([]*types.PrometheusSdConfig, error) {
	nodes, err := disc.nodeLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	promConfigs := make([]*types.PrometheusSdConfig, 0)
	for _, node := range nodes {
		ip := node.Spec.GetAgentIP()
		if ip == "" {
			blog.Errorf("discovery %s node %s not found InnerIP", disc.module, node.GetName())
			continue
		}

		switch disc.module {
		case CadvisorModule:
			conf := &types.PrometheusSdConfig{
				Targets: []string{fmt.Sprintf("%s:%d", ip, disc.cadvisorPort)},
				Labels: map[string]string{
					DefaultBcsModuleLabelKey: disc.module,
				},
			}

			promConfigs = append(promConfigs, conf)

		case NodeexportModule:
			conf := &types.PrometheusSdConfig{
				Targets: []string{fmt.Sprintf("%s:%d", ip, disc.nodeExportPort)},
				Labels: map[string]string{
					DefaultBcsModuleLabelKey: disc.module,
				},
			}

			promConfigs = append(promConfigs, conf)
		}
	}

	return promConfigs, nil
}

func (disc *serviceMonitor) GetPromSdConfigFile(module string) string {
	return path.Join(disc.promFilePrefix, fmt.Sprintf("%s%s", module, DiscoveryFileName))
}

func (disc *serviceMonitor) RegisterEventFunc(handleFunc EventHandleFunc) {
	disc.eventHandler = handleFunc
}

func (disc *serviceMonitor) OnAdd(obj interface{}) {
	if !disc.initSuccess {
		return
	}

	disc.eventHandler(disc.module)
}

// if on update event, then don't need to update sd config
func (disc *serviceMonitor) OnUpdate(old, cur interface{}) {
	if !disc.initSuccess {
		return
	}
}

func (disc *serviceMonitor) OnDelete(obj interface{}) {
	if !disc.initSuccess {
		return
	}

	// call event handler
	disc.eventHandler(disc.module)
}

func (disc *serviceMonitor) initServiceMonitor()error{
	svrs,err := disc.serviceMonitorLister.ServiceMonitors("").List(labels.Everything())
	if err!=nil {
		blog.Errorf("List ServiceMonitors failed: %s", err.Error())
		return err
	}

	for _,svr :=range svrs {
		o := &serviceEndpoint{
			serviceM: svr,
			endpoint: make(map[string]*apisbkbcsv2.BcsEndpoint),
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
			o.endpoint[v.GetUuid()] = v
		}
		disc.svrMonitors[svr.GetUuid()] = o
	}

	return nil
}


