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

// Package systemappcheck 系统应用检查插件，检查集群中系统组件的部署状态、镜像版本和配置
package systemappcheck

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	utiltrace "k8s.io/utils/trace"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metricmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"
	internalUtil "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

// Plugin 系统应用检查插件
type Plugin struct {
	opt *Options
	pluginmanager.ClusterPlugin
}

var (
	systemAppChartVersionLabels = []string{"target", "bk_biz_id", "namespace", "chart", "version", "status", "rel"}
	systemAppImageVersionLabels = []string{"target", "bk_biz_id", "namespace", "component", "resource",
		"container", "version", "status", "rel"}
	systemAppStatusLabels = []string{"target", "bk_biz_id", "namespace", "component", "resource", "status", "rel"}
	systemAppConfigLabels = []string{"target", "bk_biz_id", "namespace", "component", "resource", "status"}
	// 应用release 状态检查
	systemAppChartVersion = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "system_app_chart_version",
		Help: "system_app_chart_version, 1 means deployed",
	}, systemAppChartVersionLabels)
	// 应用镜像版本检查
	systemAppImageVersion = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "system_app_image_version",
		Help: "system_app_image_version, 1 means ok",
	}, systemAppImageVersionLabels)
	// 应用部署状态
	systemAppStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "system_app_status",
		Help: "system_app_status, 1 means ok",
	}, systemAppStatusLabels)
	// 应用配置状态
	systemAppConfig = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "system_app_config",
		Help: "system_app_config, 1 means ok",
	}, systemAppConfigLabels)

	systemAppChartGVSList  = make(map[string][]*metricmanager.GaugeVecSet)
	systemAppImageGVSList  = make(map[string][]*metricmanager.GaugeVecSet)
	systemAppStatusGVSList = make(map[string][]*metricmanager.GaugeVecSet)
	systemAppConfigGVSList = make(map[string][]*metricmanager.GaugeVecSet)
)

func init() {
	// 注册指标
	metricmanager.Register(systemAppChartVersion)
	metricmanager.Register(systemAppImageVersion)
	metricmanager.Register(systemAppStatus)
	metricmanager.Register(systemAppConfig)
}

