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

package webhookserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/pkg/errors"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/portpoolcache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
)

// PatchOperation struct for k8s webhook patch
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// entry for allocate cloud loadbalancer port
type portEntry struct {
	PoolNamespace string `json:"poolNamespace,omitempty"`
	PoolName      string `json:"poolName,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
	Port          int    `json:"port,omitempty"`
	HostPort      bool   `json:"hostPort,omitempty"`
	ItemName      string `json:"itemName,omitempty"`
}

// combine container port info into port entry format
func getPortEntryListFromPod(pod *k8scorev1.Pod, annotationPorts []*annotationPort) ([]*portEntry, error) {
	var retPorts []*portEntry
	for _, port := range annotationPorts {
		tmpEntry := &portEntry{
			PoolNamespace: port.poolNamespace,
			PoolName:      port.poolName,
			Protocol:      port.protocol,
			HostPort:      port.hostPort,
			ItemName:      port.itemName,
		}
		if len(port.poolNamespace) == 0 {
			tmpEntry.PoolNamespace = pod.GetNamespace()
		}
		found := false
		for _, container := range pod.Spec.Containers {
			for _, containerPort := range container.Ports {
				if port.portIntOrStr == containerPort.Name {
					found = true
					tmpEntry.Port = int(containerPort.ContainerPort)
					if len(port.protocol) == 0 {
						tmpEntry.Protocol = string(containerPort.Protocol)
					}
					break
				}
				portNumber, err := strconv.Atoi(port.portIntOrStr)
				if err != nil {
					continue
				}
				if int32(portNumber) == containerPort.ContainerPort {
					found = true
					tmpEntry.Port = int(containerPort.ContainerPort)
					if len(port.protocol) == 0 {
						tmpEntry.Protocol = string(containerPort.Protocol)
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
	if err := s.handlePortAnnotationChanged(portList, portBinding); err != nil {
		return nil, err
	}
	return portBinding, nil
}

func (s *Server) handlePortAnnotationChanged(portList []*portEntry, portBinding *networkextensionv1.PortBinding) error {
	portAnnotationChanged := false
	rsPortMap := make(map[string]string)
	for _, item := range portBinding.Spec.PortBindingList {
		key := getPoolPortKey(item.PoolNamespace, item.PoolName, item.Protocol, item.RsStartPort)
		rsPortMap[key] = item.PoolItemName
	}
	// 比较原有PortBinding(通过旧Pod创建)和新Pod注解（用户分配的端口）是否一致， 如果用户更新了分配端口/协议/item等，需要删除PortBinding重建。
	lenPodPortList := 0
	for _, port := range portList {
		if port.Protocol == constant.PortPoolPortProtocolTCPUDP {
			lenPodPortList += 2
			key := getPoolPortKey(port.PoolNamespace, port.PoolName, constant.ProtocolTCP, port.Port)
			// 如果key不存在， 或item名称不一致（用户可能未指定item）
			if itemName, ok := rsPortMap[key]; !ok || (port.ItemName != "" && port.ItemName != itemName) {
				portAnnotationChanged = true
				break
			}

			key = getPoolPortKey(port.PoolNamespace, port.PoolName, constant.ProtocolUDP, port.Port)
			if itemName, ok := rsPortMap[key]; !ok || (port.ItemName != "" && port.ItemName != itemName) {
				portAnnotationChanged = true
				break
			}
		} else {
			lenPodPortList++
			key := getPoolPortKey(port.PoolNamespace, port.PoolName, port.Protocol, port.Port)
			if itemName, ok := rsPortMap[key]; !ok || (port.ItemName != "" && port.ItemName != itemName) {
				portAnnotationChanged = true
				break
			}
		}
	}
	if lenPodPortList != len(portBinding.Spec.PortBindingList) {
		portAnnotationChanged = true
	}

	if portAnnotationChanged {
		blog.Warnf("pod '%s/%s' annotation '%s' changed, need to recreate PortBinding. PortBinding: %s, pod: %s ",
			portBinding.GetNamespace(), portBinding.GetName(), constant.AnnotationForPortPoolPorts,
			common.ToJsonString(portBinding.Spec.PortBindingList), common.ToJsonString(portList))
		// 移除portBinding上的keepDuration注解，尽快删除
		if err := s.removePortBindingAnnotation(portBinding); err != nil {
			return errors.Wrapf(err, "remove portbinding '%s/%s' annotations failed, err: %s",
				portBinding.GetNamespace(), portBinding.GetName(), err.Error())
		}
		if err := s.k8sClient.Delete(context.Background(), portBinding, &client.DeleteOptions{}); err != nil {
			return errors.Wrapf(err, "delete portbinding '%s/%s' failed",
				portBinding.GetName(), portBinding.GetNamespace())
		}
		return nil
	}
	return nil
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
	return retPatches, nil
}

// patch pod by port binding object
func (s *Server) patchPodByBinding(
	pod *k8scorev1.Pod, portBinding *networkextensionv1.PortBinding) ([]PatchOperation, error) {
	blog.Infof("pod '%s/%s' reused portbinding do port inject", pod.GetNamespace(), pod.GetName())

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
		portpool := &networkextensionv1.PortPool{}
		if err1 := s.k8sClient.Get(context.Background(), k8stypes.NamespacedName{
			Namespace: poolNamespace,
			Name:      poolName,
		}, portpool); err1 != nil {
			return PatchOperation{}, fmt.Errorf("get port pool[%s/%s] failed, err %s", poolNamespace, poolName, err1.Error())
		}
		var poolItem *networkextensionv1.PortPoolItem = nil
		for _, item := range portpool.Spec.PoolItems {
			if item.ItemName == portPoolItemStatusList[index].ItemName {
				poolItem = item
				break
			}
		}
		if poolItem == nil {
			return PatchOperation{}, fmt.Errorf("port pool item[%s/%s/%s] not found", poolNamespace, poolName,
				portPoolItemStatusList[index].ItemName)
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
				RsStartPort:           portEntry.Port,
				HostPort:              portEntry.HostPort,
				External:              portPoolItemStatusList[index].External,
				UptimeCheck:           poolItem.UptimeCheck,
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
				Name:  constant.EnvVIPsPrefixForPortPoolPort + item.Protocol + "_" + strconv.Itoa(portEntry.Port),
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
