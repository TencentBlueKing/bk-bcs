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
	"encoding/json"
	"strings"

	"bk-bcs/bcs-common/common/blog"

	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8SPod k8s pod in storage
type K8SPod struct {
	k8smetav1.ObjectMeta `json:"metadata"`
	Spec                 k8scorev1.PodSpec   `json:"spec"`
	Status               k8scorev1.PodStatus `json:"status"`
}

func convertK8STime(k8sTime string) string {
	output := strings.Replace(k8sTime, "T", " ", -1)
	output = output[0:19]
	return output
}

// ConvertK8SPod convert k8s pod to bk cmdb pod
func ConvertK8SPod(cluster string, kPod *K8SPod) (*Pod, error) {
	newPod := &Pod{}
	newPod.PodName = kPod.GetName()
	newPod.PodNamespace = kPod.GetNamespace()
	newPod.PodCluster = cluster
	newPod.PodClusterType = ClusterTypeK8S
	newPod.PodUUID = string(kPod.GetUID())

	if len(kPod.OwnerReferences) != 0 {
		newPod.WorkloadName = kPod.OwnerReferences[0].Name
		newPod.WorkloadType = kPod.OwnerReferences[0].Kind
	}

	labelBytes, err := json.Marshal(kPod.Labels)
	if err != nil {
		blog.Errorf("k8d pod encoding labels failed, err %s", err.Error())
		return nil, err
	}
	newPod.PodLabels = string(labelBytes)

	annotationBytes, err := json.Marshal(kPod.Annotations)
	if err != nil {
		blog.Errorf("k8d pod encoding annotation failed, err %s", err.Error())
		return nil, err
	}
	newPod.PodAnnotations = string(annotationBytes)

	newPod.PodIP = kPod.Status.PodIP
	newPod.PodStatus = string(kPod.Status.Phase)

	if kPod.Spec.HostNetwork {
		newPod.PodNetworkMode = K8SNetworkModeHost
	} else {
		newPod.PodNetworkMode = K8SNetworkModeCni
	}
	newPod.PodNetworkType = K8SNetworkTypeCni

	var containerData ContainerData
	for _, c := range kPod.Spec.Containers {
		cStr, err := json.Marshal(&c)
		if err != nil {
			blog.Errorf("k8s pod encoding container failed, err %s", err.Error())
			return nil, err
		}
		containerData.Data = append(containerData.Data, string(cStr))
	}
	for _, c := range kPod.Status.ContainerStatuses {
		cStr, err := json.Marshal(&c)
		if err != nil {
			blog.Errorf("k8s pod encoding container status failed, err %s", err.Error())
			return nil, err
		}
		containerData.Status = append(containerData.Status, string(cStr))
	}

	volumesStr, err := json.Marshal(kPod.Spec.Volumes)
	if err != nil {
		blog.Errorf("k8s pod encoding volumes failed, err %s", err.Error())
		return nil, err
	}
	newPod.PodVolumes = string(volumesStr)

	newPod.PodCreateTime = convertK8STime(kPod.CreationTimestamp.String())
	newPod.PodStartTime = convertK8STime(kPod.Status.StartTime.String())

	return newPod, nil
}
