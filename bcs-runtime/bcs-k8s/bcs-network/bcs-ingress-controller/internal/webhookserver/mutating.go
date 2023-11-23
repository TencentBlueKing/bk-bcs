/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package webhookserver

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/eventer"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/portpoolcache"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// PatchOperation struct for k8s webhook patch
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// entry for allocate cloud loadbalancer port
type portEntry struct {
	poolNamespace string
	poolName      string
	protocol      string
	port          int
	hostPort      bool
	itemName      string
}

// combine container port info into port entry format
func getPortEntryListFromPod(pod *k8scorev1.Pod, annotationPorts []*annotationPort) ([]*portEntry, error) {
	var retPorts []*portEntry
	for _, port := range annotationPorts {
		tmpEntry := &portEntry{
			poolNamespace: port.poolNamespace,
			poolName:      port.poolName,
			protocol:      port.protocol,
			hostPort:      port.hostPort,
			itemName:      port.itemName,
		}
		if len(port.poolNamespace) == 0 {
			tmpEntry.poolNamespace = pod.GetNamespace()
		}
		found := false
		for _, container := range pod.Spec.Containers {
			for _, containerPort := range container.Ports {
				if port.portIntOrStr == containerPort.Name {
					found = true
					tmpEntry.port = int(containerPort.ContainerPort)
					if len(port.protocol) == 0 {
						tmpEntry.protocol = string(containerPort.Protocol)
					}
					break
				}
				portNumber, err := strconv.Atoi(port.portIntOrStr)
				if err != nil {
					continue
				}
				if int32(portNumber) == containerPort.ContainerPort {
					found = true
					tmpEntry.port = int(containerPort.ContainerPort)
					if len(port.protocol) == 0 {
						tmpEntry.protocol = string(containerPort.Protocol)
					}
					break
				}
			}
		}
		if !found {
			// if ports in annotations are not found in container ports, return err
			return nil, fmt.Errorf("port %s not found with container port name or port number", port.portIntOrStr)
		}
		retPorts = append(retPorts, tmpEntry)
	}
	return retPorts, nil
}

// remove allocated port item from port pool cache
func (s *Server) cleanAllocatedResource(items [][]portpoolcache.AllocatedPortItem) {
	for _, list := range items {
		for _, item := range list {
			s.poolCache.ReleasePortBinding(item.PoolKey, item.PoolItemKey, item.Protocol, item.StartPort, item.EndPort)
		}
	}
}

// check existed port binding
func (s *Server) checkExistedPortBinding(name, namespace string, portList []*portEntry, annotation map[string]string) (
	*networkextensionv1.PortBinding, error) {
	portBinding := &networkextensionv1.PortBinding{}
	if err := s.k8sClient.Get(context.Background(), k8stypes.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, portBinding); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "get portbinding '%s/%s' failed", namespace, name)
	}
	// to prevent pod from reusing the old portbinding without keep duration
	if !isPortBindingKeepDurationExisted(portBinding) {
		return nil, errors.Errorf("found previous uncleaned portbinding '%s/%s' and need wait",
			portBinding.GetName(), portBinding.GetNamespace())
	}
	// 用户移除了Pod上的KeepDuration标记
	if !isKeepDurationExisted(annotation) {
		blog.Infof("remove '%s/%s' keep duration annotate, delete portBinding quickly", namespace, name)
		// 移除portBinding上的keepDuration注解，尽快删除
		if err := s.removePortBindingAnnotation(portBinding); err != nil {
			return nil, errors.Wrapf(err, "remove portbinding '%s/%s' annotations failed, err: %s",
				portBinding.GetNamespace(), portBinding.GetName(), err.Error())
		}
		if err := s.k8sClient.Delete(context.Background(), portBinding, &client.DeleteOptions{}); err != nil {
			return nil, errors.Wrapf(err, "delete portbinding '%s/%s' failed",
				portBinding.GetName(), portBinding.GetNamespace())
		}
		return nil, nil
	}

	// to prevent pod from reusing the old portbinding which is being deleted
	if portBinding.DeletionTimestamp != nil {
		return nil, errors.Errorf("portbinding %s/%s is deleting",
			portBinding.GetName(), portBinding.GetNamespace())
	}
	rsPortMap := make(map[string]struct{})
	for _, item := range portBinding.Spec.PortBindingList {
		key := getPoolPortKey(item.PoolNamespace, item.PoolName, item.PoolItemName, item.Protocol, item.RsStartPort)
		if _, ok := rsPortMap[key]; !ok {
			rsPortMap[key] = struct{}{}
		}
	}
	for _, port := range portList {
		key := getPoolPortKey(port.poolNamespace, port.poolNamespace, port.itemName, port.protocol, port.port)
		if _, ok := rsPortMap[key]; !ok {
			blog.Warnf("port '%d' is not in portbinding '%s/%s', need to delete portbinding first",
				port.port, portBinding.GetName(), portBinding.GetNamespace())
			// 移除portBinding上的keepDuration注解，尽快删除
			if err := s.removePortBindingAnnotation(portBinding); err != nil {
				return nil, errors.Wrapf(err, "remove portbinding '%s/%s' annotations failed, err: %s",
					portBinding.GetNamespace(), portBinding.GetName(), err.Error())
			}
			if err := s.k8sClient.Delete(context.Background(), portBinding, &client.DeleteOptions{}); err != nil {
				return nil, errors.Wrapf(err, "delete portbinding '%s/%s' failed",
					portBinding.GetName(), portBinding.GetNamespace())
			}
			blog.Warnf("portbinding '%s/%s' is deleted because of port '%+v' not in it",
				portBinding.GetName(), portBinding.GetNamespace(), port)
			return nil, nil
		}
	}
	return portBinding, nil
}

