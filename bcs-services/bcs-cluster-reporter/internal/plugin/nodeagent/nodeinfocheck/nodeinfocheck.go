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
	"os"
	"path"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/api/qcloud"
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
	nodeMetadataLabel  = []string{"node", "item", "value"}
	nodeMetadataMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_metadata",
		Help: "node_metadata, 1 means OK",
	}, nodeMetadataLabel)
)

func init() {
	metric_manager.Register(nodeMetadataMetric)
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
	result := plugin_manager.CheckResult{
		Items:        make([]plugin_manager.CheckItem, 0, 0),
		InfoItemList: make([]plugin_manager.InfoItem, 0, 0),
	}
	p.CheckLock.Lock()
	klog.Infof("start %s", p.Name())
	defer func() {
		klog.Infof("end %s", p.Name())
		p.CheckLock.Unlock()
	}()

	p.ready = false

	nodeconfig := plugin_manager.Pm.GetConfig().NodeConfig
	nodeName := nodeconfig.NodeName
	gvsList := make([]*metric_manager.GaugeVecSet, 0, 0)

	// qcloudinfo
	_, err := os.Stat(path.Join(nodeconfig.HostPath, "/etc/cloud/cloud.cfg"))
	if err != nil {
		klog.Error("now only support qcloud cvm get nodeinfo, skip, %s", err.Error())
	} else {
		nodeMetadata, err := qcloud.GetQcloudNodeMetadata()
		if err != nil {
			klog.Errorf("get cvm info failed: %s", err.Error())
			if !os.IsNotExist(err) {
				checkItem := plugin_manager.CheckItem{
					ItemName:   pluginName,
					ItemTarget: nodeName,
					Normal:     false,
					Detail:     fmt.Sprintf("get nodeinfo failed: %s", err.Error()),
					Status:     errorStatus,
				}
				gvsList = append(gvsList, &metric_manager.GaugeVecSet{
					Labels: []string{nodeName, ZoneItemType, errorStatus}, Value: float64(1),
				})
				result.Items = append(result.Items, checkItem)
			}
		} else {
			klog.Infof("node metadata is %v", *nodeMetadata)
			gvsList = append(gvsList, &metric_manager.GaugeVecSet{
				Labels: []string{nodeName, ZoneItemType, nodeMetadata.Zone}, Value: float64(1),
			})
			result.InfoItemList = append(result.InfoItemList, plugin_manager.InfoItem{
				ItemName: ZoneItemType,
				Labels:   map[string]string{"type": ZoneItemType},
				Result:   nodeMetadata.Zone,
			})

			gvsList = append(gvsList, &metric_manager.GaugeVecSet{
				Labels: []string{nodeName, RegionItemType, nodeMetadata.Region}, Value: float64(1),
			})
			result.InfoItemList = append(result.InfoItemList, plugin_manager.InfoItem{
				ItemName: RegionItemType,
				Labels:   map[string]string{"type": RegionItemType},
				Result:   nodeMetadata.Region,
			})

			gvsList = append(gvsList, &metric_manager.GaugeVecSet{
				Labels: []string{nodeName, InstanceTypeItemType, nodeMetadata.InstanceType}, Value: float64(1),
			})
			result.InfoItemList = append(result.InfoItemList, plugin_manager.InfoItem{
				ItemName: InstanceTypeItemType,
				Labels:   map[string]string{"type": InstanceTypeItemType},
				Result:   nodeMetadata.InstanceType,
			})
		}
	}

	metric_manager.RefreshMetric(nodeMetadataMetric, gvsList)
	p.Result = result

	if !p.ready {
		p.ready = true
	}
}

// GetResult return check result by cluster ID
func (p *Plugin) GetResult(string) plugin_manager.CheckResult {
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
