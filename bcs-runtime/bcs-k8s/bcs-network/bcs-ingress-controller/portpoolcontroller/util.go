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

package portpoolcontroller

import (
	netextv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
)

// return true if listener's labels are correct
func checkListenerLabels(labels map[string]string, portPoolName, itemName string) bool {
	poolNameLabel := common.GetPortPoolListenerLabelKey(portPoolName, itemName)
	if v, ok := labels[poolNameLabel]; !ok || v != netextv1.LabelValueForPortPoolItemName {
		return false
	}

	if v, ok := labels[netextv1.LabelKeyForOwnerName]; !ok || v != portPoolName {
		return false
	}

	return true
}

func getPortPoolStatus(pool *netextv1.PortPool) string {
	statusReady := true
	for _, ts := range pool.Status.PoolItemStatuses {
		if ts.Status != constant.PortPoolItemStatusReady {
			statusReady = false
			break
		}
	}
	if statusReady {
		return constant.PortPoolStatusReady
	}
	return constant.PortPoolStatusNotReady
}
