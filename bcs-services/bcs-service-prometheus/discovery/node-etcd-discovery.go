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
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-mesos/pkg/client/informers"
	"bk-bcs/bcs-mesos/pkg/client/internalclientset"
	bkbcsv2 "bk-bcs/bcs-mesos/pkg/client/lister/bkbcs/v2"
	"bk-bcs/bcs-services/bcs-service-prometheus/types"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/clientcmd"
)

type nodeEtcdDiscovery struct {
	kubeconfig     string
	sdFilePath     string
	cadvisorPort   int
	nodeExportPort int
	module         string

	eventHandler EventHandleFunc
	nodeLister   bkbcsv2.AgentLister
	initSuccess  bool
	promFilePrefix string
}

// new nodeEtcdDiscovery for discovery node cadvisor targets
func NewNodeEtcdDiscovery(kubeconfig string, promFilePrefix, module string, cadvisorPort, nodeExportPort int) (Discovery, error) {
	disc := &nodeEtcdDiscovery{
		kubeconfig:     kubeconfig,
		promFilePrefix: promFilePrefix,
		cadvisorPort:   cadvisorPort,
		nodeExportPort: nodeExportPort,
		module:         module,
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
	disc.nodeLister = internalFactory.Bkbcs().V2().Agents().Lister()
	internalFactory.Start(stopCh)
	// Wait for all caches to sync.
	internalFactory.WaitForCacheSync(stopCh)
	blog.Infof("build internalClientset for config %s success", disc.kubeconfig)

	go disc.syncTickerPromSdConfig()
	disc.initSuccess = true
	disc.eventHandler(disc.module)
	return nil
}

func (disc *nodeEtcdDiscovery) GetPrometheusSdConfig(module string) ([]*types.PrometheusSdConfig, error) {
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

func (disc *nodeEtcdDiscovery) GetPromSdConfigFile(module string) string {
	return path.Join(disc.promFilePrefix, fmt.Sprintf("%s%s", module, DiscoveryFileName))
}

func (disc *nodeEtcdDiscovery) RegisterEventFunc(handleFunc EventHandleFunc) {
	disc.eventHandler = handleFunc
}

func (disc *nodeEtcdDiscovery) OnAdd(obj interface{}) {
	if !disc.initSuccess {
		return
	}

	disc.eventHandler(disc.module)
}

// if on update event, then don't need to update sd config
func (disc *nodeEtcdDiscovery) OnUpdate(old, cur interface{}) {
	if !disc.initSuccess {
		return
	}
}

func (disc *nodeEtcdDiscovery) OnDelete(obj interface{}) {
	if !disc.initSuccess {
		return
	}

	// call event handler
	disc.eventHandler(disc.module)
}

func (disc *nodeEtcdDiscovery) syncTickerPromSdConfig() {
	ticker := time.NewTicker(time.Minute * 5)

	select {
	case <-ticker.C:
		blog.V(3).Infof("ticker sync prometheus service discovery config")
		disc.eventHandler(disc.module)
	}
}
