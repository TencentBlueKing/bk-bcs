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

// Package capacitycheck xxx
package capacitycheck

import (
	"bytes"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metric_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"k8s.io/kubectl/pkg/rawhttp"
)

// Plugin xxx
type Plugin struct {
	opt            *Options
	testYamlString string
	plugin_manager.ClusterPlugin
}

var (
	clusterGVSMap   = make(map[string][]*metric_manager.GaugeVecSet)
	clusterCapacity = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: ClusterCapacityMetricName,
		Help: ClusterCapacityMetricName,
	}, []string{"target", "bk_biz_id", "item", "status"})

	routinePool = util.NewRoutinePool(20)
)

func init() {
	metric_manager.Register(clusterCapacity)
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

	p.Result = make(map[string]plugin_manager.CheckResult)
	p.ReadyMap = make(map[string]bool)

	interval := p.opt.Interval
	if interval == 0 {
		interval = 60
	}

	if runMode == plugin_manager.RunModeDaemon {
		go func() {
			for {
				if p.CheckLock.TryLock() {
					p.CheckLock.Unlock()
					if p.opt.Synchronization {
						plugin_manager.Pm.Lock()
					}
					go p.Check()
				} else {
					klog.V(3).Infof("the former clustercheck didn't over, skip in this loop")
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
	} else if runMode == plugin_manager.RunModeOnce {
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

func int64Ptr(i int64) *int64 { return &i }

// Check xxx
func (p *Plugin) Check() {
	start := time.Now()
	p.CheckLock.Lock()
	klog.Infof("start %s", p.Name())
	defer func() {
		klog.Infof("end %s", p.Name())
		if p.opt.Synchronization {
			plugin_manager.Pm.UnLock()
		}
		p.CheckLock.Unlock()
		metric_manager.SetCommonDurationMetric([]string{"clustercheck", "", "", ""}, start)
	}()

	clusterConfigs := plugin_manager.Pm.GetConfig().ClusterConfigs

	wg := sync.WaitGroup{}

	// 遍历所有集群
	for _, cluster := range clusterConfigs {
		wg.Add(1)
		routinePool.Add(1)

		plugin_manager.Pm.Ready("systemappcheck,nodecheck", cluster.ClusterID)
		go func(cluster *plugin_manager.ClusterConfig) {
			cluster.Lock()
			klog.Infof("start capacitycheck for %s", cluster.ClusterID)
			gvsList := make([]*metric_manager.GaugeVecSet, 0, 0)

			p.WriteLock.Lock()
			p.ReadyMap[cluster.ClusterID] = false
			p.WriteLock.Unlock()

			defer func() {
				wg.Done()
				routinePool.Done()
				p.WriteLock.Lock()
				p.ReadyMap[cluster.ClusterID] = true
				p.WriteLock.Unlock()
				cluster.Unlock()
				klog.Infof("end capacitycheck for %s", cluster.ClusterID)
			}()

			clusterResult := plugin_manager.CheckResult{
				Items:        make([]plugin_manager.CheckItem, 0, 0),
				InfoItemList: make([]plugin_manager.InfoItem, 0, 0),
			}

			defer func() {
				p.WriteLock.Lock()
				for key, val := range clusterResult.Items {
					val.ItemName = StringMap[val.ItemName]
					val.ItemTarget = StringMap[val.ItemTarget]
					val.Status = StringMap[val.Status]
					clusterResult.Items[key] = val
				}

				p.Result[cluster.ClusterID] = clusterResult
				p.WriteLock.Unlock()
			}()

			// 获取apiserver的metric指标
			metricFamilies, err := GetApiserverMetric(cluster.Config)
			if err != nil {
				klog.Errorf("get cluster %s metric failed: %s", cluster.ClusterID, err.Error())
				return
			}

			// 获取集群 各类resource的object数量
			resourceList := []string{"pods", "nodes", "services", "configmaps"}
			for _, resource := range resourceList {
				objectNum, err := GetObjectNum(metricFamilies, resource)
				if err != nil {
					klog.Errorf("get cluster %s %s failed: %s", cluster.ClusterID, resource, err.Error())
					continue
				}

				clusterResult.InfoItemList = append(clusterResult.InfoItemList, plugin_manager.InfoItem{
					ItemName: fmt.Sprintf("%s num", resource),
					Result:   objectNum,
				})

				switch resource {
				case "services":
					cluster.ServiceNum = objectNum
				case "nodes":
					cluster.NodeNum = objectNum
				}

				gvsList = append(gvsList, &metric_manager.GaugeVecSet{
					Labels: []string{cluster.ClusterID, cluster.BusinessID, fmt.Sprintf("%s num", resource), NormalStatus},
					Value:  float64(objectNum),
				})
			}

			// 获取集群的service网段信息
			if _, _, err := net.ParseCIDR(cluster.ServiceCidr); err == nil {
				mask, _ := strconv.Atoi(strings.Split(cluster.ServiceCidr, "/")[1])
				cluster.ServiceMaxNum = 1 << uint(32-mask)

				clusterResult.InfoItemList = append(clusterResult.InfoItemList, plugin_manager.InfoItem{
					ItemName: ServiceMaxNumCheckItemType,
					Result:   cluster.ServiceMaxNum,
				})

				clusterResult.InfoItemList = append(clusterResult.InfoItemList, plugin_manager.InfoItem{
					ItemName: ServiceCidrCheckItemType,
					Result:   cluster.ServiceCidr,
				})

				gvsList = append(gvsList, &metric_manager.GaugeVecSet{
					Labels: []string{cluster.ClusterID, cluster.BusinessID, ServiceNumCheckItemType, NormalStatus},
					Value:  float64(cluster.ServiceMaxNum - cluster.ServiceNum),
				})
			} else {
				klog.Errorf("%s parse service cidr %s failed: %s", cluster.ClusterID, cluster.ServiceCidr, err.Error())
			}

			// 获取集群还可以分配的cidr数量
			if len(cluster.Cidr) > 0 {
				totalIPNum := 0
				nodePodNum := math.Pow(2, float64(32-cluster.MaskSize))
				for _, cidr := range cluster.Cidr {
					mask, _ := strconv.Atoi(strings.Split(cidr, "/")[1])
					ipNum := math.Pow(2, float64(32-mask))
					totalIPNum = totalIPNum + int(ipNum)
				}
				totalIPNum = totalIPNum - cluster.ServiceMaxNum

				maxNodeNum := totalIPNum / int(nodePodNum)

				// cidr允许的最大节点数
				clusterResult.InfoItemList = append(clusterResult.InfoItemList, plugin_manager.InfoItem{
					ItemName: NodeCidrNumCheckItemType,
					Result:   maxNodeNum,
				})

				clusterResult.InfoItemList = append(clusterResult.InfoItemList, plugin_manager.InfoItem{
					ItemName: NodeMaxPodCheckItemType,
					Result:   nodePodNum,
				})

				gvsList = append(gvsList, &metric_manager.GaugeVecSet{
					Labels: []string{cluster.ClusterID, cluster.BusinessID, NodeCidrNumCheckItemType, NormalStatus},
					Value:  float64(maxNodeNum),
				})
			}

			// master检查
			checkItemList, masterGVSList := GetMasterCheckResult(cluster)
			gvsList = append(gvsList, masterGVSList...)
			clusterResult.Items = append(clusterResult.Items, checkItemList...)

			// node检查
			infoItemList, masterGVSList := GetNodeCheckResult(cluster)
			gvsList = append(gvsList, masterGVSList...)
			clusterResult.InfoItemList = append(clusterResult.InfoItemList, infoItemList...)

			p.WriteLock.Lock()
			metric_manager.DeleteMetric(clusterCapacity, clusterGVSMap[cluster.ClusterID])
			metric_manager.SetMetric(clusterCapacity, gvsList)
			clusterGVSMap[cluster.ClusterID] = gvsList
			p.WriteLock.Unlock()
		}(cluster)

	}
	wg.Wait()

	p.WriteLock.Lock()
	for clusterID, _ := range p.ReadyMap {
		if _, ok := clusterConfigs[clusterID]; !ok {
			p.ReadyMap[clusterID] = false
			klog.Infof("delete cluster %s", clusterID)
		}
	}

	for clusterID, ready := range p.ReadyMap {
		if !ready {
			delete(p.ReadyMap, clusterID)
			metric_manager.DeleteMetric(clusterCapacity, clusterGVSMap[clusterID])
			delete(clusterGVSMap, clusterID)
		}
	}
	p.WriteLock.Unlock()
}

// GetApiserverMetric Get metric from apiserver api
func GetApiserverMetric(config *rest.Config) (map[string]*io_prometheus_client.MetricFamily, error) {
	metricServer := config.Host + "/metrics"
	out := &bytes.Buffer{}
	o := genericclioptions.IOStreams{In: &bytes.Buffer{}, Out: out, ErrOut: &bytes.Buffer{}}

	config.GroupVersion = &schema.GroupVersion{Group: "mygroup", Version: "v1"}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: serializer.NewCodecFactory(runtime.NewScheme())}
	c, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, err
	}
	err = rawhttp.RawGet(c, o, metricServer)
	if err != nil {
		return nil, err
	} else {
		var parser expfmt.TextParser
		metricFamilies, err := parser.TextToMetricFamilies(out)
		if err != nil {
			return nil, err
		}

		return metricFamilies, nil
	}

}

// GetObjectNum Get object number by metric data
func GetObjectNum(metricFamilies map[string]*io_prometheus_client.MetricFamily, resource string) (int, error) {
	for key, metricFamily := range metricFamilies {
		if key == "etcd_object_counts" || key == "apiserver_storage_objects" {
			for _, metric := range metricFamily.Metric {
				for _, label := range metric.Label {
					if *label.Name == "resource" && *label.Value == resource {
						return int(*metric.Gauge.Value), nil
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("not found %s metric", resource)

}

// GetMasterCheckResult Check master info and generate check result
func GetMasterCheckResult(clusterInfo *plugin_manager.ClusterConfig) ([]plugin_manager.CheckItem, []*metric_manager.GaugeVecSet) {
	gvsList := make([]*metric_manager.GaugeVecSet, 0, 0)
	checkItemList := make([]plugin_manager.CheckItem, 0, 0)

	checkItem := plugin_manager.CheckItem{
		ItemName:   pluginName,
		ItemTarget: MasterTarget,
		Status:     NormalStatus,
		Normal:     len(clusterInfo.Master) >= 3,
		Level:      plugin_manager.WARNLevel,
		Tags:       nil,
	}
	if len(clusterInfo.Master) < 3 {
		checkItem.Status = MasterNumHAErrorStatus
		checkItem.Detail = fmt.Sprintf(StringMap[MasterNumDetailFormart], len(clusterInfo.Master))
		gvsList = append(gvsList, &metric_manager.GaugeVecSet{
			Labels: []string{clusterInfo.ClusterID, clusterInfo.BusinessID, MasterNumItemType, MasterNumHAErrorStatus},
			Value:  float64(len(clusterInfo.Master))})
	} else {
		gvsList = append(gvsList, &metric_manager.GaugeVecSet{
			Labels: []string{clusterInfo.ClusterID, clusterInfo.BusinessID, MasterNumItemType, NormalStatus},
			Value:  float64(len(clusterInfo.Master))})
	}

	checkItemList = append(checkItemList, checkItem)

	zoneNum := make(map[string]int)
	for _, master := range clusterInfo.Master {
		if clusterInfo.NodeInfo[master].Zone == "" {
			continue
		}
		zoneNum[clusterInfo.NodeInfo[master].Zone] = zoneNum[clusterInfo.NodeInfo[master].Zone] + 1
	}

	for zone, num := range zoneNum {
		if (num)*2 >= len(clusterInfo.Master) {
			checkItem = plugin_manager.CheckItem{
				ItemName:   pluginName,
				ItemTarget: MasterTarget,
				Status:     MasterZoneHAErrorStatus,
				Normal:     (num)*2 < len(clusterInfo.Master),
				Detail:     fmt.Sprintf(StringMap[MasterZoneDetailFormart], zone, num),
				Level:      plugin_manager.WARNLevel,
				Tags:       nil,
			}
			checkItemList = append(checkItemList, checkItem)

			gvsList = append(gvsList, &metric_manager.GaugeVecSet{
				Labels: []string{clusterInfo.ClusterID, clusterInfo.BusinessID, MasterZoneItemType, MasterZoneHAErrorStatus},
				Value:  1})
			break
		}
	}

	return checkItemList, gvsList
}

// GetNodeCheckResult Check node info and generate check result
func GetNodeCheckResult(clusterInfo *plugin_manager.ClusterConfig) ([]plugin_manager.InfoItem, []*metric_manager.GaugeVecSet) {
	gvsList := make([]*metric_manager.GaugeVecSet, 0, 0)
	infoItemList := make([]plugin_manager.InfoItem, 0, 0)

	zoneNum := make(map[string]int)
	for _, nodeInfo := range clusterInfo.NodeInfo {
		if nodeInfo.Zone == "" {
			continue
		}
		zoneNum[nodeInfo.Zone] = zoneNum[nodeInfo.Zone] + 1
	}

	for zone, num := range zoneNum {
		infoItemList = append(infoItemList, plugin_manager.InfoItem{
			ItemName: pluginName,
			Labels:   map[string]string{"zone": zone},
			Result:   fmt.Sprintf("%d", num),
		})
	}

	typeNum := make(map[string]int)
	for _, nodeInfo := range clusterInfo.NodeInfo {
		if nodeInfo.Type == "" {
			continue
		}
		typeNum[nodeInfo.Type] = typeNum[nodeInfo.Type] + 1
	}

	for nodeType, num := range typeNum {
		infoItemList = append(infoItemList, plugin_manager.InfoItem{
			ItemName: pluginName,
			Labels:   map[string]string{"type": nodeType},
			Result:   fmt.Sprintf("%d", num),
		})
	}

	return infoItemList, gvsList
}

// Ready return true if cluster check is over
func (p *Plugin) Ready(clusterID string) bool {
	p.WriteLock.Lock()
	defer p.WriteLock.Unlock()
	return p.ReadyMap[clusterID]
}

// GetResult return check result by cluster ID
func (p *Plugin) GetResult(clusterID string) plugin_manager.CheckResult {
	return p.Result[clusterID]
}