// Setup 初始化插件配置并启动检查循环
func (p *Plugin) Setup(configFilePath string, runMode string) error {
	p.opt = &Options{}

	err := internalUtil.ReadorInitConf(configFilePath, p.opt, initContent)
	if err != nil {
		return err
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

	checkOpt := pluginmanager.CheckOption{}

	if runMode == pluginmanager.RunModeDaemon {
		go func() {
			for {
				if p.CheckLock.TryLock() {
					p.CheckLock.Unlock()

					// 重载opt
					opt := &Options{}
					err = internalUtil.ReadorInitConf(configFilePath, opt, initContent)
					if err != nil {
						klog.Errorf("reload config failed: %s", err.Error())
					} else {
						if err = p.opt.Validate(); err != nil {
							klog.Errorf("validate config failed: %s", err.Error())
						} else {
							if reflect.DeepEqual(&p.opt, &opt) {
								p.opt = opt
								klog.Info("config reload success")
							}
						}
					}
					// 执行检查
					go p.Check(checkOpt)
				} else {
					klog.V(3).Infof("the former systemappcheck didn't over, skip in this loop")
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
	} else if runMode == pluginmanager.RunModeOnce {
		p.Check(checkOpt)
	}

	return nil
}

// Stop 停止插件运行
func (p *Plugin) Stop() error {
	p.StopChan <- 1
	klog.Infof("plugin %s stopped", p.Name())
	return nil
}

// Name 返回插件名称
func (p *Plugin) Name() string {
	return pluginName
}

// Check 执行系统应用检查
func (p *Plugin) Check(option pluginmanager.CheckOption) {
	start := time.Now()
	p.CheckLock.Lock()
	klog.Infof("start %s", p.Name())
	defer func() {
		klog.Infof("end %s", p.Name())
		p.CheckLock.Unlock()
		if option.ClusterIDs == nil || len(option.ClusterIDs) == 0 {
			metricmanager.SetCommonDurationMetric([]string{p.PluginName, "", "", ""}, start)
		}
	}()

	clusterConfigs := make(map[string]*pluginmanager.ClusterConfig)
	if option.ClusterIDs == nil || len(option.ClusterIDs) == 0 {
		clusterConfigs = pluginmanager.Pm.GetConfig().ClusterConfigs
		// clean deleted cluster metric data
		p.WriteLock.Lock()
		for clusterID, _ := range p.ReadyMap {
			if _, ok := clusterConfigs[clusterID]; !ok {
				delete(p.ReadyMap, clusterID)
				metricmanager.DeleteMetric(systemAppChartVersion, systemAppChartGVSList[clusterID])
				metricmanager.DeleteMetric(systemAppImageVersion, systemAppImageGVSList[clusterID])
				metricmanager.DeleteMetric(systemAppStatus, systemAppStatusGVSList[clusterID])
				metricmanager.DeleteMetric(systemAppConfig, systemAppConfigGVSList[clusterID])
				delete(systemAppChartGVSList, clusterID)
				delete(systemAppImageGVSList, clusterID)
				delete(systemAppStatusGVSList, clusterID)
				delete(systemAppConfigGVSList, clusterID)
				delete(p.Result, clusterID)
				klog.Infof("delete cluster %s", clusterID)
			}
		}
		p.WriteLock.Unlock()
	} else {
		for _, clusterID := range option.ClusterIDs {
			if cluster, ok := pluginmanager.Pm.GetConfig().ClusterConfigs[clusterID]; ok {
				clusterConfigs[clusterID] = cluster
			}
		}
	}

	p.checkClusters(clusterConfigs, option)
}

func (p *Plugin) checkClusters(clusterConfigs map[string]*pluginmanager.ClusterConfig, option pluginmanager.CheckOption) {
	wg := sync.WaitGroup{}
	// 默认获取以下这些namespace中的应用
	namespaceList := p.opt.Namespaces
	for _, cluster := range clusterConfigs {
		wg.Add(1)
		if len(option.ClusterIDs) > 0 {
			pluginmanager.Pm.AddCheck()
		} else {
			pluginmanager.Pm.Add()
		}

		go func(cluster *pluginmanager.ClusterConfig) {
			cluster.Lock()
			klog.Infof("start %s for %s", p.Name(), cluster.ClusterID)

			clusterID := cluster.ClusterID
			clientSet := cluster.ClientSet
			config := cluster.Config
			clusterResult := pluginmanager.CheckResult{
				Items: make([]pluginmanager.CheckItem, 0, 0),
			}

			trace := utiltrace.New("systemappcheck", utiltrace.Field{"target", clusterID})

			defer func() {
				cluster.Unlock()
				klog.Infof("end systemappcheck for %s", clusterID)
				wg.Done()
				if len(option.ClusterIDs) > 0 {
					pluginmanager.Pm.DoneCheck()
				} else {
					pluginmanager.Pm.Done()
				}
				p.WriteLock.Lock()
				p.ReadyMap[cluster.ClusterID] = true
				p.WriteLock.Unlock()
				trace.LogIfLong(20 * time.Second)
			}()

			p.WriteLock.Lock()
			p.ReadyMap[cluster.ClusterID] = false
			// return if all nodes are master.
			if len(cluster.Master) == cluster.NodeNum {
				metricmanager.DeleteMetric(systemAppChartVersion, systemAppChartGVSList[clusterID])
				metricmanager.DeleteMetric(systemAppImageVersion, systemAppImageGVSList[clusterID])
				metricmanager.DeleteMetric(systemAppStatus, systemAppStatusGVSList[clusterID])
				metricmanager.DeleteMetric(systemAppConfig, systemAppConfigGVSList[clusterID])
				p.Result[clusterID] = clusterResult
				p.WriteLock.Unlock()
				return
			}
			p.WriteLock.Unlock()

			loopSystemAppChartGVSList := make([]*metricmanager.GaugeVecSet, 0, 0)
			loopSystemAppImageGVSList := make([]*metricmanager.GaugeVecSet, 0, 0)
			loopSystemAppStatusGVSList := make([]*metricmanager.GaugeVecSet, 0, 0)
			loopSystemAppConfigGVSList := make([]*metricmanager.GaugeVecSet, 0, 0)

			clusterVersion, err := k8s.GetK8sVersion(clientSet)
			if err != nil {
				klog.Errorf("%s GetK8sVersion failed: %s", clusterID, err.Error())
				return
			}

			// don't check for v1.8 cluster
			if strings.Contains(clusterVersion, "v1.8") {
				klog.Infof("%s version is %s, skip", clusterID, clusterVersion)
				return
			}

			// 检查静态pod配置，支持deep
			podCheckItemList, staticPodGVSList, err := CheckStaticPod(cluster, option.DeepCheck)
			if err != nil {
				klog.Errorf("%s CheckStaticPod failed %s", clusterID, err.Error())
			} else {
				loopSystemAppConfigGVSList = append(loopSystemAppConfigGVSList, staticPodGVSList...)
				clusterResult.Items = append(clusterResult.Items, podCheckItemList...)
			}
			trace.Step("check static pod")

			// 检查svc配置
			svcCheckItemList, GVSList, err := CheckService(cluster, clusterID)
			if err != nil {
				klog.Errorf("%s CheckService failed %s", clusterID, err.Error())
			} else {
				// 生成异常配置指标
				loopSystemAppConfigGVSList = append(loopSystemAppConfigGVSList, GVSList...)
				clusterResult.Items = append(clusterResult.Items, svcCheckItemList...)
			}

			// 检查tke服务的配置
			if cluster.ClusterType == pluginmanager.TKECluster {
				// 检查TKE应用配置
				CheckTKENetwork(cluster)
			}

			trace.Step("check service")

			getter := k8s.GetRestClientGetterByConfig(config)

			// 获取配置文件中的component列表(有可能不是helm方式部署的)
			clusterComponent := make([]Component, 0, 0)
			for index := range p.opt.Components {
				if !strings.EqualFold(p.opt.Components[index].Resource, deployment) &&
					!strings.EqualFold(p.opt.Components[index].Resource, daemonset) &&
					!strings.EqualFold(p.opt.Components[index].Resource, statefulset) {
					klog.Errorf("unsupported resource type %s", p.opt.Components[index].Resource)
					continue
				}
				clusterComponent = append(clusterComponent, p.opt.Components[index])
			}

			// 遍历各个namespace下的releases，基于release进行应用状态的检测
			releaseResult := CheckRelease(namespaceList, getter, cluster, p.opt.ComponentVersionConf)
			clusterResult.Items = append(clusterResult.Items, releaseResult.CheckItems...)
			loopSystemAppChartGVSList = append(loopSystemAppChartGVSList, releaseResult.ChartGVSList...)
			loopSystemAppStatusGVSList = append(loopSystemAppStatusGVSList, releaseResult.StatusGVSList...)
			loopSystemAppImageGVSList = append(loopSystemAppImageGVSList, releaseResult.ImageGVSList...)
			loopSystemAppConfigGVSList = append(loopSystemAppConfigGVSList, releaseResult.ConfigGVSList...)

			// 如果component列表中的应用已经通过helm部署了，则不再需要单独确认
			for _, component := range releaseResult.ComponentList {
				for index, c := range clusterComponent {
					if c.Name == component.Name &&
						c.Namespace == component.Namespace &&
						strings.EqualFold(c.Resource, component.Resource) {
						clusterComponent = append(clusterComponent[:index], clusterComponent[index+1:]...)
						break
					}
				}
			}
			trace.Step("check release")

			// 检查指定组件
			for _, component := range clusterComponent {
				compResult, err := CheckComponent(component, cluster, "", p.opt.ComponentVersionConf)
				if err != nil {
					klog.Errorf("%s %s CheckComponent failed: %s", clusterID, component.Name, err.Error())
				}

				clusterResult.Items = append(clusterResult.Items, compResult.CheckItems...)
				loopSystemAppStatusGVSList = append(loopSystemAppStatusGVSList, compResult.StatusGVSList...)
				loopSystemAppImageGVSList = append(loopSystemAppImageGVSList, compResult.ImageGVSList...)
				loopSystemAppConfigGVSList = append(loopSystemAppConfigGVSList, compResult.ConfigGVSList...)
			}
			trace.Step("check component")

			// 检查特定systemapp的config
			workCheckItemList, workGVSList := CheckSystemWorkLoadConfig(cluster)
			loopSystemAppConfigGVSList = append(loopSystemAppConfigGVSList, workGVSList...)
			clusterResult.Items = append(clusterResult.Items, workCheckItemList...)
			trace.Step("check config")

			// 刷新metric
			// get former metric data
			p.WriteLock.Lock()
			// 删除上一次检查的指标
			metricmanager.DeleteMetric(systemAppChartVersion, systemAppChartGVSList[clusterID])
			metricmanager.DeleteMetric(systemAppImageVersion, systemAppImageGVSList[clusterID])
			metricmanager.DeleteMetric(systemAppStatus, systemAppStatusGVSList[clusterID])
			metricmanager.DeleteMetric(systemAppConfig, systemAppConfigGVSList[clusterID])

			systemAppChartGVSList[clusterID] = loopSystemAppChartGVSList
			systemAppImageGVSList[clusterID] = loopSystemAppImageGVSList
			systemAppStatusGVSList[clusterID] = loopSystemAppStatusGVSList
			systemAppConfigGVSList[clusterID] = loopSystemAppConfigGVSList
			// get new metric data
			p.WriteLock.Unlock()

			// 写入指标
			metricmanager.SetMetric(systemAppChartVersion, loopSystemAppChartGVSList)
			metricmanager.SetMetric(systemAppImageVersion, loopSystemAppImageGVSList)
			metricmanager.SetMetric(systemAppStatus, loopSystemAppStatusGVSList)
			metricmanager.SetMetric(systemAppConfig, loopSystemAppConfigGVSList)

			for key, val := range clusterResult.Items {
				val.ItemName = StringMap[val.ItemName]
				if _, ok := StringMap[val.ItemTarget]; ok {
					val.ItemTarget = StringMap[val.ItemTarget]
				}
				val.Status = StringMap[val.Status]
				clusterResult.Items[key] = val
			}

			p.WriteLock.Lock()
			p.Result[clusterID] = clusterResult
			p.WriteLock.Unlock()

			trace.Step("refresh metric")
		}(cluster)
	}
	wg.Wait()
}

// CheckRelease 检查release详情
func CheckRelease(namespaceList []string, getter *k8s.RESTClientGetter, cluster *pluginmanager.ClusterConfig, componentVersionConfList []ComponentVersionConf) ReleaseCheckResult {
	var syncLock sync.Mutex
	result := ReleaseCheckResult{
		CheckItems:    make([]pluginmanager.CheckItem, 0),
		ChartGVSList:  make([]*metricmanager.GaugeVecSet, 0),
		StatusGVSList: make([]*metricmanager.GaugeVecSet, 0),
		ImageGVSList:  make([]*metricmanager.GaugeVecSet, 0),
		ConfigGVSList: make([]*metricmanager.GaugeVecSet, 0),
		ComponentList: make([]Component, 0),
	}

	var wg sync.WaitGroup

	// 遍历指定的namespace
	for _, namespace := range namespaceList {
		wg.Add(1)
		go func(namespace string) {
			defer wg.Done()

			actionConfig := new(action.Configuration)
			if err := actionConfig.Init(getter, namespace, os.Getenv("HELM_DRIVER"), klog.Infof); err != nil {
				klog.Errorf("%s Config helm client failed: %s", cluster.ClusterID, err.Error())
				return
			}

			// 获取release列表
			client := action.NewList(actionConfig)
			client.Deployed = true
			relList, err := client.Run()
			if err != nil {
				klog.Errorf("%s helm get deployed chart failed: %s", cluster.ClusterID, err.Error())
				return
			}

			// 基于helm release进行检查
			for _, rel := range relList {
				// 生成对应的release 状态的checkitem
				syncLock.Lock()
				checkItem := pluginmanager.CheckItem{
					ItemName:   pluginName,
					ItemTarget: rel.Name,
					Status:     rel.Info.Status.String(),
					Detail:     "",
					Level:      pluginmanager.WARNLevel,
					Normal:     rel.Info.Status.String() == ChartVersionNormalStatus,
					Tags:       map[string]string{"component": rel.Name},
				}

				if checkItem.Normal {
					checkItem.Status = NormalStatus
				}
				result.CheckItems = append(result.CheckItems, checkItem)

				// 生成release状态指标
				result.ChartGVSList = append(result.ChartGVSList, &metricmanager.GaugeVecSet{
					Labels: []string{cluster.ClusterID, cluster.BusinessID, namespace, rel.Chart.Name(), rel.Chart.Metadata.Version, rel.Info.Status.String(), rel.Name},
					Value:  1,
				})
				syncLock.Unlock()

				manifest := rel.Manifest
				// exclude non-workload resource manifest
				resourceManifestList := resourceManifestRe.Split(manifest, -1)

				workloadTypeList := []string{deployment, daemonset, statefulset}
				workLoadManifestMap := make(map[string][]string)

				// 获取各个workload的yaml文件
				for _, workloadType := range workloadTypeList {
					workloadRe, _ := regexp.Compile(fmt.Sprintf("\nkind: %s", workloadType))
					for _, resourceManifest := range resourceManifestList {
						if workloadRe.MatchString(resourceManifest) {
							if _, ok := workLoadManifestMap[workloadType]; !ok {
								workLoadManifestMap[workloadType] = make([]string, 0, 0)
							}
							workLoadManifestMap[workloadType] = append(workLoadManifestMap[workloadType], resourceManifest)
						}
					}
				}

				// 通过workload的yaml文件 获取各个应用的状态
				for _, worloadManifestList := range workLoadManifestMap {
					for _, worloadManifest := range worloadManifestList {
						component := Component{Namespace: namespace}
						err = GetComponentFromManifest(worloadManifest, &component)
						if err != nil {
							klog.Errorf("GetComponentFromManifest %s failed: %s", rel.Name, err.Error())
							continue
						}

						// tke daemonset告警过多，先不进行检查
						if !strings.Contains(strings.ToLower(component.Name), "bk") &&
							!strings.Contains(strings.ToLower(component.Name), "bcs") && component.Resource == daemonset {
							continue
						}

						// 获取release下workload的详细信息
						compResult, err := CheckComponent(component, cluster, rel.Name, componentVersionConfList)
						if err != nil {
							klog.Errorf("%s %s %s CheckComponent failed: %s", cluster.ClusterID, rel.Name, component.Name, err.Error())
						}
						syncLock.Lock()
						result.ComponentList = append(result.ComponentList, component)
						result.CheckItems = append(result.CheckItems, compResult.CheckItems...)
						result.StatusGVSList = append(result.StatusGVSList, compResult.StatusGVSList...)
						result.ImageGVSList = append(result.ImageGVSList, compResult.ImageGVSList...)
						result.ConfigGVSList = append(result.ConfigGVSList, compResult.ConfigGVSList...)
						syncLock.Unlock()
					}
				}
			}
		}(namespace)
	}

	wg.Wait()
	return result
}

// GetStatus get workload ready status
func GetStatus(updatedReplicas int, availableReplicas int, replicas int, nodeNum int) string {
	// 如果节点数较少则只期望可用副本数>=nodenum
	if nodeNum < replicas && availableReplicas >= nodeNum {
		return NormalStatus
		// 期望副本数=可用副本数=最新副本数
	} else if availableReplicas == replicas && updatedReplicas == replicas {
		return NormalStatus
	} else {
		return AppStatusNotReadyStatus
	}

}

// ComponentCheckResult 组件检查结果，封装 CheckComponent 的返回值
type ComponentCheckResult struct {
	CheckItems    []pluginmanager.CheckItem
	StatusGVSList []*metricmanager.GaugeVecSet
	ImageGVSList  []*metricmanager.GaugeVecSet
	ConfigGVSList []*metricmanager.GaugeVecSet
}

// newEmptyComponentCheckResult 创建空的组件检查结果
func newEmptyComponentCheckResult() ComponentCheckResult {
	return ComponentCheckResult{
		CheckItems:    make([]pluginmanager.CheckItem, 0),
		StatusGVSList: make([]*metricmanager.GaugeVecSet, 0),
		ImageGVSList:  make([]*metricmanager.GaugeVecSet, 0),
		ConfigGVSList: make([]*metricmanager.GaugeVecSet, 0),
	}
}

// ReleaseCheckResult release 检查结果，封装 CheckRelease 的返回值
type ReleaseCheckResult struct {
	CheckItems       []pluginmanager.CheckItem
	ChartGVSList     []*metricmanager.GaugeVecSet
	StatusGVSList    []*metricmanager.GaugeVecSet
	ImageGVSList     []*metricmanager.GaugeVecSet
	ConfigGVSList    []*metricmanager.GaugeVecSet
	ComponentList    []Component
}

// CheckComponent 检查应用workload详情
func CheckComponent(component Component, cluster *pluginmanager.ClusterConfig, rel string,
	componentVersionConfList []ComponentVersionConf) (ComponentCheckResult, error) {
	// daemonset可能按需部署，比如vcluster不部署daemonset
	if component.Name == "kube-proxy" && cluster.ClusterType == "virtual" {
		return newEmptyComponentCheckResult(), nil
	} else if component.Name == "kube-proxy" && cluster.ALLEKLET {
		return newEmptyComponentCheckResult(), nil
	}

	result := newEmptyComponentCheckResult()

	// 检查workload status, 是否ready
	status, containerImageList, err := getWorkLoadStatus(component.Resource, component.Name, component.Namespace, cluster)
	if err != nil {
		// 有的业务有些组件可能并不需要或以其它形式部署，如果并非helm部署且不存在，不再认为异常
		if status == AppNotFoundStatus && rel == "" {
			return result, nil
		}

		result.CheckItems = append(result.CheckItems, pluginmanager.CheckItem{
			ItemName:   SystemAppStatusCheckItemName,
			ItemTarget: component.Name,
			Status:     status,
			Level:      pluginmanager.WARNLevel,
			Detail:     err.Error(),
			Normal:     status == NormalStatus,
			Tags:       map[string]string{"component": component.Name},
		})

		result.StatusGVSList = append(result.StatusGVSList, &metricmanager.GaugeVecSet{
			Labels: []string{cluster.ClusterID, cluster.BusinessID, component.Namespace, component.Name, component.Resource, status, rel},
			Value:  1,
		})
		return result, err
	}

	if status != NormalStatus {
		klog.Infof("%s %s %s status is %s", cluster.ClusterID, component.Name, component.Namespace, status)
		// 如果找不到对应的workload，则直接返回
		if status == AppNotFoundStatus {
			return result, err
		}
	}

	result.CheckItems = append(result.CheckItems, pluginmanager.CheckItem{
		ItemName:   SystemAppStatusCheckItemName,
		ItemTarget: component.Name,
		Status:     status,
		Detail:     "",
		Level:      pluginmanager.WARNLevel,
		Normal:     status == NormalStatus,
		Tags:       map[string]string{"component": component.Name},
	})
	result.StatusGVSList = append(result.StatusGVSList, &metricmanager.GaugeVecSet{
		Labels: []string{cluster.ClusterID, cluster.BusinessID, component.Namespace, component.Name, component.Resource, status, rel},
		Value:  1,
	})

	// 检查workload镜像
	for _, ci := range containerImageList {
		status = GetImageStatus(ci, componentVersionConfList)
		result.CheckItems = append(result.CheckItems, pluginmanager.CheckItem{
			ItemName:   SystemAppImageVersionCheckItemName,
			ItemTarget: component.Name,
			Status:     NormalStatus,
			Detail:     "",
			Level:      pluginmanager.WARNLevel,
			Normal:     status == NormalStatus,
			Tags:       map[string]string{"component": component.Name},
		})

		result.ImageGVSList = append(result.ImageGVSList, &metricmanager.GaugeVecSet{
			Labels: []string{cluster.ClusterID, cluster.BusinessID, component.Namespace, component.Name, component.Resource, ci.container, ci.image, status, rel},
			Value:  1,
		})
	}

	return result, err
}

// isRetryableError 判断错误是否为可重试的临时性错误，如超时、服务暂时不可用等
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	// 超时类错误
	if strings.Contains(errMsg, "Timeout") || strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
		return true
	}
	// apiserver 暂时不可用（503），如 "the server is currently unable to handle the request"
	if strings.Contains(errMsg, "unable to handle the request") || strings.Contains(errMsg, "ServiceUnavailable") {
		return true
	}
	// 连接拒绝/重置等网络抖动
	if strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "connection reset") {
		return true
	}
	return false
}

// workloadFetchResult workload 获取结果
type workloadFetchResult struct {
	containerImages []containerImage
	status          string
	err             error
}

// handleWorkloadError 统一处理 workload 获取错误，返回对应的 status
func handleWorkloadError(err error) string {
	if err == nil {
		return NormalStatus
	}
	if strings.Contains(err.Error(), "not found") {
		return AppNotFoundStatus
	}
	return AppErrorStatus
}

// fetchWorkload 根据资源类型获取 workload 并返回检查结果，消除三种资源类型的重复代码
func fetchWorkload(workloadType, workloadName, namespace string, cluster *pluginmanager.ClusterConfig, ctx context.Context) workloadFetchResult {
	switch strings.ToLower(workloadType) {
	case strings.ToLower(deployment):
		deploy, err := cluster.ClientSet.AppsV1().Deployments(namespace).Get(ctx, workloadName, metav1.GetOptions{
			ResourceVersion: "0",
		})
		if err != nil {
			return workloadFetchResult{status: handleWorkloadError(err), err: err}
		}
		containerImages, status, err := getDeployCheckResult(deploy, cluster)
		return workloadFetchResult{containerImages: containerImages, status: status, err: err}

	case strings.ToLower(daemonset):
		ds, err := cluster.ClientSet.AppsV1().DaemonSets(namespace).Get(ctx, workloadName, metav1.GetOptions{
			ResourceVersion: "0",
		})
		if err != nil {
			return workloadFetchResult{status: handleWorkloadError(err), err: err}
		}
		containerImages, status, err := getDSCheckResult(ds, cluster)
		return workloadFetchResult{containerImages: containerImages, status: status, err: err}

	case strings.ToLower(statefulset):
		sts, err := cluster.ClientSet.AppsV1().StatefulSets(namespace).Get(ctx, workloadName, metav1.GetOptions{
			ResourceVersion: "0",
		})
		if err != nil {
			return workloadFetchResult{status: handleWorkloadError(err), err: err}
		}
		containerImages, status, err := getSTSCheckResult(sts, cluster)
		return workloadFetchResult{containerImages: containerImages, status: status, err: err}

	default:
		klog.Errorf("unsupported resource type %s", workloadType)
		return workloadFetchResult{status: AppErrorStatus, err: fmt.Errorf("unsupported resource type %s", workloadType)}
	}
}

// getWorkLoadStatus 获取workload的状态信息，支持对临时性错误进行重试
func getWorkLoadStatus(workloadType string, workloadName string, namespace string, cluster *pluginmanager.ClusterConfig) (string, []containerImage, error) {
	const maxRetry = 3
	ctx := internalUtil.GetCtx(10 * time.Second)

	for i := 0; i < maxRetry; i++ {
		result := fetchWorkload(workloadType, workloadName, namespace, cluster, ctx)

		if result.err == nil {
			return result.status, result.containerImages, nil
		}

		if isRetryableError(result.err) {
			klog.Warningf("namespace: %s workload: %s GetWorkLoad retryable error (attempt %d/%d): %s",
				namespace, workloadName, i+1, maxRetry, result.err.Error())
			time.Sleep(time.Duration(i+1) * 2 * time.Second)
			ctx = internalUtil.GetCtx(20 * time.Second)
			continue
		}

		return result.status, nil, fmt.Errorf("namespace: %s worload: %s GetWorkLoad failed: %s", namespace, workloadName, result.err.Error())
	}

	return AppErrorStatus, nil, fmt.Errorf("namespace: %s worload: %s GetWorkLoad failed after %d retries", namespace, workloadName, maxRetry)
}

// GetComponentFromManifest get component by parsing workload manifest
func GetComponentFromManifest(worloadManifest string, component *Component) error {
	var objMap map[string]interface{}
	if err := yaml.Unmarshal([]byte(worloadManifest), &objMap); err != nil {
		return fmt.Errorf("Error unmarshalling YAML: %v", err)
	}
	if _, ok := objMap["kind"]; !ok {
		return fmt.Errorf("wrong workload yaml, no kind: %s", worloadManifest)
	}
	if _, ok := objMap["metadata"]; !ok {
		return fmt.Errorf("wrong workload yaml, no metadata%s", worloadManifest)
	}

	if _, ok := objMap["metadata"].(map[interface{}]interface{}); ok {
		if _, ok = objMap["metadata"].(map[interface{}]interface{})["namespace"]; ok {
			if namespace, ok := objMap["metadata"].(map[interface{}]interface{})["namespace"].(string); ok {
				component.Namespace = namespace
			}
		}
	} else {
		return fmt.Errorf("wrong workload yaml, wrong metadata type %s", worloadManifest)
	}

	component.Resource = objMap["kind"].(string)
	component.Name = objMap["metadata"].(map[interface{}]interface{})["name"].(string)

	return nil
}

type containerImage struct {
	container string
	image     string
}

// getSTSCheckResult 检查statefulset类型workload
func getSTSCheckResult(sts *v1.StatefulSet, cluster *pluginmanager.ClusterConfig) ([]containerImage, string, error) {
	containerImageList := make([]containerImage, 0, 0)
	for _, container := range sts.Spec.Template.Spec.Containers {
		containerImageList = append(containerImageList, containerImage{
			container: container.Name,
			image:     container.Image,
		})
	}

	// 获取workload当前是否ready
	status := GetStatus(
		int(sts.Status.UpdatedReplicas),
		int(sts.Status.ReadyReplicas),
		int(sts.Status.Replicas), cluster.NodeNum)

	if status != NormalStatus {
		return containerImageList, status, nil
	}

	// 检测workload当前资源使用情况
	status, err := CheckPodMetric(sts.Spec.Template.Spec.Containers, cluster, sts.Namespace, sts.Spec.Selector.MatchLabels)
	return containerImageList, status, err
}

// getDSCheckResult 检查daemonset类型workload
func getDSCheckResult(ds *v1.DaemonSet, cluster *pluginmanager.ClusterConfig) ([]containerImage, string, error) {
	containerImageList := make([]containerImage, 0, 0)
	for _, container := range ds.Spec.Template.Spec.Containers {
		containerImageList = append(containerImageList, containerImage{
			container: container.Name,
			image:     container.Image,
		})
	}

	// 获取workload当前是否ready
	status := GetStatus(
		int(ds.Status.UpdatedNumberScheduled),
		int(ds.Status.NumberReady),
		int(ds.Status.DesiredNumberScheduled), cluster.NodeNum)
	if status != NormalStatus {
		return containerImageList, status, nil
	}

	// ds不检测资源使用情况
	//status, err := CheckPodMetric(ds.Spec.Template.Spec.Containers, ms, ds.Namespace, ds.Spec.Selector.MatchLabels)
	return containerImageList, status, nil
}

// getDeployCheckResult 检查deployment类型workload
func getDeployCheckResult(deploy *v1.Deployment, cluster *pluginmanager.ClusterConfig) ([]containerImage, string, error) {
	containerImageList := make([]containerImage, 0, 0)
	for _, container := range deploy.Spec.Template.Spec.Containers {
		containerImageList = append(containerImageList, containerImage{
			container: container.Name,
			image:     container.Image,
		})
	}

	// 获取workload当前是否ready
	status := GetStatus(
		int(deploy.Status.UpdatedReplicas),
		int(deploy.Status.ReadyReplicas),
		int(deploy.Status.Replicas), cluster.NodeNum)
	if status != NormalStatus {
		return containerImageList, status, nil
	}

	// 检测workload当前资源使用情况
	status, err := CheckPodMetric(deploy.Spec.Template.Spec.Containers, cluster, deploy.Namespace, deploy.Spec.Selector.MatchLabels)
	return containerImageList, status, err
}

// CheckPodMetric check pod resource metric, generate high load gvs
func CheckPodMetric(containerList []corev1.Container, cluster *pluginmanager.ClusterConfig, namespace string, matchLabels map[string]string) (string, error) {
	if cluster.MetricSet == nil {
		return NormalStatus, nil
	}
	// 基于workload的label获取pod metric
	podMetricList, err := cluster.MetricSet.MetricsV1beta1().PodMetricses(namespace).List(internalUtil.GetCtx(15*time.Second),
		metav1.ListOptions{
			ResourceVersion: "0",
			LabelSelector:   metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: matchLabels})})

	if err != nil {
		klog.Infof("%s get metric failed: %s", cluster.ClusterID, err.Error())
		return NormalStatus, nil
	}

	for _, container := range containerList {
		for _, podMetric := range podMetricList.Items {
			for _, containerMetric := range podMetric.Containers {
				if container.Name != containerMetric.Name {
					continue
				}

				// 如果使用率大于95%则返回异常status
				if container.Resources.Limits.Memory().MilliValue() != 0 {
					memoryUsagePercent := containerMetric.Usage.Memory().MilliValue() * 100 / container.Resources.Limits.Memory().MilliValue()
					if memoryUsagePercent > 95 {
						return AppStatusMemoryHighStatus, nil
					}
				}

				if container.Resources.Limits.Cpu().MilliValue() != 0 {
					cpuUsagePercent := containerMetric.Usage.Cpu().MilliValue() * 100 / container.Resources.Limits.Cpu().MilliValue()
					if cpuUsagePercent > 95 {
						return AppStatusCpuHighStatus, nil
					}
				}
			}
		}

	}

	return NormalStatus, nil
}

