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

package v1

import (
	status "github.com/Tencent/bk-bcs/bcs-common/common/types"
	netservicetypes "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
	deploymentType "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

const (
	// BcsStorageDynamicTypeApplication xxx
	BcsStorageDynamicTypeApplication = "application"
	// BcsStorageDynamicTypeProcess xxx
	BcsStorageDynamicTypeProcess = "process"
	// BcsStorageDynamicTypeTaskGroup xxx
	BcsStorageDynamicTypeTaskGroup = "taskgroup"
	// BcsStorageDynamicTypeConfigMap xxx
	BcsStorageDynamicTypeConfigMap = "configmap"
	// BcsStorageDynamicTypeSecret xxx
	BcsStorageDynamicTypeSecret = "secret"
	// BcsStorageDynamicTypeService xxx
	BcsStorageDynamicTypeService = "service"
	// BcsStorageDynamicTypeEndpoint xxx
	BcsStorageDynamicTypeEndpoint = "endpoint"
	// BcsStorageDynamicTypeDeployment xxx
	BcsStorageDynamicTypeDeployment = "deployment"
	// BcsStorageDynamicTypeNamespace xxx
	BcsStorageDynamicTypeNamespace = "namespace"
	// BcsStorageDynamicTypeIPPoolStatic xxx
	BcsStorageDynamicTypeIPPoolStatic = "ippoolstatic"
	// BcsStorageDynamicTypeIPPoolStaticDetail xxx
	BcsStorageDynamicTypeIPPoolStaticDetail = "ippoolstaticdetail"
)

// ApplicationSet xxx
type ApplicationSet struct {
	Data status.BcsReplicaControllerStatus `json:"data"`
}

// ProcessSet xxx
type ProcessSet struct {
	Data status.BcsReplicaControllerStatus `json:"data"`
}

// TaskGroupSet xxx
type TaskGroupSet struct {
	Data status.BcsPodStatus `json:"data"`
}

// ConfigMapSet xxx
type ConfigMapSet struct {
	Data status.BcsConfigMap `json:"data"`
}

// SecretSet xxx
type SecretSet struct {
	Data status.BcsSecret `json:"data"`
}

// ServiceSet xxx
type ServiceSet struct {
	Data status.BcsService `json:"data"`
}

// EndpointSet xxx
type EndpointSet struct {
	Data status.BcsEndpoint `json:"data"`
}

// DeploymentSet xxx
type DeploymentSet struct {
	Data deploymentType.Deployment `json:"data"`
}

// IPPoolStatic is netservice ip pool resources object.
type IPPoolStatic struct {
	// Data includes poolnum/activeip/availableip/reservedip.
	Data netservicetypes.NetStatic `json:"data"`
}

// IPPoolStaticDetail is netservice ip pool resources detail object.
type IPPoolStaticDetail struct {
	// Data includes cluster hosts/available/reserved/active ip pool informations.
	Data []*netservicetypes.NetPool `json:"data"`
}

// ApplicationList xxx
type ApplicationList []*ApplicationSet

// ProcessList xxx
type ProcessList []*ProcessSet

// TaskGroupList xxx
type TaskGroupList []*TaskGroupSet

// ConfigMapList xxx
type ConfigMapList []*ConfigMapSet

// SecretList xxx
type SecretList []*SecretSet

// ServiceList xxx
type ServiceList []*ServiceSet

// EndpointList xxx
type EndpointList []*EndpointSet

// DeploymentList xxx
type DeploymentList []*DeploymentSet

// IPPoolStaticList xxx
type IPPoolStaticList []*IPPoolStatic

// IPPoolStaticDetailList xxx
type IPPoolStaticDetailList []*IPPoolStaticDetail

// Len in list-sort for ApplicationList
func (l ApplicationList) Len() int { return len(l) }

// Less in list-sort for ApplicationList
func (l ApplicationList) Less(i, j int) bool { return l[i].Data.NameSpace > l[j].Data.NameSpace }

// Swap in list-sort for ApplicationList
func (l ApplicationList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

// Len in list-sort for ProcessList
func (l ProcessList) Len() int { return len(l) }

// Less in list-sort for ProcessList
func (l ProcessList) Less(i, j int) bool { return l[i].Data.NameSpace > l[j].Data.NameSpace }

// Swap in list-sort for ProcessList
func (l ProcessList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

// Len in list-sort for TaskGroupList
func (l TaskGroupList) Len() int { return len(l) }

// Less in list-sort for TaskGroupList
func (l TaskGroupList) Less(i, j int) bool { return l[i].Data.NameSpace > l[j].Data.NameSpace }

// Swap in list-sort for TaskGroupList
func (l TaskGroupList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

// Len in list-sort for ConfigMapList
func (l ConfigMapList) Len() int { return len(l) }

// Less in list-sort for ConfigMapList
func (l ConfigMapList) Less(i, j int) bool { return l[i].Data.NameSpace > l[j].Data.NameSpace }

// Swap in list-sort for ConfigMapList
func (l ConfigMapList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

// Len in list-sort for SecretList
func (l SecretList) Len() int { return len(l) }

// Less in list-sort for SecretList
func (l SecretList) Less(i, j int) bool { return l[i].Data.NameSpace > l[j].Data.NameSpace }

// Swap in list-sort for SecretList
func (l SecretList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

// Len in list-sort for ServiceList
func (l ServiceList) Len() int { return len(l) }

// Less in list-sort for ServiceList
func (l ServiceList) Less(i, j int) bool { return l[i].Data.NameSpace > l[j].Data.NameSpace }

// Swap in list-sort for ServiceList
func (l ServiceList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

// Len in list-sort for EndpointList
func (l EndpointList) Len() int { return len(l) }

// Less in list-sort for EndpointList
func (l EndpointList) Less(i, j int) bool { return l[i].Data.NameSpace > l[j].Data.NameSpace }

// Swap in list-sort for EndpointList
func (l EndpointList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

// Len in list-sort for DeploymentList
func (l DeploymentList) Len() int { return len(l) }

// Less in list-sort for DeploymentList
func (l DeploymentList) Less(i, j int) bool {
	return l[i].Data.ObjectMeta.NameSpace > l[j].Data.ObjectMeta.NameSpace
}

// Swap in list-sort for DeploymentList
func (l DeploymentList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
