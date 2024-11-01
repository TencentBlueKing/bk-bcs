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

	"github.com/hashicorp/go-version"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metric_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"
	internalUtil "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

// Plugin xxx
type Plugin struct {
	opt *Options
	plugin_manager.ClusterPlugin
}

var (
	systemAppChartVersionLabels = []string{"target", "bk_biz_id", "namespace", "chart", "version", "status", "rel"}
	systemAppImageVersionLabels = []string{"target", "bk_biz_id", "namespace", "component", "resource",
		"container", "version", "status", "rel"}
	systemAppStatusLabels = []string{"target", "bk_biz_id", "namespace", "component", "resource", "status", "rel"}
	systemAppConfigLabels = []string{"target", "bk_biz_id", "namespace", "component", "resource", "status"}
	systemAppChartVersion = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "system_app_chart_version",
		Help: "system_app_chart_version, 1 means deployed",
	}, systemAppChartVersionLabels)
	systemAppImageVersion = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "system_app_image_version",
		Help: "system_app_image_version, 1 means ok",
	}, systemAppImageVersionLabels)
	systemAppStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "system_app_status",
		Help: "system_app_status, 1 means ok",
	}, systemAppStatusLabels)
	systemAppConfig = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "system_app_config",
		Help: "system_app_config, 1 means ok",
	}, systemAppConfigLabels)

	systemAppChartGVSList  = make(map[string][]*metric_manager.GaugeVecSet)
	systemAppImageGVSList  = make(map[string][]*metric_manager.GaugeVecSet)
	systemAppStatusGVSList = make(map[string][]*metric_manager.GaugeVecSet)
	systemAppConfigGVSList = make(map[string][]*metric_manager.GaugeVecSet)
)

