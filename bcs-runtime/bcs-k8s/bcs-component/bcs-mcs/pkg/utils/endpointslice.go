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

package utils

import discoveryv1beta1 "k8s.io/api/discovery/v1beta1"

//FindNeedDeleteEndpointSlice find need delete endpoint slice
func FindNeedDeleteEndpointSlice(newEndpointSlices []*discoveryv1beta1.EndpointSlice, oldEndpointSlices []discoveryv1beta1.EndpointSlice) []discoveryv1beta1.EndpointSlice {
	needDeleteEndpointSlices := make([]discoveryv1beta1.EndpointSlice, 0)
	if len(oldEndpointSlices) == 0 {
		return nil
	}
	for _, oldEndpointSlice := range oldEndpointSlices {
		find := false
	innerLoop:
		for _, newEndpointSlice := range newEndpointSlices {
			generateEndpointSliceName := GenerateEndpointSliceName(newEndpointSlice.Name, newEndpointSlice.Labels[ConfigClusterLabel])
			if generateEndpointSliceName == oldEndpointSlice.Name {
				find = true
				break innerLoop
			}
		}
		if !find {
			needDeleteEndpointSlices = append(needDeleteEndpointSlices, oldEndpointSlice)
		}
	}
	return needDeleteEndpointSlices
}
