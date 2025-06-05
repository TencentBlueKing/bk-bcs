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

// Package processcheck xxx
package processcheck

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metricmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/types/process"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

var (
	processStatusLabels = []string{"name", "status", "node"}
	processStatus       = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "process_status",
		Help: "process_status",
	}, processStatusLabels)
	processCPU = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "agent_process_cpu_seconds_total",
		Help: "agent_process_cpu_seconds_total, 1 means OK",
	}, []string{"name", "node"})
	abnormalProcessStatusMap = make(map[int32]process.ProcessStatus)

	processCPUCheckFlag    = false
	processCPUCheckMap     = make(map[int32]int32)
	processCPUCheckMapLock sync.Mutex
)

func init() {
	metricmanager.Register(processStatus)
	metricmanager.Register(processCPU)
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
	ProcessInfo   []process.ProcessInfo
	ProcessStatus []process.ProcessStatus
}

// ProcessCheckResult xxx
type ProcessCheckResult struct {
	Node   string `yaml:"node"`
	Status string `yaml:"status"`
	Name   string `yaml:"name"`
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
					go p.Check(pluginmanager.CheckOption{})
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
		p.Check(pluginmanager.CheckOption{})
	}

	return nil
}

// Stop xxx
func (p *Plugin) Stop() error {
	p.StopChan <- 1
	processCPUCheckFlag = false
	klog.Infof("plugin %s stopped", p.Name())
	return nil
}

// Name xxx
func (p *Plugin) Name() string {
	return pluginName
}