// inject port pool item info into pod annotations and envs
func (s *Server) mutatingPod(pod *k8scorev1.Pod) ([]PatchOperation, error) {
	if pod.Annotations == nil {
		return nil, fmt.Errorf("pod annotation is nil")
	}
	// get port info that should be injected
	annotationValue, ok := pod.Annotations[constant.AnnotationForPortPoolPorts]
	if !ok {
		return nil, fmt.Errorf("pod lack annotation key %s", constant.AnnotationForPortPoolPorts)
	}
	annotationPorts, err := parserAnnotation(annotationValue)
	if err != nil {
		return nil, fmt.Errorf("parse annotation failed, err %s", err.Error())
	}
	portEntryList, err := getPortEntryListFromPod(pod, annotationPorts)
	if err != nil {
		return nil, fmt.Errorf("match container ports for pod %s/%s failed, err %s",
			pod.GetName(), pod.GetNamespace(), err.Error())
	}

	// check for existed port binding
	portBinding, err := s.checkExistedPortBinding(pod.GetName(), pod.GetNamespace(), portEntryList, pod.GetAnnotations())
	if err != nil {
		return nil, errors.Wrapf(err, checkPortBindingExistedFailed)
	}
	if portBinding != nil {
		blog.Infof("pod '%s/%s' reuse portbinding", pod.GetName(), pod.GetNamespace())
		return s.patchPodByBinding(pod, portBinding)
	}

	blog.Infof("pod '%s/%s' do port inject", pod.GetNamespace(), pod.GetName())

	portPoolItemStatusList, portItemListArr, err := s.portAllocate(portEntryList)
	if err != nil {
		return nil, err
	}

	// patch annotations
	var retPatches []PatchOperation
	annotationPortsPatch, err := s.generatePortsAnnotationPatch(
		pod.Annotations, portPoolItemStatusList, portItemListArr, portEntryList)
	if err != nil {
		s.cleanAllocatedResource(portItemListArr)
		return nil, fmt.Errorf("generate ports of port pool annotations failed, err %s", err.Error())
	}
	retPatches = append(retPatches, annotationPortsPatch)

	// patch readiness gate
	if _, ok := pod.Annotations[constant.AnnotationForPortPoolReadinessGate]; ok {
		readinessGatePatch, err := s.generatePodReadinessGate(pod)
		if err != nil {
			s.cleanAllocatedResource(portItemListArr)
			return nil, fmt.Errorf("generate pod readiness gate failed, err %s", err.Error())
		}
		retPatches = append(retPatches, readinessGatePatch)
	}

	// patch envs
	for index, initContainer := range pod.Spec.InitContainers {
		envPatch := s.generateContainerEnvPatch(
			constant.PathPathInitContainerEnv, index, initContainer,
			portPoolItemStatusList, portItemListArr, portEntryList)
		retPatches = append(retPatches, envPatch)
	}
	for index, container := range pod.Spec.Containers {
		envPatch := s.generateContainerEnvPatch(
			constant.PatchPathContainerEnv, index, container, portPoolItemStatusList, portItemListArr, portEntryList)
		retPatches = append(retPatches, envPatch)
	}

	for _, itemList := range portItemListArr {
		for _, item := range itemList {
			blog.Infof("pod '%s/%s' allocate port from cache: %v", pod.GetNamespace(), pod.GetName(), item)
		}
	}
	go s.handleForPodCreateFailed(pod, portItemListArr)
	return retPatches, nil
}

