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

// Package randhostport xxx
package randhostport

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corev1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/pluginmanager"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/pluginutil"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/types"
)

func init() {
	p := &HostPortInjector{}
	pluginmanager.Register(pluginName, p)
}

// HostPortInjectorConfig config of host port injector
type HostPortInjectorConfig struct {
	StartPort         uint64 `json:"startPort"`
	EndPort           uint64 `json:"endPort"`
	EnableOrderAssign bool   `json:"enableOrderAssign"`
	Kubeconfig        string `json:"kubeconfig"`
}

// HostPortInjector host port injector
type HostPortInjector struct {
	k8sClient *kubernetes.Clientset
	conf      *HostPortInjectorConfig
	stopCh    chan struct{}

	podLister corev1lister.PodLister

	portCache *PortCache

	nextAssignPort uint64
}

// AnnotationKey returns key of the randhostport plugin for hook server to identify
func (hpi *HostPortInjector) AnnotationKey() string {
	return pluginAnnotationKey
}

// Init init host port injector kubeclient
func (hpi *HostPortInjector) Init(configFilePath string) error {
	var err error
	var fileBytes []byte
	var k8sClient *kubernetes.Clientset
	fileBytes, err = ioutil.ReadFile(configFilePath)
	if err != nil {
		blog.Errorf("load config file %s failed, err %s", configFilePath, err.Error())
		return fmt.Errorf("load config file %s failed, err %s", configFilePath, err.Error())
	}
	newConfig := &HostPortInjectorConfig{}
	err = json.Unmarshal(fileBytes, &newConfig)
	if err != nil {
		blog.Errorf("decode config %s failed, err %s", string(fileBytes), err.Error())
		return fmt.Errorf("decode config %s failed, err %s", string(fileBytes), err.Error())
	}
	hpi.conf = newConfig

	var restConfig *rest.Config
	// init k8s client
	if len(hpi.conf.Kubeconfig) == 0 {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return fmt.Errorf("use InCluster restConfig failed, err %s", err.Error())
		}
	} else {
		restConfig, err = clientcmd.BuildConfigFromFlags("", hpi.conf.Kubeconfig)
		if err != nil {
			return fmt.Errorf("build restConfig by file %s failed, err %s", hpi.conf.Kubeconfig, err.Error())
		}
	}

	if !hpi.conf.EnableOrderAssign {
		if err = hpi.initCache(); err != nil {
			return fmt.Errorf("init cache failed, err %s", err.Error())
		}
	} else {
		portPeriod := hpi.conf.EndPort - hpi.conf.StartPort
		hpi.nextAssignPort = uint64(rand.Int63n(int64(portPeriod)) + int64(hpi.conf.StartPort)) // nolint
	}

	blog.Infof("randhostport plugin init cache successfully")

	if !hpi.conf.EnableOrderAssign {
		k8sClient, err = kubernetes.NewForConfig(restConfig)
		if err != nil {
			return fmt.Errorf("build kubeClient failed, err %s", err.Error())
		}
		hpi.k8sClient = k8sClient
		corev1InformerFactory := kubeinformers.NewSharedInformerFactory(hpi.k8sClient, 0)
		podInformer := corev1InformerFactory.Core().V1().Pods().Informer()
		podLister := corev1InformerFactory.Core().V1().Pods().Lister()
		podInformer.AddEventHandler(hpi)
		hpi.podLister = podLister
		hpi.stopCh = make(chan struct{})
		corev1InformerFactory.Start(hpi.stopCh)
		if !cache.WaitForCacheSync(hpi.stopCh, podInformer.HasSynced) {
			return fmt.Errorf("pod cache synced failed")
		}
	}
	blog.Infof("randhostport plugin wait k8s informer cache synced successfully")
	return nil
}

// nolint always nil
func (hpi *HostPortInjector) initCache() error {
	hpi.portCache = NewPortCache()
	for i := hpi.conf.StartPort; i <= hpi.conf.EndPort; i++ {
		hpi.portCache.PushPortEntry(&PortEntry{
			Port:     i,
			Quantity: uint64(0),
		})
	}
	return nil
}

