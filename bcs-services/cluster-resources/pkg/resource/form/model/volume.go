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

// PVCVolume xxx
type PVCVolume struct {
	Name     string `structs:"name"`
	PVCName  string `structs:"pvcName"`
	ReadOnly bool   `structs:"readOnly"`
}

// HostPathVolume xxx
type HostPathVolume struct {
	Name string `structs:"name"`
	Path string `structs:"path"`
	Type string `structs:"type"`
}

// CMVolume xxx
type CMVolume struct {
	Name        string      `structs:"name"`
	DefaultMode string      `structs:"defaultMode"`
	CMName      string      `structs:"cmName"`
	Items       []KeyToPath `structs:"items"`
}

// SecretVolume xxx
type SecretVolume struct {
	Name        string      `structs:"name"`
	DefaultMode string      `structs:"defaultMode"`
	SecretName  string      `structs:"secretName"`
	Items       []KeyToPath `structs:"items"`
}

// KeyToPath xxx
type KeyToPath struct {
	Key  string `structs:"key"`
	Path string `structs:"path"`
}

// EmptyDirVolume xxx
type EmptyDirVolume struct {
	Name string `structs:"name"`
}

// NFSVolume xxx
type NFSVolume struct {
	Name     string `structs:"name"`
	Path     string `structs:"path"`
	Server   string `structs:"server"`
	ReadOnly bool   `structs:"readOnly"`
}
