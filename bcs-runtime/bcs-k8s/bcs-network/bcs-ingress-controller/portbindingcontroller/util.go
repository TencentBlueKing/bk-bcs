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

package portbindingcontroller

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	netpkgcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8scorev1 "k8s.io/api/core/v1"
)

// if port binding is expired
func isPortBindingExpired(portBinding *networkextensionv1.PortBinding) (bool, error) {
	keepTimeStr, ok := portBinding.Annotations[networkextensionv1.PortPoolBindingAnnotationKeyKeepDuration]
	if !ok {
		// always return true when no keeptime annotation
		return true, nil
	}
	keepDuration, err := time.ParseDuration(keepTimeStr)
	if err != nil {
		return false, fmt.Errorf("parse keep duration string %s failed, err %s", keepTimeStr, err.Error())
	}
	updateTime, err := netpkgcommon.ParseTimeString(portBinding.Status.UpdateTime)
	if err != nil {
		return false, fmt.Errorf("parse update time string %s failed, err %s",
			portBinding.Status.UpdateTime, err.Error())
	}
	if time.Now().After(updateTime.Add(keepDuration)) {
		return true, nil
	}
	return false, nil
}

// check for pod annotation
func checkPortPoolAnnotationForPod(pod *k8scorev1.Pod) bool {
	_, ok := pod.Annotations[constant.AnnotationForPortPool]
	return ok
}

// get pod host port by container port
func getPodHostPortByPort(pod *k8scorev1.Pod, port int32) int32 {
	for _, container := range pod.Spec.Containers {
		for _, cp := range container.Ports {
			if cp.ContainerPort == port {
				return cp.HostPort
			}
		}
	}
	return 0
}
