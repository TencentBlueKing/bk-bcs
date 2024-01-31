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

// Package clustercheck xxx
package clustercheck

import (
	"context"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metric_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"

	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog"
)

// Plugin xxx
type Plugin struct {
	stopChan       chan int
	opt            *Options
	checkLock      sync.Mutex
	testYamlString string
}

var (
	// NewGaugeVec creates a new GaugeVec based on the provided GaugeOpts and
	// partitioned by the given label names.
	clusterAvailability = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cluster_availability",
		Help: "cluster_availability, 1 means OK",
	}, []string{"target", "target_biz", "status"})
	// NewGaugeVec creates a new GaugeVec based on the provided GaugeOpts and
	// partitioned by the given label names.
	clusterCheckDuration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cluster_check_duration_seconds",
		Help: "cluster_check_duration_seconds, 1 means OK",
	}, []string{"target", "target_biz", "step"})
	clusterAvailabilityMap     = make(map[string][]*prometheus.GaugeVec)
	unstructuredObj            = &unstructured.Unstructured{}
	clusterAvailabilityMapLock sync.Mutex
)

func init() {
	metric_manager.Register(clusterAvailability)
	metric_manager.Register(clusterCheckDuration)
}

// Setup xxx
func (p *Plugin) Setup(configFilePath string) error {
	configFileBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("read clustercheck config file %s failed, err %s", configFilePath, err.Error())
	}
	p.opt = &Options{}
	if err = json.Unmarshal(configFileBytes, p.opt); err != nil {
		if err = yaml.Unmarshal(configFileBytes, p.opt); err != nil {
			return fmt.Errorf("decode clustercheck config file %s failed, err %s", configFilePath, err.Error())
		}
	}

	if err = p.opt.Validate(); err != nil {
		return err
	}

	unstructuredObj.SetUnstructuredContent(p.opt.TestYaml)

	// 给测试workload添加标签
	kind := unstructuredObj.GetKind()
	switch strings.ToLower(kind) {
	case "replicaset":
		fallthrough
	case "deployment":
		fallthrough
	case "job":
		objectMap := unstructuredObj.Object
		updateNestedMap(objectMap, []string{"spec", "template", "metadata", "labels", "bcs-cluster-reporter"},
			"bcs-cluster-reporter")
		//updateNestedMap(objectMap, []string{"spec", "selector", "matchLabels", "bcs-cluster-reporter"},
		//	"bcs-cluster-reporter")
		//updateNestedMap(objectMap, []string{"spec", "selector", "matchLabels", "bcs-cluster-reporter"},
		//	"bcs-cluster-reporter")
		//klog.Info(objectMap)
		unstructuredObj.SetUnstructuredContent(objectMap)
	default:
		klog.Fatalf("workload %s type is %s, not supported, please use job, deployment, replicaset",
			unstructuredObj.GetName(), kind)
	}

	interval := p.opt.Interval
	if interval == 0 {
		interval = 60
	}

	go func() {
		for {
			if p.checkLock.TryLock() {
				p.checkLock.Unlock()
				if p.opt.Synchronization {
					plugin_manager.Pm.Lock()
				}
				go p.Check()
			} else {
				klog.V(3).Infof("the former clustercheck didn't over, skip in this loop")
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
	return "clustercheck"
}

func int64Ptr(i int64) *int64 { return &i }

// Check xxx
func (p *Plugin) Check() {
	start := time.Now()
	p.checkLock.Lock()
	klog.Infof("start %s", p.Name())
	defer func() {
		klog.Infof("end %s", p.Name())
		if p.opt.Synchronization {
			plugin_manager.Pm.UnLock()
		}
		p.checkLock.Unlock()
		metric_manager.SetCommonDurationMetric([]string{"clustercheck", "", "", ""}, start)
	}()

	// 根据internal来调整超时时间的长短
	interval := p.opt.Interval

	wg := sync.WaitGroup{}
	clusterChecktGaugeVecSetList := make([]*metric_manager.GaugeVecSet, 0, 0)
	clusterCheckDurationGaugeVecSetList := make([]*metric_manager.GaugeVecSet, 0, 0)
	for _, cluster := range plugin_manager.Pm.GetConfig().ClusterConfigs {
		wg.Add(1)
		config := cluster.Config
		clusterId := cluster.ClusterID
		clusterbiz := cluster.BusinessID
		plugin_manager.Pm.Add()
		go func() {
			status := "error"
			defer func() {
				if r := recover(); r != nil {
					klog.Errorf("%s clustercheck failed: %s, stack: %v\n", clusterId, r, string(debug.Stack()))
					// GetClientsetByConfig cluster k8s client
					clientSet1, _ := k8s.GetClientsetByConfig(config)
					var responseContentType string
					body, _ := clientSet1.RESTClient().Get().
						AbsPath("/apis").
						SetHeader("Accept", "application/json").
						Do(context.TODO()).
						ContentType(&responseContentType).
						Raw()
					klog.V(3).Infof("Try get apis for %s: %s", clusterId, string(body))
					status = "panic"
				}

				plugin_manager.Pm.Done()
				wg.Done()
			}()

			klog.Infof("start clustercheck for %s", clusterId)
			workloadToScheduleCost, workloadToPodCost, worloadToRunningCost, err := testClusterByCreateUnstructuredObj(
				unstructuredObj, config, &status, interval, clusterId)

			clusterChecktGaugeVecSetList = append(clusterChecktGaugeVecSetList,
				&metric_manager.GaugeVecSet{Labels: []string{clusterId, clusterbiz, status}, Value: float64(1)})
			clusterCheckDurationGaugeVecSetList = append(clusterCheckDurationGaugeVecSetList,
				&metric_manager.GaugeVecSet{Labels: []string{clusterId, clusterbiz, "create_pod"},
					Value: float64(workloadToPodCost) / 1000000000})
			clusterCheckDurationGaugeVecSetList = append(clusterCheckDurationGaugeVecSetList,
				&metric_manager.GaugeVecSet{Labels: []string{clusterId, clusterbiz, "schedule_pod"},
					Value: float64(workloadToScheduleCost) / 1000000000})
			clusterCheckDurationGaugeVecSetList = append(clusterCheckDurationGaugeVecSetList,
				&metric_manager.GaugeVecSet{Labels: []string{clusterId, clusterbiz, "start_pod"},
					Value: float64(worloadToRunningCost) / 1000000000})

			// 集群单独路径的指标配置
			clusterAvailabilityMapLock.Lock()
			if _, ok := clusterAvailabilityMap[clusterId]; !ok {
				clusterAvailabilityMap[clusterId] = make([]*prometheus.GaugeVec, 0, 0)
				clusterAvailabilityMap[clusterId] = append(clusterAvailabilityMap[clusterId],
					prometheus.NewGaugeVec(prometheus.GaugeOpts{
						Name: "cluster_availability",
						Help: "cluster_availability, 1 means OK",
					}, []string{"target", "target_biz", "status"}),
					prometheus.NewGaugeVec(prometheus.GaugeOpts{
						Name: "cluster_check_duration_seconds",
						Help: "cluster_check_duration_seconds",
					}, []string{"target", "target_biz", "step"}))

				for index := range clusterAvailabilityMap[clusterId] {
					metric_manager.MM.RegisterSeperatedMetric(clusterId, clusterAvailabilityMap[clusterId][index])
				}
			}
			clusterAvailabilityMapLock.Unlock()

			metric_manager.SetMetric(clusterAvailabilityMap[clusterId][0], []*metric_manager.GaugeVecSet{
				&metric_manager.GaugeVecSet{Labels: []string{clusterId, clusterbiz, status}, Value: float64(1)},
			})
			metric_manager.SetMetric(clusterAvailabilityMap[clusterId][1], []*metric_manager.GaugeVecSet{
				&metric_manager.GaugeVecSet{
					Labels: []string{clusterId, clusterbiz, "create_pod"}, Value: float64(workloadToPodCost) / 1000000000},
				&metric_manager.GaugeVecSet{
					Labels: []string{clusterId, clusterbiz, "schedule_pod"}, Value: float64(workloadToScheduleCost) / 1000000000},
				&metric_manager.GaugeVecSet{
					Labels: []string{clusterId, clusterbiz, "start_pod"}, Value: float64(worloadToRunningCost) / 1000000000},
			})

			if err != nil {
				klog.Errorf("%s testClusterByCreateUnstructuredObj failed: %s", clusterId, err.Error())
			}
			klog.Infof("end clustercheck for %s", clusterId)
			klog.V(6).Infof("%s clustercheck result %s", clusterId, status)
		}()
	}
	wg.Wait()
	metric_manager.SetMetric(clusterAvailability, clusterChecktGaugeVecSetList)
	metric_manager.SetMetric(clusterCheckDuration, clusterCheckDurationGaugeVecSetList)

	// 去掉已经不存在的集群的指标
	for clusterId := range clusterAvailabilityMap {
		deleted := true
		for _, cluster := range plugin_manager.Pm.GetConfig().ClusterConfigs {
			if clusterId == cluster.ClusterID {
				deleted = false
				break
			}
		}
		if deleted {
			delete(clusterAvailabilityMap, clusterId)
		}
	}
}

// test ClusterByCreateUnstructuredObj
func testClusterByCreateUnstructuredObj(unstructuredObj *unstructured.Unstructured, config *rest.Config, status *string,
	interval int, clusterID string) (
	workloadToScheduleCost, workloadToPodCost, worloadToRunningCost time.Duration, err error) {
	workloadToScheduleCost = time.Duration(0)
	workloadToPodCost = time.Duration(0)
	worloadToRunningCost = time.Duration(0)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(interval/6)*time.Second)
	defer cancelFunc()
	namespace := unstructuredObj.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	clientSet, err := k8s.GetClientsetByConfig(config)
	if err != nil {
		*status = "配置失败" // nolint
		err = fmt.Errorf("GetClientsetByConfig failed: %s", err.Error())
		return
	}

	if clientSet == nil {
		*status = "配置失败" // nolint
		err = fmt.Errorf("Get clientSet failed %s", err.Error())
		return
	}

	clusterUnstructuredObj := unstructuredObj.DeepCopy()
	clusterGVK := clusterUnstructuredObj.GroupVersionKind()

	// 获取k8s集群version,确认集群是否可访问
	_, err = k8s.GetK8sVersion(clientSet)
	if err != nil {
		*status = "访问集群失败" // 访问集群失败
		err = fmt.Errorf("GetK8sVersion failed: %s", err.Error())
		return
	}

	// 确认test yaml的命名空间是否存在
	_, err = clientSet.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		*status = "命名空间不存在"
		_, createError := clientSet.CoreV1().Namespaces().Create(ctx, &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}, metav1.CreateOptions{})
		if createError != nil {
			klog.Errorf("create namespace failed: %s", createError.Error())
		}
		err = fmt.Errorf("get target resource namespace failed: %s", err.Error())
		return
	}

	// Returns copy of current discovery client that will only
	// receive the legacy discovery format, or pointer to current
	// discovery client if it does not support legacy-only discovery.
	discoveryInterface := clientSet.Discovery().WithLegacy()
	if discoveryInterface == nil {
		*status = "配置失败" // nolint
		err = fmt.Errorf("Get discoveryInterface failed %s", err.Error())
		return
	}
	// discoveryInterface.ServerGroupsAndResources()
	groupResource, err := restmapper.GetAPIGroupResources(discoveryInterface)
	if err != nil {
		*status = "配置失败" // nolint
		err = fmt.Errorf("GetAPIGroupResources failed %s", err.Error())
	}
	mapper := restmapper.NewDiscoveryRESTMapper(groupResource)
	mapping, err := mapper.RESTMapping(clusterGVK.GroupKind(), clusterGVK.Version)
	if err != nil {
		*status = "配置失败" // nolint
		err = fmt.Errorf("RESTMapping failed %s", err.Error())
		return
	}

	dynamicConfig, err := dynamic.NewForConfig(config)
	if err != nil {
		*status = "配置失败"
		err = fmt.Errorf("create dynamicConfig %s", err.Error())
		return
	}
	clusterUnstructuredObj.SetName("bcs-blackbox-job-" + time.Now().Format("150405"))

	dri := dynamicConfig.Resource(mapping.Resource).Namespace(namespace)
	defer func() {
		go func() {
			backgroundDeletion := metav1.DeletePropagationBackground

			podList, err := clientSet.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
				ResourceVersion: "0",
				LabelSelector:   "bcs-cluster-reporter=bcs-cluster-reporter",
			})
			if err != nil {
				klog.Errorf("%s get pod failed %s", clusterID, err.Error())
			} else {
				for _, pod := range podList.Items {
					if pod.Status.Phase != "Completed" && pod.Status.Phase != "Succeeded" && time.Now().
						Unix()-pod.CreationTimestamp.Unix() < 600 {
						continue
					}

					klog.Infof("%s start to delete targetPod %s", clusterID, pod.Name)
					err = clientSet.CoreV1().Pods(namespace).Delete(context.Background(), pod.Name, metav1.DeleteOptions{
						GracePeriodSeconds: int64Ptr(0),
					})
					if err != nil {
						klog.Errorf("%s delete pod failed %s", clusterID, err.Error())
					}
				}
			}

			jobList, err := clientSet.BatchV1().Jobs(namespace).List(context.Background(), metav1.ListOptions{
				ResourceVersion: "0",
				LabelSelector:   "bcs-cluster-reporter=bcs-cluster-reporter",
			})
			if err != nil {
				klog.Errorf("%s get job failed %s", clusterID, err.Error())
			} else {
				for _, job := range jobList.Items {
					klog.Infof("%s start to delete job %s", clusterID, job.Name)
					err = clientSet.BatchV1().Jobs(namespace).Delete(context.Background(), job.Name, metav1.DeleteOptions{
						GracePeriodSeconds: int64Ptr(5),
						PropagationPolicy:  &backgroundDeletion,
					})
					if err != nil {
						klog.Errorf("%s delete job failed %s", clusterID, err.Error())
					}
				}
			}

		}()
	}()

	*status, workloadToScheduleCost, workloadToPodCost, worloadToRunningCost, err =
		getWatchStatus(clientSet, clusterUnstructuredObj, dri, namespace, interval, clusterID, context.Background())

	return // nolint
}

