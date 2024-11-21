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

// Package constant xxx
package constant

const (
	// ServiceName BCS 服务名称
	ServiceName = "bcs-project-manager"
	// ModuleName module name
	ModuleName = "bcsproject"
	// ServiceDomain 域名，用于注册到APISIX
	ServiceDomain = "project.bkbcs.tencent.com"
	// ClusterManagerDomain 用于发现 ClusterManager 服务
	ClusterManagerDomain = "clustermanager.bkbcs.tencent.com"
	// DefaultConfigPath 配置路径
	DefaultConfigPath = "./bcs-project-manager.yaml"
	// MicroMetaKeyHTTPPort 初始化micro服务需要的httpport
	MicroMetaKeyHTTPPort = "httpport"

	// TimeLayout time layout
	TimeLayout = "2006-01-02 15:04:05"

	// AnnotationKeyProjectCode annotation key projectCode
	AnnotationKeyProjectCode = "io.tencent.bcs.projectcode"

	// AnnotationKeyVcluster annotation key vcluster clusterID
	AnnotationKeyVcluster = "io.tencent.bcs.vcluster"

	// AnnotationKeyCreator annotation key projectCode
	AnnotationKeyCreator = "io.tencent.bcs.creator"

	// MaxMsgSize grpc限制的message的最大值
	MaxMsgSize int = 50 * 1024 * 1024

	// AnonymousUsername 匿名用户
	AnonymousUsername = "anonymous"

	// NamespaceSyncLockPrefix etcd distributed lock prefix for sync namespace
	NamespaceSyncLockPrefix = "namespace-sync"
)

const (
	// MetadataCookiesKey 在 GoMicro Metadata 中，Cookie 的键名
	MetadataCookiesKey = "Grpcgateway-Cookie"
	// LangCookieName 语言版本 Cookie 名称
	LangCookieName = "blueking_language"
)
