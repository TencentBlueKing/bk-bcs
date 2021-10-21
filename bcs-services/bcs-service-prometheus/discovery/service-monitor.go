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
	"context"
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	apisbkbcsv2 "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"
	apismonitorv1 "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/monitor/v1"
	internalclientset "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/clientset/versioned"
	informers "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/informers/externalversions"
	bkbcsv2 "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/listers/bkbcs/v2"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-service-prometheus/types"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type serviceMonitor struct {
	sync.RWMutex
	kubeconfig     string
	sdFilePath     string
	module         string
	promFilePrefix string

	eventHandler EventHandleFunc
	//endpoints
	endpointLister   bkbcsv2.BcsEndpointLister
	endpointInformer cache.SharedIndexInformer
	//service monitor
	//serviceMonitorLister monitorv1.ServiceMonitorLister
	serviceMonitorInformer cache.SharedIndexInformer
	//apiextensions clientset
	extensionClientset *apiextensionsclient.Clientset
	//local cache for combination of servicemonitor & bcsendpoint
	svrMonitors map[string]*serviceEndpoint
}

type serviceEndpoint struct {
	serviceM  *apismonitorv1.ServiceMonitor
	endpoints map[string]*apisbkbcsv2.BcsEndpoint
	cPorts    map[string]apismonitorv1.Endpoint
}

func (s *serviceEndpoint) getPrometheusConfigs() []*types.PrometheusSdConfig {
	promConfigs := make([]*types.PrometheusSdConfig, 0)
	//traverse BcsEndpoints, corresponding to a separate BcsService
	for _, bcsEndpoint := range s.endpoints {
		conf := &types.PrometheusSdConfig{
			Targets: make([]string, 0),
			Labels:  make(map[string]string),
		}
		for k, v := range bcsEndpoint.Labels {
			r, _ := regexp.Compile("[a-zA-Z_][a-zA-Z0-9_]*")
			rk := r.FindAllString(k, 1)
			if len(rk) != 1 || rk[0] != k {
				blog.Warnf("BcsEndpoint(%s) Label(%s: %s) is invalid, skip", bcsEndpoint.GetUuid(), k, v)
				continue
			}
			conf.Labels[k] = v
		}

		//append ServiceMonitor Identity
		conf.Labels["ServiceMonitor"] = s.serviceM.GetUuid()
		//append BcsEndpoint Identity
		conf.Labels["BcsEndpoint"] = bcsEndpoint.GetUuid()
		//append namespace
		conf.Labels["namespace"] = bcsEndpoint.Namespace
		conf.Labels["job"] = fmt.Sprintf("%s/%s/0", bcsEndpoint.Namespace, s.serviceM.Name)
		//conf.Labels["name"] = s.serviceM.Name
		for _, endpoint := range bcsEndpoint.Spec.Endpoints {
			for _, cPort := range endpoint.Ports {
				_, ok := s.cPorts[cPort.Name]
				if !ok {
					blog.V(5).Infof("BcsEndpoint(%s) endpoint(%s) port(%s) don't matched, and continue",
						bcsEndpoint.GetUuid(), endpoint.ContainerIP, cPort.Name)
					continue
				}
				//if container network=Host
				if endpoint.NetworkMode == "HOST" {
					conf.Targets = append(conf.Targets, fmt.Sprintf("%s:%d", endpoint.NodeIP, cPort.ContainerPort))
				} else {
					conf.Targets = append(conf.Targets, fmt.Sprintf("%s:%d", endpoint.ContainerIP, cPort.ContainerPort))
				}
			}
		}
		promConfigs = append(promConfigs, conf)
	}

	return promConfigs
}

// NewServiceMonitor new serviceMonitor for discovery node cadvisor targets
func NewServiceMonitor(kubeconfig string, promFilePrefix, module string) (Discovery, error) {
	disc := &serviceMonitor{
		kubeconfig:     kubeconfig,
		module:         module,
		promFilePrefix: promFilePrefix,
		svrMonitors:    make(map[string]*serviceEndpoint),
	}

	return disc, nil
}

