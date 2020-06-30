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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	moduleDiscovery "github.com/Tencent/bk-bcs/bcs-common/pkg/module-discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-sd-prometheus/types"
)

type bcsMesosDiscovery struct {
	zkAddr     string
	sdFilePath string

	eventHandler    EventHandleFunc
	moduleDiscovery moduleDiscovery.ModuleDiscovery
	module          string
}

// new bcs mesos module service discovery
func NewBcsMesosDiscovery(zkAddr string, promFilePrefix string, module string) (Discovery, error) {
	disc := &bcsMesosDiscovery{
		zkAddr:     zkAddr,
		sdFilePath: path.Join(promFilePrefix, fmt.Sprintf("%s%s", module, DiscoveryFileName)),
		module:     module,
	}

	return disc, nil
}

// start discovery
func (disc *bcsMesosDiscovery) Start() error {
	var err error
	disc.moduleDiscovery, err = moduleDiscovery.NewMesosDiscovery(disc.zkAddr)
	if err != nil {
		return err
	}
	disc.moduleDiscovery.RegisterEventFunc(disc.handleEventFunc)
	go disc.syncTickerPromSdConfig()
	return nil
}

// get the discovery key
func (disc *bcsMesosDiscovery) GetDiscoveryKey() string {
	return disc.module
}

// get prometheus service discovery config
func (disc *bcsMesosDiscovery) GetPrometheusSdConfig() ([]*types.PrometheusSdConfig, error) {
	promConfigs := make([]*types.PrometheusSdConfig, 0)
	servs, err := disc.moduleDiscovery.GetModuleServers(disc.module)
	if err != nil {
		blog.Errorf("discovery %s get disc.module %s error %s", disc.module, disc.module, err.Error())
		return nil, err
	}

	for _, serv := range servs {
		var conf *types.PrometheusSdConfig
		switch disc.module {
		case commtypes.BCS_MODULE_SCHEDULER:
			ser, ok := serv.(*commtypes.SchedulerServInfo)
			if !ok {
				blog.Errorf("discovery %s disc.module %s failed convert to SchedulerServInfo", disc.module, disc.module)
				break
			}

			conf = &types.PrometheusSdConfig{
				Targets: []string{fmt.Sprintf("%s:%d", ser.IP, ser.MetricPort)},
				Labels: map[string]string{
					DefaultBcsModuleLabelKey: disc.module,
				},
			}

		case commtypes.BCS_MODULE_MESOSAPISERVER:
			ser, ok := serv.(*commtypes.BcsMesosApiserverInfo)
			if !ok {
				blog.Errorf("discovery %s disc.module %s failed convert to MesosDriverServInfo", disc.module, disc.module)
				break
			}

			conf = &types.PrometheusSdConfig{
				Targets: []string{fmt.Sprintf("%s:%d", ser.IP, ser.MetricPort)},
				Labels: map[string]string{
					DefaultBcsModuleLabelKey: disc.module,
				},
			}

		case commtypes.BCS_MODULE_MESOSDATAWATCH:
			ser, ok := serv.(*commtypes.MesosDataWatchServInfo)
			if !ok {
				blog.Errorf("discovery %s disc.module %s failed convert to MesosDataWatchServInfo", disc.module, disc.module)
				break
			}

			conf = &types.PrometheusSdConfig{
				Targets: []string{fmt.Sprintf("%s:%d", ser.IP, ser.MetricPort)},
				Labels: map[string]string{
					DefaultBcsModuleLabelKey: disc.module,
				},
			}

		case commtypes.BCS_MODULE_DNS:
			ser, ok := serv.(*commtypes.DNSInfo)
			if !ok {
				blog.Errorf("discovery %s disc.module %s failed convert to DNSInfo", disc.module, disc.module)
				break
			}

			conf = &types.PrometheusSdConfig{
				Targets: []string{fmt.Sprintf("%s:%d", ser.IP, ser.MetricPort)},
				Labels: map[string]string{
					DefaultBcsModuleLabelKey: disc.module,
				},
			}

		case commtypes.BCS_MODULE_LOADBALANCE:
			ser, ok := serv.(*commtypes.LoadBalanceInfo)
			if !ok {
				blog.Errorf("discovery %s disc.module %s failed convert to DNSInfo", disc.module, disc.module)
				break
			}

			conf = &types.PrometheusSdConfig{
				Targets: []string{fmt.Sprintf("%s:%d", ser.IP, ser.MetricPort)},
				Labels: map[string]string{
					DefaultBcsModuleLabelKey: disc.module,
				},
			}

		default:
			blog.Errorf("discovery %s disc.module %s not found", disc.module, disc.module)
		}

		if conf != nil {
			promConfigs = append(promConfigs, conf)
		}

	}

	return promConfigs, nil
}

// get prometheus sd config file path
func (disc *bcsMesosDiscovery) GetPromSdConfigFile() string {
	return disc.sdFilePath
}

//register event handle function
func (disc *bcsMesosDiscovery) RegisterEventFunc(handleFunc EventHandleFunc) {
	disc.eventHandler = handleFunc
}

func (disc *bcsMesosDiscovery) handleEventFunc(module string) {
	blog.Infof("discovery %s handle module %s event", disc.module, module)
	disc.eventHandler(disc.GetDiscoveryKey())
}

func (disc *bcsMesosDiscovery) syncTickerPromSdConfig() {
	ticker := time.NewTicker(time.Minute * 5)

	select {
	case <-ticker.C:
		blog.V(3).Infof("ticker sync prometheus service discovery config")
		disc.eventHandler(disc.GetDiscoveryKey())
	}
}
