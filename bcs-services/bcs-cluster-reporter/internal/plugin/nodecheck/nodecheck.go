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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metricmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/diskcheck"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/netcheck"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/nodeinfocheck"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/timecheck"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	yaml "gopkg.in/yaml.v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/dnscheck"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/hwcheck"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/processcheck"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

// Plugin xxx
type Plugin struct {
	opt            *Options
	testYamlString string
	pluginmanager.ClusterPlugin
}

var (
	nodeAvailabilityLabels = []string{"target", "bk_biz_id", "item", "item_target", "status"}
	nodeAvailability       = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cluster_node_availability",
		Help: "cluster_node_availability, 1 means OK",
	}, nodeAvailabilityLabels)
	nodeAvailabilityGaugeVecSetMap = make(map[string][]*metricmanager.GaugeVecSet)
)

func init() {
	metricmanager.Register(nodeAvailability)
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

	p.Result = make(map[string]pluginmanager.CheckResult)
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
		metricmanager.SetCommonDurationMetric([]string{p.Name(), "", "", ""}, start)
	}()

	clusterConfigs := pluginmanager.Pm.GetConfig().ClusterConfigs

	wg := sync.WaitGroup{}

	// 遍历所有集群
	for _, cluster := range clusterConfigs {
		wg.Add(1)
		pluginmanager.Pm.Add()

		go func(cluster *pluginmanager.ClusterConfig) {
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
				pluginmanager.Pm.Done()
				p.WriteLock.Lock()
				p.ReadyMap[cluster.ClusterID] = true
				p.WriteLock.Unlock()
				wg.Done()
				klog.Infof("end nodecheck for %s", cluster.ClusterID)
			}()
			clusterResult := pluginmanager.CheckResult{
				Items: make([]pluginmanager.CheckItem, 0, 0),
			}

			clientSet, _ := k8s.GetClientsetByConfig(config)
			cmList, err := clientSet.CoreV1().ConfigMaps(nodeagentNamespace).List(util.GetCtx(10*time.Second), v1.ListOptions{
				ResourceVersion: "0",
			})
			if err != nil {
				klog.Errorf("get nodeagent configmap from cluster %s failed: %s", clusterId, err.Error())
				return
			}

			nodeAvailabilityGVSMap := make(map[string][]*metricmanager.GaugeVecSet)
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

				nodeinfo := make(map[string]pluginmanager.PluginInfo)
				err = yaml.Unmarshal([]byte(configmap.Data["nodeinfo"]), nodeinfo)
				if err != nil {
					//klog.Errorf("unmarshal %s nodeinfo %s failed: %s", clusterId, configmap.Name, err.Error())
					continue
				}

				// 获取节点的checkitem并生成metric的map
				nodeInfo := plugin.NodeInfo{}
				checkItemList, infoItemList, nodeGVSMap := checkNodePluginResult(nodeinfo, strings.TrimSuffix(configmap.Name, "-v1"), clusterId, clusterbiz, &nodeInfo)
				// 一个节点每类异常指标只能有一个
				for name, list := range nodeGVSMap {
					if len(list) > 1 {
						nodeGVSMap[name] = list[:1]
					}
				}

				clusterResult.Items = append(clusterResult.Items, checkItemList...)
				clusterResult.InfoItemList = append(clusterResult.InfoItemList, infoItemList...)
				for key, nodeGVSList := range nodeGVSMap {
					if _, ok := nodeAvailabilityGVSMap[key]; !ok {
						nodeAvailabilityGVSMap[key] = make([]*metricmanager.GaugeVecSet, 0, 0)
					}
					nodeAvailabilityGVSMap[key] = append(nodeAvailabilityGVSMap[key], nodeGVSList...)
				}

				cluster.NodeInfo[nodeName] = nodeInfo
			}

			nodeAvailabilityGaugeVecSetList := make([]*metricmanager.GaugeVecSet, 0, 0)
			for key, gvsList := range nodeAvailabilityGVSMap {
				if len(gvsList) == 0 {
					nodeAvailabilityGaugeVecSetList = append(nodeAvailabilityGaugeVecSetList, &metricmanager.GaugeVecSet{
						Labels: []string{clusterId, clusterbiz, key, "node", normalStatus},
						Value:  1,
					})
				} else {
					for _, gaugeVecSet := range gvsList {
						nodeAvailabilityGaugeVecSetList = append(nodeAvailabilityGaugeVecSetList, gaugeVecSet)
					}
				}
			}

			p.WriteLock.Lock()
			metricmanager.DeleteMetric(nodeAvailability, nodeAvailabilityGaugeVecSetMap[clusterId])
			nodeAvailabilityGaugeVecSetMap[clusterId] = nodeAvailabilityGaugeVecSetList
			metricmanager.SetMetric(nodeAvailability, nodeAvailabilityGaugeVecSetMap[clusterId])
			p.Result[clusterId] = clusterResult
			p.WriteLock.Unlock()
		}(cluster)
	}

	wg.Wait()

	// clean deleted cluster data
	for clusterID, _ := range p.ReadyMap {
		if _, ok := clusterConfigs[clusterID]; !ok {
			metricmanager.DeleteMetric(nodeAvailability, nodeAvailabilityGaugeVecSetMap[clusterID])
			delete(p.ReadyMap, clusterID)
			delete(nodeAvailabilityGaugeVecSetMap, clusterID)
			delete(p.Result, clusterID)
			klog.Infof("delete cluster %s", clusterID)
		}
	}
}

