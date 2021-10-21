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
	"encoding/json"
	"fmt"
	"path"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	moduleDiscovery "github.com/Tencent/bk-bcs/bcs-common/pkg/module-discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-service-prometheus/types"
)

type bcsMesosDiscovery struct {
	zkAddr     string
	sdFilePath string

	eventHandler    EventHandleFunc
	moduleDiscovery moduleDiscovery.ModuleDiscovery
	module          []string
	promFilePrefix  string
}

// NewBcsDiscovery new bcs module service discovery
func NewBcsDiscovery(zkAddr string, promFilePrefix string, module []string) (Discovery, error) {
	disc := &bcsMesosDiscovery{
		zkAddr:         zkAddr,
		promFilePrefix: promFilePrefix,
		module:         module,
	}

	return disc, nil
}

// start discovery
func (disc *bcsMesosDiscovery) Start() error {
	var err error
	disc.moduleDiscovery, err = moduleDiscovery.NewDiscoveryV2(disc.zkAddr, disc.module)
	if err != nil {
		return err
	}
	disc.moduleDiscovery.RegisterEventFunc(disc.handleEventFunc)
	go disc.syncTickerPromSdConfig()
	return nil
}

// get prometheus service discovery config
func (disc *bcsMesosDiscovery) GetPrometheusSdConfig(module string) ([]*types.PrometheusSdConfig, error) {
	promConfigs := make([]*types.PrometheusSdConfig, 0)
	servs, err := disc.moduleDiscovery.GetModuleServers(module)
	if err != nil {
		blog.Errorf("discovery %s get disc.module %s error %s", module, module, err.Error())
		return nil, err
	}

	for _, serv := range servs {
		//serv is string object
		data, _ := serv.(string)
		var servInfo *commtypes.ServerInfo
		err = json.Unmarshal([]byte(data), &servInfo)
		if err != nil {
			blog.Errorf("getModuleAddr Unmarshal data(%s) to commtypes.BcsMesosApiserverInfo failed: %s", data, err.Error())
			continue
		}
		conf := &types.PrometheusSdConfig{
			Targets: []string{fmt.Sprintf("%s:%d", servInfo.IP, servInfo.MetricPort)},
			Labels: map[string]string{
				DefaultBcsModuleLabelKey: module,
			},
		}
		promConfigs = append(promConfigs, conf)
	}

	return promConfigs, nil
}

// get prometheus sd config file path
func (disc *bcsMesosDiscovery) GetPromSdConfigFile(module string) string {
	return path.Join(disc.promFilePrefix, fmt.Sprintf("%s%s", module, DiscoveryFileName))
}

//register event handle function
func (disc *bcsMesosDiscovery) RegisterEventFunc(handleFunc EventHandleFunc) {
	disc.eventHandler = handleFunc
}

func (disc *bcsMesosDiscovery) handleEventFunc(module string) {
	blog.Infof("discovery %s handle module %s event", disc.module, module)
	disc.eventHandler(Info{Module: module, Key: module})
}

func (disc *bcsMesosDiscovery) syncTickerPromSdConfig() {
	for _, module := range disc.module {
		disc.eventHandler(Info{Module: module, Key: module})
	}
	ticker := time.NewTicker(time.Minute * 5)
	select {
	case <-ticker.C:
		blog.V(3).Infof("ticker sync prometheus service discovery config")
		for _, module := range disc.module {
			disc.eventHandler(Info{Module: module, Key: module})
		}
	}
}
