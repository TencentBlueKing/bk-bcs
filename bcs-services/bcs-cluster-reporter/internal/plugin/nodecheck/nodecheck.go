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

// Package nodecheck xxx
package nodecheck

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/diskcheck"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/netcheck"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/nodeinfocheck"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/timecheck"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metric_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/dnscheck"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/hwcheck"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/processcheck"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

// Plugin xxx
type Plugin struct {
	opt            *Options
	testYamlString string
	plugin_manager.ClusterPlugin
}

var (
	nodeAvailabilityLabels = []string{"target", "bk_biz_id", "item", "item_target", "status"}
	nodeAvailability       = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cluster_node_availability",
		Help: "cluster_node_availability, 1 means OK",
	}, nodeAvailabilityLabels)
	nodeAvailabilityGaugeVecSetMap = make(map[string][]*metric_manager.GaugeVecSet)
)

func init() {
	metric_manager.Register(nodeAvailability)
}

// Setup xxx
func (p *Plugin) Setup(configFilePath string, runMode string) error {
	p.opt = &Options{}
	err := util.ReadorInitConf(configFilePath, p.opt, initContent)
	if err != nil {
		return fmt.Errorf("read clustercheck config file %s failed, err %s", configFilePath, err.Error())
	}

	if err = p.opt.Validate(); err != nil {
		return err
	}

	p.Result = make(map[string]plugin_manager.CheckResult)
	p.ReadyMap = make(map[string]bool)

	interval := p.opt.Interval
	if interval == 0 {
		interval = 60
	}

	if runMode == "daemon" {
		go func() {
			for {
				if p.CheckLock.TryLock() {
					p.CheckLock.Unlock()
					go p.Check()
				} else {
					klog.V(3).Infof("the former %s didn't over, skip in this loop", p.Name())
				}
				select {
				case result := <-p.StopChan:
					klog.V(3).Infof("stop plugin %s by signal %d", p.Name(), result)
					return
				case <-time.After(time.Duration(interval) * time.Second):
					continue
				}
			}
		}()
	} else if runMode == "once" {
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
	start := time.Now()
	p.CheckLock.Lock()
	klog.Infof("start %s", p.Name())
	defer func() {
		klog.Infof("end %s", p.Name())
		p.CheckLock.Unlock()
		metric_manager.SetCommonDurationMetric([]string{p.Name(), "", "", ""}, start)
	}()

	clusterConfigs := plugin_manager.Pm.GetConfig().ClusterConfigs

	wg := sync.WaitGroup{}

	// 遍历所有集群
	for _, cluster := range clusterConfigs {
		wg.Add(1)
		plugin_manager.Pm.Add()

		go func(cluster *plugin_manager.ClusterConfig) {
			cluster.Lock()
			klog.Infof("start nodecheck for %s", cluster.ClusterID)

			p.WriteLock.Lock()
			p.ReadyMap[cluster.ClusterID] = false
			p.WriteLock.Unlock()

			config := cluster.Config
			clusterId := cluster.ClusterID
			clusterbiz := cluster.BusinessID

			defer func() {
				cluster.Unlock()
				plugin_manager.Pm.Done()
				p.WriteLock.Lock()
				p.ReadyMap[cluster.ClusterID] = true
				p.WriteLock.Unlock()
				wg.Done()
				klog.Infof("end nodecheck for %s", cluster.ClusterID)
			}()
			clusterResult := plugin_manager.CheckResult{
				Items: make([]plugin_manager.CheckItem, 0, 0),
			}

			clientSet, _ := k8s.GetClientsetByConfig(config)
			cmList, err := clientSet.CoreV1().ConfigMaps(nodeagentNamespace).List(util.GetCtx(10*time.Second), v1.ListOptions{
				ResourceVersion: "0",
			})
			if err != nil {
				klog.Errorf("get nodeagent configmap from cluster %s failed: %s", clusterId, err.Error())
				return
			}

			nodeAvailabilityGVSMap := make(map[string][]*metric_manager.GaugeVecSet)
			//遍历该集群的nodeagent configmap
			klog.Infof("%s nodeagent configmap num: %d", clusterId, len(cmList.Items))
			for _, configmap := range cmList.Items {
				if !strings.HasSuffix(configmap.Name, "-v1") {
					continue
				}
				if _, ok := configmap.Data["nodeinfo"]; !ok {
					continue
				}

				// 检查更新时间
				if _, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", configmap.Data["updateTime"]); err == nil {
					//if time.Now().Sub(updateTime) > time.Hour*24 {
					//	continue
					//}
				} else {
					continue
				}
				nodeName := strings.TrimSuffix(configmap.Name, "-v1")

				nodeinfo := make(map[string]plugin_manager.PluginInfo)
				err = yaml.Unmarshal([]byte(configmap.Data["nodeinfo"]), nodeinfo)
				if err != nil {
					klog.Errorf("unmarshal %s nodeinfo %s failed: %s", clusterId, configmap.Name, err.Error())
					continue
				}

				// 获取节点的checkitem并生成metric的map
				nodeInfo := plugin.NodeInfo{}
				checkItemList, infoItemList, nodeGVSMap := checkNodePluginResult(nodeinfo, strings.TrimSuffix(configmap.Name, "-v1"), clusterId, clusterbiz, &nodeInfo)

				clusterResult.Items = append(clusterResult.Items, checkItemList...)
				clusterResult.InfoItemList = append(clusterResult.InfoItemList, infoItemList...)
				for key, nodeGVSList := range nodeGVSMap {
					if _, ok := nodeAvailabilityGVSMap[key]; !ok {
						nodeAvailabilityGVSMap[key] = make([]*metric_manager.GaugeVecSet, 0, 0)
					}
					nodeAvailabilityGVSMap[key] = append(nodeAvailabilityGVSMap[key], nodeGVSList...)
				}

				cluster.NodeInfo[nodeName] = nodeInfo
			}

			nodeAvailabilityGaugeVecSetList := make([]*metric_manager.GaugeVecSet, 0, 0)
			for key, gvsList := range nodeAvailabilityGVSMap {
				if len(gvsList) == 0 {
					nodeAvailabilityGaugeVecSetList = append(nodeAvailabilityGaugeVecSetList, &metric_manager.GaugeVecSet{
						Labels: []string{clusterId, clusterbiz, key, "node", normalStatus},
						Value:  1,
					})
				} else {
					// 如果大于1,如果有异常则只增加一条异常的指标，否则增加一条正常的指标
					nodeAvailabilityGaugeVecSetList = append(nodeAvailabilityGaugeVecSetList, gvsList[0])
					for _, gvs := range gvsList {
						if gvs.Labels[4] != normalStatus {
							nodeAvailabilityGaugeVecSetList[len(nodeAvailabilityGaugeVecSetList)-1] = gvs
							break
						}
					}
				}
			}

			p.WriteLock.Lock()
			metric_manager.DeleteMetric(nodeAvailability, nodeAvailabilityGaugeVecSetMap[clusterId])
			nodeAvailabilityGaugeVecSetMap[clusterId] = nodeAvailabilityGaugeVecSetList
			metric_manager.SetMetric(nodeAvailability, nodeAvailabilityGaugeVecSetMap[clusterId])
			p.Result[clusterId] = clusterResult
			p.WriteLock.Unlock()
		}(cluster)
	}

	wg.Wait()

	// clean deleted cluster data
	for clusterID, _ := range p.ReadyMap {
		if _, ok := clusterConfigs[clusterID]; !ok {
			p.ReadyMap[clusterID] = false
			klog.Infof("delete cluster %s", clusterID)
		}
	}

	for clusterID, ready := range p.ReadyMap {
		if !ready {
			delete(p.ReadyMap, clusterID)
			delete(nodeAvailabilityGaugeVecSetMap, clusterID)
			delete(p.Result, clusterID)
			metric_manager.DeleteMetric(nodeAvailability, nodeAvailabilityGaugeVecSetMap[clusterID])
		}
	}
}

