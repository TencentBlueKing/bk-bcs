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

// Package hwcheck xxx
package hwcheck

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metricmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"
	"os"
	"path"
	"strings"
	"time"

	"github.com/jaypipes/ghw"
	"github.com/jaypipes/ghw/pkg/option"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

var (
	deviceStatusLabels  = []string{"id", "name", "node", "revision"}
	hardwareErrorLabels = []string{"type", "node"}
	deviceStatus        = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "device_status",
		Help: "device_status",
	}, deviceStatusLabels)

	hardwareError = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hardware_error_count",
		Help: "hardware_error_count",
	}, hardwareErrorLabels)
)

func init() {
	metricmanager.Register(deviceStatus)
	metricmanager.Register(hardwareError)
}

// Plugin XXX
type Plugin struct {
	opt    *Options
	ready  bool
	Detail Detail
	pluginmanager.NodePlugin
}

// Detail XXX
type Detail struct {
}

// Setup xxx
func (p *Plugin) Setup(configFilePath string, runMode string) error {
	p.opt = &Options{}
	err := util.ReadorInitConf(configFilePath, p.opt, initContent)
	//err := util.ReadFromStr(p.opt, initContent)
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

	node := pluginmanager.Pm.GetConfig().NodeConfig

	logFileConfigList := make([]LogFileConfig, 0, 0)
	for _, logFileConfig := range p.opt.LogFileConfigList {
		logFileConfig.logFile = util.NewLogFile(path.Join(node.HostPath, logFileConfig.Path))
		if logFileConfig.logFile == nil {
			klog.Errorf("%s no such file or directory, skip", logFileConfig.Path)
			continue
		}

		logFileConfig.logFile.SetSearchKey(logFileConfig.KeyWordList)

		logFileConfigList = append(logFileConfigList, logFileConfig)
	}

	p.opt.LogFileConfigList = logFileConfigList

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

// Check check for hardware problem
func (p *Plugin) Check() {
	result := make([]pluginmanager.CheckItem, 0, 0)
	p.CheckLock.Lock()
	klog.Infof("start %s", p.Name())
	defer func() {
		klog.Infof("end %s", p.Name())
		p.CheckLock.Unlock()
	}()

	node := pluginmanager.Pm.GetConfig().NodeConfig
	nodeName := node.NodeName
	p.ready = false

	deviceList, err := GetDeviceStatus(node.HostPath)
	if err != nil {
		klog.Errorf(err.Error())
		return
	}

	deviceStatusGaugeVecSetList := make([]*metricmanager.GaugeVecSet, 0, 0)
	for _, device := range deviceList {
		deviceStatusGaugeVecSetList = append(deviceStatusGaugeVecSetList, &metricmanager.GaugeVecSet{
			Labels: []string{device.Address, strings.Replace(device.Vendor.Name, " ", "_", -1), nodeName, device.Revision},
			Value:  float64(1),
		})
		if device.Revision == "ff" {
			result = append(result, pluginmanager.CheckItem{
				ItemName:   pluginName,
				ItemTarget: nodeName,
				Normal:     false,
				Detail:     fmt.Sprintf("device %s %s revision is %s", device.Vendor.Name, device.Address, device.Revision),
				Status:     ffStatus,
			})
		}

	}

	metricmanager.SetMetric(deviceStatus, deviceStatusGaugeVecSetList)

	hardwareErrorGVSList := make([]*metricmanager.GaugeVecSet, 0, 0)
	for _, logFileConfig := range p.opt.LogFileConfigList {
		logList, err := logFileConfig.logFile.CheckNewEntriesOnce()
		if err != nil {
			klog.Errorf(err.Error())
		} else {
			for _, key := range logFileConfig.KeyWordList {
				count := 0
				for _, line := range logList {
					if strings.Contains(line, key) {
						count++
						//hardwareError.WithLabelValues(logFileConfig.Rule, nodeName).Add(1)
						//break
					}
				}

				hardwareErrorGVSList = append(hardwareErrorGVSList, &metricmanager.GaugeVecSet{
					Labels: []string{logFileConfig.Rule, nodeName},
					Value:  float64(count),
				})

				if count > 0 {
					result = append(result, pluginmanager.CheckItem{
						ItemName:   pluginName,
						ItemTarget: nodeName,
						Normal:     false,
						Detail:     fmt.Sprintf("%s found %s in logfile %s", logFileConfig.Rule, key, logFileConfig.Path),
						Status:     logMatchedStatus,
					})
				}
			}
		}
	}
	metricmanager.RefreshMetric(hardwareError, hardwareErrorGVSList)

	p.Result = pluginmanager.CheckResult{
		Items: result,
	}

	if !p.ready {
		p.ready = true
	}
}

// GetDeviceStatus xxx
func GetDeviceStatus(hostPath string) ([]*ghw.PCIDevice, error) {
	pciInfo, err := ghw.PCI(&option.Option{
		Chroot: &hostPath,
	})
	if err != nil {
		return nil, err
	}

	deviceList := make([]*ghw.PCIDevice, 0, 0)
	for _, device := range pciInfo.Devices {
		file, openErr := os.Open(fmt.Sprintf("/sys/bus/pci/devices/%s/config", device.Address))
		if openErr != nil {
			klog.Errorf("Error opening file: %s", openErr.Error())
			continue
		}
		defer file.Close()

		revision := make([]byte, 1)
		_, err = file.ReadAt(revision, 8) // Revision ID is at offset 8
		if err != nil {
			klog.Errorf("Error reading file: %s", err.Error())
			continue
		}

		if fmt.Sprintf("%x", revision[0]) == "ff" {
			device.Revision = fmt.Sprintf("%x", revision[0])
			deviceList = append(deviceList, device)
		}

		klog.Infof("%s %s %s %s status is %x", device.Address, device.Vendor.Name, device.Class.Name, device.Product.Name, revision[0])
	}

	return deviceList, nil
}

// GetResult xxx
func (p *Plugin) GetResult(string) pluginmanager.CheckResult {
	return p.Result
}

// Execute xxx
func (p *Plugin) Execute() {
	p.Check()
}

// GetDetail xxx
func (p *Plugin) GetDetail() interface{} {
	return p.Detail
}

// Ready xxx
func (p *Plugin) Ready(string) bool {
	return p.ready
}
