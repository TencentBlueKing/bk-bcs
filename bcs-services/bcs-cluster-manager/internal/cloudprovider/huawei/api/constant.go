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

// Package api xxx
package api

// CCE cluster status
const (
	// Available 集群可用
	Available = "Available"
	// Unavailable 集群不可用
	Unavailable = "Unavailable"
	// Creating 创建中
	Creating = "Creating"
	// Deleting 删除中
	Deleting = "Deleting"
	// Error 错误
	Error = "Error"
)

// cluster type
const (
	// VirtualMachine CCE集群
	VirtualMachine = "VirtualMachine"
	// ARM64 鲲鹏集群
	ARM64 = "ARM64"
)

const (
	// JobPhaseInitializing 初始化
	JobPhaseInitializing = "Initializing"
	// JobPhaseRunning 运行中
	JobPhaseRunning = "Running"
	// JobPhaseFailed 失败
	JobPhaseFailed = "Failed"
	// JobPhaseSuccess 成功
	JobPhaseSuccess = "Success"
)

const (
	// NodePoolIdKey nodePool key
	NodePoolIdKey = "kubernetes.io/node-pool.id"
	// NodePoolCordonTaintKey nodePool cordon taint key
	NodePoolCordonTaintKey = "node.cloudprovider.kubernetes.io/uninitialized"
)

// CCE node status
const (
	// NodeBuild 创建中，表示节点正处于创建过程中。
	NodeBuild = "Build"
	// NodeInstalling 纳管中，表示节点正处于纳管过程中
	NodeInstalling = "Installing"
	// NodeUpgrading 升级中，表示节点正处于升级过程中。
	NodeUpgrading = "Upgrading"
	// NodeActive 正常，表示节点处于正常状态
	NodeActive = "Active"
	// NodeAbnormal 异常，表示节点处于异常状态
	NodeAbnormal = "Abnormal"
	// NodeDeleting 删除中，表示节点正处于删除过程中
	NodeDeleting = "Deleting"
	// NodeError 故障，表示节点处于故障状态
	NodeError = "Error"
)

const (
	// NodePoolSynchronizing 伸缩中（节点池当前节点数未达到预期，且无伸缩中的节点) ;空值 可用（节点池当前节点数已达到预期，且无伸缩中的节点）
	NodePoolSynchronizing = "Synchronizing"
	// NodePoolSynchronized 伸缩等待中（节点池当前节点数未达到预期，或者存在伸缩中的节点）
	NodePoolSynchronized = "Synchronized"
	// NodePoolSoldOut 节点池当前不可扩容（兼容字段，标记节点池资源售罄、资源配额不足等不可扩容状态）
	NodePoolSoldOut = "SoldOut"
	// NodePoolDeleting 删除中
	NodePoolDeleting = "Deleting"
	// NodePoolError 错误
	NodePoolError = "Error"
)

const (
	// SubnetScopeCenter center-表示作用域为中心
	SubnetScopeCenter = "center"
	// SubnetScopeAz {azId}表示作用域为具体的AZ
	SubnetScopeAz = "azId"
)

const (
	// ClusterInstallAddonsExternalInstall cluster.install.addons.external/install
	ClusterInstallAddonsExternalInstall = "cluster.install.addons.external/install"
	// ClusterInstallAddonsExternalInstallValue xxx
	ClusterInstallAddonsExternalInstallValue = `[{"addonTemplateName":"icagent",
	"extendParam":{"logSwitch":"false","tDSEnable":"true"}}]`

	// ClusterInstallAddonsInstall cluster.install.addons/install
	ClusterInstallAddonsInstall = "cluster.install.addons/install"
	// ClusterInstallAddonsInstallValue xxx
	ClusterInstallAddonsInstallValue = `[{"addonTemplateName":"coredns",
	"values":{"flavor":{"name":20000,"recommend_cluster_flavor_types":["xlarge"],
	"replicas":4,"resources":[{"limitsCpu":"2000m","limitsMem":"2048Mi",
	"name":"coredns","requestsCpu":"2000m","requestsMem":"2048Mi"}],
	"category":["CCE","Turbo"]}}},{"addonTemplateName":"everest"},
	{"addonTemplateName":"node-local-dns"},{"addonTemplateName":"npd"}]`
)

const (
	ContainerNetworkModeVpcRouter = "vpc-router"
	ContainerNetworkModeOverlayL2 = "overlay_l2"
)

const (
	ChargemodeTraffic   = "traffic"
	ChargemodeBandwidth = "bandwidth"
)
