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

package constants

// k8s 资源类型
const (
	// Node ...
	Node = "Node"

	// NS ...
	NS = "Namespace"

	// Deploy ...
	Deploy = "Deployment"
	// RS ...
	RS = "ReplicaSet"
	// DS ...
	DS = "DaemonSet"
	// STS ...
	STS = "StatefulSet"
	// CJ ...
	CJ = "CronJob"
	// Job ...
	Job = "Job"
	// Po ...
	Po = "Pod"

	// Ing ...
	Ing = "Ingress"
	// SVC ...
	SVC = "Service"
	// EP ...
	EP = "Endpoints"

	// CM ...
	CM = "ConfigMap"
	// Secret ...
	Secret = "Secret"

	// PV ...
	PV = "PersistentVolume"
	// PVC ...
	PVC = "PersistentVolumeClaim"
	// SC ...
	SC = "StorageClass"

	// SA ...
	SA = "ServiceAccount"

	// HPA ...
	HPA = "HorizontalPodAutoscaler"

	// CRD ...
	CRD = "CustomResourceDefinition"
	// CObj ...
	CObj = "CustomObject"
)

// BCS 提供自定义类型
const (
	// GDeploy ...
	GDeploy = "GameDeployment"

	// GSTS ...
	GSTS = "GameStatefulSet"

	// HookTmpl ...
	HookTmpl = "HookTemplate"

	// HookRun ...
	HookRun = "HookRun"
)

const (
	// WatchTimeout 执行资源变动 Watch 超时时间 30 分钟
	WatchTimeout = 30 * 60
)

const (
	// NamespacedScope 命名空间维度
	NamespacedScope = "Namespaced"

	// ClusterScope 集群维度
	ClusterScope = "Cluster"
)

const (
	// MasterNodeLabelKey Node 存在该标签，且值为 "true" 说明是 master
	MasterNodeLabelKey = "node-role.kubernetes.io/master"
)

const (
	// EditModeAnnoKey 资源被编辑的模式，表单为 form，Key 不存在或 Manifest 则为 Yaml 模式
	EditModeAnnoKey = "io.tencent.bcs.editFormat"
)

const (
	// CreatorAnnoKey 创建者
	CreatorAnnoKey = "io.tencent.paas.creator"

	// UpdaterAnnoKey 更新者，为保持与 bcs-ui 中的一致，还是使用 updator（typo）
	UpdaterAnnoKey = "io.tencent.paas.updator"
)

const (
	// EditModeForm 资源编辑模式 - 表单
	EditModeForm = "form"
	// EditModeYaml 资源编辑模式 - Yaml
	EditModeYaml = "yaml"
)

const (
	// MetricResCPU 指标资源：CPU
	MetricResCPU = "cpu"
	// MetricResMem 指标资源：内存
	MetricResMem = "memory"
	// MetricResEphemeralStorage 指标资源：临时存储
	MetricResEphemeralStorage = "ephemeral-storage"
)

const (
	// BCSNetworkApiVersion BCS Network CRD apiVersion
	BCSNetworkApiVersion = "networkextension.bkbcs.tencent.com/v1"
)

// RemoveResVersionKinds 更新时强制移除 resourceVersion 的资源类型
// 添加 HPA 原因是，HPA 每次做扩缩容操作，集群均会更新资源（rv），过于频繁导致用户编辑态的 rv 过期 & 冲突导致无法更新
// 理论上所有资源都可能会有这样的问题，不止其他用户操作，集群也可能操作导致 rv 过期，但是因 HPA 过于频繁，因此这里配置需要移除 rv
// 注意：该行为相当于强制更新，会覆盖集群中的该资源，另外需要注意的是 service 这类资源必须指定 resourceVersion，否则报错如下：
// Service "service-xxx" is invalid: metadata.resourceVersion: Invalid value: "": must be specified for an update
var RemoveResVersionKinds = []string{HPA}

const (
	// CurrentRevision 当前deployment 版本
	CurrentRevision = "current_revision"
	// RolloutRevision 要回滚的deployment 版本
	RolloutRevision = "rollout_revision"
	// Revision deployment 版本
	Revision = "deployment.kubernetes.io/revision"
	// ChangeCause deployment更新原因
	ChangeCause = "deployment.kubernetes.io/change-cause"

	// STSChangeCause StatefulSet 更新原因
	STSChangeCause = "statefulset.kubernetes.io/change-cause"
	// DSChangeCause DaemonSet 更新原因
	DSChangeCause = "daemonset.kubernetes.io/change-cause"
)
