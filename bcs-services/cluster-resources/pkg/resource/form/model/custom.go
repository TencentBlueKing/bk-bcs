/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package model

// HookTmpl HookTemplate 表单化建模
type HookTmpl struct {
	Metadata Metadata     `structs:"metadata"`
	Spec     HookTmplSpec `structs:"spec"`
}

// HookTmplSpec ...
type HookTmplSpec struct {
	Args    []HookTmplArg    `structs:"args"`
	Policy  string           `structs:"policy"`
	Metrics []HookTmplMetric `structs:"metrics"`
}

// HookTmplArg ...
type HookTmplArg struct {
	Key   string `structs:"key"`
	Value string `structs:"value"`
}

// HookTmplMetric ...
type HookTmplMetric struct {
	Name     string `structs:"name"`
	HookType string `structs:"hookType"`
	// web provider
	URL         string `structs:"url"`
	JSONPath    string `structs:"jsonPath"`
	TimeoutSecs int64  `structs:"timeoutSecs"`
	// prometheus provider
	Address string `structs:"address"`
	Query   string `structs:"query"`
	// kubernetes provider
	Function string          `structs:"function"`
	Fields   []HookTmplField `structs:"fields"`
	// common fields
	Count               int64  `structs:"count"`
	Interval            int    `structs:"interval"`
	SuccessConditionExp string `structs:"successConditionExp"`
	SuccessPolicy       string `structs:"successPolicy"`
	SuccessCnt          int64  `structs:"successCnt"`
}

// HookTmplField ...
type HookTmplField struct {
	Key   string `structs:"key"`
	Value string `structs:"value"`
}

// GDeploy GameDeployment 表单化建模
type GDeploy struct {
	Metadata       Metadata       `structs:"metadata"`
	Spec           GDeploySpec    `structs:"spec"`
	Volume         WorkloadVolume `structs:"volume"`
	ContainerGroup ContainerGroup `structs:"containerGroup"`
}

// GDeploySpec ...
type GDeploySpec struct {
	Replicas        GDeployReplicas        `structs:"replicas"`
	GracefulManage  GDeployGracefulManage  `structs:"gracefulManage"`
	DeletionProtect GDeployDeletionProtect `structs:"deletionProtect"`
	NodeSelect      NodeSelect             `structs:"nodeSelect"`
	Affinity        Affinity               `structs:"affinity"`
	Toleration      Toleration             `structs:"toleration"`
	Networking      Networking             `structs:"networking"`
	Security        PodSecurityCtx         `structs:"security"`
	Other           SpecOther              `structs:"other"`
}

// GDeployReplicas ...
type GDeployReplicas struct {
	Cnt             int64  `structs:"cnt"`             // 副本数量
	UpdateStrategy  string `structs:"updateStrategy"`  // 更新策略（RollingUpdate/InplaceUpdate）
	MaxSurge        int64  `structs:"maxSurge"`        // 最大调度 Pod 数量
	MSUnit          string `structs:"msUnit"`          // 最大调度 Pod 数量单位（个/%）
	MaxUnavailable  int64  `structs:"maxUnavailable"`  // 最大不可用数量
	MUAUnit         string `structs:"muaUnit"`         // 最大不可用数量单位（个/%）
	MinReadySecs    int64  `structs:"minReadySecs"`    // 最小就绪时间
	Partition       int64  `structs:"partition"`       // 保留旧版本示例数量
	GracePeriodSecs int64  `structs:"gracePeriodSecs"` // 原地升级优雅更新时间
}

// GDeployGracefulManage 优雅删除/更新
type GDeployGracefulManage struct {
	PreDeleteHook   GDeployHookSpec `structs:"preDeleteHook"`
	PreInplaceHook  GDeployHookSpec `structs:"preInplaceHook"`
	PostInplaceHook GDeployHookSpec `structs:"postInplaceHook"`
}

// GDeployHookSpec ...
type GDeployHookSpec struct {
	Enabled  bool          `structs:"enabled"`
	TmplName string        `structs:"tmplName"`
	Args     []HookCallArg `structs:"args"`
}

// HookCallArg 调用 Hook 时传入的参数
type HookCallArg struct {
	Key   string `structs:"key"`
	Value string `structs:"value"`
}

// GDeployDeletionProtect 删除保护
type GDeployDeletionProtect struct {
	Policy string `structs:"policy"`
}
