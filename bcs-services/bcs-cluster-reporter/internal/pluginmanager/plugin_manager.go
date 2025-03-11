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
	"k8s.io/klog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Pm xxx
	Pm           *PluginManager
	clusterTotal *prometheus.GaugeVec
)

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

// PluginManager xxx
type PluginManager struct {
	plugins           map[string]Plugin
	config            *Config
	configLock        sync.Mutex
	concurrencyLock   sync.Mutex
	routinePool       *util.RoutinePool
	clusterReportList map[string]map[string]string
}

// Register xxx
func (pm *PluginManager) Register(plugin Plugin) {
	pm.plugins[plugin.Name()] = plugin
}

// GetPlugin xxx
func (pm *PluginManager) GetPlugin(plugin string) Plugin {
	if p, ok := pm.plugins[plugin]; ok {
		return p
	} else {
		return nil
	}
}

// GetPluginstr xxx
func (pm *PluginManager) GetPluginstr() string {
	result := ""
	for name, _ := range pm.plugins {
		result = fmt.Sprintf("%s,%s", name, result)
	}
	result = strings.TrimSuffix(result, ",")
	return result
}

// SetConfig configure pluginmanager by config file
func (pm *PluginManager) SetConfig(config *Config) {
	pm.configLock.Lock()
	defer pm.configLock.Unlock()
	if config != nil {
		pm.config = config
	}

	clusterTotal.WithLabelValues().Set(float64(len(pm.config.ClusterConfigs)))
}

// GetConfig xxx
func (pm *PluginManager) GetConfig() *Config {
	pm.configLock.Lock()
	defer pm.configLock.Unlock()
	return pm.config
}

// SetClusterReport xxx
func (pm *PluginManager) SetClusterReport(clusterID, name, report string) {
	pm.clusterReportList[clusterID][name] = report
}

// SetupPlugin xxx
func (pm *PluginManager) SetupPlugin(plugins string, pluginDir string, runMode string) error {
	var wg sync.WaitGroup
	for _, plugin := range strings.Split(plugins, ",") {
		if p := pm.GetPlugin(plugin); p == nil {
			return fmt.Errorf("Get Plugin %s failed, nil result", plugin)
		} else {
			wg.Add(1)
			go func(plugin string) {
				err := p.Setup(filepath.Join(pluginDir, plugin+".conf"), runMode)
				if err != nil {
					klog.Fatalf("Setup plugin %s failed: %s", p.Name(), err.Error())
				}
				wg.Done()
			}(plugin)
		}
	}

	for plugin, _ := range pm.plugins {
		if !strings.Contains(plugins, plugin) {
			delete(pm.plugins, plugin)
		}
	}
	wg.Wait()
	return nil
}

// Lock xxx
func (pm *PluginManager) Lock() {
	pm.concurrencyLock.Lock()
}

// UnLock xxx
func (pm *PluginManager) UnLock() {
	pm.concurrencyLock.Unlock()
}

// Add xxx
func (pm *PluginManager) Add() {
	pm.routinePool.Add(1)
}

// Done xxx
func (pm *PluginManager) Done() {
	pm.routinePool.Done()
}

// StopPlugin xxx
func (pm *PluginManager) StopPlugin(plugins string) error {
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
func NewPluginManager() *PluginManager {
	return &PluginManager{
		routinePool:       util.NewRoutinePool(80),
		plugins:           make(map[string]Plugin),
		clusterReportList: make(map[string]map[string]string),
	}
}

// Ready xxx
func (pm *PluginManager) Ready(pluginStr string, targetID string) bool {
	for _, plugin := range strings.Split(pluginStr, ",") {
		p := pm.GetPlugin(plugin)
		if p == nil {
			continue
		}
		for {
			if p.Ready(targetID) {
				break
			}

			time.Sleep(5 * time.Second)
			if targetID != "" && targetID != "node" {
				klog.Infof("%s for %s is not ready", plugin, targetID)
				if _, ok := pm.GetConfig().ClusterConfigs[targetID]; !ok {
					return false
				}
			} else {
				klog.Infof("%s is not ready", plugin)
			}

		}
	}
	return true
}

// GetClusterResult xxx
func (pm *PluginManager) GetClusterResult(pluginStr string, clusterID string) map[string]CheckResult {
	Pm.Ready(pluginStr, clusterID)
	result := make(map[string]CheckResult)
	for _, plugin := range strings.Split(pluginStr, ",") {
		p := pm.GetPlugin(plugin)
		result[plugin] = p.GetResult(clusterID)
	}
	return result
}

// GetNodeResult xxx
func (pm *PluginManager) GetNodeResult(pluginStr string) map[string]CheckResult {
	result := make(map[string]CheckResult)
	for _, plugin := range strings.Split(pluginStr, ",") {
		p := pm.GetPlugin(plugin)
		if p == nil {
			continue
		}
		result[plugin] = p.GetResult("")
	}
	return result
}

// GetNodeDetail XXX
func (pm *PluginManager) GetNodeDetail(pluginStr string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, plugin := range strings.Split(pluginStr, ",") {
		p := pm.GetPlugin(plugin)
		if p == nil {
			continue
		}
		result[plugin] = p.GetDetail()
	}
	return result
}