// checkNodePluginResult 解析node check PluginInfo
func checkNodePluginResult(nodeinfo map[string]pluginmanager.PluginInfo, nodeName string, clusterId, clusterbiz string, nodeInfo *plugin.NodeInfo) ([]pluginmanager.CheckItem, []pluginmanager.InfoItem, map[string][]*metricmanager.GaugeVecSet) {
	checkItemList := make([]pluginmanager.CheckItem, 0, 0)
	infoItemList := make([]pluginmanager.InfoItem, 0, 0)
	nodeGVSMap := make(map[string][]*metricmanager.GaugeVecSet)

	// 所有节点检测项，不管正常与否都应该返回对应checkitem
	for name, pluginInfo := range nodeinfo {
		for _, checkItem := range pluginInfo.Result.Items {
			if _, ok := nodeGVSMap[checkItem.ItemName]; !ok {
				nodeGVSMap[checkItem.ItemName] = make([]*metricmanager.GaugeVecSet, 0, 0)
			}

			nodeGVSMap[checkItem.ItemName] = append(nodeGVSMap[checkItem.ItemName], &metricmanager.GaugeVecSet{
				Labels: []string{clusterId, clusterbiz, checkItem.ItemName, "node", checkItem.Status},
				Value:  1,
			})
		}

		switch name {
		case "processcheck":
			pluginCheckItemList, gvsList, err := getProcessCheckResult(pluginInfo, nodeName, clusterId, clusterbiz)
			if err != nil {
				klog.Errorf(err.Error())
				continue
			}
			checkItemList = append(checkItemList, pluginCheckItemList...)
			nodeGVSMap[processConfigCheckItem] = gvsList

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
			checkItemList = append(checkItemList, getNodeinfoCheckResult(pluginInfo, nodeInfo)...)
		}
	}

	return checkItemList, infoItemList, nodeGVSMap
}

func getProcessCheckResult(pluginInfo pluginmanager.PluginInfo, nodeName, clusterId, clusterbiz string) ([]pluginmanager.CheckItem, []*metricmanager.GaugeVecSet, error) {
	checkItemList := make([]pluginmanager.CheckItem, 0)
	nodeGVSMap := make([]*metricmanager.GaugeVecSet, 0)
	for _, checkItem := range pluginInfo.Result.Items {
		checkItem.ItemName = processcheck.StringMap[checkItem.ItemName]
		checkItem.Status = processcheck.StringMap[checkItem.Status]
		checkItemList = append(checkItemList, checkItem)
	}
	detailBytes, err := yaml.Marshal(pluginInfo.Detail)
	if err != nil {
		return checkItemList, nodeGVSMap, err
	}

	detail := processcheck.Detail{}
	err = yaml.Unmarshal(detailBytes, &detail)
	if err != nil {
		return checkItemList, nodeGVSMap, err
	}

	// 检查进程配置，生成checkitem
	processResult := checkProcess(detail, nodeName)
	checkItemList = append(checkItemList, processResult...)

	for index, checkItem := range processResult {
		nodeGVSMap = append(nodeGVSMap, &metricmanager.GaugeVecSet{
			Labels: []string{clusterId, clusterbiz, checkItem.ItemName, "node", checkItem.Status},
			Value:  1,
		})

		checkItem.ItemName = StringMap[checkItem.ItemName]
		checkItem.Status = StringMap[checkItem.Status]
		processResult[index] = checkItem
	}

	return checkItemList, nodeGVSMap, nil
}

func getNodeinfoCheckResult(pluginInfo pluginmanager.PluginInfo, nodeInfo *plugin.NodeInfo) []pluginmanager.CheckItem {
	checkItemList := make([]pluginmanager.CheckItem, 0)

	for _, checkItem := range pluginInfo.Result.Items {
		checkItem.ItemName = nodeinfocheck.StringMap[checkItem.ItemName]
		checkItem.Status = nodeinfocheck.StringMap[checkItem.Status]
		checkItemList = append(checkItemList, checkItem)
	}

	for _, infoItem := range pluginInfo.Result.InfoItemList {
		switch infoItem.ItemName {
		case nodeinfocheck.ZoneItemType:
			nodeInfo.Zone = infoItem.Result.(string)
		case nodeinfocheck.RegionItemType:
			nodeInfo.Region = infoItem.Result.(string)
		case nodeinfocheck.InstanceTypeItemType:
			nodeInfo.Type = infoItem.Result.(string)
		}
	}

	return checkItemList
}

// Ready xxx
func (p *Plugin) Ready(clusterID string) bool {
	p.WriteLock.Lock()
	defer p.WriteLock.Unlock()
	return p.ReadyMap[clusterID]
}

// GetResult xxx
func (p *Plugin) GetResult(s string) pluginmanager.CheckResult {
	return p.Result[s]
}

// GetDetail xxx
func (p *Plugin) GetDetail() interface{} {
	return nil
}
