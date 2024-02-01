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

// Package masterpodcheck xxx
package masterpodcheck

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metric_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"

	"github.com/dlclark/regexp2"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog"
)

var (
	// master Pod Name List
	masterPodNameList = []string{
		"kube-apiserver",
		"kube-controller-manager",
		"kube-scheduler",
		"etcd",
		"cloud-controller-manager",
	}
)

// Plugin xxx
type Plugin struct {
	stopChan  chan int
	opt       *Options
	checkLock sync.Mutex
}

var (
	// NewGaugeVec creates a new GaugeVec based on the provided GaugeOpts and
	// partitioned by the given label names.
	masterPodCheck = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "master_pod_check",
		Help: "the result of master pod configuration consistency check operation, 1 means ok",
	}, []string{"target", "target_biz", "status", "component", "detection_item"})
	masterPodCheckMap = make(map[string]*prometheus.GaugeVec)
	masterPodMapLock  sync.Mutex
)

// MetricLabel xxx
type MetricLabel struct {
	Target        string
	TargetBiz     string
	Status        string
	Component     string
	DetectionItem string
}

// ToLabelList xxx
func (l *MetricLabel) ToLabelList() []string {
	result := make([]string, 0, 0)
	result = append(result, l.Target)
	result = append(result, l.TargetBiz)
	result = append(result, l.Status)
	result = append(result, l.Component)
	result = append(result, l.DetectionItem)
	return result
}

func init() {
	metric_manager.Register(masterPodCheck)
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
			return fmt.Errorf("decode masterpodcheck config file %s failed, err %s", configFilePath, err.Error())
		}
	}
	if err = p.opt.Validate(); err != nil {
		return err
	}

	p.stopChan = make(chan int)
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
				klog.V(3).Infof("the former masterpodcheck didn't over, skip in this loop")
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
	return "masterpodcheck"
}

// Check xxx
func (p *Plugin) Check() {
	start := time.Now()
	p.checkLock.Lock()
	klog.Infof("start %s", p.Name())
	defer func() {
		klog.Infof("end %s", p.Name())
		plugin_manager.Pm.UnLock()
		p.checkLock.Unlock()
		metric_manager.SetCommonDurationMetric([]string{"masterpodcheck", "", "", ""}, start)
	}()

	wg := sync.WaitGroup{}
	masterPodCheckGaugeVecSetList := make([]*metric_manager.GaugeVecSet, 0, 0)
	// check master pod
	for _, cluster := range plugin_manager.Pm.GetConfig().ClusterConfigs {
		wg.Add(1)
		config := cluster.Config
		clusterId := cluster.ClusterID
		clusterbiz := cluster.BusinessID
		plugin_manager.Pm.Add()
		go func() {
			defer func() {
				klog.V(9).Infof("end masterpodcheck for %s", clusterId)
				wg.Done()
				plugin_manager.Pm.Done()
			}()
			klog.V(9).Infof("start masterpodcheck for %s", clusterId)
			// GetClientsetByConfig cluster k8s client
			clientSet, err := k8s.GetClientsetByConfig(config)
			if err != nil {
				klog.Errorf("%s GetClientsetByClusterId failed: %s", clusterId, err.Error())
				return
			}

			// GetK8sVersion get cluster k8s version
			clusterVersion, err := k8s.GetK8sVersion(clientSet)
			if err != nil {
				klog.Errorf("%s GetK8sVersion failed: %s", clusterId, err.Error())
			}

			// GetPods get namespace pods
			ALLPodList, err := k8s.GetPods(clientSet, "kube-system", v1.ListOptions{}, "")
			if err != nil {
				klog.Errorf("%s GetPods failed: %s", clusterId, err.Error())
			}

			// 筛选静态pod
			// 去掉IP匹配规则"(([1-9]?[0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])"
			podList := make([]corev1.Pod, 0, 0)
			for _, pod := range ALLPodList {
				if strings.Contains(pod.Name, pod.Spec.NodeName) {
					podList = append(podList, pod)
				}
			}

			clusterResult := make([]*metric_manager.GaugeVecSet, 0, 0)
			for _, masterPod := range masterPodNameList {
				var nodeLabelSelector string

				if clusterVersion == "" || podList == nil {
					masterPodCheckGaugeVecSetList = append(masterPodCheckGaugeVecSetList,
						&metric_manager.GaugeVecSet{Labels: []string{clusterId, clusterbiz, "访问集群失败", masterPod, "all"}, Value: 1})
					continue
				}

				if masterPod == "etcd" {
					nodeLabelSelector = "kubernetes.io/node-role-etcd=true"
				} else if masterPod == "cloud-controller-manager" {
					nodeLabelSelector = "node-role.kubernetes.io/master=true"
				} else {
					nodeLabelSelector = "node-role.kubernetes.io/master"
				}

				result := p.checkMasterPod(clientSet, podList, masterPod, nodeLabelSelector, clusterId, clusterbiz, config)
				masterPodCheckGaugeVecSetList = append(masterPodCheckGaugeVecSetList, result...)
				clusterResult = append(clusterResult, result...)

			}

			// 集群单独路径的指标配置
			masterPodMapLock.Lock()
			if _, ok := masterPodCheckMap[clusterId]; !ok {
				masterPodCheckMap[clusterId] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Name: "master_pod_check",
					Help: "the result of master pod configuration consistency check operation, 1 means ok",
				}, []string{"target", "target_biz", "status", "component", "detection_item"})
				metric_manager.MM.RegisterSeperatedMetric(clusterId, masterPodCheckMap[clusterId])
			}
			metric_manager.SetMetric(masterPodCheckMap[clusterId], clusterResult)
			masterPodMapLock.Unlock()
		}()
	}

	wg.Wait()
	// reset metric value
	klog.Infof("length is %d", len(masterPodCheckGaugeVecSetList))
	metric_manager.SetMetric(masterPodCheck, masterPodCheckGaugeVecSetList)
}

