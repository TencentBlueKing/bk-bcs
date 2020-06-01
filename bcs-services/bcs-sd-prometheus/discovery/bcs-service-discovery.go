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
	commtypes "bk-bcs/bcs-common/common/types"
	moduleDiscovery "bk-bcs/bcs-common/pkg/module-discovery"
	"bk-bcs/bcs-services/bcs-sd-prometheus/types"
)

type bcsServiceDiscovery struct {
	zkAddr     string
	sdFilePath string

	eventHandler    EventHandleFunc
	moduleDiscovery moduleDiscovery.ModuleDiscovery
	module          string
}

// new bcs service module service discovery
func NewBcsServiceDiscovery(zkAddr string, promFilePrefix string, module string) (Discovery, error) {
	disc := &bcsServiceDiscovery{
		zkAddr:     zkAddr,
		sdFilePath: path.Join(promFilePrefix, fmt.Sprintf("%s%s", module, DiscoveryFileName)),
		module:     module,
	}

	return disc, nil
}

// start
func (disc *bcsServiceDiscovery) Start() error {
	var err error
	disc.moduleDiscovery, err = moduleDiscovery.NewServiceDiscovery(disc.zkAddr)
	if err != nil {
		return err
	}
	disc.moduleDiscovery.RegisterEventFunc(disc.handleEventFunc)
	go disc.syncTickerPromSdConfig()
	return nil
}

func (disc *bcsServiceDiscovery) GetDiscoveryKey() string {
	return disc.module
}

func (disc *bcsServiceDiscovery) GetPrometheusSdConfig() ([]*types.PrometheusSdConfig, error) {
	promConfigs := make([]*types.PrometheusSdConfig, 0)
	servs, err := disc.moduleDiscovery.GetModuleServers(disc.module)
	if err != nil {
		blog.Errorf("discovery %s get module %s error %s", disc.module, disc.module, err.Error())
		return nil, err
	}

	for _, serv := range servs {
		var conf *types.PrometheusSdConfig
		switch disc.module {
		case commtypes.BCS_MODULE_APISERVER:
			ser, ok := serv.(*commtypes.APIServInfo)
			if !ok {
				blog.Errorf("discovery %s module %s failed convert to APIServInfo", disc.module, disc.module)
				break
			}

			conf = &types.PrometheusSdConfig{
				Targets: []string{fmt.Sprintf("%s:%d", ser.IP, ser.MetricPort)},
				Labels: map[string]string{
					DefaultBcsModuleLabelKey: disc.module,
				},
			}

		case commtypes.BCS_MODULE_STORAGE:
			ser, ok := serv.(*commtypes.BcsStorageInfo)
			if !ok {
				blog.Errorf("discovery %s module %s failed convert to BcsStorageInfo", disc.module, disc.module)
				break
			}

			conf = &types.PrometheusSdConfig{
				Targets: []string{fmt.Sprintf("%s:%d", ser.IP, ser.MetricPort)},
				Labels: map[string]string{
					DefaultBcsModuleLabelKey: disc.module,
				},
			}

		case commtypes.BCS_MODULE_NETSERVICE:
			ser, ok := serv.(*commtypes.NetServiceInfo)
			if !ok {
				blog.Errorf("discovery %s module %s failed convert to NetServiceInfo", disc.module, disc.module)
				break
			}

			conf = &types.PrometheusSdConfig{
				Targets: []string{fmt.Sprintf("%s:%d", ser.IP, ser.MetricPort)},
				Labels: map[string]string{
					DefaultBcsModuleLabelKey: disc.module,
				},
			}

		default:
			blog.Errorf("discovery %s module %s not found", disc.module, disc.module)
		}

		if conf != nil {
			promConfigs = append(promConfigs, conf)
		}

	}

	return promConfigs, nil
}

func (disc *bcsServiceDiscovery) GetPromSdConfigFile() string {
	return disc.sdFilePath
}

func (disc *bcsServiceDiscovery) RegisterEventFunc(handleFunc EventHandleFunc) {
	disc.eventHandler = handleFunc
}

func (disc *bcsServiceDiscovery) handleEventFunc(module string) {
	blog.Infof("discovery %s handle module %s event", disc.module, module)
	disc.eventHandler(disc.GetDiscoveryKey())
}

func (disc *bcsServiceDiscovery) syncTickerPromSdConfig() {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	select {
	case <-ticker.C:
		blog.V(3).Infof("ticker sync prometheus service discovery config")
		disc.eventHandler(disc.GetDiscoveryKey())
	}
}
