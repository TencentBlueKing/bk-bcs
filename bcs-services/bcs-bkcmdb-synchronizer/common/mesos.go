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
	commtypes "bk-bcs/bcs-common/common/types"
)

func convertMesosTime(k8sTime string) string {
	output := strings.Replace(k8sTime, "T", " ", -1)
	output = output[0:19]
	return output
}

// ConvertMesosPod convert mesos pod to bk cmdb pod
func ConvertMesosPod(mesosPod *commtypes.BcsPodStatus) (*Pod, error) {
	newPod := &Pod{}
	newPod.PodName = mesosPod.Name
	newPod.PodNamespace = mesosPod.NameSpace
	newPod.PodCluster = mesosPod.ClusterName
	newPod.PodClusterType = ClusterTypeMesos
	newPod.PodUUID = mesosPod.Name
	newPod.WorkloadType = "Application"
	newPod.WorkloadName = mesosPod.RcName

	labelStr, err := json.Marshal(mesosPod.ObjectMeta.Labels)
	if err != nil {
		blog.Errorf("mesos taskgroup encoding labels failed, err %s", err.Error())
		return nil, err
	}
	newPod.PodLabels = string(labelStr)
	newPod.PodIP = mesosPod.PodIP
	newPod.PodNetworkMode = mesosPod.ContainerStatuses[0].Network
	if strings.ToLower(newPod.PodNetworkMode) == MesosNetworkModeBridge ||
		strings.ToLower(newPod.PodNetworkMode) == MesosNetworkModeHost {
		newPod.PodNetworkType = MesosNetworkTypeCnm
	} else {
		newPod.PodNetworkType = MesosNetworkTypeCni
	}

	containersStr, err := json.Marshal(mesosPod.ContainerStatuses)
	if err != nil {
		blog.Errorf("mesos taskgroup encoding container statuses failed, err %s", err.Error())
		return nil, err
	}
	newPod.PodContainers = string(containersStr)

	newPod.PodVolumes = ""
	newPod.PodStatus = string(mesosPod.Status)
	newPod.PodCreateTime = convertMesosTime(mesosPod.CreationTimestamp.String())

	return newPod, nil
}