func init() {
	metric_manager.Register(systemAppChartVersion)
	metric_manager.Register(systemAppImageVersion)
	metric_manager.Register(systemAppStatus)
	metric_manager.Register(systemAppConfig)
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
					plugin_manager.Pm.Lock()

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
		plugin_manager.Pm.UnLock()
		p.CheckLock.Unlock()
		metric_manager.SetCommonDurationMetric([]string{"systemappcheck", "", "", ""}, start)
	}()

	wg := sync.WaitGroup{}
	// 默认获取以下这些namespace中的应用
	namespaceList := p.opt.Namespaces

	for _, cluster := range plugin_manager.Pm.GetConfig().ClusterConfigs {
		wg.Add(1)
		plugin_manager.Pm.Add()

		go func(cluster *plugin_manager.ClusterConfig) {
			cluster.Lock()
			klog.Infof("start nodecheck for %s", cluster.ClusterID)

			clusterId := cluster.ClusterID
			clusterbiz := cluster.BusinessID
			clientSet := cluster.ClientSet
			config := cluster.Config
			clusterResult := plugin_manager.CheckResult{
				Items: make([]plugin_manager.CheckItem, 0, 0),
			}

			trace := utiltrace.New("systemappcheck", utiltrace.Field{"target", clusterId})

			defer func() {
				cluster.Unlock()
				klog.Infof("end systemappcheck for %s", clusterId)
				wg.Done()
				plugin_manager.Pm.Done()
				p.WriteLock.Lock()
				p.ReadyMap[cluster.ClusterID] = true
				p.WriteLock.Unlock()
				trace.LogIfLong(20 * time.Second)
			}()

			p.WriteLock.Lock()
			p.ReadyMap[cluster.ClusterID] = false
			p.WriteLock.Unlock()

			loopSystemAppChartGVSList := make([]*metric_manager.GaugeVecSet, 0, 0)
			loopSystemAppImageGVSList := make([]*metric_manager.GaugeVecSet, 0, 0)
			loopSystemAppStatusGVSList := make([]*metric_manager.GaugeVecSet, 0, 0)
			loopSystemAppConfigGVSList := make([]*metric_manager.GaugeVecSet, 0, 0)

			metricsClient, err := metricsclientset.NewForConfig(config)
			if err != nil {
				klog.Errorf("%s GetClientsetByClusterId failed: %s", clusterId, err.Error())
				return
			}

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
			if cluster.ClusterType == plugin_manager.TKECluster {
				svcCheckItemList, GVSList, err := CheckService(clientSet, clusterId)
				if err != nil {
					klog.Errorf("%s CheckService failed %s", clusterId, err.Error())
				} else {
					// 生成异常配置指标
					for _, gvs := range GVSList {
						gvs.Labels = append([]string{clusterId, clusterbiz}, gvs.Labels...)
						loopSystemAppConfigGVSList = append(loopSystemAppConfigGVSList, gvs)
					}
					clusterResult.Items = append(clusterResult.Items, svcCheckItemList...)
				}

				// 检查TKE应用配置
				CheckTKENetwork(cluster)
				trace.Step("check service")
			}

			getter := k8s.GetRestClientGetterByConfig(config)

			// 获取配置文件中的component列表(有可能不是helm方式部署的)
			clusterComponent := make([]Component, 0, 0)
			for index, _ := range p.opt.Components {
				switch strings.ToLower(p.opt.Components[index].Resource) {
				case strings.ToLower(deployment):
					p.opt.Components[index].Resource = deployment
					clusterComponent = append(clusterComponent, p.opt.Components[index])
				case strings.ToLower(daemonset):
					p.opt.Components[index].Resource = daemonset
					clusterComponent = append(clusterComponent, p.opt.Components[index])
				case strings.ToLower(statefulset):
					p.opt.Components[index].Resource = statefulset
					clusterComponent = append(clusterComponent, p.opt.Components[index])
				default:
					klog.Errorf("unsupported resource type %s", p.opt.Components[index].Resource)
					continue
				}

			}

			// Get releases，基于release进行检测
			if len(namespaceList) > 0 {
				checkItemList, relStatusGVSList, statusGVSList, imageGVSList, configGVSList, componentList :=
					CheckRelease(namespaceList, getter, clusterId, clusterbiz, clientSet, metricsClient, p.opt.ComponentVersionConf)
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
							c.Resource == component.Resource {
							clusterComponent = append(clusterComponent[:index], clusterComponent[index+1:]...)
							break
						}
					}
				}
			}
			trace.Step("check release")

			// 检查指定组件
			for _, component := range clusterComponent {
				checkItemList, statusGVSList, imageGVSList, configGVSList, status, err :=
					CheckComponent(component, clientSet, metricsClient, clusterId, clusterbiz, "", p.opt.ComponentVersionConf)
				if err != nil {
					klog.Errorf("%s %s CheckComponent failed: %s", clusterId, component.Name, err.Error())
				}

				// 如果应用没有workload和helm release，则认为不需要检查这个应用
				if status == APPNotfoundAppStatus {
					continue
				}

				clusterResult.Items = append(clusterResult.Items, checkItemList...)
				loopSystemAppStatusGVSList = append(loopSystemAppStatusGVSList, statusGVSList...)
				loopSystemAppImageGVSList = append(loopSystemAppImageGVSList, imageGVSList...)
				loopSystemAppConfigGVSList = append(loopSystemAppConfigGVSList, configGVSList...)
			}
			trace.Step("check component")

			// 检查特定systemapp的config
			workCheckItemList := CheckSystemWorkLoadConfig(clientSet)
			for index, checkItem := range workCheckItemList {
				loopSystemAppConfigGVSList = append(loopSystemAppConfigGVSList, &metric_manager.GaugeVecSet{
					Labels: []string{clusterId, clusterbiz, "kube-system", checkItem.ItemTarget, "app", checkItem.Status},
					Value:  1,
				})
				workCheckItemList[index] = checkItem
			}
			trace.Step("check config")

			clusterResult.Items = append(clusterResult.Items, workCheckItemList...)

			// 刷新metric
			// get former metric data
			p.WriteLock.Lock()
			// delete former metric
			if _, ok := systemAppChartGVSList[cluster.ClusterID]; ok {
				metric_manager.DeleteMetric(systemAppChartVersion, systemAppChartGVSList[clusterId])
				metric_manager.DeleteMetric(systemAppImageVersion, systemAppImageGVSList[clusterId])
				metric_manager.DeleteMetric(systemAppStatus, systemAppStatusGVSList[clusterId])
				metric_manager.DeleteMetric(systemAppConfig, systemAppConfigGVSList[clusterId])
			}

			systemAppChartGVSList[clusterId] = loopSystemAppChartGVSList
			systemAppImageGVSList[clusterId] = loopSystemAppImageGVSList
			systemAppStatusGVSList[clusterId] = loopSystemAppStatusGVSList
			systemAppConfigGVSList[clusterId] = loopSystemAppConfigGVSList
			// get new metric data
			for key, val := range clusterResult.Items {
				val.ItemName = StringMap[val.ItemName]
				if _, ok := StringMap[val.ItemTarget]; ok {
					val.ItemTarget = StringMap[val.ItemTarget]
				}
				val.Status = StringMap[val.Status]
				clusterResult.Items[key] = val
			}
			p.Result[clusterId] = clusterResult
			p.WriteLock.Unlock()

			metric_manager.SetMetric(systemAppChartVersion, loopSystemAppChartGVSList)
			metric_manager.SetMetric(systemAppImageVersion, loopSystemAppImageGVSList)
			metric_manager.SetMetric(systemAppStatus, loopSystemAppStatusGVSList)
			metric_manager.SetMetric(systemAppConfig, loopSystemAppConfigGVSList)

			trace.Step("refresh metric")
		}(cluster)
	}
	wg.Wait()

	// clean deleted cluster data
	clusterConfigs := plugin_manager.Pm.GetConfig().ClusterConfigs
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
			metric_manager.DeleteMetric(systemAppChartVersion, systemAppChartGVSList[clusterID])
			metric_manager.DeleteMetric(systemAppImageVersion, systemAppImageGVSList[clusterID])
			metric_manager.DeleteMetric(systemAppStatus, systemAppStatusGVSList[clusterID])
			metric_manager.DeleteMetric(systemAppConfig, systemAppConfigGVSList[clusterID])
			delete(systemAppChartGVSList, clusterID)
			delete(systemAppImageGVSList, clusterID)
			delete(systemAppStatusGVSList, clusterID)
			delete(systemAppConfigGVSList, clusterID)
			delete(p.Result, clusterID)
		}
	}
	p.WriteLock.Unlock()
}