const (
	compensationPodCreateFailedDuration = 10
)

// handleForPodCreateFailed 补偿 GameWorkload 创建 Pod 失败导致缓存中的端口泄漏问题。
// 当确认创建失败后，将已分配的端口清理掉
func (s *Server) handleForPodCreateFailed(pod *k8scorev1.Pod, portItemListArr [][]portpoolcache.AllocatedPortItem) {
	defer func() {
		if r := recover(); r != nil {
			blog.Errorf("pod '%s/%s' check failed create failed: %v, occurred a panic: %s\n",
				pod.GetNamespace(), pod.GetName(), r, string(debug.Stack()))
		}
	}()

	var workloadName string
	for i := range pod.OwnerReferences {
		kind := pod.OwnerReferences[i].Kind
		if kind == eventer.KindGameDeployment || kind == eventer.KindGameStatefulSet {
			workloadName = pod.OwnerReferences[i].Name
			break
		}
	}
	if workloadName == "" {
		return
	}
	traceID := uuid.New().String()
	triggered := make(chan struct{})
	s.eventWatcher.RegisterEventHook(eventer.HookKindPodCreateFailed, traceID, func(event *k8scorev1.Event) {
		if event.InvolvedObject.Namespace != pod.GetNamespace() || event.InvolvedObject.Name != workloadName {
			return
		}
		// NOTE: 目前通过判断 event.Message 中是否存在 Pod 的名字来判断是否创建失败了
		if !strings.Contains(event.Message, pod.Name) {
			return
		}
		// NOTE: 如果因为 portBinding 已存在导致的 Pod FailedCreate，则不需要回收
		if strings.Contains(event.Message, checkPortBindingExistedFailed) {
			close(triggered)
			return
		}
		blog.Warnf("pod '%s/%s' workload '%s' check pod create failed got create failed event: %s",
			pod.GetNamespace(), pod.GetName(), workloadName, event.Message)

		tmpPod := new(k8scorev1.Pod)
		if err := s.k8sClient.Get(context.Background(), k8stypes.NamespacedName{
			Name:      pod.GetName(),
			Namespace: pod.GetNamespace(),
		}, tmpPod); err != nil {
			// 如果确实发现 Pod 不存在则对分配的端口进行回收
			if k8serrors.IsNotFound(err) {
				blog.Warnf("pod '%s/%s' not exist when check failed create, so delete the port it allocated",
					pod.GetNamespace(), pod.GetName())
				s.poolCache.Lock()
				s.cleanAllocatedResource(portItemListArr)
				s.poolCache.Unlock()
				close(triggered)
			} else {
				blog.Errorf("pod '%s/%s' check failed create query failed: %s",
					pod.GetNamespace(), pod.GetName(), err.Error())
			}
			return
		}
		blog.Infof("pod '%s/%s' actually exist, so don't need to delete the port it allocated.",
			pod.GetNamespace(), pod.GetName())
	})
	timeout := time.After(compensationPodCreateFailedDuration * time.Second)
	select {
	case <-triggered:
		s.eventWatcher.UnRegisterEventHook(eventer.HookKindPodCreateFailed, traceID)
	case <-timeout:
		s.eventWatcher.UnRegisterEventHook(eventer.HookKindPodCreateFailed, traceID)
	}
}

