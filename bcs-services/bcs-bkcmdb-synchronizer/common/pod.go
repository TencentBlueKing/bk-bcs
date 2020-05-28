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

package common

import (
	"fmt"
)

const (
	// ClusterTypeK8S k8s type cluster
	ClusterTypeK8S = "k8s"
	// ClusterTypeMesos mesos type cluster
	ClusterTypeMesos = "mesos"

	// K8SNetworkTypeCni network type for k8s
	K8SNetworkTypeCni = "cni"
	// K8SNetworkModeCni network mode for k8s
	K8SNetworkModeCni = "cni"
	// K8SNetworkModeHost network mode for host
	K8SNetworkModeHost = "host"

	// MesosNetworkModeBridge bridge mode for mesos network
	MesosNetworkModeBridge = "bridge"
	// MesosNetworkModeHost host mode for mesos network
	MesosNetworkModeHost = "host"
	// MesosNetworkTypeCnm cnm network
	MesosNetworkTypeCnm = "cnm"
	// MesosNetworkTypeCni cni network
	MesosNetworkTypeCni = "cni"
)

// Labels pod labels
type Labels map[string]string

// Annotations pod annotations
type Annotations map[string]string

// ContainerData container data
type ContainerData struct {
	Data   []string `json:"containers"`
	Status []string `json:"containerStatuses"`
}

// Pod struct for both mesos taskgroup and k8s pod
type Pod struct {
	BizID          int64  `json:"bk_biz_id" mapstructure:"bk_biz_id"`
	ModuleID       int64  `json:"bk_module_id" mapstructure:"bk_module_id"`
	CloudID        int64  `json:"bk_cloud_id" mapstructure:"bk_cloud_id"`
	HostInnerIP    string `json:"bk_host_innerip" mapstructure:"bk_host_innerip"`
	PodName        string `json:"bk_pod_name" mapstructure:"bk_pod_name"`
	PodNamespace   string `json:"bk_pod_namespace" mapstructure:"bk_pod_namespace"`
	PodCluster     string `json:"bk_pod_cluster" mapstructure:"bk_pod_cluster"`
	PodClusterType string `json:"bk_pod_clustertype" mapstructure:"bk_pod_clustertype"`
	PodUUID        string `json:"bk_pod_uuid" mapstructure:"bk_pod_uuid"`
	WorkloadType   string `json:"bk_pod_workloadtype" mapstructure:"bk_pod_workloadtype"`
	WorkloadName   string `json:"bk_pod_workloadname" mapstructure:"bk_pod_workloadname"`
	PodLabels      string `json:"bk_pod_labels" mapstructure:"bk_pod_labels"`
	PodAnnotations string `json:"bk_pod_annotations" mapstructure:"bk_pod_annotations"`
	PodIP          string `json:"bk_pod_ip" mapstructure:"bk_pod_ip"`
	PodNetworkMode string `json:"bk_pod_networkmode" mapstructure:"bk_pod_networkmode"`
	PodNetworkType string `json:"bk_pod_networktype" mapstructure:"bk_pod_networktype"`
	PodContainers  string `json:"bk_pod_containers" mapstructure:"bk_pod_containers"`
	PodVolumes     string `json:"bk_pod_volumes" mapstructure:"bk_pod_volumes"`
	PodStatus      string `json:"bk_pod_status" mapstructure:"bk_pod_status"`
	PodCreateTime  string `json:"bk_pod_create_time" mapstructure:"bk_pod_create_time"`
	PodStartTime   string `json:"bk_pod_start_time" mapstructure:"bk_pod_start_time"`
}