// Start start up service monitor feature
func (disc *serviceMonitor) Start() error {
	cfg, err := clientcmd.BuildConfigFromFlags("", disc.kubeconfig)
	if err != nil {
		blog.Errorf("build kubeconfig %s error %s", disc.kubeconfig, err.Error())
		return err
	}
	//apiextensions clientset for creating BcsLogConfig Crd
	disc.extensionClientset, err = apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		blog.Errorf("build apiextension client by kubeconfig % error %s", disc.kubeconfig, err.Error())
		return err
	}
	//create BcsLogConfig Crd
	err = disc.createserviceMonitor()
	if err != nil {
		return err
	}
	stopCh := make(chan struct{})
	//internal clientset for informer serviceMonitor Crd
	internalClientset, err := internalclientset.NewForConfig(cfg)
	if err != nil {
		blog.Errorf("build internal clientset by kubeconfig %s error %s", disc.kubeconfig, err.Error())
		return err
	}
	internalFactory := informers.NewSharedInformerFactory(internalClientset, 0)
	disc.endpointInformer = internalFactory.Bkbcs().V2().BcsEndpoints().Informer()
	disc.endpointLister = internalFactory.Bkbcs().V2().BcsEndpoints().Lister()
	blog.Infof("build bkbcsClientset for config %s success", disc.kubeconfig)

	//init monitor clientset
	disc.serviceMonitorInformer = internalFactory.Monitor().V1().ServiceMonitors().Informer()
	internalFactory.Start(stopCh)
	// Wait for all caches to sync.
	internalFactory.WaitForCacheSync(stopCh)
	blog.Infof("build monitorClientset for config %s success", disc.kubeconfig)

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
	return nil
}

// create crd of ServiceMonitor
func (disc *serviceMonitor) createserviceMonitor() error {
	blog.Infof("start create ServiceMonitor")
	serviceMonitorPlural := "servicemonitors"
	serviceMonitorFullName := "servicemonitors" + "." + apismonitorv1.SchemeGroupVersion.Group
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceMonitorFullName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   apismonitorv1.SchemeGroupVersion.Group,   // serviceMonitorsGroup,
			Version: apismonitorv1.SchemeGroupVersion.Version, // serviceMonitorsVersion,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   serviceMonitorPlural,
				Kind:     reflect.TypeOf(apismonitorv1.ServiceMonitor{}).Name(),
				ListKind: reflect.TypeOf(apismonitorv1.ServiceMonitorList{}).Name(),
			},
		},
	}

	_, err := disc.extensionClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(
		context.Background(), crd, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			blog.Infof("serviceMonitor Crd is already exists")
			return nil
		}
		blog.Errorf("create serviceMonitor Crd error %s", err.Error())
		return err
	}
	blog.Infof("create serviceMonitor Crd success")
	return nil
}

func (disc *serviceMonitor) GetPrometheusSdConfig(key string) ([]*types.PrometheusSdConfig, error) {
	disc.Lock()
	defer disc.Unlock()

	svrMonitor, ok := disc.svrMonitors[key]
	if !ok {
		return nil, fmt.Errorf("ServiceMonitor(%s) Not Found", key)
	}

	return svrMonitor.getPrometheusConfigs(), nil
}

func (disc *serviceMonitor) GetPromSdConfigFile(key string) string {
	return strings.ToLower(path.Join(disc.promFilePrefix, fmt.Sprintf("%s_%s%s", key, disc.module, DiscoveryFileName)))
}

func (disc *serviceMonitor) RegisterEventFunc(handleFunc EventHandleFunc) {
	disc.eventHandler = handleFunc
}

func (disc *serviceMonitor) OnServiceMonitorAdd(obj interface{}) {
	serviceM, ok := obj.(*apismonitorv1.ServiceMonitor)
	if !ok {
		blog.Errorf("cannot convert to *apismonitorv1.ServiceMonitor: %v", obj)
		return
	}
	by, _ := json.Marshal(serviceM)
	blog.Infof("receive ServiceMonitor(%s) Data(%s) Add event", serviceM.GetUuid(), string(by))
	disc.handlerServiceMonitorChanged(serviceM)
}

