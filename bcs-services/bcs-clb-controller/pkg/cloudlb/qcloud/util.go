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
 *
 */

package qcloud

import (
	loadbalance "bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/network/v1"
	"strings"
)

// GetClusterIDPostfix get postfix from cluster id
func GetClusterIDPostfix(clusterid string) string {
	if len(clusterid) == 0 {
		return ""
	}
	strs := strings.Split(clusterid, "-")
	if len(strs) == 1 {
		return clusterid
	}
	return strs[len(strs)-1]
}

// GetBackendsSegment get segment from a big string slice
func GetBackendsSegment(strs []*loadbalance.Backend, cur, segmentLen int) []*loadbalance.Backend {
	if len(strs) == 0 || cur < 0 || segmentLen < 0 || cur > len(strs) {
		return nil
	}
	ret := make([]*loadbalance.Backend, 0)
	realLen := segmentLen
	if cur+realLen > len(strs) {
		realLen = len(strs) - cur
	}
	ret = append(ret, strs[cur:cur+realLen]...)
	return ret
}
