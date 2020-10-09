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

package pluginManager

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	typesplugin "github.com/Tencent/bk-bcs/bcs-common/common/plugin"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/pluginManager/config"
	bcsplugin "github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/pluginManager/plugin"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/pluginManager/plugin/dynamicPlugin"
)

//PluginManager plugin manager
type PluginManager struct {
	lock sync.RWMutex

	pluginNames []string
	pluginDir   string
	confs       map[string]*config.PluginConfig

	plugins map[string]bcsplugin.Plugin
}

// NewPluginManager create plugin manager
func NewPluginManager(pluginNames []string, pluginDir string) (*PluginManager, error) {
	var err error
	dir := pluginDir
	if dir == "" {
		dir, err = getCurrentDirectory()
		if err != nil {
			return nil, err
		}
	}

	p := &PluginManager{
		pluginDir:   dir,
		plugins:     make(map[string]bcsplugin.Plugin),
		confs:       make(map[string]*config.PluginConfig),
		pluginNames: pluginNames,
	}

	p.initPlugins()

	return p, nil
}

func getCurrentDirectory() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/plugin", dir), nil
}

func (p *PluginManager) initPlugins() {
	paths := make([]string, 0)

	filepath.Walk(fmt.Sprintf("%s/conf", p.pluginDir),
		func(path string, f os.FileInfo, err error) error {
			if f == nil {
				return err
			}

			if f.IsDir() {
				return nil
			}

			paths = append(paths, path)
			return nil

		})

	for _, path := range paths {
		conf, err := config.NewConfig(path)
		if err != nil {
			blog.Errorf("init config %s error %s", path, err.Error())
			continue
		}

		var ok bool

		for _, name := range p.pluginNames {
			if name == conf.Name {
				ok = true
				break
			}
		}

		if !ok {
			blog.Errorf("plugin %s is disable", path)
			continue
		}

		p.confs[conf.Name] = conf
	}

	for _, conf := range p.confs {
		var err error
		var plugin bcsplugin.Plugin

		switch conf.Type {
		case config.DynamicPluginType:
			plugin, err = dynamicPlugin.NewDynamicPlugin(p.pluginDir, conf)

		default:
			err = fmt.Errorf("plugin type %s is invalid", conf.Type)
		}

		if err != nil {
			blog.Errorf("NewDynamicPlugin error %s", err.Error())
			continue
		}

		blog.Infof("init plugin %s success", conf.Name)
		p.plugins[conf.Name] = plugin
	}

	blog.Infof("initPlugins done")
}

// GetHostAttributes get mesos slave dynamic attributes
func (p *PluginManager) GetHostAttributes(para *typesplugin.HostPluginParameter) (map[string]*typesplugin.HostAttributes, error) {

	hosts := make(map[string]*typesplugin.HostAttributes)

	for _, ip := range para.Ips {
		attr := &typesplugin.HostAttributes{
			Ip: ip,
		}

		hosts[ip] = attr
	}

	for name, plugin := range p.plugins {
		blog.Infof("plugin %s start GetHostAttributes...", name)

		var isError bool

		attributes, err := plugin.GetHostAttributes(para)
		if err != nil {
			blog.Errorf("plugin %s GetHostAttributes ips %v error %s", name, para.Ips, err.Error())
			isError = true
		}

		if isError {
			if conf, ok := p.confs[name]; ok {

				for _, ip := range para.Ips {
					hosts[ip].Attributes = conf.DefaultAtrrs
				}

			}

			continue
		}

		for ip, attr := range attributes {
			if attr == nil {
				blog.Errorf("plugin %s get ip %s attributes is nil", name, ip)
				continue
			}

			hosts[ip] = attr
		}

		blog.Infof("plugin %s start GetHostAttributes done", name)
	}

	return hosts, nil
}
