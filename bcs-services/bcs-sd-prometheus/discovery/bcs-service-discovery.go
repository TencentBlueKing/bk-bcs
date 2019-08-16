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

const (
	DefaultbcsServiceDiscoveryKey      = "bcsServiceDiscovery"
	DefaultbcsServiceDiscoveryFileName = "bcs_service_sd_config.json"
)

type bcsServiceDiscovery struct {
	zkAddr     string
	key        string
	sdFilePath string

	eventHandler    EventHandleFunc
	moduleDiscovery moduleDiscovery.ModuleDiscovery
	modules         []string
}

// new bcs service module service discovery
func NewBcsServiceDiscovery(zkAddr string, promFilePrefix string) (Discovery, error) {
	disc := &bcsServiceDiscovery{
		zkAddr:     zkAddr,
		key:        DefaultbcsServiceDiscoveryKey,
		sdFilePath: path.Join(promFilePrefix, DefaultbcsServiceDiscoveryFileName),
		modules: []string{
			commtypes.BCS_MODULE_APISERVER, commtypes.BCS_MODULE_STORAGE, commtypes.BCS_MODULE_NETSERVICE,
		},
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
	return disc.key
}

func (disc *bcsServiceDiscovery) GetPrometheusSdConfig() ([]*types.PrometheusSdConfig, error) {
	promConfigs := make([]*types.PrometheusSdConfig, 0)
	for _, module := range disc.modules {
		servs, err := disc.moduleDiscovery.GetModuleServers(module)
		if err != nil {
			blog.Errorf("discovery %s get module %s error %s", disc.key, module, err.Error())
			continue
		}

		for _, serv := range servs {
			var conf *types.PrometheusSdConfig
			switch module {
			case commtypes.BCS_MODULE_APISERVER:
				ser, ok := serv.(*commtypes.APIServInfo)
				if !ok {
					blog.Errorf("discovery %s module %s failed convert to APIServInfo", disc.key, module)
					break
				}

				conf = &types.PrometheusSdConfig{
					Targets: []string{fmt.Sprintf("%s:%d", ser.IP, ser.MetricPort)},
					Labels: map[string]string{
						DefaultBcsModuleLabelKey: module,
					},
				}

			case commtypes.BCS_MODULE_STORAGE:
				ser, ok := serv.(*commtypes.BcsStorageInfo)
				if !ok {
					blog.Errorf("discovery %s module %s failed convert to BcsStorageInfo", disc.key, module)
					break
				}

				conf = &types.PrometheusSdConfig{
					Targets: []string{fmt.Sprintf("%s:%d", ser.IP, ser.MetricPort)},
					Labels: map[string]string{
						DefaultBcsModuleLabelKey: module,
					},
				}

			case commtypes.BCS_MODULE_NETSERVICE:
				ser, ok := serv.(*commtypes.NetServiceInfo)
				if !ok {
					blog.Errorf("discovery %s module %s failed convert to NetServiceInfo", disc.key, module)
					break
				}

				conf = &types.PrometheusSdConfig{
					Targets: []string{fmt.Sprintf("%s:%d", ser.IP, ser.MetricPort)},
					Labels: map[string]string{
						DefaultBcsModuleLabelKey: module,
					},
				}

			default:
				blog.Errorf("discovery %s module %s not found", disc.key, module)
			}

			if conf != nil {
				promConfigs = append(promConfigs, conf)
			}

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
	blog.Infof("discovery %s handle module %s event", disc.key, module)
	disc.eventHandler(disc.GetDiscoveryKey())
}

func (disc *bcsServiceDiscovery) syncTickerPromSdConfig() {
	ticker := time.NewTicker(time.Minute * 5)

	select {
	case <-ticker.C:
		blog.V(3).Infof("ticker sync prometheus service discovery config")
		disc.eventHandler(disc.GetDiscoveryKey())
	}
}