// CheckRelease
func CheckRelease(namespaceList []string, getter *k8s.RESTClientGetter, clusterID, clusterBiz string,
	clientSet *kubernetes.Clientset, metricsClient *metricsclientset.Clientset, componentVersionConfList []ComponentVersionConf) (
	[]plugin_manager.CheckItem, []*metric_manager.GaugeVecSet, []*metric_manager.GaugeVecSet, []*metric_manager.GaugeVecSet, []*metric_manager.GaugeVecSet, []Component) {

	var syncLock sync.Mutex
	checkItemList := make([]plugin_manager.CheckItem, 0, 0)
	relStatusGvsList := make([]*metric_manager.GaugeVecSet, 0, 0)
	relStatusGVSList := make([]*metric_manager.GaugeVecSet, 0, 0)
	relImageGVSList := make([]*metric_manager.GaugeVecSet, 0, 0)
	relConfigGVSList := make([]*metric_manager.GaugeVecSet, 0, 0)
	componentList := make([]Component, 0, 0)

	var wg sync.WaitGroup
	for _, namespace := range namespaceList {
		wg.Add(1)
		go func(namespace string) {
			defer wg.Done()

			actionConfig := new(action.Configuration)
			if err := actionConfig.Init(getter, namespace, os.Getenv("HELM_DRIVER"), klog.Infof); err != nil {
				klog.Errorf("%s Config helm client failed: %s", clusterID, err.Error())
				return
			}

			// 获取release列表
			client := action.NewList(actionConfig)
			client.Deployed = true
			// client.AllNamespaces = true
			relList, err := client.Run()
			if err != nil {
				klog.Errorf("%s helm get deployed chart failed: %s", clusterID, err.Error())
				return
			}

			// 基于helm release进行检查
			for _, rel := range relList {
				// 生成对应的release checkitem
				syncLock.Lock()
				checkItem := plugin_manager.CheckItem{
					ItemName:   pluginName,
					ItemTarget: rel.Name,
					Status:     rel.Info.Status.String(),
					Detail:     "",
					Level:      plugin_manager.WARNLevel,
					Normal:     rel.Info.Status.String() == ChartVersionNormalStatus,
					Tags:       map[string]string{"component": rel.Name},
				}

				if checkItem.Normal {
					checkItem.Status = NormalStatus
				}
				checkItemList = append(checkItemList, checkItem)

				relStatusGvsList = append(relStatusGvsList, &metric_manager.GaugeVecSet{
					Labels: []string{clusterID, clusterBiz, namespace, rel.Chart.Name(), rel.Chart.AppVersion(), rel.Info.Status.String(), rel.Name},
					Value:  1,
				})
				syncLock.Unlock()

				manifest := rel.Manifest
				// exclude non-workload resource manifest
				resourceRe := regexp.MustCompile(`(?m)^---$`)
				resourceManifestList := resourceRe.Split(manifest, -1)

				workloadTypeList := []string{"Deployment", "DaemonSet", "StatefulSet"}
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

				// 获取各个应用的状态
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
						relCheckItemList, statusGVSList, imageGVSList, configGVSList, _, err :=
							CheckComponent(component, clientSet, metricsClient, clusterID, clusterBiz, rel.Name, componentVersionConfList)
						if err != nil {
							klog.Errorf("%s %s %s CheckComponent failed: %s", clusterID, rel.Name, component.Name, err.Error())
						}
						syncLock.Lock()
						componentList = append(componentList, component)
						checkItemList = append(checkItemList, relCheckItemList...)
						relStatusGVSList = append(relStatusGVSList, statusGVSList...)
						relImageGVSList = append(relImageGVSList, imageGVSList...)
						relConfigGVSList = append(relConfigGVSList, configGVSList...)
						syncLock.Unlock()
					}
				}
			}
		}(namespace)
	}

	wg.Wait()
	return checkItemList, relStatusGvsList, relStatusGVSList, relImageGVSList, relConfigGVSList, componentList
}

