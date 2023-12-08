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

// Deploy Deployment 表单化建模
type Deploy struct {
	Metadata       Metadata       `structs:"metadata"`
	Spec           DeploySpec     `structs:"spec"`
	Volume         WorkloadVolume `structs:"volume"`
	ContainerGroup ContainerGroup `structs:"containerGroup"`
}

// DeploySpec xxx
type DeploySpec struct {
	Replicas       DeployReplicas `structs:"replicas"`
	NodeSelect     NodeSelect     `structs:"nodeSelect"`
	Affinity       Affinity       `structs:"affinity"`
	Toleration     Toleration     `structs:"toleration"`
	Networking     Networking     `structs:"networking"`
	Security       PodSecurityCtx `structs:"security"`
	ReadinessGates ReadinessGates `structs:"readinessGates"`
	Other          SpecOther      `structs:"other"`
}

// DeployReplicas xxx
type DeployReplicas struct {
	Cnt                  int64  `structs:"cnt"`                  // 副本数量
	UpdateStrategy       string `structs:"updateStrategy"`       // 更新策略
	MaxSurge             int64  `structs:"maxSurge"`             // 最大调度 Pod 数量
	MSUnit               string `structs:"msUnit"`               // 最大调度 Pod 数量单位（个/%）
	MaxUnavailable       int64  `structs:"maxUnavailable"`       // 最大不可用数量
	MUAUnit              string `structs:"muaUnit"`              // 最大不可用数量单位（个/%）
	MinReadySecs         int64  `structs:"minReadySecs"`         // 最小就绪时间
	ProgressDeadlineSecs int64  `structs:"progressDeadlineSecs"` // 进程截止时间
}

// WorkloadVolume xxx
type WorkloadVolume struct {
	PVC       []PVCVolume      `structs:"pvc"`
	HostPath  []HostPathVolume `structs:"hostPath"`
	ConfigMap []CMVolume       `structs:"configMap"`
	Secret    []SecretVolume   `structs:"secret"`
	EmptyDir  []EmptyDirVolume `structs:"emptyDir"`
	NFS       []NFSVolume      `structs:"nfs"`
}

// ContainerGroup xxx
type ContainerGroup struct {
	InitContainers []Container `structs:"initContainers"`
	Containers     []Container `structs:"containers"`
}

// DS DaemonSet 表单化建模
type DS struct {
	Metadata       Metadata       `structs:"metadata"`
	Spec           DSSpec         `structs:"spec"`
	Volume         WorkloadVolume `structs:"volume"`
	ContainerGroup ContainerGroup `structs:"containerGroup"`
}

// DSSpec xxx
type DSSpec struct {
	Replicas   DSReplicas     `structs:"replicas"`
	NodeSelect NodeSelect     `structs:"nodeSelect"`
	Affinity   Affinity       `structs:"affinity"`
	Toleration Toleration     `structs:"toleration"`
	Networking Networking     `structs:"networking"`
	Security   PodSecurityCtx `structs:"security"`
	Other      SpecOther      `structs:"other"`
}

// DSReplicas xxx
type DSReplicas struct {
	UpdateStrategy string `structs:"updateStrategy"`
	MaxUnavailable int64  `structs:"maxUnavailable"`
	MUAUnit        string `structs:"muaUnit"`
	MinReadySecs   int64  `structs:"minReadySecs"`
}

// STS StatefulSet 表单化建模
type STS struct {
	Metadata       Metadata       `structs:"metadata"`
	Spec           STSSpec        `structs:"spec"`
	Volume         WorkloadVolume `structs:"volume"`
	ContainerGroup ContainerGroup `structs:"containerGroup"`
}

// STSSpec xxx
type STSSpec struct {
	Replicas        STSReplicas        `structs:"replicas"`
	VolumeClaimTmpl STSVolumeClaimTmpl `structs:"volumeClaimTmpl"`
	NodeSelect      NodeSelect         `structs:"nodeSelect"`
	Affinity        Affinity           `structs:"affinity"`
	Toleration      Toleration         `structs:"toleration"`
	Networking      Networking         `structs:"networking"`
	Security        PodSecurityCtx     `structs:"security"`
	ReadinessGates  ReadinessGates     `structs:"readinessGates"`
	Other           SpecOther          `structs:"other"`
}