func (disc *serviceMonitor) validateServiceMonitor(serviceM *apismonitorv1.ServiceMonitor) bool {
	if len(serviceM.Spec.Endpoints) == 0 {
		blog.Errorf("ServiceMonitor(%s) len(Endpoints)==0, then ignore it", serviceM.GetUuid())
		return false
	}

	if len(serviceM.Spec.Selector.MatchLabels) == 0 {
		blog.Errorf("ServiceMonitor(%s) len(Selector.MatchLabels)==0, then ignore it", serviceM.GetUuid())
		return false
	}

	return true
}

// if on update event, then don't need to update sd config
func (disc *serviceMonitor) OnServiceMonitorUpdate(old, cur interface{}) {
	serviceM, ok := cur.(*apismonitorv1.ServiceMonitor)
	if !ok {
		blog.Errorf("cannot convert to *apismonitorv1.ServiceMonitor: %v", cur)
		return
	}
	by, _ := json.Marshal(serviceM)
	blog.Infof("receive ServiceMonitor(%s) Data(%s) Update event", serviceM.GetUuid(), string(by))
	disc.handlerServiceMonitorChanged(serviceM)
}

// handlerServiceMonitorChanged recreate relationship between ServiceMonitor & BcsEndpoint
// no matter AddEvent or UpdateEvent
func (disc *serviceMonitor) handlerServiceMonitorChanged(serviceM *apismonitorv1.ServiceMonitor) {
	if !disc.validateServiceMonitor(serviceM) {
		return
	}
	o := &serviceEndpoint{
		serviceM:  serviceM,
		endpoints: make(map[string]*apisbkbcsv2.BcsEndpoint),
		cPorts:    make(map[string]apismonitorv1.Endpoint),
	}
	for _, endpoint := range serviceM.Spec.Endpoints {
		o.cPorts[endpoint.Port] = endpoint
		blog.Infof("ServiceMonitor(%s) have endpoint(%s:%s)", serviceM.GetUuid(), endpoint.Port, endpoint.Path)
	}
	rms := labels.NewSelector()
	reqs, err := serviceM.GetSelector()
	if err != nil {
		blog.Errorf("ServiceMonitor(%s) selector definition err, %s. skip handling", serviceM.GetUuid(), err.Error())
		return
	}
	for _, o := range reqs {
		rms.Add(o)
	}
	endpoints, err := disc.endpointLister.BcsEndpoints(serviceM.Namespace).List(rms)
	if err != nil {
		blog.Errorf("ServiceMonitor(%s) get Endpoints failed: %s", serviceM.GetUuid(), err.Error())
		return
	}
	for _, v := range endpoints {
		if !serviceM.Match(v.Labels) {
			blog.V(5).Infof("ServiceMonitor(%s) don't match BcsEndpoint(%s), and continue", serviceM.GetUuid(), v.GetUuid())
			continue
		}
		o.endpoints[v.GetUuid()] = v
		blog.Infof("ServiceMonitor(%s) add selected BcsEndpoint(%s) success", serviceM.GetUuid(), v.GetUuid())
	}
	disc.Lock()
	disc.svrMonitors[serviceM.GetUuid()] = o
	disc.Unlock()
	blog.Infof("handle recreate ServiceMonitor(%s) success", serviceM.GetUuid())

	go disc.eventHandler(Info{Module: disc.module, Key: serviceM.GetUuid()})
}

func (disc *serviceMonitor) OnServiceMonitorDelete(obj interface{}) {
	serviceM, ok := obj.(*apismonitorv1.ServiceMonitor)
	if !ok {
		blog.Errorf("cannot convert to *apismonitorv1.ServiceMonitor: %v", obj)
		return
	}
	blog.Infof("receive ServiceMonitor(%s) Delete event", serviceM.GetUuid())
	disc.Lock()
	delete(disc.svrMonitors, serviceM.GetUuid())
	disc.Unlock()
	blog.Infof("handle Delete event ServiceMonitor(%s) success", serviceM.GetUuid())
	// call event handler
	go disc.eventHandler(Info{Module: disc.module, Key: serviceM.GetUuid()})
}

