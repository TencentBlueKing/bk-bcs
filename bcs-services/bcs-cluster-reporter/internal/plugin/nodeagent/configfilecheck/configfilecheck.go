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

// Package configfilecheck xxx
package configfilecheck

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metricmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"path"
	"time"

	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

var (
	checkRuleStatusLabels = []string{"name", "status", "node"}
	checkRuleStatus       = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "check_rule_status",
		Help: "check_rule_status",
	}, checkRuleStatusLabels)
)

func init() {
	metricmanager.Register(checkRuleStatus)
}

// Plugin xxx
type Plugin struct {
	opt   *Options
	ready bool

	// detail用来记录详细的检查信息到configmap，提供给cluster-reporter做进一步分析
	Detail Detail
	pluginmanager.NodePlugin
}

// Detail xxx
type Detail struct {
	ConfigFileMap map[string]string
}

// Setup xxx
func (p *Plugin) Setup(configFilePath string, runMode string) error {
	p.opt = &Options{}

	err := util.ReadorInitConf(configFilePath, p.opt, initContent)
	if err != nil {
		return err
	}

	if err = p.opt.Validate(); err != nil {
		return err
	}

	interval := p.opt.Interval
	if interval == 0 {
		interval = 60
	}

	p.Detail = Detail{}

	// run as daemon
	if runMode == pluginmanager.RunModeDaemon {
		go func() {
			for {
				if p.CheckLock.TryLock() {
					p.CheckLock.Unlock()
					go p.Check()
				} else {
					klog.Infof("the former %s didn't over, skip in this loop", p.Name())
				}
				select {
				case result := <-p.StopChan:
					klog.Infof("stop plugin %s by signal %d", p.Name(), result)
					return
				case <-time.After(time.Duration(interval) * time.Second):
					continue
				}
			}
		}()
	} else if runMode == pluginmanager.RunModeOnce {
		p.Check()
	}

	return nil
}

// Stop xxx
func (p *Plugin) Stop() error {
	p.StopChan <- 1
	klog.Infof("plugin %s stopped", p.Name())
	return nil
}

// Name xxx
func (p *Plugin) Name() string {
	return pluginName
}

// Check xxx
func (p *Plugin) Check() {
	p.CheckLock.Lock()
	klog.Infof("start %s", p.Name())
	defer func() {
		klog.Infof("end %s", p.Name())
		p.CheckLock.Unlock()
		p.ready = true
	}()

	p.ready = false

	nodeconfig := pluginmanager.Pm.GetConfig().NodeConfig
	nodeName := nodeconfig.NodeName

	configFileMap := make(map[string]string, 0)
	for _, filePath := range p.opt.FilePaths {
		filePath = path.Join(util.GetHostPath(), filePath)

		// 读取文件内容
		content, err := os.ReadFile(filePath)
		if err != nil {
			klog.Errorf("read file %s failed: %s", filePath, err.Error())
			configFileMap[filePath] = fmt.Sprintf("read file %s failed: %s", filePath, err.Error())
			continue
		}

		configFileMap[filePath] = string(content)
	}

	result := make([]pluginmanager.CheckItem, 0, 0)
	configFileGaugeVecSetList := make([]*metricmanager.GaugeVecSet, 0, 0)
	for _, checkRule := range p.opt.CheckRules {
		status, err := checkRule.Check()
		checkItem := pluginmanager.CheckItem{
			ItemName:   pluginName,
			ItemTarget: nodeName,
			Normal:     status == NormalStatus,
			Level:      pluginmanager.WARNLevel,
			Status:     status,
		}
		if err != nil {
			klog.Errorf("check rule failed: %s", err.Error())
			checkItem.Detail = fmt.Sprintf("check rule failed: %s", err.Error())
			configFileGaugeVecSetList = append(configFileGaugeVecSetList, &metricmanager.GaugeVecSet{
				Labels: []string{checkRule.RuleName, status, nodeName}, Value: float64(1),
			})
		}

		result = append(result, checkItem)
	}

	p.Result = pluginmanager.CheckResult{
		Items: result,
	}

	p.Detail.ConfigFileMap = configFileMap

	// return result
	metricmanager.RefreshMetric(checkRuleStatus, configFileGaugeVecSetList)
}

// Ready xxx
func (p *Plugin) Ready(string) bool {
	return p.ready
}

// GetResult xxx
func (p *Plugin) GetResult(string) pluginmanager.CheckResult {
	return p.Result
}

// GetDetail xxx
func (p *Plugin) GetDetail() interface{} {
	return p.Detail
}

// Execute xxx
func (p *Plugin) Execute() {
	p.Check()
}

// GetString xxx
func (p *Plugin) GetString(key string) string {
	return StringMap[key]
}
