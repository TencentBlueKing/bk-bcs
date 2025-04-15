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

// Package systemappcheck xxx
package systemappcheck

import (
	"fmt"
	utiltrace "k8s.io/utils/trace"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metricmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"
	internalUtil "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
	"github.com/hashicorp/go-version"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// Plugin xxx
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

// Setup xxx
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

	if runMode == pluginmanager.RunModeDaemon {
		go func() {
			for {
				if p.CheckLock.TryLock() {
					p.CheckLock.Unlock()
					pluginmanager.Pm.Lock()

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
					go p.Check()
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
		pluginmanager.Pm.UnLock()
		p.CheckLock.Unlock()
		metricmanager.SetCommonDurationMetric([]string{"systemappcheck", "", "", ""}, start)
	}()

	wg := sync.WaitGroup{}
	// 默认获取以下这些namespace中的应用
	namespaceList := p.opt.Namespaces

	for _, cluster := range pluginmanager.Pm.GetConfig().ClusterConfigs {
		wg.Add(1)
		pluginmanager.Pm.Add()

		go func(cluster *pluginmanager.ClusterConfig) {
			cluster.Lock()
			klog.Infof("start systemappcheck for %s", cluster.ClusterID)

			clusterId := cluster.ClusterID
			clientSet := cluster.ClientSet
			config := cluster.Config
			clusterResult := pluginmanager.CheckResult{
				Items: make([]pluginmanager.CheckItem, 0, 0),
			}

			trace := utiltrace.New("systemappcheck", utiltrace.Field{"target", clusterId})

			defer func() {
				cluster.Unlock()
				klog.Infof("end systemappcheck for %s", clusterId)
				wg.Done()
				pluginmanager.Pm.Done()
				p.WriteLock.Lock()
				p.ReadyMap[cluster.ClusterID] = true
				p.WriteLock.Unlock()
				trace.LogIfLong(20 * time.Second)
			}()

			p.WriteLock.Lock()
			p.ReadyMap[cluster.ClusterID] = false
			p.WriteLock.Unlock()

			loopSystemAppChartGVSList := make([]*metricmanager.GaugeVecSet, 0, 0)
			loopSystemAppImageGVSList := make([]*metricmanager.GaugeVecSet, 0, 0)
			loopSystemAppStatusGVSList := make([]*metricmanager.GaugeVecSet, 0, 0)
			loopSystemAppConfigGVSList := make([]*metricmanager.GaugeVecSet, 0, 0)

			clusterVersion, err := k8s.GetK8sVersion(clientSet)
			if err != nil {
				klog.Errorf("%s GetK8sVersion failed: %s", clusterId, err.Error())
				return
			}

			// don't check for v1.8 cluster
			if strings.Contains(clusterVersion, "v1.8") {
				klog.Infof("%s version is %s, skip", clusterId, clusterVersion)
				return
			}

			// 检查静态pod配置
			podCheckItemList, staticPodGVSList, err := CheckStaticPod(cluster)
			if err != nil {
				klog.Errorf("%s CheckStaticPod failed %s", clusterId, err.Error())
			} else {
				loopSystemAppConfigGVSList = append(loopSystemAppConfigGVSList, staticPodGVSList...)
				clusterResult.Items = append(clusterResult.Items, podCheckItemList...)
			}
			trace.Step("check static pod")

			// 检查svc配置
			svcCheckItemList, GVSList, err := CheckService(cluster, clusterId)
			if err != nil {
				klog.Errorf("%s CheckService failed %s", clusterId, err.Error())
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
			for index, _ := range p.opt.Components {
				if strings.EqualFold(p.opt.Components[index].Resource, deployment) &&
					strings.EqualFold(p.opt.Components[index].Resource, daemonset) &&
					strings.EqualFold(p.opt.Components[index].Resource, statefulset) {
					klog.Errorf("unsupported resource type %s", p.opt.Components[index])
					continue
				}
				clusterComponent = append(clusterComponent, p.opt.Components[index])
			}

			// 遍历各个namespace下的releases，基于release进行应用状态的检测
			checkItemList, relStatusGVSList, statusGVSList, imageGVSList, configGVSList, componentList :=
				CheckRelease(namespaceList, getter, cluster, p.opt.ComponentVersionConf)
			clusterResult.Items = append(clusterResult.Items, checkItemList...)
			loopSystemAppChartGVSList = append(loopSystemAppChartGVSList, relStatusGVSList...)
			loopSystemAppStatusGVSList = append(loopSystemAppStatusGVSList, statusGVSList...)
			loopSystemAppImageGVSList = append(loopSystemAppImageGVSList, imageGVSList...)
			loopSystemAppConfigGVSList = append(loopSystemAppConfigGVSList, configGVSList...)

			// 如果component列表中的应用已经通过helm部署了，则不再需要单独确认
			for _, component := range componentList {
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
				componentCheckItemList, componentStatusGVSList, componentImageGVSList, componentConfigGVSList, err :=
					CheckComponent(component, cluster, "", p.opt.ComponentVersionConf)
				if err != nil {
					klog.Errorf("%s %s CheckComponent failed: %s", clusterId, component.Name, err.Error())
				}

				clusterResult.Items = append(clusterResult.Items, componentCheckItemList...)
				loopSystemAppStatusGVSList = append(loopSystemAppStatusGVSList, componentStatusGVSList...)
				loopSystemAppImageGVSList = append(loopSystemAppImageGVSList, componentImageGVSList...)
				loopSystemAppConfigGVSList = append(loopSystemAppConfigGVSList, componentConfigGVSList...)
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
			if _, ok := systemAppChartGVSList[cluster.ClusterID]; ok {
				metricmanager.DeleteMetric(systemAppChartVersion, systemAppChartGVSList[clusterId])
				metricmanager.DeleteMetric(systemAppImageVersion, systemAppImageGVSList[clusterId])
				metricmanager.DeleteMetric(systemAppStatus, systemAppStatusGVSList[clusterId])
				metricmanager.DeleteMetric(systemAppConfig, systemAppConfigGVSList[clusterId])
			}

			systemAppChartGVSList[clusterId] = loopSystemAppChartGVSList
			systemAppImageGVSList[clusterId] = loopSystemAppImageGVSList
			systemAppStatusGVSList[clusterId] = loopSystemAppStatusGVSList
			systemAppConfigGVSList[clusterId] = loopSystemAppConfigGVSList
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
			p.Result[clusterId] = clusterResult
			p.WriteLock.Unlock()

			trace.Step("refresh metric")
		}(cluster)
	}
	wg.Wait()

	// clean deleted cluster metric data
	clusterConfigs := pluginmanager.Pm.GetConfig().ClusterConfigs
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
}

// CheckRelease 检查release详情
func CheckRelease(namespaceList []string, getter *k8s.RESTClientGetter, cluster *pluginmanager.ClusterConfig, componentVersionConfList []ComponentVersionConf) (
	[]pluginmanager.CheckItem, []*metricmanager.GaugeVecSet, []*metricmanager.GaugeVecSet, []*metricmanager.GaugeVecSet, []*metricmanager.GaugeVecSet, []Component) {

	var syncLock sync.Mutex
	checkItemList := make([]pluginmanager.CheckItem, 0, 0)
	relStatusGvsList := make([]*metricmanager.GaugeVecSet, 0, 0)
	relComponentStatusGVSList := make([]*metricmanager.GaugeVecSet, 0, 0)
	relImageGVSList := make([]*metricmanager.GaugeVecSet, 0, 0)
	relConfigGVSList := make([]*metricmanager.GaugeVecSet, 0, 0)
	componentList := make([]Component, 0, 0)

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
			// client.AllNamespaces = true
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
				checkItemList = append(checkItemList, checkItem)

				// 生成release状态指标
				relStatusGvsList = append(relStatusGvsList, &metricmanager.GaugeVecSet{
					Labels: []string{cluster.ClusterID, cluster.BusinessID, namespace, rel.Chart.Name(), rel.Chart.Metadata.Version, rel.Info.Status.String(), rel.Name},
					Value:  1,
				})
				syncLock.Unlock()

				manifest := rel.Manifest
				// exclude non-workload resource manifest
				resourceRe := regexp.MustCompile(`(?m)^---$`)
				resourceManifestList := resourceRe.Split(manifest, -1)

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
						relCheckItemList, statusGVSList, imageGVSList, configGVSList, err :=
							CheckComponent(component, cluster, rel.Name, componentVersionConfList)
						if err != nil {
							klog.Errorf("%s %s %s CheckComponent failed: %s", cluster.ClusterID, rel.Name, component.Name, err.Error())
						}
						syncLock.Lock()
						componentList = append(componentList, component)
						checkItemList = append(checkItemList, relCheckItemList...)
						relComponentStatusGVSList = append(relComponentStatusGVSList, statusGVSList...)
						relImageGVSList = append(relImageGVSList, imageGVSList...)
						relConfigGVSList = append(relConfigGVSList, configGVSList...)
						syncLock.Unlock()
					}
				}
			}
		}(namespace)
	}

	wg.Wait()
	return checkItemList, relStatusGvsList, relComponentStatusGVSList, relImageGVSList, relConfigGVSList, componentList
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

// CheckComponent 检查应用workload详情
func CheckComponent(component Component, cluster *pluginmanager.ClusterConfig, rel string,
	componentVersionConfList []ComponentVersionConf) (
	[]pluginmanager.CheckItem, []*metricmanager.GaugeVecSet, []*metricmanager.GaugeVecSet, []*metricmanager.GaugeVecSet, error) {
	// daemonset可能按需部署，比如vcluster不部署daemonset
	if component.Name == "kube-proxy" && cluster.ClusterType == "virtual" {
		return nil, nil, nil, nil, nil
	} else if component.Name == "kube-proxy" && cluster.ALLEKLET {
		return nil, nil, nil, nil, nil
	}

	checkItemList := make([]pluginmanager.CheckItem, 0, 0)
	statusGVSList := make([]*metricmanager.GaugeVecSet, 0, 0)
	imageGVSList := make([]*metricmanager.GaugeVecSet, 0, 0)
	configGVSList := make([]*metricmanager.GaugeVecSet, 0, 0)

	// 检查workload status, 是否ready
	status, containerImageList, err := getWorkLoadStatus(component.Resource, component.Name, component.Namespace, cluster)
	if err != nil {
		checkItemList = append(checkItemList, pluginmanager.CheckItem{
			ItemName:   SystemAppStatusCheckItemName,
			ItemTarget: component.Name,
			Status:     status,
			Level:      pluginmanager.WARNLevel,
			Detail:     err.Error(),
			Normal:     status == NormalStatus,
			Tags:       map[string]string{"component": component.Name},
		})

		statusGVSList = append(statusGVSList, &metricmanager.GaugeVecSet{
			Labels: []string{cluster.ClusterID, cluster.BusinessID, component.Namespace, component.Name, component.Resource, status, rel},
			Value:  1,
		})
		return checkItemList, statusGVSList, imageGVSList, configGVSList, err
	}

	if status != NormalStatus {
		klog.Infof("%s %s %s status is %s", cluster.ClusterID, component.Name, component.Namespace, status)
		// 如果找不到对应的workload，则直接返回
		if status == APPNotfoundAppStatus {
			return checkItemList, statusGVSList, imageGVSList, configGVSList, err
		}
	}

	checkItemList = append(checkItemList, pluginmanager.CheckItem{
		ItemName:   SystemAppStatusCheckItemName,
		ItemTarget: component.Name,
		Status:     status,
		Detail:     "",
		Level:      pluginmanager.WARNLevel,
		Normal:     status == NormalStatus,
		Tags:       map[string]string{"component": component.Name},
	})
	statusGVSList = append(statusGVSList, &metricmanager.GaugeVecSet{
		Labels: []string{cluster.ClusterID, cluster.BusinessID, component.Namespace, component.Name, component.Resource, status, rel},
		Value:  1,
	})

	// 检查workload镜像
	for _, ci := range containerImageList {
		status = GetImageStatus(ci, componentVersionConfList)
		checkItemList = append(checkItemList, pluginmanager.CheckItem{
			ItemName:   SystemAppImageVersionCheckItemName,
			ItemTarget: component.Name,
			Status:     NormalStatus,
			Detail:     "",
			Level:      pluginmanager.WARNLevel,
			Normal:     status == NormalStatus,
			Tags:       map[string]string{"component": component.Name},
		})

		imageGVSList = append(imageGVSList, &metricmanager.GaugeVecSet{
			Labels: []string{cluster.ClusterID, cluster.BusinessID, component.Namespace, component.Name, component.Resource, ci.container, ci.image, status, rel},
			Value:  1,
		})
	}

	return checkItemList, statusGVSList, imageGVSList, configGVSList, err
}

// getWorkLoadStatus xxx
func getWorkLoadStatus(workloadType string, workloadName string, namespace string, cluster *pluginmanager.ClusterConfig) (string, []containerImage, error) {
	var containerImageList []containerImage
	var status = NormalStatus
	var err error
	ctx := internalUtil.GetCtx(10 * time.Second)
	switch strings.ToLower(workloadType) {
	case strings.ToLower(deployment):
		var deploy *v1.Deployment
		deploy, err = cluster.ClientSet.AppsV1().Deployments(namespace).Get(ctx, workloadName, metav1.GetOptions{
			ResourceVersion: "0",
		})

		if err != nil {
			status = AppErrorStatus
			if strings.Contains(err.Error(), "not found") {
				status = APPNotfoundAppStatus
			}
		} else {
			containerImageList, status, err = getDeployCheckResult(deploy, cluster)
		}

	case strings.ToLower(daemonset):
		var ds *v1.DaemonSet
		ds, err = cluster.ClientSet.AppsV1().DaemonSets(namespace).Get(ctx, workloadName, metav1.GetOptions{
			ResourceVersion: "0",
		})

		if err != nil {
			status = AppErrorStatus
			if strings.Contains(err.Error(), "not found") {
				status = APPNotfoundAppStatus
			}
		} else {
			containerImageList, status, err = getDSCheckResult(ds, cluster)
		}

	case strings.ToLower(statefulset):
		var sts *v1.StatefulSet
		sts, err = cluster.ClientSet.AppsV1().StatefulSets(namespace).Get(ctx, workloadName, metav1.GetOptions{
			ResourceVersion: "0",
		})

		if err != nil {
			status = AppErrorStatus
			if strings.Contains(err.Error(), "not found") {
				status = APPNotfoundAppStatus
			}
		} else {
			containerImageList, status, err = getSTSCheckResult(sts, cluster)
		}
	default:
		klog.Errorf("unsupported resource type %s", workloadType)
	}

	if err != nil {
		return status, nil, fmt.Errorf("namespace: %s worload: %s GetWorkLoad failed: %s", namespace, workloadName, err.Error())
	}

	return status, containerImageList, nil

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
		if strings.Contains(err.Error(), "the server could not find the requested resource") {
			return NormalStatus, nil
		} else {
			return AppMetricErrorStatus, err
		}
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
				return ImageStatusNeedUpgrade
			}

			if imageVersion.LessThan(niceToUpgradeVersion) {
				return ImageStatusNiceToUpgrade
			}
		}

		return NormalStatus
	}
	return NormalStatus
}
