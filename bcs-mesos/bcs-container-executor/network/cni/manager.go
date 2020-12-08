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

package cni

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/container"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/logs"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/network"

	"github.com/containernetworking/cni/libcni"
)

//NewNetManager create cni plugin manager
func NewNetManager(binpath, confpath string) network.NetManager {
	if confpath == "" {
		return nil
	}
	//check confpath exist
	if exist, _ := util.FileExists(confpath); !exist {
		return nil
	}
	manager := &PluginManager{
		binDir:  binpath,
		confDir: confpath,
		plugins: make(map[string]network.NetworkPlugin),
	}
	return manager
}

//PluginManager manager for all cni plugins
type PluginManager struct {
	binDir  string                           //executable binary  directory for CNI plugin
	confDir string                           //config directory for CNI configuration
	plugins map[string]network.NetworkPlugin //all plugins
}

//Init loading all configuration in directory
func (manager *PluginManager) Init() error {
	logs.Infof("CNI plugin manager init plugin's configuration under %s\n", manager.confDir)
	//list all .json or .conf file
	confFiles, err := libcni.ConfFiles(manager.confDir, []string{".conf", ".json"})
	if err != nil {
		logs.Errorf("CNI plugin manager load configuration in %s faile: %s\n", manager.confDir, err.Error())
		return err
	}
	if len(confFiles) == 0 {
		logs.Infof("No CNI configuration in %s\n", manager.confDir)
		return nil
	}
	//reading all configuration & create CNIPlugin
	for _, file := range confFiles {
		pluginName, plugin := NewPlugin(manager.binDir, file)
		if plugin == nil {
			logs.Errorf("CNI plugin manager create plugin by config file %s failed!\n", file)
			//return fmt.Errorf("Create %s plugin error", file)
			continue
		}
		if _, ok := manager.plugins[pluginName]; ok {
			//plugin name conflict, init failed
			logs.Errorf("CNI plugin named %s conflict in config file %s\n", pluginName, file)
			return fmt.Errorf("conflict plugin name: %s", file)
		}
		if perr := plugin.Init(""); perr != nil {
			logs.Errorf("CNI plugin manager init plugin %s err: %s\n", pluginName, perr.Error())
			return perr
		}
		logs.Infof("CNI plugin manager add plugin %s success.", pluginName)
		manager.plugins[pluginName] = plugin
	}
	return nil
}

//Stop manager stop if necessary
func (manager *PluginManager) Stop() {
	//empty
}

//GetPlugin get plugin by name
func (manager *PluginManager) GetPlugin(name string) network.NetworkPlugin {
	if plugin, ok := manager.plugins[name]; ok {
		return plugin
	}
	return nil
}

//AddPlugin Add plugin to manager dynamic if necessary
func (manager *PluginManager) AddPlugin(name string, plugin network.NetworkPlugin) error {
	if _, ok := manager.plugins[name]; ok {
		return fmt.Errorf("conflict plugin with name %s", name)
	}
	manager.plugins[name] = plugin
	return nil
}

//SetUpPod for setting Pod network interface
func (manager *PluginManager) SetUpPod(podInfo container.Pod) error {
	logs.Infof("CNI plugin manager ADD pod %s network with %s\n", podInfo.GetContainerID(), podInfo.GetNetworkName())
	if podInfo.GetNetworkName() == "none" {
		logs.Infoln("CNI plugin manager skip pod setup because of none network mode")
		return nil
	}
	if podInfo.GetNetworkName() == "host" {
		addr := util.GetIPAddress()
		podInfo.SetIPAddr(addr)
		logs.Infoln(fmt.Sprintf("CNI plugin manager skip pod setup with host mode, setting pod ipaddr %s", addr))
		return nil
	}

	if plugin, ok := manager.plugins[podInfo.GetNetworkName()]; ok {
		if pErr := plugin.SetUpPod(podInfo); pErr != nil {
			logs.Errorf("CNI plugin %s/%s SetUpPod %s err: %s\n", podInfo.GetNetworkName(), plugin.Name(), podInfo.GetContainerID(), pErr.Error())
			return pErr
		}
		logs.Infof("CNI plugin manager ADD pod %s network with %s success.\n", podInfo.GetContainerID(), podInfo.GetNetworkName())
		return nil
	}
	logs.Errorf("CNI plugin Manager get no plugin named %s\n", podInfo.GetNetworkName())
	return fmt.Errorf("No CNI plugin name called %s", podInfo.GetNetworkName())
}

//TearDownPod for release pod network resource
func (manager *PluginManager) TearDownPod(podInfo container.Pod) error {
	logs.Infof("CNI plugin manager DEL pod %s network with %s\n", podInfo.GetContainerID(), podInfo.GetNetworkName())
	if podInfo.GetNetworkName() == "host" || podInfo.GetNetworkName() == "none" {
		logs.Infoln("CNI plugin manager skip host pod teardown...")
		return nil
	}
	if plugin, ok := manager.plugins[podInfo.GetNetworkName()]; ok {
		if pErr := plugin.TearDownPod(podInfo); pErr != nil {
			logs.Errorf("CNI plugin %s/%s TearDownPod %s err: %s\n", podInfo.GetNetworkName(), plugin.Name(), podInfo.GetContainerID(), pErr.Error())
			return pErr
		}
		logs.Infof("CNI plugin manager DEL pod %s network with %s success.\n", podInfo.GetContainerID(), podInfo.GetNetworkName())
		return nil
	}
	logs.Errorf("CNI plugin Manager get no plugin named %s\n", podInfo.GetNetworkName())
	return fmt.Errorf("No CNI plugin name called %s", podInfo.GetNetworkName())
}
