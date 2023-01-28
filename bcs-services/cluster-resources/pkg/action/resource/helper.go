/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package resource

import (
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

// ScalableResKinds 允许扩缩容的资源类型
var ScalableResKinds = []string{resCsts.Deploy, resCsts.STS, resCsts.GDeploy, resCsts.GSTS}

// ReschedulableResKinds 允许重新调度的资源类型（不含 Pod）
var ReschedulableResKinds = []string{resCsts.Deploy, resCsts.STS, resCsts.GDeploy, resCsts.GSTS}

func isScalable(kind string) bool {
	return slice.StringInSlice(kind, ScalableResKinds)
}

func isReschedulable(kind string) bool {
	return slice.StringInSlice(kind, ReschedulableResKinds)
}
