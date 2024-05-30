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

package calchandler

import (
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/calculator"
)

// ClusterRate defines the instance
type ClusterRate struct {
	OriginalRate   *PackingRate `json:"originalRate"`
	OptimizedRate  *PackingRate `json:"optimizedRate"`
	OptimizedNodes []string     `json:"optimizedNodes"`
}

// PackingRate defines the instance
type PackingRate struct {
	TotalRate *RateObj `json:"totalRate"`

	NodePods        map[string][]*calculator.PodItem `json:"nodePods"`
	NodePackingRate map[string]*RateObj              `json:"nodePackingRate"`
}

// RateObj defines the instance
type RateObj struct {
	MemVal      float64 `json:"memVal"`
	MemCapacity float64 `json:"memCapacity"`
	CpuVal      float64 `json:"cpuVal"`
	CpuCapacity float64 `json:"cpuCapacity"`

	Mem float64 `json:"mem"`
	Cpu float64 `json:"cpu"`
}
