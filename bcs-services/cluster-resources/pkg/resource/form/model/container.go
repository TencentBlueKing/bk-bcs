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

// Container ...
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

// ContainerBasic ...
type ContainerBasic struct {
	Name       string `structs:"name"`
	Image      string `structs:"image"`
	PullPolicy string `structs:"pullPolicy"`
}

// ContainerCommand ...
type ContainerCommand struct {
	WorkingDir string   `structs:"workingDir"`
	Stdin      bool     `structs:"stdin"`
	StdinOnce  bool     `structs:"stdinOnce"`
	Tty        bool     `structs:"tty"`
	Command    []string `structs:"command"`
	Args       []string `structs:"args"`
}

// ContainerService ...
type ContainerService struct {
	Ports []ContainerPort `structs:"ports"`
}

// ContainerPort ...
type ContainerPort struct {
	Name          string `structs:"name"`
	ContainerPort int64  `structs:"containerPort"`
	Protocol      string `structs:"protocol"`
	HostPort      int64  `structs:"hostPort"`
}

// ContainerEnvs ...
type ContainerEnvs struct {
	Vars []EnvVar `structs:"vars"`
}

// EnvVar ...
type EnvVar struct {
	Type   string `structs:"type"`
	Name   string `structs:"name"`
	Source string `structs:"source"`
	Value  string `structs:"value"`
}

// ContainerHealthz ...
type ContainerHealthz struct {
	ReadinessProbe Probe `structs:"readinessProbe"`
	LivenessProbe  Probe `structs:"livenessProbe"`
}

// Probe ...
type Probe struct {
	PeriodSecs       int64    `structs:"periodSecs"`
	InitialDelaySecs int64    `structs:"initialDelaySecs"`
	TimeoutSecs      int64    `structs:"timeoutSecs"`
	SuccessThreshold int64    `structs:"successThreshold"`
	FailureThreshold int64    `structs:"failureThreshold"`
	Type             string   `structs:"type"`
	Path             string   `structs:"path"`
	Port             int64    `structs:"port"`
	Command          []string `structs:"command"`
}

// ContainerRes ...
type ContainerRes struct {
	Requests ResRequirement `structs:"requests"`
	Limits   ResRequirement `structs:"limits"`
}

// ResRequirement ...
type ResRequirement struct {
	CPU    int `structs:"cpu"`
	Memory int `structs:"memory"`
}

// SecurityCtx ...
type SecurityCtx struct {
	Privileged               bool         `structs:"privileged"`
	AllowPrivilegeEscalation bool         `structs:"allowPrivilegeEscalation"`
	RunAsNonRoot             bool         `structs:"runAsNonRoot"`
	ReadOnlyRootFilesystem   bool         `structs:"readOnlyRootFilesystem"`
	RunAsUser                int64        `structs:"runAsUser"`
	RunAsGroup               int64        `structs:"runAsGroup"`
	ProcMount                string       `structs:"procMount"`
	Capabilities             Capabilities `structs:"capabilities"`
	SELinuxOpt               SELinuxOpt   `structs:"seLinuxOpt" mapstructure:"seLinuxOptions"`
}

// Capabilities ...
type Capabilities struct {
	Add  []string `structs:"add"`
	Drop []string `structs:"drop"`
}

// ContainerMount ...
type ContainerMount struct {
	Volumes []MountVolume `structs:"volumes"`
}

// MountVolume ...
type MountVolume struct {
	Name      string `structs:"name"`
	MountPath string `structs:"mountPath"`
	SubPath   string `structs:"subPath"`
	ReadOnly  bool   `structs:"readOnly"`
}
