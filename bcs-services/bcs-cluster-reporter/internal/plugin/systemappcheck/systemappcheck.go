/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package systempodcheck

import (
	"context"
	"fmt"
	"os"
	"regexp"
	goruntime "runtime"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metric_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"

	"github.com/containerd/containerd/pkg/cri/util"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

// Plugin xxx
type Plugin struct {
	stopChan  chan int
	opt       *Options
	checkLock sync.Mutex
}

var (
	systemAppChartVersion = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "system_app_chart_version",
		Help: "system_app_chart_version, 1 means deployed",
	}, []string{"target", "target_biz", "namespace", "chart", "version", "status"})
	systemAppImageVersion = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "system_app_image_version",
		Help: "system_app_image_version, 1 means ok",
	}, []string{"target", "target_biz", "namespace", "chart", "component", "resource", "container", "version", "status"})
	systemAppChartMap = make(map[string]*prometheus.GaugeVec)
	systemAppImageMap = make(map[string]*prometheus.GaugeVec)
	systemAppMapLock  sync.Mutex
)

func init() {
	metric_manager.Register(systemAppChartVersion)
	metric_manager.Register(systemAppImageVersion)
}

// Setup xxx
func (p *Plugin) Setup(configFilePath string) error {
	configFileBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("read systemappcheck config file %s failed, err %s", configFilePath, err.Error())
	}
	p.opt = &Options{}
	if err = json.Unmarshal(configFileBytes, p.opt); err != nil {
		if err = yaml.Unmarshal(configFileBytes, p.opt); err != nil {
			return fmt.Errorf("decode systemappcheck config file %s failed, err %s", configFilePath, err.Error())
		}
	}
	if err = p.opt.Validate(); err != nil {
		return err
	}

	interval := p.opt.Interval
	if interval == 0 {
		interval = 60
	}

	go func() {
		for {
			if p.checkLock.TryLock() {
				p.checkLock.Unlock()
				plugin_manager.Pm.Lock()
				go p.Check()
			} else {
				klog.V(3).Infof("the former systemappcheck didn't over, skip in this loop")
			}
			select {
			case result := <-p.stopChan:
				klog.V(3).Infof("stop plugin %s by signal %d", p.Name(), result)
				return
			case <-time.After(time.Duration(interval) * time.Second):
				continue
			}
		}
	}()

	return nil
}

// Stop xxx
func (p *Plugin) Stop() error {
	p.stopChan <- 1
	klog.Infof("plugin %s stopped", p.Name())
	return nil
}

// Name xxx
func (p *Plugin) Name() string {
	return "systemappcheck"
}