// Handle handles webhook request of host port injector
func (hpi *HostPortInjector) Handle(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	// when the kind is not Pod, ignore hook
	if req.Kind.Kind != "Pod" {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	if req.Operation != v1beta1.Create {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}

	started := time.Now()
	pod := &corev1.Pod{}
	if err := json.Unmarshal(req.Object.Raw, pod); err != nil {
		blog.Errorf("cannot decode raw object %s to pod, err %s", string(req.Object.Raw), err.Error())
		metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionResponse(err)
	}
	// Deal with potential empty fileds, e.g., when the pod is created by a deployment
	if pod.ObjectMeta.Namespace == "" {
		pod.ObjectMeta.Namespace = req.Namespace
	}
	if !hpi.injectRequired(pod) {
		return &v1beta1.AdmissionResponse{
			Allowed: true,
			PatchType: func() *v1beta1.PatchType {
				pt := v1beta1.PatchTypeJSONPatch
				return &pt
			}(),
		}
	}

	patches, err := hpi.injectToPod(pod)
	if err != nil {
		blog.Errorf("inject to pod %s/%s failed, err %s", pod.GetName(), pod.GetNamespace(), err.Error())
		metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionResponse(err)
	}
	patchesBytes, err := json.Marshal(patches)
	if err != nil {
		blog.Errorf("encoding patches failed, err %s", err.Error())
		metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionResponse(err)
	}

	metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusSuccess, started)
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchesBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

// injectRequired xxx
// check if pod injection needed
func (hpi *HostPortInjector) injectRequired(pod *corev1.Pod) bool {
	if value, ok := pod.Annotations[pluginAnnotationKey]; !ok || value != pluginAnnotationValue {
		blog.Warnf("Pod %s/%s has no expected annoation key & value", pod.Name, pod.Namespace)
		return false
	}
	return true
}

// nolint funlen
func (hpi *HostPortInjector) injectToPod(pod *corev1.Pod) ([]types.PatchOperation, error) {
	portStrs := getPortStringsFromPodAnnotations(pod.Annotations)
	if len(portStrs) == 0 {
		return nil, fmt.Errorf("pod %s/%s does not specify container port to inject random hostport",
			pod.GetName(), pod.GetNamespace())
	}
	alsoChangeContainerPort := false
	if _, ok := pod.Annotations[pluginContainerPortsAnnotationKey]; ok {
		alsoChangeContainerPort = true
	}
	// to collect how many port should be injected
	containerPortsIndexList := make([][]int, len(pod.Spec.Containers))
	containerPortList := make([]int32, 0)
	needInjectCount := 0

	for containerIndex, container := range pod.Spec.Containers {
		for portIndex, containerPort := range container.Ports {
			for _, portStr := range portStrs {
				// 根据声明的port名字或者端口号去匹配容器端口
				// 1. 先按名字匹配
				// 2. 再按端口号匹配
				if portStr == containerPort.Name {
					containerPortsIndexList[containerIndex] = append(containerPortsIndexList[containerIndex], portIndex)
					containerPortList = append(containerPortList, containerPort.ContainerPort)
					needInjectCount++
					break
				}
				// 如果不能转成数字，则跳过，说明用户声明的是端口名字，名字不匹配。
				portNumber, err := strconv.Atoi(portStr)
				if err != nil {
					continue
				}
				if int32(portNumber) == containerPort.ContainerPort {
					containerPortsIndexList[containerIndex] = append(containerPortsIndexList[containerIndex], portIndex)
					containerPortList = append(containerPortList, containerPort.ContainerPort)
					needInjectCount++
					break
				}
			}
		}
	}
	if needInjectCount != len(portStrs) {
		return nil, fmt.Errorf("not all ports %v in annotation match ports in container", portStrs)
	}

	var hostPorts []*PortEntry
	if !hpi.conf.EnableOrderAssign {
		// get rand host port
		hpi.portCache.Lock()
		for i := 0; i < needInjectCount; i++ {
			portEntry := hpi.portCache.PopPortEntry()
			hostPorts = append(hostPorts, portEntry)
		}
		for _, hostPort := range hostPorts {
			hostPort.Quantity++
			hpi.portCache.PushPortEntry(hostPort)
		}
		hpi.portCache.Unlock()
	} else {
		// get order host port
		for i := 0; i < needInjectCount; i++ {
			portEntry := hpi.getOrderPortEntry()
			blog.Infof("get order portentry %d for pod %s/%s", portEntry.Port, pod.GetName(), pod.GetNamespace())
			hostPorts = append(hostPorts, portEntry)
		}
	}
	var retPatches []types.PatchOperation
	// patch affinity
	retPatches = append(retPatches, hpi.generateAffinityPath(pod, hostPorts))
	// patch label
	retPatches = append(retPatches, hpi.generateLabelPatch(pod, hostPorts))
	// patch container port
	hostPortCount := 0

	// containerPort=>hostPort
	hostPortMapping := make(map[uint64]uint64, len(containerPortsIndexList))
	for containerIndex, portIndexList := range containerPortsIndexList {
		for _, portIndex := range portIndexList {
			containerPort := pod.Spec.Containers[containerIndex].Ports[portIndex].ContainerPort
			// inject hostport into container port
			retPatches = append(retPatches, types.PatchOperation{
				Path:  fmt.Sprintf(PatchPathContainerHostPort, containerIndex, portIndex),
				Op:    PatchOperationAdd,
				Value: hostPorts[hostPortCount].Port,
			})
			if alsoChangeContainerPort {
				retPatches = append(retPatches, types.PatchOperation{
					Path:  fmt.Sprintf(PatchPathContainerContainerPort, containerIndex, portIndex),
					Op:    PatchOperationAdd,
					Value: hostPorts[hostPortCount].Port,
				})
				hostPortMapping[hostPorts[hostPortCount].Port] = hostPorts[hostPortCount].Port
			} else {
				hostPortMapping[uint64(containerPort)] = hostPorts[hostPortCount].Port
			}
			hostPortCount++
		}
		// inject all hostport envs into all containers
		envs := pod.Spec.Containers[containerIndex].Env
		envPatch := hpi.generateEnvPatch(PatchPathContainerEnv, containerIndex, envs, containerPortList, hostPorts)
		retPatches = append(retPatches, envPatch)
	}
	// injecto all hostport envs into all init containers
	for containerIndex, container := range pod.Spec.InitContainers {
		envPatch := hpi.generateEnvPatch(
			PatchPathInitContainerEnv, containerIndex, container.Env, containerPortList, hostPorts)
		retPatches = append(retPatches, envPatch)
	}

	// inject hostport into pod annotations
	retPatches = append(retPatches, hpi.generateAnnotationsPatch(pod, hostPortMapping))
	fmt.Println(retPatches)

	return retPatches, nil
}

// generateAnnotationsPatch generate patch for pod annotations
func (hpi *HostPortInjector) generateAnnotationsPatch(
	pod *corev1.Pod, hostPortMapping map[uint64]uint64) types.PatchOperation {
	annotations := pod.Annotations
	op := PatchOperationReplace
	if len(annotations) == 0 {
		op = PatchOperationAdd
		annotations = make(map[string]string)
	}
	for containerPort, hostPort := range hostPortMapping {
		annotations[fmt.Sprintf(annotationsRandHostportPrefix+"%d", containerPort)] = strconv.FormatUint(hostPort, 10)
	}
	return types.PatchOperation{
		Path:  PatchPathPodAnnotations,
		Op:    op,
		Value: annotations,
	}
}

func (hpi *HostPortInjector) generateEnvPatch(
	patchPath string, index int,
	envs []corev1.EnvVar, containerPortList []int32, hostPorts []*PortEntry) types.PatchOperation {
	envPatchOp := PatchOperationReplace
	if len(envs) == 0 {
		envPatchOp = PatchOperationAdd
	}
	for tmpIndex, containerPort := range containerPortList {
		envs = append(envs, corev1.EnvVar{
			Name:  envRandHostportPrefix + strconv.FormatInt(int64(containerPort), 10),
			Value: strconv.FormatUint(hostPorts[tmpIndex].Port, 10),
		})
	}

	envs = append(envs, corev1.EnvVar{
		Name: envRandHostportHostIP,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "status.hostIP",
			},
		},
	},
	)
	envs = append(envs, corev1.EnvVar{
		Name: envRandHostportPodName,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "metadata.name",
			},
		},
	},
	)
	envs = append(envs, corev1.EnvVar{
		Name: envRandHostportPodNamespace,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "metadata.namespace",
			},
		},
	},
	)

	return types.PatchOperation{
		Path:  fmt.Sprintf(patchPath, index),
		Op:    envPatchOp,
		Value: envs,
	}
}