// STSReplicas xxx
type STSReplicas struct {
	SVCName        string `structs:"svcName"`
	Cnt            int64  `structs:"cnt"`
	UpdateStrategy string `structs:"updateStrategy"`
	PodManPolicy   string `structs:"podManPolicy"`
	Partition      int64  `structs:"partition"`
}

// STSVolumeClaimTmpl xxx
type STSVolumeClaimTmpl struct {
	Claims []VolumeClaim `structs:"claims"`
}

// VolumeClaim xxx
type VolumeClaim struct {
	PVCName     string   `structs:"pvcName"`
	ClaimType   string   `structs:"claimType"`
	PVName      string   `structs:"pvName"`
	SCName      string   `structs:"scName"`
	StorageSize int      `structs:"storageSize"`
	AccessModes []string `structs:"accessModes"`
}

// CJ CronJob 表单化建模
type CJ struct {
	Metadata       Metadata       `structs:"metadata"`
	Spec           CJSpec         `structs:"spec"`
	Volume         WorkloadVolume `structs:"volume"`
	ContainerGroup ContainerGroup `structs:"containerGroup"`
}

// CJSpec xxx
type CJSpec struct {
	JobManage  CJJobManage    `structs:"jobManage"`
	NodeSelect NodeSelect     `structs:"nodeSelect"`
	Affinity   Affinity       `structs:"affinity"`
	Toleration Toleration     `structs:"toleration"`
	Networking Networking     `structs:"networking"`
	Security   PodSecurityCtx `structs:"security"`
	Other      SpecOther      `structs:"other"`
}

// CJJobManage xxx
type CJJobManage struct {
	Schedule                   string `structs:"schedule"`                   // 调度规则
	ConcurrencyPolicy          string `structs:"concurrencyPolicy"`          // 并发策略
	Suspend                    bool   `structs:"suspend"`                    // 暂停
	Completions                int64  `structs:"completions"`                // 需完成数
	Parallelism                int64  `structs:"parallelism"`                // 并发数
	BackoffLimit               int64  `structs:"backoffLimit"`               // 重试次数
	ActiveDDLSecs              int64  `structs:"activeDDLSecs"`              // 活跃终止时间
	SuccessfulJobsHistoryLimit int64  `structs:"successfulJobsHistoryLimit"` // 历史累计成功数
	FailedJobsHistoryLimit     int64  `structs:"failedJobsHistoryLimit"`     // 历史累计失败数
	StartingDDLSecs            int64  `structs:"startingDDLSecs"`            // 运行截止时间
}

// Job 表单化建模
type Job struct {
	Metadata       Metadata       `structs:"metadata"`
	Spec           JobSpec        `structs:"spec"`
	Volume         WorkloadVolume `structs:"volume"`
	ContainerGroup ContainerGroup `structs:"containerGroup"`
}

// JobSpec xxx
type JobSpec struct {
	JobManage  JobManage      `structs:"jobManage"`
	NodeSelect NodeSelect     `structs:"nodeSelect"`
	Affinity   Affinity       `structs:"affinity"`
	Toleration Toleration     `structs:"toleration"`
	Networking Networking     `structs:"networking"`
	Security   PodSecurityCtx `structs:"security"`
	Other      SpecOther      `structs:"other"`
}

// JobManage xxx
type JobManage struct {
	Completions   int64 `structs:"completions"`   // 需完成数
	Parallelism   int64 `structs:"parallelism"`   // 并发数
	BackoffLimit  int64 `structs:"backoffLimit"`  // 重试次数
	ActiveDDLSecs int64 `structs:"activeDDLSecs"` // 活跃终止时间
}

// Po Pod 表单化建模
type Po struct {
	Metadata       Metadata       `structs:"metadata"`
	Spec           PoSpec         `structs:"spec"`
	Volume         WorkloadVolume `structs:"volume"`
	ContainerGroup ContainerGroup `structs:"containerGroup"`
}

// PoSpec xxx
type PoSpec struct {
	NodeSelect     NodeSelect     `structs:"nodeSelect"`
	Affinity       Affinity       `structs:"affinity"`
	Toleration     Toleration     `structs:"toleration"`
	Networking     Networking     `structs:"networking"`
	Security       PodSecurityCtx `structs:"security"`
	ReadinessGates ReadinessGates `structs:"readinessGates"`
	Other          SpecOther      `structs:"other"`
}