// Check xxx
//
//	how to check resource version without chart
//
// informer 改造
//
//	内存使用优化
func (p *Plugin) Check() {
	start := time.Now()
	p.checkLock.Lock()
	klog.Infof("start %s", p.Name())
	defer func() {
		klog.Infof("end %s", p.Name())
		plugin_manager.Pm.UnLock()
		p.checkLock.Unlock()
		metric_manager.SetCommonDurationMetric([]string{"systemappcheck", "", "", ""}, start)
	}()

	wg := sync.WaitGroup{}
	chartGaugeVecSetList := make([]*metric_manager.GaugeVecSet, 0, 0)
	imageGaugeVecSetList := make([]*metric_manager.GaugeVecSet, 0, 0)
	for _, cluster := range plugin_manager.Pm.GetConfig().ClusterConfigs {
		wg.Add(1)
		clusterId := cluster.ClusterID
		clusterbiz := cluster.BusinessID
		config := cluster.Config
		plugin_manager.Pm.Add()
		go func() {
			defer func() {
				klog.V(9).Infof("end systemappcheck for %s", clusterId)
				wg.Done()
				plugin_manager.Pm.Done()
			}()

			klog.V(9).Infof("start systemappcheck for %s", clusterId)

			clientSet, err := k8s.GetClientsetByConfig(config)
			if err != nil {
				klog.Errorf("%s GetClientsetByClusterId failed: %s", clusterId, err.Error())
				return
			}
			_, err = k8s.GetK8sVersion(clientSet)
			if err != nil {
				klog.Errorf("%s GetK8sVersion failed: %s", clusterId, err.Error())
				return
			}

			getter := k8s.GetRestClientGetterByConfig(config)

			// 获取配置文件中的component列表(有可能不是helm方式部署的)
			var clusterComponent []Component
			err = util.DeepCopy(&clusterComponent, p.opt.Components)
			if err != nil {
				klog.Errorf("%s DeepCopy failed: %s", clusterId, err.Error())
				return
			}

			// Get releases
			// 默认获取以下这些namespace中的应用
			namespaceList := []string{
				"kube-system", "bcs-system", "bk-system", "default", "bkmonitor-operator"}
			chartCheckResult := make([]*metric_manager.GaugeVecSet, 0, 0)
			for _, namespace := range namespaceList {
				actionConfig := new(action.Configuration)
				if err := actionConfig.Init(getter, namespace, os.Getenv("HELM_DRIVER"), klog.Infof); err != nil {
					klog.Errorf("Config helm client failed: %s", err.Error())
					return
				}

				// 获取release列表
				client := action.NewList(actionConfig)
				client.Deployed = true
				client.AllNamespaces = true
				relList, err := client.Run()
				if err != nil {
					klog.Errorf("%s helm get deployed chart failed: %s", clusterId, err.Error())
					return
				}

				for _, rel := range relList {
					// 生成对应的metric配置

					chartCheckResult = append(chartCheckResult, &metric_manager.GaugeVecSet{
						Labels: []string{
							clusterId,
							clusterbiz,
							rel.Namespace,
							rel.Name,
							rel.Chart.AppVersion(),
							rel.Info.Status.String()},
						Value: 1})

					manifest := rel.Manifest
					// exclude non-workload resource manifest
					resourceManifestList := strings.Split(manifest, "---")
					workLoadManifestList := make([]string, 0, 0)

					re, _ := regexp.Compile("\nkind: Deployment|\nkind: DaemonSet|\nkind: StatefulSet")
					for _, resourceManifest := range resourceManifestList {
						if re.MatchString(resourceManifest) {
							workLoadManifestList = append(workLoadManifestList, resourceManifest)
						}
					}

					if len(workLoadManifestList) == 0 {
						klog.V(9).Infof("%s %s manifest workload is nil", clusterId, rel.Name)
						continue
					}

					// 查询对应的资源对象
					for _, worloadManifest := range workLoadManifestList {
						var objMap map[string]interface{}
						if err := yaml.Unmarshal([]byte(worloadManifest), &objMap); err != nil {
							fmt.Printf("Error unmarshalling YAML: %v\n", err)
							continue
						}
						if _, ok := objMap["kind"]; !ok {
							klog.Errorf("%s wrong workload yaml, no kind, rel: %s, workload: %s", clusterId, rel.Name, worloadManifest)
							continue
						}
						if _, ok := objMap["metadata"]; !ok {
							klog.Errorf("%s wrong workload yaml, no name, rel: %s, workload: %s", clusterId, rel.Name, worloadManifest)
							continue
						}

						kind := objMap["kind"].(string)
						name := objMap["metadata"].(map[interface{}]interface{})["name"].(string)
						ctx, _ := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)

						var workload runtime.Object
						switch kind {
						case "Deployment":
							deploy, err := clientSet.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
							deploy.TypeMeta.Kind = "Deployment"
							if err != nil {
								klog.Errorf("%s deployment %s not found in namespace %s, release %s", clusterId, name, namespace, rel.Name)
								continue
							}
							workload = deploy.DeepCopyObject()
						case "DaemonSet":
							ds, err := clientSet.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
							ds.TypeMeta.Kind = "DaemonSet"
							if err != nil {
								klog.Errorf("%s daemonset %s not found in namespace %s, release %s", clusterId, name, namespace, rel.Name)
								continue
							}
							workload = ds.DeepCopyObject()
						case "StatefulSet":
							sts, err := clientSet.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
							sts.TypeMeta.Kind = "StatefulSet"
							if err != nil {
								klog.Errorf("%s statefulset %s not found in namespace %s, release %s", clusterId, name, namespace, rel.Name)
								continue
							}
							workload = sts.DeepCopyObject()
						}

						if workload == nil {
							klog.Infof("Unknown resource type: %s", kind)
							continue
						}

						// 如果component列表中的应用已经通过helm部署了，则不再需要单独确认
						for index, component := range clusterComponent {
							if component.Name == name &&
								component.Namespace == namespace &&
								strings.ToLower(component.Resource) ==
									strings.ToLower(kind) {
								clusterComponent = append(clusterComponent[:index], clusterComponent[index+1:]...)
							}
						}

						resourceGaugeVecSet := GetResourceGaugeVecSet(clusterId, clusterbiz, workload, *rel)
						if resourceGaugeVecSet == nil {
							resourceGaugeVecSet = &metric_manager.GaugeVecSet{
								Labels: []string{
									clusterId,
									clusterbiz,
									rel.Namespace,
									rel.Name,
									name,
									kind,
									"", "",
									"notready"}, Value: 1}
						}
						imageGaugeVecSetList = append(imageGaugeVecSetList, resourceGaugeVecSet)
					}
				}
			}

			chartGaugeVecSetList = append(chartGaugeVecSetList, chartCheckResult...)

			// get result of CheckComponents
			componentImageGaugeVecSetList, _ := p.CheckComponents(clusterId, clusterbiz, clientSet, clusterComponent)
			imageGaugeVecSetList = append(imageGaugeVecSetList, componentImageGaugeVecSetList...)

			// 集群单独路径的指标配置
			systemAppMapLock.Lock()
			if _, ok := systemAppChartMap[clusterId]; !ok {
				systemAppChartMap[clusterId] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Name: "system_app_chart_version",
					Help: "system_app_chart_version, 1 means deployed",
				}, []string{"target", "target_biz", "namespace", "chart", "version", "status"})
				metric_manager.MM.RegisterSeperatedMetric(clusterId, systemAppChartMap[clusterId])
			}

			if _, ok := systemAppImageMap[clusterId]; !ok {
				systemAppImageMap[clusterId] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Name: "system_app_image_version",
					Help: "system_app_image_version, 1 means ok",
				}, []string{"target", "target_biz", "namespace", "chart", "component", "resource", "container", "version",
					"status"})
				metric_manager.MM.RegisterSeperatedMetric(clusterId, systemAppImageMap[clusterId])
			}

			metric_manager.SetMetric(systemAppChartMap[clusterId], chartCheckResult)
			metric_manager.SetMetric(systemAppImageMap[clusterId], componentImageGaugeVecSetList)
			systemAppMapLock.Unlock()
		}()
	}

	wg.Wait()
	// reset metric value
	metric_manager.SetMetric(systemAppChartVersion, chartGaugeVecSetList)
	metric_manager.SetMetric(systemAppImageVersion, imageGaugeVecSetList)
	goruntime.GC()
}

