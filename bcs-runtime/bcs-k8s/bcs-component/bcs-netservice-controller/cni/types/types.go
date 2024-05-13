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
 */

// Package cni xxx
package cni

import "github.com/containernetworking/cni/pkg/types"

const (
	// ToPodRulePriority ip rules priority and leave 512 gap for future
	ToPodRulePriority = 512
	// FromPodRulePriority 1024 is reserved for (ip rule not to <vpc's subnet> table main)
	FromPodRulePriority = 1536

	// VethPrefix xxx
	VethPrefix = "eni"
)

// K8SArgs is the valid CNI_ARGS used for Kubernetes
type K8SArgs struct {
	types.CommonArgs

	// K8S_POD_NAME is pod's name
	K8S_POD_NAME types.UnmarshallableString

	// K8S_POD_NAMESPACE is pod's namespace
	K8S_POD_NAMESPACE types.UnmarshallableString

	// K8S_POD_INFRA_CONTAINER_ID is pod's container id
	K8S_POD_INFRA_CONTAINER_ID types.UnmarshallableString
}