// get Watch Status
// nolint
func getWatchStatus(clientSet *kubernetes.Clientset, clusterUnstructuredObj *unstructured.Unstructured,
	dri dynamic.ResourceInterface, namespace string, interval int, clusterID string, ctx context.Context) (status string,
	workloadToScheduleCost, workloadToPodCost, worloadToRunningCost time.Duration, err error) {
	// TimeoutSeconds: int64Ptr(500)
	watchInterface, err := clientSet.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{ResourceVersion: "0",
		LabelSelector: "bcs-cluster-reporter=bcs-cluster-reporter", TimeoutSeconds: int64Ptr(int64(interval / 6))})
	if err != nil {
		status = "watch失败"
		err = fmt.Errorf("%s watch failed %s", clusterID, err.Error())
		return
	}
	watchStartTime := time.Now()

	defer func() {
		go func() {
			if watchInterface != nil {
				watchInterface.Stop()
			}
		}()
	}()

	testObj, err := dri.Create(ctx, clusterUnstructuredObj, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Create failed %s", err.Error())
		if strings.Contains(err.Error(), "already exists") {
			time.Sleep(5 * time.Second)
			status = "workload已存在"
		} else {
			status = "创建workload失败"
		}
		return
	}

	createStartTime := testObj.GetCreationTimestamp().Time

	createPodFlag := false

	for {
		select {
		case e, ok := <-watchInterface.ResultChan():
			if !ok {
				watchInterface.Stop()
				if status != "ok" {
					status = "watch关闭"
					klog.Infof("%s restart watch", clusterID)
					// may miss some events here
					watchInterface, err = clientSet.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{ResourceVersion: "0",
						LabelSelector: "bcs-cluster-reporter=bcs-cluster-reporter", TimeoutSeconds: int64Ptr(int64(interval / 6))})
					if err != nil {
						status = "watch失败"
						err = fmt.Errorf("%s watch failed %s", clusterID, err.Error())
						return
					}
				}
			} else if pod, ok := e.Object.(*v1.Pod); ok {
				if !createPodFlag {
					workloadToPodCost = pod.CreationTimestamp.Sub(createStartTime)
					createPodFlag = true
				}

				if strings.Contains(pod.Name, clusterUnstructuredObj.GetName()) && createStartTime.Unix() <= pod.CreationTimestamp.Unix() {
					if pod.Spec.NodeName != "" {
						status = "ok"

						// pod调度成功耗时
						klog.V(6).Infof("cluster schedule pod successful")

						workloadToScheduleCost, worloadToRunningCost = getPodLifeCycleTimePoint(pod, createStartTime)
					} else {
						for _, condition := range pod.Status.Conditions {
							if strings.Contains(condition.Message, "nodes are available") {
								status = "无可用节点"
								return
							}
						}
					}
				}
			}
		case <-time.After(time.Duration(interval/4) * time.Second):
			if status != "ok" {
				if !createPodFlag {
					status = "创建pod超时"
					klog.Errorf("create pod timeout")
				} else {
					status = "调度超时"
					klog.Errorf("watch timeout")
				}
			}
			return
		}

		if time.Since(watchStartTime).Seconds() > float64(interval/6) {
			if status != "ok" {
				if !createPodFlag {
					status = "创建pod超时"
					klog.Errorf("timeout waiting pod created %s %f s", status, time.Since(watchStartTime).Seconds())
				} else {
					klog.Errorf("timeout waiting pod scheduled %s %f s", status, time.Since(watchStartTime).Seconds())
					status = "调度超时"
				}
			}
			return
		}
	}
}