// GetStatus
func GetStatus(updatedReplicas int, availableReplicas int, replicas int) string {
	if replicas > 0 && availableReplicas == replicas && updatedReplicas == replicas {
		return "ready"
	} else {
		return "notready"
	}

}

// CheckComponents check specified component status
func (p *Plugin) CheckComponents(
	clusterId string, clusterbiz string, clientSet *kubernetes.Clientset, componentList []Component) (
	[]*metric_manager.GaugeVecSet, error) {
	imageGaugeVecSetList := make([]*metric_manager.GaugeVecSet, 0, 0)
	for _, component := range componentList {
		var workload runtime.Object
		var err error
		switch component.Resource {
		case "Deployment", "deployment":
			workload, err = clientSet.AppsV1().Deployments(component.Namespace).Get(context.Background(), component.Name,
				metav1.GetOptions{})
			if err != nil {
				klog.Errorf("%s get %s %s %s failed: %s",
					clusterId, component.Namespace, component.Resource, component.Name, err.Error())
				continue
			}

		case "DaemonSet", "daemonSet":
			workload, err = clientSet.AppsV1().DaemonSets(component.Namespace).Get(context.Background(), component.Name,
				metav1.GetOptions{})
			if err != nil {
				klog.Errorf("%s get %s %s %s failed: %s",
					clusterId, component.Namespace, component.Resource, component.Name, err.Error())
				continue
			}
		case "StatefulSet", "statefulset":
			workload, err = clientSet.AppsV1().StatefulSets(component.Namespace).Get(context.Background(), component.Name,
				metav1.GetOptions{})
			if err != nil {
				klog.Errorf("%s get %s %s %s failed: %s",
					clusterId, component.Namespace, component.Resource, component.Name, err.Error())
				continue
			}
		}
		if workload == nil {
			klog.Infof("%s Unknown resource type: %s", clusterId, component.Resource)
			continue
		}

		// unstrObj, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(workload)
		resourceGaugeVecSet := GetResourceGaugeVecSet(clusterId, clusterbiz, workload, release.Release{
			Namespace: component.Namespace,
			Name:      "nonchart",
		})
		if resourceGaugeVecSet == nil {
			resourceGaugeVecSet = &metric_manager.GaugeVecSet{
				Labels: []string{
					clusterId, clusterbiz, component.Namespace, "nonchart", component.Name, workload.GetObjectKind().
						GroupVersionKind().Kind,
					"", "",
					"notready"}, Value: 1}
		}
		imageGaugeVecSetList = append(imageGaugeVecSetList, resourceGaugeVecSet)
	}

	return imageGaugeVecSetList, nil
}

