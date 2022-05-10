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

package common

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

// GetLbRegionAndName from {region}:{lbID}
func GetLbRegionAndName(lbName string) (string, string, error) {
	if a, err := arn.Parse(lbName); err == nil {
		return a.Region, lbName, nil
	}
	// for lb name without region, we use default region
	if !strings.Contains(lbName, constant.DelimiterForLbID) {
		return "", lbName, nil
	}
	idStrs := strings.Split(lbName, constant.DelimiterForLbID)
	if len(idStrs) != 2 {
		return "", "", fmt.Errorf("lb name %s is invalid", lbName)
	}
	return idStrs[0], idStrs[1], nil
}

// example: arn:aws:elasticloadbalancing:us-west-1:1234567:loadbalancer/net/name/xxx
// return: region-lb-name
func tryParseARNFromLbID(lbID string) string {
	if a, err := arn.Parse(lbID); err == nil {
		names := strings.Split(a.Resource, "/")
		if len(names) != 4 {
			return ""
		}
		return a.Region + "-" + names[2]
	}
	return ""
}

// GetListenerName generate listener name with lb id and port number
func GetListenerName(lbID string, port int) string {
	if a := tryParseARNFromLbID(lbID); len(a) != 0 {
		lbID = a
	}
	return lbID + "-" + strconv.Itoa(port)
}

// GetListenerNameWithProtocol generate listener key with lbid, protocol and port number
func GetListenerNameWithProtocol(lbID, protocol string, startPort, endPort int) string {
	if a := tryParseARNFromLbID(lbID); len(a) != 0 {
		lbID = a
	}
	if endPort <= 0 {
		return lbID + "-" + strings.ToLower(protocol) + "-" + strconv.Itoa(startPort)
	}
	return lbID + "-" + strings.ToLower(protocol) + "-" + strconv.Itoa(startPort) + "-" + strconv.Itoa(endPort)
}

// GetSegmentListenerNameWithProtocol generate segment listener name by protocol
func GetSegmentListenerNameWithProtocol(lbID, protocol string, startPort, endPort int) string {
	if a := tryParseARNFromLbID(lbID); len(a) != 0 {
		lbID = a
	}
	return lbID + "-" + strings.ToLower(protocol) + "-" + strconv.Itoa(startPort) + "-" + strconv.Itoa(endPort)
}

// GetSegmentListenerName generate listener for port segment
func GetSegmentListenerName(lbID string, startPort, endPort int) string {
	if a := tryParseARNFromLbID(lbID); len(a) != 0 {
		lbID = a
	}
	return lbID + "-" + strconv.Itoa(startPort) + "-" + strconv.Itoa(endPort)
}

// GetNamespacedNameKey get key by name and namespace
func GetNamespacedNameKey(name, ns string) string {
	return name + "/" + ns
}

// GetPortPoolListenerLabelKey get key for port pool listener label
// example: pool1/md5(item1)
// because item1 is an anomaly string, so we use md5 to encode it
func GetPortPoolListenerLabelKey(portPoolName, itemName string) string {
	return portPoolName + "/" + fmt.Sprintf("%x", (md5.Sum([]byte(itemName))))
}
