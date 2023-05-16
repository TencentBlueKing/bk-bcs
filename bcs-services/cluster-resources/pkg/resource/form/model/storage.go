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

// PV PersistentVolume 表单化建模
type PV struct {
	Metadata Metadata `structs:"metadata"`
	Spec     PVSpec   `structs:"spec"`
}

// PVSpec ...
type PVSpec struct {
	Type        string   `structs:"type"`
	SCName      string   `structs:"scName"`
	StorageSize int      `structs:"storageSize"`
	AccessModes []string `structs:"accessModes"`
	// Local Volume
	LocalPath string `structs:"localPath"`
	// HostPath
	HostPath     string `structs:"hostPath"`
	HostPathType string `structs:"hostPathType"`
	// NFS Share
	NFSPath     string `structs:"nfsPath"`
	NFSServer   string `structs:"nfsServer"`
	NFSReadOnly bool   `structs:"nfsReadOnly"`
}

// PVC PersistentVolumeClaim 表单化建模
type PVC struct {
	Metadata Metadata `structs:"metadata"`
	Spec     PVCSpec  `structs:"spec"`
}

// PVCSpec ...
type PVCSpec struct {
	ClaimType   string   `structs:"claimType"`
	PVName      string   `structs:"pvName"`
	SCName      string   `structs:"scName"`
	StorageSize int      `structs:"storageSize"`
	AccessModes []string `structs:"accessModes"`
}

// SC StorageClass 表单化建模
type SC struct {
	Metadata Metadata `structs:"metadata"`
	Spec     SCSpec   `structs:"spec"`
}

// SCSpec ...
type SCSpec struct {
	SetAsDefault      bool      `structs:"setAsDefault"`
	Provisioner       string    `structs:"provisioner"`
	VolumeBindingMode string    `structs:"volumeBindingMode"`
	ReclaimPolicy     string    `structs:"reclaimPolicy"`
	Params            []SCParam `structs:"params"`
	MountOpts         []string  `structs:"mountOpts"`
}

// SCParam ...
type SCParam struct {
	Key   string `structs:"key"`
	Value string `structs:"value"`
}
