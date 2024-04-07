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

package model

// Container 容器配置
type Container struct {
	Basic    ContainerBasic   `structs:"basic"`
	Command  ContainerCommand `structs:"command"`
	Service  ContainerService `structs:"service"`
	Envs     ContainerEnvs    `structs:"envs"`
	Healthz  ContainerHealthz `structs:"healthz"`
	Resource ContainerRes     `structs:"resource"`
	Security SecurityCtx      `structs:"security"`
	Mount    ContainerMount   `structs:"mount"`
}

// ContainerBasic 容器基础信息
type ContainerBasic struct {
	Name       string `structs:"name"`
	Image      string `structs:"image"`
	PullPolicy string `structs:"pullPolicy"`
}

// ContainerCommand 容器命令
type ContainerCommand struct {
	WorkingDir string   `structs:"workingDir"`
	Stdin      bool     `structs:"stdin"`
	StdinOnce  bool     `structs:"stdinOnce"`
	Tty        bool     `structs:"tty"`
	Command    []string `structs:"command"`
	Args       []string `structs:"args"`
}

// ContainerService 容器网络服务
type ContainerService struct {
	Ports []ContainerPort `structs:"ports"`
}

// ContainerPort 端口
type ContainerPort struct {
	Name          string `structs:"name"`
	ContainerPort int64  `structs:"containerPort"`
	Protocol      string `structs:"protocol"`
	HostPort      int64  `structs:"hostPort"`
}

// ContainerEnvs 环境配置
type ContainerEnvs struct {
	Vars []EnvVar `structs:"vars"`
}

// EnvVar 环境变量
type EnvVar struct {
	Type   string `structs:"type"`
	Name   string `structs:"name"`
	Source string `structs:"source"`
	Value  string `structs:"value"`
}

// ContainerHealthz 健康检查
type ContainerHealthz struct {
	ReadinessProbe Probe `structs:"readinessProbe"`
	LivenessProbe  Probe `structs:"livenessProbe"`
}

// Probe 探针
type Probe struct {
	Enabled          bool     `structs:"enabled"`          // 是否启用
	PeriodSecs       int64    `structs:"periodSecs"`       // 检查间隔
	InitialDelaySecs int64    `structs:"initialDelaySecs"` // 初始延时
	TimeoutSecs      int64    `structs:"timeoutSecs"`      // 超时时间
	SuccessThreshold int64    `structs:"successThreshold"` // 成功阈值
	FailureThreshold int64    `structs:"failureThreshold"` // 失败阈值
	Type             string   `structs:"type"`
	Path             string   `structs:"path"`
	Port             int64    `structs:"port"`
	Command          []string `structs:"command"`
}

// ContainerRes 资源配置
type ContainerRes struct {
	Requests ResRequirement `structs:"requests"`
	Limits   ResRequirement `structs:"limits"`
}

// ResRequirement 资源集
type ResRequirement struct {
	CPU              int        `structs:"cpu"`
	Memory           int        `structs:"memory"`
	EphemeralStorage int        `structs:"ephemeral-storage"`
	Extra            []ResExtra `structs:"extra"`
}

// ResExtra 额外资源集
type ResExtra struct {
	Key   string `structs:"key"`
	Value string `structs:"value"`
}

// SecurityCtx 安全性上下文
type SecurityCtx struct {
	Privileged               bool         `structs:"privileged"`               // 特权模式
	AllowPrivilegeEscalation bool         `structs:"allowPrivilegeEscalation"` // 允许提权
	RunAsNonRoot             bool         `structs:"runAsNonRoot"`             // 以非 Root 方式运行
	ReadOnlyRootFilesystem   bool         `structs:"readOnlyRootFilesystem"`   // 只读 Root 文件系统
	RunAsUser                int64        `structs:"runAsUser"`                // 运行用户
	RunAsGroup               int64        `structs:"runAsGroup"`               // 用户组
	ProcMount                string       `structs:"procMount"`                // 掩码挂载
	Capabilities             Capabilities `structs:"capabilities"`             // 权限信息
	SELinuxOpt               SELinuxOpt   `structs:"seLinuxOpt" mapstructure:"seLinuxOptions"`
}

// Capabilities 特殊权限
type Capabilities struct {
	Add  []string `structs:"add"`
	Drop []string `structs:"drop"`
}

// ContainerMount 挂载卷配置
type ContainerMount struct {
	Volumes []MountVolume `structs:"volumes"`
}

// MountVolume 挂载卷
type MountVolume struct {
	Name      string `structs:"name"`
	MountPath string `structs:"mountPath"`
	SubPath   string `structs:"subPath"`
	ReadOnly  bool   `structs:"readOnly"`
}