// check Master Pod
func (p *Plugin) checkMasterPod(clientSet *kubernetes.Clientset, podList []corev1.Pod, podName string,
	nodeLabelSelector string, clusterId string, clusterBiz string, config *rest.Config) []*metric_manager.GaugeVecSet {
	result := make([]*metric_manager.GaugeVecSet, 0, 0)

	metricLabel := MetricLabel{
		Target:    clusterId,
		TargetBiz: clusterBiz,
		Component: podName,
	}
	var err error

	// 获取本次要检查的master pod列表
	masterPodList := make([]corev1.Pod, 0, 0)
	for _, pod := range podList {
		if strings.Contains(pod.Name, podName) {
			masterPodList = append(masterPodList, pod)

			// 检查pod label
			metricLabel.DetectionItem = "pod_label_check"
			if pod.Labels == nil || len(pod.Labels) == 0 {
				metricLabel.Status = "notok"
				result = append(result,
					&metric_manager.GaugeVecSet{Labels: metricLabel.ToLabelList(), Value: 1})
			}
		}
	}

	// master pod的实例数检测
	metricLabel.DetectionItem = "实例数检测"
	metricLabel.Status = p.checkPodNum(clientSet, nodeLabelSelector, masterPodList, podName)
	result = append(result,
		&metric_manager.GaugeVecSet{Labels: metricLabel.ToLabelList(), Value: 1})

	// pod如果只有一个则无需检查一致性
	if len(masterPodList) > 1 {
		if podName == "kube-scheduler" {
			metricLabel.DetectionItem = "配置文件一致性"
			metricLabel.Status, err = p.checkSchedulerPolicy(clientSet, config, masterPodList)
			if err != nil {
				klog.Errorf("%s checkSchedulerPolicy failed: %s", clusterId, err.Error())
				result = append(result,
					&metric_manager.GaugeVecSet{Labels: metricLabel.ToLabelList(), Value: 1})
			} else {
				result = append(result,
					&metric_manager.GaugeVecSet{Labels: metricLabel.ToLabelList(), Value: 1})
			}
		}
		metricLabel.DetectionItem = "配置一致性"
		metricLabel.Status = p.checkStaticPodConsistency(masterPodList)
		result = append(result,
			&metric_manager.GaugeVecSet{Labels: metricLabel.ToLabelList(), Value: 1})
	}

	// 检查配置文件中的其他检测项
	for _, checkConfig := range p.opt.CheckConfigs {
		if checkConfig.Name == podName {
			metricLabel.DetectionItem = checkConfig.DetectionItem
			for _, pod := range masterPodList {
				checkResult, err := p.checkPodConfig(pod, checkConfig.ConfigPath, checkConfig.ConfigRegex)
				if err != nil {
					klog.Errorf("%s %s checkPodConfig failed: %s", clusterId, podName, err.Error())
					metricLabel.Status = "检测失败"
					result = append(result,
						&metric_manager.GaugeVecSet{Labels: metricLabel.ToLabelList(), Value: 1})
				} else if !checkResult {
					metricLabel.Status = checkConfig.Status
					result = append(result,
						&metric_manager.GaugeVecSet{Labels: metricLabel.ToLabelList(), Value: 1})
					break
				} else if checkResult {
					metricLabel.Status = "ok"
					result = append(result,
						&metric_manager.GaugeVecSet{Labels: metricLabel.ToLabelList(), Value: 1})
				}
			}
		}
	}

	return result
}

