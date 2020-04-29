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
 *
 */

package custom

import (
	"bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-k8s/bcs-k8s-driver/client"
	"bk-bcs/bcs-k8s/bcs-k8s-driver/kubedriver/options"
	"fmt"

	restful "github.com/emicklei/go-restful"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

type ClusterResourceStatus struct {
	Capacity ResourceStatus
	Limit    ResourceStatus
	Request  ResourceStatus
}

type ResourceStatus struct {
	Cpu    float64
	Memory float64
}

type ClusterResourceAPIHandler struct {
	clientSet *kubernetes.Clientset
}

func (cph *ClusterResourceAPIHandler) Handler(request *restful.Request, response *restful.Response) {
	nodes, err := cph.clientSet.CoreV1().Nodes().List(metav1.ListOptions{})

	if err != nil {
		CustomServerErrorResponse(response, "Get node list failed")
		return
	}

	bcsClusterResourceStatus := types.BcsClusterResource{
		DiskTotal: float64(0),
		DiskUsed:  float64(0),
	}

	allPods := cph.FetchAllPods()
	if allPods == nil {
		CustomServerErrorResponse(response, "Get pod list failed")
		return
	}
	// agent info
	var agentInfo []types.BcsClusterAgentInfo
OutLoop:
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" && condition.Status != "True" || node.ObjectMeta.Labels["node-role.kubernetes.io/master"] == "true" {
				continue OutLoop
			}
		}
		// compose ip for agent info
		for _, addrInfo := range node.Status.Addresses {
			if addrInfo.Type == "InternalIP" {
				agentInfo = append(agentInfo, types.BcsClusterAgentInfo{
					IP: addrInfo.Address,
				})
			}
		}
		nodesStatus := cph.describeNodeMetrc(&node)
		bcsClusterResourceStatus.MemTotal += nodesStatus.Capacity.Memory
		bcsClusterResourceStatus.CpuTotal += nodesStatus.Capacity.Cpu

	}
	PodsStatus := cph.describePodsMetric(allPods)

	bcsClusterResourceStatus.MemUsed = PodsStatus.Request.Memory
	bcsClusterResourceStatus.CpuUsed = PodsStatus.Request.Cpu
	bcsClusterResourceStatus.Agents = agentInfo

	CustomSuccessResponse(response, "Success", bcsClusterResourceStatus)
	return
}

func (cph *ClusterResourceAPIHandler) Config(KubeMasterURL string, TLSConfig options.TLSConfig) error {
	cph.clientSet = client.NewClientSet(KubeMasterURL, TLSConfig)
	if cph.clientSet == nil {
		return fmt.Errorf("failed to get k8s clientSet")
	}
	return nil
}

func (cph *ClusterResourceAPIHandler) describeNodeMetrc(node *v1.Node) *ClusterResourceStatus {
	allocatable := node.Status.Capacity
	//if len(node.Status.Allocatable) > 0 {
	//	allocatable = node.Status.Allocatable
	//}

	return &ClusterResourceStatus{
		Capacity: ResourceStatus{
			Cpu:    float64(allocatable.Cpu().MilliValue() / 1000),
			Memory: float64(allocatable.Memory().MilliValue()) / (1024 * 1024 * 1024 * 1024),
		},
	}
}

func (cph *ClusterResourceAPIHandler) describePodsMetric(nodeNonTerminatedPodsList *v1.PodList) *ClusterResourceStatus {
	reqs, limits := getPodsTotalRequestsAndLimits(nodeNonTerminatedPodsList)
	cpuReqs, cpuLimits, memoryReqs, memoryLimits := reqs[v1.ResourceCPU], limits[v1.ResourceCPU], reqs[v1.ResourceMemory], limits[v1.ResourceMemory]
	return &ClusterResourceStatus{
		Limit: ResourceStatus{
			Cpu:    float64(cpuLimits.MilliValue() / 1000),
			Memory: float64(memoryLimits.MilliValue()) / (1024 * 1024 * 1024 * 1024),
		},
		Request: ResourceStatus{
			Cpu:    float64(cpuReqs.MilliValue()) / 1000,
			Memory: float64(memoryReqs.MilliValue()) / (1024 * 1024 * 1024 * 1024),
		},
	}
}

func (cph *ClusterResourceAPIHandler) FetchAllPods() (allPodsInNode *v1.PodList) {
	allPodsInNode = &v1.PodList{}
	fieldSelector, err := fields.ParseSelector(
		"status.phase!=" + string(v1.PodSucceeded) + ",status.phase!=" + string(v1.PodFailed))
	if err != nil {
		return nil
	}
	// get all pods
	allPodsInNode, err = cph.clientSet.CoreV1().Pods("").List(metav1.ListOptions{FieldSelector: fieldSelector.String()})
	if err != nil {
		return nil
	}
	return allPodsInNode
}

func PodRequestsAndLimits(pod *v1.Pod) (reqs map[v1.ResourceName]resource.Quantity, limits map[v1.ResourceName]resource.Quantity) {
	reqs, limits = map[v1.ResourceName]resource.Quantity{}, map[v1.ResourceName]resource.Quantity{}
	for _, container := range pod.Spec.Containers {
		for name, quantity := range container.Resources.Requests {
			if value, ok := reqs[name]; !ok {
				reqs[name] = *quantity.Copy()
			} else {
				value.Add(quantity)
				reqs[name] = value
			}
		}
		for name, quantity := range container.Resources.Limits {
			if value, ok := limits[name]; !ok {
				limits[name] = *quantity.Copy()
			} else {
				value.Add(quantity)
				limits[name] = value
			}
		}
	}
	// init containers define the minimum of any resource
	for _, container := range pod.Spec.InitContainers {
		for name, quantity := range container.Resources.Requests {
			value, ok := reqs[name]
			if !ok {
				reqs[name] = *quantity.Copy()
				continue
			}
			if quantity.Cmp(value) > 0 {
				reqs[name] = *quantity.Copy()
			}
		}
		for name, quantity := range container.Resources.Limits {
			value, ok := limits[name]
			if !ok {
				limits[name] = *quantity.Copy()
				continue
			}
			if quantity.Cmp(value) > 0 {
				limits[name] = *quantity.Copy()
			}
		}
	}
	return reqs, limits
}

func getPodsTotalRequestsAndLimits(podList *v1.PodList) (reqs map[v1.ResourceName]resource.Quantity, limits map[v1.ResourceName]resource.Quantity) {
	reqs, limits = map[v1.ResourceName]resource.Quantity{}, map[v1.ResourceName]resource.Quantity{}
	for _, pod := range podList.Items {
		podReqs, podLimits := PodRequestsAndLimits(&pod)
		for podReqName, podReqValue := range podReqs {
			if value, ok := reqs[podReqName]; !ok {
				reqs[podReqName] = *podReqValue.Copy()
			} else {
				value.Add(podReqValue)
				reqs[podReqName] = value
			}
		}
		for podLimitName, podLimitValue := range podLimits {
			if value, ok := limits[podLimitName]; !ok {
				limits[podLimitName] = *podLimitValue.Copy()
			} else {
				value.Add(podLimitValue)
				limits[podLimitName] = value
			}
		}
	}
	return reqs, limits
}
