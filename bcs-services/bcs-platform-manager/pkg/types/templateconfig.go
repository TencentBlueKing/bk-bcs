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

// Package types pod types
package types

// CreateTemplateConfigReq create template config request
type CreateTemplateConfigReq struct {
	BusinessID          string               `json:"businessID"`
	ProjectID           string               `json:"projectID"`
	ClusterID           string               `json:"clusterID"`
	Provider            string               `json:"provider"`
	ConfigType          string               `json:"configType"`
	CloudTemplateConfig *CloudTemplateConfig `json:"cloudTemplateConfig"`
}

// CloudTemplateConfig cloud template config
type CloudTemplateConfig struct {
	CloudNetworkTemplateConfig *CloudNetworkTemplateConfig `json:"cloudNetworkTemplateConfig"`
}

// CloudNetworkTemplateConfig cloud network template config
type CloudNetworkTemplateConfig struct {
	CidrSteps         []*EnvCidrStep `json:"cidrSteps"`
	ServiceSteps      []uint32       `json:"serviceSteps"`
	PerNodePodNum     []uint32       `json:"perNodePodNum"`
	UnderlaySteps     []uint32       `json:"underlaySteps"`
	UnderlayAutoSteps []uint32       `json:"underlayAutoSteps"`
}

// EnvCidrStep env cidr step
type EnvCidrStep struct {
	Env  string `json:"env"`
	Step uint32 `json:"step"`
}

// DeleteTemplateConfigReq delete template config request
type DeleteTemplateConfigReq struct {
	TemplateConfigID string `json:"cloudID" in:"path=templateConfigID"`
	BusinessID       string `json:"businessID" in:"query=businessID"`
	ProjectID        string `json:"projectID" in:"query=projectID"`
}
