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
	"fmt"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func getPortEnvValue(startPort, endPort int, vipList []string) string {
	portString := strconv.Itoa(startPort)
	if endPort > startPort {
		portString = portString + "-" + strconv.Itoa(endPort)
	}
	var vipString string
	if len(vipList) == 1 {
		vipString = vipList[0]
	} else {
		vipString = strings.Join(vipList, ",")
	}
	return vipString + ":" + portString
}

func getLbIDFromRegionID(lbIDStr string) (string, error) {
	var err error
	var lbID string
	if strings.Contains(lbIDStr, constant.DelimiterForLbID) {
		_, lbID, err = common.GetLbRegionAndName(lbIDStr)
		if err != nil {
			return "", fmt.Errorf("lbIDStr %s is invalid", lbIDStr)
		}
		return lbID, nil
	}
	return lbIDStr, nil
}

func getPoolKey(name, ns string) string {
	return fmt.Sprintf("%s/%s", name, ns)
}

func getPoolPortKey(poolNs, poolName, protocol string, port int) string {
	return fmt.Sprintf("%s/%s/%s/%d", poolNs, poolName, protocol, port)
}

func parsePoolKey(key string) (string, string, error) {
	if !strings.Contains(key, "/") {
		return "", "", fmt.Errorf("invaid pool key %s", key)
	}
	strs := strings.Split(key, "/")
	if len(strs) != 2 {
		return "", "", fmt.Errorf("invaid pool key %s", key)
	}
	return strs[0], strs[1], nil
}

func isProtocolValid(protocol string) bool {
	switch protocol {
	case constant.ProtocolTCP, constant.ProtocolUDP, constant.PortPoolPortProtocolTCPUDP,
		constant.ProtocolTCPSSL:
		return true
	default:
		return false
	}
}

func isPortBindingKeepDurationExisted(portBinding *networkextensionv1.PortBinding) bool {
	if portBinding == nil {
		return false
	}
	_, ok := portBinding.Annotations[networkextensionv1.PortPoolBindingAnnotationKeyKeepDuration]
	return ok
}

func isKeepDurationExisted(anno map[string]string) bool {
	if anno == nil {
		return false
	}
	_, ok := anno[networkextensionv1.PortPoolBindingAnnotationKeyKeepDuration]
	return ok
}