// Check xxx
func (p *Plugin) Check(option pluginmanager.CheckOption) {
	result := make([]pluginmanager.CheckItem, 0, 0)
	p.CheckLock.Lock()
	klog.Infof("start %s", p.Name())
	defer func() {
		klog.Infof("end %s", p.Name())
		p.CheckLock.Unlock()
	}()

	nodeconfig := pluginmanager.Pm.GetConfig().NodeConfig
	nodeName := nodeconfig.NodeName
	config := pluginmanager.Pm.GetConfig()

	if strings.Contains(nodeconfig.Node.Status.NodeInfo.ContainerRuntimeVersion, "containerd") {
		p.opt.Processes = append(p.opt.Processes, ProcessCheckConfig{Name: "containerd", ConfigFile: "/etc/containerd/config.toml"})
	} else {
		p.opt.Processes = append(p.opt.Processes, ProcessCheckConfig{Name: "dockerd", ConfigFile: "/etc/docker/daemon.json"})
	}
	p.opt.Processes = removeDuplicates(p.opt.Processes)

	processInfoList := make([]process.ProcessInfo, 0, 0)
	processGaugeVecSetList := make([]*metricmanager.GaugeVecSet, 0, 0)

	processStatusList, err := process.GetProcessStatus()
	if err != nil {
		klog.Errorf("Get process status failed: %s", err.Error())
	}

	// 检测所有进程状态
	newAbnormalProcessStatusMap := make(map[int32]process.ProcessStatus)
	abnormalProcessStatusList := make([]process.ProcessStatus, 0, 0)
	processCPUGVSList := make([]*metricmanager.GaugeVecSet, 0, 0)

	// 找到状态异常的进程
	for _, pstatus := range processStatusList {
		if pstatus.Status == "D" || pstatus.Status == "Z" {
			klog.Infof("status of process %d %s is %s", pstatus.Pid, pstatus.Name, pstatus.Status)
			// 避免如正常的等待IO被计入D状态进程
			newAbnormalProcessStatusMap[pstatus.Pid] = pstatus
			if abnormalProcessStatus, ok := abnormalProcessStatusMap[pstatus.Pid]; ok && pstatus.Status == "D" {
				if abnormalProcessStatus.Pid == pstatus.Pid && abnormalProcessStatus.CreateTime == pstatus.CreateTime && abnormalProcessStatus.CpuTime == pstatus.CpuTime {
					// cputime didn't increase, means process stayed in D status in this interval
					processGaugeVecSetList = append(processGaugeVecSetList, &metricmanager.GaugeVecSet{
						Labels: []string{pstatus.Name, pstatus.Status, nodeName}, Value: float64(1),
					})

					result = append(result, pluginmanager.CheckItem{
						ItemName:   pluginName,
						ItemTarget: nodeName,
						Normal:     false,
						Detail:     fmt.Sprintf("%s process %s status is %s", nodeName, pstatus.Name, pstatus.Status),
						Level:      pluginmanager.WARNLevel,
						Status:     dStatus,
					})
					abnormalProcessStatusList = append(abnormalProcessStatusList, pstatus)
				}
			} else if pstatus.Status == "Z" {
				processGaugeVecSetList = append(processGaugeVecSetList, &metricmanager.GaugeVecSet{
					Labels: []string{pstatus.Name, pstatus.Status, nodeName}, Value: float64(1),
				})
				result = append(result, pluginmanager.CheckItem{
					ItemName:   pluginName,
					ItemTarget: nodeName,
					Normal:     false,
					Detail:     fmt.Sprintf("%s process %s status is %s", nodeName, pstatus.Name, pstatus.Status),
					Level:      pluginmanager.WARNLevel,
					Status:     zStatus,
				})
				abnormalProcessStatusList = append(abnormalProcessStatusList, pstatus)
			}
		}
	}

	// 进程cpu采点记录
	RecordProcessCpu([]string{"kswapd"}, processStatusList)

	if processCPUCheckFlag {
		go func() {
			for {
				processCPUCheckMapLock.Lock()
				defer func() {
					processCPUCheckMapLock.Unlock()
				}()
				if !processCPUCheckFlag {
					return
				}

				for _, pid := range processCPUCheckMap {
					pStatus, err := process.GetProcessStatusByPID(pid)
					if err != nil {
						klog.Errorf(err.Error())
						continue
					}

					processCPUGVSList = append(processCPUGVSList, &metricmanager.GaugeVecSet{
						Labels: []string{pStatus.Name, nodeName},
						Value:  pStatus.CpuTime,
					})
				}

				metricmanager.RefreshMetric(processCPU, processCPUGVSList)
				time.Sleep(time.Second * 30)

			}
		}()
	}

	abnormalProcessStatusMap = newAbnormalProcessStatusMap

	checkItem := pluginmanager.CheckItem{
		ItemName: pluginName,
		Normal:   true,
		Detail:   "",
		Level:    pluginmanager.WARNLevel,
		Status:   NormalStatus,
	}
	// status中只记录异常的进程状态
	p.Detail.ProcessStatus = abnormalProcessStatusList

	result1, processGaugeVecSetList1, processInfoList1 := p.checkProcess(config)
	result = append(result, result1...)
	processGaugeVecSetList = append(processGaugeVecSetList, processGaugeVecSetList1...)
	processInfoList = append(processInfoList, processInfoList1...)

	if len(processGaugeVecSetList) == 0 {
		checkItem.ItemTarget = nodeName
		processGaugeVecSetList = append(processGaugeVecSetList, &metricmanager.GaugeVecSet{
			Labels: []string{"", NormalStatus, nodeName}, Value: float64(1),
		})

		result = append(result, checkItem)
	}

	// info中记录所有指定的进程信息
	p.Detail.ProcessInfo = processInfoList

	p.Result = pluginmanager.CheckResult{
		Items: result,
	}

	if !p.ready {
		p.ready = true
	}
	// return result
	metricmanager.RefreshMetric(processStatus, processGaugeVecSetList)
}