// ToMapInterface to format map[string]interface{}
func (p *Pod) ToMapInterface() map[string]interface{} {
	ret := make(map[string]interface{})
	ret["bk_biz_id"] = p.BizID
	ret["bk_module_id"] = p.ModuleID
	ret["bk_cloud_id"] = p.CloudID
	ret["bk_host_innerip"] = p.HostInnerIP
	ret["bk_pod_name"] = p.PodName
	ret["bk_pod_namespace"] = p.PodNamespace
	ret["bk_pod_cluster"] = p.PodCluster
	ret["bk_pod_clustertype"] = p.PodClusterType
	ret["bk_pod_uuid"] = p.PodUUID
	ret["bk_pod_workloadtype"] = p.WorkloadType
	ret["bk_pod_workloadname"] = p.WorkloadName
	ret["bk_pod_labels"] = p.PodLabels
	ret["bk_pod_annotations"] = p.PodAnnotations
	ret["bk_pod_ip"] = p.PodIP
	ret["bk_pod_networkmode"] = p.PodNetworkMode
	ret["bk_pod_networktype"] = p.PodNetworkType
	ret["bk_pod_containers"] = p.PodContainers
	ret["bk_pod_volumes"] = p.PodVolumes
	ret["bk_pod_status"] = p.PodStatus
	ret["bk_pod_create_time"] = p.PodCreateTime
	ret["bk_pod_start_time"] = p.PodStartTime
	return ret
}

// MetadataString string for metadata
func (p *Pod) MetadataString() string {
	return fmt.Sprintf("cluster:%s, ns:%s, name:%s, module:%d", p.PodCluster, p.PodNamespace, p.PodName, p.ModuleID)
}

// GetUpdatedField get updated field
func (p *Pod) GetUpdatedField(updated *Pod) (bool, map[string]interface{}) {
	ret := make(map[string]interface{})
	if p.ModuleID != updated.ModuleID {
		ret["bk_module_id"] = updated.ModuleID
	}
	if p.HostInnerIP != updated.HostInnerIP {
		ret["bk_host_innerip"] = updated.HostInnerIP
	}
	if p.PodLabels != updated.PodLabels {
		ret["bk_pod_labels"] = updated.PodLabels
	}
	if p.PodAnnotations != updated.PodAnnotations {
		ret["bk_pod_annotations"] = updated.PodAnnotations
	}
	if p.PodIP != updated.PodIP {
		ret["bk_pod_ip"] = updated.PodIP
	}
	if p.PodNetworkMode != updated.PodNetworkMode {
		ret["bk_pod_networkmode"] = updated.PodNetworkMode
	}
	if p.PodNetworkType != updated.PodNetworkType {
		ret["bk_pod_networktype"] = updated.PodNetworkType
	}
	if p.PodContainers != updated.PodContainers {
		ret["bk_pod_containers"] = updated.PodContainers
	}
	if p.PodVolumes != updated.PodVolumes {
		ret["bk_pod_volumes"] = updated.PodVolumes
	}
	if p.PodStatus != updated.PodStatus {
		ret["bk_pod_status"] = updated.PodStatus
	}
	if p.PodCreateTime != updated.PodCreateTime {
		ret["bk_pod_create_time"] = updated.PodCreateTime
	}
	if p.PodStartTime != updated.PodStartTime {
		ret["bk_pod_start_time"] = updated.PodStartTime
	}

	if len(ret) == 0 {
		return false, ret
	}
	return true, ret
}

// PodUpdateInfo info to update a pod
type PodUpdateInfo struct {
	Condition map[string]interface{}
	Data      map[string]interface{}
}

// GetDiffPods get different pods between two pods maps
func GetDiffPods(oldPodMap map[string]*Pod, newPodMap map[string]*Pod) (adds []*Pod, updates []*Pod, deletes []*Pod) {
	for k, v := range newPodMap {
		if old, ok := oldPodMap[k]; ok {
			if flag, _ := old.GetUpdatedField(newPodMap[k]); flag {
				updates = append(updates, v)
			}
		} else {
			adds = append(adds, v)
		}
	}

	for k, v := range oldPodMap {
		if _, ok := newPodMap[k]; !ok {
			deletes = append(deletes, v)
		}
	}
	return
}