// Ready return true if cluster check is over
func (p *Plugin) Ready(clusterID string) bool {
	p.WriteLock.Lock()
	defer p.WriteLock.Unlock()
	return p.ReadyMap[clusterID]
}

// GetResult return check result by cluster ID
func (p *Plugin) GetResult(s string) pluginmanager.CheckResult {
	return p.Result[s]
}

// GetImageStatus check container image version
func GetImageStatus(image containerImage, componentVersionConfList []ComponentVersionConf) string {
	for _, versionConf := range componentVersionConfList {
		if versionConf.Name != image.container {
			continue
		}

		images := strings.Split(image.image, ":")
		if len(images) != 2 {
			// image format: mirrors.tencent.com/xxx: v1.29.0-alpha.1
			return ImageStatusUnknown
		}

		imageVersion, err := version.NewVersion(images[1])
		if err != nil {
			return NormalStatus
		}

		if versionConf.NeedUpgrade != "" {
			needUpgradeVersion, err := version.NewVersion(versionConf.NeedUpgrade)
			if err != nil {
				return ImageStatusUnknown
			}

			if imageVersion.LessThan(needUpgradeVersion) {
				return ImageStatusNeedUpgrade
			}
		}

		if versionConf.NiceToUpgrade != "" {
			niceToUpgradeVersion, err := version.NewVersion(versionConf.NiceToUpgrade)
			if err != nil {
				return ImageStatusUnknown
			}

			if imageVersion.LessThan(niceToUpgradeVersion) {
				return ImageStatusNiceToUpgrade
			}
		}

		return NormalStatus
	}
	return NormalStatus
}
