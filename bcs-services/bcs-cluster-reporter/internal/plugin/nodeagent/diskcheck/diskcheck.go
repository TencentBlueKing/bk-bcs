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

// Package diskcheck xxx
package diskcheck

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/moby/sys/mountinfo"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metric_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

// Plugin xxx
type Plugin struct {
	opt   *Options
	ready bool
	plugin_manager.NodePlugin
	Detail Detail
}

// Detail xxx
type Detail struct {
}

var (
	fsAvailabilityLabels = []string{"mountpoint", "node", "status"}
	fsAvailability       = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fs_availability",
		Help: "fs_availability, 1 means OK",
	}, fsAvailabilityLabels)
)

func init() {
	metric_manager.Register(fsAvailability)
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

	p.StopChan = make(chan int)
	interval := p.opt.Interval
	if interval == 0 {
		interval = 60
	}

	if err != nil {
		klog.Fatalf("%s get incluster config failed, only can run as incluster mode", p.Name())
	}

	// run as daemon
	if runMode == plugin_manager.RunModeDaemon {
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
	} else if runMode == plugin_manager.RunModeOnce {
		p.Check()
	}

	return nil
}

// Stop xxx
func (p *Plugin) Stop() error {
	p.CheckLock.Lock()
	p.StopChan <- 1
	klog.Infof("plugin %s stopped", p.Name())
	p.CheckLock.Unlock()
	return nil
}

// Name xxx
func (p *Plugin) Name() string {
	return pluginName
}

// Check xxx
func (p *Plugin) Check() {
	result := make([]plugin_manager.CheckItem, 0, 0)
	p.CheckLock.Lock()
	klog.Infof("start %s", p.Name())
	defer func() {
		klog.Infof("end %s", p.Name())
		p.CheckLock.Unlock()
	}()

	p.ready = false

	node := plugin_manager.Pm.GetConfig().NodeConfig
	nodeName := node.NodeName
	fsGaugeVecSetList := make([]*metric_manager.GaugeVecSet, 0, 0)

	mountInfoList, err := GetFSMountInfo(node.HostPath)
	if err != nil {
		klog.Errorf(err.Error())
		return
	}

	for _, mountInfo := range mountInfoList {
		err = TestFS(node.HostPath, mountInfo.Mountpoint)
		if err != nil {
			klog.Infof("test fs %s failed: %s", mountInfo.Mountpoint, err.Error())
			fsGaugeVecSetList = append(fsGaugeVecSetList, &metric_manager.GaugeVecSet{
				Labels: []string{mountInfo.Mountpoint, nodeName, "notok"}, Value: float64(1),
			})
			result = append(result, plugin_manager.CheckItem{
				ItemName:   pluginName,
				ItemTarget: nodeName,
				Normal:     false,
				Detail:     fmt.Sprintf("testfs %s failed: %s", mountInfo.Mountpoint, err.Error()),
				Status:     testFailStatus,
				Level:      plugin_manager.WARNLevel,
			})

		} else {
			klog.Infof("test fs %s success", mountInfo.Mountpoint)
		}
	}

	if len(fsGaugeVecSetList) == 0 {
		fsGaugeVecSetList = append(fsGaugeVecSetList, &metric_manager.GaugeVecSet{
			Labels: []string{"/", nodeName, NormalStatus}, Value: float64(1),
		})

		result = append(result, plugin_manager.CheckItem{
			ItemName:   pluginName,
			ItemTarget: nodeName,
			Normal:     true,
			Detail:     "",
			Status:     NormalStatus,
			Level:      plugin_manager.WARNLevel,
		})
	}

	metric_manager.RefreshMetric(fsAvailability, fsGaugeVecSetList)
	p.Result = plugin_manager.CheckResult{
		Items: result,
	}

	if !p.ready {
		p.ready = true
	}
}

func GetFSMountInfo(hostPath string) ([]*mountinfo.Info, error) {
	mountInfoList, err := GetProcessMountInfo(hostPath, 1)
	if err != nil {
		return nil, err
	}

	fsMountInfoList := make([]*mountinfo.Info, 0, 0)
	for _, mountInfo := range mountInfoList {
		if mountInfo.FSType == "ext4" || mountInfo.FSType == "xfs" {
			if mountInfo.Root == "/" {
				fsMountInfoList = append(fsMountInfoList, mountInfo)
			}
		}
	}
	return fsMountInfoList, err
}

func TestFS(hostPath, path string) error {
	rand.Seed(time.Now().UnixNano())
	fileName := filepath.Join(hostPath, path, fmt.Sprintf("%d", rand.Intn(1000))+".nodeagent")
	file, err := os.Create(fileName)
	defer func() {
		if file != nil {
			err = file.Close()
			if err != nil {
				klog.Error(err.Error())
			}
		}

		err = os.Remove(fileName)
		if err != nil {
			klog.Error(err.Error())
		}
	}()

	if err != nil {
		return err
	}

	_, err = file.WriteString("test")
	if err != nil {
		return err
	}

	return nil
}

func GetProcessMountInfo(hostPath string, pid int32) ([]*mountinfo.Info, error) {
	f, err := os.Open(fmt.Sprintf("%s/proc/%d/mountinfo", hostPath, pid))
	if err != nil {
		return nil, err
	}

	mountInfoList, err := mountinfo.GetMountsFromReader(f, nil)
	if err != nil {
		return nil, err
	}

	return mountInfoList, nil
}

func (p *Plugin) GetResult(string) plugin_manager.CheckResult {
	return p.Result
}

func (p *Plugin) Execute() {
	p.Check()
}

func (p *Plugin) GetDetail() interface{} {
	return p.Detail
}

func (p *Plugin) Ready(string) bool {
	return p.ready
}

func (p *Plugin) GetString(key string) string {
	return StringMap[key]
}
