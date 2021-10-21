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

package controller

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-service-prometheus/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-service-prometheus/discovery"
)

const (
	// ServiceMonitorModule name
	ServiceMonitorModule = "ServiceMonitor"
)

// PrometheusController controller for prometheus service discovery
type PrometheusController struct {
	promFilePrefix string
	clusterID      string
	conf           *config.Config

	discoverys     map[string]discovery.Discovery
	mesosModules   []string
	serviceModules []string
	nodeModules    []string
	serviceMonitor string
}

// NewPrometheusController new prometheus controller
func NewPrometheusController(conf *config.Config) *PrometheusController {
	prom := &PrometheusController{
		conf:           conf,
		clusterID:      conf.ClusterID,
		promFilePrefix: conf.PromFilePrefix,
		discoverys:     make(map[string]discovery.Discovery),
		mesosModules: []string{commtypes.BCS_MODULE_SCHEDULER, commtypes.BCS_MODULE_MESOSDATAWATCH, commtypes.BCS_MODULE_MESOSAPISERVER,
			commtypes.BCS_MODULE_DNS, commtypes.BCS_MODULE_LOADBALANCE},
		serviceModules: []string{commtypes.BCS_MODULE_APISERVER, commtypes.BCS_MODULE_STORAGE, commtypes.BCS_MODULE_NETSERVICE},
		nodeModules:    []string{discovery.CadvisorModule, discovery.NodeexportModule},
		serviceMonitor: ServiceMonitorModule,
	}
	if len(conf.ServiceModules) > 0 {
		prom.serviceModules = conf.ServiceModules
	}
	if len(conf.ClusterModules) > 0 {
		prom.mesosModules = conf.ClusterModules
	}

	return prom
}

// Start to work update prometheus sd config
func (prom *PrometheusController) Start() error {
	//init bcs mesos module discovery
	if prom.conf.EnableMesos {
		dis, err := discovery.NewBcsDiscovery(prom.conf.ClusterZk, prom.promFilePrefix, prom.mesosModules)
		if err != nil {
			blog.Errorf("NewBcsDiscovery ClusterZk %s error %s", prom.conf.ClusterZk, err.Error())
			return err
		}
		//register event handle function
		dis.RegisterEventFunc(prom.handleDiscoveryEvent)
		for _, module := range prom.mesosModules {
			prom.discoverys[module] = dis
		}
		err = dis.Start()
		if err != nil {
			blog.Errorf("mesosDiscovery start failed: %s", err.Error())
			return err
		}
	}

	//init node discovery
	if prom.conf.EnableNode {
		for _, module := range prom.nodeModules {
			var nodeDiscovery discovery.Discovery
			var err error
			if prom.conf.Kubeconfig != "" {
				nodeDiscovery, err = discovery.NewNodeEtcdDiscovery(prom.conf.Kubeconfig, prom.promFilePrefix, module, prom.conf.CadvisorPort, prom.conf.NodeExportPort)
			} else {
				zkAddr := strings.Split(prom.conf.ClusterZk, ",")
				nodeDiscovery, err = discovery.NewNodeZkDiscovery(zkAddr, prom.promFilePrefix, module, prom.conf.CadvisorPort, prom.conf.NodeExportPort)

			}
			if err != nil {
				blog.Errorf("NewNodeDiscovery ClusterZk %s error %s", prom.conf.ClusterZk, err.Error())
				return err
			}
			//register event handle function
			nodeDiscovery.RegisterEventFunc(prom.handleDiscoveryEvent)
			prom.discoverys[module] = nodeDiscovery
			err = nodeDiscovery.Start()
			if err != nil {
				blog.Errorf("nodeDiscovery start failed: %s", err.Error())
			}
		}
	}

	//init bcs service module discovery
	if prom.conf.EnableService {
		serviceDiscovery, err := discovery.NewBcsDiscovery(prom.conf.ServiceZk, prom.promFilePrefix, prom.serviceModules)
		if err != nil {
			blog.Errorf("NewBcsDiscovery ClusterZk %s error %s", prom.conf.ServiceZk, err.Error())
			return err
		}
		//register event handle function
		serviceDiscovery.RegisterEventFunc(prom.handleDiscoveryEvent)
		for _, module := range prom.serviceModules {
			prom.discoverys[module] = serviceDiscovery
		}
		err = serviceDiscovery.Start()
		if err != nil {
			blog.Errorf("serviceDiscovery start failed: %s", err.Error())
			return err
		}
	}

	//init taskgroup ServiceMonitor discovery
	if prom.conf.EnableServiceMonitor {
		serviceDiscovery, err := discovery.NewServiceMonitor(prom.conf.Kubeconfig, prom.promFilePrefix, prom.serviceMonitor)
		if err != nil {
			blog.Errorf("NewBcsDiscovery ClusterZk %s error %s", prom.conf.ServiceZk, err.Error())
			return err
		}
		//register event handle function
		serviceDiscovery.RegisterEventFunc(prom.handleDiscoveryEvent)
		prom.discoverys[prom.serviceMonitor] = serviceDiscovery
		err = serviceDiscovery.Start()
		if err != nil {
			blog.Errorf("serviceDiscovery start failed: %s", err.Error())
			return err
		}
	}

	return nil
}

func (prom *PrometheusController) handleDiscoveryEvent(dInfo discovery.Info) {
	blog.Infof("discovery %s service discovery config changed", dInfo.Module)
	disc, ok := prom.discoverys[dInfo.Module]
	if !ok {
		blog.Errorf("not found discovery %s", dInfo.Module)
		return
	}

	sdConfig, err := disc.GetPrometheusSdConfig(dInfo.Key)
	if err != nil {
		//if serviceMonitor Not Found, then delete the specify config file
		if strings.Contains(err.Error(), "Not Found") {
			prom.deletePrometheusConfigFile(dInfo)
			return
		}
		blog.Errorf("discovery %s get prometheus service discovery config error %s", dInfo.Key, err.Error())
		return
	}
	by, _ := json.Marshal(sdConfig)

	file, err := os.OpenFile(disc.GetPromSdConfigFile(dInfo.Key), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		blog.Errorf("open/create file %s error %s", disc.GetPromSdConfigFile(dInfo.Key), err.Error())
		return
	}
	defer file.Close()

	err = file.Truncate(0)
	if err != nil {
		blog.Errorf("Truncate file %s error %s", disc.GetPromSdConfigFile(dInfo.Key), err.Error())
		return
	}
	_, err = file.Write(by)
	if err != nil {
		blog.Errorf("write file %s error %s", disc.GetPromSdConfigFile(dInfo.Key), err.Error())
		return
	}

	blog.Infof("discovery %s write config file %s success", dInfo.Key, disc.GetPromSdConfigFile(dInfo.Key))
}

func (prom *PrometheusController) deletePrometheusConfigFile(dInfo discovery.Info) {
	disc, ok := prom.discoverys[dInfo.Module]
	if !ok {
		blog.Errorf("not found discovery %s", dInfo.Module)
		return
	}
	cFile := disc.GetPromSdConfigFile(dInfo.Key)
	err := os.Remove(cFile)
	if err != nil {
		blog.Errorf("remove config file(%s) error %s", cFile, err.Error())
		return
	}
	blog.Infof("remove config file(%s) success", cFile)
}