// GetStatus xxx
func GetStatus(updatedReplicas int, availableReplicas int, replicas int) string {
	if availableReplicas == replicas && updatedReplicas == replicas {
		return NormalStatus
	} else {
		return AppStatusNotReadyStatus
	}

}

// CheckComponent xxx
func CheckComponent(component Component, clientSet *kubernetes.Clientset, metricsClient *metricsclientset.Clientset, clusterId, clusterbiz, rel string,
	componentVersionConfList []ComponentVersionConf) (
	[]plugin_manager.CheckItem, []*metric_manager.GaugeVecSet, []*metric_manager.GaugeVecSet, []*metric_manager.GaugeVecSet, string, error) {
	checkItemList := make([]plugin_manager.CheckItem, 0, 0)
	statusGVSList := make([]*metric_manager.GaugeVecSet, 0, 0)
	imageGVSList := make([]*metric_manager.GaugeVecSet, 0, 0)
	configGVSList := make([]*metric_manager.GaugeVecSet, 0, 0)

	// 检查workload status
	status, containerImageList, err := getWorkLoadStatus(component.Resource, component.Name, component.Namespace, clientSet, metricsClient)
	if err != nil {
		checkItemList = append(checkItemList, plugin_manager.CheckItem{
			ItemName:   SystemAppStatusCheckItemName,
			ItemTarget: component.Name,
			Status:     status,
			Level:      plugin_manager.WARNLevel,
			Detail:     err.Error(),
			Normal:     status == NormalStatus,
			Tags:       map[string]string{"component": component.Name},
		})

		statusGVSList = append(statusGVSList, &metric_manager.GaugeVecSet{
			Labels: []string{clusterId, clusterbiz, component.Namespace, component.Name, component.Resource, status, rel},
			Value:  1,
		})
		return checkItemList, statusGVSList, imageGVSList, configGVSList, status, err
	}

	if status != NormalStatus {
		klog.Infof("%s %s %s status is %s", clusterId, component.Name, component.Namespace, status)
	}

	checkItemList = append(checkItemList, plugin_manager.CheckItem{
		ItemName:   SystemAppStatusCheckItemName,
		ItemTarget: component.Name,
		Status:     status,
		Detail:     "",
		Level:      plugin_manager.WARNLevel,
		Normal:     status == NormalStatus,
		Tags:       map[string]string{"component": component.Name},
	})
	statusGVSList = append(statusGVSList, &metric_manager.GaugeVecSet{
		Labels: []string{clusterId, clusterbiz, component.Namespace, component.Name, component.Resource, status, rel},
		Value:  1,
	})

	// 检查workload镜像
	for _, ci := range containerImageList {
		status = GetImageStatus(ci, componentVersionConfList)
		checkItemList = append(checkItemList, plugin_manager.CheckItem{
			ItemName:   SystemAppImageVersionCheckItemName,
			ItemTarget: component.Name,
			Status:     NormalStatus,
			Detail:     "",
			Level:      plugin_manager.WARNLevel,
			Normal:     status == NormalStatus,
			Tags:       map[string]string{"component": component.Name},
		})

		imageGVSList = append(imageGVSList, &metric_manager.GaugeVecSet{
			Labels: []string{clusterId, clusterbiz, component.Namespace, component.Name, component.Resource, ci.container, ci.image, status, rel},
			Value:  1,
		})
	}

	// 检查workload配置
	status, err = GetWorkLoadConfigStatus(component.Resource, component.Name, component.Namespace, clientSet)
	if err != nil {
		checkItemList = append(checkItemList, plugin_manager.CheckItem{
			ItemName:   SystemAppConfigCheckItem,
			ItemTarget: component.Name,
			Status:     status,
			Normal:     status == NormalStatus,
			Detail:     err.Error(),
			Tags:       map[string]string{"component": component.Name},
			Level:      plugin_manager.WARNLevel,
		})

		configGVSList = append(configGVSList, &metric_manager.GaugeVecSet{
			Labels: []string{clusterId, clusterbiz, component.Namespace, component.Name, component.Resource, status},
			Value:  1,
		})
	}

	return checkItemList, statusGVSList, imageGVSList, configGVSList, "", err
}

