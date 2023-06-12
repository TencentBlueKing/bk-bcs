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

package task

import (
	"strings"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// Passwd flag
var Passwd = []string{"password", "passwd"}

func strContains(ipList []string, ip string) bool {
	for i := range ipList {
		if strings.EqualFold(ipList[i], ip) {
			return true
		}
	}
	return false
}

func hiddenTaskPassword(task *proto.Task) {
	if task != nil && len(task.Steps) > 0 {
		for i := range task.Steps {
			for k := range task.Steps[i].Params {
				if k == cloudprovider.BkSopsTaskUrlKey.String() {
					continue
				}
				delete(task.Steps[i].Params, k)
			}
		}
	}

	if task != nil && len(task.CommonParams) > 0 {
		for k, v := range task.CommonParams {
			if utils.StringInSlice(strings.ToLower(k), Passwd) || utils.StringContainInSlice(v, Passwd) {
				delete(task.CommonParams, k)
			}
		}
	}
}
