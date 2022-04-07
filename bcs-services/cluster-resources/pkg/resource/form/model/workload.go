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
	APIVersion     string         `structs:"apiVersion"`
	Kind           string         `structs:"kind"`
	Metadata       Metadata       `structs:"metadata"`
	Spec           DeploySpec     `structs:"spec"`
	Volume         WorkloadVolume `structs:"volume"`
	ContainerGroup ContainerGroup `structs:"containerGroup"`
}

// DeploySpec ...
type DeploySpec struct {
	Replicas   DeployReplicas `structs:"replicas"`
	NodeSelect NodeSelect     `structs:"nodeSelect"`
	Affinity   Affinity       `structs:"affinity"`
	Toleration Toleration     `structs:"toleration"`
	Networking Networking     `structs:"networking"`
	Security   PodSecurityCtx `structs:"security"`
	Other      SpecOther      `structs:"other"`
}

// DeployReplicas ...
type DeployReplicas struct {
	Cnt                  int64  `structs:"cnt"`
	UpdateStrategy       string `structs:"updateStrategy"`
	MaxSurge             int64  `structs:"maxSurge"`
	MSUnit               string `structs:"msUnit"`
	MaxUnavailable       int64  `structs:"maxUnavailable"`
	MUAUnit              string `structs:"muaUnit"`
	MinReadySecs         int64  `structs:"minReadySecs"`
	ProgressDeadlineSecs int64  `structs:"progressDeadlineSecs"`
}

// WorkloadVolume ...
type WorkloadVolume struct {
	PVC       []PVCVolume      `structs:"pvc"`
	HostPath  []HostPathVolume `structs:"hostPath"`
	ConfigMap []CMVolume       `structs:"configMap"`
	Secret    []SecretVolume   `structs:"secret"`
	EmptyDir  []EmptyDirVolume `structs:"emptyDir"`
	NFS       []NFSVolume      `structs:"nfs"`
}

// ContainerGroup ...
type ContainerGroup struct {
	InitContainers []Container `structs:"initContainers"`
	Containers     []Container `structs:"containers"`
}