// check Pod Num
func (p *Plugin) checkPodNum(clientSet *kubernetes.Clientset, nodeLabelSelector string, masterPodList []corev1.Pod,
	podName string) string {
	masterPodNum := len(masterPodList)

	// ensure number of master node
	// nolint
	if podName != "etcd" {
		ctx := context.Background()
		nodeList, err := clientSet.CoreV1().Nodes().List(ctx, v1.ListOptions{
			LabelSelector:   nodeLabelSelector,
			ResourceVersion: "0",
		})
		if err != nil {
			return "访问集群失败"
		}
		masterNum := len(nodeList.Items)
		if masterNum == 0 {
			return "无master节点"
		} else if masterNum == 1 {
			return "单master节点"
		}
		if masterNum != masterPodNum {
			return "节点pod数量不等"
		}
	}
	return "ok"
}

// check Static Pod Consistency
func (p *Plugin) checkStaticPodConsistency(podList []corev1.Pod) string {
	podSpecList := make(map[string]corev1.PodSpec)
	argsList := make(map[string]map[string][]string)
	for _, pod := range podList {
		spec := pod.Spec
		spec.NodeName = ""
		for index, container := range spec.Containers {
			if argsList[pod.Name] == nil {
				argsList[pod.Name] = make(map[string][]string)
			}

			sort.Strings(container.Args)
			argsList[pod.Name][container.Name] = container.Args
			spec.Containers[index].Args = nil
		}
		podSpecList[pod.Name] = spec
	}

	// 对比容器配置以外的spec是否一致
	var sampleSpec corev1.PodSpec
	var sampleName string
	// nolint
	for podName, spec := range podSpecList {
		if reflect.DeepEqual(sampleSpec, corev1.PodSpec{}) {
			sampleSpec = spec
			sampleName = podName
		} else {
			if !reflect.DeepEqual(sampleSpec, spec) {
				klog.V(9).Infof("%s: %s not equal %s: %s", podName, spec.String(), sampleName, sampleSpec.String())
				return "配置不一致"
			}
		}
	}

	// 对比容器命令行参数是否一致
	var samplePodName string
	for podName := range argsList {
		if samplePodName == "" {
			samplePodName = podName
			break
		}
	}
	if samplePodName == "" {
		return "无有效pod"
	}

	for podName, containers := range argsList {
		if podName == samplePodName {
			continue
		}

		// nolint
		for containerName, args := range containers {
			if sampleArgs, ok := argsList[samplePodName][containerName]; !ok {
				klog.Infof("pod %s doesn't have container %s", samplePodName, containerName)
				return "配置不一致"
			} else {
				err := checkArguments(args, sampleArgs)
				if err != nil {
					klog.Infof("pod %s container %s doesn't equal pod %s : %s",
						samplePodName, containerName, podName, err.Error())
					return "配置不一致"
				}
			}
		}
	}

	return "ok"
}

// check Arguments
func checkArguments(argList1 []string, argList2 []string) error {
	if len(argList1) != len(argList2) {
		return fmt.Errorf("length not equal")
	} else {
		for index, arg1 := range argList1 {
			arg2 := argList2[index]
			if arg1 != arg2 {
				// exclude ip address
				re, _ := regexp.Compile(
					"(([1-9]?[0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])")
				arg1WithoutIp := re.ReplaceAllString(arg1, "")
				arg2WithoutIp := re.ReplaceAllString(arg2, "")
				if arg1WithoutIp != arg2WithoutIp {
					return fmt.Errorf("arg %s not equal", arg1WithoutIp)
				}
			}
		}
	}

	return nil
}

