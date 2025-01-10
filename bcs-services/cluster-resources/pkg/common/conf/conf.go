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

// Package conf xxx
package conf

const (
	// ServiceDomain 服务域名
	ServiceDomain = "clusterresources.bkbcs.tencent.com"
	// DefaultConfPath 默认配置存放路径
	DefaultConfPath = "etc/conf.yaml"
	// ProjectMgrServiceName 项目管理服务名
	ProjectMgrServiceName = "project.bkbcs.tencent.com"
	// ClusterMgrServiceName 集群管理服务名
	ClusterMgrServiceName = "clustermanager.bkbcs.tencent.com"
	// ProjectCodeAnnoKey 命名空间所属 projectcode 注解 key 的默认值
	ProjectCodeAnnoKey = "io.tencent.bcs.projectcode"
	// LangCookieName 语言版本 Cookie 名称
	LangCookieName = "blueking_language"
	// MaxGrpcMsgSize 单请求/响应体最大尺寸 64MB
	MaxGrpcMsgSize = 64 * 1024 * 1024
)