// getWorkLoadStatus xxx
func getWorkLoadStatus(workloadType string, workloadName string, namespace string, clientSet *kubernetes.Clientset, metricsClient *metricsclientset.Clientset) (string, []containerImage, error) {
	var containerImageList []containerImage
	var status = NormalStatus
	var err error
	ctx := internalUtil.GetCtx(10 * time.Second)
	switch workloadType {
	case deployment:
		var deploy *v1.Deployment
		deploy, err = clientSet.AppsV1().Deployments(namespace).Get(ctx, workloadName, metav1.GetOptions{
			ResourceVersion: "0",
		})

		if err != nil {
			status = AppErrorStatus
			if strings.Contains(err.Error(), "not found") {
				status = APPNotfoundAppStatus
			}
		} else {
			containerImageList, status, err = getDeployCheckResult(deploy, metricsClient)
		}

	case daemonset:
		var ds *v1.DaemonSet
		ds, err = clientSet.AppsV1().DaemonSets(namespace).Get(ctx, workloadName, metav1.GetOptions{
			ResourceVersion: "0",
		})

		if err != nil {
			status = AppErrorStatus
			if strings.Contains(err.Error(), "not found") {
				status = APPNotfoundAppStatus
			}
		} else {
			containerImageList, status, err = getDSCheckResult(ds, metricsClient)
		}

	case statefulset:
		var sts *v1.StatefulSet
		sts, err = clientSet.AppsV1().StatefulSets(namespace).Get(ctx, workloadName, metav1.GetOptions{
			ResourceVersion: "0",
		})

		if err != nil {
			status = AppErrorStatus
			if strings.Contains(err.Error(), "not found") {
				status = APPNotfoundAppStatus
			}
		} else {
			containerImageList, status, err = getSTSCheckResult(sts, metricsClient)
		}
	}

	if err != nil {
		return status, nil, fmt.Errorf("namespace: %s worload: %s GetWorkLoad failed: %s", namespace, workloadName, err.Error())
	}

	return status, containerImageList, nil

}

