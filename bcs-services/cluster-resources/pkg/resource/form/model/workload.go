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

// Deploy Deployment 表单化建模
type Deploy struct {
	Metadata       Metadata
	Spec           DeploySpec
	Volume         WorkloadVolume
	ContainerGroup ContainerGroup
}

// DeploySpec ...
type DeploySpec struct {
	Replicas   DeployReplicas
	NodeSelect NodeSelect
	Affinity   Affinity
	Toleration Toleration
	Networking Networking
	Security   PodSecurityCtx
	Other      SpecOther
}

// DeployReplicas ...
type DeployReplicas struct {
	Cnt                  int64
	UpdateStrategy       string
	MaxSurge             int64
	MSUnit               string
	MaxUnavailable       int64
	MUAUnit              string
	MinReadySecs         int64
	ProgressDeadlineSecs int64
}

// NodeSelect ...
type NodeSelect struct {
	Type     string
	NodeName string
	Selector []NodeSelector
}

// NodeSelector ...
type NodeSelector struct {
	Key   string
	Value string
}

// Affinity ...
type Affinity struct {
	NodeAffinity []NodeAffinity
	PodAffinity  []PodAffinity
}

// NodeAffinity ...
type NodeAffinity struct {
	Priority string
	Weight   int64
	Selector NodeAffinitySelector
}

// NodeAffinitySelector ...
type NodeAffinitySelector struct {
	Expressions []ExpSelector
	Fields      []FieldSelector
}

// ExpSelector ...
type ExpSelector struct {
	Key    string
	Op     string
	Values string
}

// FieldSelector ...
type FieldSelector struct {
	Key    string
	Op     string
	Values string
}

// PodAffinity ...
type PodAffinity struct {
	Type        string
	Priority    string
	Namespaces  []string
	Weight      int64
	TopologyKey string
	Selector    PodAffinitySelector
}

// PodAffinitySelector ...
type PodAffinitySelector struct {
	Expressions []ExpSelector
	Labels      []LabelSelector
}

// LabelSelector ...
type LabelSelector struct {
	Key   string
	Value string
}

// Toleration ...
type Toleration struct {
	Rules []TolerationRule
}

// TolerationRule ...
type TolerationRule struct {
	Key            string
	Op             string `mapstructure:"operator"`
	Value          string
	Effect         string
	TolerationSecs int64 `mapstructure:"tolerationSeconds"`
}

// Networking ...
type Networking struct {
	DNSPolicy             string
	HostIPC               bool
	HostNetwork           bool
	HostPID               bool
	ShareProcessNamespace bool
	HostName              string
	Subdomain             string
	NameServers           []string
	Searches              []string
	DNSResolverOpts       []DNSResolverOpt
	HostAliases           []HostAlias
}

// DNSResolverOpt ...
type DNSResolverOpt struct {
	Name  string
	Value string
}

// HostAlias ...
type HostAlias struct {
	IP    string
	Alias string
}

// PodSecurityCtx ...
type PodSecurityCtx struct {
	RunAsUser    int64
	RunAsNonRoot bool
	RunAsGroup   int64
	FSGroup      int64
	SELinuxOpt   SELinuxOpt `mapstructure:"seLinuxOptions"`
}

// SELinuxOpt ...
type SELinuxOpt struct {
	Level string
	Role  string
	Type  string
	User  string
}

// SpecOther ...
type SpecOther struct {
	RestartPolicy              string
	TerminationGracePeriodSecs int64
	ImagePullSecrets           []string
	SAName                     string
}

// WorkloadVolume ...
type WorkloadVolume struct {
	PVC       []PVCVolume
	HostPath  []HostPathVolume
	ConfigMap []CMVolume
	Secret    []SecretVolume
	EmptyDir  []EmptyDirVolume
	NFS       []NFSVolume
}

// PVCVolume ...
type PVCVolume struct {
	Name     string
	PVCName  string
	ReadOnly bool
}

// HostPathVolume ...
type HostPathVolume struct {
	Name string
	Path string
	Type string
}

// CMVolume ...
type CMVolume struct {
	Name        string
	DefaultMode int64
	CMName      string
	Items       []KeyToPath
}

// SecretVolume ...
type SecretVolume struct {
	Name        string
	DefaultMode int64
	SecretName  string
	Items       []KeyToPath
}

// KeyToPath ...
type KeyToPath struct {
	Key  string
	Path string
}

// EmptyDirVolume ...
type EmptyDirVolume struct {
	Name string
}

// NFSVolume ...
type NFSVolume struct {
	Name     string
	Path     string
	Server   string
	ReadOnly bool
}

// ContainerGroup ...
type ContainerGroup struct {
	InitContainers []Container
	Containers     []Container
}

// Container ...
type Container struct {
	Basic    ContainerBasic
	Command  ContainerCommand
	Service  ContainerService
	Envs     ContainerEnvs
	Healthz  ContainerHealthz
	Resource ContainerRes
	Security SecurityCtx
	Mount    ContainerMount
}

// ContainerBasic ...
type ContainerBasic struct {
	Name       string
	Image      string
	PullPolicy string
}

// ContainerCommand ...
type ContainerCommand struct {
	WorkingDir string
	Stdin      bool
	StdinOnce  bool
	Tty        bool
	Command    []string
	Args       []string
}

// ContainerService ...
type ContainerService struct {
	Ports []ContainerPort
}

// ContainerPort ...
type ContainerPort struct {
	Name          string
	ContainerPort int64
	Protocol      string
	HostPort      int64
}

// ContainerEnvs ...
type ContainerEnvs struct {
	Vars []EnvVars
}

// EnvVars ...
type EnvVars struct {
	Type   string
	Name   string
	Source string
	Value  string
}

// ContainerHealthz ...
type ContainerHealthz struct {
	ReadinessProbe Probe
	LivenessProbe  Probe
}

// Probe ...
type Probe struct {
	PeriodSecs       int64
	InitialDelaySecs int64
	TimeoutSecs      int64
	SuccessThreshold int64
	FailureThreshold int64
	Type             string
	Path             string
	Port             int64
	Command          []string
}

// ContainerRes ...
type ContainerRes struct {
	Requests ResRequirement
	Limits   ResRequirement
}

// ResRequirement ...
type ResRequirement struct {
	CPU    int
	Memory int
}

// SecurityCtx ...
type SecurityCtx struct {
	Privileged               bool
	AllowPrivilegeEscalation bool
	RunAsNonRoot             bool
	ReadOnlyRootFilesystem   bool
	RunAsUser                int64
	RunAsGroup               int64
	ProcMount                string
	Capabilities             Capabilities
	SELinuxOpt               SELinuxOpt `mapstructure:"seLinuxOptions"`
}

// Capabilities ...
type Capabilities struct {
	Add  []string
	Drop []string
}

// ContainerMount ...
type ContainerMount struct {
	Volumes []MountVolume
}

// MountVolume ...
type MountVolume struct {
	Name      string
	MountPath string
	SubPath   string
	ReadOnly  bool
}