func checkNodePluginResult(nodeinfo map[string]plugin_manager.PluginInfo, nodeName string, clusterId, clusterbiz string, nodeInfo *plugin.NodeInfo) ([]plugin_manager.CheckItem, []plugin_manager.InfoItem, map[string][]*metric_manager.GaugeVecSet) {
	checkItemList := make([]plugin_manager.CheckItem, 0, 0)
	infoItemList := make([]plugin_manager.InfoItem, 0, 0)
	nodeGVSMap := make(map[string][]*metric_manager.GaugeVecSet)

	// 所有节点检测项，不管正常与否都应该返回对应checkitem
	for name, pluginInfo := range nodeinfo {
		for _, checkItem := range pluginInfo.Result.Items {
			if _, ok := nodeGVSMap[checkItem.ItemName]; !ok {
				nodeGVSMap[checkItem.ItemName] = make([]*metric_manager.GaugeVecSet, 0, 0)
			}

			nodeGVSMap[checkItem.ItemName] = append(nodeGVSMap[checkItem.ItemName], &metric_manager.GaugeVecSet{
				Labels: []string{clusterId, clusterbiz, checkItem.ItemName, "node", checkItem.Status},
				Value:  1,
			})
		}

		switch name {
		case "processcheck":
			for _, checkItem := range pluginInfo.Result.Items {
				checkItem.ItemName = processcheck.StringMap[checkItem.ItemName]
				checkItem.Status = processcheck.StringMap[checkItem.Status]
				checkItemList = append(checkItemList, checkItem)
			}
			detailBytes, err := yaml.Marshal(pluginInfo.Detail)
			if err != nil {
				klog.Errorf(err.Error())
				continue
			}

			detail := processcheck.Detail{}
			err = yaml.Unmarshal(detailBytes, &detail)
			if err != nil {
				klog.Errorf(err.Error())
				continue
			}

			// 检查进程配置，生成checkitem
			processResult := checkProcess(detail, nodeName)
			checkItemList = append(checkItemList, processResult...)

			for index, checkItem := range processResult {
				if _, ok := nodeGVSMap[checkItem.ItemName]; !ok {
					nodeGVSMap[checkItem.ItemName] = make([]*metric_manager.GaugeVecSet, 0, 0)

				}
				nodeGVSMap[checkItem.ItemName] = append(nodeGVSMap[checkItem.ItemName], &metric_manager.GaugeVecSet{
					Labels: []string{clusterId, clusterbiz, checkItem.ItemName, "node", checkItem.Status},
					Value:  1,
				})

				checkItem.ItemName = StringMap[checkItem.ItemName]
				checkItem.Status = StringMap[checkItem.Status]
				processResult[index] = checkItem
			}

		case "dnscheck":
			for _, checkItem := range pluginInfo.Result.Items {
				checkItem.ItemName = dnscheck.StringMap[checkItem.ItemName]
				checkItem.Status = dnscheck.StringMap[checkItem.Status]
				checkItemList = append(checkItemList, checkItem)
			}
		case "timecheck":
			for _, checkItem := range pluginInfo.Result.Items {
				checkItem.ItemName = timecheck.StringMap[checkItem.ItemName]
				checkItem.Status = timecheck.StringMap[checkItem.Status]
				checkItemList = append(checkItemList, checkItem)
			}
		case "netcheck":
			for _, checkItem := range pluginInfo.Result.Items {
				checkItem.ItemName = netcheck.StringMap[checkItem.ItemName]
				checkItem.Status = netcheck.StringMap[checkItem.Status]
				checkItemList = append(checkItemList, checkItem)
			}
		case "diskcheck":
			for _, checkItem := range pluginInfo.Result.Items {
				checkItem.ItemName = diskcheck.StringMap[checkItem.ItemName]
				checkItem.Status = diskcheck.StringMap[checkItem.Status]
				checkItemList = append(checkItemList, checkItem)
			}
		case "hwcheck":
			for _, checkItem := range pluginInfo.Result.Items {
				checkItem.ItemName = hwcheck.StringMap[checkItem.ItemName]
				checkItem.Status = hwcheck.StringMap[checkItem.Status]
				checkItemList = append(checkItemList, checkItem)
			}
		case "nodeinfocheck":
			for _, checkItem := range pluginInfo.Result.Items {
				checkItem.ItemName = nodeinfocheck.StringMap[checkItem.ItemName]
				checkItem.Status = nodeinfocheck.StringMap[checkItem.Status]
				checkItemList = append(checkItemList, checkItem)
			}

			for _, infoItem := range pluginInfo.Result.InfoItemList {

				if infoItem.ItemName == nodeinfocheck.ZoneItemType {
					nodeInfo.Zone = infoItem.Result.(string)
				} else if infoItem.ItemName == nodeinfocheck.RegionItemType {
					nodeInfo.Region = infoItem.Result.(string)
				} else if infoItem.ItemName == nodeinfocheck.InstanceTypeItemType {
					nodeInfo.Type = infoItem.Result.(string)
				}
			}
		}
	}

	return checkItemList, infoItemList, nodeGVSMap
}

// Ready xxx
func (p *Plugin) Ready(clusterID string) bool {
	p.WriteLock.Lock()
	defer p.WriteLock.Unlock()
	return p.ReadyMap[clusterID]
}

// GetResult xxx
func (p *Plugin) GetResult(s string) plugin_manager.CheckResult {
	return p.Result[s]
}

// GetDetail xxx
func (p *Plugin) GetDetail() interface{} {
	return nil
}