// patch pod by port binding object
func (s *Server) patchPodByBinding(
	pod *k8scorev1.Pod, portBinding *networkextensionv1.PortBinding) ([]PatchOperation, error) {
	blog.Infof("pod '%s' reused portbinding do port inject", pod.GetNamespace(), pod.GetName())

	// patch annotations
	var retPatches []PatchOperation
	annotationPortsPatch, err := s.generatePortsAnnoPatchByBinding(pod, portBinding)
	if err != nil {
		return nil, fmt.Errorf("generate annotations ports by portbinding failed, err %s", err.Error())
	}
	retPatches = append(retPatches, annotationPortsPatch)

	// patch readiness gate
	if _, ok := pod.Annotations[constant.AnnotationForPortPoolReadinessGate]; ok {
		readinessGatePatch, err := s.generatePodReadinessGate(pod)
		if err != nil {
			return nil, fmt.Errorf("generate pod readiness gate failed, err %s", err.Error())
		}
		retPatches = append(retPatches, readinessGatePatch)
	}

	// patch envs
	for index, initContainer := range pod.Spec.InitContainers {
		envPatch := s.generateContainerEnvPatchByBinding(
			constant.PathPathInitContainerEnv, index, initContainer, portBinding)
		retPatches = append(retPatches, envPatch)
	}
	for index, container := range pod.Spec.Containers {
		envPatch := s.generateContainerEnvPatchByBinding(
			constant.PatchPathContainerEnv, index, container, portBinding)
		retPatches = append(retPatches, envPatch)
	}
	return retPatches, nil
}

// generate container annotations patch object by existed portbinding object
func (s *Server) generatePortsAnnoPatchByBinding(
	pod *k8scorev1.Pod, portBinding *networkextensionv1.PortBinding) (PatchOperation, error) {
	portValues, err := json.Marshal(portBinding.Spec.PortBindingList)
	if err != nil {
		return PatchOperation{}, fmt.Errorf("encoding portbinding list to json failed, err %s", err)
	}
	annotations := pod.Annotations
	op := constant.PatchOperationReplace
	if len(annotations) == 0 {
		op = constant.PatchOperationAdd
		annotations = make(map[string]string)
	}
	annotations[constant.AnnotationForPortPoolBindings] = string(portValues)
	annotations[constant.AnnotationForPortPoolBindingStatus] = constant.AnnotationForPodStatusNotReady
	return PatchOperation{
		Path:  constant.PatchPathPodAnnotations,
		Op:    op,
		Value: annotations,
	}, nil
}

// generate container environments patch object by portbinding
func (s *Server) generateContainerEnvPatchByBinding(
	patchPath string, index int, container k8scorev1.Container,
	portBinding *networkextensionv1.PortBinding) PatchOperation {
	envs := container.Env
	envPatchOp := constant.PatchOperationReplace
	if len(envs) == 0 {
		envPatchOp = constant.PatchOperationAdd
	}
	for _, binding := range portBinding.Spec.PortBindingList {
		vipList := genVipList(binding.PoolItemLoadBalancers)
		envs = append(envs, k8scorev1.EnvVar{
			Name:  constant.EnvVIPsPrefixForPortPoolPort + binding.Protocol + "_" + strconv.Itoa(binding.RsStartPort),
			Value: getPortEnvValue(binding.StartPort, binding.EndPort, vipList),
		})
	}
	return PatchOperation{
		Path:  fmt.Sprintf(patchPath, index),
		Op:    envPatchOp,
		Value: envs,
	}
}

