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

package generator

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// SplitRegionLBID get region and lbid from regionLBID
func SplitRegionLBID(regionLBID string) (string, string, error) {
	strs := strings.Split(regionLBID, ":")
	if len(strs) == 1 {
		return "", strs[0], nil
	}
	if len(strs) == 2 {
		return strs[0], strs[1], nil
	}
	return "", "", fmt.Errorf("regionLBID %s invalid", regionLBID)
}

// GetListenerName generate listener name with lb id and port number
func GetListenerName(lbID string, port int) string {
	return lbID + "-" + strconv.Itoa(port)
}

// GetListenerNameWithProtocol generate listener key with lbid, protocol and port number
func GetListenerNameWithProtocol(lbID, protocol string, port int) string {
	return lbID + "-" + strings.ToLower(protocol) + "-" + strconv.Itoa(port)
}

// GetSegmentListenerName generate listener for port segment
func GetSegmentListenerName(lbID string, startPort, endPort int) string {
	return lbID + "-" + strconv.Itoa(startPort) + "-" + strconv.Itoa(endPort)
}

// GetPodIndex get pod index
func GetPodIndex(podName string) (int, error) {
	nameStrs := strings.Split(podName, "-")
	if len(nameStrs) < 2 {
		blog.Errorf("")
	}
	podNumberStr := nameStrs[len(nameStrs)-1]
	podIndex, err := strconv.Atoi(podNumberStr)
	if err != nil {
		blog.Errorf("get stateful set pod index failed from podName %s, err %s", podName, err.Error())
		return -1, fmt.Errorf("get stateful set pod index failed from podName %s, err %s", podName, err.Error())
	}
	return podIndex, nil
}

// GetDiffListeners get diff between two listener arrays
func GetDiffListeners(existedListeners, newListeners []networkextensionv1.Listener) (
	[]networkextensionv1.Listener, []networkextensionv1.Listener,
	[]networkextensionv1.Listener, []networkextensionv1.Listener) {

	existedListenerMap := make(map[string]networkextensionv1.Listener)
	for _, listener := range existedListeners {
		existedListenerMap[listener.GetName()] = listener
	}
	newListenerMap := make(map[string]networkextensionv1.Listener)
	for _, listener := range newListeners {
		newListenerMap[listener.GetName()] = listener
	}

	var adds []networkextensionv1.Listener
	var dels []networkextensionv1.Listener
	var olds []networkextensionv1.Listener
	var news []networkextensionv1.Listener

	for _, listener := range newListeners {
		existedListener, ok := existedListenerMap[listener.GetName()]
		if !ok {
			adds = append(adds, listener)
			continue
		}
		if !reflect.DeepEqual(listener.Spec, existedListener.Spec) {
			olds = append(olds, existedListener)
			news = append(news, listener)
			continue
		}
	}

	for _, listener := range existedListeners {
		_, ok := newListenerMap[listener.GetName()]
		if !ok {
			dels = append(dels, listener)
			continue
		}
	}
	return adds, dels, olds, news
}

// GetSpecLabelSelectorFromMap get spec.selector from k8s runtime.Object
func GetSpecLabelSelectorFromMap(m map[string]interface{}) *k8smetav1.LabelSelector {
	spec, ok := m["spec"]
	if !ok {
		blog.Warnf("no spec")
		return nil
	}
	specMap, ok := spec.(map[string]interface{})
	if !ok {
		blog.Warnf("spec is not map[string]interface")
		return nil
	}
	selector, ok := specMap["selector"]
	if !ok {
		blog.Warnf("has no selector")
		return nil
	}

	selectorBytes, err := json.Marshal(selector)
	if err != nil {
		blog.Warnf("json mashal %+v failed, err %s", selector, err.Error())
		return nil
	}

	selectorObj := &k8smetav1.LabelSelector{}
	err = json.Unmarshal(selectorBytes, selectorObj)
	if err != nil {
		blog.Warnf("json unmashal %s to LabelSelector failed, err %s", selectorObj, err.Error())
		return nil
	}
	return selectorObj
}

// isPodOwner to see whether obj with certain kind and name is owner of pod
func isPodOwner(kind, name string, pod *k8scorev1.Pod) bool {
	if pod == nil {
		return false
	}
	for _, ownerRef := range pod.OwnerReferences {
		if ownerRef.Kind == kind && ownerRef.Name == name {
			return true
		}
	}
	return false
}

// MatchLbStr check region info format
func MatchLbStrWithId(lbId string) bool {
	// should not include space and newline
	if strings.Contains(lbId, "\n") || strings.Contains(lbId, " ") {
		return false
	}

	// match ap-xxxxx:lb-xxxxx
	match, _ := regexp.MatchString(constant.LoadBalanceCheckFormatWithApLbID, lbId)
	if match {
		return true
	}

	// match lb-xxxxx
	match, _ = regexp.MatchString(constant.LoadBalanceCheckFormat, lbId)
	return match
}

// MatchLbStr check region info format
func MatchLbStrWithName(lbName string) bool {
	// should not include space and newline
	if strings.Contains(lbName, "\n") || strings.Contains(lbName, " ") {
		return false
	}

	// match ap-xxxxx:lbname
	match, _ := regexp.MatchString(constant.LoadBalanceCheckFormatWithApLbName, lbName)
	if match {
		return true
	}

	// match lbname
	return lbName != ""
}
