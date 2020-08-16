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
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// GetListenerName generate listener name with lb id and port number
func GetListenerName(lbID string, port int) string {
	return lbID + "-" + strconv.Itoa(port)
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
		return nil, fmt.Errorf("get stateful set pod index failed from podName %s, err %s", podName, err.Error())
	}
}