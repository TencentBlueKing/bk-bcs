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

// Package timecheck xxx
package timecheck

import (
	"fmt"
	"strings"
	"time"

	"github.com/beevik/ntp"
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

type Detail struct {
}

var (
	ntpAvailabilityLabels = []string{"node", "status"}
	ntpAvailability       = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ntp_availability",
		Help: "ntp_availability, 1 means OK",
	}, ntpAvailabilityLabels)
)

func init() {
	metric_manager.Register(ntpAvailability)
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

	nodeconfig := plugin_manager.Pm.GetConfig().NodeConfig
	nodeName := nodeconfig.NodeName
	p.ready = false

	var gaugeVecSet *metric_manager.GaugeVecSet

	servers := strings.Split(p.opt.TimeServers, ",")
	offset, err := GetTimeOffset(servers[rand.Intn(len(servers)-1)])
	if err != nil {
		klog.Errorf("get time offset failed: %s", err.Error())
		gaugeVecSet = &metric_manager.GaugeVecSet{
			Labels: []string{nodeName, timeErrorStatus},
			Value:  0,
		}
		result = append(result, plugin_manager.CheckItem{
			ItemName:   pluginName,
			ItemTarget: nodeName,
			Normal:     false,
			Detail:     fmt.Sprintf("get time offset failed: %s", err.Error()),
			Status:     timeErrorStatus,
		})
	} else {
		klog.Infof("%s result is %.8fs", p.Name(), float64(offset/time.Second))

		if offset > 3*time.Second {
			result = append(result, plugin_manager.CheckItem{
				ItemName:   pluginName,
				ItemTarget: nodeName,
				Level:      plugin_manager.RISKLevel,
				Normal:     false,
				Detail:     fmt.Sprintf("%s offset is %v", nodeName, offset),
				Status:     timeOffsetErrorStatus,
			})

			gaugeVecSet = &metric_manager.GaugeVecSet{
				Labels: []string{nodeName, timeOffsetErrorStatus},
				Value:  float64(offset) / float64(time.Second),
			}
		}
	}

	if len(result) == 0 {
		gaugeVecSet = &metric_manager.GaugeVecSet{
			Labels: []string{nodeName, "ok"},
			Value:  float64(offset) / float64(time.Second),
		}

		result = append(result, plugin_manager.CheckItem{
			ItemName:   pluginName,
			ItemTarget: nodeName,
			Level:      plugin_manager.RISKLevel,
			Status:     plugin_manager.NormalStatus,
			Normal:     true,
			Detail:     fmt.Sprintf("%s offset is %v", nodeName, offset),
		})

	}
	metric_manager.RefreshMetric(ntpAvailability, []*metric_manager.GaugeVecSet{gaugeVecSet})
	p.Result = plugin_manager.CheckResult{
		Items: result,
	}

	if !p.ready {
		p.ready = true
	}

}

func GetTimeOffset(timeserver string) (time.Duration, error) {
	localTime := time.Now()

	ntpServer := timeserver

	ntpTime, err := ntp.Time(ntpServer)
	if err != nil {
		return 0, err
	}

	diff := ntpTime.Sub(localTime)

	return diff, nil
}

func (p *Plugin) Ready(string) bool {
	return p.ready
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

func (p *Plugin) GetString(key string) string {
	return StringMap[key]
}
