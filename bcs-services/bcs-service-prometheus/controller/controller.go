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
	"sync"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-services/bcs-service-prometheus/config"
	"bk-bcs/bcs-services/bcs-service-prometheus/discovery"
)

type PrometheusController struct {
	sync.RWMutex

	promFilePrefix string
	clusterId      string
	conf           *config.Config

	discoverys map[string]discovery.Discovery
}

func NewPrometheusController(conf *config.Config) *PrometheusController {
	prom := &PrometheusController{
		conf:           conf,
		promFilePrefix: conf.PromFilePrefix,
		discoverys:     make(map[string]discovery.Discovery),
	}

	return prom
}

func (prom *PrometheusController) Start() error {
	//init bcs mesos module discovery
	bcsDiscovery, err := discovery.NewBcsDiscovery(prom.conf.ServiceZk, prom.promFilePrefix)
	if err != nil {
		blog.Errorf("NewBcsDiscovery ClusterZk %s error %s", prom.conf.ServiceZk, err.Error())
		return err
	}
	err = bcsDiscovery.Start()
	if err != nil {
		blog.Errorf("nodeDiscovery start failed: %s", err.Error())
	}
	//register event handle function
	bcsDiscovery.RegisterEventFunc(prom.handleDiscoveryEvent)
	prom.discoverys[bcsDiscovery.GetDiscoveryKey()] = bcsDiscovery

	return nil
}

func (prom *PrometheusController) handleDiscoveryEvent(discoveryKey string) {
	prom.Lock()
	defer prom.Unlock()

	blog.Infof("discovery %s service discovery config changed", discoveryKey)
	disc, ok := prom.discoverys[discoveryKey]
	if !ok {
		blog.Errorf("not found discovery %s", discoveryKey)
		return
	}

	sdConfig, err := disc.GetPrometheusSdConfig()
	if err != nil {
		blog.Errorf("discovery %s get prometheus service discovery config error %s", discoveryKey, err.Error())
		return
	}
	by, _ := json.Marshal(sdConfig)

	file, err := os.OpenFile(disc.GetPromSdConfigFile(), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		blog.Errorf("open/create file %s error %s", disc.GetPromSdConfigFile(), err.Error())
		return
	}
	defer file.Close()

	err = file.Truncate(0)
	if err != nil {
		blog.Errorf("Truncate file %s error %s", disc.GetPromSdConfigFile(), err.Error())
		return
	}
	_, err = file.Write(by)
	if err != nil {
		blog.Errorf("write file %s error %s", disc.GetPromSdConfigFile(), err.Error())
		return
	}

	blog.Infof("discovery %s write config file %s success", discoveryKey, disc.GetPromSdConfigFile())
}