// GetWorkLoadConfigStatus xxx
func GetWorkLoadConfigStatus(workloadType, workloadName, namespace string, clientSet *kubernetes.Clientset) (string, error) {
	var err error
	var status string
	ctx := internalUtil.GetCtx(10 * time.Second)
	switch workloadType {
	case deployment:
		var deploy *v1.Deployment
		deploy, err = clientSet.AppsV1().Deployments(namespace).Get(ctx, workloadName, metav1.GetOptions{
			ResourceVersion: "0",
		})

		if err == nil {
			klog.V(9.).Infof("%s is ok", deploy.Name)
		}

	case daemonset:
		var ds *v1.DaemonSet
		ds, err = clientSet.AppsV1().DaemonSets(namespace).Get(ctx, workloadName, metav1.GetOptions{
			ResourceVersion: "0",
		})
		if err == nil {
			klog.V(9.).Infof("%s is ok", ds.Name)
		}

	case statefulset:
		var sts *v1.StatefulSet
		sts, err = clientSet.AppsV1().StatefulSets(namespace).Get(ctx, workloadName, metav1.GetOptions{
			ResourceVersion: "0",
		})
		if err == nil {
			klog.V(9.).Infof("%s is ok", sts.Name)
		}

	}

	if err != nil {
		if status != "" {
			return status, err
		}
		if strings.Contains(err.Error(), "not found") {
			return ConfigNotFoundStatus, err
		}
		return ConfigOtherErrorStatus, err
	}

	return NormalStatus, nil
}

// GetComponentFromManifest xxx
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

// CheckComponents check specified component status
func (p *Plugin) CheckComponents(
	clusterId string, clusterbiz string, clientSet *kubernetes.Clientset, metricsClient *metricsclientset.Clientset, componentList []Component) (
	[]*metric_manager.GaugeVecSet, []*metric_manager.GaugeVecSet, error) {
	imageGaugeVecSetList := make([]*metric_manager.GaugeVecSet, 0, 0)
	statusGaugeVecSetList := make([]*metric_manager.GaugeVecSet, 0, 0)
	ctx := internalUtil.GetCtx(10 * time.Second)
	for _, component := range componentList {
		var workload runtime.Object
		switch component.Resource {
		case deployment:
			deploy, err := clientSet.AppsV1().Deployments(component.Namespace).Get(ctx, component.Name,
				metav1.GetOptions{ResourceVersion: "0"})

			if err != nil {
				klog.Errorf("%s get %s %s %s failed: %s",
					clusterId, component.Namespace, component.Resource, component.Name, err.Error())
				continue
			}
			deploy.TypeMeta.Kind = deployment
			workload = deploy.DeepCopyObject()

		case daemonset:
			ds, err := clientSet.AppsV1().DaemonSets(component.Namespace).Get(ctx, component.Name,
				metav1.GetOptions{ResourceVersion: "0"})
			if err != nil {
				klog.Errorf("%s get %s %s %s failed: %s",
					clusterId, component.Namespace, component.Resource, component.Name, err.Error())
				continue
			}
			ds.TypeMeta.Kind = daemonset
			workload = ds.DeepCopyObject()

		case statefulset:
			sts, err := clientSet.AppsV1().StatefulSets(component.Namespace).Get(ctx, component.Name,
				metav1.GetOptions{ResourceVersion: "0"})
			if err != nil {
				klog.Errorf("%s get %s %s %s failed: %s",
					clusterId, component.Namespace, component.Resource, component.Name, err.Error())
				continue
			}
			sts.TypeMeta.Kind = statefulset
			workload = sts.DeepCopyObject()
		default:
			klog.Errorf("%s %s type is %s", clusterId, component.Name, component.Resource)
		}
		if workload == nil {
			klog.Infof("%s Unknown resource type: %s", clusterId, component.Resource)
			continue
		}

		// unstrObj, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(workload)
		containerImageGaugeVecSetList, statusGaugeVecSet := GetComponentGaugeVecSet(clusterId, clusterbiz, workload, release.Release{
			Namespace: component.Namespace,
			Name:      "nonchart",
		}, metricsClient)
		if len(containerImageGaugeVecSetList) == 0 {
			containerImageGaugeVecSetList = append(containerImageGaugeVecSetList, &metric_manager.GaugeVecSet{
				Labels: []string{
					clusterId, clusterbiz, component.Namespace, "nonchart", component.Name, workload.GetObjectKind().
						GroupVersionKind().Kind,
					"", ""}, Value: 1})
		}
		if statusGaugeVecSet == nil {
			statusGaugeVecSet = &metric_manager.GaugeVecSet{
				Labels: []string{
					clusterId, clusterbiz, component.Namespace, "nonchart", component.Name, workload.GetObjectKind().
						GroupVersionKind().Kind, "unknown"}, Value: 1}
		}
		imageGaugeVecSetList = append(imageGaugeVecSetList, containerImageGaugeVecSetList...)
		statusGaugeVecSetList = append(statusGaugeVecSetList, statusGaugeVecSet)
	}

	return imageGaugeVecSetList, statusGaugeVecSetList, nil
}