// generateAffinityPath xxx
// generate pod affinity patch
func (hpi *HostPortInjector) generateAffinityPath(pod *corev1.Pod, hostPorts []*PortEntry) types.PatchOperation {
	var affinity *corev1.Affinity
	op := PatchOperationReplace
	if pod.Spec.Affinity == nil {
		op = PatchOperationAdd
		affinity = &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{},
		}
	} else if pod.Spec.Affinity.PodAntiAffinity == nil {
		affinity = pod.Spec.Affinity
		affinity.PodAntiAffinity = &corev1.PodAntiAffinity{}
	} else {
		affinity = pod.Spec.Affinity
	}
	for _, hostPort := range hostPorts {
		affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(
			affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, corev1.PodAffinityTerm{
				LabelSelector: k8smetav1.SetAsLabelSelector(labels.Set(map[string]string{
					strconv.FormatUint(
						hostPort.Port, 10) + podHostportLabelSuffix: strconv.FormatUint(
						hostPort.Port, 10),
				})),
				TopologyKey: "kubernetes.io/hostname",
			})
	}
	return types.PatchOperation{
		Path:  PatchPathAffinity,
		Op:    op,
		Value: affinity,
	}
}

// generateLabelPatch xxx
// generate pod label patch
func (hpi *HostPortInjector) generateLabelPatch(pod *corev1.Pod, hostPorts []*PortEntry) types.PatchOperation {
	labels := pod.Labels
	op := PatchOperationReplace
	if len(labels) == 0 {
		op = PatchOperationAdd
		labels = make(map[string]string)
	}
	labels[podHostportLabelFlagKey] = podHostportLabelFlagValue
	for _, hostPort := range hostPorts {
		labels[strconv.FormatUint(hostPort.Port, 10)+podHostportLabelSuffix] = strconv.FormatUint(hostPort.Port, 10)
	}
	return types.PatchOperation{
		Path:  PatchPathPodLabel,
		Op:    op,
		Value: labels,
	}
}

