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

package dynamicPlugin

import (
	"fmt"
	"plugin"
	"time"

	typesplugin "github.com/Tencent/bk-bcs/bcs-common/common/plugin"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/pluginManager/config"
	bcsplugin "github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/pluginManager/plugin"
)

type dynamicPlugin struct {
	currentDir string
	name       string

	conf *config.PluginConfig

	timeout int

	initErr error

	getHostAttributes func(*typesplugin.HostPluginParameter) (map[string]*typesplugin.HostAttributes, error)
}

// NewDynamicPlugin loading plugin according configuration
func NewDynamicPlugin(dir string, conf *config.PluginConfig) (bcsplugin.Plugin, error) {
	p := &dynamicPlugin{
		currentDir: dir,
		conf:       conf,
		name:       conf.Name,
		timeout:    conf.Timeout,
	}

	err := p.initPlugin()
	if err != nil {
		err = fmt.Errorf("plugin %s init plugin error %s", p.name, err.Error())
		p.initErr = err
	}

	return p, nil
}

func (p *dynamicPlugin) initPlugin() error {

	pluginPath := fmt.Sprintf("%s/bin/%s/%s.so", p.currentDir, p.name, p.name)
	outPlugin, err := plugin.Open(pluginPath)
	if err != nil {
		err = fmt.Errorf("init plugin %s error %s", pluginPath, err.Error())
		return err
	}

	pluginInit, err := outPlugin.Lookup("Init")
	if err != nil {
		err = fmt.Errorf("plugin %s lookup func Init error %s", pluginPath, err.Error())
		return err
	}

	initFunc, ok := pluginInit.(func(*typesplugin.InitPluginParameter) error)
	if !ok {
		err = fmt.Errorf("plugin %s func Init convert to function error %s", pluginPath, err.Error())
		return err
	}

	pluginPara := &typesplugin.InitPluginParameter{
		ConfPath: fmt.Sprintf("%s/bin/%s", p.currentDir, p.name),
	}

	err = initFunc(pluginPara)
	if err != nil {
		return err
	}

	outScheduler, err := outPlugin.Lookup("GetHostAttributes")
	if err != nil {
		err = fmt.Errorf("plugin %s lookup func GetHostAttributes error %s", pluginPath, err.Error())
		return err
	}

	schedulerFunc, ok := outScheduler.(func(*typesplugin.HostPluginParameter) (map[string]*typesplugin.HostAttributes, error))
	if !ok {
		err = fmt.Errorf("plugin %s func GetHostAttributes convert to function error %s", pluginPath, err.Error())
		return err
	}

	p.getHostAttributes = schedulerFunc

	return nil
}

// GetHostAttributes interface implementation
func (p *dynamicPlugin) GetHostAttributes(para *typesplugin.HostPluginParameter) (map[string]*typesplugin.HostAttributes, error) {

	if p.initErr != nil {
		return nil, p.initErr
	}

	chAttr := make(chan map[string]*typesplugin.HostAttributes, 1)
	chErr := make(chan error, 1)

	go func() {
		attrs, err := p.getHostAttributes(para)
		chAttr <- attrs
		chErr <- err
	}()

	ticker := time.NewTicker(time.Second * time.Duration(p.timeout))
	defer ticker.Stop()

	select {
	case <-ticker.C:
		return nil, fmt.Errorf("plugin %s GetHostAttributes timeout %ds", p.name, p.timeout)

	case err := <-chErr:
		if err != nil {
			return nil, err
		}
	}

	return <-chAttr, nil
}
