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

package conflicthandler

import (
	"strings"

	mapset "github.com/deckarep/golang-set"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
)

func getRegionLBID(lbID, defaultRegion string) (string, error) {
	region, ID, err := common.GetLbRegionAndName(lbID)
	if err != nil {
		blog.Error(err.Error())
		return "", err
	}
	if region == "" {
		region = defaultRegion
	}
	return common.BuildRegionName(region, ID), nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isProtocolConflict return true if conflict
func isProtocolConflict(isTCPUDPReuse bool, protocols, otherProtocols []string) bool {
	if isTCPUDPReuse == false {
		return true
	}

	// only TCP & UDP reuse is valid
	availableProtocolSet := mapset.NewThreadUnsafeSet()
	availableProtocolSet.Add(constant.ProtocolUDP)
	availableProtocolSet.Add(constant.ProtocolTCP)

	for _, protocol := range protocols {
		protocolUpper := strings.ToUpper(protocol)
		if !availableProtocolSet.Contains(protocolUpper) {
			return true
		}
		availableProtocolSet.Remove(protocolUpper)
	}

	for _, protocol := range otherProtocols {
		protocolUpper := strings.ToUpper(protocol)
		if !availableProtocolSet.Contains(protocolUpper) {
			return true
		}
		availableProtocolSet.Remove(protocolUpper)
	}

	return false
}
