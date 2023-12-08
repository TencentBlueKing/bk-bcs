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

package middleware

import (
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
)

type applicationSort []v1alpha1.Application

// Len defines the length of application sort
func (p applicationSort) Len() int {
	return len(p)
}

// Less defines the less of application sort
func (p applicationSort) Less(i, j int) bool {
	collectI := p[i].Annotations[common.ApplicationCollectAnnotation]
	collectJ := p[j].Annotations[common.ApplicationCollectAnnotation]
	if collectI != "" && collectJ == "" {
		return true
	}
	if collectI == "" && collectJ != "" {
		return false
	}
	return p[j].CreationTimestamp.Before(&p[i].CreationTimestamp)
}

// Swap defines the swap of application sort
func (p applicationSort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
