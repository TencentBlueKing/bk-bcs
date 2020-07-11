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

package util

import "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"

func GetMesosAgentInnerIP(attributes []*mesos.Attribute) (string, bool) {
	ip := ""
	ok := false

	for _, attribute := range attributes {
		if attribute.GetName() == "InnerIP" {
			ip = attribute.Text.GetValue()
			ok = true
			break
		}
	}

	return ip, ok
}

func ParseMesosResources(resources []*mesos.Resource) (float64, float64, float64, string) {
	var cpus, mem, disk float64
	var port string
	for _, res := range resources {
		if res.GetName() == "cpus" {
			cpus += *res.GetScalar().Value
		}
		if res.GetName() == "mem" {
			mem += *res.GetScalar().Value
		}
		if res.GetName() == "disk" {
			disk += *res.GetScalar().Value
		}
		if res.GetName() == "ports" {
			port = res.GetRanges().String()
		}
	}

	return cpus, mem, disk, port
}
