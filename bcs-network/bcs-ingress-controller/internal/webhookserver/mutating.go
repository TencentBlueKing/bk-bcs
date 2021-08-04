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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/portpoolcache"

	k8scorev1 "k8s.io/api/core/v1"
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
}

func getPortEntryListFromPod(pod *k8scorev1.Pod, annotationPorts []*annotationPort) ([]*portEntry, error) {
	var retPorts []*portEntry
	for _, port := range annotationPorts {
		tmpEntry := &portEntry{
			poolNamespace: port.poolNamespace,
			poolName:      port.poolName,
			protocol:      port.protocol,
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
			return nil, fmt.Errorf("port %s not found with container port name or port number", port.portIntOrStr)
		}
		retPorts = append(retPorts, tmpEntry)
	}
	return retPorts, nil
}

func (s *Server) cleanAllocatedResource(items [][]portpoolcache.AllocatedPortItem) {
	for _, list := range items {
		for _, item := range list {
			s.poolCache.ReleasePortBinding(item.PoolKey, item.PoolItemKey, item.Protocol, item.StartPort, item.EndPort)
		}
	}
}

func (s *Server) mutatingPod(pod *k8scorev1.Pod) ([]PatchOperation, error) {
	if pod.Annotations == nil {
		return nil, fmt.Errorf("pod annotation is nil")
	}
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
	var portPoolItemStatusList []*networkextensionv1.PortPoolItemStatus
	var portItemListArr [][]portpoolcache.AllocatedPortItem

	s.poolCache.Lock()
	defer s.poolCache.Unlock()
	for _, portEntry := range portEntryList {
		poolKey := getPoolKey(portEntry.poolName, portEntry.poolNamespace)
		var portPoolItemStatus *networkextensionv1.PortPoolItemStatus
		var err error
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

	var retPatches []PatchOperation
	annotationPortsPatch, err := s.generatePortsAnnotationPatch(
		pod, portPoolItemStatusList, portItemListArr, portEntryList)
	if err != nil {
		s.cleanAllocatedResource(portItemListArr)
		return nil, fmt.Errorf("generate ports of port pool annotations failed, err %s", err.Error())
	}
	retPatches = append(retPatches, annotationPortsPatch)

	if _, ok := pod.Annotations[constant.AnnotationForPortPoolReadinessGate]; ok {
		readinessGatePatch, err := s.generatePodReadinessGate(pod)
		if err != nil {
			s.cleanAllocatedResource(portItemListArr)
			return nil, fmt.Errorf("generate pod readiness gate failed, err %s", err.Error())
		}
		retPatches = append(retPatches, readinessGatePatch)
	}

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
			portString := strconv.Itoa(item.StartPort)
			if item.EndPort > item.StartPort {
				portString = portString + "-" + strconv.Itoa(item.EndPort)
			}
			var vipString string
			if len(vipList) == 1 {
				vipString = vipList[0]
			} else {
				vipString = strings.Join(vipList, ",")
			}
			envs = append(envs, k8scorev1.EnvVar{
				Name:  constant.EnvVIPsPrefixForPortPoolPort + item.Protocol + "_" + strconv.Itoa(portEntry.port),
				Value: vipString + ":" + portString,
			})
		}
	}
	return PatchOperation{
		Path:  fmt.Sprintf(patchPath, index),
		Op:    envPatchOp,
		Value: envs,
	}
}

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