func (p *Plugin) checkProcess(config *pluginmanager.Config) ([]pluginmanager.CheckItem, []*metricmanager.GaugeVecSet, []process.ProcessInfo) {
	result := make([]pluginmanager.CheckItem, 0, 0)
	processGaugeVecSetList := make([]*metricmanager.GaugeVecSet, 0)
	processInfoList := make([]process.ProcessInfo, 0)

	checkItem := pluginmanager.CheckItem{
		ItemName: pluginName,
		Normal:   true,
		Detail:   "",
		Level:    pluginmanager.WARNLevel,
		Status:   NormalStatus,
	}

	// 检测opt中指定的进程，记录进程详细信息
	for _, pcc := range p.opt.Processes {
		checkItem.ItemTarget = config.NodeConfig.NodeName
		processInfo, processErr := process.GetProcessInfo(pcc.Name, 0)
		if processErr != nil {
			klog.Errorf("Get process %s info failed: %s", pcc.Name, processErr.Error())
			checkItem.Detail = fmt.Sprintf("Get process %s info failed: %s", pcc.Name, processErr.Error())
			checkItem.Normal = false
			checkItem.Status = processOtherErrorStatus

			result = append(result, checkItem)

			processGaugeVecSetList = append(processGaugeVecSetList, &metricmanager.GaugeVecSet{
				Labels: []string{pcc.Name, processOtherErrorStatus, config.NodeConfig.NodeName}, Value: float64(1),
			})

			continue
		}

		if pcc.Name == "kubelet" {
			kubeletParams := make(map[string]string)
			for _, param := range processInfo.Params {
				if strings.HasPrefix(param, "--") && strings.Contains(param, "=") {
					param = strings.TrimPrefix(param, "--")
					key := strings.SplitN(param, "=", 2)[0]
					value := strings.SplitN(param, "=", 2)[1]
					kubeletParams[key] = value
				} else {
					param = strings.TrimPrefix(param, "--")
					kubeletParams[param] = ""
				}
				config.NodeConfig.KubeletParams = kubeletParams
				pluginmanager.Pm.SetConfig(config)
			}
		}

		if pcc.ConfigFile != "" {
			configFile, configFileErr := getConfigFile(pcc)
			if configFileErr != nil {
				klog.Errorf(configFileErr.Error())

				checkItem.Normal = false
				checkItem.Detail = fmt.Sprintf("Get process %s info failed: %s", pcc.Name, configFileErr.Error())
				checkItem.ItemTarget = config.NodeConfig.NodeName
				checkItem.Status = processOtherErrorStatus

				result = append(result, checkItem)

				processGaugeVecSetList = append(processGaugeVecSetList, &metricmanager.GaugeVecSet{
					Labels: []string{pcc.Name, processOtherErrorStatus, config.NodeConfig.NodeName}, Value: float64(1),
				})
			} else if pcc.ConfigFile != "" {
				processInfo.ConfigFiles[pcc.ConfigFile] = configFile
			}
		}

		if processInfo != nil {
			processInfoList = append(processInfoList, *processInfo)
		}
	}

	return result, processGaugeVecSetList, processInfoList
}

// RecordProcessCpu xxx
func RecordProcessCpu(processNameList []string, processStatusList []process.ProcessStatus) {
	for _, pstatus := range processStatusList {
		for _, name := range processNameList {
			if strings.Contains(pstatus.Name, name) {
				if _, ok := processCPUCheckMap[pstatus.Pid]; !ok {
					processCPUCheckMapLock.Lock()
					processCPUCheckMap[pstatus.Pid] = pstatus.Pid
					processCPUCheckMapLock.Unlock()
				}
			}
		}
	}
}

// GetConfigfile xxx
func getConfigFile(p ProcessCheckConfig) (string, error) {
	if p.ConfigFile != "" {
		data, err := os.ReadFile(path.Join(os.Getenv("HOST_PATH"), p.ConfigFile))
		if err != nil {
			return "", err
		}
		return string(data), nil
	} else {
		return "", nil
	}
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
	p.Check(pluginmanager.CheckOption{})
}

// GetString xxx
func (p *Plugin) GetString(key string) string {
	return StringMap[key]
}