func (disc *serviceMonitor) OnEndpointsAdd(obj interface{}) {
	endpoint, ok := obj.(*apisbkbcsv2.BcsEndpoint)
	if !ok {
		blog.Errorf("cannot convert to *apisbkbcsv2.BcsEndpoint: %v", obj)
		return
	}
	by, _ := json.Marshal(endpoint)
	blog.Infof("receive BcsEndpoint(%s) Data(%s) Add event", endpoint.GetUuid(), string(by))
	disc.Lock()
	defer disc.Unlock()
	for _, sm := range disc.svrMonitors {
		serviceM := sm.serviceM
		if serviceM.Namespace != endpoint.Namespace || !serviceM.Match(endpoint.Labels) {
			blog.V(5).Infof("ServiceMonitor(%s) don't match BcsEndpoint(%s), and continue", serviceM.GetUuid(), endpoint.GetUuid())
			continue
		}
		sm.endpoints[endpoint.GetUuid()] = endpoint
		blog.Infof("ServiceMonitor(%s) add selected BcsEndpoint(%s) success", serviceM.GetUuid(), endpoint.GetUuid())
		// call event handler
		go disc.eventHandler(Info{Module: disc.module, Key: serviceM.GetUuid()})
	}
}

// if on update event, then don't need to update sd config
func (disc *serviceMonitor) OnEndpointsUpdate(old, cur interface{}) {
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
	by, _ := json.Marshal(curEndpoint)
	blog.Infof("receive BcsEndpoint(%s) Data(%s) Update event", curEndpoint.GetUuid(), string(by))
	disc.Lock()
	defer disc.Unlock()
	for _, sm := range disc.svrMonitors {
		serviceM := sm.serviceM
		if serviceM.Namespace != curEndpoint.Namespace || !serviceM.Match(curEndpoint.Labels) {
			blog.V(5).Infof("ServiceMonitor(%s) don't match BcsEndpoint(%s), and continue", serviceM.GetUuid(), curEndpoint.GetUuid())
			continue
		}
		sm.endpoints[curEndpoint.GetUuid()] = curEndpoint
		blog.Infof("ServiceMonitor(%s) update selected BcsEndpoint(%s) success", serviceM.GetUuid(), curEndpoint.GetUuid())
		// call event handler
		go disc.eventHandler(Info{Module: disc.module, Key: serviceM.GetUuid()})
	}
}

func checkEndpointsChanged(old, cur commtypes.BcsEndpoint) bool {
	if len(old.Endpoints) != len(cur.Endpoints) {
		return true
	}
	endpoints := make(map[string]bool)
	for _, in := range old.Endpoints {
		endpoints[in.ContainerIP] = false
	}
	for _, in := range cur.Endpoints {
		endpoints[in.ContainerIP] = true
	}
	for k, v := range endpoints {
		if !v {
			blog.Infof("BcsEndpoint(%s) ContainerIP(%s) changed", k)
			return true
		}
	}

	return false
}

func (disc *serviceMonitor) OnEndpointsDelete(obj interface{}) {
	endpoint, ok := obj.(*apisbkbcsv2.BcsEndpoint)
	if !ok {
		blog.Errorf("cannot convert to *apisbkbcsv2.BcsEndpoint: %v", obj)
		return
	}
	blog.Infof("receive BcsEndpoint(%s) Delete event", endpoint.GetUuid())
	disc.Lock()
	defer disc.Unlock()
	for _, sm := range disc.svrMonitors {
		serviceM := sm.serviceM
		if serviceM.Namespace != endpoint.Namespace || !serviceM.Match(endpoint.Labels) {
			blog.V(5).Infof("ServiceMonitor(%s) don't match BcsEndpoint(%s), and continue", serviceM.GetUuid(), endpoint.GetUuid())
			continue
		}
		delete(sm.endpoints, endpoint.GetUuid())
		blog.Infof("ServiceMonitor(%s) delete selected BcsEndpoint(%s) success", serviceM.GetUuid(), endpoint.GetUuid())
		// call event handler
		go disc.eventHandler(Info{Module: disc.module, Key: serviceM.GetUuid()})
	}
}
