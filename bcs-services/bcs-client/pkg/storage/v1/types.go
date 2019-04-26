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
	status "bk-bcs/bcs-common/common/types"
	deploymentType "bk-bcs/bcs-mesos/bcs-scheduler/src/types"
)

const (
	BcsStorageDynamicTypeApplication = "application"
	BcsStorageDynamicTypeProcess     = "process"
	BcsStorageDynamicTypeTaskGroup   = "taskgroup"
	BcsStorageDynamicTypeConfigMap   = "configmap"
	BcsStorageDynamicTypeSecret      = "secret"
	BcsStorageDynamicTypeService     = "service"
	BcsStorageDynamicTypeEndpoint    = "endpoint"
	BcsStorageDynamicTypeDeployment  = "deployment"
	BcsStorageDynamicTypeNamespace   = "namespace"
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

type ApplicationList []*ApplicationSet
type ProcessList []*ProcessSet
type TaskGroupList []*TaskGroupSet
type ConfigMapList []*ConfigMapSet
type SecretList []*SecretSet
type ServiceList []*ServiceSet
type EndpointList []*EndpointSet
type DeploymentList []*DeploymentSet