// get Pod Life Cycle Time Point
func getPodLifeCycleTimePoint(pod *v1.Pod, createStartTime time.Time) (workloadToScheduleCost, worloadToRunningCost time.Duration) {
	workloadToScheduleCost = 0
	worloadToRunningCost = 0
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodScheduled && condition.Status == v1.ConditionTrue {
			if workloadToScheduleCost == 0 {
				workloadToScheduleCost = condition.LastTransitionTime.Sub(createStartTime)
			}

		} else if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
			if worloadToRunningCost == 0 {
				worloadToRunningCost = condition.LastTransitionTime.Sub(createStartTime)
			}
		}
	}

	if pod.Status.Phase == "Running" {
		if worloadToRunningCost == 0 {
			worloadToRunningCost = time.Since(createStartTime)
		}
	} else if pod.Status.Phase == "Completed" || pod.Status.Phase == "Succeeded" {
		if worloadToRunningCost == 0 {
			worloadToRunningCost = time.Since(createStartTime)
		}
		return workloadToScheduleCost, worloadToRunningCost
	}

	return workloadToScheduleCost, worloadToRunningCost
}

// update Nested Map
func updateNestedMap(obj map[string]interface{}, keyPath []string, newValue interface{}) {
	if len(keyPath) == 1 {
		obj[keyPath[0]] = newValue
		return
	}

	nestedObj, ok := obj[keyPath[0]].(map[string]interface{})
	if !ok {
		nestedObj = make(map[string]interface{})
		obj[keyPath[0]] = nestedObj
	}

	updateNestedMap(nestedObj, keyPath[1:], newValue)
}