type containerImage struct {
	container string
	image     string
}

// getSTSCheckResult xxx
func getSTSCheckResult(sts *v1.StatefulSet, ms *metricsclientset.Clientset) ([]containerImage, string, error) {
	containerImageList := make([]containerImage, 0, 0)
	for _, container := range sts.Spec.Template.Spec.Containers {
		containerImageList = append(containerImageList, containerImage{
			container: container.Name,
			image:     container.Image,
		})
	}

	status := GetStatus(
		int(sts.Status.UpdatedReplicas),
		int(sts.Status.ReadyReplicas),
		int(sts.Status.Replicas))

	if status != NormalStatus {
		return containerImageList, status, nil
	}

	status, err := CheckPodMetric(sts.Spec.Template.Spec.Containers, ms, sts.Namespace, sts.Spec.Selector.MatchLabels)
	return containerImageList, status, err
}

// getDSCheckResult xxx
func getDSCheckResult(ds *v1.DaemonSet, ms *metricsclientset.Clientset) ([]containerImage, string, error) {
	containerImageList := make([]containerImage, 0, 0)
	for _, container := range ds.Spec.Template.Spec.Containers {
		containerImageList = append(containerImageList, containerImage{
			container: container.Name,
			image:     container.Image,
		})
	}

	status := GetStatus(
		int(ds.Status.UpdatedNumberScheduled),
		int(ds.Status.NumberReady),
		int(ds.Status.DesiredNumberScheduled))
	if status != NormalStatus {
		return containerImageList, status, nil
	}

	// ds不检测资源使用情况
	//status, err := CheckPodMetric(ds.Spec.Template.Spec.Containers, ms, ds.Namespace, ds.Spec.Selector.MatchLabels)
	return containerImageList, status, nil
}

// getDeployCheckResult xxx
func getDeployCheckResult(deploy *v1.Deployment, ms *metricsclientset.Clientset) ([]containerImage, string, error) {
	containerImageList := make([]containerImage, 0, 0)
	for _, container := range deploy.Spec.Template.Spec.Containers {
		containerImageList = append(containerImageList, containerImage{
			container: container.Name,
			image:     container.Image,
		})
	}

	status := GetStatus(
		int(deploy.Status.UpdatedReplicas),
		int(deploy.Status.ReadyReplicas),
		int(deploy.Status.Replicas))
	if status != NormalStatus {
		return containerImageList, status, nil
	}

	status, err := CheckPodMetric(deploy.Spec.Template.Spec.Containers, ms, deploy.Namespace, deploy.Spec.Selector.MatchLabels)
	return containerImageList, status, err
}

