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

// Package nodeinfocheck xxx
package nodeinfocheck

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metricmanager"
	pluginmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"
	"os"
	"path"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/api/qcloud"
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
	nodeMetadataLabel  = []string{"node", "item", "value"}
	nodeMetadataMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_metadata",
		Help: "node_metadata, 1 means OK",
	}, nodeMetadataLabel)
)

func init() {
	metricmanager.Register(nodeMetadataMetric)
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

// Check for node's platform info
func (p *Plugin) Check(option pluginmanager.CheckOption) {
	result := pluginmanager.CheckResult{
		Items:        make([]pluginmanager.CheckItem, 0, 0),
		InfoItemList: make([]pluginmanager.InfoItem, 0, 0),
	}
	p.CheckLock.Lock()
	klog.Infof("start %s", p.Name())
	defer func() {
		klog.Infof("end %s", p.Name())
		p.CheckLock.Unlock()
	}()

	p.ready = false

	nodeconfig := pluginmanager.Pm.GetConfig().NodeConfig
	nodeName := nodeconfig.NodeName
	gvsList := make([]*metricmanager.GaugeVecSet, 0, 0)

	// qcloudinfo
	_, err := os.Stat(path.Join(nodeconfig.HostPath, "/etc/cloud/cloud.cfg"))
	if err != nil {
		klog.Error("now only support qcloud cvm get nodeinfo, skip, %s", err.Error())
	} else {
		nodeMetadata, err := qcloud.GetQcloudNodeMetadata()
		if err != nil {
			klog.Errorf("get cvm info failed: %s", err.Error())
			if !os.IsNotExist(err) {
				checkItem := pluginmanager.CheckItem{
					ItemName:   pluginName,
					ItemTarget: nodeName,
					Normal:     false,
					Detail:     fmt.Sprintf("get nodeinfo failed: %s", err.Error()),
					Status:     errorStatus,
				}
				gvsList = append(gvsList, &metricmanager.GaugeVecSet{
					Labels: []string{nodeName, ZoneItemType, errorStatus}, Value: float64(1),
				})
				result.Items = append(result.Items, checkItem)
			}
		} else {
			klog.Infof("node metadata is %v", *nodeMetadata)
			gvsList = append(gvsList, &metricmanager.GaugeVecSet{
				Labels: []string{nodeName, ZoneItemType, nodeMetadata.Zone}, Value: float64(1),
			})
			result.InfoItemList = append(result.InfoItemList, pluginmanager.InfoItem{
				ItemName: ZoneItemType,
				Labels:   map[string]string{"type": ZoneItemType},
				Result:   nodeMetadata.Zone,
			})

			gvsList = append(gvsList, &metricmanager.GaugeVecSet{
				Labels: []string{nodeName, RegionItemType, nodeMetadata.Region}, Value: float64(1),
			})
			result.InfoItemList = append(result.InfoItemList, pluginmanager.InfoItem{
				ItemName: RegionItemType,
				Labels:   map[string]string{"type": RegionItemType},
				Result:   nodeMetadata.Region,
			})

			gvsList = append(gvsList, &metricmanager.GaugeVecSet{
				Labels: []string{nodeName, InstanceTypeItemType, nodeMetadata.InstanceType}, Value: float64(1),
			})
			result.InfoItemList = append(result.InfoItemList, pluginmanager.InfoItem{
				ItemName: InstanceTypeItemType,
				Labels:   map[string]string{"type": InstanceTypeItemType},
				Result:   nodeMetadata.InstanceType,
			})
		}
	}

	metricmanager.RefreshMetric(nodeMetadataMetric, gvsList)
	p.Result = result

	if !p.ready {
		p.ready = true
	}
}

// GetResult return check result by cluster ID
func (p *Plugin) GetResult(string) pluginmanager.CheckResult {
	return p.Result
}

// GetDetail xxx
func (p *Plugin) GetDetail() interface{} {
	return p.Detail
}

// Ready return true if cluster check is over
func (p *Plugin) Ready(string) bool {
	return p.ready
}