// Close do close action, clean
func (hpi *HostPortInjector) Close() error {
	hpi.stopCh <- struct{}{}
	return nil
}

// OnAdd add event callback
func (hpi *HostPortInjector) OnAdd(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		blog.Warnf("added obj %v is not Pod type", obj)
		return
	}
	hostports := getHostPortByLabels(pod.Labels)
	for _, portNumber := range hostports {
		selector := labels.SelectorFromSet(labels.Set(map[string]string{
			strconv.FormatUint(portNumber, 10) + podHostportLabelSuffix: strconv.FormatUint(portNumber, 10),
		}))
		pods, err := hpi.podLister.List(selector)
		if err != nil {
			blog.Errorf("list pod by selector %s failed, err %s", selector, err.Error())
		}
		hpi.portCache.Lock()
		hpi.portCache.PushPortEntry(&PortEntry{
			Port:     portNumber,
			Quantity: uint64(len(pods)),
		})
		hpi.portCache.Unlock()
		blog.V(5).Infof("update portentry %d quantity to %d successfully", portNumber, len(pods))
	}
}

// OnUpdate update event callback
func (hpi *HostPortInjector) OnUpdate(newObj, oldObj interface{}) {
	// do nothing
}

// OnDelete delete event callback
func (hpi *HostPortInjector) OnDelete(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		blog.Warnf("delete obj %v is not Pod type", obj)
		return
	}
	hostports := getHostPortByLabels(pod.Labels)
	for _, portNumber := range hostports {
		hpi.portCache.Lock()
		if err := hpi.portCache.DecPortQuantity(portNumber); err != nil {
			blog.Warnf("decrease port %d quantity failed, err %s", portNumber, err.Error())
		}
		hpi.portCache.Unlock()
		blog.V(5).Infof("descrease portentry %d quantity successfully", portNumber)
	}
}

func (hpi *HostPortInjector) getOrderPortEntry() *PortEntry {
	portEntry := &PortEntry{
		Port: hpi.nextAssignPort,
	}
	hpi.nextAssignPort++
	if hpi.nextAssignPort > hpi.conf.EndPort {
		hpi.nextAssignPort = hpi.conf.StartPort
	}
	return portEntry
}