// check Scheduler Policy
func (p *Plugin) checkSchedulerPolicy(clientSet *kubernetes.Clientset, restConfig *rest.Config,
	podList []corev1.Pod) (string, error) {
	if len(podList) == 0 {
		klog.Infof("no kube-scheduler pods were found")
		return "访问集群失败", nil // nolint
	} else if len(podList) == 1 {
		return "单实例", nil
	}

	var filePath string
	for _, item := range podList {
		for _, arg := range item.Spec.Containers[0].Args {
			if strings.Contains(arg, "policy-config-file") || strings.Contains(arg, "--config") {
				filePath = strings.Split(arg, "=")[1]
				break
			}
		}
	}

	if filePath == "" {
		return "ok", nil
	}
	return p.checkPodFileConsistency(restConfig, clientSet, podList, "kube-scheduler", filePath)
}

// check Pod File Consistency
func (p *Plugin) checkPodFileConsistency(restConfig *rest.Config, clientSet *kubernetes.Clientset, podList []corev1.Pod,
	containerName string, filePath string) (string, error) {
	ctx := context.Background()

	var sampleFile string
	var sampleName string
	// var sampleName string
	for _, pod := range podList {
		req := clientSet.CoreV1().RESTClient().Post().Resource("pods").Name(pod.Name).
			Namespace("kube-system").SubResource("exec").Param("container", containerName).
			VersionedParams(&corev1.PodExecOptions{
				Command: []string{"cat", filePath},
				Stdin:   false,
				Stdout:  true,
				Stderr:  true,
				TTY:     false,
			}, scheme.ParameterCodec)

		exec, err := remotecommand.NewSPDYExecutor(restConfig, "POST", req.URL())
		if err != nil {
			return "访问集群失败", fmt.Errorf("NewSPDYExecutor failed: %s", err.Error()) // nolint
		}

		var stdout, stderr bytes.Buffer
		// StreamWithContext
		if err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdin:  nil,
			Stdout: &stdout,
			Stderr: &stderr,
			Tty:    false,
		}); err != nil {
			return "error", fmt.Errorf("Stream failed: %s %s", err.Error(), stderr.String())
		}
		klog.V(6).Infof("%s %s", stdout.String(), stderr.String())
		errMsg := stderr.String()
		execMsg := stdout.String()
		if errMsg != "" {
			klog.Infof("Exec failed: %s", errMsg)
			return "访问集群失败", nil // nolint
		} else if execMsg == "" {
			klog.Infof("%s is blank", filePath)
			return "访问集群失败", nil // nolint
		}

		if sampleFile == "" {
			sampleFile = execMsg
			sampleName = pod.Name
		} else {
			if sampleFile != execMsg {
				klog.Infof("pod %s policy %s doesn't equal pod %s policy %s",
					sampleName, sampleFile, pod.Name, execMsg)
				return "配置不一致", nil // nolint
			}
		}
	}
	return "ok", nil
}

// check Pod Config
func (p *Plugin) checkPodConfig(obj interface{}, path string, regex string) (bool, error) {
	value := reflect.ValueOf(obj)
	fields := strings.Split(path, ".")
	for _, field := range fields {
		if value.Kind() != reflect.Struct && value.Kind() != reflect.Slice && value.Kind() != reflect.Pointer {
			return false, fmt.Errorf("invalid field %s in path %s %s", field, path, value)
		}
		if value.Kind() == reflect.Slice {
			indexStr := strings.Trim(field, "[]")
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return false, fmt.Errorf("invalid field %s in path %s %s", field, path, value)
			}
			if value.Len() >= index+1 {
				value = value.Index(index)
			} else {
				return true, nil
			}
		} else if value.Kind() == reflect.Struct {
			// 如果当前值是结构体，就按照字段名获取对应属性值
			fieldValue := value.FieldByName(field)
			if !fieldValue.IsValid() {
				return false, fmt.Errorf("invalid field %s in path %s %s", field, path, value)
			}
			value = fieldValue
		} else if value.Kind() == reflect.Pointer && !value.IsNil() {
			value = reflect.Indirect(value)
			// 如果当前值是结构体，就按照字段名获取对应属性值
			fieldValue := value.FieldByName(field)
			if !fieldValue.IsValid() {
				return false, fmt.Errorf("invalid field %s in path %s %s", field, path, value)
			}
			value = fieldValue
		} else if value.IsNil() {
			return true, nil
		}
	}

	result := value.Interface().(string)
	reg, err := regexp2.Compile(regex, 0)
	if err != nil {
		return false, err
	}
	return reg.MatchString(result)
}