// generate container annotations patch object by port entry list
func (s *Server) generatePortsAnnotationPatch(annotations map[string]string,
	portPoolItemStatusList []*networkextensionv1.PortPoolItemStatus,
	portItemList [][]portpoolcache.AllocatedPortItem,
	portEntryList []*portEntry) (PatchOperation, error) {
	var generatedPortList []*networkextensionv1.PortBindingItem
	for index, portEntry := range portEntryList {
		poolName, poolNamespace, err := parsePoolKey(portItemList[index][0].PoolKey)
		if err != nil {
			return PatchOperation{}, fmt.Errorf(
				"parse pool key failed when generatePortsAnnotationPatch, err %s", err.Error())
		}
		for _, item := range portItemList[index] {
			tmpPort := &networkextensionv1.PortBindingItem{
				PoolName:              poolName,
				PoolNamespace:         poolNamespace,
				PoolItemName:          portPoolItemStatusList[index].ItemName,
				LoadBalancerIDs:       portPoolItemStatusList[index].LoadBalancerIDs,
				PoolItemLoadBalancers: portPoolItemStatusList[index].PoolItemLoadBalancers,
				Protocol:              item.Protocol,
				StartPort:             item.StartPort,
				EndPort:               item.EndPort,
				RsStartPort:           portEntry.port,
				HostPort:              portEntry.hostPort,
				External:              portPoolItemStatusList[index].External,
			}
			generatedPortList = append(generatedPortList, tmpPort)
		}
	}
	portValues, err := json.Marshal(generatedPortList)
	if err != nil {
		return PatchOperation{}, err
	}
	op := constant.PatchOperationReplace
	if len(annotations) == 0 {
		op = constant.PatchOperationAdd
		annotations = make(map[string]string)
	}
	annotations[constant.AnnotationForPortPoolBindings] = string(portValues)
	annotations[constant.AnnotationForPortPoolBindingStatus] = constant.AnnotationForPodStatusNotReady
	return PatchOperation{
		Path:  constant.PatchPathPodAnnotations,
		Op:    op,
		Value: annotations,
	}, nil
}

// generate container environments patch object by port entry list
func (s *Server) generateContainerEnvPatch(
	patchPath string, index int, container k8scorev1.Container,
	portPoolItemStatusList []*networkextensionv1.PortPoolItemStatus,
	portItemList [][]portpoolcache.AllocatedPortItem,
	portEntryList []*portEntry) PatchOperation {

	envs := container.Env
	envPatchOp := constant.PatchOperationReplace
	if len(envs) == 0 {
		envPatchOp = constant.PatchOperationAdd
	}
	for index, portEntry := range portEntryList {
		vipList := genVipList(portPoolItemStatusList[index].PoolItemLoadBalancers)
		for _, item := range portItemList[index] {
			envs = append(envs, k8scorev1.EnvVar{
				Name:  constant.EnvVIPsPrefixForPortPoolPort + item.Protocol + "_" + strconv.Itoa(portEntry.port),
				Value: getPortEnvValue(item.StartPort, item.EndPort, vipList),
			})
		}
	}
	return PatchOperation{
		Path:  fmt.Sprintf(patchPath, index),
		Op:    envPatchOp,
		Value: envs,
	}
}

// generate pod readiness gate patch object
func (s *Server) generatePodReadinessGate(pod *k8scorev1.Pod) (PatchOperation, error) {
	readinessGates := pod.Spec.ReadinessGates
	op := constant.PatchOperationReplace
	if len(readinessGates) == 0 {
		op = constant.PatchOperationAdd
	}
	readinessGates = append(readinessGates, k8scorev1.PodReadinessGate{
		ConditionType: constant.ConditionTypeBcsIngressPortBinding,
	})
	return PatchOperation{
		Path:  constant.PatchPathPodReadinessGate,
		Op:    op,
		Value: readinessGates,
	}, nil
}

func (s *Server) removePortBindingAnnotation(portBinding *networkextensionv1.PortBinding) error {
	patchStruct := []PatchOperation{
		{
			Op:   "remove",
			Path: constant.PatchPathPodAnnotations + "/" + networkextensionv1.PortPoolBindingAnnotationKeyKeepDuration,
		},
	}

	patchBytes, err := json.Marshal(patchStruct)
	if err != nil {
		blog.Errorf("marshal patch struct failed, err: %s", err)
		return err
	}

	if err = s.k8sClient.Patch(context.TODO(), portBinding, client.RawPatch(k8stypes.JSONPatchType,
		patchBytes)); err != nil {
		blog.Errorf("patch portbinding '%s/%s' failed, err: %s", portBinding.GetNamespace(), portBinding.GetName(), err)
		return err
	}

	return nil
}

func genVipList(lbs []*networkextensionv1.IngressLoadBalancer) []string {
	var vipList []string
	for _, lbObj := range lbs {
		if len(lbObj.IPs) != 0 {
			vipList = append(vipList, lbObj.IPs...)
		}
		if len(lbObj.DNSName) != 0 {
			vipList = append(vipList, lbObj.DNSName)
		}
	}

	return vipList
}
