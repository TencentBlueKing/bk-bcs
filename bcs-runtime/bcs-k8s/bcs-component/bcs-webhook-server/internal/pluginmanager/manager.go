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
 */

// Package pluginmanager xxx
package pluginmanager

import (
	"fmt"
	"path/filepath"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/plugin"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/options"
)

var (
	kubernetesHookPlugins map[string]plugin.Interface
	mesosHookPlugins      map[string]plugin.MesosPlugin
)

func init() {
	kubernetesHookPlugins = make(map[string]plugin.Interface)
	mesosHookPlugins = make(map[string]plugin.MesosPlugin)
}

// Register register plugin
func Register(name string, p plugin.Interface) {
	if len(name) == 0 {
		blog.Fatalf("plugin name cannot be empty")
		return
	}
	if _, ok := kubernetesHookPlugins[name]; ok {
		blog.Fatalf("plugin with name %s already exists", name)
		return
	}
	kubernetesHookPlugins[name] = p
}

// RegisterMesos register mesos plugin
func RegisterMesos(name string, mp plugin.MesosPlugin) {
	if len(name) == 0 {
		blog.Fatalf("mesos plugin name cannot be empty")
		return
	}
	if _, ok := mesosHookPlugins[name]; ok {
		blog.Fatalf("mesos plugin with name %s already exists", name)
		return
	}
	mesosHookPlugins[name] = mp
}

// Manager manager for plugins
type Manager struct {
	clusterMode            string
	configDir              string
	activePlugins          []plugin.Interface
	activePluginNames      []string
	activeMesosPlugins     []plugin.MesosPlugin
	activeMesosPluginNames []string
}

// NewManager create new manager
func NewManager(mode, configDir string) *Manager {
	return &Manager{
		clusterMode:            mode,
		configDir:              configDir,
		activePlugins:          make([]plugin.Interface, 0),
		activePluginNames:      make([]string, 0),
		activeMesosPlugins:     make([]plugin.MesosPlugin, 0),
		activeMesosPluginNames: make([]string, 0),
	}
}

// InitPlugins init plugins with given names
func (m *Manager) InitPlugins(names []string) error {
	for _, name := range names {
		switch m.clusterMode {
		case options.EngineTypeKubernetes:
			p, found := kubernetesHookPlugins[name]
			if !found {
				return fmt.Errorf("plugin with name %s not found", name)
			}
			configFilePath := filepath.Join(m.configDir, getPluginFile(name))
			blog.Infof("activate plugin %s with config file %s", name, configFilePath)
			err := p.Init(configFilePath)
			if err != nil {
				return fmt.Errorf("activate plugin %s with config file %s failed, err %s",
					name, configFilePath, err.Error())
			}
			m.activePlugins = append(m.activePlugins, p)
			m.activePluginNames = append(m.activePluginNames, name)
		case options.EngineTypeMesos:
			p, found := mesosHookPlugins[name]
			if !found {
				return fmt.Errorf("mesos plugin with name %s not found", name)
			}
			configFilePath := filepath.Join(m.configDir, getPluginFile(name))
			blog.Infof("activate mesos plugin %s with config file %s", name, configFilePath)
			err := p.Init(configFilePath)
			if err != nil {
				return fmt.Errorf("activate mesos plugin %s with config file %s failed, err %s",
					name, configFilePath, err.Error())
			}
			m.activeMesosPlugins = append(m.activeMesosPlugins, p)
			m.activeMesosPluginNames = append(m.activeMesosPluginNames, name)
		default:
			return fmt.Errorf("unsupported cluster mode %s", m.clusterMode)
		}
	}
	return nil
}

// GetKubernetesPlugins get k8s plugins
func (m *Manager) GetKubernetesPlugins() []plugin.Interface {
	return m.activePlugins
}

// GetKubernetesPluginNames get k8s plugin names
func (m *Manager) GetKubernetesPluginNames() []string {
	return m.activePluginNames
}

// GetMesosPlugins get mesos plugins
func (m *Manager) GetMesosPlugins() []plugin.MesosPlugin {
	return m.activeMesosPlugins
}

// GetMesosPluginNames get mesos plugin names
func (m *Manager) GetMesosPluginNames() []string {
	return m.activeMesosPluginNames
}

// ClosePlugins call plugin Close function
func (m *Manager) ClosePlugins() error {
	if m.clusterMode == options.EngineTypeKubernetes {
		for name, p := range m.activePlugins {
			if err := p.Close(); err != nil {
				blog.Warnf("close plugin %s failed, err %s", name, err.Error())
			}
		}
	}
	return nil
}
