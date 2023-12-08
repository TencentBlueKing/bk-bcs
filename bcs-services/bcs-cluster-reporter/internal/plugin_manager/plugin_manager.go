/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package plugin_manager
package plugin_manager

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metric_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Pm xxx
	Pm           *pluginManager
	clusterTotal *prometheus.GaugeVec
)

// Plugin xxx
type Plugin interface {
	Name() string
	Setup(configFilePath string) error
	Stop() error
}

func init() {
	Pm = NewPluginManager()
	// set default metric
	clusterTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cluster_total_num",
		Help: "cluster_total_num",
	}, []string{})

	prometheus.MustRegister(clusterTotal)
}

// Register xxx
func Register(plugin Plugin) {
	Pm.Register(plugin)
}

type pluginManager struct {
	plugins         map[string]Plugin
	config          *Config
	configLock      sync.Mutex
	concurrencyLock sync.Mutex
	routinePool     *util.RoutinePool
}

func (pm *pluginManager) Register(plugin Plugin) {
	pm.plugins[plugin.Name()] = plugin
}

func (pm *pluginManager) GetPlugin(plugin string) Plugin {
	if p, ok := pm.plugins[plugin]; ok {
		return p
	} else {
		return nil
	}
}

// SetConfig configure pluginmanager by config file
func (pm *pluginManager) SetConfig(config *Config) {
	pm.configLock.Lock()
	defer pm.configLock.Unlock()
	if config != nil {
		pm.config = config
	}

	for _, cluster := range config.ClusterConfigs {
		metric_manager.MM.SetSeperatedMetric(cluster.ClusterID)
	}

	clusterTotal.WithLabelValues().Set(float64(len(pm.config.ClusterConfigs)))
}

func (pm *pluginManager) GetConfig() *Config {
	pm.configLock.Lock()
	defer pm.configLock.Unlock()
	return pm.config
}

func (pm *pluginManager) SetupPlugin(plugins string, pluginDir string) error {
	for _, plugin := range strings.Split(plugins, ",") {
		if p := pm.GetPlugin(plugin); p == nil {
			return fmt.Errorf("Get Plugin %s failed, nil result", plugin)
		} else {
			err := p.Setup(filepath.Join(pluginDir, plugin+".conf"))
			if err != nil {
				return fmt.Errorf("Setup plugin %s failed: %s", p.Name(), err.Error())
			}
		}
	}
	return nil
}

func (pm *pluginManager) Lock() {
	pm.concurrencyLock.Lock()
}

func (pm *pluginManager) UnLock() {
	pm.concurrencyLock.Unlock()
}

func (pm *pluginManager) Add() {
	pm.routinePool.Add(1)
}

func (pm *pluginManager) Done() {
	pm.routinePool.Done()
}

func (pm *pluginManager) StopPlugin(plugins string) error {
	for _, plugin := range strings.Split(plugins, ",") {
		if p := pm.GetPlugin(plugin); p == nil {
			return fmt.Errorf("Get Plugin %s failed, nil result", plugin)
		} else {
			err := p.Stop()
			if err != nil {
				return fmt.Errorf("StopPlugin plugin %s failed: %s", p.Name(), err.Error())
			}
		}
	}
	return nil
}

// NewPluginManager xxx
func NewPluginManager() *pluginManager {
	return &pluginManager{
		routinePool: util.NewRoutinePool(40),
		plugins:     make(map[string]Plugin),
	}
}