// GetResourceGaugeVecSet generate GaugeVecSet from workload status
func GetResourceGaugeVecSet(
	clusterId string, clusterbiz string, object runtime.Object, rel release.Release) *metric_manager.GaugeVecSet {

	var resourceGaugeVecSet *metric_manager.GaugeVecSet
	// unstr, ok := object.(*unstructured.Unstructured)
	// if !ok {
	//	klog.Errorf("attempt to decode non-Unstructured object: %s", object)
	//	return nil
	// }
	objectMap, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	unstr := &unstructured.Unstructured{Object: objectMap}

	kind := unstr.GetKind()
	klog.Infof(kind)
	switch strings.ToLower(kind) {
	case "deployment":
		deploy := &v1.Deployment{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(objectMap, deploy)
		if err != nil {
			klog.Errorf("DefaultUnstructuredConverter failed: %s", err.Error())
			return nil
		}

		for _, container := range deploy.Spec.Template.Spec.Containers {
			resourceGaugeVecSet = &metric_manager.GaugeVecSet{
				Labels: []string{clusterId, clusterbiz, rel.Namespace, rel.Name, deploy.Name, kind,
					container.Name, container.Image,
					GetStatus(
						int(deploy.Status.UpdatedReplicas),
						int(deploy.Status.ReadyReplicas),
						int(deploy.Status.Replicas))}, Value: 1}
		}
		break
	case "statefulset":
		sts := &v1.StatefulSet{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(objectMap, sts)
		if err != nil {
			klog.Errorf("DefaultUnstructuredConverter failed: %s", err.Error())
			return nil
		}
		for _, container := range sts.Spec.Template.Spec.Containers {
			resourceGaugeVecSet = &metric_manager.GaugeVecSet{
				Labels: []string{clusterId, clusterbiz, rel.Namespace, rel.Name, sts.Name, kind,
					container.Name, container.Image,
					GetStatus(
						int(sts.Status.UpdatedReplicas),
						int(sts.Status.ReadyReplicas),
						int(sts.Status.Replicas))}, Value: 1}
		}
		break
	case "daemonset":
		ds := &v1.DaemonSet{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(objectMap, ds)
		if err != nil {
			klog.Errorf("DefaultUnstructuredConverter failed: %s", err.Error())
			return nil
		}
		for _, container := range ds.Spec.Template.Spec.Containers {
			resourceGaugeVecSet = &metric_manager.GaugeVecSet{
				Labels: []string{clusterId, clusterbiz, rel.Namespace, rel.Name, ds.Name, kind,
					container.Name, container.Image,
					GetStatus(
						int(ds.Status.UpdatedNumberScheduled),
						int(ds.Status.NumberReady),
						int(ds.Status.DesiredNumberScheduled))}, Value: 1}
		}
		break
	default:
		klog.V(6).Infof("%s type is %s", unstr.GetName(), kind)
	}
	return resourceGaugeVecSet
}
