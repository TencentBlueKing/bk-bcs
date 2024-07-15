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

// Package helmManger helm-service
package helmManger

import "github.com/golang/protobuf/jsonpb"

const (
	// installAddonsApi post ( projectID + clusterID )
	installAddonsApi = "/helmmanager/v1/projects/%s/clusters/%s/addons"

	// uninstallAddonsApi delete (projectID + clusterID + name )
	uninstallAddonsApi = "/helmmanager/v1/projects/%s/clusters/%s/addons/%s"

	// upgradeAddonsApi put (projectID + clusterID + name )
	upgradeAddonsApi = "/helmmanager/v1/projects/%s/clusters/%s/addons/%s"

	// getAddonsDetailApi get (projectID + clusterID + name )
	getAddonsDetailApi = "/helmmanager/v1/projects/%s/clusters/%s/addons/%s"

	// listAddonsApi  get ( projectID + clusterID )
	listAddonsApi = "/helmmanager/v1/projects/%s/clusters/%s/addons"
)

// InstallAddonsRequest body
//
// ProjectID 项目编码(英文缩写), 全局唯一, 长度不能超过64字符
// ***不能为空***
//
// ClusterID 所在的集群ID
// ***不能为空***
//
// Name 组件名称
// ***不能为空***
//
// Version 组件版本
// ***不能为空***
//
// Values values.yaml
// ***不能为空***
type InstallAddonsRequest struct {
	// RequestID 请求ID
	RequestID string

	// ProjectID 项目编码(英文缩写), 全局唯一, 长度不能超过64字符
	// ***不能为空***
	ProjectID string `json:"projectID"`

	// ClusterID 所在的集群ID
	// ***不能为空***
	ClusterID string `json:"clusterID,omitempty"`

	// Name 组件名称
	// ***不能为空***
	Name string `json:"name,omitempty"`

	// Version 组件版本
	// ***不能为空***
	Version string `json:"version,omitempty"`

	// Values values.yaml
	// ***不能为空***
	Values string `json:"values,omitempty"`
}

// UninstallAddonsRequest  body
//
// ProjectID 项目ID
// ***不能为空***
//
// ClusterID 所在的集群ID
// ***不能为空***
//
// Name 组件名称
// ***不能为空***
type UninstallAddonsRequest struct {
	// RequestID 请求ID
	RequestID string

	// ProjectID 项目编码(英文缩写), 全局唯一, 长度不能超过64字符
	// ***不能为空***
	ProjectID string `json:"projectID"`

	// ClusterID 所在的集群ID
	// ***不能为空***
	ClusterID string `json:"clusterID,omitempty"`

	// Name 组件名称
	// ***不能为空***
	Name string `json:"name,omitempty"`
}

// UpgradeAddonsRequest update body.
//
// ProjectID 项目编码(英文缩写), 全局唯一, 长度不能超过64字符
// ***不能为空***
//
// ClusterID 所在的集群ID
// ***不能为空***
//
// Name 组件名称
// ***不能为空***
//
// Version 组件版本
// ***不能为空***
//
// Values values.yaml
// ***不能为空***
type UpgradeAddonsRequest struct {
	// RequestID 请求ID
	RequestID string

	// ProjectID 项目编码(英文缩写), 全局唯一, 长度不能超过64字符
	// ***不能为空***
	ProjectID string `json:"projectID"`

	// ClusterID 所在的集群ID
	// ***不能为空***
	ClusterID string `json:"clusterID,omitempty"`

	// Name 组件名称
	// ***不能为空***
	Name string `json:"name,omitempty"`

	// Version 组件版本
	// ***不能为空***
	Version string `json:"version,omitempty"`

	// Values values.yaml
	// ***不能为空***
	Values string `json:"values,omitempty"`
}

// GetAddonsDetailRequest  body
//
// ProjectID 项目ID
// ***不能为空***
//
// ClusterID 所在的集群ID
// ***不能为空***
//
// Name 组件名称
// ***不能为空***
type GetAddonsDetailRequest struct {
	// RequestID 请求ID
	RequestID string

	// ProjectID 项目编码(英文缩写), 全局唯一, 长度不能超过64字符
	// ***不能为空***
	ProjectID string `json:"projectID"`

	// ClusterID 所在的集群ID
	// ***不能为空***
	ClusterID string `json:"clusterID,omitempty"`

	// Name 组件名称
	// ***不能为空***
	Name string `json:"name,omitempty"`
}

// ListAddonsRequest  body
//
// ProjectID 项目编码(英文缩写), 全局唯一, 长度不能超过64字符
// ***不能为空***
//
// ClusterID 所在的集群ID
// ***不能为空***
type ListAddonsRequest struct {
	// RequestID 请求ID
	RequestID string

	// ProjectID 项目编码(英文缩写), 全局唯一, 长度不能超过64字符
	// ***不能为空***
	ProjectID string `json:"projectID"`

	// ClusterID 所在的集群ID
	// ***不能为空***
	ClusterID string `json:"clusterID,omitempty"`
}

var (
	// pbMarshaller 创建一个jsonpb.Marshaler
	pbMarshaller = new(jsonpb.Marshaler)
)
