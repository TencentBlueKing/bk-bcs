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
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/portpoolcache"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
func (s *Server) checkExistedPortBinding(pod *k8scorev1.Pod, portList []*portEntry) (
	*networkextensionv1.PortBinding, error) {
	portBinding := &networkextensionv1.PortBinding{}
	if err := s.k8sClient.Get(context.Background(), k8stypes.NamespacedName{
		Name:      pod.GetName(),
		Namespace: pod.GetNamespace(),
	}, portBinding); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}
		blog.Warnf("get portbinding %s/%s failed, err %s",
			pod.GetName(), pod.GetNamespace(), err.Error())
		return nil, fmt.Errorf("get portbinding %s/%s failed, err %s",
			pod.GetName(), pod.GetNamespace(), err.Error())
	}
	// to prevent pod from reusing the old portbinding without keep duration
	if !isPortBindingKeepDurationExisted(portBinding) {
		blog.Warnf("found previous uncleaned portbinding %s/%s, wait",
			portBinding.GetName(), portBinding.GetNamespace())
		return nil, fmt.Errorf("found previous uncleaned portbinding %s/%s, wait",
			portBinding.GetName(), portBinding.GetNamespace())
	}
	// to prevent pod from reusing the old portbinding which is being deleted
	if portBinding.DeletionTimestamp != nil {
		blog.Warnf("portbinding %s/%s is deleting",
			portBinding.GetName(), portBinding.GetNamespace())
		return nil, fmt.Errorf("portbinding %s/%s is deleting",
			portBinding.GetName(), portBinding.GetNamespace())
	}
	rsPortMap := make(map[int]struct{})
	for _, item := range portBinding.Spec.PortBindingList {
		if _, ok := rsPortMap[item.RsStartPort]; !ok {
			rsPortMap[item.RsStartPort] = struct{}{}
		}
	}
	for _, port := range portList {
		if _, ok := rsPortMap[port.port]; !ok {
			blog.Warnf("port %d is not in portbinding %s/%s, to delete portbinding first",
				port.port, portBinding.GetName(), portBinding.GetNamespace())
			if err := s.k8sClient.Delete(context.Background(), portBinding, &client.DeleteOptions{}); err != nil {
				return nil, fmt.Errorf(
					"port %d is not in portbinding %s/%s, to delete portbinding first, but delete failed, err %s",
					port.port, portBinding.GetName(), portBinding.GetNamespace(), err.Error())
			}
			return nil, fmt.Errorf("port %d is not in portbinding %s/%s, to delete portbinding first",
				port.port, portBinding.GetName(), portBinding.GetNamespace())
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
	portBinding, err := s.checkExistedPortBinding(pod, portEntryList)
	if err != nil {
		return nil, err
	}
	if portBinding != nil {
		blog.Infof("pod %s/%s reuse portbinding", pod.GetName(), pod.GetNamespace())
		return s.patchPodByBinding(pod, portBinding)
	}

	var portPoolItemStatusList []*networkextensionv1.PortPoolItemStatus
	var portItemListArr [][]portpoolcache.AllocatedPortItem

	// allocate port from pool cache
	s.poolCache.Lock()
	defer s.poolCache.Unlock()
	for _, portEntry := range portEntryList {
		poolKey := getPoolKey(portEntry.poolName, portEntry.poolNamespace)
		var portPoolItemStatus *networkextensionv1.PortPoolItemStatus
		var err error
		// deal with TCP_UDP protocol
		// for TCP_UDP protocol, one container port needs both TCP listener port and UDP listener port
		if portEntry.protocol == constant.PortPoolPortProtocolTCPUDP {
			var cachePortItemMap map[string]portpoolcache.AllocatedPortItem
			portPoolItemStatus, cachePortItemMap, err = s.poolCache.AllocateAllProtocolPortBinding(poolKey)
			if err != nil {
				s.cleanAllocatedResource(portItemListArr)
				return nil, fmt.Errorf("allocate protocol %s port from pool %s failed, err %s",
					portEntry.protocol, poolKey, err.Error())
			}
			var tmpPortItemList []portpoolcache.AllocatedPortItem
			for _, cachePortItem := range cachePortItemMap {
				tmpPortItemList = append(tmpPortItemList, cachePortItem)
			}
			portItemListArr = append(portItemListArr, tmpPortItemList)
			portPoolItemStatusList = append(portPoolItemStatusList, portPoolItemStatus)
		} else {
			// deal with TCP protocol and UDP protocol
			var cachePortItem portpoolcache.AllocatedPortItem
			portPoolItemStatus, cachePortItem, err = s.poolCache.AllocatePortBinding(poolKey, portEntry.protocol)
			if err != nil {
				s.cleanAllocatedResource(portItemListArr)
				return nil, fmt.Errorf("allocate protocol %s port from pool %s failed, err %s",
					portEntry.protocol, poolKey, err.Error())
			}
			portItemListArr = append(portItemListArr, []portpoolcache.AllocatedPortItem{cachePortItem})
			portPoolItemStatusList = append(portPoolItemStatusList, portPoolItemStatus)
		}
	}

	// patch annotations
	var retPatches []PatchOperation
	annotationPortsPatch, err := s.generatePortsAnnotationPatch(
		pod, portPoolItemStatusList, portItemListArr, portEntryList)
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
	return retPatches, nil
}

// patch pod by port binding object
func (s *Server) patchPodByBinding(
	pod *k8scorev1.Pod, portBinding *networkextensionv1.PortBinding) ([]PatchOperation, error) {

	// patch annotations
	var retPatches []PatchOperation
	annotationPortsPatch, err := s.generatePortsAnnotationPatchByBinding(pod, portBinding)
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
func (s *Server) generatePortsAnnotationPatchByBinding(
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
		var vipList []string
		for _, lbObj := range binding.PoolItemLoadBalancers {
			vipList = append(vipList, lbObj.IPs...)
		}
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
func (s *Server) generatePortsAnnotationPatch(pod *k8scorev1.Pod,
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
			}
			generatedPortList = append(generatedPortList, tmpPort)
		}
	}
	portValues, err := json.Marshal(generatedPortList)
	if err != nil {
		return PatchOperation{}, err
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
		var vipList []string
		for _, lbObj := range portPoolItemStatusList[index].PoolItemLoadBalancers {
			vipList = append(vipList, lbObj.IPs...)
		}
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
