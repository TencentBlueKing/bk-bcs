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
	commDiscovery "bk-bcs/bcs-common/pkg/discovery"
	"bk-bcs/bcs-services/bcs-sd-prometheus/types"
)

type nodeDiscovery struct {
	zkAddr         []string
	sdFilePath     string
	cadvisorPort   int
	nodeExportPort int
	module         string

	eventHandler   EventHandleFunc
	nodeController commDiscovery.NodeController
	initSuccess    bool
}

// new nodeDiscovery for discovery node cadvisor targets
func NewNodeDiscovery(zkAddr []string, promFilePrefix, module string, cadvisorPort, nodeExportPort int) (Discovery, error) {
	disc := &nodeDiscovery{
		zkAddr:         zkAddr,
		sdFilePath:     path.Join(promFilePrefix, fmt.Sprintf("%s%s", module, DiscoveryFileName)),
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

func (disc *nodeDiscovery) Start() error {
	var err error
	disc.nodeController, err = commDiscovery.NewNodeController(disc.zkAddr, disc)
	if err != nil {
		return err
	}

	go disc.syncTickerPromSdConfig()
	disc.initSuccess = true
	disc.eventHandler(disc.module)
	return nil
}

func (disc *nodeDiscovery) GetDiscoveryKey() string {
	return disc.module
}

func (disc *nodeDiscovery) GetPrometheusSdConfig() ([]*types.PrometheusSdConfig, error) {
	nodes, err := disc.nodeController.List(commDiscovery.EverythingSelector())
	if err != nil {
		return nil, err
	}

	promConfigs := make([]*types.PrometheusSdConfig, 0)
	for _, node := range nodes {
		ip := node.GetAgentIP()
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

func (disc *nodeDiscovery) GetPromSdConfigFile() string {
	return disc.sdFilePath
}

func (disc *nodeDiscovery) RegisterEventFunc(handleFunc EventHandleFunc) {
	disc.eventHandler = handleFunc
}

func (disc *nodeDiscovery) OnAdd(obj interface{}) {
	if !disc.initSuccess {
		return
	}

	disc.eventHandler(disc.module)
}

// if on update event, then don't need to update sd config
func (disc *nodeDiscovery) OnUpdate(old, cur interface{}) {
	if !disc.initSuccess {
		return
	}
}

func (disc *nodeDiscovery) OnDelete(obj interface{}) {
	if !disc.initSuccess {
		return
	}

	// call event handler
	disc.eventHandler(disc.module)
}

func (disc *nodeDiscovery) syncTickerPromSdConfig() {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	select {
	case <-ticker.C:
		blog.V(3).Infof("ticker sync prometheus service discovery config")
		disc.eventHandler(disc.module)
	}
}
