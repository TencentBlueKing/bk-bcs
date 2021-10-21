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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8SContainerPort simplified for k8s container port
type K8SContainerPort struct {
	Name          string `json:"name"`
	HostPort      int32  `json:"hostPort"`
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol"`
	HostIP        string `json:"hostIP"`
}

// K8SContainer simplified data for k8s container
type K8SContainer struct {
	Name      string                         `json:"name"`
	Image     string                         `json:"image"`
	Ports     []K8SContainerPort             `json:"ports,omitempty"`
	Resources k8scorev1.ResourceRequirements `json:"resources,omitempty"`
}

// K8SPodSpec simplified data of k8s pod spec
type K8SPodSpec struct {
	Containers   []K8SContainer      `json:"containers,omitempty"`
	DNSPolicy    k8scorev1.DNSPolicy `json:"dnsPolicy,omitempty"`
	NodeSelector map[string]string   `json:"nodeSelector"`
	NodeName     string              `json:"nodeName"`
	HostNetwork  bool                `json:"hostNetwork"`
	Hostname     string              `json:"hostname"`
}

// K8SContainerStatus simplified data of container status
type K8SContainerStatus struct {
	Name         string `json:"name"`
	Ready        bool   `json:"ready"`
	RestartCount int32  `json:"restartCount"`
	ContainerID  string `json:"containerID"`
}

// K8SPodStatus simplified data of k8s pod status
type K8SPodStatus struct {
	Phase             string               `json:"phase"`
	Reason            string               `json:"reason"`
	HostIP            string               `json:"hostIP"`
	PodIP             string               `json:"podIP"`
	StartTime         *k8smetav1.Time      `json:"startTime"`
	ContainerStatuses []K8SContainerStatus `json:"containerStatuses"`
}

// K8SPod k8s pod in storage
type K8SPod struct {
	ID                   string `json:"id"`
	k8smetav1.ObjectMeta `json:"metadata"`
	Spec                 K8SPodSpec   `json:"spec"`
	Status               K8SPodStatus `json:"status"`
}

// GetCreationTimestamp get creation timestamp
func (kp *K8SPod) GetCreationTimestamp() time.Time {
	return kp.ObjectMeta.CreationTimestamp.Time
}

// SetCreationTimestamp set creation timestamp
func (kp *K8SPod) SetCreationTimestamp(t time.Time) {
	kp.ObjectMeta.CreationTimestamp.Time = t
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

	if kPod.Annotations != nil {
		annotationBytes, err := json.Marshal(kPod.Annotations)
		if err != nil {
			blog.Errorf("k8d pod encoding annotation failed, err %s", err.Error())
			return nil, err
		}
		newPod.PodAnnotations = string(annotationBytes)
	} else {
		newPod.PodAnnotations = "{}"
	}

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

	newPod.PodCreateTime = convertK8STime(kPod.CreationTimestamp.String())
	if kPod.Status.StartTime != nil {
		newPod.PodStartTime = convertK8STime(kPod.Status.StartTime.String())
	}

	return newPod, nil
}
