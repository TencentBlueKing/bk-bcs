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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	apisbkbcsv2 "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"
	internalclientset "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/clientset/versioned"
	informers "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/informers/externalversions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-service-prometheus/types"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type nodeEtcdDiscovery struct {
	sync.RWMutex
	kubeconfig     string
	sdFilePath     string
	cadvisorPort   int
	nodeExportPort int
	module         string

	eventHandler   EventHandleFunc
	nodeInformer   cache.SharedIndexInformer
	initSuccess    bool
	promFilePrefix string
	nodes          map[string]struct{}
}

// NewNodeEtcdDiscovery new nodeEtcdDiscovery for discovery node cadvisor targets
func NewNodeEtcdDiscovery(kubeconfig string, promFilePrefix, module string, cadvisorPort, nodeExportPort int) (Discovery, error) {
	disc := &nodeEtcdDiscovery{
		kubeconfig:     kubeconfig,
		promFilePrefix: promFilePrefix,
		cadvisorPort:   cadvisorPort,
		nodeExportPort: nodeExportPort,
		module:         module,
		nodes:          make(map[string]struct{}),
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

// Start node discovery from etcd storage
func (disc *nodeEtcdDiscovery) Start() error {
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
	disc.nodeInformer = internalFactory.Bkbcs().V2().Agents().Informer()
	internalFactory.Start(stopCh)
	// Wait for all caches to sync.
	internalFactory.WaitForCacheSync(stopCh)
	blog.Infof("build internalClientset for config %s success", disc.kubeconfig)
	disc.nodeInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    disc.OnAdd,
			UpdateFunc: disc.OnUpdate,
			DeleteFunc: disc.OnDelete,
		},
	)
	return nil
}

func (disc *nodeEtcdDiscovery) GetPrometheusSdConfig(module string) ([]*types.PrometheusSdConfig, error) {
	promConfigs := make([]*types.PrometheusSdConfig, 0)
	for nodeIP := range disc.nodes {
		switch disc.module {
		case CadvisorModule:
			conf := &types.PrometheusSdConfig{
				Targets: []string{fmt.Sprintf("%s:%d", nodeIP, disc.cadvisorPort)},
				Labels: map[string]string{
					DefaultBcsModuleLabelKey: disc.module,
				},
			}

			promConfigs = append(promConfigs, conf)

		case NodeexportModule:
			conf := &types.PrometheusSdConfig{
				Targets: []string{fmt.Sprintf("%s:%d", nodeIP, disc.nodeExportPort)},
				Labels: map[string]string{
					DefaultBcsModuleLabelKey: disc.module,
				},
			}

			promConfigs = append(promConfigs, conf)
		}
	}

	return promConfigs, nil
}

func (disc *nodeEtcdDiscovery) GetPromSdConfigFile(module string) string {
	return path.Join(disc.promFilePrefix, fmt.Sprintf("%s%s", module, DiscoveryFileName))
}

func (disc *nodeEtcdDiscovery) RegisterEventFunc(handleFunc EventHandleFunc) {
	disc.eventHandler = handleFunc
}

// OnAdd add event handler
func (disc *nodeEtcdDiscovery) OnAdd(obj interface{}) {
	agent, ok := obj.(*apisbkbcsv2.Agent)
	if !ok {
		blog.Errorf("cannot convert to *apisbkbcsv2.Agent: %v", obj)
		return
	}
	blog.Infof("receive Agent(%s) Add event", agent.Name)
	ip := agent.Spec.GetAgentIP()
	if ip == "" {
		blog.Errorf("node %s not found InnerIP", agent.GetName())
		return
	}
	disc.nodes[ip] = struct{}{}

	disc.eventHandler(Info{Module: disc.module, Key: disc.module})
}

// OnUpdate if on update event, then don't need to update sd config
func (disc *nodeEtcdDiscovery) OnUpdate(old, cur interface{}) {
	//do nothing
}

// OnDelete delete event handler
func (disc *nodeEtcdDiscovery) OnDelete(obj interface{}) {
	agent, ok := obj.(*apisbkbcsv2.Agent)
	if !ok {
		blog.Errorf("cannot convert to *apisbkbcsv2.Agent: %v", obj)
		return
	}
	blog.Infof("receive Agent(%s) Delete event", agent.Name)
	ip := agent.Spec.GetAgentIP()
	if ip == "" {
		blog.Errorf("node %s not found InnerIP", agent.GetName())
		return
	}
	delete(disc.nodes, ip)

	// call event handler
	disc.eventHandler(Info{Module: disc.module, Key: disc.module})
}
