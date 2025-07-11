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

// Package common 提供mesh manager的常量定义
package common

const (
	// ServiceDomain domain name for service
	ServiceDomain = "meshmanager.bkbcs.tencent.com"
	// HelmManagerServiceDomain domain name for helm manager
	HelmManagerServiceDomain = "helmmanager.bkbcs.tencent.com"
	// ProjectManagerServiceName project manager service name
	ProjectManagerServiceName = "project.bkbcs.tencent.com"

	// MicroMetaKeyHTTPPort http port in micro-service meta
	MicroMetaKeyHTTPPort = "httpport"

	// MetaKeyHTTPPort TODO
	MetaKeyHTTPPort = "httpport"

	// LangContectKey lang context key
	LangContectKey string = "lang"
)

const (
	// ComponentIstiod istiod组件
	ComponentIstiod = "istiod"
	// ComponentIstioBase istio base组件
	ComponentIstioBase = "base"
	// ComponentIstioGateway istio gateway组件
	ComponentIstioGateway = "gateway"

	// IstioInstallModePrimary 主集群安装
	IstioInstallModePrimary = "primary"
	// IstioInstallModeRemote 远程集群安装
	IstioInstallModeRemote = "remote"
	// IstioNamespace istio命名空间
	IstioNamespace = "istio-system"

	// IstioInstallBaseName istio base安装名称
	IstioInstallBaseName = "bcs-istio-base"
	// IstioInstallIstiodName istiod安装名称
	IstioInstallIstiodName = "bcs-istio-istiod"
	// IstioInstallIstioGatewayName istio gateway安装名称
	IstioInstallIstioGatewayName = "bcs-istio-ingress-gateway"
)

const (
	// StringTrue 字符串true
	StringTrue = "true"
	// StringFalse 字符串false
	StringFalse = "false"
)

const (
	// ControlPlaneModeHosting 托管控制面
	ControlPlaneModeHosting = "hosting"
	// ControlPlaneModeIndependent 独立控制面
	ControlPlaneModeIndependent = "independent"

	// MultiClusterModePrimaryRemote 主从架构
	MultiClusterModePrimaryRemote = "primaryPemote"
	// MultiClusterModeMultiPrimary 主从架构
	MultiClusterModeMultiPrimary = "multiPrimary"

	// AccessLogEncodingJSON 日志编码json
	AccessLogEncodingJSON = "json"
	// AccessLogEncodingTEXT 日志编码text
	AccessLogEncodingTEXT = "text"

	// EnvPilotHTTP10 是否开启HTTP1.0
	EnvPilotHTTP10 = "PILOT_HTTP10"
)

// IstioStatus 控制面状态
const (
	// IstioStatusRunning 运行中
	IstioStatusRunning = "running"
	// IstioStatusInstalling 安装中
	IstioStatusInstalling = "installing"
	// IstioStatusInstalled 安装完成
	IstioStatusInstalled = "installed"
	// IstioStatusFailed 安装失败
	IstioStatusInstallFailed = "install-failed"
	// IstioStatusUninstalling 卸载中
	IstioStatusUninstalling = "uninstalling"
	// IstioStatusUninstalled 卸载完成
	IstioStatusUninstalled = "uninstalled"
	// IstioStatusUninstallingFailed 卸载失败
	IstioStatusUninstallingFailed = "uninstalling-failed"
	// IstioStatusUpdating 配置更新中
	IstioStatusUpdating = "updating"
	// IstioStatusUpdateFailed 配置更新失败
	IstioStatusUpdateFailed = "update-failed"
)

// shared cluster
const (
	// AnnotationKeyProjectCode namespace 的 projectcode 注解 key 默认值
	AnnotationKeyProjectCode = "io.tencent.bcs.projectcode"
)

// MeshManager接口常量
const (
	// MeshManagerInstallIstio 安装Istio接口
	MeshManagerInstallIstio = "MeshManager.InstallIstio"
	// MeshManagerUpdateIstio 更新Istio接口
	MeshManagerUpdateIstio = "MeshManager.UpdateIstio"
	// MeshManagerDeleteIstio 删除Istio接口
	MeshManagerDeleteIstio = "MeshManager.DeleteIstio"
	// MeshManagerGetIstioDetail 获取Istio详情接口
	MeshManagerGetIstioDetail = "MeshManager.GetIstioDetail"
	// MeshManagerListIstio 获取Istio列表接口
	MeshManagerListIstio = "MeshManager.ListIstio"
)
