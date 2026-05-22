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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metricmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"

	"github.com/beevik/ntp"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

// Plugin xxx
type Plugin struct {
	opt   *Options
	ready bool
	pluginmanager.NodePlugin
	Detail Detail
}

// Detail xxx
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
	metricmanager.Register(ntpAvailability)
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
	p.ready = false

	var gaugeVecSet *metricmanager.GaugeVecSet

	servers := strings.Split(p.opt.TimeServers, ",")
	offset, err := GetTimeOffset(servers[rand.Intn(len(servers)-1)])
	if err != nil {
		klog.Errorf("get time offset failed: %s", err.Error())
		gaugeVecSet = &metricmanager.GaugeVecSet{
			Labels: []string{nodeName, timeErrorStatus},
			Value:  0,
		}
		result = append(result, pluginmanager.CheckItem{
			ItemName:   pluginName,
			ItemTarget: nodeName,
			Normal:     false,
			Detail:     fmt.Sprintf("get time offset failed: %s", err.Error()),
			Status:     timeErrorStatus,
		})
	} else {
		klog.Infof("%s result is %.8fs", p.Name(), float64(offset/time.Second))

		if offset > 3*time.Second {
			result = append(result, pluginmanager.CheckItem{
				ItemName:   pluginName,
				ItemTarget: nodeName,
				Level:      pluginmanager.RISKLevel,
				Normal:     false,
				Detail:     fmt.Sprintf("%s offset is %v", nodeName, offset),
				Status:     timeOffsetErrorStatus,
			})

			gaugeVecSet = &metricmanager.GaugeVecSet{
				Labels: []string{nodeName, timeOffsetErrorStatus},
				Value:  float64(offset) / float64(time.Second),
			}
		}
	}

	if len(result) == 0 {
		gaugeVecSet = &metricmanager.GaugeVecSet{
			Labels: []string{nodeName, "ok"},
			Value:  float64(offset) / float64(time.Second),
		}

		result = append(result, pluginmanager.CheckItem{
			ItemName:   pluginName,
			ItemTarget: nodeName,
			Level:      pluginmanager.RISKLevel,
			Status:     pluginmanager.NormalStatus,
			Normal:     true,
			Detail:     fmt.Sprintf("%s offset is %v", nodeName, offset),
		})

	}
	metricmanager.RefreshMetric(ntpAvailability, []*metricmanager.GaugeVecSet{gaugeVecSet})
	p.Result = pluginmanager.CheckResult{
		Items: result,
	}

	if !p.ready {
		p.ready = true
	}

}

// GetTimeOffset xxx
func GetTimeOffset(timeserver string) (time.Duration, error) {
	// 使用 NTP 算法的时钟偏移而不是直接相减，以抵消 RTT 的影响
	opts := ntp.QueryOptions{
		Timeout: 3 * time.Second,
		TTL:     0, // 使用系统默认 TTL；若网络要求可按需设置
		// LocalAddress: "", // 如需绑定特定本地地址可设置
	}

	resp, err := ntp.QueryWithOptions(timeserver, opts)
	if err != nil {
		return 0, err
	}

	// 基本校验：响应有效且未被标记为未同步
	if err := resp.Validate(); err != nil {
		return 0, fmt.Errorf("invalid ntp response: %w", err)
	}

	// 可选质量门槛（根据环境调参）：
	// - 过大的根距离（root distance, 近似上界误差）
	// - 过高的 Stratum
	// - 过大的 RTT
	const (
		maxRootDistance = 250 * time.Millisecond
		maxRTT          = 500 * time.Millisecond
		maxStratum      = 15 // 16 表示未同步
	)
	if resp.RootDistance > maxRootDistance {
		klog.Errorf("root distance too large: %v", resp.RootDistance)
	}
	if resp.RTT > maxRTT {
		klog.Errorf("rtt too large: %v", resp.RTT)
	}
	if resp.Stratum > maxStratum {
		klog.Errorf("stratum too high: %d", resp.Stratum)
	}

	// 使用 NTP 算法计算的时钟偏移（已考虑网络延迟）
	return resp.ClockOffset, nil
}

// Ready xxx
func (p *Plugin) Ready(string) bool {
	return p.ready
}

// GetResult xxx
func (p *Plugin) GetResult(string) pluginmanager.CheckResult {
	return p.Result
}

// Execute xxx
func (p *Plugin) Execute() {
	p.Check(pluginmanager.CheckOption{})
}

// GetDetail xxx
func (p *Plugin) GetDetail() interface{} {
	return p.Detail
}