// CheckPodMetric xxx
func CheckPodMetric(containerList []corev1.Container, ms *metricsclientset.Clientset, namespace string, matchLabels map[string]string) (string, error) {
	podMetricList, err := ms.MetricsV1beta1().PodMetricses(namespace).List(internalUtil.GetCtx(15*time.Second),
		metav1.ListOptions{
			ResourceVersion: "0",
			LabelSelector:   metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: matchLabels})})

	if err != nil {
		if strings.Contains(err.Error(), "the server could not find the requested resource") {
			klog.Infof(err.Error())
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

// GetResourceGaugeVecSet generate GaugeVecSet from workload status
func GetComponentGaugeVecSet(
	clusterId string, clusterbiz string, object runtime.Object, rel release.Release, cs *metricsclientset.Clientset) ([]*metric_manager.GaugeVecSet, *metric_manager.GaugeVecSet) {

	imageGaugeVecSetList := make([]*metric_manager.GaugeVecSet, 0, 0)
	var statusGaugeVecSet *metric_manager.GaugeVecSet
	objectMap, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	unstr := &unstructured.Unstructured{Object: objectMap}

	kind := unstr.GetKind()
	var containers []corev1.Container
	namespace := ""
	var labels map[string]string
	switch kind {
	case deployment:
		deploy := &v1.Deployment{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(objectMap, deploy)
		if err != nil {
			klog.Errorf("DefaultUnstructuredConverter failed: %s", err.Error())
			return nil, nil
		}

		for _, container := range deploy.Spec.Template.Spec.Containers {
			imageGaugeVecSetList = append(imageGaugeVecSetList, &metric_manager.GaugeVecSet{
				Labels: []string{clusterId, clusterbiz, rel.Namespace, rel.Name, deploy.Name, kind,
					container.Name, container.Image}, Value: 1})
		}

		status := GetStatus(
			int(deploy.Status.UpdatedReplicas),
			int(deploy.Status.ReadyReplicas),
			int(deploy.Status.Replicas))

		statusGaugeVecSet = &metric_manager.GaugeVecSet{
			Labels: []string{clusterId, clusterbiz, rel.Namespace, rel.Name, deploy.Name, kind,
				status}, Value: 1}

		containers = deploy.Spec.Template.Spec.Containers
		namespace = deploy.Namespace
		labels = deploy.Spec.Selector.MatchLabels

		break
	case statefulset:
		sts := &v1.StatefulSet{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(objectMap, sts)
		if err != nil {
			klog.Errorf("DefaultUnstructuredConverter failed: %s", err.Error())
			return nil, nil
		}
		for _, container := range sts.Spec.Template.Spec.Containers {
			imageGaugeVecSetList = append(imageGaugeVecSetList, &metric_manager.GaugeVecSet{
				Labels: []string{clusterId, clusterbiz, rel.Namespace, rel.Name, sts.Name, kind,
					container.Name, container.Image}, Value: 1})
		}
		statusGaugeVecSet = &metric_manager.GaugeVecSet{
			Labels: []string{clusterId, clusterbiz, rel.Namespace, rel.Name, sts.Name, kind,
				GetStatus(
					int(sts.Status.UpdatedReplicas),
					int(sts.Status.ReadyReplicas),
					int(sts.Status.Replicas))}, Value: 1}
		containers = sts.Spec.Template.Spec.Containers
		namespace = sts.Namespace
		labels = sts.Spec.Selector.MatchLabels
		break
	case daemonset:
		ds := &v1.DaemonSet{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(objectMap, ds)
		if err != nil {
			klog.Errorf("DefaultUnstructuredConverter failed: %s", err.Error())
			return nil, nil
		}
		for _, container := range ds.Spec.Template.Spec.Containers {
			imageGaugeVecSetList = append(imageGaugeVecSetList, &metric_manager.GaugeVecSet{
				Labels: []string{clusterId, clusterbiz, rel.Namespace, rel.Name, ds.Name, kind,
					container.Name, container.Image}, Value: 1})
		}

		statusGaugeVecSet = &metric_manager.GaugeVecSet{
			Labels: []string{clusterId, clusterbiz, rel.Namespace, rel.Name, ds.Name, kind,
				GetStatus(
					int(ds.Status.UpdatedNumberScheduled),
					int(ds.Status.NumberReady),
					int(ds.Status.DesiredNumberScheduled))}, Value: 1}
		containers = ds.Spec.Template.Spec.Containers
		namespace = ds.Namespace
		labels = ds.Spec.Selector.MatchLabels
		break
	default:
		klog.Errorf("%s %s type is %s", clusterId, unstr.GetName(), kind)
		return nil, nil
	}

	if statusGaugeVecSet.Labels[0] == NormalStatus {
		status, err := CheckPodMetric(containers, cs, namespace, labels)
		if err != nil {
			klog.Errorf("%s get podmetric failed: %s", clusterId, err.Error())

			statusGaugeVecSet = &metric_manager.GaugeVecSet{
				Labels: []string{clusterId, clusterbiz, rel.Namespace, rel.Name, objectMap["Name"].(string), kind,
					status}, Value: 1}
		}
	}

	return imageGaugeVecSetList, statusGaugeVecSet
}

// Ready return true if cluster check is over
func (p *Plugin) Ready(clusterID string) bool {
	p.WriteLock.Lock()
	defer p.WriteLock.Unlock()
	return p.ReadyMap[clusterID]
}

// GetResult return check result by cluster ID
func (p *Plugin) GetResult(s string) plugin_manager.CheckResult {
	return p.Result[s]
}

// GetImageStatus xxx
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
