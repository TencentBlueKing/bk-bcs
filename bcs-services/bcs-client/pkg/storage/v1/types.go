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
	BcsStorageDynamicTypeApplication        = "application"
	BcsStorageDynamicTypeProcess            = "process"
	BcsStorageDynamicTypeTaskGroup          = "taskgroup"
	BcsStorageDynamicTypeConfigMap          = "configmap"
	BcsStorageDynamicTypeSecret             = "secret"
	BcsStorageDynamicTypeService            = "service"
	BcsStorageDynamicTypeEndpoint           = "endpoint"
	BcsStorageDynamicTypeDeployment         = "deployment"
	BcsStorageDynamicTypeNamespace          = "namespace"
	BcsStorageDynamicTypeIPPoolStatic       = "ippoolstatic"
	BcsStorageDynamicTypeIPPoolStaticDetail = "ippoolstaticdetail"
)

type ApplicationSet struct {
	Data status.BcsReplicaControllerStatus `json:"data"`
}

type ProcessSet struct {
	Data status.BcsReplicaControllerStatus `json:"data"`
}

type TaskGroupSet struct {
	Data status.BcsPodStatus `json:"data"`
}

type ConfigMapSet struct {
	Data status.BcsConfigMap `json:"data"`
}

type SecretSet struct {
	Data status.BcsSecret `json:"data"`
}

type ServiceSet struct {
	Data status.BcsService `json:"data"`
}

type EndpointSet struct {
	Data status.BcsEndpoint `json:"data"`
}

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

type ApplicationList []*ApplicationSet
type ProcessList []*ProcessSet
type TaskGroupList []*TaskGroupSet
type ConfigMapList []*ConfigMapSet
type SecretList []*SecretSet
type ServiceList []*ServiceSet
type EndpointList []*EndpointSet
type DeploymentList []*DeploymentSet
type IPPoolStaticList []*IPPoolStatic
type IPPoolStaticDetailList []*IPPoolStaticDetail

// Len in list-sort for ApplicationList
func (l ApplicationList) Len() int           { return len(l) }
// Less in list-sort for ApplicationList
func (l ApplicationList) Less(i, j int) bool { return l[i].Data.NameSpace > l[j].Data.NameSpace }
// Swap in list-sort for ApplicationList
func (l ApplicationList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

// Len in list-sort for ProcessList
func (l ProcessList) Len() int               { return len(l) }
// Less in list-sort for ProcessList
func (l ProcessList) Less(i, j int) bool     { return l[i].Data.NameSpace > l[j].Data.NameSpace }
// Swap in list-sort for ProcessList
func (l ProcessList) Swap(i, j int)          { l[i], l[j] = l[j], l[i] }

// Len in list-sort for TaskGroupList
func (l TaskGroupList) Len() int             { return len(l) }
// Less in list-sort for TaskGroupList
func (l TaskGroupList) Less(i, j int) bool   { return l[i].Data.NameSpace > l[j].Data.NameSpace }
// Swap in list-sort for TaskGroupList
func (l TaskGroupList) Swap(i, j int)        { l[i], l[j] = l[j], l[i] }

// Len in list-sort for ConfigMapList
func (l ConfigMapList) Len() int             { return len(l) }
// Less in list-sort for ConfigMapList
func (l ConfigMapList) Less(i, j int) bool   { return l[i].Data.NameSpace > l[j].Data.NameSpace }
// Swap in list-sort for ConfigMapList
func (l ConfigMapList) Swap(i, j int)        { l[i], l[j] = l[j], l[i] }

// Len in list-sort for SecretList
func (l SecretList) Len() int                { return len(l) }
// Less in list-sort for SecretList
func (l SecretList) Less(i, j int) bool      { return l[i].Data.NameSpace > l[j].Data.NameSpace }
// Swap in list-sort for SecretList
func (l SecretList) Swap(i, j int)           { l[i], l[j] = l[j], l[i] }

// Len in list-sort for ServiceList
func (l ServiceList) Len() int               { return len(l) }
// Less in list-sort for ServiceList
func (l ServiceList) Less(i, j int) bool     { return l[i].Data.NameSpace > l[j].Data.NameSpace }
// Swap in list-sort for ServiceList
func (l ServiceList) Swap(i, j int)          { l[i], l[j] = l[j], l[i] }

// Len in list-sort for EndpointList
func (l EndpointList) Len() int              { return len(l) }
// Less in list-sort for EndpointList
func (l EndpointList) Less(i, j int) bool    { return l[i].Data.NameSpace > l[j].Data.NameSpace }
// Swap in list-sort for EndpointList
func (l EndpointList) Swap(i, j int)         { l[i], l[j] = l[j], l[i] }

// Len in list-sort for DeploymentList
func (l DeploymentList) Len() int            { return len(l) }
// Less in list-sort for DeploymentList
func (l DeploymentList) Less(i, j int) bool {
	return l[i].Data.ObjectMeta.NameSpace > l[j].Data.ObjectMeta.NameSpace
}
// Swap in list-sort for DeploymentList
func (l DeploymentList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
